package protoversion

import (
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

type apiServer struct {
	version *Version
}

func newAPIServer(version *Version) *apiServer {
	return &apiServer{version}
}

func (a *apiServer) GetVersion(_ context.Context, _ *google_protobuf.Empty) (*Version, error) {
	return a.version, nil
}
