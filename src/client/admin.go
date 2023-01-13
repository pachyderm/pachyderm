package client

import (
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

// InspectCluster retrieves cluster state
func (c APIClient) InspectCluster() (*admin.ClusterInfo, error) {
	clusterInfo, err := c.AdminAPIClient.InspectCluster(c.Ctx(), &admin.InspectClusterRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to inspect cluster")
	}
	return clusterInfo, nil
}

// InspectCluster retrieves cluster state, and sends the server its version for the server to
// validate.
func (c APIClient) InspectClusterWithVersion(v *versionpb.Version) (*admin.ClusterInfo, error) {
	clusterInfo, err := c.AdminAPIClient.InspectCluster(c.Ctx(), &admin.InspectClusterRequest{
		ClientVersion: v,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to inspect cluster")
	}
	return clusterInfo, nil
}
