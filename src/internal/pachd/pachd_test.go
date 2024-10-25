package pachd

import (
	"testing"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/pjs"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestFull(t *testing.T) {
	ctx := pctx.TestContext(t)
	pc := NewTestPachd(t)
	res, err := pc.VersionAPIClient.GetVersion(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	t.Log(res)
}

func TestPJSWorkerAuth(t *testing.T) {
	ctx := pctx.TestContext(t)
	pc := NewTestPachd(t, PJSWorkerAuthOption(auth.HashToken("iampjs")))
	pc.SetAuthToken(auth.HashToken("iampjs"))
	_, err := pc.ListQueue(ctx, &pjs.ListQueueRequest{})
	require.NoError(t, err)
}
