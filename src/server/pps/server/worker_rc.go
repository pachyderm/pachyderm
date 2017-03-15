package server

import (
	"fmt"
	"math"
	"strings"

	client "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

// Parameters used when creating the kubernetes replication controller in charge
// of a job or pipeline's workers
type workerOptions struct {
	rcName string // Name of the replication controller managing workers

	userImage    string            // The user's pipeline/job image
	labels       map[string]string // k8s labels attached to the RC and workers
	parallelism  int32             // Number of replicas the RC maintains
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
	// TODO: deal with name collision
	return fmt.Sprintf("pipeline-%s-v%d", strings.ToLower(name), version)
}

// JobRcName generates the name of the k8s replication controller that manages
// an orphan job's workers
func JobRcName(id string) string {
	// k8s won't allow RC names that contain upper-case letters
	// TODO: deal with name collision
	return fmt.Sprintf("job-%s", strings.ToLower(id))
}

func (a *apiServer) workerPodSpec(options *workerOptions) api.PodSpec {
	pullPolicy := a.workerImagePullPolicy
	if pullPolicy == "" {
		pullPolicy = "IfNotPresent"
	}
	return api.PodSpec{
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
				Name:    "user",
				Image:   options.userImage,
				Command: []string{"/pach-bin/guest.sh"},
				SecurityContext: &api.SecurityContext{
					Privileged: &trueVal, // god is this dumb
				},
				ImagePullPolicy: api.PullPolicy(pullPolicy),
				Env:             options.workerEnv,
				VolumeMounts:    options.volumeMounts,
			},
		},
		RestartPolicy:    "Always",
		Volumes:          options.volumes,
		ImagePullSecrets: options.imagePullSecrets,
	}
}

// GetExpectedNumWorkers computes the expected number of workers that pachyderm will start given
// the ParallelismSpec 'spec'.
//
// This is only exported for testing
func GetExpectedNumWorkers(kubeClient *kube.Client, spec *pps.ParallelismSpec) (uint64, error) {
	fmt.Println("Getting expected num workers...")
	coefficient := 0.0 // Used if [spec.Strategy == PROPORTIONAL] or [spec.Constant == 0]
	if spec == nil {
		// Unset ParallelismSpec is handled here. Currently we start one worker per
		// node
		coefficient = 1.0
	} else if spec.Strategy == pps.ParallelismSpec_CONSTANT {
		if spec.Constant > 0 {
			fmt.Println("Returning constant: ", spec.Constant)
			return spec.Constant, nil
		}
		// Zero-initialized ParallelismSpec is handled here. Currently we start one
		// worker per node
		coefficient = 1
	} else if spec.Strategy == pps.ParallelismSpec_COEFFICIENT {
		coefficient = spec.Coefficient
	} else {
		return 0, fmt.Errorf("Unable to interpret ParallelismSpec strategy %s", spec.Strategy)
	}
	if coefficient == 0.0 {
		return 0, fmt.Errorf("Ended up with coefficient == 0 (no workers) after interpreting ParallelismSpec %s", spec.Strategy)
	}

	// Start ('coefficient' * 'nodes') workers. Determine number of workers
	nodeList, err := kubeClient.Nodes().List(api.ListOptions{})
	fmt.Println("nodeList: %+v", nodeList)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve node list from k8s to determine parallelism: %v", err)
	}
	if len(nodeList.Items) == 0 {
		return 0, fmt.Errorf("pachyderm.pps.jobserver: no k8s nodes found")
	}
	result := math.Floor(coefficient * float64(len(nodeList.Items)))
	return uint64(math.Max(result, 1)), nil
}

func (a *apiServer) getWorkerOptions(rcName string, parallelism int32, transform *pps.Transform) *workerOptions {
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
	var imagePullSecrets []api.LocalObjectReference
	for _, secret := range transform.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, api.LocalObjectReference{Name: secret})
	}

	return &workerOptions{
		rcName:           rcName,
		labels:           labels,
		parallelism:      int32(parallelism),
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
	_, err := a.kubeClient.ReplicationControllers(a.namespace).Create(rc)
	fmt.Println("Error from a.kubeClient.ReplicationControllers(a.namespace).Create(rc): ", err)
	return err
}
