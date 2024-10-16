package cmds

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestListSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	if err := tu.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create snapshot
		pachctl create snapshot
		pachctl create snapshot
		pachctl list snapshot
		`,
	).Run(); err != nil {
		t.Fatalf("list snapshot RPC: %v", err)
	}
}

// there are no tests for delete. because delete gRPC implementation has not been
// merged into master
