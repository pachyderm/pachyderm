//go:build k8s

package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
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

type upgradeFunc func(t *testing.T, ctx context.Context, c *client.APIClient, fromVersion string)

// runs the upgrade test from all versions specified in "fromVersions" against the local image
func upgradeTest(suite *testing.T, ctx context.Context, parallelOK bool, numPachds int, fromVersions []string, preUpgrade upgradeFunc, postUpgrade upgradeFunc) {
	k := testutil.GetKubeClient(suite)
	for _, from := range fromVersions {
		from := from // suite.Run runs in a background goroutine if parallelOK is true
		suite.Run(fmt.Sprintf("UpgradeFrom_%s", from), func(t *testing.T) {
			if parallelOK {
				t.Parallel()
			}
			ns, portOffset := minikubetestenv.ClaimCluster(t)
			t.Logf("starting preUpgrade; version %v, namespace %v", from, ns)
			valuesOverridden, strValuesOverridden := helmValuesPreGoCDK(numPachds)
			preUpgrade(t, ctx, minikubetestenv.InstallRelease(t,
				context.Background(),
				ns,
				k,
				&minikubetestenv.DeployOpts{
					Version:            from,
					DisableLoki:        true,
					PortOffset:         portOffset,
					UseLeftoverCluster: false,
					ValueOverrides:     valuesOverridden,
					ValuesStrOverrides: strValuesOverridden,
				}), from)
			t.Logf("preUpgrade done; starting postUpgrade")
			postUpgrade(t, ctx, minikubetestenv.UpgradeRelease(t,
				context.Background(),
				ns,
				k,
				&minikubetestenv.DeployOpts{
					PortOffset:     portOffset,
					CleanupAfter:   true,
					ValueOverrides: map[string]string{"pachw.minReplicas": "1", "pachw.maxReplicas": "5", "pachd.replicas": strconv.Itoa(numPachds)},
				}), from)
			t.Logf("postUpgrade done")
		})
	}
}

// helmValuesPreGoCDK returns two maps. The first is for overriding SetValues, the second is for
// overriding SetStrValues.
func helmValuesPreGoCDK(numPachds int) (map[string]string, map[string]string) {
	return map[string]string{
			"pachw.minReplicas": "1",
			"pachw.maxReplicas": "5",
			"pachd.replicas":    strconv.Itoa(numPachds),
			// We are using "old" minio values here to pass CI tests. Current configurations has enabled gocdk by default,
			// so to make UpgradeTest work, we overried configuration with these "old" minio values.
			"pachd.storage.gocdkEnabled":   "false",
			"pachd.storage.backend":        "MINIO",
			"pachd.storage.minio.bucket":   "pachyderm-test",
			"pachd.storage.minio.endpoint": "minio.default.svc.cluster.local:9000",
			"pachd.storage.minio.id":       "minioadmin",
			"pachd.storage.minio.secret":   "minioadmin",
		},
		map[string]string{
			"pachd.storage.minio.signature": "",
			"pachd.storage.minio.secure":    "false",
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
		"2.7.6",
		"2.8.5",
	}
	montage := func(fromVersion string) string {
		repo := montageRepo
		if fromVersion < "2.5.0" {
			// We use a long pipeline name (gt 64 chars) to test whether our auth tokens,
			// which originally had a 64 limit, can handle the upgrade which adds the project names to the subject key.
			repo += "01234567890123456789012345678901234567890"
		}
		return repo

	}
	upgradeTest(t, pctx.TestContext(t), true /* parallelOK */, 2, fromVersions,
		func(t *testing.T, ctx context.Context, c *client.APIClient, from string) { /* preUpgrade */
			c = testutil.AuthenticatedPachClient(t, c, upgradeSubject)
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, imagesRepo))
			require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
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
				c.CreatePipeline(pfs.DefaultProjectName,
					montage(from),
					"dpokidov/imagemagick:7.1.0-23",
					[]string{"sh"}, /* cmd */
					[]string{fmt.Sprintf("montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find -L /pfs/%s /pfs/%s -type f | sort) /pfs/out/montage.png", imagesRepo, edgesRepo)}, /* stdin */
					nil, /* parallelismSpec */
					&pps.Input{Cross: []*pps.Input{
						{Pfs: &pps.PFSInput{Repo: imagesRepo, Glob: "/"}},
						{Pfs: &pps.PFSInput{Repo: edgesRepo, Glob: "/"}},
					}},
					"master", /* outputBranch */
					false,    /* update */
				))

			require.NoError(t, c.WithModifyFileClient(client.NewCommit(pfs.DefaultProjectName, imagesRepo, "master", "" /* commitID */), func(mf client.ModifyFile) error {
				return errors.EnsureStack(mf.PutFileURL("/liberty.png", "https://docs.pachyderm.com/images/mldm/opencv/liberty.jpg", false))
			}))

			commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, montage(from), "master", "")
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(ctx, 4*time.Minute)
			defer cancel()
			t.Log("before upgrade: waiting for montage commit")
			commitInfos, err := c.WithCtx(ctx).WaitCommitSetAll(commitInfo.Commit.Id)
			t.Log("before upgrade: wait is done")
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Repo, client.NewRepo(pfs.DefaultProjectName, montage(from))) {
					require.NoError(t, c.GetFile(info.Commit, "montage.png", &buf))
				}
			}
		},
		func(t *testing.T, ctx context.Context, c *client.APIClient, from string) { /* postUpgrade */
			c = testutil.AuthenticateClient(t, c, upgradeSubject)
			state, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
			require.NoError(t, err)
			require.Equal(t, enterprise.State_ACTIVE, state.State)
			// check provenance migration
			commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, montage(from), "master", "")
			require.NoError(t, err)
			require.Equal(t, 3, len(commitInfo.DirectProvenance))
			for _, p := range commitInfo.DirectProvenance {
				if p.Repo.Type == pfs.SpecRepoType { // spec commit should be in a different commit set
					require.NotEqual(t, commitInfo.Commit.Id, p.Id, "commit %q with provenance %q", commitInfo.Commit.String(), p.String())
				} else {
					require.Equal(t, commitInfo.Commit.Id, p.Id, "commit %q with provenance %q", commitInfo.Commit.String(), p.String())
				}
			}
			// check DAG still works with new commits
			require.NoError(t, c.WithModifyFileClient(client.NewCommit(pfs.DefaultProjectName, imagesRepo, "master", ""), func(mf client.ModifyFile) error {
				return errors.EnsureStack(mf.PutFileURL("/kitten.png", "https://docs.pachyderm.com/images/mldm/opencv/kitten.jpg", false))
			}))
			commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, montage(from), "master", "")
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()
			t.Log("after upgrade: waiting for montage commit")
			commitInfos, err := c.WithCtx(ctx).WaitCommitSetAll(commitInfo.Commit.Id)
			t.Log("after upgrade: wait is done")
			require.NoError(t, err)
			var buf bytes.Buffer
			for _, info := range commitInfos {
				if proto.Equal(info.Commit.Repo, client.NewRepo(pfs.DefaultProjectName, montage(from))) {
					require.NoError(t, c.GetFile(info.Commit, "montage.png", &buf))
				}
			}
			require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
				if resp.Error != "" {
					return errors.Errorf(resp.Error)
				}
				return nil
			}))
		},
	)
}

func TestUpgradeMultiProjectJoins(t *testing.T) {
	if skip {
		t.Skip("Skipping upgrade test")
	}
	fromVersions := []string{"2.7.4", "2.8.1"}
	files := []string{"file1", "file2", "file3", "file4"}
	upgradeTest(t, pctx.TestContext(t), true /* parallelOK */, 1, fromVersions,
		func(t *testing.T, ctx context.Context, c *client.APIClient, _ string) { // preUpgrade
			c = testutil.AuthenticatedPachClient(t, c, upgradeSubject)
			// Create projects
			require.NoError(t, c.CreateProject("project1"))
			// Create repos
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo1"))
			// Create pipelines
			require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName, "pipeline1", "", []string{"bash"} /* command */, []string{"cp /pfs/in/* /pfs/out/"} /* stdin */, nil, &pps.Input{Pfs: &pps.PFSInput{Name: "in", Project: pfs.DefaultProjectName, Repo: "repo1", Glob: "/*"}}, "master", false))
			require.NoError(t, c.CreatePipeline(
				"project1",
				"pipeline1",
				"",
				[]string{"bash"},
				[]string{`
					for f in /pfs/in1/*; do
						cat $f /pfs/in2/$(basename $f) > /pfs/out/$(basename $f)
					done
				`},
				nil, /* parallelism spec */
				&pps.Input{Join: []*pps.Input{
					{Pfs: &pps.PFSInput{Name: "in1", Project: pfs.DefaultProjectName, Repo: "repo1", Glob: "/(*)", JoinOn: "$1"}},
					{Pfs: &pps.PFSInput{Name: "in2", Project: pfs.DefaultProjectName, Repo: "pipeline1", Glob: "/(*)", JoinOn: "$1"}},
				}},
				"master",
				false))
			require.NoError(t, c.CreatePipeline("project1", "pipeline2", "", []string{"bash"} /* command */, []string{"cp /pfs/in/* /pfs/out/"} /* stdin */, nil, &pps.Input{Pfs: &pps.PFSInput{Name: "in", Project: "project1", Repo: "pipeline1", Glob: "/*"}}, "master", false))
			require.NoError(t, c.CreatePipeline(
				"project1",
				"pipeline3",
				"",
				[]string{"bash"},
				[]string{`
					for f in /pfs/in1/*; do
						cat $f /pfs/in2/$(basename $f) > /pfs/out/$(basename $f)
					done
				`},
				nil, /* parallelism spec */
				&pps.Input{Join: []*pps.Input{
					{Pfs: &pps.PFSInput{Name: "in1", Project: "project1", Repo: "pipeline1", Glob: "/(*)", JoinOn: "$1"}},
					{Pfs: &pps.PFSInput{Name: "in2", Project: "project1", Repo: "pipeline2", Glob: "/(*)", JoinOn: "$1"}},
				}},
				"master",
				false))
			require.NoError(t, c.WithModifyFileClient(client.NewCommit(pfs.DefaultProjectName, "repo1", "master", ""), func(mf client.ModifyFile) error {
				for _, f := range files {
					if err := mf.PutFile("/"+f, strings.NewReader(f)); err != nil {
						return err
					}
				}
				return nil
			}))
			t.Log("before upgrade: waiting for commit")
			commitInfo, err := c.WaitCommit("project1", "pipeline3", "master", "")
			require.NoError(t, err)
			var buf bytes.Buffer
			for _, f := range files {
				require.NoError(t, c.GetFile(commitInfo.Commit, "/"+f, &buf))
				require.Equal(t, strings.Repeat(f, 4), buf.String()) // repeats 4 times because we concatenated twice
				buf.Reset()
			}
		},
		func(t *testing.T, ctx context.Context, c *client.APIClient, _ string) { // postUpgrade
			c = testutil.AuthenticatedPachClient(t, c, upgradeSubject)
			commitInfo, err := c.InspectCommit("project1", "pipeline3", "master", "")
			require.NoError(t, err)

			var buf bytes.Buffer
			for _, f := range files {
				require.NoError(t, c.GetFile(commitInfo.Commit, "/"+f, &buf))
				require.Equal(t, strings.Repeat(f, 4), buf.String()) // repeats 4 times because we concatenated twice
				buf.Reset()
			}
			require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
				if resp.Error != "" {
					return errors.Errorf(resp.Error)
				}
				return nil
			}))
		})
}

func TestUpgradeLoad(t *testing.T) {
	if skip {
		t.Skip("Skipping upgrade test")
	}
	fromVersions := []string{"2.7.2", "2.8.0"}
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
	upgradeTest(t, pctx.TestContext(t), false /* parallelOK */, 1, fromVersions,
		func(t *testing.T, ctx context.Context, c *client.APIClient, _ string) {
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
		func(t *testing.T, ctx context.Context, c *client.APIClient, _ string) {
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
			require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
				if resp.Error != "" {
					return errors.Errorf(resp.Error)
				}
				return nil
			}))
		},
	)
}
