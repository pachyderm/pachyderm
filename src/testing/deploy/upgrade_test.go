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

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"golang.org/x/mod/semver"
	"google.golang.org/protobuf/proto"
)

const (
	imagesRepo     = "images"
	edgesRepo      = "edges"
	montageRepo    = "montage"
	upgradeSubject = "upgrade_client"
)

skip := true

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

func TestUpgradeTrigger(t *testing.T) {
	if skip {
		t.Skip("Skipping upgrade test")
	}
	fromVersions := []string{
		"2.7.6",
		"2.8.5",
	}

	type ExpectedCommitCount struct {
		preTrigger1  int
		preTrigger2  int
		postTrigger1 int
		postTrigger2 int
	}

	getExpectedCommitCountFromVersion := func(version string) ExpectedCommitCount {
		expectedCommitMap := map[string]ExpectedCommitCount{
			"v2.7": {
				// 2.7 is a special case where the commit structure changes after the upgrade
				preTrigger1:  23,
				preTrigger2:  12,
				postTrigger1: 33,
				postTrigger2: 17,
			},
			"v2.8": {
				// trigger 1 is two empty commits and 11 data commits
				// trigger 2 is one inital commit and "every other" data commit, so 6 total
				preTrigger1:  13,
				preTrigger2:  6,
				postTrigger1: 13,
				postTrigger2: 6,
			},
		}
		lookup := semver.MajorMinor("v" + version)
		return expectedCommitMap[lookup]
	}

	dataRepo := "TestTrigger_data"
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	pipeline1 := "TestTrigger1"
	pipeline2 := "TestTrigger2"

	logCommits := func(t *testing.T, c *client.APIClient, commits []*pfs.CommitInfo) {
		var buf bytes.Buffer
		for i, commit := range commits {
			err := c.GetFile(commit.Commit, "/hello", &buf)
			commitFile := buf.String()
			buf.Reset()
			if err != nil || commitFile == "" {
				commitFile = "no file"
			}
			t.Logf("	commit %d: id:%s, file: %s", len(commits)-i, commit.Commit.Id, commitFile)
		}
	}

	upgradeTest(t, pctx.TestContext(t), true /* parallelOK */, 1, fromVersions,
		func(t *testing.T, ctx context.Context, c *client.APIClient, from string) { /* preUpgrade */
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
			// after 2.7.x pachyderm doesn't come with a "master" branch anymore, so we create it in this test
			_, err := c.PfsAPIClient.CreateBranch(c.Ctx(), &pfs.CreateBranchRequest{
				Branch: &pfs.Branch{Repo: &pfs.Repo{Name: dataRepo, Type: pfs.UserRepoType, Project: &pfs.Project{Name: pfs.DefaultProjectName}}, Name: "master"},
			})
			require.NoError(t, err)
			require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
				pipeline1,
				"",
				[]string{"bash"},
				[]string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
				&pps.ParallelismSpec{
					Constant: 1,
				},
				client.NewPFSInputOpts(dataRepo, pfs.DefaultProjectName, dataRepo, "trigger", "/*", "", "", false, false, &pfs.Trigger{
					Branch:  "master",
					Commits: 1,
				}),
				"",
				false,
			))
			require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
				pipeline2,
				"",
				[]string{"bash"},
				[]string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", pipeline1),
				},
				&pps.ParallelismSpec{
					Constant: 1,
				},
				client.NewPFSInputOpts(pipeline1, pfs.DefaultProjectName, pipeline1, "trigger", "/*", "", "", false, false, &pfs.Trigger{
					Branch:  "master",
					Commits: 2,
				}),
				"",
				false,
			))
			for i := 0; i < 11; i++ {
				require.NoError(t, c.PutFile(dataCommit, "/hello", strings.NewReader(fmt.Sprintf("hello world %v", i))))
			}
			require.NoError(t, err)
			expectedCommitCount := getExpectedCommitCountFromVersion(from)
			require.NoErrorWithinTRetry(t, 2*time.Minute, func() error {
				commits, err := c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline1), nil, nil, 0)
				require.NoError(t, err)
				if err != nil {
					return err
				}
				t.Logf("comparing commit trigger1 sizes %d/%d", len(commits), expectedCommitCount.preTrigger1)
				logCommits(t, c, commits)
				if got, want := len(commits), expectedCommitCount.preTrigger1; got != want {
					return errors.Errorf("trigger1 not ready; got %v commits, want %v commits", got, want)
				}
				return nil
			})
			require.NoErrorWithinTRetry(t, 2*time.Minute, func() error {
				commits, err := c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline2), nil, nil, 0)
				require.NoError(t, err)
				if err != nil {
					return err
				}
				t.Logf("comparing commit trigger2 sizes %d/%d", len(commits), expectedCommitCount.preTrigger2)
				logCommits(t, c, commits)
				if got, want := len(commits), expectedCommitCount.preTrigger2; got != want {
					return errors.Errorf("trigger2 not ready; got %v commits, want %v commits", got, want)
				}
				return nil
			})
		},
		func(t *testing.T, ctx context.Context, c *client.APIClient, from string) { /* postUpgrade */
			for i := 0; i < 10; i++ {
				require.NoError(t, c.PutFile(dataCommit, "/hello", strings.NewReader(fmt.Sprintf("hello world post %v", i))))
			}
			latestDataCI, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
			require.NoError(t, err)
			if semver.Compare("v"+from, "v2.8.0") < 0 {
				// these alias commits only exist before v2.8.x
				require.NoErrorWithinTRetryConstant(t, 5*time.Minute, func() error {
					ci, err := c.InspectCommit(pfs.DefaultProjectName, pipeline2, "master", "")
					require.NoError(t, err)
					aliasCI, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "", ci.Commit.Id)
					if err != nil {
						return err
					}
					if got, want := latestDataCI.Commit.Id, aliasCI.Commit.Id; got != want {
						return errors.Errorf("not ready alias commit: %v latest data commit: %v", aliasCI.Commit.Id, latestDataCI.Commit.Id)
					}
					return nil
				}, 10*time.Second)
			}
			expectedCommitCount := getExpectedCommitCountFromVersion(from)
			commits, err := c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline1), nil, nil, 0)
			require.NoError(t, err)
			t.Logf("comparing commit trigger1 post %d/%d", len(commits), expectedCommitCount.postTrigger1)
			logCommits(t, c, commits)
			require.Equal(t, expectedCommitCount.postTrigger1, len(commits))
			commits, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline2), nil, nil, 0)
			require.NoError(t, err)
			t.Logf("comparing commit trigger2 post %d/%d", len(commits), expectedCommitCount.postTrigger2)
			logCommits(t, c, commits)
			require.Equal(t, expectedCommitCount.postTrigger2, len(commits))
			if semver.Compare("v"+from, "v2.8.0") < 0 {
				// parent branch default/TestTrigger_data@trigger commit is not in direct provenance of head of branch default/TestTrigger1@master
				require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
					if resp.Error != "" {
						return errors.Errorf(resp.Error)
					}
					return nil
				}))
			}
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
				return errors.EnsureStack(mf.PutFileURL("/liberty.png", "https://docs.pachyderm.com/images/opencv/liberty.jpg", false))
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
				return errors.EnsureStack(mf.PutFileURL("/kitten.png", "https://docs.pachyderm.com/images/opencv/kitten.jpg", false))
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
