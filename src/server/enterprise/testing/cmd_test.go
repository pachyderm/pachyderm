// testing contains integration tests which run against two servers: a pachd, and an enterprise server.
// By contrast, the tests in the server package run against a single pachd.
package testing

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
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
	require.NoError(t, ec.DeleteAllEnterprise())
}

// TestRegisterPachd tests registering a pachd with the enterprise server when auth is disabled
func TestRegisterPachd(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address pachd.default:650
		pachctl enterprise get-state | match ACTIVE
		pachctl license list-clusters \
		  | match 'id: {{.id}}' \
		  | match -v 'last_heartbeat: <nil>'
		`,
		"id", tu.UniqueString("cluster"),
		"license", tu.GetTestEnterpriseCode(t),
	).Run())
}
