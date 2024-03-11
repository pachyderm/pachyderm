package client

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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
func (c APIClient) InspectClusterWithVersionAndProject(ctx context.Context, v *versionpb.Version, p *pfs.Project) (*admin.ClusterInfo, error) {
	clusterInfo, err := c.AdminAPIClient.InspectCluster(ctx, &admin.InspectClusterRequest{
		ClientVersion:  v,
		CurrentProject: p,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to inspect cluster")
	}
	return clusterInfo, nil
}
