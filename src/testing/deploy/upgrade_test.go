//go:build k8s

package main

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

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
	imagesRepo     = "images"
	edgesRepo      = "edges"
	montageRepo    = "montage"
	upgradeSubject = "upgrade_client"
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
func TestUpgradeOpenCVWithAuth(t *testing.T) {
	if skip {
		t.Skip("Skipping upgrade test")
	}
	fromVersions := []string{
		"2.0.4",
		"2.1.0",
		"2.2.0",
		"2.3.9",
		"2.4.3",
	}
	// We use a long pipeline name (gt 64 chars) to test whether our auth tokens,
	// which originally had a 64 limit, can handle the upgrade which adds the project names to the subject key.
	montage := montageRepo + "01234567890123456789012345678901234567890"
	upgradeTest(t, context.Background(), fromVersions,
		func(t *testing.T, c *client.APIClient) { /* preUpgrade */
			c = testutil.AuthenticatedPachClient(t, c, upgradeSubject)
			require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName, imagesRepo))
			require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
				edgesRepo,
				"pachyderm/opencv:1.0",
				[]string{"python3", "/edges.py"}, /* cmd */
				nil,                              /* stdin */
				nil,                              /* parallelismSpec */
				&pps.Input{Pfs: &pps.PFSInput{Glob: "/", Repo: imagesRepo}},
				"master", /* outputBranch */
				false,    /* update */
			))
			require.NoError(t,
				c.CreateProjectPipeline(pfs.DefaultProjectName,
					montage,
					"dpokidov/imagemagick:7.1.0-23",
					[]string{"sh"}, /* cmd */
					[]string{"montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png"}, /* stdin */
					nil, /* parallelismSpec */
					&pps.Input{Cross: []*pps.Input{
						{Pfs: &pps.PFSInput{Repo: imagesRepo, Glob: "/"}},
						{Pfs: &pps.PFSInput{Repo: edgesRepo, Glob: "/"}},
					}},
					"master", /* outputBranch */
					false,    /* update */
				))

			require.NoError(t, c.WithModifyFileClient(client.NewProjectCommit(pfs.DefaultProjectName, imagesRepo, "master", "" /* commitID */), func(mf client.ModifyFile) error {
				return errors.EnsureStack(mf.PutFileURL("/liberty.png", "http://i.imgur.com/46Q8nDz.png", false))
			}))

			commitInfo, err := c.InspectProjectCommit(pfs.DefaultProjectName, montage, "master", "")
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			commitInfos, err := c.WithCtx(ctx).WaitCommitSetAll(commitInfo.Commit.ID)
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Branch.Repo, client.NewProjectRepo(pfs.DefaultProjectName, montage)) {
					require.NoError(t, c.GetFile(info.Commit, "montage.png", &buf))
				}
			}
		},
		func(t *testing.T, c *client.APIClient) { /* postUpgrade */
			c = testutil.AuthenticateClient(t, c, upgradeSubject)
			state, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
			require.NoError(t, err)
			require.Equal(t, enterprise.State_ACTIVE, state.State)
			require.NoError(t, c.WithModifyFileClient(client.NewProjectCommit(pfs.DefaultProjectName, imagesRepo, "master", ""), func(mf client.ModifyFile) error {
				return errors.EnsureStack(mf.PutFileURL("/kitten.png", "https://i.imgur.com/g2QnNqa.png", false))
			}))

			commitInfo, err := c.InspectProjectCommit(pfs.DefaultProjectName, montage, "master", "")
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			commitInfos, err := c.WithCtx(ctx).WaitCommitSetAll(commitInfo.Commit.ID)
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Branch.Repo, client.NewProjectRepo(pfs.DefaultProjectName, montage)) {
					require.NoError(t, c.GetFile(info.Commit, "montage.png", &buf))
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
