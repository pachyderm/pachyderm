package server_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/logs"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
)

func TestEmptyRequest(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	logsClient := logs.NewAPIClient(env.PachClient.ClientConn())
	respStream, err := logsClient.GetLogs(ctx, &logs.GetLogsRequest{})
	require.NoError(t, err, "logs.GetLogs request error")
	resp, err := respStream.Recv()
	require.NoError(t, err, "logs.GetLogs stream error")
	require.True(t, resp.GetLog().GetPpsLogMessage().Message == "GetLogs dummy response", "unexpected message in response")
}
