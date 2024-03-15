package cmds

import (
	"context"
	"strconv"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func mockInspectCluster(env *realenv.RealEnv) {
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		clusterInfo := admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}
		return &clusterInfo, nil
	})
}

func TestGetLogs_noauth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnvWithIdentity(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	c := env.PachClient
	mockInspectCluster(env)

	// TODO(CORE-2123): check for real logs
	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl logs2 | match "GetLogs dummy response"`,
	).Run())
}

func TestGetLogs_nonadmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnvWithIdentity(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	c := env.PachClient
	mockInspectCluster(env)
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	alice := testutil.UniqueString("robot:alice")
	aliceClient := testutil.AuthenticatedPachClient(t, c, alice, peerPort)

	// TODO(CORE-2123): check for real logs
	require.NoError(t, testutil.PachctlBashCmdCtx(aliceClient.Ctx(), t, aliceClient, `
		pachctl logs2 | match "GetLogs dummy response"`,
	).Run())
}

func TestGetLogs_admin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnvWithIdentity(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	c := env.PachClient
	mockInspectCluster(env)
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	adminClient := testutil.AuthenticatedPachClient(t, c, auth.RootUser, peerPort)

	// TODO(CORE-2123): check for real logs
	require.NoError(t, testutil.PachctlBashCmdCtx(adminClient.Ctx(), t, adminClient, `
		pachctl logs2 | match "GetLogs dummy response"`,
	).Run())
}
