package server

import (
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

// APIServer represents and APIServer
type APIServer interface {
	admin.APIServer
}

// NewAPIServer returns a new admin.APIServer
func NewAPIServer(env serviceenv.ServiceEnv) APIServer {
	return &apiServer{
		Logger: log.NewLogger("admin.API"),
		clusterInfo: &admin.ClusterInfo{
			ID:           env.ClusterID(),
			DeploymentID: env.Config().DeploymentID,
		},
	}
}
