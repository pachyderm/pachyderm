package cmds

import (
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
