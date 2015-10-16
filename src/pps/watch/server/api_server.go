package server

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pachyderm.com/pachyderm/src/pps/watch"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

var (
	emptyInstance = &google_protobuf.Empty{}
)

type apiServer struct {
	pfsAPIClient     pfs.ApiClient
	persistAPIClient persist.APIClient
}

func newAPIServer(
	pfsAPIClient pfs.ApiClient,
	persistAPIClient persist.APIClient,
) *apiServer {
	return &apiServer{
		pfsAPIClient,
		persistAPIClient,
	}
}

func (a *apiServer) Start(ctx context.Context, request *google_protobuf.Empty) (*google_protobuf.Empty, error) {
	return emptyInstance, nil
}

func (a *apiServer) RegisterChangeEvent(ctx context.Context, request *watch.ChangeEvent) (*google_protobuf.Empty, error) {
	return emptyInstance, nil
}
