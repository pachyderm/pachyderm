//go:build k8s

package cmds

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	// TODO(CORE-2123): check for real logs
	require.NoError(t, testutil.PachctlBashCmd(t, c, `
		pachctl logs2 | match "GetLogs dummy response"`,
	).Run())
}
