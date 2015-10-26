package server

import (
	"strconv"

	"go.pachyderm.com/pachyderm/src/pkg/deploy"
	"golang.org/x/net/context"

	"go.pedge.io/google-protobuf"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

//TODO these names are a bit unwieldy
var (
	emptyInstance      = &google_protobuf.Empty{}
	pfsdImage          = "pachyderm/pfsd"
	rolerImage         = "pachyderm/pfs-roler"
	ppsdImage          = "pachyderm/ppsd"
	btrfsImage         = "pachyderm_btrfs"
	etcdImage          = "gcr.io/google_containers/etcd:2.0.12"
	rethinkImage       = "rethinkdb:2.1.5"
	trueVal            = true
	etcdRcName         = "etcd-rc"
	etcdServiceName    = "etcd"
	rethinkRcName      = "rethink-rc"
	rethinkServiceName = "rethink"
	pfsdRcName         = "pfsd-rc"
	rolerRcName        = "roler-rc"
	pfsdServiceName    = "pfsd"
	ppsdRcName         = "ppsd-rc"
	ppsdServiceName    = "ppsd"
)

type apiServer struct {
	client *client.Client
}

func newAPIServer(client *client.Client) APIServer {
	return &apiServer{client}
}

func (a *apiServer) CreateCluster(ctx context.Context, request *deploy.CreateClusterRequest) (*google_protobuf.Empty, error) {
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(etcdReplicationController()); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(etcdService()); err != nil {
		return nil, err
	}
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(rethinkReplicationController()); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(rethinkService()); err != nil {
		return nil, err
	}
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(
		pfsdRc(
			request.Nodes,
			request.Shards,
			request.Replicas,
		),
	); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(pfsdService()); err != nil {
		return nil, err
	}
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(
		rolerRc(
			request.Shards,
			request.Replicas,
		),
	); err != nil {
		return nil, err
	}
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(
		ppsdRc(
			request.Nodes,
		),
	); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(ppsService()); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *apiServer) UpdateCluster(ctx context.Context, request *deploy.UpdateClusterRequest) (*google_protobuf.Empty, error) {
	return emptyInstance, nil
}

func (a *apiServer) InspectCluster(ctx context.Context, request *deploy.InspectClusterRequest) (*deploy.ClusterInfo, error) {
	return nil, nil
}

func (a *apiServer) ListCluster(ctx context.Context, request *deploy.ListClusterRequest) (*deploy.ClusterInfos, error) {
	return nil, nil
}

func (a *apiServer) DeleteCluster(ctx context.Context, request *deploy.DeleteClusterRequest) (*google_protobuf.Empty, error) {
	if err := a.client.ReplicationControllers(api.NamespaceDefault).Delete(etcdRcName); err != nil {
		return nil, err
	}
	if err := a.client.Services(api.NamespaceDefault).Delete(etcdServiceName); err != nil {
		return nil, err
	}
	if err := a.client.ReplicationControllers(api.NamespaceDefault).Delete(rethinkRcName); err != nil {
		return nil, err
	}
	if err := a.client.Services(api.NamespaceDefault).Delete(rethinkServiceName); err != nil {
		return nil, err
	}
	if err := a.client.ReplicationControllers(api.NamespaceDefault).Delete(pfsdRcName); err != nil {
		return nil, err
	}
	if err := a.client.Services(api.NamespaceDefault).Delete(pfsdServiceName); err != nil {
		return nil, err
	}
	if err := a.client.Services(api.NamespaceDefault).Delete(rolerRcName); err != nil {
		return nil, err
	}
	if err := a.client.ReplicationControllers(api.NamespaceDefault).Delete(ppsdRcName); err != nil {
		return nil, err
	}
	if err := a.client.Services(api.NamespaceDefault).Delete(ppsdServiceName); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func pfsdRc(nodes uint64, shards uint64, replicas uint64) *api.ReplicationController {
	app := "pfsd"
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: pfsdRcName,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: int(nodes),
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name: "pfsd",
					Labels: map[string]string{
						"app": app,
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  "pfsd",
							Image: pfsdImage,
							Env: []api.EnvVar{
								{
									Name:  "PFS_DRIVER_ROOT",
									Value: "/pfs/btrfs",
								},
								{
									Name:  "BTRFS_DEVICE",
									Value: "/pfs-img/btrfs.img",
								},
								{
									Name:  "PFS_NUM_SHARDS",
									Value: strconv.FormatUint(shards, 10),
								},
								{
									Name:  "PFS_NUM_REPLICAS",
									Value: strconv.FormatUint(replicas, 10),
								},
							},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 650,
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
							VolumeMounts: []api.VolumeMount{
								{
									Name:      "pfs-disk",
									MountPath: "/pfs/btrfs",
								},
							},
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
						},
					},
					Volumes: []api.Volume{
						{
							Name: "pfs-disk",
							//api.VolumeSource{
							//	GCEPersistentDisk: &api.GCEPersistentDiskVolumeSource{
							//		PDName: "pch-pfs",
							//		FSType: "btrfs",
							//	},
							//},
						},
					},
				},
			},
		},
	}
}

func pfsdService() *api.Service {
	app := "pfsd"
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: pfsdServiceName,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: api.ServiceSpec{
			Selector: map[string]string{
				"app": app,
			},
			Ports: []api.ServicePort{
				{
					Port: 650,
					Name: "api-grpc-port",
				},
				{
					Port: 750,
					Name: "api-http-port",
				},
			},
		},
	}
}

func rolerRc(shards uint64, replicas uint64) *api.ReplicationController {
	app := "roler"
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: rolerRcName,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name: "rolerkj",
					Labels: map[string]string{
						"app": app,
					},
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
								{
									Name:  "PFS_NUM_REPLICAS",
									Value: strconv.FormatUint(replicas, 10),
								},
							},
						},
					},
				},
			},
		},
	}
}

func ppsdRc(nodes uint64) *api.ReplicationController {
	app := "ppsd"
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: ppsdRcName,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name: "ppsd",
					Labels: map[string]string{
						"app": app,
					},
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
									Name:          "api-grpc-port",
								},
								{
									ContainerPort: 1051,
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

func ppsService() *api.Service {
	app := "ppsd"
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: ppsdServiceName,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: api.ServiceSpec{
			Selector: map[string]string{
				"app": app,
			},
			Ports: []api.ServicePort{
				{
					Port: 651,
					Name: "api-grpc-port",
				},
			},
		},
	}
}

func etcdReplicationController() *api.ReplicationController {
	app := "etcd"
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: etcdRcName,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name: "etcd-pod",
					Labels: map[string]string{
						"app": app,
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  "etcd",
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

func etcdService() *api.Service {
	app := "etcd"
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: etcdServiceName,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: api.ServiceSpec{
			Selector: map[string]string{
				"app": app,
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

func rethinkReplicationController() *api.ReplicationController {
	app := "rethink"
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: rethinkRcName,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name: "rethink-pod",
					Labels: map[string]string{
						"app": app,
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  "rethink",
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

func rethinkService() *api.Service {
	app := "rethink"
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: rethinkServiceName,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: api.ServiceSpec{
			Selector: map[string]string{
				"app": app,
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
