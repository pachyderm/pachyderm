package cmds

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

// TestTransaction runs a straightforward end-to-end test of starting, adding
// to and finishing a transaction.
func TestTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	repo := tu.UniqueString("TestTransaction-repo")
	setup := tu.BashCmd(`
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
	require.YesError(t, tu.BashCmd("pachctl inspect repo {{.repo}}", "repo", repo).Run())

	// Check that finishing the transaction creates the requested objects
	require.NoError(t, tu.BashCmd(`
		pachctl finish transaction
		pachctl inspect repo {{.repo}}
		pachctl inspect commit {{.repo}}@master
		pachctl inspect commit {{.repo}}@{{.commit}}
		`,
		"repo", repo,
		"commit", commit,
	).Run())
}

func startTransaction(t *testing.T) string {
	output, err := (*exec.Cmd)(tu.BashCmd("pachctl start transaction")).Output()
	require.NoError(t, err)

	strs := strings.Split(strings.TrimSpace(string(output)), " ")
	return strs[len(strs)-1]
}

func requireTransactionDoesNotExist(t *testing.T, txn string) {
	output, err := (*exec.Cmd)(tu.BashCmd("pachctl inspect transaction {{.txn}} 2>&1", "txn", txn)).Output()
	require.YesError(t, err)
	expected := fmt.Sprintf("%s not found", txn)
	require.True(t, strings.Contains(string(output), expected))
}

func TestDeleteActiveTransaction(t *testing.T) {
	// Start then delete a transaction
	txn := startTransaction(t)
	require.NoError(t, tu.BashCmd("pachctl delete transaction").Run())

	// Check that the transaction no longer exists
	requireTransactionDoesNotExist(t, txn)
}

func TestDeleteInactiveTransaction(t *testing.T) {
	// Start, stop, then delete a transaction
	txn := startTransaction(t)
	require.NoError(t, tu.BashCmd("pachctl stop transaction").Run())
	require.NoError(t, tu.BashCmd("pachctl delete transaction {{.txn}}", "txn", txn).Run())

	// Check that the transaction no longer exists
	requireTransactionDoesNotExist(t, txn)
}

func TestDeleteSpecificTransaction(t *testing.T) {
	// Start two transactions, delete the first one, make sure the correct one is deleted
	txn1 := startTransaction(t)
	txn2 := startTransaction(t)

	// Check that both transactions exist
	require.NoError(t, tu.BashCmd("pachctl inspect transaction {{.txn}}", "txn", txn1).Run())
	require.NoError(t, tu.BashCmd("pachctl inspect transaction {{.txn}}", "txn", txn2).Run())

	require.NoError(t, tu.BashCmd("pachctl delete transaction {{.txn}}", "txn", txn1).Run())

	// Check that only the second transaction exists
	requireTransactionDoesNotExist(t, txn1)
	require.NoError(t, tu.BashCmd("pachctl inspect transaction {{.txn}}", "txn", txn2).Run())
}
