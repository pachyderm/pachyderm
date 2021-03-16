// testing contains integration tests which run against two servers: a pachd, and an enterprise server.
// By contrast, the tests in the server package run against a single pachd.
package testing

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
)

func resetClusterState(t *testing.T) {
	ec, err := client.NewEnterpriseClientForTest()
	require.NoError(t, err)

	c, err := client.NewForTest()
	require.NoError(t, err)

	// Set the root token, in case a previous test failed
	c.SetAuthToken(tu.RootToken)
	ec.SetAuthToken(tu.RootToken)

	require.NoError(t, c.DeleteAll())
	require.NoError(t, ec.DeleteAll())
}

// TestRegisterPachd tests registering a pachd with the enterprise server when auth is disabled
func TestRegisterPachd(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	ec, err := client.NewEnterpriseClientForTest()
	require.NoError(t, err)

	c, err := client.NewForTest()
	require.NoError(t, err)

	code := tu.GetTestEnterpriseCode(t)

	_, err = ec.License.Activate(ec.Ctx(),
		&license.ActivateRequest{
			ActivationCode: code,
		})
	require.NoError(t, err)

	_, err = ec.License.AddCluster(ec.Ctx(),
		&license.AddClusterRequest{
			Id:      "pachd",
			Secret:  "pachd",
			Address: "pachd.default:650",
		})
	if err != nil && !license.IsErrDuplicateClusterID(err) {
		require.NoError(t, err)
	}

	_, err = c.Enterprise.Activate(c.Ctx(),
		&enterprise.ActivateRequest{
			Id:            "pachd",
			Secret:        "pachd",
			LicenseServer: "pach-enterprise.enterprise:650",
		})
	require.NoError(t, err)

	state, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_ACTIVE, state.State)

	clusters, err := ec.License.ListClusters(ec.Ctx(), &license.ListClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(clusters.Clusters))
}
