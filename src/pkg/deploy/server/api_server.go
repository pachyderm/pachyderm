package server

import (
	"fmt"

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
	ppsdImage          = "pachyderm/ppsd"
	btrfsImage         = "pachyderm_btrfs"
	etcdImage          = "gcr.io/google_containers/etcd:2.0.12"
	rethinkImage       = "rethinkdb:2.1.5"
	trueVal            = true
	etcdRcName         = "etcd-rc"
	etcdServiceName    = "etcd"
	rethinkRcName      = "rethink-rc"
	rethinkServiceName = "rethink"
)

func pfsdRcName(name string) string {
	return fmt.Sprintf("pfsd-rc-%s", name)
}

func pfsdServiceName(name string) string {
	return fmt.Sprintf("pfsd-%s", name)
}

func ppsdRcName(name string) string {
	return fmt.Sprintf("ppsd-rc-%s", name)
}

func ppsdServiceName(name string) string {
	return fmt.Sprintf("ppsd-%s", name)
}

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
			request.Cluster.Name,
			request.Nodes,
			request.Shards,
			request.Replicas,
		),
	); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(pfsdService(request.Cluster.Name)); err != nil {
		return nil, err
	}
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(
		ppsdRc(
			request.Cluster.Name,
			request.Nodes,
		),
	); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(ppsService(request.Cluster.Name)); err != nil {
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
	if err := a.client.ReplicationControllers(api.NamespaceDefault).Delete(pfsdRcName(request.Cluster.Name)); err != nil {
		return nil, err
	}
	if err := a.client.Services(api.NamespaceDefault).Delete(pfsdServiceName(request.Cluster.Name)); err != nil {
		return nil, err
	}
	if err := a.client.ReplicationControllers(api.NamespaceDefault).Delete(ppsdRcName(request.Cluster.Name)); err != nil {
		return nil, err
	}
	if err := a.client.Services(api.NamespaceDefault).Delete(ppsdServiceName(request.Cluster.Name)); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func pfsdRc(name string, nodes uint64, shards uint64, replicas uint64) *api.ReplicationController {
	app := fmt.Sprintf("pfsd-%s", name)
	return &api.ReplicationController{
		unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		api.ObjectMeta{
			Name: pfsdRcName(name),
			Labels: map[string]string{
				"app": app,
			},
		},
		api.ReplicationControllerSpec{
			Replicas: int(nodes),
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				api.ObjectMeta{
					Name: fmt.Sprintf("pfsd-%s", name),
					Labels: map[string]string{
						"app": app,
					},
				},
				api.PodSpec{
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
		api.ReplicationControllerStatus{},
	}
}

func pfsdService(name string) *api.Service {
	app := fmt.Sprintf("pfsd-%s", name)
	return &api.Service{
		unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		api.ObjectMeta{
			Name: pfsdServiceName(name),
			Labels: map[string]string{
				"app": app,
			},
		},
		api.ServiceSpec{
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
		api.ServiceStatus{},
	}
}

func ppsdRc(name string, nodes uint64) *api.ReplicationController {
	app := fmt.Sprintf("ppsd-%s", name)
	return &api.ReplicationController{
		unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		api.ObjectMeta{
			Name: ppsdRcName(name),
			Labels: map[string]string{
				"app": app,
			},
		},
		api.ReplicationControllerSpec{
			Replicas: int(nodes),
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				api.ObjectMeta{
					Name: fmt.Sprintf("ppsd-%s", name),
					Labels: map[string]string{
						"app": app,
					},
				},
				api.PodSpec{
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
		api.ReplicationControllerStatus{},
	}
}

func ppsService(name string) *api.Service {
	app := fmt.Sprintf("ppsd-%s", name)
	return &api.Service{
		unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		api.ObjectMeta{
			Name: ppsdServiceName(name),
			Labels: map[string]string{
				"app": app,
			},
		},
		api.ServiceSpec{
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
		api.ServiceStatus{},
	}
}

func etcdReplicationController() *api.ReplicationController {
	app := "etcd"
	return &api.ReplicationController{
		unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		api.ObjectMeta{
			Name: etcdRcName,
			Labels: map[string]string{
				"app": app,
			},
		},
		api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				api.ObjectMeta{
					Name: "etcd-pod",
					Labels: map[string]string{
						"app": app,
					},
				},
				api.PodSpec{
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
		api.ReplicationControllerStatus{},
	}
}

func etcdService() *api.Service {
	app := "etcd"
	return &api.Service{
		unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		api.ObjectMeta{
			Name: etcdServiceName,
			Labels: map[string]string{
				"app": app,
			},
		},
		api.ServiceSpec{
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
		api.ServiceStatus{},
	}
}

func rethinkReplicationController() *api.ReplicationController {
	app := "rethink"
	return &api.ReplicationController{
		unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		api.ObjectMeta{
			Name: rethinkRcName,
			Labels: map[string]string{
				"app": app,
			},
		},
		api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				api.ObjectMeta{
					Name: "rethink-pod",
					Labels: map[string]string{
						"app": app,
					},
				},
				api.PodSpec{
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
		api.ReplicationControllerStatus{},
	}
}

func rethinkService() *api.Service {
	app := "rethink"
	return &api.Service{
		unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		api.ObjectMeta{
			Name: rethinkServiceName,
			Labels: map[string]string{
				"app": app,
			},
		},
		api.ServiceSpec{
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
		api.ServiceStatus{},
	}
}
