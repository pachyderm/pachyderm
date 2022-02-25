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
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	inputRepo      string = "input"
	outputRepo     string = "output"
	upgradeSubject string = "upgrade_client"
)

var (
	fromVersions = []string{
		"2.0.4",
		"2.0.5",
		"2.1.0",
	}
)

// runs the upgrade test from all versions specified in "fromVersions" against the local image
func upgradeTest(suite *testing.T, ctx context.Context, preUpgrade func(*testing.T, *client.APIClient), postUpgrade func(*testing.T, *client.APIClient)) {
	k := testutil.GetKubeClient(suite)
	for _, from := range fromVersions {
		suite.Run(fmt.Sprintf("UpgradeFrom_%s", from), func(t *testing.T) {
			preUpgrade(t, minikubetestenv.InstallRelease(t,
				context.Background(),
				k,
				&minikubetestenv.DeployOpts{
					AuthUser: upgradeSubject,
					Version:  from,
				}))
			postUpgrade(t, minikubetestenv.UpgradeRelease(t,
				context.Background(),
				k,
				&minikubetestenv.DeployOpts{
					AuthUser:     upgradeSubject,
					CleanupAfter: true,
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
	upgradeTest(t, context.Background(),
		func(t *testing.T, c *client.APIClient) {
			require.NoError(t, c.CreateRepo(inputRepo))
			require.NoError(t,
				c.CreatePipeline(outputRepo,
					"busybox",
					[]string{"sh"},
					[]string{"cp /pfs/input/* /pfs/out/;"},
					nil,
					&pps.Input{Pfs: &pps.PFSInput{Glob: "/", Repo: inputRepo}},
					"master",
					false,
				))
			require.NoError(t, c.WithModifyFileClient(client.NewCommit(inputRepo, "master", ""), func(mf client.ModifyFile) error {
				return errors.EnsureStack(mf.PutFile("foo", strings.NewReader("foo")))
			}))

			commitInfo, err := c.InspectCommit(outputRepo, "master", "")
			require.NoError(t, err)
			commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.ID)
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Branch.Repo, client.NewRepo(outputRepo)) {
					require.NoError(t, c.GetFile(info.Commit, "foo", &buf))
					require.Equal(t, "foo", buf.String())
				}
			}
		},

		func(t *testing.T, c *client.APIClient) {
			state, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
			require.NoError(t, err)
			require.Equal(t, enterprise.State_ACTIVE, state.State)
			require.NoError(t, c.WithModifyFileClient(client.NewCommit(inputRepo, "master", ""), func(mf client.ModifyFile) error {
				return errors.EnsureStack(mf.PutFile("bar", strings.NewReader("bar")))
			}))

			commitInfo, err := c.InspectCommit(outputRepo, "master", "")
			require.NoError(t, err)
			commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.ID)
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Branch.Repo, client.NewRepo(outputRepo)) {
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
