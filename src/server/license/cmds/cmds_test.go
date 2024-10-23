//go:build k8s

package cmds

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/testutilpachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestClusterCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateEnterprise(t, c)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl license add-cluster --id {{.id}} --address grpc://localhost:1653
		pachctl license list-clusters \
                  | match 'id: {{.id}}' \
                  | match 'address: grpc://localhost:1653' \
                  | match 'id: localhost' \
                  | match 'address: grpc://localhost:1650'
		pachctl license update-cluster --id {{.id}} --address grpc://127.0.0.1:1650
		pachctl license list-clusters \
                  | match 'address: grpc://127.0.0.1:1650' \
                  | match 'address: grpc://localhost:1650'
		pachctl license delete-cluster --id {{.id}}
		pachctl license list-clusters \
		  | match -v 'address: grpc://127.0.0.1:1650' \
		  | match -v 'id: {{.id}}'
		`,
		"id", uuid.UniqueString("cluster"),
	).Run())
}
