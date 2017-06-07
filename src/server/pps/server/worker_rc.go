package server

import (
	"fmt"
	"strings"

	client "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// Parameters used when creating the kubernetes replication controller in charge
// of a job or pipeline's workers
type workerOptions struct {
	rcName string // Name of the replication controller managing workers

	userImage    string            // The user's pipeline/job image
	labels       map[string]string // k8s labels attached to the Deployment and workers
	parallelism  int32             // Number of replicas the RC maintains
	resources    *api.ResourceList // Resources requested by pipeline/job pods
	workerEnv    []api.EnvVar      // Environment vars set in the user container
	volumes      []api.Volume      // Volumes that we expose to the user container
	volumeMounts []api.VolumeMount // Paths where we mount each volume in 'volumes'

	// Secrets that we mount in the worker container (e.g. for reading/writing to
	// s3)
	imagePullSecrets []api.LocalObjectReference
}

// PipelineRcName generates the name of the k8s replication controller that
// manages a pipeline's workers
func PipelineRcName(name string, version uint64) string {
	// k8s won't allow RC names that contain upper-case letters
	// or underscores
	// TODO: deal with name collision
	name = strings.Replace(name, "_", "-", -1)
	return fmt.Sprintf("pipeline-%s-v%d", strings.ToLower(name), version)
}

func (a *apiServer) workerPodSpec(options *workerOptions) api.PodSpec {
	pullPolicy := a.workerImagePullPolicy
	if pullPolicy == "" {
		pullPolicy = "IfNotPresent"
	}
	// Disable block caching for the sidecar.  We still want PFS cache
	// because it caches things like commit IDs which typically has a
	// very high hit rate since workers tend to be processing a small
	// set commits at a time.
	sidecarEnv := []api.EnvVar{{
		Name:  "BLOCK_CACHE_BYTES",
		Value: "256M",
	}, {
		Name:  "PFS_CACHE_BYTES",
		Value: "10M",
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
	var sidecarVolumeMounts []api.VolumeMount
	if a.storageHostPath != "" {
		options.volumes = append(options.volumes, api.Volume{
			Name: storageVolumeName,
			VolumeSource: api.VolumeSource{
				HostPath: &api.HostPathVolumeSource{
					Path: a.storageHostPath,
				},
			},
		})

		sidecarVolumeMounts = []api.VolumeMount{
			{
				Name:      storageVolumeName,
				MountPath: a.storageRoot,
			},
		}
	}
	secretVolume, secretMount, err := assets.GetSecretVolumeAndMount(a.storageBackend)
	if err == nil {
		options.volumes = append(options.volumes, secretVolume)
		sidecarVolumeMounts = append(sidecarVolumeMounts, secretMount)
	}
	podSpec := api.PodSpec{
		InitContainers: []api.Container{
			{
				Name:            "init",
				Image:           a.workerImage,
				Command:         []string{"/pach/worker.sh"},
				ImagePullPolicy: api.PullPolicy(pullPolicy),
				Env:             options.workerEnv,
				VolumeMounts:    options.volumeMounts,
			},
		},
		Containers: []api.Container{
			{
				Name:    client.PPSWorkerUserContainerName,
				Image:   options.userImage,
				Command: []string{"/pach-bin/guest.sh"},
				SecurityContext: &api.SecurityContext{
					Privileged: &trueVal, // god is this dumb
				},
				ImagePullPolicy: api.PullPolicy(pullPolicy),
				Env:             options.workerEnv,
				VolumeMounts:    options.volumeMounts,
			},
			{
				Name:            client.PPSWorkerSidecarContainerName,
				Image:           a.workerSidecarImage,
				Command:         []string{"/pachd", "--mode", "pfs"},
				ImagePullPolicy: api.PullPolicy(pullPolicy),
				Env:             sidecarEnv,
				VolumeMounts:    sidecarVolumeMounts,
			},
		},
		RestartPolicy:    "Always",
		Volumes:          options.volumes,
		ImagePullSecrets: options.imagePullSecrets,
	}
	if options.resources != nil {
		podSpec.Containers[0].Resources = api.ResourceRequirements{
			Requests: *options.resources,
		}
	}
	return podSpec
}

func (a *apiServer) getWorkerOptions(rcName string, parallelism int32, resources *api.ResourceList, transform *pps.Transform) *workerOptions {
	labels := labels(rcName)
	userImage := transform.Image
	if userImage == "" {
		userImage = DefaultUserImage
	}

	var workerEnv []api.EnvVar
	for name, value := range transform.Env {
		workerEnv = append(
			workerEnv,
			api.EnvVar{
				Name:  name,
				Value: value,
			},
		)
	}
	// We use Kubernetes' "Downward API" so the workers know their IP
	// addresses, which they will then post on etcd so the job managers
	// can discover the workers.
	workerEnv = append(workerEnv, api.EnvVar{
		Name: client.PPSWorkerIPEnv,
		ValueFrom: &api.EnvVarSource{
			FieldRef: &api.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	})
	workerEnv = append(workerEnv, api.EnvVar{
		Name: client.PPSPodNameEnv,
		ValueFrom: &api.EnvVarSource{
			FieldRef: &api.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})
	// Set the etcd prefix env
	workerEnv = append(workerEnv, api.EnvVar{
		Name:  client.PPSEtcdPrefixEnv,
		Value: a.etcdPrefix,
	})

	var volumes []api.Volume
	var volumeMounts []api.VolumeMount
	for _, secret := range transform.Secrets {
		volumes = append(volumes, api.Volume{
			Name: secret.Name,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: secret.Name,
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      secret.Name,
			MountPath: secret.MountPath,
		})
	}

	volumes = append(volumes, api.Volume{
		Name: "pach-bin",
		VolumeSource: api.VolumeSource{
			EmptyDir: &api.EmptyDirVolumeSource{},
		},
	})
	volumeMounts = append(volumeMounts, api.VolumeMount{
		Name:      "pach-bin",
		MountPath: "/pach-bin",
	})

	volumes = append(volumes, api.Volume{
		Name: client.PPSWorkerVolume,
		VolumeSource: api.VolumeSource{
			EmptyDir: &api.EmptyDirVolumeSource{},
		},
	})
	volumeMounts = append(volumeMounts, api.VolumeMount{
		Name:      client.PPSWorkerVolume,
		MountPath: client.PPSInputPrefix,
	})
	if resources != nil && resources.NvidiaGPU() != nil && !resources.NvidiaGPU().IsZero() {
		volumes = append(volumes, api.Volume{
			Name: "root-lib",
			VolumeSource: api.VolumeSource{
				HostPath: &api.HostPathVolumeSource{
					Path: "/usr/lib",
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      "root-lib",
			MountPath: "/rootfs/usr/lib",
		})
	}
	var imagePullSecrets []api.LocalObjectReference
	for _, secret := range transform.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, api.LocalObjectReference{Name: secret})
	}

	return &workerOptions{
		rcName:           rcName,
		labels:           labels,
		parallelism:      int32(parallelism),
		resources:        resources,
		userImage:        userImage,
		workerEnv:        workerEnv,
		volumes:          volumes,
		volumeMounts:     volumeMounts,
		imagePullSecrets: imagePullSecrets,
	}
}

func (a *apiServer) createWorkerRc(options *workerOptions) error {
	rc := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   options.rcName,
			Labels: options.labels,
		},
		Spec: api.ReplicationControllerSpec{
			Selector: options.labels,
			Replicas: options.parallelism,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   options.rcName,
					Labels: options.labels,
				},
				Spec: a.workerPodSpec(options),
			},
		},
	}
	if _, err := a.kubeClient.ReplicationControllers(a.namespace).Create(rc); err != nil {
		if !isAlreadyExistsErr(err) {
			return err
		}
	}

	service := &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   options.rcName,
			Labels: options.labels,
		},
		Spec: api.ServiceSpec{
			Selector: options.labels,
			Ports: []api.ServicePort{
				{
					Port: client.PPSWorkerPort,
					Name: "grpc-port",
				},
			},
		},
	}

	if _, err := a.kubeClient.Services(a.namespace).Create(service); err != nil {
		if !isAlreadyExistsErr(err) {
			return err
		}
	}

	return nil
}
