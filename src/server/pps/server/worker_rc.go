package server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	client "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/worker"

	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kube "k8s.io/client-go/kubernetes"
)

const (
	pipelineNameLabel    = "pipelineName"
	pachVersionLabel     = "version"
	specCommitLabel      = "specCommit"
	hashedAuthTokenLabel = "authTokenHash"
)

// Parameters used when creating the kubernetes replication controller in charge
// of a job or pipeline's workers
type workerOptions struct {
	rcName string // Name of the replication controller managing workers

	userImage        string              // The user's pipeline/job image
	labels           map[string]string   // k8s labels attached to the RC and workers
	annotations      map[string]string   // k8s annotations attached to the RC and workers
	parallelism      int32               // Number of replicas the RC maintains
	cacheSize        string              // Size of cache that sidecar uses
	resourceRequests *v1.ResourceList    // Resources requested by pipeline/job pods
	resourceLimits   *v1.ResourceList    // Resources requested by pipeline/job pods
	workerEnv        []v1.EnvVar         // Environment vars set in the user container
	volumes          []v1.Volume         // Volumes that we expose to the user container
	volumeMounts     []v1.VolumeMount    // Paths where we mount each volume in 'volumes'
	etcdPrefix       string              // the prefix in etcd to use
	schedulingSpec   *pps.SchedulingSpec // the SchedulingSpec for the pipeline
	podSpec          string

	// Secrets that we mount in the worker container (e.g. for reading/writing to
	// s3)
	imagePullSecrets []v1.LocalObjectReference
	service          *pps.Service
}

func (a *apiServer) workerPodSpec(options *workerOptions) (v1.PodSpec, error) {
	pullPolicy := a.workerImagePullPolicy
	if pullPolicy == "" {
		pullPolicy = "IfNotPresent"
	}
	sidecarEnv := []v1.EnvVar{{
		Name:  "BLOCK_CACHE_BYTES",
		Value: options.cacheSize,
	}, {
		Name:  "PFS_CACHE_SIZE",
		Value: "16",
	}, {
		Name:  "PACH_ROOT",
		Value: a.storageRoot,
	}, {
		Name:  "STORAGE_BACKEND",
		Value: a.storageBackend,
	}}
	// This only happens in local deployment.  We want the workers to be
	// able to read from/write to the hostpath volume as well.
	storageVolumeName := "pach-disk"
	var sidecarVolumeMounts []v1.VolumeMount
	if a.storageHostPath != "" {
		options.volumes = append(options.volumes, v1.Volume{
			Name: storageVolumeName,
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: a.storageHostPath,
				},
			},
		})

		sidecarVolumeMounts = []v1.VolumeMount{
			{
				Name:      storageVolumeName,
				MountPath: a.storageRoot,
			},
		}
	}
	userVolumeMounts := options.volumeMounts
	secretVolume, secretMount := assets.GetBackendSecretVolumeAndMount(a.storageBackend)
	options.volumes = append(options.volumes, secretVolume)
	options.volumeMounts = append(options.volumeMounts, secretMount)
	sidecarVolumeMounts = append(sidecarVolumeMounts, secretMount)
	userVolumeMounts = append(userVolumeMounts, secretMount)

	// Explicitly set CPU, MEM and DISK requests to zero because some cloud
	// providers set their own defaults which are usually not what we want.
	cpuZeroQuantity := resource.MustParse("0")
	memZeroQuantity := resource.MustParse("0M")
	memSidecarQuantity := resource.MustParse(options.cacheSize)

	if !a.noExposeDockerSocket {
		options.volumes = append(options.volumes, v1.Volume{
			Name: "docker",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/var/run/docker.sock",
				},
			},
		})
		userVolumeMounts = append(userVolumeMounts, v1.VolumeMount{
			Name:      "docker",
			MountPath: "/var/run/docker.sock",
		})
	}
	zeroVal := int64(0)
	workerImage := a.workerImage
	resp, err := a.getPachClient().Enterprise.GetState(context.Background(), &enterprise.GetStateRequest{})
	if err != nil {
		return v1.PodSpec{}, err
	}
	if resp.State != enterprise.State_ACTIVE {
		workerImage = assets.AddRegistry("", workerImage)
	}
	podSpec := v1.PodSpec{
		InitContainers: []v1.Container{
			{
				Name:            "init",
				Image:           workerImage,
				Command:         []string{"/pach/worker.sh"},
				ImagePullPolicy: v1.PullPolicy(pullPolicy),
				VolumeMounts:    options.volumeMounts,
			},
		},
		Containers: []v1.Container{
			{
				Name:            client.PPSWorkerUserContainerName,
				Image:           options.userImage,
				Command:         []string{"/pach-bin/worker"},
				ImagePullPolicy: v1.PullPolicy(pullPolicy),
				Env:             options.workerEnv,
				VolumeMounts:    userVolumeMounts,
			},
			{
				Name:            client.PPSWorkerSidecarContainerName,
				Image:           a.workerSidecarImage,
				Command:         []string{"/pachd", "--mode", "sidecar"},
				ImagePullPolicy: v1.PullPolicy(pullPolicy),
				Env:             sidecarEnv,
				VolumeMounts:    sidecarVolumeMounts,
				Resources: v1.ResourceRequirements{
					Requests: map[v1.ResourceName]resource.Quantity{
						v1.ResourceCPU:    cpuZeroQuantity,
						v1.ResourceMemory: memSidecarQuantity,
					},
				},
			},
		},
		RestartPolicy:                 "Always",
		Volumes:                       options.volumes,
		ImagePullSecrets:              options.imagePullSecrets,
		TerminationGracePeriodSeconds: &zeroVal,
		SecurityContext:               &v1.PodSecurityContext{RunAsUser: &zeroVal},
		ServiceAccountName:            assets.ServiceAccountName,
	}
	if options.schedulingSpec != nil {
		podSpec.NodeSelector = options.schedulingSpec.NodeSelector
		podSpec.PriorityClassName = options.schedulingSpec.PriorityClassName
	}
	resourceRequirements := v1.ResourceRequirements{
		Requests: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    cpuZeroQuantity,
			v1.ResourceMemory: memZeroQuantity,
		},
	}
	if options.resourceRequests != nil {
		resourceRequirements.Requests = *options.resourceRequests
	}
	if options.resourceLimits != nil {
		resourceRequirements.Limits = *options.resourceLimits
	}
	podSpec.Containers[0].Resources = resourceRequirements
	if options.podSpec != "" {
		if err := json.Unmarshal([]byte(options.podSpec), &podSpec); err != nil {
			return v1.PodSpec{}, err
		}
	}
	return podSpec, nil
}

// we don't want to expose pipeline auth tokens, so we hash them before
// returning them to use as a label
func hashAuthToken(token string) string {
	hashBytes := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hashBytes[:])
}

// commitIDFromB64 converts 'id' from a base64-string (in the RC label) to a hex
// string that can be compared with its stored format
func commitIDFromB64(label string) (string, error) {
	bts, err := base64.RawURLEncoding.DecodeString(label)
	if err != nil {
		return "", fmt.Errorf("couldn't decode commit ID %q as hex: %v", label, err)
	}
	return hex.EncodeToString(bts), nil
}

// commitIDToB64 converts 'id' from a hex string to a base64-string (to shorten
// it, so that it fits in the RC label)
func commitIDToB64(id string) (string, error) {
	bts, err := hex.DecodeString(id)
	if err != nil {
		return "", fmt.Errorf("couldn't decode commit ID %q as hex: %v", id, err)
	}
	return base64.RawURLEncoding.EncodeToString(bts), nil
}

func (a *apiServer) getWorkerOptions(ptr *pps.EtcdPipelineInfo, pipelineInfo *pps.PipelineInfo) (*workerOptions, error) {
	pipelineName := pipelineInfo.Pipeline.Name
	pipelineVersion := pipelineInfo.Version
	parallelism := 0
	var resourceRequests *v1.ResourceList
	var resourceLimits *v1.ResourceList
	if pipelineInfo.ResourceRequests != nil {
		var err error
		resourceRequests, err = ppsutil.GetRequestsResourceListFromPipeline(pipelineInfo)
		if err != nil {
			return nil, err
		}
	}
	if pipelineInfo.ResourceLimits != nil {
		var err error
		resourceLimits, err = ppsutil.GetLimitsResourceListFromPipeline(pipelineInfo)
		if err != nil {
			return nil, err
		}
	}

	transform := pipelineInfo.Transform
	cacheSize := pipelineInfo.CacheSize
	service := pipelineInfo.Service
	schedulingSpec := pipelineInfo.SchedulingSpec
	podSpec := pipelineInfo.PodSpec
	specCommitB64, err := commitIDToB64(ptr.SpecCommit.ID)
	if err != nil {
		return nil, err
	}

	rcName := ppsutil.PipelineRcName(pipelineName, pipelineVersion)
	labels := labels(rcName)
	labels[pipelineNameLabel] = pipelineName
	labels[pachVersionLabel] = version.PrettyVersion()
	labels[specCommitLabel] = specCommitB64
	labels[hashedAuthTokenLabel] = hashAuthToken(ptr.AuthToken)
	userImage := transform.Image
	if userImage == "" {
		userImage = DefaultUserImage
	}

	var workerEnv []v1.EnvVar
	for name, value := range transform.Env {
		workerEnv = append(
			workerEnv,
			v1.EnvVar{
				Name:  name,
				Value: value,
			},
		)
	}
	// We use Kubernetes' "Downward API" so the workers know their IP
	// addresses, which they will then post on etcd so the job managers
	// can discover the workers.
	workerEnv = append(workerEnv, v1.EnvVar{
		Name: client.PPSWorkerIPEnv,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	})
	workerEnv = append(workerEnv, v1.EnvVar{
		Name: client.PPSPodNameEnv,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})
	// Set the etcd prefix env
	workerEnv = append(workerEnv, v1.EnvVar{
		Name:  client.PPSEtcdPrefixEnv,
		Value: a.etcdPrefix,
	})
	// Pass along the namespace
	workerEnv = append(workerEnv, v1.EnvVar{
		Name:  client.PPSNamespaceEnv,
		Value: a.namespace,
	})
	workerEnv = append(workerEnv, v1.EnvVar{
		Name:  client.PPSSpecCommitEnv,
		Value: ptr.SpecCommit.ID,
	})
	workerEnv = append(workerEnv, v1.EnvVar{
		Name:  client.PPSPipelineNameEnv,
		Value: pipelineInfo.Pipeline.Name,
	})

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
	if resourceLimits != nil && resourceLimits.NvidiaGPU() != nil && !resourceLimits.NvidiaGPU().IsZero() {
		volumes = append(volumes, v1.Volume{
			Name: "root-lib",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/usr/lib",
				},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      "root-lib",
			MountPath: "/rootfs/usr/lib",
		})
	}
	var imagePullSecrets []v1.LocalObjectReference
	for _, secret := range transform.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, v1.LocalObjectReference{Name: secret})
	}
	if a.imagePullSecret != "" {
		imagePullSecrets = append(imagePullSecrets, v1.LocalObjectReference{Name: a.imagePullSecret})
	}

	annotations := map[string]string{"pipelineName": pipelineName}
	if a.iamRole != "" {
		annotations["iam.amazonaws.com/role"] = a.iamRole
	}

	return &workerOptions{
		rcName:           rcName,
		labels:           labels,
		annotations:      annotations,
		parallelism:      int32(parallelism),
		resourceRequests: resourceRequests,
		resourceLimits:   resourceLimits,
		userImage:        userImage,
		workerEnv:        workerEnv,
		volumes:          volumes,
		volumeMounts:     volumeMounts,
		imagePullSecrets: imagePullSecrets,
		cacheSize:        cacheSize,
		service:          service,
		schedulingSpec:   schedulingSpec,
		podSpec:          podSpec,
	}, nil
}

func (a *apiServer) createWorkerSvcAndRc(ctx context.Context, ptr *pps.EtcdPipelineInfo, pipelineInfo *pps.PipelineInfo) (retErr error) {
	log.Infof("PPS master: upserting workers for %q", pipelineInfo.Pipeline.Name)
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/pps.Master/CreateWorkerRC", "pipeline", pipelineInfo.Pipeline.Name)
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)
	options, err := a.getWorkerOptions(ptr, pipelineInfo)

	if err != nil {
		return err
	}
	podSpec, err := a.workerPodSpec(options)
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
	if _, err := a.kubeClient.CoreV1().ReplicationControllers(a.namespace).Create(rc); err != nil {
		if !isAlreadyExistsErr(err) {
			return err
		}
	}
	serviceAnnotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   strconv.Itoa(worker.PrometheusPort),
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
			Selector: options.labels,
			Ports: []v1.ServicePort{
				{
					Port: client.PPSWorkerPort,
					Name: "grpc-port",
				},
				{
					Port: worker.PrometheusPort,
					Name: "prometheus-metrics",
				},
			},
		},
	}
	if _, err := a.kubeClient.CoreV1().Services(a.namespace).Create(service); err != nil {
		if !isAlreadyExistsErr(err) {
			return err
		}
	}

	if options.service != nil {
		service := &v1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   options.rcName + "-user",
				Labels: options.labels,
			},
			Spec: v1.ServiceSpec{
				Selector: options.labels,
				Type:     v1.ServiceTypeNodePort,
				Ports: []v1.ServicePort{
					{
						Port:       options.service.ExternalPort,
						TargetPort: intstr.FromInt(int(options.service.InternalPort)),
						Name:       "user-port",
						NodePort:   options.service.ExternalPort,
					},
				},
			},
		}
		if _, err := a.kubeClient.CoreV1().Services(a.namespace).Create(service); err != nil {
			if !isAlreadyExistsErr(err) {
				return err
			}
		}
	}

	// True if the pipeline has a git input
	var hasGitInput bool
	pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) {
		if input.Git != nil {
			hasGitInput = true
		}
	})
	if hasGitInput {
		if err := a.checkOrDeployGithookService(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) checkOrDeployGithookService(ctx context.Context) error {
	_, err := getGithookService(a.kubeClient, a.namespace)
	if err != nil {
		if _, ok := err.(*errGithookServiceNotFound); ok {
			svc := assets.GithookService(a.namespace)
			_, err = a.kubeClient.CoreV1().Services(a.namespace).Create(svc)
			return err
		}
		return err
	}
	// service already exists
	return nil
}

func getGithookService(kubeClient *kube.Clientset, namespace string) (*v1.Service, error) {
	labels := map[string]string{
		"app":   "githook",
		"suite": suite,
	}
	serviceList, err := kubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(labels)),
	})
	if err != nil {
		return nil, err
	}
	if len(serviceList.Items) != 1 {
		return nil, &errGithookServiceNotFound{
			fmt.Errorf("expected 1 githook service but found %v", len(serviceList.Items)),
		}
	}
	return &serviceList.Items[0], nil
}
