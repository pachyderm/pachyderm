package client

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Health health checks pachd, it returns an error if pachd isn't healthy.
func (c APIClient) Health() error {
	var req grpc_health_v1.HealthCheckRequest

	response, err := c.healthClient.Check(c.Ctx(), &req)
	if err != nil {
		return errors.Errorf("health check errored %w", err)
	}
	if response.Status == grpc_health_v1.HealthCheckResponse_NOT_SERVING {
		return errors.Errorf("server not ready")
	}
	return nil
}
