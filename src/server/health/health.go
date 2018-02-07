package health

import (
	"fmt"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/health"
	"golang.org/x/net/context"
)

type HealthServer interface {
	health.HealthServer
	Ready()
}

// NewHealthServer returns a new health server
func NewHealthServer() HealthServer {
	return &healthServer{}
}

type healthServer struct {
	ready bool
}

func (h *healthServer) Health(context.Context, *types.Empty) (*types.Empty, error) {
	if !h.ready {
		return nil, fmt.Errorf("server not ready")
	}
	return &types.Empty{}, nil
}

func (h *healthServer) Ready() {
	h.ready = true
}
