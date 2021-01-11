package server

import (
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"

	"golang.org/x/net/context"
)

type apiServer struct {
	log.Logger
	clusterInfo *admin.ClusterInfo
}

func (a *apiServer) InspectCluster(ctx context.Context, request *types.Empty) (*admin.ClusterInfo, error) {
	return a.clusterInfo, nil
}
