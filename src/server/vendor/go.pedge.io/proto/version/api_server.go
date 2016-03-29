package protoversion

import (
	"time"

	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
)

type apiServer struct {
	protorpclog.Logger
	version *Version
	options APIServerOptions
}

func newAPIServer(version *Version, options APIServerOptions) *apiServer {
	return &apiServer{protorpclog.NewLogger("protoversion.API"), version, options}
}

func (a *apiServer) GetVersion(ctx context.Context, request *google_protobuf.Empty) (response *Version, err error) {
	if !a.options.DisableLogging {
		defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	}
	return a.version, nil
}
