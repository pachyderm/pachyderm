package server

import (
	"strconv"

	"github.com/pachyderm/pachyderm/src/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/pkg/provider"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
	"golang.org/x/net/context"

	"go.pedge.io/google-protobuf"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

//TODO these names are a bit unwieldy
var (
	emptyInstance      = &google_protobuf.Empty{}
	suite              = "pachyderm"
	defaultDiskSizeGb  = int64(1000)
	pfsdImage          = "pachyderm/pfsd"
	rolerImage         = "pachyderm/pfs-roler"
	ppsdImage          = "pachyderm/ppsd"
	pachImage          = "pachyderm/pach"
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
	sandboxRcName      = "sandbox-rc"
	pfsdServiceName    = "pfsd"
	ppsdRcName         = "ppsd-rc"
	ppsdServiceName    = "ppsd"
)

type apiServer struct {
	client   *client.Client
	provider provider.Provider
}

func newAPIServer(client *client.Client, provider provider.Provider) APIServer {
	return &apiServer{client, provider}
}

func (a *apiServer) CreateCluster(ctx context.Context, request *deploy.CreateClusterRequest) (*google_protobuf.Empty, error) {
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(etcdRc()); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(etcdService()); err != nil {
		return nil, err
	}
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(rethinkRc()); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(rethinkService()); err != nil {
		return nil, err
	}
	if a.provider != nil {
		diskNames, err := a.createDisks(ctx, request.Nodes)
		if err != nil {
			return nil, err
		}
		persistentVolumes := persistantVolumes(diskNames)
		for _, persistantVolume := range persistentVolumes {
			if _, err := a.client.PersistentVolumes().Create(persistantVolume); err != nil {
				return nil, err
			}
		}
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
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(sandboxRc()); err != nil {
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

func (a *apiServer) createDisks(ctx context.Context, nodes uint64) ([]string, error) {
	var names []string
	for i := uint64(0); i < nodes; i++ {
		name := "disk-" + uuid.NewWithoutDashes()
		if err := a.provider.CreateDisk(name, defaultDiskSizeGb); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func persistantVolumes(names []string) []*api.PersistentVolume {
	var result []*api.PersistentVolume
	for _, name := range names {
		result = append(result, &api.PersistentVolume{
			TypeMeta: unversioned.TypeMeta{
				Kind:       "PersistentVolume",
				APIVersion: "v1",
			},
			ObjectMeta: api.ObjectMeta{
				Name: name,
			},
			Spec: api.PersistentVolumeSpec{
				Capacity: api.ResourceList{
					"storage": *resource.NewQuantity(defaultDiskSizeGb*1000*1000*1000, resource.DecimalSI),
				},
				PersistentVolumeSource: api.PersistentVolumeSource{
					GCEPersistentDisk: &api.GCEPersistentDiskVolumeSource{
						PDName: name,
						FSType: "btrfs",
					},
				},
				AccessModes: []api.PersistentVolumeAccessMode{
					api.ReadWriteOnce,
					api.ReadOnlyMany,
				},
			},
		})
	}
	return result
}

func pfsdRc(nodes uint64, shards uint64, replicas uint64) *api.ReplicationController {
	app := "pfsd"
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   pfsdRcName,
			Labels: labels(app),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: int(nodes),
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   "pfsd",
					Labels: labels(app),
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
			Name:   pfsdServiceName,
			Labels: labels(app),
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
			Name:   rolerRcName,
			Labels: labels(app),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   "roler",
					Labels: labels(app),
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
			Name:   ppsdRcName,
			Labels: labels(app),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   "ppsd",
					Labels: labels(app),
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
			Name:   ppsdServiceName,
			Labels: labels(app),
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

func sandboxRc() *api.ReplicationController {
	app := "sandbox"
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   sandboxRcName,
			Labels: labels(app),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   "sandbox-pod",
					Labels: labels(app),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:    "sandbox",
							Image:   pachImage,
							Command: []string{"/pach", "mount"},
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
						},
					},
				},
			},
		},
	}
}

func etcdRc() *api.ReplicationController {
	app := "etcd"
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   etcdRcName,
			Labels: labels(app),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   "etcd-pod",
					Labels: labels(app),
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
			Name:   etcdServiceName,
			Labels: labels(app),
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

func rethinkRc() *api.ReplicationController {
	app := "rethink"
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkRcName,
			Labels: labels(app),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app": app,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   "rethink-pod",
					Labels: labels(app),
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
			Name:   rethinkServiceName,
			Labels: labels(app),
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

func labels(app string) map[string]string {
	return map[string]string{
		"app":   app,
		"suite": suite,
	}
}
