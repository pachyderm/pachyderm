package server

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/provider"
	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

//TODO these names are a bit unwieldy
var (
	emptyInstance = &google_protobuf.Empty{}
)

type apiServer struct {
	client   *client.Client
	provider provider.Provider
}

func newAPIServer(client *client.Client, provider provider.Provider) APIServer {
	return &apiServer{client, provider}
}

func (a *apiServer) CreateCluster(ctx context.Context, request *deploy.CreateClusterRequest) (*google_protobuf.Empty, error) {
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(assets.EtcdRc()); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(assets.EtcdService()); err != nil {
		return nil, err
	}
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(assets.RethinkRc()); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(assets.RethinkService()); err != nil {
		return nil, err
	}
	if _, err := a.client.ReplicationControllers(api.NamespaceDefault).Create(assets.PachdRc(request.Shards, false)); err != nil {
		return nil, err
	}
	if _, err := a.client.Services(api.NamespaceDefault).Create(assets.PachdService()); err != nil {
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
