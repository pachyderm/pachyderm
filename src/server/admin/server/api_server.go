package server

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/sirupsen/logrus"
)

// Env is the set of dependencies required by an APIServer
type Env struct {
	ClusterID string
	Config    *serviceenv.Configuration
	Logger    *logrus.Logger
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv) Env {
	return Env{
		ClusterID: senv.ClusterID(),
		Config:    senv.Config(),
		Logger:    senv.Logger(),
	}
}

// APIServer represents an APIServer
type APIServer interface {
	admin.APIServer
}

// NewAPIServer returns a new admin.APIServer
func NewAPIServer(env Env) APIServer {
	return &apiServer{
		clusterInfo: &admin.ClusterInfo{
			ID:           env.ClusterID,
			DeploymentID: env.Config.DeploymentID,
		},
	}
}

type apiServer struct {
	clusterInfo *admin.ClusterInfo
}

func (a *apiServer) InspectCluster(ctx context.Context, request *types.Empty) (*admin.ClusterInfo, error) {
	return a.clusterInfo, nil
}
