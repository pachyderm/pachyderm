package watch

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

var (
	emptyInstance = &google_protobuf.Empty{}
)

type apiServer struct {
	ppsAPIClient     pps.APIClient
	pfsAPIClient     pfs.ApiClient
	persistAPIClient persist.APIClient
}

func newAPIServer(
	ppsAPIClient pps.APIClient,
	pfsAPIClient pfs.ApiClient,
	persistAPIClient persist.APIClient,
) *apiServer {
	return &apiServer{
		ppsAPIClient,
		pfsAPIClient,
		persistAPIClient,
	}
}

func (a *apiServer) RegisterChangeEvent(ctx context.Context, request *ChangeEvent) (*google_protobuf.Empty, error) {
	return emptyInstance, nil
}
