//go:build k8s

package cmds

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

// TestTransaction runs a straightforward end-to-end test of starting, adding
// to and finishing a transaction.
func TestTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	repo := tu.UniqueString("TestTransaction-repo")
	branch := tu.UniqueString("Test")
	setup := tu.PachctlBashCmd(t, c, `
		pachctl start transaction > /dev/null
		pachctl create repo {{.repo}}
		pachctl create branch {{.repo}}@{{.branch}}
		`,
		"repo", repo,
		"branch", branch,
	)

	output, err := (*exec.Cmd)(setup).Output()
	require.NoError(t, err)
	strs := strings.Split(strings.TrimSpace(string(output)), " ")
	txn := strs[len(strs)-1]
	t.Logf("DNJ TODO output: %v --- commit: %v", string(output), txn)

	// The repo shouldn't exist yet
	require.YesError(t, tu.PachctlBashCmd(t, c, "pachctl inspect repo {{.repo}}", "repo", repo).Run())

	// Check that finishing the transaction creates the requested objects
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl finish transaction {{.txn}}
		pachctl inspect repo {{.repo}}
		pachctl inspect commit {{.repo}}@{{.branch}}
		`,
		"txn", txn,
		"repo", repo,
		"branch", branch,
	).Run())
}

func startTransaction(t *testing.T, c *client.APIClient) string {
	output, err := (*exec.Cmd)(tu.PachctlBashCmd(t, c, "pachctl start transaction")).Output()
	require.NoError(t, err)

	strs := strings.Split(strings.TrimSpace(string(output)), " ")
	return strs[len(strs)-1]
}

func requireTransactionDoesNotExist(t *testing.T, c *client.APIClient, txn string) {
	output, err := (*exec.Cmd)(tu.PachctlBashCmd(t, c, "pachctl inspect transaction {{.txn}} 2>&1", "txn", txn)).Output()
	require.YesError(t, err)
	expected := fmt.Sprintf("not found")
	require.True(t, strings.Contains(string(output), expected))
}

func TestDeleteActiveTransaction(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	// Start then delete a transaction
	txn := startTransaction(t, c)
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl delete transaction").Run())

	// Check that the transaction no longer exists
	requireTransactionDoesNotExist(t, c, txn)
}

func TestDeleteInactiveTransaction(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	// Start, stop, then delete a transaction
	txn := startTransaction(t, c)
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl stop transaction").Run())
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl delete transaction {{.txn}}", "txn", txn).Run())

	// Check that the transaction no longer exists
	requireTransactionDoesNotExist(t, c, txn)
}
