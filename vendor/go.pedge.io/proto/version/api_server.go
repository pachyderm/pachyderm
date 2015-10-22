package protoversion

import (
	"time"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
)

type apiServer struct {
	protorpclog.Logger
	version *Version
}

func newAPIServer(version *Version) *apiServer {
	return &apiServer{protorpclog.NewLogger("protoversion.API"), version}
}

func (a *apiServer) GetVersion(ctx context.Context, request *google_protobuf.Empty) (response *Version, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.version, nil
}
