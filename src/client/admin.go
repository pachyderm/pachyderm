package client

import (
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// InspectCluster retrieves cluster state
func (c APIClient) InspectCluster() (*admin.ClusterInfo, error) {
	clusterInfo, err := c.AdminAPIClient.InspectCluster(c.Ctx(), &emptypb.Empty{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to inspect cluster")
	}
	return clusterInfo, nil
}
