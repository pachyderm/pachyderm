package server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	client "github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	workerstats "github.com/pachyderm/pachyderm/v2/src/server/worker/stats"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

const (
	appLabel                     = "app"
	pipelineNameLabel            = "pipelineName"
	pipelineVersionLabel         = "pipelineVersion"
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
	UploadConcurrencyLimitEnvVar = "STORAGE_UPLOAD_CONCURRENCY_LIMIT"
)

// Parameters used when creating the kubernetes replication controller in charge
// of a job or pipeline's workers
type workerOptions struct {
	rcName        string // Name of the replication controller managing workers
	specCommit    string // Pipeline spec commit ID (needed for s3 inputs)
	s3GatewayPort int32  // s3 gateway port (if any s3 pipeline inputs)

	userImage             string                // The user's pipeline/job image
	labels                map[string]string     // k8s labels attached to the RC and workers
	annotations           map[string]string     // k8s annotations attached to the RC and workers
	parallelism           int32                 // Number of replicas the RC maintains
	resourceRequests      *v1.ResourceList      // Resources requested by pipeline/job pods
	resourceLimits        *v1.ResourceList      // Resources requested by pipeline/job pods, applied to the user and init containers
	sidecarResourceLimits *v1.ResourceList      // Resources requested by pipeline/job pods, applied to the sidecar container
	workerEnv             []v1.EnvVar           // Environment vars set in the user container
	volumes               []v1.Volume           // Volumes that we expose to the user container
	volumeMounts          []v1.VolumeMount      // Paths where we mount each volume in 'volumes'
	postgresSecret        *v1.SecretKeySelector // the reference to the postgres password
	schedulingSpec        *pps.SchedulingSpec   // the SchedulingSpec for the pipeline
	podSpec               string
	podPatch              string

	// Secrets that we mount in the worker container (e.g. for reading/writing to
	// s3)
	imagePullSecrets []v1.LocalObjectReference
	service          *pps.Service
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

func (kd *kubeDriver) workerPodSpec(options *workerOptions, pipelineInfo *pps.PipelineInfo) (v1.PodSpec, error) {
	pullPolicy := kd.config.WorkerImagePullPolicy
	if pullPolicy == "" {
		pullPolicy = "IfNotPresent"
	}

	// Environment variables that are shared between both containers
	commonEnv := []v1.EnvVar{{
		Name:  "PACH_ROOT",
		Value: kd.config.StorageRoot,
	}, {
		Name:  "PACH_NAMESPACE",
		Value: kd.namespace,
	}, {
		Name:  "STORAGE_BACKEND",
		Value: kd.config.StorageBackend,
	}, {
		Name:  "POSTGRES_USER",
		Value: kd.config.PostgresUser,
	}, {
		Name: "POSTGRES_PASSWORD",
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: options.postgresSecret,
		},
	}, {
		Name:  "POSTGRES_DATABASE",
		Value: kd.config.PostgresDBName,
	}, {
		Name:  "PG_BOUNCER_HOST",
		Value: kd.config.PGBouncerHost,
	}, {
		Name:  "PG_BOUNCER_PORT",
		Value: strconv.FormatInt(int64(kd.config.PGBouncerPort), 10),
	}, {
		Name:  client.PeerPortEnv,
		Value: strconv.FormatUint(uint64(kd.config.PeerPort), 10),
	}, {
		Name:  client.PPSSpecCommitEnv,
		Value: options.specCommit,
	}, {
		Name:  client.PPSPipelineNameEnv,
		Value: pipelineInfo.Pipeline.Name,
	}, {
		Name:  "LOKI_SERVICE_HOST",
		Value: kd.config.LokiHost,
	}, {
		Name:  "LOKI_SERVICE_PORT",
		Value: kd.config.LokiPort,
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

	// Set up sidecar env vars
	sidecarEnv := []v1.EnvVar{{
		Name:  "PORT",
		Value: strconv.FormatUint(uint64(kd.config.Port), 10),
	}, {
		Name: "PACHD_POD_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	}, {
		Name:  "GC_PERCENT",
		Value: strconv.FormatInt(int64(kd.config.GCPercent), 10),
	}}

	sidecarEnv = append(sidecarEnv, kd.getStorageEnvVars(pipelineInfo)...)
	sidecarEnv = append(sidecarEnv, commonEnv...)
	sidecarEnv = append(sidecarEnv, kd.getEgressSecretEnvVars(pipelineInfo)...)

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
			Value: kd.etcdPrefix,
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
			Value: strconv.FormatUint(uint64(kd.config.PPSWorkerPort), 10),
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
	if kd.config.DisableCommitProgressCounter {
		sidecarEnv = append(sidecarEnv, v1.EnvVar{Name: "DISABLE_COMMIT_PROGRESS_COUNTER", Value: "true"})
		workerEnv = append(workerEnv, v1.EnvVar{Name: "DISABLE_COMMIT_PROGRESS_COUNTER", Value: "true"})
	}
	if kd.config.LokiLogging {
		sidecarEnv = append(sidecarEnv, v1.EnvVar{Name: "LOKI_LOGGING", Value: "true"})
		workerEnv = append(workerEnv, v1.EnvVar{Name: "LOKI_LOGGING", Value: "true"})
	}
	if p := kd.config.GoogleCloudProfilerProject; p != "" {
		sidecarEnv = append(sidecarEnv, v1.EnvVar{Name: "GOOGLE_CLOUD_PROFILER_PROJECT", Value: p})
		workerEnv = append(workerEnv, v1.EnvVar{Name: "GOOGLE_CLOUD_PROFILER_PROJECT", Value: p})
	}

	// This only happens in local deployment.  We want the workers to be
	// able to read from/write to the hostpath volume as well.
	storageVolumeName := "pach-disk"
	var sidecarVolumeMounts []v1.VolumeMount
	userVolumeMounts := make([]v1.VolumeMount, len(options.volumeMounts))
	copy(userVolumeMounts, options.volumeMounts)
	if kd.config.StorageHostPath != "" {
		options.volumes = append(options.volumes, v1.Volume{
			Name: storageVolumeName,
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: kd.config.StorageHostPath,
				},
			},
		})
		storageMount := v1.VolumeMount{
			Name:      storageVolumeName,
			MountPath: kd.config.StorageRoot,
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
			MountPath: kd.config.StorageRoot,
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
		certSecretVolume, certSecretMount := getTLSCertSecretVolumeAndMount(kd.config.TLSCertSecretName, path)
		options.volumes = append(options.volumes, certSecretVolume)
		sidecarVolumeMounts = append(sidecarVolumeMounts, certSecretMount)
	}

	// mount secret for spouts using pachctl
	if pipelineInfo.Details.Spout != nil {
		pachctlSecretVolume, pachctlSecretMount := getPachctlSecretVolumeAndMount("spout-pachctl-secret-" + pipelineInfo.Pipeline.Name)
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

	workerImage := kd.config.WorkerImage
	pachSecurityCtx := &v1.SecurityContext{
		RunAsUser:                int64Ptr(1000),
		RunAsGroup:               int64Ptr(1000),
		AllowPrivilegeEscalation: pointer.BoolPtr(false),
		ReadOnlyRootFilesystem:   pointer.BoolPtr(true),
		Capabilities:             &v1.Capabilities{Drop: []v1.Capability{"all"}},
	}
	var userSecurityCtx *v1.SecurityContext
	var podSecurityContext *v1.PodSecurityContext
	userStr := pipelineInfo.Details.Transform.User
	if !kd.config.EnableWorkerSecurityContexts {
		pachSecurityCtx = nil
		podSecurityContext = nil
	} else if kd.config.WorkerUsesRoot {
		pachSecurityCtx = &v1.SecurityContext{RunAsUser: int64Ptr(0)}
		userSecurityCtx = &v1.SecurityContext{RunAsUser: int64Ptr(0)}
		podSecurityContext = nil
	} else if userStr != "" {
		// This is to allow the user to be set in the pipeline spec.
		if i, err := strconv.ParseInt(userStr, 10, 64); err != nil {
			kd.logger.Warnf("could not parse user %q into int: %v", userStr, err)
		} else {
			// hard coded security settings besides uid/gid.
			podSecurityContext = &v1.PodSecurityContext{
				RunAsUser:    int64Ptr(i),
				RunAsGroup:   int64Ptr(i),
				FSGroup:      int64Ptr(i),
				RunAsNonRoot: pointer.BoolPtr(true),
				SeccompProfile: &v1.SeccompProfile{
					Type: v1.SeccompProfileType("RuntimeDefault"),
				}}
			userSecurityCtx = &v1.SecurityContext{
				RunAsUser:                int64Ptr(i),
				RunAsGroup:               int64Ptr(i),
				AllowPrivilegeEscalation: pointer.BoolPtr(false),
				ReadOnlyRootFilesystem:   pointer.BoolPtr(true),
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
				Command:         []string{"/pach-bin/worker"},
				ImagePullPolicy: v1.PullPolicy(pullPolicy),
				Env:             workerEnv,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    cpuZeroQuantity,
						v1.ResourceMemory: memDefaultQuantity,
					},
				},
				VolumeMounts:    userVolumeMounts,
				SecurityContext: userSecurityCtx,
			},
			{
				Name:            client.PPSWorkerSidecarContainerName,
				Image:           kd.config.WorkerSidecarImage,
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
		AutomountServiceAccountToken: pointer.BoolPtr(true),
		RestartPolicy:                 "Always",
		Volumes:                       options.volumes,
		ImagePullSecrets:              options.imagePullSecrets,
		TerminationGracePeriodSeconds: int64Ptr(0),
		SecurityContext:               podSecurityContext,
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

func (kd *kubeDriver) getStorageEnvVars(pipelineInfo *pps.PipelineInfo) []v1.EnvVar {
	vars := []v1.EnvVar{
		{Name: UploadConcurrencyLimitEnvVar, Value: strconv.Itoa(kd.config.StorageUploadConcurrencyLimit)},
		{Name: client.PPSPipelineNameEnv, Value: pipelineInfo.Pipeline.Name},
	}
	return vars
}

func (kd *kubeDriver) getEgressSecretEnvVars(pipelineInfo *pps.PipelineInfo) []v1.EnvVar {
	result := []v1.EnvVar{}
	egress := pipelineInfo.Details.Egress
	if egress != nil && egress.GetSqlDatabase() != nil && egress.GetSqlDatabase().GetSecret() != nil {
		secret := egress.GetSqlDatabase().GetSecret()
		result = append(result, v1.EnvVar{
			Name: "PACHYDERM_SQL_PASSWORD", // TODO avoid hardcoding this
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: secret.Name},
					Key:                  secret.Key,
				},
			},
		})
	}
	return result
}

// We don't want to expose pipeline auth tokens, so we hash it. This will be
// visible to any user with k8s cluster access
// Note: This hash shouldn't be used for authentication in any way. We just use
// this to detect if an auth token has been added/changed
// Note: ptr.AuthToken is a pachyderm-generated UUID, and wouldn't appear in any
// rainbow tables
func hashAuthToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func (kd *kubeDriver) getWorkerOptions(ctx context.Context, pipelineInfo *pps.PipelineInfo) (*workerOptions, error) {
	pipelineName := pipelineInfo.Pipeline.Name
	pipelineVersion := pipelineInfo.Version
	var resourceRequests *v1.ResourceList
	var resourceLimits *v1.ResourceList
	var sidecarResourceLimits *v1.ResourceList
	if pipelineInfo.Details.ResourceRequests != nil {
		var err error
		resourceRequests, err = ppsutil.GetRequestsResourceListFromPipeline(pipelineInfo)
		if err != nil {
			return nil, errors.Wrapf(err, "could not determine resource request")
		}
	}
	if pipelineInfo.Details.ResourceLimits != nil {
		var err error
		resourceLimits, err = ppsutil.GetLimitsResourceList(pipelineInfo.Details.ResourceLimits)
		if err != nil {
			return nil, errors.Wrapf(err, "could not determine resource limit")
		}
	}
	if pipelineInfo.Details.SidecarResourceLimits != nil {
		var err error
		sidecarResourceLimits, err = ppsutil.GetLimitsResourceList(pipelineInfo.Details.SidecarResourceLimits)
		if err != nil {
			return nil, errors.Wrapf(err, "could not determine sidecar resource limit")
		}
	}

	transform := pipelineInfo.Details.Transform
	labels := pipelineLabels(pipelineName, pipelineVersion)
	labels[pipelineNameLabel] = pipelineName
	userImage := transform.Image
	if userImage == "" {
		userImage = DefaultUserImage
	}

	workerEnv := []v1.EnvVar{{
		Name:  client.PPSPipelineNameEnv,
		Value: pipelineInfo.Pipeline.Name,
	}}
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

	volumes = append(volumes, v1.Volume{
		Name: client.PPSWorkerVolume,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})
	volumeMounts = append(volumeMounts, v1.VolumeMount{
		Name:      client.PPSWorkerVolume,
		MountPath: client.PPSInputPrefix,
	})
	var imagePullSecrets []v1.LocalObjectReference
	for _, secret := range transform.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, v1.LocalObjectReference{Name: secret})
	}
	if kd.config.ImagePullSecrets != "" {
		secrets := strings.Split(kd.config.ImagePullSecrets, ",")
		for _, secret := range secrets {
			imagePullSecrets = append(imagePullSecrets, v1.LocalObjectReference{Name: secret})
		}
	}

	annotations := map[string]string{
		pipelineNameLabel:            pipelineName,
		pachVersionAnnotation:        version.PrettyVersion(),
		pipelineVersionAnnotation:    strconv.FormatUint(pipelineInfo.Version, 10),
		pipelineSpecCommitAnnotation: pipelineInfo.SpecCommit.ID,
		hashedAuthTokenAnnotation:    hashAuthToken(pipelineInfo.AuthToken),
	}

	// add the user's custom metadata (annotations and labels).
	metadata := pipelineInfo.Details.GetMetadata()
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
	if pipelineInfo.Details.Spout != nil && pipelineInfo.Details.Service != nil {
		return nil, errors.New("only one of pipeline.service or pipeline.spout can be set")
	} else if pipelineInfo.Details.Spout != nil && pipelineInfo.Details.Spout.Service != nil {
		service = pipelineInfo.Details.Spout.Service
	} else {
		service = pipelineInfo.Details.Service
	}
	var s3GatewayPort int32
	if ppsutil.ContainsS3Inputs(pipelineInfo.Details.Input) || pipelineInfo.Details.S3Out {
		s3GatewayPort = int32(kd.config.S3GatewayPort)
	}

	// Get the reference to the postgres secret used by the current pod
	podName := kd.config.PachdPodName
	selfPodInfo, err := kd.kubeClient.CoreV1().Pods(kd.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	var postgresSecretRef *v1.SecretKeySelector
	for _, container := range selfPodInfo.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "POSTGRES_PASSWORD" && envVar.ValueFrom != nil && envVar.ValueFrom.SecretKeyRef != nil {
				postgresSecretRef = envVar.ValueFrom.SecretKeyRef
			}
		}
	}
	if postgresSecretRef == nil {
		return nil, errors.New("could not load the existing postgres secret reference from kubernetes")
	}

	// Generate options for new RC
	return &workerOptions{
		rcName:                ppsutil.PipelineRcName(pipelineName, pipelineVersion),
		s3GatewayPort:         s3GatewayPort,
		specCommit:            pipelineInfo.SpecCommit.ID,
		labels:                labels,
		annotations:           annotations,
		parallelism:           int32(0), // pipelines start w/ 0 workers & are scaled up
		resourceRequests:      resourceRequests,
		resourceLimits:        resourceLimits,
		sidecarResourceLimits: sidecarResourceLimits,
		userImage:             userImage,
		workerEnv:             workerEnv,
		volumes:               volumes,
		volumeMounts:          volumeMounts,
		postgresSecret:        postgresSecretRef,
		imagePullSecrets:      imagePullSecrets,
		service:               service,
		schedulingSpec:        pipelineInfo.Details.SchedulingSpec,
		podSpec:               pipelineInfo.Details.PodSpec,
		podPatch:              pipelineInfo.Details.PodPatch,
	}, nil
}

func (kd *kubeDriver) createWorkerPachctlSecret(ctx context.Context, pipelineInfo *pps.PipelineInfo) error {
	var cfg config.Config
	err := cfg.InitV2()
	if err != nil {
		return errors.Wrapf(err, "error initializing V2 for config")
	}
	_, context, err := cfg.ActiveContext(true)
	if err != nil {
		return errors.Wrapf(err, "error getting the active context")
	}
	context.SessionToken = pipelineInfo.AuthToken
	context.PachdAddress = "localhost:1653"

	rawConfig, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "error marshaling the config")
	}
	s := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "spout-pachctl-secret-" + pipelineInfo.Pipeline.Name,
			Labels: spoutLabels(pipelineInfo.Pipeline.Name),
		},
		Data: map[string][]byte{
			"config.json": rawConfig,
		},
	}
	labels := s.GetLabels()
	labels["pipelineName"] = pipelineInfo.Pipeline.Name
	s.SetLabels(labels)

	// send RPC to k8s to create the secret there
	if _, err := kd.kubeClient.CoreV1().Secrets(kd.namespace).Create(ctx, &s, metav1.CreateOptions{}); err != nil {
		if !errutil.IsAlreadyExistError(err) {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

// noValidOptions error may be returned by createWorkerSvcAndRc to indicate that
// getWorkerOptions returned an error to it (getWorkerOptions does not return
// noValidOptions). This is a mechanism for createWorkerSvcAndRc to signal to
// its caller not to retry
type noValidOptionsErr struct {
	error
}

func (kd *kubeDriver) createWorkerSvcAndRc(ctx context.Context, pipelineInfo *pps.PipelineInfo) (retErr error) {
	log.Infof("PPS master: upserting workers for %q", pipelineInfo.Pipeline.Name)
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/pps.Master/CreateWorkerRC", // ctx never used, but we want the right one in scope for future uses
		"pipeline", pipelineInfo.Pipeline.Name)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	// create pachctl secret used in spouts
	if pipelineInfo.Details.Spout != nil {
		if err := kd.createWorkerPachctlSecret(ctx, pipelineInfo); err != nil {
			return err
		}
	}

	options, err := kd.getWorkerOptions(ctx, pipelineInfo)
	if err != nil {
		return noValidOptionsErr{err}
	}
	podSpec, err := kd.workerPodSpec(options, pipelineInfo)
	if err != nil {
		return err
	}
	rc := &v1.ReplicationController{
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
	if _, err := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace).Create(ctx, rc, metav1.CreateOptions{}); err != nil {
		if !errutil.IsAlreadyExistError(err) {
			return errors.EnsureStack(err)
		}
	}
	serviceAnnotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   strconv.Itoa(workerstats.PrometheusPort),
	}

	service := &v1.Service{
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
					Port: int32(kd.config.PPSWorkerPort),
					Name: "grpc-port",
				},
				{
					Port: workerstats.PrometheusPort,
					Name: "prom-metrics",
				},
			},
		},
	}
	if _, err := kd.kubeClient.CoreV1().Services(kd.namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
		if !errutil.IsAlreadyExistError(err) {
			return errors.EnsureStack(err)
		}
	}

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
		service := &v1.Service{
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
		}
		if _, err := kd.kubeClient.CoreV1().Services(kd.namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
			if !errutil.IsAlreadyExistError(err) {
				return errors.EnsureStack(err)
			}
		}
	}

	return nil
}

// GetBackendSecretVolumeAndMount returns a properly configured Volume and
// VolumeMount object
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

func int64Ptr(x int64) *int64 {
	return &x
}
