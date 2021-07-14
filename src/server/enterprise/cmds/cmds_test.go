package cmds

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestActivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	code := tu.GetTestEnterpriseCode(t)

	require.NoError(t, tu.BashCmd(`echo {{.license}} | pachctl license activate`,
		"license", code).Run())

	require.NoError(t, tu.BashCmd(`pachctl enterprise get-state | match ACTIVE`).Run())
}

func TestManuallyJoinLicenseServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate --no-register
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://localhost:1653 --pachd-address grpc://localhost:1653
		pachctl enterprise get-state | match ACTIVE
		pachctl license list-clusters \
		  | match 'id: {{.id}}' \
		  | match -v 'last_heartbeat: <nil>'
		`,
		"id", tu.UniqueString("cluster"),
		"license", tu.GetTestEnterpriseCode(t),
	).Run())
}
