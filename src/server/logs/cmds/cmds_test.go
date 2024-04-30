package cmds

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

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

func TestGetLogs_default_noauth(t *testing.T) {
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

	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl logs2 | match "99 foo"`,
	).Run())
}

// TODO: need to replace this test with real loki that can filter by data for specific pipelines and non-admin users
// func TestGetLogs_default_nonadmin(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}

// 	var (
// 		ctx          = pctx.TestContext(t)
// 		buildEntries = func() []loki.Entry {
// 			var entries []loki.Entry
// 			for i := -99; i <= 0; i++ {
// 				entries = append(entries, loki.Entry{
// 					Timestamp: time.Now().Add(time.Duration(i) * time.Second),
// 					Line:      fmt.Sprintf("%v bar", i),
// 				})
// 			}
// 			return entries
// 		}
// 		env = realEnvWithLoki(ctx, t, buildEntries())
// 		c   = env.PachClient
// 	)
// 	mockInspectCluster(env)
// 	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
// 	alice := testutil.UniqueString("robot:alice")
// 	aliceClient := testutil.AuthenticatedPachClient(t, c, alice, peerPort)

// 	require.NoError(t, testutil.PachctlBashCmdCtx(aliceClient.Ctx(), t, aliceClient, `
// 		pachctl logs2 | match "98 bar"`,
// 	).Run())
// }

func TestGetLogs_default_admin(t *testing.T) {
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
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	adminClient := testutil.AuthenticatedPachClient(t, c, auth.RootUser, peerPort)

	require.NoError(t, testutil.PachctlBashCmdCtx(adminClient.Ctx(), t, adminClient, `
		pachctl logs2 | match "12 baz"`,
	).Run())
}

func TestGetLogs_pipeline_noauth(t *testing.T) {
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
					Line:      fmt.Sprintf("%v quux", i),
				})
			}
			return entries
		}
		env = realEnvWithLoki(ctx, t, buildEntries())
		c   = env.PachClient
	)

	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create repo {{.RepoName}}
		pachctl create pipeline <<EOF
		{
			"pipeline": {
				"project": {
					"name": "{{.ProjectName | js}}"
				},
				"name": "{{.PipelineName | js}}"
			},
			"transform": {
				"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
			},
			"input": {
				"pfs": {
					"project": "default",
					"repo": "{{.RepoName | js}}",
					"glob": "/*",
					"name": "in"
				}
			},
			"resource_requests": {
				"cpu": null,
				"disk": "187Mi"
			},
			"autoscaling": false
		}
		EOF
		pachctl logs2 | match "24 quux"`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline",
	).Run())
}

// TODO: need to replace this test with real loki that can filter by data for specific pipelines and non-admin users
// func TestGetLogs_pipeline_user(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}

// 	var (
// 		ctx          = pctx.TestContext(t)
// 		buildEntries = func() []loki.Entry {
// 			var entries []loki.Entry
// 			for i := -99; i <= 0; i++ {
// 				entries = append(entries, loki.Entry{
// 					Timestamp: time.Now().Add(time.Duration(i) * time.Second),
// 					Line:      fmt.Sprintf("%v foo", i),
// 				})
// 			}
// 			return entries
// 		}
// 		env = realEnvWithLoki(ctx, t, buildEntries())
// 		c   = env.PachClient
// 	)
// 	mockInspectCluster(env)
// 	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
// 	alice := testutil.UniqueString("robot:alice")
// 	aliceClient := testutil.AuthenticatedPachClient(t, c, alice, peerPort)

// 	require.NoError(t, testutil.PachctlBashCmdCtx(aliceClient.Ctx(), t, aliceClient, `
// 		pachctl create repo {{.RepoName}}
// 		pachctl create pipeline <<EOF
// 		{
// 			"pipeline": {
// 				"project": {
// 					"name": "{{.ProjectName | js}}"
// 				},
// 				"name": "{{.PipelineName | js}}"
// 			},
// 			"transform": {
// 				"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
// 			},
// 			"input": {
// 				"pfs": {
// 					"project": "default",
// 					"repo": "{{.RepoName | js}}",
// 					"glob": "/*",
// 					"name": "in"
// 				}
// 			},
// 			"resource_requests": {
// 				"cpu": null,
// 				"disk": "187Mi"
// 			},
// 			"autoscaling": false
// 		}
// 		EOF
// 		pachctl logs2 --pipeline pipeline --project default| match "64 foo"`,
// 		"ProjectName", pfs.DefaultProjectName,
// 		"RepoName", "input",
// 		"PipelineName", "pipeline",
// 	).Run())
// }

func TestGetLogs_project_noauth(t *testing.T) {
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
					Line:      fmt.Sprintf("%v %s", i, t.Name()),
				})
			}
			return entries
		}
		env = realEnvWithLoki(ctx, t, buildEntries())
		c   = env.PachClient
	)

	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create repo {{.RepoName}}
		pachctl create pipeline <<EOF
		{
			"pipeline": {
				"project": {
					"name": "{{.ProjectName | js}}"
				},
				"name": "{{.PipelineName | js}}"
			},
			"transform": {
				"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
			},
			"input": {
				"pfs": {
					"project": "default",
					"repo": "{{.RepoName | js}}",
					"glob": "/*",
					"name": "in"
				}
			},
			"resource_requests": {
				"cpu": null,
				"disk": "187Mi"
			},
			"autoscaling": false
		}
		EOF
		pachctl logs2 | match "16 {{.TestName}}"`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline",
		"TestName", t.Name(),
	).Run())
}

// TODO: need to replace this test with real loki that can filter by data for specific pipelines and non-admin users
// func TestGetLogs_project_user(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}

// 	var (
// 		ctx          = pctx.TestContext(t)
// 		buildEntries = func() []loki.Entry {
// 			var entries []loki.Entry
// 			for i := -99; i <= 0; i++ {
// 				entries = append(entries, loki.Entry{
// 					Timestamp: time.Now().Add(time.Duration(i) * time.Second),
// 					Line:      fmt.Sprintf("%v %s", i, t.Name()),
// 				})
// 			}
// 			return entries
// 		}
// 		env = realEnvWithLoki(ctx, t, buildEntries())
// 		c   = env.PachClient
// 	)
// 	mockInspectCluster(env)
// 	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
// 	alice := testutil.UniqueString("robot:alice")
// 	aliceClient := testutil.AuthenticatedPachClient(t, c, alice, peerPort)

// 	require.NoError(t, testutil.PachctlBashCmdCtx(aliceClient.Ctx(), t, aliceClient, `
// 		pachctl create repo {{.RepoName}}
// 		pachctl create pipeline <<EOF
// 		{
// 			"pipeline": {
// 				"project": {
// 					"name": "{{.ProjectName | js}}"
// 				},
// 				"name": "{{.PipelineName | js}}"
// 			},
// 			"transform": {
// 				"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
// 			},
// 			"input": {
// 				"pfs": {
// 					"project": "default",
// 					"repo": "{{.RepoName | js}}",
// 					"glob": "/*",
// 					"name": "in"
// 				}
// 			},
// 			"resource_requests": {
// 				"cpu": null,
// 				"disk": "187Mi"
// 			},
// 			"autoscaling": false
// 		}
// 		EOF
// 		pachctl logs2 | match "8 {{.TestName}}"`,
// 		"ProjectName", pfs.DefaultProjectName,
// 		"RepoName", "input",
// 		"PipelineName", "pipeline",
// 		"TestName", t.Name(),
// 	).Run())
// }

func TestGetLogs_combination_error(t *testing.T) {
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
					Line:      fmt.Sprintf("%v %s", i, t.Name()),
				})
			}
			return entries
		}
		env = realEnvWithLoki(ctx, t, buildEntries())
		c   = env.PachClient
	)
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	alice := testutil.UniqueString("robot:alice")
	aliceClient := testutil.AuthenticatedPachClient(t, c, alice, peerPort)

	require.YesError(t, testutil.PachctlBashCmdCtx(aliceClient.Ctx(), t, aliceClient, `
		pachctl create repo {{.RepoName}}
		pachctl create pipeline <<EOF
		{
			"pipeline": {
				"project": {
					"name": "{{.ProjectName | js}}"
				},
				"name": "{{.PipelineName | js}}"
			},
			"transform": {
				"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
			},
			"input": {
				"pfs": {
					"project": "default",
					"repo": "{{.RepoName | js}}",
					"glob": "/*",
					"name": "in"
				}
			},
			"resource_requests": {
				"cpu": null,
				"disk": "187Mi"
			},
			"autoscaling": false
		}
		EOF
		pachctl logs2 --logql '{}' --pipeline {{.PipelineName}} | match "1 {{.TestName}}"`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline",
		"TestName", t.Name(),
	).Run())

	require.YesError(t, testutil.PachctlBashCmdCtx(aliceClient.Ctx(), t, aliceClient, `
		pachctl logs2 --logql '{}' --project {{.ProjectName}}} | match "16 {{.TestName}}"`,
		"ProjectName", pfs.DefaultProjectName,
		"TestName", t.Name(),
	).Run())
}

func TestGetLogs_from_to(t *testing.T) {
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

	to := time.Now().Add(-time.Second)
	from := to.Add(-time.Hour)
	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl logs2 --from {{.From}} --to {{.To}}| match "99 foo"`,
		"From", from.Format(time.RFC3339Nano),
		"To", to.Format(time.RFC3339Nano),
	).Run())
}

func TestGetLogs_limit(t *testing.T) {
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

	to := time.Now().Add(-time.Second)
	from := to.Add(-time.Hour)
	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl logs2 --from {{.From}} --to {{.To}} --limit 1| match "99 foo"`,
		"From", from.Format(time.RFC3339Nano),
		"To", to.Format(time.RFC3339Nano),
	).Run())
	require.YesError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl logs2 --from {{.From}} --to {{.To}} --limit 1| match "98 foo"`,
		"From", from.Format(time.RFC3339Nano),
		"To", to.Format(time.RFC3339Nano),
	).Run())
}
