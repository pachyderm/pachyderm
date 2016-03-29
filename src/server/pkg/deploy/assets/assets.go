package assets

import (
	"fmt"
	"io"
	"strconv"

	"github.com/ugorji/go/codec"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

var (
	suite              = "pachyderm"
	pachdImage         = "pachyderm/pachd"
	etcdImage          = "gcr.io/google_containers/etcd:2.0.12"
	rethinkImage       = "rethinkdb:2.1.5"
	serviceAccountName = "pachyderm"
	etcdName           = "etcd"
	pachdName          = "pachd"
	rethinkName        = "rethink"
	amazonSecretName   = "amazon-secret"
	trueVal            = true
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
func PachdRc(shards uint64, secrets bool) *api.ReplicationController {
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
	if secrets {
		volumes = append(volumes, api.Volume{
			Name: "amazon-secret",
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: amazonSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      "amazon-secret",
			MountPath: "/amazon-secret",
		})
	}
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
			Replicas: 1,
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
							},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 650,
									Protocol:      "TCP",
									Name:          "api-grpc-port",
								},
								{
									ContainerPort: 750,
									Name:          "api-http-port",
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
				{
					Port:     750,
					Name:     "api-http-port",
					NodePort: 30750,
				},
			},
		},
	}
}

func EtcdRc() *api.ReplicationController {
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
			Replicas: 1,
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

func RethinkRc() *api.ReplicationController {
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkName,
			Labels: labels(rethinkName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
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
						//TODO this needs to be real storage
					},
				},
			},
		},
	}
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
			Selector: map[string]string{
				"app": rethinkName,
			},
			Ports: []api.ServicePort{
				{
					Port: 8080,
					Name: "admin-port",
				},
				{
					Port: 28015,
					Name: "driver-port",
				},
				{
					Port: 29015,
					Name: "cluster-port",
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

// WriteAssets creates the assets in a dir. It expects dir to already exist.
func WriteAssets(w io.Writer, shards uint64, secrets bool) {
	encoder := codec.NewEncoder(w, &codec.JsonHandle{Indent: 2})

	ServiceAccount().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	EtcdRc().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	EtcdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	RethinkService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	RethinkRc().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	PachdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	PachdRc(shards, secrets).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

}

func WriteAmazonSecret(w io.Writer, bucket string, id string, secret string, token string, region string) {
	encoder := codec.NewEncoder(w, &codec.JsonHandle{Indent: 2})
	AmazonSecret(bucket, id, secret, token, region).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

func labels(name string) map[string]string {
	return map[string]string{
		"app":   name,
		"suite": suite,
	}
}
