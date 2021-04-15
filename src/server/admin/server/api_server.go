package server

import (
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"google.golang.org/protobuf/types/known/emptypb"

	"golang.org/x/net/context"
)

type apiServer struct {
	log.Logger
	clusterInfo *admin.ClusterInfo
}

func (a *apiServer) InspectCluster(ctx context.Context, request *emptypb.Empty) (*admin.ClusterInfo, error) {
	return a.clusterInfo, nil
}
