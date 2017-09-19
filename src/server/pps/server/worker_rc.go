package server

import (
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
	labels       map[string]string // k8s labels attached to the RC and workers
	annotations  map[string]string // k8s annotations attached to the RC and workers
	parallelism  int32             // Number of replicas the RC maintains
	cacheSize    string            // Size of cache that sidecar uses
	resources    *api.ResourceList // Resources requested by pipeline/job pods
	workerEnv    []api.EnvVar      // Environment vars set in the user container
	volumes      []api.Volume      // Volumes that we expose to the user container
	volumeMounts []api.VolumeMount // Paths where we mount each volume in 'volumes'

	// Secrets that we mount in the worker container (e.g. for reading/writing to
	// s3)
	imagePullSecrets []api.LocalObjectReference
}

func (a *apiServer) workerPodSpec(options *workerOptions) api.PodSpec {
	pullPolicy := a.workerImagePullPolicy
	if pullPolicy == "" {
		pullPolicy = "IfNotPresent"
	}
	// TODO: make the cache sizes configurable
	sidecarEnv := []api.EnvVar{{
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
	userVolumeMounts := options.volumeMounts
	secretVolume, secretMount, err := assets.GetSecretVolumeAndMount(a.storageBackend)
	if err == nil {
		options.volumes = append(options.volumes, secretVolume)
		options.volumeMounts = append(options.volumeMounts, secretMount)
		sidecarVolumeMounts = append(sidecarVolumeMounts, secretMount)
		userVolumeMounts = append(userVolumeMounts, secretMount)
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
				VolumeMounts:    userVolumeMounts,
			},
			{
				Name:            client.PPSWorkerSidecarContainerName,
				Image:           a.workerSidecarImage,
				Command:         []string{"/pachd", "--mode", "sidecar"},
				ImagePullPolicy: api.PullPolicy(pullPolicy),
				Env:             sidecarEnv,
				VolumeMounts:    sidecarVolumeMounts,
			},
		},
		RestartPolicy:                 "Always",
		Volumes:                       options.volumes,
		ImagePullSecrets:              options.imagePullSecrets,
		TerminationGracePeriodSeconds: &zeroVal,
	}
	if options.resources != nil {
		podSpec.Containers[0].Resources = api.ResourceRequirements{
			Requests: *options.resources,
		}
	}
	return podSpec
}

func (a *apiServer) getWorkerOptions(pipelineName string, rcName string, parallelism int32, resources *api.ResourceList, transform *pps.Transform, cacheSize string) *workerOptions {
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
	// Pass along the namespace
	workerEnv = append(workerEnv, api.EnvVar{
		Name:  client.PPSNamespaceEnv,
		Value: a.namespace,
	})

	var volumes []api.Volume
	var volumeMounts []api.VolumeMount
	for _, secret := range transform.Secrets {
		if secret.MountPath != "" {
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
		for ev, key := range secret.EnvVarToKey {
			workerEnv = append(workerEnv, api.EnvVar{
				Name: ev,
				ValueFrom: &api.EnvVarSource{
					SecretKeyRef: &api.SecretKeySelector{
						LocalObjectReference: api.LocalObjectReference{
							Name: secret.Name,
						},
						Key: key,
					},
				},
			})
		}
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
		MountPath: client.PPSScratchSpace,
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
		annotations:      map[string]string{"pipelineName": pipelineName},
		parallelism:      int32(parallelism),
		resources:        resources,
		userImage:        userImage,
		workerEnv:        workerEnv,
		volumes:          volumes,
		volumeMounts:     volumeMounts,
		imagePullSecrets: imagePullSecrets,
		cacheSize:        cacheSize,
	}
}

func (a *apiServer) createWorkerRc(options *workerOptions) error {
	rc := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:        options.rcName,
			Labels:      options.labels,
			Annotations: options.annotations,
		},
		Spec: api.ReplicationControllerSpec{
			Selector: options.labels,
			Replicas: options.parallelism,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:        options.rcName,
					Labels:      options.labels,
					Annotations: options.annotations,
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
