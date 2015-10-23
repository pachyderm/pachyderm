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

var (
	emptyInstance = &google_protobuf.Empty{}
	pfsdImage     = "pachyderm/pfsd"
)

type apiServer struct {
	client *client.Client
}

func newAPIServer(client *client.Client) APIServer {
	return &apiServer{client}
}

func (a *apiServer) CreateCluster(ctx context.Context, request *deploy.CreateClusterRequest) (*google_protobuf.Empty, error) {
	_, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(
		pfsReplicationController(
			request.Cluster.Name,
			request.Nodes,
			request.Shards,
			request.Replicas,
		),
	)
	if err != nil {
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
	return emptyInstance, nil
}

func pfsReplicationController(name string, nodes uint64, shards uint64, replicas uint64) *api.ReplicationController {
	app := fmt.Sprintf("pfsd-%s", name)
	return &api.ReplicationController{
		unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		api.ObjectMeta{
			Name: fmt.Sprintf("pfsd-rc-%s", name),
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
