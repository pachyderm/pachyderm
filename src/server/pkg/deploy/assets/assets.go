package assets

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/ugorji/go/codec"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	api "k8s.io/kubernetes/pkg/api/v1"
	extensions "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
)

var (
	suite                  = "pachyderm"
	volumeSuite            = "pachyderm-pps-storage"
	pachdImage             = "pachyderm/pachd"
	etcdImage              = "gcr.io/google_containers/etcd:2.0.12"
	rethinkImage           = "rethinkdb:2.3.3"
	serviceAccountName     = "pachyderm"
	etcdName               = "etcd"
	pachdName              = "pachd"
	rethinkName            = "rethink"
	rethinkVolumeName      = "rethink-volume"
	rethinkVolumeClaimName = "rethink-volume-claim"
	amazonSecretName       = "amazon-secret"
	googleSecretName       = "google-secret"
	initName               = "pachd-init"
	trueVal                = true
)

type backend int

const (
	localBackend backend = iota
	amazonBackend
	googleBackend
)

func ServiceAccount() *api.ServiceAccount {
	return &api.ServiceAccount{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   serviceAccountName,
			Labels: labels(""),
		},
	}
}

//PachdRc TODO secrets is only necessary because dockerized kube chokes on them
func PachdRc(shards uint64, backend backend, hostPath string) *api.ReplicationController {
	volumes := []api.Volume{
		{
			Name: "pach-disk",
		},
	}
	volumeMounts := []api.VolumeMount{
		{
			Name:      "pach-disk",
			MountPath: "/pach",
		},
	}
	readinessProbe := &api.Probe{
		Handler: api.Handler{
			Exec: &api.ExecAction{
				Command: []string{
					"./pachd",
					"--readiness-check",
				},
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      1,
	}
	var backendEnvVar string
	switch backend {
	case localBackend:
		if hostPath == "" {
			hostPath = "/tmp/pach"
		}
		volumes[0].HostPath = &api.HostPathVolumeSource{
			Path: hostPath,
		}
	case amazonBackend:
		backendEnvVar = server.AmazonBackendEnvVar
		volumes = append(volumes, api.Volume{
			Name: amazonSecretName,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: amazonSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      amazonSecretName,
			MountPath: "/" + amazonSecretName,
		})
	case googleBackend:
		backendEnvVar = server.GoogleBackendEnvVar
		volumes = append(volumes, api.Volume{
			Name: googleSecretName,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: googleSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      googleSecretName,
			MountPath: "/" + googleSecretName,
		})
	}
	replicas := int32(2)
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   pachdName,
			Labels: labels(pachdName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: &replicas,
			Selector: map[string]string{
				"app": pachdName,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   pachdName,
					Labels: labels(pachdName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  pachdName,
							Image: pachdImage,
							Env: []api.EnvVar{
								{
									Name:  "PACH_ROOT",
									Value: "/pach",
								},
								{
									Name:  "NUM_SHARDS",
									Value: strconv.FormatUint(shards, 10),
								},
								{
									Name:  "STORAGE_BACKEND",
									Value: backendEnvVar,
								},
								{
									Name: "PACHD_POD_NAMESPACE",
									ValueFrom: &api.EnvVarSource{
										FieldRef: &api.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
							},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 650,
									Protocol:      "TCP",
									Name:          "api-grpc-port",
								},
								{
									ContainerPort: 1050,
									Name:          "trace-port",
								},
							},
							VolumeMounts: volumeMounts,
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
							ReadinessProbe:  readinessProbe,
							ImagePullPolicy: "IfNotPresent",
						},
					},
					ServiceAccountName: serviceAccountName,
					Volumes:            volumes,
				},
			},
		},
	}
}

func PachdService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   pachdName,
			Labels: labels(pachdName),
		},
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": pachdName,
			},
			Ports: []api.ServicePort{
				{
					Port:     650,
					Name:     "api-grpc-port",
					NodePort: 30650,
				},
			},
		},
	}
}

func EtcdRc() *api.ReplicationController {
	replicas := int32(1)
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   etcdName,
			Labels: labels(etcdName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: &replicas,
			Selector: map[string]string{
				"app": etcdName,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   etcdName,
					Labels: labels(etcdName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  etcdName,
							Image: etcdImage,
							//TODO figure out how to get a cluster of these to talk to each other
							Command: []string{"/usr/local/bin/etcd", "--bind-addr=0.0.0.0:2379", "--data-dir=/var/etcd/data"},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 2379,
									Name:          "client-port",
								},
								{
									ContainerPort: 2380,
									Name:          "peer-port",
								},
							},
							VolumeMounts: []api.VolumeMount{
								{
									Name:      "etcd-storage",
									MountPath: "/var/data/etcd",
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
					Volumes: []api.Volume{
						{
							Name: "etcd-storage",
						},
					},
				},
			},
		},
	}
}

func EtcdService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   etcdName,
			Labels: labels(etcdName),
		},
		Spec: api.ServiceSpec{
			Selector: map[string]string{
				"app": etcdName,
			},
			Ports: []api.ServicePort{
				{
					Port: 2379,
					Name: "client-port",
				},
				{
					Port: 2380,
					Name: "peer-port",
				},
			},
		},
	}
}

func RethinkRc(backend backend, volume string) *api.ReplicationController {
	replicas := int32(1)
	spec := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkName,
			Labels: labels(rethinkName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: &replicas,
			Selector: map[string]string{
				"app": rethinkName,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   rethinkName,
					Labels: labels(rethinkName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  rethinkName,
							Image: rethinkImage,
							//TODO figure out how to get a cluster of these to talk to each other
							Command: []string{"rethinkdb", "-d", "/var/rethinkdb/data", "--bind", "all"},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 8080,
									Name:          "admin-port",
								},
								{
									ContainerPort: 28015,
									Name:          "driver-port",
								},
								{
									ContainerPort: 29015,
									Name:          "cluster-port",
								},
							},
							VolumeMounts: []api.VolumeMount{
								{
									Name:      "rethink-storage",
									MountPath: "/var/rethinkdb/data",
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
					Volumes: []api.Volume{
						{
							Name: "rethink-storage",
						},
					},
				},
			},
		},
	}

	if backend != localBackend && volume != "" {
		spec.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim = &api.PersistentVolumeClaimVolumeSource{
			ClaimName: rethinkVolumeClaimName,
		}
	}

	return spec
}

func RethinkService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkName,
			Labels: labels(rethinkName),
		},
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": rethinkName,
			},
			Ports: []api.ServicePort{
				{
					Port:     8080,
					Name:     "admin-port",
					NodePort: 32080,
				},
				{
					Port:     28015,
					Name:     "driver-port",
					NodePort: 32081,
				},
				{
					Port:     29015,
					Name:     "cluster-port",
					NodePort: 32082,
				},
			},
		},
	}
}

func InitJob() *extensions.Job {
	return &extensions.Job{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Job",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   initName,
			Labels: labels(initName),
		},
		Spec: extensions.JobSpec{
			Selector: &extensions.LabelSelector{
				MatchLabels: labels(initName),
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   initName,
					Labels: labels(initName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  initName,
							Image: pachdImage,
							Env: []api.EnvVar{
								{
									Name:  "PACH_ROOT",
									Value: "/pach",
								},
								{
									Name:  "INIT",
									Value: "true",
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
					RestartPolicy: "OnFailure",
				},
			},
		},
	}
}

func AmazonSecret(bucket string, id string, secret string, token string, region string) *api.Secret {
	return &api.Secret{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   amazonSecretName,
			Labels: labels(amazonSecretName),
		},
		Data: map[string][]byte{
			"bucket": []byte(bucket),
			"id":     []byte(id),
			"secret": []byte(secret),
			"token":  []byte(token),
			"region": []byte(region),
		},
	}
}

func GoogleSecret(bucket string) *api.Secret {
	return &api.Secret{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   googleSecretName,
			Labels: labels(googleSecretName),
		},
		Data: map[string][]byte{
			"bucket": []byte(bucket),
		},
	}
}

func RethinkVolume(backend backend, name string, size int) *api.PersistentVolume {
	spec := &api.PersistentVolume{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkVolumeName,
			Labels: volumeLabels(rethinkVolumeName),
		},
		Spec: api.PersistentVolumeSpec{
			Capacity: map[api.ResourceName]resource.Quantity{
				"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
			},
			AccessModes:                   []api.PersistentVolumeAccessMode{api.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: api.PersistentVolumeReclaimRetain,
		},
	}

	switch backend {
	case amazonBackend:
		spec.Spec.PersistentVolumeSource = api.PersistentVolumeSource{
			AWSElasticBlockStore: &api.AWSElasticBlockStoreVolumeSource{
				FSType:   "ext4",
				VolumeID: name,
			},
		}
	case googleBackend:
		spec.Spec.PersistentVolumeSource = api.PersistentVolumeSource{
			GCEPersistentDisk: &api.GCEPersistentDiskVolumeSource{
				FSType: "ext4",
				PDName: name,
			},
		}
	default:
		panic("cannot generate volume spec for unknown backend")
	}

	return spec
}

func RethinkVolumeClaim(size int) *api.PersistentVolumeClaim {
	return &api.PersistentVolumeClaim{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkVolumeClaimName,
			Labels: volumeLabels(rethinkVolumeClaimName),
		},
		Spec: api.PersistentVolumeClaimSpec{
			Resources: api.ResourceRequirements{
				Requests: map[api.ResourceName]resource.Quantity{
					"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
				},
			},
			AccessModes: []api.PersistentVolumeAccessMode{api.ReadWriteOnce},
		},
	}
}

// WriteAssets creates the assets in a dir. It expects dir to already exist.
func WriteAssets(w io.Writer, shards uint64, backend backend, volumeName string, volumeSize int, hostPath string) {
	encoder := codec.NewEncoder(w, &codec.JsonHandle{Indent: 2})

	ServiceAccount().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	if backend != localBackend && volumeName != "" {
		RethinkVolume(backend, volumeName, volumeSize).CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
		RethinkVolumeClaim(volumeSize).CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
	}

	EtcdRc().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	EtcdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	RethinkService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	RethinkRc(backend, volumeName).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	InitJob().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	PachdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	PachdRc(shards, backend, hostPath).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

func WriteLocalAssets(w io.Writer, shards uint64, hostPath string) {
	WriteAssets(w, shards, localBackend, "", 0, hostPath)
}

func WriteAmazonAssets(w io.Writer, shards uint64, bucket string, id string, secret string, token string, region string, volumeName string, volumeSize int) {
	WriteAssets(w, shards, amazonBackend, volumeName, volumeSize, "")
	encoder := codec.NewEncoder(w, &codec.JsonHandle{Indent: 2})
	AmazonSecret(bucket, id, secret, token, region).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

func WriteGoogleAssets(w io.Writer, shards uint64, bucket string, volumeName string, volumeSize int) {
	WriteAssets(w, shards, googleBackend, volumeName, volumeSize, "")
	encoder := codec.NewEncoder(w, &codec.JsonHandle{Indent: 2})
	GoogleSecret(bucket).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

func labels(name string) map[string]string {
	return map[string]string{
		"app":   name,
		"suite": suite,
	}
}

func volumeLabels(name string) map[string]string {
	return map[string]string{
		"app":   name,
		"suite": volumeSuite,
	}
}
