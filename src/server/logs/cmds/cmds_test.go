package cmds

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
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

func realEnvWithLoki(ctx context.Context, t testing.TB, entries []loki.Entry) *realenv.RealEnv {
	srv := httptest.NewServer(&lokiutil.FakeServer{
		Entries: entries,
	})
	t.Cleanup(func() { srv.Close() })
	return realenv.NewRealEnvWithIdentity(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption,
		func(c *pachconfig.Configuration) {
			u, err := url.Parse(srv.URL)
			if err != nil {
				panic(err)
			}
			c.LokiHost, c.LokiPort = u.Hostname(), u.Port()
		})
}

func TestGetLogs_noauth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	var (
		ctx          = pctx.TestContext(t)
		buildEntries = func() []loki.Entry {
			var entries []loki.Entry
			for i := -99; i <= 0; i++ {
				entries = append(entries, loki.Entry{
					Timestamp: time.Now().Add(time.Duration(i) * time.Second),
					Line:      fmt.Sprintf("%v foo", i),
				})
			}
			return entries
		}
		env = realEnvWithLoki(ctx, t, buildEntries())
		c   = env.PachClient
	)
	mockInspectCluster(env)

	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl logs2 | match "99 foo"`,
	).Run())
}

func TestGetLogs_nonadmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	var (
		ctx          = pctx.TestContext(t)
		buildEntries = func() []loki.Entry {
			var entries []loki.Entry
			for i := -99; i <= 0; i++ {
				entries = append(entries, loki.Entry{
					Timestamp: time.Now().Add(time.Duration(i) * time.Second),
					Line:      fmt.Sprintf("%v bar", i),
				})
			}
			return entries
		}
		env = realEnvWithLoki(ctx, t, buildEntries())
		c   = env.PachClient
	)
	mockInspectCluster(env)
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	alice := testutil.UniqueString("robot:alice")
	aliceClient := testutil.AuthenticatedPachClient(t, c, alice, peerPort)

	require.NoError(t, testutil.PachctlBashCmdCtx(aliceClient.Ctx(), t, aliceClient, `
		pachctl logs2 | match "98 bar"`,
	).Run())
}

func TestGetLogs_admin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	var (
		ctx          = pctx.TestContext(t)
		buildEntries = func() []loki.Entry {
			var entries []loki.Entry
			for i := -99; i <= 0; i++ {
				entries = append(entries, loki.Entry{
					Timestamp: time.Now().Add(time.Duration(i) * time.Second),
					Line:      fmt.Sprintf("%v baz", i),
				})
			}
			return entries
		}
		env = realEnvWithLoki(ctx, t, buildEntries())
		c   = env.PachClient
	)
	mockInspectCluster(env)
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	adminClient := testutil.AuthenticatedPachClient(t, c, auth.RootUser, peerPort)

	require.NoError(t, testutil.PachctlBashCmdCtx(adminClient.Ctx(), t, adminClient, `
		pachctl logs2 | match "12 baz"`,
	).Run())
}
