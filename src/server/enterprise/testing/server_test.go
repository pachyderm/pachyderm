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

// TestRegisterPachd tests registering a pachd with the enterprise server when auth is disabled
func TestRegisterPachd(t *testing.T) {
	ec, err := client.NewEnterpriseClientForTest()
	require.NoError(t, err)

	c, err := client.NewForTest()
	require.NoError(t, err)

	code := tu.GetTestEnterpriseCode(t)

	_, err = ec.License.Activate(c.Ctx(),
		&license.ActivateRequest{
			ActivationCode: code,
		})
	require.NoError(t, err)

	_, err = ec.License.AddCluster(c.Ctx(),
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
}

// TestRegisterPachdAuthenticated tests registering a pachd with the enterprise server when auth is enabled.
func TestRegisterPachdAuthenticated(t *testing.T) {

}
