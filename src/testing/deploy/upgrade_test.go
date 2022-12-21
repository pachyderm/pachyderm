//go:build k8s

package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	proto "github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	inputRepo      string = "input"
	outputRepo     string = "output"
	upgradeSubject string = "upgrade_client"
)

var skip bool

// runs the upgrade test from all versions specified in "fromVersions" against the local image
func upgradeTest(suite *testing.T, ctx context.Context, fromVersions []string, preUpgrade func(*testing.T, *client.APIClient), postUpgrade func(*testing.T, *client.APIClient)) {
	k := testutil.GetKubeClient(suite)
	for _, from := range fromVersions {
		suite.Run(fmt.Sprintf("UpgradeFrom_%s", from), func(t *testing.T) {
			t.Parallel()
			ns, portOffset := minikubetestenv.ClaimCluster(t)
			minikubetestenv.PutNamespace(t, ns)
			preUpgrade(t, minikubetestenv.InstallRelease(t,
				context.Background(),
				ns,
				k,
				&minikubetestenv.DeployOpts{
					Version:     from,
					DisableLoki: true,
					PortOffset:  portOffset,
					// For 2.3 -> future upgrades, we'll want to delete these
					// overrides.  They became the default (instead of random)
					// in the 2.3 alpha cycle.
					ValueOverrides: map[string]string{
						"global.postgresql.postgresqlPassword":         "insecure-user-password",
						"global.postgresql.postgresqlPostgresPassword": "insecure-root-password",
					},
				}))
			postUpgrade(t, minikubetestenv.UpgradeRelease(t,
				context.Background(),
				ns,
				k,
				&minikubetestenv.DeployOpts{
					WaitSeconds:  10,
					CleanupAfter: true,
					PortOffset:   portOffset,
				}))
		})
	}
}

// pre-upgrade:
// - create a pipeline "output" that copies contents from repo "input"
// - create file input@master:/foo
// post-upgrade:
// - create file input@master:/bar
// - verify output@master:/bar and output@master:/foo still exists
func TestUpgradeSimple(t *testing.T) {
	if skip {
		t.Skip("Skipping upgrade test")
	}
	fromVersions := []string{
		"2.0.4",
		"2.1.0",
		"2.2.0",
		"2.3.9",
	}
	upgradeTest(t, context.Background(), fromVersions,
		func(t *testing.T, c *client.APIClient) {
			c = testutil.AuthenticatedPachClient(t, c, upgradeSubject)
			require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName, inputRepo))
			require.NoError(t,
				c.CreateProjectPipeline(pfs.DefaultProjectName, outputRepo,
					"busybox",
					[]string{"sh"},
					[]string{"cp /pfs/input/* /pfs/out/;"},
					nil,
					&pps.Input{Pfs: &pps.PFSInput{Glob: "/", Repo: inputRepo}},
					"master",
					false,
				))
			require.NoError(t, c.WithModifyFileClient(client.NewProjectCommit(pfs.DefaultProjectName, inputRepo, "master", ""), func(mf client.ModifyFile) error {
				return errors.EnsureStack(mf.PutFile("foo", strings.NewReader("foo")))
			}))

			commitInfo, err := c.InspectProjectCommit(pfs.DefaultProjectName, outputRepo, "master", "")
			require.NoError(t, err)
			commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.ID)
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Branch.Repo, client.NewProjectRepo(pfs.DefaultProjectName, outputRepo)) {
					require.NoError(t, c.GetFile(info.Commit, "foo", &buf))
					require.Equal(t, "foo", buf.String())
				}
			}
		},

		func(t *testing.T, c *client.APIClient) {
			c = testutil.AuthenticateClient(t, c, upgradeSubject)
			state, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
			require.NoError(t, err)
			require.Equal(t, enterprise.State_ACTIVE, state.State)
			require.NoError(t, c.WithModifyFileClient(client.NewProjectCommit(pfs.DefaultProjectName, inputRepo, "master", ""), func(mf client.ModifyFile) error {
				return errors.EnsureStack(mf.PutFile("bar", strings.NewReader("bar")))
			}))

			commitInfo, err := c.InspectProjectCommit(pfs.DefaultProjectName, outputRepo, "master", "")
			require.NoError(t, err)
			commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.ID)
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Branch.Repo, client.NewProjectRepo(pfs.DefaultProjectName, outputRepo)) {
					require.NoError(t, c.GetFile(info.Commit, "foo", &buf))
					require.Equal(t, "foo", buf.String())

					buf.Reset()

					require.NoError(t, c.GetFile(info.Commit, "bar", &buf))
					require.Equal(t, "bar", buf.String())
				}
			}
		},
	)
}

func TestUpgradeLoad(t *testing.T) {
	if skip {
		t.Skip("Skipping upgrade test")
	}
	fromVersions := []string{"2.3.9"}
	dagSpec := `
default-load-test-source:
default-load-test-pipeline: default-load-test-source
`
	loadSpec := `
count: 5
modifications:
  - count: 5
    putFile:
      count: 5
      source: "random"
fileSources:
  - name: "random"
    random:
      sizes:
        - min: 1000
          max: 10000
          prob: 30
        - min: 10000
          max: 100000
          prob: 40
        - min: 1000000
          max: 10000000
          prob: 30
validator:
  frequency:
    count: 1
`
	var stateID string
	upgradeTest(t, context.Background(), fromVersions,
		func(t *testing.T, c *client.APIClient) {
			c = testutil.AuthenticatedPachClient(t, c, upgradeSubject)
			resp, err := c.PpsAPIClient.RunLoadTest(c.Ctx(), &pps.RunLoadTestRequest{
				DagSpec:  dagSpec,
				LoadSpec: loadSpec,
			})
			require.NoError(t, err)
			require.Equal(t, "", resp.Error)
			stateID = resp.StateId
		},
		func(t *testing.T, c *client.APIClient) {
			c = testutil.AuthenticateClient(t, c, upgradeSubject)
			resp, err := c.PpsAPIClient.RunLoadTest(c.Ctx(), &pps.RunLoadTestRequest{
				DagSpec:  dagSpec,
				LoadSpec: loadSpec,
				StateId:  stateID,
			})
			require.NoError(t, err)
			require.Equal(t, "", resp.Error)
		},
	)
}
