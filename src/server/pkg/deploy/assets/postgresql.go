package assets

import (
	"path/filepath"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	postgresImage                   = "postgres:11.3"
	postgresName                    = "postgres"
	postgresVolumeName              = "postgres-volume"
	postgresVolumeClaimName         = "postgres-storage"
	defaultPostgresStorageClassName = "postgres-storage-class"
)

// PostgresStorageClass creates a storage class used for dynamic volume
// provisioning.  Currently dynamic volume provisioning only works
// on AWS and GCE.
func PostgresStorageClass(opts *AssetOpts, backend backend) (interface{}, error) {
	return makeStorageClass(opts, backend, defaultPostgresStorageClassName, labels(postgresName))
}

// PostgresVolume creates a persistent volume backed by a volume with name "name"
func PostgresVolume(persistentDiskBackend backend, opts *AssetOpts,
	hostPath string, name string, size int) (*v1.PersistentVolume, error) {
	return makePersistentVolume(persistentDiskBackend, opts, hostPath, name, size, postgresVolumeName, labels(postgresName))
}

// PostgresVolumeClaim creates a persistent volume claim of 'size' GB.
//
// Note that if you're controlling Postgres with a Stateful Set, this is
// unnecessary (the stateful set controller will create PVCs automatically).
func PostgresVolumeClaim(size int, opts *AssetOpts) *v1.PersistentVolumeClaim {
	return makeVolumeClaim(size, opts, postgresVolumeName, postgresVolumeClaimName, labels(postgresName))
}

// PostgresDeployment generates a Deployment for the pachyderm postgres instance.
func PostgresDeployment(opts *AssetOpts, hostPath string) *apps.Deployment {
	cpu := resource.MustParse(opts.PostgresCPURequest)
	mem := resource.MustParse(opts.PostgresMemRequest)
	var volumes []v1.Volume
	if hostPath == "" {
		volumes = []v1.Volume{
			{
				Name: "postgres-storage",
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: postgresVolumeClaimName,
					},
				},
			},
		}
	} else {
		volumes = []v1.Volume{
			{
				Name: "postgres-storage",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: filepath.Join(hostPath, "postgres"),
					},
				},
			},
		}
	}
	resourceRequirements := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    cpu,
			v1.ResourceMemory: mem,
		},
	}
	if !opts.NoGuaranteed {
		resourceRequirements.Limits = v1.ResourceList{
			v1.ResourceCPU:    cpu,
			v1.ResourceMemory: mem,
		}
	}
	image := postgresImage
	if opts.Registry != "" {
		image = AddRegistry(opts.Registry, postgresImage)
	}
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: objectMeta(postgresName, labels(postgresName), nil, opts.Namespace),
		Spec: apps.DeploymentSpec{
			Replicas: replicas(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(postgresName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: objectMeta(postgresName, labels(postgresName), nil, opts.Namespace),
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  postgresName,
							Image: image,
							//TODO figure out how to get a cluster of these to talk to each other
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 5432,
									Name:          "client-port",
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "postgres-storage",
									MountPath: "/var/lib/postgresql/data",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							Resources:       resourceRequirements,
							Env: []v1.EnvVar{
								{Name: "POSTGRES_DB", Value: "pgc"},
								{Name: "POSTGRES_USER", Value: "pachyderm"},
								{Name: "POSTGRES_PASSWORD", Value: "elephantastic"},
							},
						},
					},
					Volumes:          volumes,
					ImagePullSecrets: imagePullSecrets(opts),
				},
			},
		},
	}
}

// PostgresService generates a Service for the pachyderm postgres instance.
func PostgresService(local bool, opts *AssetOpts) *v1.Service {
	var clientNodePort int32
	if local {
		clientNodePort = 32228
	}
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(postgresName, labels(postgresName), nil, opts.Namespace),
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": postgresName,
			},
			Ports: []v1.ServicePort{
				{
					Port:     5432,
					Name:     "client-port",
					NodePort: clientNodePort,
				},
			},
		},
	}
}
