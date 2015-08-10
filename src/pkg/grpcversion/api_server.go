package grpcversion

import (
	"github.com/peter-edge/go-google-protobuf"
	"golang.org/x/net/context"
)

type apiServer struct {
	getVersionResponse *GetVersionResponse
}

func newAPIServer(version *Version) *apiServer {
	return &apiServer{&GetVersionResponse{version}}
}

func (a *apiServer) GetVersion(_ context.Context, _ *google_protobuf.Empty) (*GetVersionResponse, error) {
	return a.getVersionResponse, nil
}
