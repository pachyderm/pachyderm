package client

import (
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"google.golang.org/protobuf/types/known/emptypb"
)

// InspectCluster retrieves cluster state
func (c APIClient) InspectCluster() (*admin.ClusterInfo, error) {
	clusterInfo, err := c.AdminAPIClient.InspectCluster(c.Ctx(), &emptypb.Empty{})
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return clusterInfo, nil
}
