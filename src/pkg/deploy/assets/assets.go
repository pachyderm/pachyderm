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
	pfsdImage          = "pachyderm/pfsd"
	rolerImage         = "pachyderm/pfs-roler"
	ppsdImage          = "pachyderm/ppsd"
	objdImage          = "pachyderm/objd"
	etcdImage          = "gcr.io/google_containers/etcd:2.0.12"
	rethinkImage       = "rethinkdb:2.1.5"
	serviceAccountName = "pachyderm"
	pfsdName           = "pfsd"
	rolerName          = "roler"
	ppsdName           = "ppsd"
	objdName           = "objd"
	etcdName           = "etcd"
	rethinkName        = "rethink"
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

func ObjdRc() *api.ReplicationController {
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   objdName,
			Labels: labels(objdName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": objdName,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   objdName,
					Labels: labels(objdName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  objdName,
							Image: objdImage,
							Env: []api.EnvVar{
								{
									Name:  "OBJ_ROOT",
									Value: "/obj",
								},
							},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 652,
									Protocol:      "TCP",
									HostIP:        "0.0.0.0",
									Name:          "api-grpc-port",
								},
								{
									ContainerPort: 752,
									Name:          "api-http-port",
								},
								{
									ContainerPort: 1052,
									Name:          "trace-port",
								},
							},
							VolumeMounts: []api.VolumeMount{
								{
									Name:      "obj-disk",
									MountPath: "/obj",
								},
							},
						},
					},
					Volumes: []api.Volume{
						{
							Name: "obj-disk",
						},
					},
				},
			},
		},
	}
}

func ObjdService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   objdName,
			Labels: labels(objdName),
		},
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": objdName,
			},
			Ports: []api.ServicePort{
				{
					Port: 652,
					Name: "api-grpc-port",
				},
				{
					Port: 752,
					Name: "api-http-port",
				},
			},
		},
	}
}

func PfsdRc(shards uint64) *api.ReplicationController {
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   pfsdName,
			Labels: labels(pfsdName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": pfsdName,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   pfsdName,
					Labels: labels(pfsdName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  pfsdName,
							Image: pfsdImage,
							Env: []api.EnvVar{
								{
									Name:  "PFS_NUM_SHARDS",
									Value: strconv.FormatUint(shards, 10),
								},
							},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 650,
									Protocol:      "TCP",
									HostIP:        "0.0.0.0",
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
						},
					},
				},
			},
		},
	}
}

func PfsdService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   pfsdName,
			Labels: labels(pfsdName),
		},
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": pfsdName,
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

func RolerRc(shards uint64) *api.ReplicationController {
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rolerName,
			Labels: labels(rolerName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": rolerName,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   "roler",
					Labels: labels(rolerName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  "roler",
							Image: rolerImage,
							Env: []api.EnvVar{
								{
									Name:  "PFS_NUM_SHARDS",
									Value: strconv.FormatUint(shards, 10),
								},
							},
						},
					},
				},
			},
		},
	}
}

func PpsdRc() *api.ReplicationController {
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   ppsdName,
			Labels: labels(ppsdName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": ppsdName,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   "ppsd",
					Labels: labels(ppsdName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  "ppsd",
							Image: ppsdImage,
							Env:   []api.EnvVar{},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 651,
									Protocol:      "TCP",
									HostIP:        "0.0.0.0",
									Name:          "api-grpc-port",
								},
								{
									ContainerPort: 1051,
									Name:          "trace-port",
								},
							},
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
						},
					},
					ServiceAccountName: serviceAccountName,
				},
			},
		},
	}
}

func PpsdService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   ppsdName,
			Labels: labels(ppsdName),
		},
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": ppsdName,
			},
			Ports: []api.ServicePort{
				{
					Port:     651,
					Name:     "api-grpc-port",
					NodePort: 30651,
				},
				{
					Port:     751,
					Name:     "api-http-port",
					NodePort: 30751,
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

// WriteAssets creates the assets in a dir. It expects dir to already exist.
func WriteAssets(w io.Writer, shards uint64) {
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

	ObjdRc().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	ObjdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	PfsdRc(uint64(shards)).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	PfsdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	PpsdRc().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	PpsdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	RolerRc(uint64(shards)).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

func labels(name string) map[string]string {
	return map[string]string{
		"app":   name,
		"suite": suite,
	}
}
