package health

import (
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/health"
	"golang.org/x/net/context"
)

// NewHealthServer returns a new health server
func NewHealthServer() health.HealthServer {
	return &healthServer{}
}

type healthServer struct{}

func (*healthServer) Health(context.Context, *types.Empty) (*types.Empty, error) {
	return &types.Empty{}, nil
}
