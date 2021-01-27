package assets

import (
	"fmt"
	"path"

	"github.com/pachyderm/pachyderm/src/internal/dbutil"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Refactor the stateful set setup to better capture the shared functionality between the etcd / postgres setup.
// New / existing features that apply to both should be captured in one place.
// TODO: Move off of kubernetes Deployment object entirely since it is not well suited for stateful applications.
// The primary motivation for this would be to avoid the deadlock that can occur when using a ReadWriteOnce volume mount
// with a kubernetes Deployment.

var (
	postgresImage = "postgres:13.0-alpine"

	postgresHeadlessServiceName     = "postgres-headless"
	postgresName                    = "postgres"
	postgresVolumeName              = "postgres-volume"
	postgresInitVolumeName          = "postgres-init"
	postgresInitConfigMapName       = "postgres-init-cm"
	postgresVolumeClaimName         = "postgres-storage"
	defaultPostgresStorageClassName = "postgres-storage-class"
)

// PostgresOpts are options that are applicable to postgres.
type PostgresOpts struct {
	Nodes  int
	Volume string

	// CPURequest is the amount of CPU (in cores) we request for each
	// postgres node. If empty, assets.go will choose a default size.
	CPURequest string

	// MemRequest is the amount of memory we request for each postgres
	// node. If empty, assets.go will choose a default size.
	MemRequest string

	// StorageClassName is the name of an existing StorageClass to use when
	// creating a StatefulSet for dynamic postgres storage. If unset, a new
	// StorageClass will be created for the StatefulSet.
	StorageClassName string
}

// PostgresDeployment generates a Deployment for the pachyderm postgres instance.
func PostgresDeployment(opts *AssetOpts, hostPath string) *apps.Deployment {
	cpu := resource.MustParse(opts.PostgresOpts.CPURequest)
	mem := resource.MustParse(opts.PostgresOpts.MemRequest)
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
						Path: path.Join(hostPath, "postgres"),
					},
				},
			},
		}
	}
	volumes = append(volumes, v1.Volume{
		Name: postgresInitVolumeName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{Name: postgresInitConfigMapName},
			},
		},
	})
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
			APIVersion: "apps/v1",
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
									MountPath: "/var/lib/postgresql",
								},
								{
									Name:      postgresInitVolumeName,
									MountPath: "/docker-entrypoint-initdb.d",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							Resources:       resourceRequirements,
							Env: []v1.EnvVar{
								// TODO: Figure out how we want to handle auth in real deployments.
								// The auth has been removed for now to allow PFS tests to run against
								// a deployed Postgres instance.
								{Name: "POSTGRES_DB", Value: dbutil.DefaultDBName},
								{Name: "POSTGRES_HOST_AUTH_METHOD", Value: "trust"},
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

// PostgresStorageClass creates a storage class used for dynamic volume
// provisioning.  Currently dynamic volume provisioning only works
// on AWS and GCE.
func PostgresStorageClass(opts *AssetOpts, backend Backend) (interface{}, error) {
	return makeStorageClass(opts, backend, defaultPostgresStorageClassName, labels(postgresName))
}

// PostgresHeadlessService returns a headless postgres service, which is only for DNS
// resolution.
func PostgresHeadlessService(opts *AssetOpts) *v1.Service {
	ports := []v1.ServicePort{
		{
			Name: "client-port",
			Port: 5432,
		},
	}
	return makeHeadlessService(opts, postgresName, postgresHeadlessServiceName, ports)
}

// PostgresStatefulSet returns a stateful set that manages an etcd cluster
func PostgresStatefulSet(opts *AssetOpts, backend Backend, diskSpace int) interface{} {
	mem := resource.MustParse(opts.PostgresOpts.MemRequest)
	cpu := resource.MustParse(opts.PostgresOpts.CPURequest)
	volumes := []v1.Volume{
		v1.Volume{
			Name: postgresInitVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{Name: postgresInitConfigMapName},
				},
			},
		},
	}
	var pvcTemplates []interface{}
	switch backend {
	case GoogleBackend, AmazonBackend:
		storageClassName := opts.PostgresOpts.StorageClassName
		if storageClassName == "" {
			storageClassName = defaultPostgresStorageClassName
		}
		pvcTemplates = []interface{}{
			map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":   postgresVolumeClaimName,
					"labels": labels(postgresName),
					"annotations": map[string]string{
						"volume.beta.kubernetes.io/storage-class": storageClassName,
					},
					"namespace": opts.Namespace,
				},
				"spec": map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"storage": resource.MustParse(fmt.Sprintf("%vGi", diskSpace)),
						},
					},
					"accessModes": []string{"ReadWriteOnce"},
				},
			},
		}
	default:
		pvcTemplates = []interface{}{
			map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      postgresVolumeClaimName,
					"labels":    labels(postgresName),
					"namespace": opts.Namespace,
				},
				"spec": map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"storage": resource.MustParse(fmt.Sprintf("%vGi", diskSpace)),
						},
					},
					"accessModes": []string{"ReadWriteOnce"},
				},
			},
		}
	}
	var imagePullSecrets []map[string]string
	if opts.ImagePullSecret != "" {
		imagePullSecrets = append(imagePullSecrets, map[string]string{"name": opts.ImagePullSecret})
	}
	// As of March 17, 2017, the Kubernetes client does not include structs for
	// Stateful Set, so we generate the kubernetes manifest using raw json.
	// TODO(msteffen): we're now upgrading our kubernetes client, so we should be
	// abe to rewrite this spec using k8s client structs
	image := postgresImage
	if opts.Registry != "" {
		image = AddRegistry(opts.Registry, postgresImage)
	}
	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "StatefulSet",
		"metadata": map[string]interface{}{
			"name":      postgresName,
			"labels":    labels(postgresName),
			"namespace": opts.Namespace,
		},
		"spec": map[string]interface{}{
			// Effectively configures a RC
			"serviceName": postgresHeadlessServiceName,
			"replicas":    int(opts.PostgresOpts.Nodes),
			"selector": map[string]interface{}{
				"matchLabels": labels(postgresName),
			},

			// pod template
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      postgresName,
					"labels":    labels(postgresName),
					"namespace": opts.Namespace,
				},
				"spec": map[string]interface{}{
					"imagePullSecrets": imagePullSecrets,
					"containers": []interface{}{
						map[string]interface{}{
							"name":  postgresName,
							"image": image,
							// TODO: Figure out how we want to handle auth in real deployments.
							// The auth has been removed for now to allow PFS tests to run against
							// a deployed Postgres instance.
							"env": []map[string]interface{}{{
								"name":  "POSTGRES_DB",
								"value": dbutil.DefaultDBName,
							}, {
								"name":  "POSTGRES_HOST_AUTH_METHOD",
								"value": "trust",
							}},
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": 5432,
									"name":          "client-port",
								},
							},
							"volumeMounts": []interface{}{
								map[string]interface{}{
									"name":      postgresVolumeClaimName,
									"mountPath": "/var/lib/postgresql",
								},
								map[string]interface{}{
									"name":      postgresInitVolumeName,
									"mountPath": "/docker-entrypoint-initdb.d",
								},
							},
							"imagePullPolicy": "IfNotPresent",
							"resources": map[string]interface{}{
								"requests": map[string]interface{}{
									string(v1.ResourceCPU):    cpu.String(),
									string(v1.ResourceMemory): mem.String(),
								},
							},
						},
					},
					"volumes": volumes,
				},
			},
			"volumeClaimTemplates": pvcTemplates,
		},
	}
}

// PostgresVolume creates a persistent volume backed by a volume with name "name"
func PostgresVolume(persistentDiskBackend Backend, opts *AssetOpts,
	hostPath string, name string, size int) (*v1.PersistentVolume, error) {
	return makePersistentVolume(opts, persistentDiskBackend, hostPath, name, size, postgresVolumeName, labels(postgresName))
}

// PostgresVolumeClaim creates a persistent volume claim of 'size' GB.
//
// Note that if you're controlling Postgres with a Stateful Set, this is
// unnecessary (the stateful set controller will create PVCs automatically).
func PostgresVolumeClaim(size int, opts *AssetOpts) *v1.PersistentVolumeClaim {
	return makeVolumeClaim(opts, size, postgresVolumeName, postgresVolumeClaimName, labels(postgresName))
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

// PostgresInitConfigMap generates a configmap which can be mounted into
// the postgres container to initialize the database.
func PostgresInitConfigMap(opts *AssetOpts) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(postgresInitConfigMapName, labels(postgresName), nil, opts.Namespace),
		Data: map[string]string{
			"init-db.sh": `
#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE dex;
    GRANT ALL PRIVILEGES ON DATABASE dex TO postgres;
EOSQL
`,
		},
	}
}
