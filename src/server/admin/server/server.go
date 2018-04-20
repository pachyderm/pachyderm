package server

import (
	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
)

// APIServer represents and APIServer
type APIServer interface {
	admin.APIServer
}

// NewAPIServer returns a new admin.APIServer
func NewAPIServer(env *serviceenv.ServiceEnv, clusterInfo *admin.ClusterInfo) APIServer {
	return &apiServer{
		Logger:      log.NewLogger("admin.API"),
		env:         env,
		clusterInfo: clusterInfo,
	}
}
