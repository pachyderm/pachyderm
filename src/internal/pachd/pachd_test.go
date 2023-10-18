package pachd

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestFull(t *testing.T) {
	ctx := pctx.TestContext(t)
	pc := NewTestPachd(t)
	res, err := pc.VersionAPIClient.GetVersion(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	t.Log(res)
}
