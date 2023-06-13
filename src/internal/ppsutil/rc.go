package ppsutil

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	jsonpatch "github.com/evanphx/json-patch"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	workerstats "github.com/pachyderm/pachyderm/v2/src/server/worker/stats"
	"github.com/pachyderm/pachyderm/v2/src/version"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
)

const (
	AppLabel                     = "app"
	PipelineProjectLabel         = "pipelineProject"
	PipelineNameLabel            = "pipelineName"
	PipelineVersionLabel         = "pipelineVersion"
	suite                        = "pachyderm"
	pipelineProjectAnnotation    = "pipelineProject"
	pipelineNameAnnotation       = "pipelineName"
	pachVersionAnnotation        = "pachVersion"
	pipelineVersionAnnotation    = "pipelineVersion"
	pipelineSpecCommitAnnotation = "specCommit"
	hashedAuthTokenAnnotation    = "authTokenHash"
	// WorkerServiceAccountEnvVar is the name of the environment variable used to tell pachd
	// what service account to assign to new worker RCs, for the purpose of
	// creating S3 gateway services.
	WorkerServiceAccountEnvVar = "WORKER_SERVICE_ACCOUNT"
	// DefaultWorkerServiceAccountName is the default value to use if WorkerServiceAccountEnvVar is
	// undefined (for compatibility purposes)
	DefaultWorkerServiceAccountName = "pachyderm-worker"
	// UploadConcurrencyLimitEnvVar is the environment variable for the upload concurrency limit.
	// EnvVar defined in src/internal/serviceenv/config.go

	DefaultUserImage = "ubuntu:20.04"
)

type K8sEnv interface {
	DefaultCPURequest() resource.Quantity
	DefaultMemoryRequest() resource.Quantity
	DefaultStorageRequest() resource.Quantity
	ImagePullSecrets() []string
	S3GatewayPort() uint16
	PostgresSecretRef(context.Context) (*v1.SecretKeySelector, error)
	ImagePullPolicy() string
	WorkerImage() string
	SidecarImage() string
	StorageRoot() string
	Namespace() string
	StorageBackend() string
	PostgresUser() string
	PostgresDatabase() string
	PGBouncerHost() string
	PGBouncerPort() uint16
	PeerPort() uint16
	LokiHost() string
	LokiPort() (uint16, error)
	SidecarPort() uint16
	InSidecars() bool
	GarbageCollectionPercent() int
	SidecarEnvVars(*pps.PipelineInfo, []v1.EnvVar) []v1.EnvVar
	EtcdPrefix() string
	PPSWorkerPort() uint16
	CommitProgressCounterDisabled() bool
	LokiLoggingEnabled() bool
	GoogleCloudProfilerProject() string
	StorageHostPath() string
	TLSSecretName() string
	WorkerSecurityContextsEnabled() bool
	WorkerUsesRoot() bool
}

// Parameters used when creating the kubernetes replication controller in charge
// of a job or pipeline's workers
type workerOptions struct {
	rcName        string // Name of the replication controller managing workers
	specCommit    string // Pipeline spec commit ID (needed for s3 inputs)
	s3GatewayPort int32  // s3 gateway port (if any s3 pipeline inputs)

	userImage               string                // The user's pipeline/job image
	labels                  map[string]string     // k8s labels attached to the RC and workers
	annotations             map[string]string     // k8s annotations attached to the RC and workers
	parallelism             int32                 // Number of replicas the RC maintains
	resourceRequests        *v1.ResourceList      // Resources requested by pipeline/job pods
	resourceLimits          *v1.ResourceList      // Resources requested by pipeline/job pods, applied to the user and init containers
	sidecarResourceLimits   *v1.ResourceList      // Resources requested by pipeline/job pods, applied to the sidecar container
	sidecarResourceRequests *v1.ResourceList      // Resources requested by pipeline/job pods, applied to the sidecar container
	workerEnv               []v1.EnvVar           // Environment vars set in the user container
	volumes                 []v1.Volume           // Volumes that we expose to the user container
	volumeMounts            []v1.VolumeMount      // Paths where we mount each volume in 'volumes'
	postgresSecret          *v1.SecretKeySelector // the reference to the postgres password
	schedulingSpec          *pps.SchedulingSpec   // the SchedulingSpec for the pipeline
	podSpec                 string
	podPatch                string

	// Secrets that we mount in the worker container (e.g. for reading/writing to
	// s3)
	imagePullSecrets []v1.LocalObjectReference
	service          *pps.Service
	tolerations      []v1.Toleration
}

func pipelineLabels(projectName, pipelineName string, pipelineVersion uint64) map[string]string {
	labels := map[string]string{
		AppLabel:             "pipeline",
		PipelineNameLabel:    pipelineName,
		PipelineVersionLabel: fmt.Sprint(pipelineVersion),
		"suite":              suite,
		"component":          "worker",
	}
	if projectName != "" {
		labels[PipelineProjectLabel] = projectName
	}
	return labels
}

func getWorkerOptions(ctx context.Context, env K8sEnv, pi *pps.PipelineInfo) (*workerOptions, error) {
	projectName := pi.Pipeline.Project.GetName()
	pipelineName := pi.Pipeline.Name
	pipelineVersion := pi.Version
	var resourceRequests *v1.ResourceList
	var resourceLimits *v1.ResourceList
	var sidecarResourceLimits *v1.ResourceList
	var sidecarResourceRequests *v1.ResourceList
	if pi.Details.ResourceRequests != nil {
		var err error
		resourceRequests, err = GetRequestsResourceListFromPipeline(ctx, pi)
		if err != nil {
			return nil, errors.Wrapf(err, "could not determine resource request")
		}
	}
	if pi.Details.ResourceLimits != nil {
		var err error
		resourceLimits, err = GetLimitsResourceList(ctx, pi.Details.ResourceLimits)
		if err != nil {
			return nil, errors.Wrapf(err, "could not determine resource limit")
		}
	}
	// Users are often surprised when their jobs that claim to require 0 CPU are scheduled to a
	// node with no free CPU resources.  This avoids that; if resources are completely omitted
	// in the pipeline spec, then we supply some reasonable defaults.  You can supply an empty
	// ResrouceRequests in your pipeline to avoid these defaults; this will explicitly request
	// the old behavior of not making any requests.
	if pi.Details.ResourceRequests == nil && pi.Details.ResourceLimits == nil {
		resourceRequests = &v1.ResourceList{
			v1.ResourceCPU:              env.DefaultCPURequest(),
			v1.ResourceMemory:           env.DefaultMemoryRequest(),
			v1.ResourceEphemeralStorage: env.DefaultStorageRequest(),
		}
		log.Info(ctx, "setting default resource requests on pipeline; supply an empty resource request ('resource_requests: {}') to opt out",
			zap.String("pipeline", pi.GetPipeline().GetName()),
			zap.Reflect("requests", resourceRequests))
	}
	if pi.Details.SidecarResourceLimits != nil {
		var err error
		sidecarResourceLimits, err = GetLimitsResourceList(ctx, pi.Details.SidecarResourceLimits)
		if err != nil {
			return nil, errors.Wrapf(err, "could not determine sidecar resource limit")
		}
	}
	if pi.Details.SidecarResourceRequests != nil {
		var err error
		sidecarResourceRequests, err = GetLimitsResourceList(ctx, pi.Details.SidecarResourceRequests)
		if err != nil {
			return nil, errors.Wrapf(err, "could not determine sidecar resource request")
		}
	}

	transform := pi.Details.Transform
	labels := pipelineLabels(projectName, pipelineName, pipelineVersion)
	userImage := transform.Image
	if userImage == "" {
		userImage = DefaultUserImage
	}

	workerEnv := []v1.EnvVar{}
	for name, value := range transform.Env {
		workerEnv = append(
			workerEnv,
			v1.EnvVar{
				Name:  name,
				Value: value,
			},
		)
	}

	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount
	for _, secret := range transform.Secrets {
		if secret.MountPath != "" {
			volumes = append(volumes, v1.Volume{
				Name: secret.Name,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: secret.Name,
					},
				},
			})
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      secret.Name,
				MountPath: secret.MountPath,
			})
		}
		if secret.EnvVar != "" {
			workerEnv = append(workerEnv, v1.EnvVar{
				Name: secret.EnvVar,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: secret.Name,
						},
						Key: secret.Key,
					},
				},
			})
		}
	}

	volumes = append(volumes, v1.Volume{
		Name: "pach-bin",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})
	volumeMounts = append(volumeMounts, v1.VolumeMount{
		Name:      "pach-bin",
		MountPath: "/pach-bin",
	})
	workerVolume := v1.Volume{
		Name: client.PPSWorkerVolume,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
	if transform.MemoryVolume {
		workerVolume.VolumeSource.EmptyDir.Medium = v1.StorageMediumMemory
	}
	volumes = append(volumes, workerVolume)
	volumeMounts = append(volumeMounts, v1.VolumeMount{
		Name:      client.PPSWorkerVolume,
		MountPath: client.PPSInputPrefix,
	})
	var imagePullSecrets []v1.LocalObjectReference
	for _, secret := range transform.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, v1.LocalObjectReference{Name: secret})
	}
	for _, secret := range env.ImagePullSecrets() {
		imagePullSecrets = append(imagePullSecrets, v1.LocalObjectReference{Name: secret})
	}

	annotations := map[string]string{
		pipelineNameAnnotation:       pipelineName,
		pachVersionAnnotation:        version.PrettyVersion(),
		pipelineVersionAnnotation:    strconv.FormatUint(pi.Version, 10),
		pipelineSpecCommitAnnotation: pi.SpecCommit.ID,
		hashedAuthTokenAnnotation:    HashedAuthToken(pi.AuthToken),
	}
	if projectName != "" {
		annotations[pipelineProjectAnnotation] = projectName
	}

	// add the user's custom metadata (annotations and labels).
	metadata := pi.Details.GetMetadata()
	if metadata != nil {
		for k, v := range metadata.Annotations {
			if annotations[k] == "" {
				annotations[k] = v
			}
		}

		for k, v := range metadata.Labels {
			if labels[k] == "" {
				labels[k] = v
			}
		}
	}

	// A service can be present either directly on the pipeline spec
	// or on the spout field of the spec.
	var service *pps.Service
	if pi.Details.Spout != nil && pi.Details.Service != nil {
		return nil, errors.New("only one of pipeline.service or pipeline.spout can be set")
	} else if pi.Details.Spout != nil && pi.Details.Spout.Service != nil {
		service = pi.Details.Spout.Service
	} else {
		service = pi.Details.Service
	}
	var s3GatewayPort int32
	if ContainsS3Inputs(pi.Details.Input) || pi.Details.S3Out {
		s3GatewayPort = int32(env.S3GatewayPort())
	}

	// // Get the reference to the postgres secret used by the current pod
	// podName := kd.config.PachdPodName
	// selfPodInfo, err := kd.kubeClient.CoreV1().Pods(kd.namespace).Get(ctx, podName, metav1.GetOptions{})
	// if err != nil {
	// 	return nil, errors.EnsureStack(err)
	// }
	// var postgresSecretRef *v1.SecretKeySelector
	// for _, container := range selfPodInfo.Spec.Containers {
	// 	for _, envVar := range container.Env {
	// 		if envVar.Name == "POSTGRES_PASSWORD" && envVar.ValueFrom != nil && envVar.ValueFrom.SecretKeyRef != nil {
	// 			postgresSecretRef = envVar.ValueFrom.SecretKeyRef
	// 		}
	// 	}
	// }
	// if postgresSecretRef == nil {
	// 	return nil, errors.New("could not load the existing postgres secret reference from kubernetes")
	// }
	postgresSecretRef, err := env.PostgresSecretRef(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get PostgresSQL secret reference")
	}
	if postgresSecretRef == nil {
		return nil, errors.New("did not get a PostgresSQL secret reference")
	}

	// Setup tolerations
	var tolerations []v1.Toleration
	var tolErr error
	for i, in := range pi.GetDetails().GetTolerations() {
		out, err := transformToleration(in)
		if err != nil {
			errors.JoinInto(&tolErr, errors.Errorf("toleration %d/%d: %v", i+1, len(pi.GetDetails().GetTolerations()), err))
			continue
		}
		tolerations = append(tolerations, out)
	}
	if tolErr != nil {
		return nil, tolErr
	}

	// Generate options for new RC
	return &workerOptions{
		rcName:                  PipelineRcName(pi),
		s3GatewayPort:           s3GatewayPort,
		specCommit:              pi.SpecCommit.ID,
		labels:                  labels,
		annotations:             annotations,
		parallelism:             int32(0), // pipelines start w/ 0 workers & are scaled up
		resourceRequests:        resourceRequests,
		resourceLimits:          resourceLimits,
		sidecarResourceLimits:   sidecarResourceLimits,
		sidecarResourceRequests: sidecarResourceRequests,
		userImage:               userImage,
		workerEnv:               workerEnv,
		volumes:                 volumes,
		volumeMounts:            volumeMounts,
		postgresSecret:          postgresSecretRef,
		imagePullSecrets:        imagePullSecrets,
		service:                 service,
		schedulingSpec:          pi.Details.SchedulingSpec,
		podSpec:                 pi.Details.PodSpec,
		podPatch:                pi.Details.PodPatch,
		tolerations:             tolerations,
	}, nil
}

// transformToleration transforms a pps.Toleration into a k8s Toleration.  It's used while creating
// the RC, and at pipeline submission to validate the provided tolerations before saving the
// pipeline.  It's in here and not in src/pps/pps.go so that the protobuf library doesn't depend on
// k8s.
func transformToleration(in *pps.Toleration) (v1.Toleration, error) {
	var out v1.Toleration
	out.Key = in.GetKey()
	out.Value = in.GetValue()
	if ts := in.GetTolerationSeconds(); ts != nil {
		out.TolerationSeconds = &ts.Value
	}
	switch in.GetEffect() { //exhaustive:enforce
	case pps.TaintEffect_ALL_EFFECTS:
		out.Effect = ""
	case pps.TaintEffect_NO_EXECUTE:
		out.Effect = v1.TaintEffectNoExecute
	case pps.TaintEffect_NO_SCHEDULE:
		out.Effect = v1.TaintEffectNoSchedule
	case pps.TaintEffect_PREFER_NO_SCHEDULE:
		out.Effect = v1.TaintEffectPreferNoSchedule
	}
	switch in.GetOperator() { //exhaustive:enforce
	case pps.TolerationOperator_EMPTY:
		return out, errors.New("cannot omit key/value comparison operator; specify EQUAL or EXISTS")
	case pps.TolerationOperator_EXISTS:
		out.Operator = v1.TolerationOpExists
	case pps.TolerationOperator_EQUAL:
		out.Operator = v1.TolerationOpEqual
	}
	return out, nil
}

// We don't want to expose pipeline auth tokens, so we hash it. This will be
// visible to any user with k8s cluster access
// Note: This hash shouldn't be used for authentication in any way. We just use
// this to detect if an auth token has been added/changed
// Note: ptr.AuthToken is a pachyderm-generated UUID, and wouldn't appear in any
// rainbow tables
func HashedAuthToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func workerPodSpec(ctx context.Context, env K8sEnv, options *workerOptions, pipelineInfo *pps.PipelineInfo) (v1.PodSpec, error) {
	pullPolicy := env.ImagePullPolicy()
	if pullPolicy == "" {
		pullPolicy = "IfNotPresent"
	}

	// Environment variables that are shared between both containers
	lokiPort, err := env.LokiPort()
	if err != nil {
		return v1.PodSpec{}, errors.Wrap(err, "could not get Loki port from k8s env")
	}
	commonEnv := []v1.EnvVar{{
		Name:  "PACH_ROOT",
		Value: env.StorageRoot(),
	}, {
		Name:  "PACH_NAMESPACE",
		Value: env.Namespace(),
	}, {
		Name:  "STORAGE_BACKEND",
		Value: env.StorageBackend(),
	}, {
		Name:  "POSTGRES_USER",
		Value: env.PostgresUser(),
	}, {
		Name: "POSTGRES_PASSWORD",
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: options.postgresSecret,
		},
	}, {
		Name:  "POSTGRES_DATABASE",
		Value: env.PostgresDatabase(),
	}, {
		Name:  "PG_BOUNCER_HOST",
		Value: env.PGBouncerHost(),
	}, {
		Name:  "PG_BOUNCER_PORT",
		Value: strconv.FormatInt(int64(env.PGBouncerPort()), 10),
	}, {
		Name:  client.PeerPortEnv,
		Value: strconv.FormatUint(uint64(env.PeerPort()), 10),
	}, {
		Name:  client.PPSSpecCommitEnv,
		Value: options.specCommit,
	}, {
		Name:  client.PPSProjectNameEnv,
		Value: pipelineInfo.Pipeline.Project.GetName(),
	}, {
		Name:  client.PPSPipelineNameEnv,
		Value: pipelineInfo.Pipeline.Name,
	}, {
		Name:  "LOKI_SERVICE_HOST",
		Value: env.LokiHost(),
	}, {
		Name:  "LOKI_SERVICE_PORT",
		Value: strconv.FormatUint(uint64(lokiPort), 10),
	}, {
		Name:  "GOCOVERDIR",
		Value: "/tmp",
	},
		// These are set explicitly below to prevent kubernetes from setting them to the service host and port.
		{
			Name:  "POSTGRES_PORT",
			Value: "",
		}, {
			Name:  "POSTGRES_HOST",
			Value: "",
		},
	}
	commonEnv = append(commonEnv, log.WorkerLogConfig.AsKubernetesEnvironment()...)
	fmt.Println("QQQ commonEnv", commonEnv)

	// Set up sidecar env vars
	sidecarEnv := []v1.EnvVar{{
		Name:  "PORT",
		Value: strconv.FormatUint(uint64(env.SidecarPort()), 10),
	}, {
		Name: "PACHD_POD_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	}, {
		Name:  "PACHW_IN_SIDECARS",
		Value: strconv.FormatBool(env.InSidecars()),
	}, {
		Name:  "GC_PERCENT",
		Value: strconv.FormatInt(int64(env.GarbageCollectionPercent()), 10),
	}}

	sidecarEnv = append(sidecarEnv, env.SidecarEnvVars(pipelineInfo, commonEnv)...)

	// Set up worker env vars
	workerEnv := append(options.workerEnv, []v1.EnvVar{
		// Set core pach env vars
		{
			Name:  "PACH_IN_WORKER",
			Value: "true",
		},
		// We use Kubernetes' "Downward API" so the workers know their IP
		// addresses, which they will then post on etcd so the job managers
		// can discover the workers.
		{
			Name: client.PPSWorkerIPEnv,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
		// Set the PPS env vars
		{
			Name:  client.PPSEtcdPrefixEnv,
			Value: env.EtcdPrefix(),
		},
		{
			Name: client.PPSPodNameEnv,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name:  client.PPSWorkerPortEnv,
			Value: strconv.FormatUint(uint64(env.PPSWorkerPort()), 10),
		},
	}...)
	workerEnv = append(workerEnv, commonEnv...)

	// Set S3GatewayPort in the worker (for user code) and sidecar (for serving)
	if options.s3GatewayPort != 0 {
		workerEnv = append(workerEnv, v1.EnvVar{
			Name:  "S3GATEWAY_PORT",
			Value: strconv.FormatUint(uint64(options.s3GatewayPort), 10),
		})
		sidecarEnv = append(sidecarEnv, v1.EnvVar{
			Name:  "S3GATEWAY_PORT",
			Value: strconv.FormatUint(uint64(options.s3GatewayPort), 10),
		})
	}
	// Propagate feature flags to worker and sidecar
	if env.CommitProgressCounterDisabled() {
		sidecarEnv = append(sidecarEnv, v1.EnvVar{Name: "DISABLE_COMMIT_PROGRESS_COUNTER", Value: "true"})
		workerEnv = append(workerEnv, v1.EnvVar{Name: "DISABLE_COMMIT_PROGRESS_COUNTER", Value: "true"})
	}
	if env.LokiLoggingEnabled() {
		sidecarEnv = append(sidecarEnv, v1.EnvVar{Name: "LOKI_LOGGING", Value: "true"})
		workerEnv = append(workerEnv, v1.EnvVar{Name: "LOKI_LOGGING", Value: "true"})
	}
	if p := env.GoogleCloudProfilerProject(); p != "" {
		sidecarEnv = append(sidecarEnv, v1.EnvVar{Name: "GOOGLE_CLOUD_PROFILER_PROJECT", Value: p})
		workerEnv = append(workerEnv, v1.EnvVar{Name: "GOOGLE_CLOUD_PROFILER_PROJECT", Value: p})
	}

	// This only happens in local deployment.  We want the workers to be
	// able to read from/write to the hostpath volume as well.
	storageVolumeName := "pach-disk"
	var sidecarVolumeMounts []v1.VolumeMount
	userVolumeMounts := make([]v1.VolumeMount, len(options.volumeMounts))
	copy(userVolumeMounts, options.volumeMounts)
	if storageHostPath := env.StorageHostPath(); storageHostPath != "" {
		options.volumes = append(options.volumes, v1.Volume{
			Name: storageVolumeName,
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: storageHostPath,
				},
			},
		})
		storageMount := v1.VolumeMount{
			Name:      storageVolumeName,
			MountPath: env.StorageRoot(),
		}
		sidecarVolumeMounts = append(sidecarVolumeMounts, storageMount)
		userVolumeMounts = append(userVolumeMounts, storageMount)
	} else {
		// `pach-dir-volume` is needed for openshift, see:
		// https://github.com/pachyderm/pachyderm/v2/issues/3404
		options.volumes = append(options.volumes, v1.Volume{
			Name: "pach-dir-volume",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		}, v1.Volume{
			Name: "user-tmp",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
		emptyDirVolumeMount := v1.VolumeMount{
			Name:      "pach-dir-volume",
			MountPath: env.StorageRoot(),
		}
		userTmpVolumeMount := v1.VolumeMount{
			Name:      "user-tmp",
			MountPath: "/tmp",
		}
		sidecarVolumeMounts = append(sidecarVolumeMounts, emptyDirVolumeMount)
		userVolumeMounts = append(userVolumeMounts, emptyDirVolumeMount, userTmpVolumeMount)
	}
	// add emptydir for /tmp to allow for read only rootfs
	options.volumes = append(options.volumes, v1.Volume{
		Name: "tmp",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		}})
	tmpDirVolumeMount := v1.VolumeMount{
		Name:      "tmp",
		MountPath: "/tmp",
	}
	secretVolume, secretMount := GetBackendSecretVolumeAndMount()
	options.volumes = append(options.volumes, secretVolume)
	sidecarVolumeMounts = append(sidecarVolumeMounts, secretMount, tmpDirVolumeMount)
	userVolumeMounts = append(userVolumeMounts, secretMount)

	// in the case the pachd is deployed with custom root certs, propagate them to the side-cars
	if path, ok := os.LookupEnv("SSL_CERT_DIR"); ok {
		sidecarEnv = append(sidecarEnv, v1.EnvVar{Name: "SSL_CERT_DIR", Value: path})
		certSecretVolume, certSecretMount := getTLSCertSecretVolumeAndMount(env.TLSSecretName(), path)
		options.volumes = append(options.volumes, certSecretVolume)
		sidecarVolumeMounts = append(sidecarVolumeMounts, certSecretMount)
	}

	// mount secret for spouts using pachctl
	if pipelineInfo.Details.Spout != nil {
		pachctlSecretVolume, pachctlSecretMount := getPachctlSecretVolumeAndMount(spoutSecretName(pipelineInfo.Pipeline))
		options.volumes = append(options.volumes, pachctlSecretVolume)
		sidecarVolumeMounts = append(sidecarVolumeMounts, pachctlSecretMount)
		userVolumeMounts = append(userVolumeMounts, pachctlSecretMount)
	}

	// Explicitly set CPU requests to zero because some cloud providers set their
	// own defaults which are usually not what we want. Mem request defaults to
	// 64M, but is overridden by the CacheSize setting for the sidecar.
	cpuZeroQuantity := resource.MustParse("0")
	memDefaultQuantity := resource.MustParse("64M")
	memSidecarQuantity := resource.MustParse("64M")

	// Get service account name for worker from env or use default
	workerServiceAccountName, ok := os.LookupEnv(WorkerServiceAccountEnvVar)
	if !ok {
		workerServiceAccountName = DefaultWorkerServiceAccountName
	}

	// possibly expose s3 gateway port in the sidecar container
	var sidecarPorts []v1.ContainerPort
	if options.s3GatewayPort != 0 {
		sidecarPorts = append(sidecarPorts, v1.ContainerPort{
			ContainerPort: options.s3GatewayPort,
		})
	}

	workerImage := env.WorkerImage()
	pachSecurityCtx := &v1.SecurityContext{
		RunAsUser:                int64Ptr(1000),
		RunAsGroup:               int64Ptr(1000),
		AllowPrivilegeEscalation: pointer.Bool(false),
		ReadOnlyRootFilesystem:   pointer.Bool(true),
		Capabilities:             &v1.Capabilities{Drop: []v1.Capability{"all"}},
	}
	var userSecurityCtx *v1.SecurityContext
	var podSecurityContext *v1.PodSecurityContext
	userStr := pipelineInfo.Details.Transform.User
	if !env.WorkerSecurityContextsEnabled() {
		pachSecurityCtx = nil
		podSecurityContext = nil
	} else if env.WorkerUsesRoot() {
		pachSecurityCtx = &v1.SecurityContext{RunAsUser: int64Ptr(0)}
		userSecurityCtx = &v1.SecurityContext{RunAsUser: int64Ptr(0)}
		podSecurityContext = nil
	} else if userStr != "" {
		// This is to allow the user to be set in the pipeline spec.
		if i, err := strconv.ParseInt(userStr, 10, 64); err != nil {
			log.Error(ctx, "could not parse user into int", zap.String("user", userStr), zap.Error(err))
		} else {
			// hard coded security settings besides uid/gid.
			podSecurityContext = &v1.PodSecurityContext{
				RunAsUser:    int64Ptr(i),
				RunAsGroup:   int64Ptr(i),
				FSGroup:      int64Ptr(i),
				RunAsNonRoot: pointer.Bool(true),
				SeccompProfile: &v1.SeccompProfile{
					Type: v1.SeccompProfileType("RuntimeDefault"),
				}}
			userSecurityCtx = &v1.SecurityContext{
				RunAsUser:                int64Ptr(i),
				RunAsGroup:               int64Ptr(i),
				AllowPrivilegeEscalation: pointer.Bool(false),
				ReadOnlyRootFilesystem:   pointer.Bool(true),
				Capabilities:             &v1.Capabilities{Drop: []v1.Capability{"all"}},
			}
		}
	}
	envFrom := []v1.EnvFromSource{
		{
			SecretRef: &v1.SecretEnvSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: client.StorageSecretName,
				},
			},
		},
	}
	podSpec := v1.PodSpec{
		InitContainers: []v1.Container{
			{
				Name:            "init",
				Image:           workerImage,
				Command:         []string{"/app/init"},
				ImagePullPolicy: v1.PullPolicy(pullPolicy),
				VolumeMounts:    options.volumeMounts,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    cpuZeroQuantity,
						v1.ResourceMemory: memDefaultQuantity,
					},
				},
				SecurityContext: pachSecurityCtx,
			},
		},
		Containers: []v1.Container{
			{
				Name:            client.PPSWorkerUserContainerName,
				Image:           options.userImage,
				Command:         []string{"/pach-bin/dumb-init", "--", "/pach-bin/worker"},
				ImagePullPolicy: v1.PullPolicy(pullPolicy),
				Env:             workerEnv,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{},
				},
				VolumeMounts:    userVolumeMounts,
				SecurityContext: userSecurityCtx,
			},
			{
				Name:            client.PPSWorkerSidecarContainerName,
				Image:           env.SidecarImage(),
				Command:         []string{"/pachd", "--mode", "sidecar"},
				ImagePullPolicy: v1.PullPolicy(pullPolicy),
				Env:             sidecarEnv,
				EnvFrom:         envFrom,
				VolumeMounts:    sidecarVolumeMounts,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    cpuZeroQuantity,
						v1.ResourceMemory: memSidecarQuantity,
					},
				},
				Ports:           sidecarPorts,
				SecurityContext: pachSecurityCtx,
			},
		},
		ServiceAccountName:            workerServiceAccountName,
		AutomountServiceAccountToken:  pointer.Bool(true),
		RestartPolicy:                 "Always",
		Volumes:                       options.volumes,
		ImagePullSecrets:              options.imagePullSecrets,
		TerminationGracePeriodSeconds: int64Ptr(0),
		SecurityContext:               podSecurityContext,
		Tolerations:                   options.tolerations,
	}
	if options.schedulingSpec != nil {
		podSpec.NodeSelector = options.schedulingSpec.NodeSelector
		podSpec.PriorityClassName = options.schedulingSpec.PriorityClassName
	}

	if options.resourceRequests != nil {
		for k, v := range *options.resourceRequests {
			podSpec.Containers[0].Resources.Requests[k] = v
		}
	}

	// Copy over some settings from the user to init container. We don't apply GPU
	// requests because of FUD. The init container shouldn't run concurrently with
	// the user container, so there should be no contention here.
	for _, k := range []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory, v1.ResourceEphemeralStorage} {
		if val, ok := podSpec.Containers[0].Resources.Requests[k]; ok {
			podSpec.InitContainers[0].Resources.Requests[k] = val
		}
	}

	if options.resourceLimits != nil {
		podSpec.InitContainers[0].Resources.Limits = make(v1.ResourceList)
		podSpec.Containers[0].Resources.Limits = make(v1.ResourceList)
		for k, v := range *options.resourceLimits {
			podSpec.InitContainers[0].Resources.Limits[k] = v
			podSpec.Containers[0].Resources.Limits[k] = v
		}
	}

	if options.sidecarResourceLimits != nil {
		podSpec.Containers[1].Resources.Limits = make(v1.ResourceList)
		for k, v := range *options.sidecarResourceLimits {
			podSpec.Containers[1].Resources.Limits[k] = v
		}
	}

	if options.sidecarResourceRequests != nil {
		podSpec.Containers[1].Resources.Requests = make(v1.ResourceList)
		for k, v := range *options.sidecarResourceRequests {
			podSpec.Containers[1].Resources.Requests[k] = v
		}
	}

	if options.podSpec != "" || options.podPatch != "" {
		jsonPodSpec, err := json.Marshal(&podSpec)
		if err != nil {
			return v1.PodSpec{}, errors.EnsureStack(err)
		}

		// the json now contained in jsonPodSpec is the authoritative copy
		// so we should deserialize in into a fresh structure
		podSpec = v1.PodSpec{}

		if options.podSpec != "" {
			jsonPodSpec, err = jsonpatch.MergePatch(jsonPodSpec, []byte(options.podSpec))
			if err != nil {
				return v1.PodSpec{}, errors.EnsureStack(err)
			}
		}
		if options.podPatch != "" {
			patch, err := jsonpatch.DecodePatch([]byte(options.podPatch))
			if err != nil {
				return v1.PodSpec{}, errors.EnsureStack(err)
			}
			jsonPodSpec, err = patch.Apply(jsonPodSpec)
			if err != nil {
				return v1.PodSpec{}, errors.EnsureStack(err)
			}
		}
		if err := json.Unmarshal(jsonPodSpec, &podSpec); err != nil {
			return v1.PodSpec{}, errors.EnsureStack(err)
		}
	}
	return podSpec, nil
}

type PipelineSpecs struct {
	Secret                *v1.Secret
	ReplicationController *v1.ReplicationController
	Services              []*v1.Service
}

func workerPachctlSecret(pipelineInfo *pps.PipelineInfo) (*v1.Secret, error) {
	var cfg config.Config
	err := cfg.InitV2()
	if err != nil {
		return nil, errors.Wrapf(err, "error initializing V2 for config")
	}
	_, context, err := cfg.ActiveContext(true)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting the active context")
	}
	context.SessionToken = pipelineInfo.AuthToken
	context.PachdAddress = "localhost:1653"

	rawConfig, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, errors.Wrapf(err, "error marshaling the config")
	}
	s := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   spoutSecretName(pipelineInfo.Pipeline),
			Labels: spoutLabels(pipelineInfo.Pipeline),
		},
		Data: map[string][]byte{
			"config.json": rawConfig,
		},
	}
	labels := s.GetLabels()
	if projectName := pipelineInfo.Pipeline.Project.GetName(); projectName != "" {
		labels[PipelineProjectLabel] = projectName
	}
	labels[PipelineNameLabel] = pipelineInfo.Pipeline.Name
	s.SetLabels(labels)

	return s, nil
}
func spoutLabels(pipeline *pps.Pipeline) map[string]string {
	m := map[string]string{
		AppLabel:          "spout",
		PipelineNameLabel: pipeline.Name,
		"suite":           suite,
		"component":       "worker",
	}
	if projectName := pipeline.Project.GetName(); projectName != "" {
		m[PipelineProjectLabel] = projectName
	}
	return m
}

func SpecsFromPipelineInfo(ctx context.Context, env K8sEnv, pi *pps.PipelineInfo) (PipelineSpecs, error) {
	var specs PipelineSpecs
	options, err := getWorkerOptions(ctx, env, pi)
	if err != nil {
		return specs, NoValidOptionsErr{err}
	}

	if pi.Details.Spout != nil {
		if specs.Secret, err = workerPachctlSecret(pi); err != nil {
			return specs, errors.Wrap(err, "")
		}
	}

	podSpec, err := workerPodSpec(ctx, env, options, pi)
	if err != nil {
		return specs, errors.Wrap(err, "could not generate worker pod spec")
	}

	specs.ReplicationController = &v1.ReplicationController{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.rcName,
			Labels:      options.labels,
			Annotations: options.annotations,
		},
		Spec: v1.ReplicationControllerSpec{
			Selector: options.labels,
			Replicas: &options.parallelism,
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        options.rcName,
					Labels:      options.labels,
					Annotations: options.annotations,
				},
				Spec: podSpec,
			},
		},
	}
	serviceAnnotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   strconv.Itoa(workerstats.PrometheusPort),
	}

	specs.Services = append(specs.Services, &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.rcName,
			Labels:      options.labels,
			Annotations: serviceAnnotations,
		},
		Spec: v1.ServiceSpec{
			Selector:  options.labels,
			ClusterIP: "None", // headless, so as not to consume an IP address in the cluster
			Ports: []v1.ServicePort{
				{
					Port: int32(env.PPSWorkerPort()),
					Name: "grpc-port",
				},
				{
					Port: workerstats.PrometheusPort,
					Name: "prom-metrics",
				},
			},
		},
	})

	if options.service != nil {
		var servicePort = []v1.ServicePort{
			{
				Port:       options.service.ExternalPort,
				TargetPort: intstr.FromInt(int(options.service.InternalPort)),
				Name:       "user-port",
			},
		}
		var serviceType = v1.ServiceType(options.service.Type)
		if serviceType == v1.ServiceTypeNodePort {
			servicePort[0].NodePort = options.service.ExternalPort
		}
		specs.Services = append(specs.Services, &v1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        options.rcName + "-user",
				Labels:      options.labels,
				Annotations: options.annotations,
			},
			Spec: v1.ServiceSpec{
				Selector: options.labels,
				Type:     serviceType,
				Ports:    servicePort,
			},
		})
	}
	return specs, nil
}

// noValidOptions error may be returned by createWorkerSvcAndRc to indicate that
// getWorkerOptions returned an error to it (getWorkerOptions does not return
// noValidOptions). This is a mechanism for createWorkerSvcAndRc to signal to
// its caller not to retry.
//
// FIXME: wrap?
type NoValidOptionsErr struct {
	error
}

// GetBackendSecretVolumeAndMount returns a properly configured Volume and
// VolumeMount object
//
// FIXME: remove
func GetBackendSecretVolumeAndMount() (v1.Volume, v1.VolumeMount) {
	return v1.Volume{
			Name: client.StorageSecretName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: client.StorageSecretName,
				},
			},
		}, v1.VolumeMount{
			Name:      client.StorageSecretName,
			MountPath: "/" + client.StorageSecretName,
		}
}

// getPachctlSecretVolumeAndMount returns a Volume and
// VolumeMount object configured for the pachctl secret (currently used in spout pipelines).
func getPachctlSecretVolumeAndMount(secret string) (v1.Volume, v1.VolumeMount) {
	return v1.Volume{
			Name: client.PachctlSecretName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
				},
			},
		}, v1.VolumeMount{
			Name:      client.PachctlSecretName,
			MountPath: "/pachctl",
		}
}

// getTLSCertSecretVolumeAndMount returns a Volume and VolumeMount object
// configured for the pach-tls secret to be stored in pipeline side-cars.
func getTLSCertSecretVolumeAndMount(secret, mountPath string) (v1.Volume, v1.VolumeMount) {
	return v1.Volume{
			Name: secret,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
				},
			},
		}, v1.VolumeMount{
			Name:      secret,
			MountPath: mountPath,
		}
}

func spoutSecretName(p *pps.Pipeline) string {
	if projectName := p.Project.GetName(); projectName != "" {
		return fmt.Sprintf("spout-pachctl-secret-%s-%s", projectName, p.Name)
	}
	return fmt.Sprintf("spout-pachctl-secret-%s", p.Name)
}

func int64Ptr(x int64) *int64 {
	return &x
}
