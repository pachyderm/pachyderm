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
	t.Skip() // DNJ - tests have not run and need fixing
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	repo := tu.UniqueString("TestTransaction-repo")
	setup := tu.PachctlBashCmd(t, c, `
		pachctl start transaction > /dev/null
		pachctl create repo {{.repo}}
		pachctl create branch {{.repo}}@master
		pachctl start commit {{.repo}}@master
		`,
		"repo", repo,
	)

	output, err := (*exec.Cmd)(setup).Output()
	require.NoError(t, err)
	commit := strings.TrimSpace(string(output))

	// The repo shouldn't exist yet
	require.YesError(t, tu.PachctlBashCmd(t, c, "pachctl inspect repo {{.repo}}", "repo", repo).Run())

	// Check that finishing the transaction creates the requested objects
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl finish transaction
		pachctl inspect repo {{.repo}}
		pachctl inspect commit {{.repo}}@master
		pachctl inspect commit {{.repo}}@{{.commit}}
		`,
		"repo", repo,
		"commit", commit,
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
	expected := fmt.Sprintf("%s not found", txn)
	require.True(t, strings.Contains(string(output), expected))
}

func TestDeleteActiveTransaction(t *testing.T) {
	t.Skip() // DNJ - tests have not run and need fixing
	c, _ := minikubetestenv.AcquireCluster(t)
	// Start then delete a transaction
	txn := startTransaction(t, c)
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl delete transaction").Run())

	// Check that the transaction no longer exists
	requireTransactionDoesNotExist(t, c, txn)
}

func TestDeleteInactiveTransaction(t *testing.T) {
	t.Skip() // DNJ - tests have not run and need fixing
	c, _ := minikubetestenv.AcquireCluster(t)
	// Start, stop, then delete a transaction
	txn := startTransaction(t, c)
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl stop transaction").Run())
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl delete transaction {{.txn}}", "txn", txn).Run())

	// Check that the transaction no longer exists
	requireTransactionDoesNotExist(t, c, txn)
}

func TestDeleteSpecificTransaction(t *testing.T) {
	t.Skip() // DNJ - tests have not run and need fixing
	c, _ := minikubetestenv.AcquireCluster(t)
	// Start two transactions, delete the first one, make sure the correct one is deleted
	txn1 := startTransaction(t, c)
	txn2 := startTransaction(t, c)

	// Check that both transactions exist
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl inspect transaction {{.txn}}", "txn", txn1).Run())
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl inspect transaction {{.txn}}", "txn", txn2).Run())

	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl delete transaction {{.txn}}", "txn", txn1).Run())

	// Check that only the second transaction exists
	requireTransactionDoesNotExist(t, c, txn1)
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl inspect transaction {{.txn}}", "txn", txn2).Run())
}
