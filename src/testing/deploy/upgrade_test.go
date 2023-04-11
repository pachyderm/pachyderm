//go:build k8s

package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
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
func upgradeTest(suite *testing.T, ctx context.Context, parallelOK bool, fromVersions []string, preUpgrade func(*testing.T, *client.APIClient), postUpgrade func(*testing.T, *client.APIClient)) {
	k := testutil.GetKubeClient(suite)
	for _, from := range fromVersions {
		from := from // suite.Run runs in a background goroutine if parallelOK is true
		suite.Run(fmt.Sprintf("UpgradeFrom_%s", from), func(t *testing.T) {
			if parallelOK {
				t.Parallel()
			}
			ns, portOffset := minikubetestenv.ClaimCluster(t)
			minikubetestenv.PutNamespace(t, ns)
			t.Logf("starting preUpgrade; version %v, namespace %v", from, ns)
			preUpgrade(t, minikubetestenv.InstallRelease(t,
				context.Background(),
				ns,
				k,
				&minikubetestenv.DeployOpts{
					Version:     from,
					DisableLoki: true,
					PortOffset:  portOffset,
				}))
			t.Logf("preUpgrade done; starting postUpgrade")
			postUpgrade(t, minikubetestenv.UpgradeRelease(t,
				context.Background(),
				ns,
				k,
				&minikubetestenv.DeployOpts{
					PortOffset: portOffset,
				}))
			t.Logf("postUpgrade done")
		})
	}
}

func TestUpgradeTrigger(t *testing.T) {
	if skip {
		t.Skip("Skipping upgrade test")
	}
	fromVersions := []string{
		"2.4.6",
		"2.5.0",
	}
	dataRepo := "TestTrigger_data"
	dataCommit := client.NewProjectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	upgradeTest(t, context.Background(), false /* parallelOK */, fromVersions,
		func(t *testing.T, c *client.APIClient) { /* preUpgrade */
			require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName, dataRepo))
			pipeline1 := "TestTrigger1"
			pipeline2 := "TestTrigger2"
			require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
				pipeline1,
				"",
				[]string{"bash"},
				[]string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
				&pps.ParallelismSpec{
					Constant: 1,
				},
				client.NewProjectPFSInputOpts(dataRepo, pfs.DefaultProjectName, dataRepo, "trigger", "/*", "", "", false, false, &pfs.Trigger{
					Branch:  "master",
					Commits: 1,
				}),
				"",
				false,
			))
			require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
				pipeline2,
				"",
				[]string{"bash"},
				[]string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", pipeline1),
				},
				&pps.ParallelismSpec{
					Constant: 1,
				},
				client.NewProjectPFSInputOpts(pipeline1, pfs.DefaultProjectName, pipeline1, "", "/*", "", "", false, false, &pfs.Trigger{
					Commits: 2,
				}),
				"",
				false,
			))
			for i := 0; i < 10; i++ {
				require.NoError(t, c.PutFile(dataCommit, "/hello", strings.NewReader("hello world")))
			}
			ci, err := c.InspectProjectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
			require.NoError(t, err)
			_, err = c.WaitCommitSetAll(ci.Commit.ID)
			require.NoError(t, err)
		},
		func(t *testing.T, c *client.APIClient) { /* postUpgrade */
			for i := 0; i < 10; i++ {
				require.NoError(t, c.PutFile(dataCommit, "/hello", strings.NewReader("hello world")))
			}
			latestDataCI, err := c.InspectProjectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
			require.NoError(t, err)
			require.NoErrorWithinTRetry(t, 2*time.Minute, func() error {
				ci, err := c.InspectProjectCommit(pfs.DefaultProjectName, "TestTrigger2", "master", "")
				require.NoError(t, err)
				aliasCI, err := c.InspectProjectCommit(pfs.DefaultProjectName, dataRepo, "", ci.Commit.ID)
				require.NoError(t, err)
				if aliasCI.Commit.ID != latestDataCI.Commit.ID {
					return errors.New("not ready")
				}
				return nil
			})
			require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
				if resp.Error != "" {
					return errors.Errorf(resp.Error)
				}
				return nil
			}))
		},
	)
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
		"2.3.9",
		"2.4.6",
	}
	// We use a long pipeline name (gt 64 chars) to test whether our auth tokens,
	// which originally had a 64 limit, can handle the upgrade which adds the project names to the subject key.
	montage := montageRepo + "01234567890123456789012345678901234567890"
	upgradeTest(t, context.Background(), true /* parallelOK */, fromVersions,
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
			t.Log("before upgrade: waiting for montage commit")
			commitInfos, err := c.WithCtx(ctx).WaitCommitSetAll(commitInfo.Commit.ID)
			t.Log("before upgrade: wait is done")
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Repo, client.NewProjectRepo(pfs.DefaultProjectName, montage)) {
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
			t.Log("after upgrade: waiting for montage commit")
			commitInfos, err := c.WithCtx(ctx).WaitCommitSetAll(commitInfo.Commit.ID)
			t.Log("after upgrade: wait is done")
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Repo, client.NewProjectRepo(pfs.DefaultProjectName, montage)) {
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
	fromVersions := []string{"2.3.9", "2.4.6"}
	dagSpec := `
default-load-test-source-1:
default-load-test-pipeline-1: default-load-test-source-1
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
	upgradeTest(t, context.Background(), false /* parallelOK */, fromVersions,
		func(t *testing.T, c *client.APIClient) {
			c = testutil.AuthenticatedPachClient(t, c, upgradeSubject)
			t.Log("before upgrade: starting load test")
			resp, err := c.PpsAPIClient.RunLoadTest(c.Ctx(), &pps.RunLoadTestRequest{
				DagSpec:  dagSpec,
				LoadSpec: loadSpec,
			})
			t.Log("before upgrade: load test done")
			require.NoError(t, err)
			require.Equal(t, "", resp.Error)
			stateID = resp.StateId
		},
		func(t *testing.T, c *client.APIClient) {
			c = testutil.AuthenticateClient(t, c, upgradeSubject)
			t.Log("after upgrade: starting load test")
			resp, err := c.PpsAPIClient.RunLoadTest(c.Ctx(), &pps.RunLoadTestRequest{
				DagSpec:  dagSpec,
				LoadSpec: loadSpec,
				StateId:  stateID,
			})
			t.Log("after upgrade: load test done")
			require.NoError(t, err)
			require.Equal(t, "", resp.Error)
		},
	)
}
