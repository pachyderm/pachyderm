package health

import (
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/health"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"golang.org/x/net/context"
)

// Server adds the Ready method to health.HealthServer.
type Server interface {
	health.HealthServer
	Ready()
}

// NewHealthServer returns a new health server
func NewHealthServer() Server {
	return &healthServer{}
}

type healthServer struct {
	ready bool
}

// Health implements the Health method for healthServer.
func (h *healthServer) Health(context.Context, *types.Empty) (*types.Empty, error) {
	if !h.ready {
		return nil, errors.Errorf("server not ready")
	}
	return &types.Empty{}, nil
}

// Ready tells pachd to start responding positively to Health requests. This
// will cause the node to pass its k8s readiness check.
func (h *healthServer) Ready() {
	h.ready = true
}
