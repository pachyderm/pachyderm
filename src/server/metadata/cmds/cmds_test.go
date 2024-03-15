package cmds

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestEditMetadata(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl edit metadata && exit 1 || exit 0
`).Run())
}
