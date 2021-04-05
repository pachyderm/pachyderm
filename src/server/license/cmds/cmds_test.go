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

	require.NoError(t, tu.BashCmd(`pachctl license get-state | match ACTIVE`).Run())
}

func TestClusterCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	tu.ActivateEnterprise(t, tu.GetPachClient(t))
	defer tu.DeleteAll(t)
	require.NoError(t, tu.BashCmd(`
		pachctl license add-cluster --id {{.id}} --address localhost:653
		pachctl license list-clusters \
                  | match 'id: {{.id}}' \
                  | match 'address: localhost:653' \
                  | match 'id: localhost' \
                  | match 'address: localhost:650'
		pachctl license update-cluster --id {{.id}} --address 127.0.0.1:650
		pachctl license list-clusters \
                  | match 'address: 127.0.0.1:650' \
                  | match 'address: localhost:650'
		pachctl license delete-cluster --id {{.id}}
		pachctl license list-clusters \
		  | match -v 'address: 127.0.0.1:650' \
		  | match -v 'id: {{.id}}'
		`,
		"id", tu.UniqueString("cluster"),
	).Run())
}
