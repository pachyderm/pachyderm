package health

import (
	"github.com/pachyderm/pachyderm/src/client/health"
	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
)

// NewHealthServer returns a new health server
func NewHealthServer() health.HealthServer {
	return &healthServer{}
}

type healthServer struct{}

func (*healthServer) Health(context.Context, *google_protobuf.Empty) (*google_protobuf.Empty, error) {
	return google_protobuf.EmptyInstance, nil
}
