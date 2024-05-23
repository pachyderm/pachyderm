//go:build k8s

package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"math"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-units"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfspretty "github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
	ppspretty "github.com/pachyderm/pachyderm/v2/src/server/pps/pretty"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"

	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

func newCountBreakFunc(maxCount int) func(func() error) error {
	var count int
	return func(cb func() error) error {
		if err := cb(); err != nil {
			return err
		}
		count++
		if count == maxCount {
			return errutil.ErrBreak
		}
		return nil
	}
}

func basicPipelineReq(name, input string) *pps.CreatePipelineRequest {
	return &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(pfs.DefaultProjectName, name),
		Transform: &pps.Transform{
			Cmd: []string{"bash"},
			Stdin: []string{
				fmt.Sprintf("cp /pfs/%s/* /pfs/out/", input),
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{
			Constant: 1,
		},
		Input: client.NewPFSInput(pfs.DefaultProjectName, input, "/*"),
	}
}

func createTestCommits(t *testing.T, repoName, branchName string, numCommits int, c *client.APIClient) {
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repoName), "pods should be available to create a repo against")

	for i := 0; i < numCommits; i++ {
		commit, err := c.StartCommit(pfs.DefaultProjectName, repoName, branchName)
		require.NoError(t, err, "creating a test commit should succeed")

		err = c.PutFile(commit, "/file", bytes.NewBufferString("file contents"))
		require.NoError(t, err, "should be able to add a file to a commit")

		err = c.FinishCommit(pfs.DefaultProjectName, repoName, branchName, commit.Id)
		require.NoError(t, err, "finishing a commit should succeed")
	}
}

func deleteEtcd(t *testing.T, ctx context.Context, namespace string) {
	t.Helper()
	kubeClient := tu.GetKubeClient(t)
	label := "app=etcd"
	etcd, err := kubeClient.CoreV1().Pods(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: label})
	require.NoError(t, err, "Error attempting to find etcd pod.")
	require.Equal(t, 1, len(etcd.Items), "etcd did not have the correct number of pods.")

	watcher, err := kubeClient.CoreV1().Pods(namespace).
		Watch(ctx, metav1.ListOptions{LabelSelector: label})
	defer watcher.Stop()
	require.NoError(t, err, "Starting etcd pod watch failed.")

	require.NoError(t, kubeClient.CoreV1().Pods(namespace).Delete(
		context.Background(),
		etcd.Items[0].ObjectMeta.Name, metav1.DeleteOptions{}), "Deleting etcd pod failed.")

	for event := range watcher.ResultChan() {
		if event.Type == watch.Deleted {
			break
		}
	}
}

// Wait for at least one pod with the given selector to be running and ready.
func waitForOnePodReady(t testing.TB, ctx context.Context, namespace string, label string) error {
	t.Helper()
	kubeClient := tu.GetKubeClient(t)
	return backoff.Retry(func() error {
		pods, err := kubeClient.CoreV1().Pods(namespace).
			List(ctx, metav1.ListOptions{LabelSelector: label})
		if err != nil {
			return errors.EnsureStack(err)
		}

		if len(pods.Items) < 1 {
			return errors.Errorf("pod with label %s has not yet been restarted.", label)
		}
		for _, item := range pods.Items {
			if item.Status.Phase == v1.PodRunning {
				for _, c := range item.Status.Conditions {
					if c.Type == v1.PodReady && c.Status == v1.ConditionTrue {
						return nil
					}
				}
			}
		}
		return errors.Errorf("one pod with label %s is not yet running and ready.", label)
	}, backoff.RetryEvery(time.Second).For(40*time.Second))
}

func TestCreatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	c = c.WithDefaultTransformUser("1000")

	projectName := tu.UniqueString("project")
	require.NoError(t, c.CreateProject(projectName))
	dataRepoName := tu.UniqueString("TestSimplePipeline_data")
	require.NoError(t, c.CreateRepo(projectName, dataRepoName))

	commit1, err := c.StartCommit(projectName, dataRepoName, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(projectName, dataRepoName, "", commit1.Id))

	pipeline := tu.UniqueString("TestSimplePipeline")
	require.NoError(t, c.CreatePipeline(projectName,
		pipeline,
		tu.DefaultTransformImage,
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepoName),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(projectName, dataRepoName, "/*"),
		"",
		false,
	))

	commitInfo, err := c.InspectCommit(projectName, pipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	// The commitset should have a commit in: data, spec, pipeline, meta
	// the last two are dependent upon the first two, so should come later
	// in topological ordering
	require.Equal(t, 4, len(commitInfos))
	var commitRepos []*pfs.Repo
	for _, info := range commitInfos {
		commitRepos = append(commitRepos, info.Commit.Repo)
	}
	require.EqualOneOf(t, commitRepos[:2], client.NewRepo(projectName, dataRepoName))
	require.EqualOneOf(t, commitRepos[:2], client.NewSystemRepo(projectName, pipeline, pfs.SpecRepoType))
	require.EqualOneOf(t, commitRepos[2:], client.NewRepo(projectName, pipeline))
	require.EqualOneOf(t, commitRepos[2:], client.NewSystemRepo(projectName, pipeline, pfs.MetaRepoType))

	var buf bytes.Buffer
	for _, info := range commitInfos {
		if proto.Equal(info.Commit.Repo, client.NewRepo(projectName, pipeline)) {
			require.NoError(t, c.GetFile(info.Commit, "file", &buf))
			require.Equal(t, "foo", buf.String())
		}
	}

	pipeline = strings.Repeat("x", 59-len(projectName)-1)
	require.NoError(t, c.CreatePipeline(projectName,
		pipeline,
		tu.DefaultTransformImage,
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepoName),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(projectName, dataRepoName, "/*"),
		"",
		false,
	), "adding a pipeline with a name summing to 59 characters (with the project name and a hyphen) should be okay")

	pipeline = strings.Repeat("x", 60-len(projectName)-1)
	require.YesError(t, c.CreatePipeline(projectName,
		pipeline,
		tu.DefaultTransformImage,
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepoName),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(projectName, dataRepoName, "/*"),
		"",
		false,
	), "adding a pipeline with a name summing to 60 characters (with the project name and a hyphen) should error")
}

func TestPipelineWithSubprocesses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	c = c.WithDefaultTransformUser("1000")

	projectName := tu.UniqueString("p")
	require.NoError(t, c.CreateProject(projectName))
	dataRepoName := tu.UniqueString("TestPipelineWithSubprocesses_data")
	require.NoError(t, c.CreateRepo(projectName, dataRepoName))

	commit1, err := c.StartCommit(projectName, dataRepoName, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "foo", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.PutFile(commit1, "bar", strings.NewReader("bar"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(projectName, dataRepoName, "", commit1.Id))

	pipeline := tu.UniqueString("TestPipelineWS")
	_, err = c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(projectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"/bin/bash"},
				Stdin: []string{
					"sleep infinity &", // sleep holds onto stdout forever, meaning we have to explicitly kill it to make progress
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepoName),
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: 1},
			Input:           client.NewPFSInput(projectName, dataRepoName, "/*"),
			DatumTimeout:    durationpb.New(5 * time.Second),
		},
	)
	require.NoError(t, err, "should create pipeline ok")

	commitInfo, err := c.InspectCommit(projectName, pipeline, "master", "")
	require.NoError(t, err, "should inspect commit ok")
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err, "should wait commit ok")

	var output *pfs.CommitInfo
	for _, info := range commitInfos {
		if proto.Equal(info.Commit.Repo, client.NewRepo(projectName, pipeline)) {
			output = info
			break
		}
	}
	require.NotNil(t, output, "output commit should have been found (got: %#v)", commitInfos)

	buf := new(bytes.Buffer)
	require.NoError(t, c.GetFile(output.Commit, "foo", buf), "should get foo without error")
	require.Equal(t, "foo", buf.String(), "content should be correct")
	buf.Reset()
	require.NoError(t, c.GetFile(output.Commit, "bar", buf), "should get bar without error")
	require.Equal(t, "bar", buf.String(), "content should be correct")
}

func TestCrossProjectPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	c = c.WithDefaultTransformUser("1000")

	inputProjectName := tu.UniqueString("input")
	require.NoError(t, c.CreateProject(inputProjectName))
	dataRepoName := tu.UniqueString("TestSimplePipeline_data")
	require.NoError(t, c.CreateRepo(inputProjectName, dataRepoName))

	commit1, err := c.StartCommit(inputProjectName, dataRepoName, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(inputProjectName, dataRepoName, "", commit1.Id))

	pipelineProjectName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreateProject(pipelineProjectName))

	pipeline := tu.UniqueString("TestPipeline")
	require.NoError(t, c.CreatePipeline(pipelineProjectName,
		pipeline,
		tu.DefaultTransformImage,
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepoName),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(inputProjectName, dataRepoName, "/*"),
		"",
		false,
	))

	commitInfo, err := c.InspectCommit(pipelineProjectName, pipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	// The commitset should have a commit in: data, spec, pipeline, meta
	// the last two are dependent upon the first two, so should come later
	// in topological ordering
	require.Equal(t, 4, len(commitInfos))
	var commitRepos []*pfs.Repo
	for _, info := range commitInfos {
		commitRepos = append(commitRepos, info.Commit.Repo)
	}
	require.EqualOneOf(t, commitRepos[:2], client.NewRepo(inputProjectName, dataRepoName))
	require.EqualOneOf(t, commitRepos[:2], client.NewSystemRepo(pipelineProjectName, pipeline, pfs.SpecRepoType))
	require.EqualOneOf(t, commitRepos[2:], client.NewRepo(pipelineProjectName, pipeline))
	require.EqualOneOf(t, commitRepos[2:], client.NewSystemRepo(pipelineProjectName, pipeline, pfs.MetaRepoType))

	var buf bytes.Buffer
	for _, info := range commitInfos {
		if proto.Equal(info.Commit.Repo, client.NewRepo(pipelineProjectName, pipeline)) {
			require.NoError(t, c.GetFile(info.Commit, "file", &buf))
			require.Equal(t, "foo", buf.String())
		}
	}
}

func TestRepoSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// create a data repo
	dataRepo := tu.UniqueString("TestRepoSize_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// create a pipeline
	pipeline := tu.UniqueString("TestRepoSize")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	// put a file without an open commit - should count towards repo size
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "file2", strings.NewReader("foo"), client.WithAppendPutFile()))

	// put a file on another branch - should not count towards repo size
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "develop", ""), "file3", strings.NewReader("foo"), client.WithAppendPutFile()))

	// put a file on an open commit - should count toward repo size
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file1", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	// wait for everything to be processed
	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	// check data repo size
	repoInfo, err := c.InspectRepo(pfs.DefaultProjectName, dataRepo)
	require.NoError(t, err)
	require.Equal(t, int64(6), repoInfo.Details.SizeBytes)

	// check pipeline repo size
	repoInfo, err = c.InspectRepo(pfs.DefaultProjectName, pipeline)
	require.NoError(t, err)
	require.Equal(t, int64(6), repoInfo.Details.SizeBytes)

	// ensure size is updated when we delete a commit
	require.NoError(t, c.DropCommitSet(commit1.Id))
	repoInfo, err = c.InspectRepo(pfs.DefaultProjectName, dataRepo)
	require.NoError(t, err)
	require.Equal(t, int64(3), repoInfo.Details.SizeBytes)
	repoInfo, err = c.InspectRepo(pfs.DefaultProjectName, pipeline)
	require.NoError(t, err)
	require.Equal(t, int64(3), repoInfo.Details.SizeBytes)
}

func TestPFSPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPFSPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("TestPFSPipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	var buf bytes.Buffer
	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit1.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buf))
	require.Equal(t, "foo", buf.String())
}

func TestPipelineWithParallelism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineWithParallelism_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 4,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	numFiles := 200
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit1.Id)
	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(outputCommit, fmt.Sprintf("file-%d", i), &buf))
		require.Equal(t, fmt.Sprintf("%d", i), buf.String())
	}
}

func TestPipelineWithLargeFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	project := tu.UniqueString("P_")
	require.NoError(t, c.CreateProject(project))

	dataRepo := tu.UniqueString("TestPipelineWithLargeFiles_data")
	require.NoError(t, c.CreateRepo(project, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(project,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		nil,
		client.NewPFSInput(project, dataRepo, "/*"),
		"",
		false,
	))

	random.SeedRand(99)
	numFiles := 10
	var fileContents []string
	commit1, err := c.StartCommit(project, dataRepo, "master")
	require.NoError(t, err)
	chunkSize := int(pfs.ChunkSize / 32) // We used to use a full ChunkSize, but it was increased which caused this test to take too long.
	for i := 0; i < numFiles; i++ {
		fileContent := random.String(chunkSize + i*units.MB)
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file-%d", i), strings.NewReader(fileContent), client.WithAppendPutFile()))
		fileContents = append(fileContents, fileContent)
	}
	require.NoError(t, c.FinishCommit(project, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	outputCommit := client.NewCommit(project, pipeline, "master", commit1.Id)

	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		fileName := fmt.Sprintf("file-%d", i)

		fileInfo, err := c.InspectFile(outputCommit, fileName)
		require.NoError(t, err)
		require.Equal(t, chunkSize+i*units.MB, int(fileInfo.SizeBytes))

		require.NoError(t, c.GetFile(outputCommit, fileName, &buf))
		// we don't wanna use the `require` package here since it prints
		// the strings, which would clutter the output.
		if fileContents[i] != buf.String() {
			t.Fatalf("file content does not match")
		}
	}
}

func TestDatumDedup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestDatumDedup_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	// This pipeline sleeps for 10 secs per datum
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"sleep 10",
		},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	commit2, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit2.Id))

	// Since we did not change the datum, the datum should not be processed
	// again, which means that the job should complete instantly.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = c.WithCtx(ctx).WaitCommitSetAll(commit2.Id)
	require.NoError(t, err)
}

func TestPipelineInputDataModification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineInputDataModification_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	var buf bytes.Buffer
	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit1.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buf))
	require.Equal(t, "foo", buf.String())

	// replace the contents of 'file' in dataRepo (from "foo" to "bar")
	commit2, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.DeleteFile(commit2, "file"))
	require.NoError(t, c.PutFile(commit2, "file", strings.NewReader("bar"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit2.Id))

	commitInfos, err = c.WaitCommitSetAll(commit2.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	buf.Reset()
	outputCommit = client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit2.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buf))
	require.Equal(t, "bar", buf.String())

	// Add a file to dataRepo
	commit3, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.DeleteFile(commit3, "file"))
	require.NoError(t, c.PutFile(commit3, "file2", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit3.Id))

	commitInfos, err = c.WaitCommitSetAll(commit3.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	outputCommit = client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit3.Id)
	require.YesError(t, c.GetFile(outputCommit, "file", &buf))
	buf.Reset()
	require.NoError(t, c.GetFile(outputCommit, "file2", &buf))
	require.Equal(t, "foo", buf.String())

	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline), client.NewCommit(pfs.DefaultProjectName, pipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))
}

func TestMultipleInputsFromTheSameBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestMultipleInputsFromTheSameBranch_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"cat /pfs/out/file",
			"cat /pfs/dirA/dirA/file >> /pfs/out/file",
			"cat /pfs/dirB/dirB/file >> /pfs/out/file",
		},
		nil,
		client.NewCrossInput(
			client.NewPFSInputOpts("dirA", pfs.DefaultProjectName, dataRepo, "", "/dirA/*", "", "", false, false, nil),
			client.NewPFSInputOpts("dirB", pfs.DefaultProjectName, dataRepo, "", "/dirB/*", "", "", false, false, nil),
		),
		"",
		false,
	))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "dirA/file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.PutFile(commit1, "dirB/file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	var buf bytes.Buffer
	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit1.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buf))
	require.Equal(t, "foo\nfoo\n", buf.String())

	commit2, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit2, "dirA/file", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit2.Id))

	commitInfos, err = c.WaitCommitSetAll(commit2.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	buf.Reset()
	outputCommit = client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit2.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buf))
	require.Equal(t, "foo\nbar\nfoo\n", buf.String())

	commit3, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit3, "dirB/file", strings.NewReader("buzz\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit3.Id))

	commitInfos, err = c.WaitCommitSetAll(commit3.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	buf.Reset()
	outputCommit = client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit3.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buf))
	require.Equal(t, "foo\nbar\nfoo\nbuzz\n", buf.String())

	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline), client.NewCommit(pfs.DefaultProjectName, pipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))
}

func TestRunPipeline(t *testing.T) {
	// TODO(2.0 optional): Run pipeline creates a dangling commit, but uses the stats branch for stats commits.
	// Since there is no relationship between the commits created by run pipeline, there should be no
	// relationship between the stats commits. So, the stats commits should also be dangling commits.
	// This might be easier to address when global IDs are implemented.
	t.Skip("Run pipeline does not work correctly with stats enabled")
	//if testing.Short() {
	//	t.Skip("Skipping integration tests in short mode")
	//}

	//c := tu.GetPachClient(t)
	//require.NoError(t, c.DeleteAll())

	//// Test on cross pipeline
	//t.Run("RunPipelineCross", func(t *testing.T) {
	//	dataRepo := tu.UniqueString("TestRunPipeline_data")
	//	require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName,dataRepo))

	//	branchA := "branchA"
	//	branchB := "branchB"

	//	pipeline := tu.UniqueString("pipeline")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	//		pipeline,
	//		"",
	//		[]string{"bash"},
	//		[]string{
	//			"cat /pfs/branch-a/file >> /pfs/out/file",
	//			"cat /pfs/branch-b/file >> /pfs/out/file",
	//			"echo ran-pipeline",
	//		},
	//		nil,
	//		client.NewCrossInput(
	//			client.NewProjectPFSInputOpts("branch-a",pfs.DefaultProjectName, dataRepo, branchA, "/*", "", "", false, false, nil),
	//			client.NewProjectPFSInputOpts("branch-b",pfs.DefaultProjectName, dataRepo, branchB, "/*", "", "", false, false, nil),
	//		),
	//		"",
	//		false,
	//	))

	//	commitA, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	require.NoError(t, c.PutFile(dataRepo, commitA.Branch.Name, commitA.ID, "/file", strings.NewReader("data A\n"), client.WithAppendPutFile()))
	//	require.NoError(t, c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA.Branch.Name, commitA.ID))

	//	commitB, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchB)
	//	require.NoError(t, err)
	//	require.NoError(t, c.PutFile(dataRepo, commitB.Branch.Name, commitB.ID, "/file", strings.NewReader("data B\n"), client.WithAppendPutFile()))
	//	require.NoError(t, c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitB.Branch.Name, commitB.ID))

	//	iter, err := c.FlushJob([]*pfs.Commit{commitA, commitB}, nil)
	//	require.NoError(t, err)
	//	require.Equal(t, 2, len(commits))
	//	buffer := bytes.Buffer{}
	//	require.NoError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.Branch.Name, commits[0].Commit.ID, "file", &buffer))
	//	require.Equal(t, "data A\ndata B\n", buffer.String())

	//	commitM, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, "master")
	//	require.NoError(t, err)
	//	require.NoError(t, c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitM.Branch.Name, commitM.ID))

	//	// we should have two jobs
	//	ji, err := c.ListProjectJob(pfs.DefaultProjectName,pipeline, nil, nil, -1, true)
	//	require.NoError(t, err)
	//	require.Equal(t, 2, len(ji))
	//	// now run the pipeline
	//	require.NoError(t, c.RunPipeline(pipeline, nil, ""))
	//	// running the pipeline should create a new job
	//	require.NoError(t, backoff.Retry(func() error {
	//		jobInfos, err := c.ListProjectJob(pfs.DefaultProjectName,pipeline, nil, nil, -1, true)
	//		require.NoError(t, err)
	//		if len(jobInfos) != 3 {
	//			return errors.Errorf("expected 3 jobs, got %d", len(jobInfos))
	//		}
	//		return nil
	//	}, backoff.NewTestingBackOff()))

	//	// now run the pipeline with non-empty provenance
	//	require.NoError(t, backoff.Retry(func() error {
	//		return c.RunPipeline(pipeline, []*pfs.Commit{
	//			client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, "branchA", commitA.ID),
	//		}, "")
	//	}, backoff.NewTestingBackOff()))

	//	// running the pipeline should create a new job
	//	require.NoError(t, backoff.Retry(func() error {
	//		jobInfos, err := c.ListProjectJob(pfs.DefaultProjectName,pipeline, nil, nil, -1, true)
	//		require.NoError(t, err)
	//		if len(jobInfos) != 4 {
	//			return errors.Errorf("expected 4 jobs, got %d", len(jobInfos))
	//		}
	//		return nil
	//	}, backoff.NewTestingBackOff()))

	//	// add some new commits with some new info
	//	commitA2, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	require.NoError(t, c.PutFile(dataRepo, commitA2.Branch.Name, commitA2.ID, "/file", strings.NewReader("data A2\n"), client.WithAppendPutFile()))
	//	require.NoError(t, c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA2.Branch.Name, commitA2.ID))

	//	commitB2, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchB)
	//	require.NoError(t, err)
	//	require.NoError(t, c.PutFile(dataRepo, commitB2.Branch.Name, commitB2.ID, "/file", strings.NewReader("data B2\n"), client.WithAppendPutFile()))
	//	require.NoError(t, c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitB2.Branch.Name, commitB2.ID))

	//	// and make sure the output file is updated appropriately
	//	iter, err = c.FlushJob([]*pfs.Commit{commitA2, commitB2}, nil)
	//	require.NoError(t, err)
	//	require.Equal(t, 2, len(commits))
	//	buffer = bytes.Buffer{}
	//	require.NoError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.Branch.Name, commits[0].Commit.ID, "file", &buffer))
	//	require.Equal(t, "data A\ndata A2\ndata B\ndata B2\n", buffer.String())

	//	// now run the pipeline provenant on the old commits
	//	require.NoError(t, c.RunPipeline(pipeline, []*pfs.Commit{
	//		client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, "branchA", commitA.ID),
	//		client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, "branchB", commitB2.ID),
	//	}, ""))

	//	// and ensure that the file now has the info from the correct versions of the commits
	//	iter, err = c.FlushJob([]*pfs.Commit{commitA, commitB2}, nil)
	//	require.NoError(t, err)
	//	require.Equal(t, 2, len(commits))
	//	buffer = bytes.Buffer{}
	//	require.NoError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.Branch.Name, commits[0].Commit.ID, "file", &buffer))
	//	require.Equal(t, "data A\ndata B\ndata B2\n", buffer.String())

	//	// make sure no commits with this provenance combination exist
	//	iter, err = c.FlushJob([]*pfs.Commit{commitA2, commitB}, nil)
	//	require.NoError(t, err)
	//	require.Equal(t, 0, len(commits))
	//})

	//// Test on pipeline with no commits
	//t.Run("RunPipelineEmpty", func(t *testing.T) {
	//	dataRepo := tu.UniqueString("TestRunPipeline_data")
	//	require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName,dataRepo))

	//	pipeline := tu.UniqueString("empty-pipeline")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	//		pipeline,
	//		"",
	//		nil,
	//		nil,
	//		nil,
	//		nil,
	//		"",
	//		false,
	//	))

	//	// we should have two jobs
	//	ji, err := c.ListProjectJob(pfs.DefaultProjectName,pipeline, nil, nil, -1, true)
	//	require.NoError(t, err)
	//	require.Equal(t, 0, len(ji))
	//	// now run the pipeline
	//	require.YesError(t, c.RunPipeline(pipeline, nil, ""))
	//})

	//// Test on unrelated branch
	//t.Run("RunPipelineUnrelated", func(t *testing.T) {
	//	dataRepo := tu.UniqueString("TestRunPipeline_data")
	//	require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName,dataRepo))

	//	branchA := "branchA"
	//	branchB := "branchB"

	//	pipeline := tu.UniqueString("unrelated-pipeline")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	//		pipeline,
	//		"",
	//		[]string{"bash"},
	//		[]string{
	//			"cat /pfs/branch-a/file >> /pfs/out/file",
	//			"cat /pfs/branch-b/file >> /pfs/out/file",
	//			"echo ran-pipeline",
	//		},
	//		nil,
	//		client.NewCrossInput(
	//			client.NewProjectPFSInputOpts("branch-a","", dataRepo, branchA, "/*", "", "", false, false, nil),
	//			client.NewProjectPFSInputOpts("branch-b","", dataRepo, branchB, "/*", "", "", false, false, nil),
	//		),
	//		"",
	//		false,
	//	))
	//	commitA, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	c.PutFile(dataRepo, commitA.Branch.Name, commitA.ID, "/file", strings.NewReader("data A\n", client.WithAppendPutFile()))
	//	c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA.Branch.Name, commitA.ID)

	//	commitM, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, "master")
	//	require.NoError(t, err)
	//	err = c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitM.Branch.Name, commitM.ID)
	//	require.NoError(t, err)

	//	require.NoError(t, c.CreateProjectBranch(pfs.DefaultProjectName,dataRepo, "unrelated", "", nil))
	//	commitU, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, "unrelated")
	//	require.NoError(t, err)
	//	err = c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitU.Branch.Name, commitU.ID)
	//	require.NoError(t, err)

	//	_, err = c.FlushJob([]*pfs.Commit{commitA, commitM, commitU}, nil)
	//	require.NoError(t, err)

	//	// now run the pipeline with unrelated provenance
	//	require.YesError(t, c.RunPipeline(pipeline, []*pfs.Commit{
	//		client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, "unrelated", commitU.ID)}, ""))
	//})

	//// Test with downstream pipeline
	//t.Run("RunPipelineDownstream", func(t *testing.T) {
	//	dataRepo := tu.UniqueString("TestRunPipeline_data")
	//	require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName,dataRepo))

	//	branchA := "branchA"
	//	branchB := "branchB"

	//	pipeline := tu.UniqueString("original-pipeline")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	//		pipeline,
	//		"",
	//		[]string{"bash"},
	//		[]string{
	//			"cat /pfs/branch-a/file >> /pfs/out/file",
	//			"cat /pfs/branch-b/file >> /pfs/out/file",
	//			"echo ran-pipeline",
	//		},
	//		nil,
	//		client.NewCrossInput(
	//			client.NewProjectPFSInputOpts("branch-a","", dataRepo, branchA, "/*", "", "", false, false, nil),
	//			client.NewProjectPFSInputOpts("branch-b","", dataRepo, branchB, "/*", "", "", false, false, nil),
	//		),
	//		"",
	//		false,
	//	))

	//	commitA, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	c.PutFile(dataRepo, commitA.Branch.Name, commitA.ID, "/file", strings.NewReader("data A\n", client.WithAppendPutFile()))
	//	c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA.Branch.Name, commitA.ID)

	//	commitB, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchB)
	//	require.NoError(t, err)
	//	c.PutFile(dataRepo, commitB.Branch.Name, commitB.ID, "/file", strings.NewReader("data B\n", client.WithAppendPutFile()))
	//	c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitB.Branch.Name, commitB.ID)

	//	iter, err := c.FlushJob([]*pfs.Commit{commitA, commitB}, nil)
	//	require.NoError(t, err)
	//	require.Equal(t, 2, len(commits))
	//	buffer := bytes.Buffer{}
	//	require.NoError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.Branch.Name, commits[0].Commit.ID, "file", &buffer))
	//	require.Equal(t, "data A\ndata B\n", buffer.String())

	//	// and make sure we can attatch a downstream pipeline
	//	downstreamPipeline := tu.UniqueString("downstream-pipeline")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	//		downstreamPipeline,
	//		"",
	//		[]string{"/bin/bash"},
	//		[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline) + " /pfs/out/"},
	//		nil,
	//		client.NewProjectPFSInput(pfs.DefaultProjectName,pipeline, "/*"),
	//		"",
	//		false,
	//	))

	//	commitA2, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	err = c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA2.Branch.Name, commitA2.ID)
	//	require.NoError(t, err)

	//	// there should be one job on the old commit for downstreamPipeline
	//	jobInfos, err := c.FlushJobAll([]*pfs.Commit{commitA}, []string{downstreamPipeline})
	//	require.NoError(t, err)
	//	require.Equal(t, 1, len(jobInfos))

	//	// now run the pipeline
	//	require.NoError(t, backoff.Retry(func() error {
	//		return c.RunPipeline(pipeline, []*pfs.Commit{
	//			client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, branchA, commitA.ID),
	//		}, "")
	//	}, backoff.NewTestingBackOff()))

	//	// the downstream pipeline shouldn't have any new jobs, since runpipeline jobs don't propagate
	//	jobInfos, err = c.FlushJobAll([]*pfs.Commit{commitA}, []string{downstreamPipeline})
	//	require.NoError(t, err)
	//	require.Equal(t, 1, len(jobInfos))

	//	// now rerun the one job that we saw
	//	require.NoError(t, backoff.Retry(func() error {
	//		return c.RunPipeline(downstreamPipeline, nil, jobInfos[0].Job.ID)
	//	}, backoff.NewTestingBackOff()))

	//	// we should now have two jobs
	//	jobInfos, err = c.FlushJobAll([]*pfs.Commit{commitA}, []string{downstreamPipeline})
	//	require.NoError(t, err)
	//	require.Equal(t, 2, len(jobInfos))
	//})

	//// Test with a downstream pipeline who's upstream has no datum, but where the downstream still needs to succeed
	//t.Run("RunPipelineEmptyUpstream", func(t *testing.T) {
	//	dataRepo := tu.UniqueString("TestRunPipeline_data")
	//	require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName,dataRepo))

	//	branchA := "branchA"
	//	branchB := "branchB"

	//	pipeline := tu.UniqueString("pipeline-downstream")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	//		pipeline,
	//		"",
	//		[]string{"bash"},
	//		[]string{
	//			"cat /pfs/branch-a/file >> /pfs/out/file",
	//			"cat /pfs/branch-b/file >> /pfs/out/file",
	//			"echo ran-pipeline",
	//		},
	//		nil,
	//		client.NewCrossInput(
	//			client.NewProjectPFSInputOpts("branch-a","", dataRepo, branchA, "/*", "", "", false, false, nil),
	//			client.NewProjectPFSInputOpts("branch-b","", dataRepo, branchB, "/*", "", "", false, false, nil),
	//		),
	//		"",
	//		false,
	//	))

	//	commitA, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	c.PutFile(dataRepo, commitA.Branch.Name, commitA.ID, "/file", strings.NewReader("data A\n", client.WithAppendPutFile()))
	//	c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA.Branch.Name, commitA.ID)

	//	iter, err := c.FlushJob([]*pfs.Commit{commitA}, nil)
	//	require.NoError(t, err)
	//	require.Equal(t, 2, len(commits))

	//	// no commit to branch-b so "file" should not exist
	//	buffer := bytes.Buffer{}
	//	require.YesError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.Branch.Name, commits[0].Commit.ID, "file", &buffer))

	//	// and make sure we can attatch a downstream pipeline
	//	downstreamPipeline := tu.UniqueString("pipelinedownstream")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	//		downstreamPipeline,
	//		"",
	//		[]string{"/bin/bash"},
	//		[]string{
	//			"cat /pfs/branch-a/file >> /pfs/out/file",
	//			fmt.Sprintf("cat /pfs/%s/file >> /pfs/out/file", pipeline),
	//			"echo ran-pipeline",
	//		},
	//		nil,
	//		client.NewUnionInput(
	//			client.NewProjectPFSInputOpts("branch-a","", dataRepo, branchA, "/*", "", "", false, false, nil),
	//			client.NewProjectPFSInput(pfs.DefaultProjectName,pipeline, "/*"),
	//		),
	//		"",
	//		false,
	//	))

	//	commitA2, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	err = c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA2.Branch.Name, commitA2.ID)
	//	require.NoError(t, err)

	//	// there should be one job on the old commit for downstreamPipeline
	//	jobInfos, err := c.FlushJobAll([]*pfs.Commit{commitA}, []string{downstreamPipeline})
	//	require.NoError(t, err)
	//	require.Equal(t, 1, len(jobInfos))

	//	// now run the pipeline
	//	require.NoError(t, backoff.Retry(func() error {
	//		return c.RunPipeline(pipeline, []*pfs.Commit{
	//			client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, branchA, commitA.ID),
	//		}, "")
	//	}, backoff.NewTestingBackOff()))

	//	buffer2 := bytes.Buffer{}
	//	require.NoError(t, c.GetFile(jobInfos[0].OutputCommit.Repo.Name, jobInfos[0].OutputCommit.Branch.Name, jobInfos[0].OutputCommit.ID, "file", &buffer2))
	//	// the union of an empty output and datA should only return a file with "data A" in it.
	//	require.Equal(t, "data A\n", buffer2.String())

	//	// add another commit to see that we can successfully do the cross and union together
	//	commitB, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchB)
	//	require.NoError(t, err)
	//	c.PutFile(dataRepo, commitB.Branch.Name, commitB.ID, "/file", strings.NewReader("data B\n", client.WithAppendPutFile()))
	//	c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitB.Branch.Name, commitB.ID)

	//	_, err = c.FlushJob([]*pfs.Commit{commitA, commitB}, nil)
	//	require.NoError(t, err)

	//	jobInfos, err = c.FlushJobAll([]*pfs.Commit{commitB}, []string{downstreamPipeline})
	//	require.NoError(t, err)
	//	require.Equal(t, 1, len(jobInfos))

	//	buffer3 := bytes.Buffer{}
	//	require.NoError(t, c.GetFile(jobInfos[0].OutputCommit.Repo.Name, jobInfos[0].OutputCommit.Branch.Name, jobInfos[0].OutputCommit.ID, "file", &buffer3))
	//	// now that we've added data to the other branch of the cross, we should see the union of data A along with the the crossed data.
	//	require.Equal(t, "data A\ndata A\ndata B\n", buffer3.String())
	//})

	//// Test on commits from the same branch
	//t.Run("RunPipelineSameBranch", func(t *testing.T) {
	//	dataRepo := tu.UniqueString("TestRunPipeline_data")
	//	require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName,dataRepo))

	//	branchA := "branchA"
	//	branchB := "branchB"

	//	pipeline := tu.UniqueString("sameBranch-pipeline")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	//		pipeline,
	//		"",
	//		[]string{"bash"},
	//		[]string{
	//			"cat /pfs/branch-a/file >> /pfs/out/file",
	//			"cat /pfs/branch-b/file >> /pfs/out/file",
	//			"echo ran-pipeline",
	//		},
	//		nil,
	//		client.NewCrossInput(
	//			client.NewProjectPFSInputOpts("branch-a","", dataRepo, branchA, "/*", "", "", false, false, nil),
	//			client.NewProjectPFSInputOpts("branch-b","", dataRepo, branchB, "/*", "", "", false, false, nil),
	//		),
	//		"",
	//		false,
	//	))
	//	commitA1, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	c.PutFile(dataRepo, commitA1.Branch.Name, commitA1.ID, "/file", strings.NewReader("data A1\n", client.WithAppendPutFile()))
	//	c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA1.Branch.Name, commitA1.ID)

	//	commitA2, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	c.PutFile(dataRepo, commitA2.Branch.Name, commitA2.ID, "/file", strings.NewReader("data A2\n", client.WithAppendPutFile()))
	//	c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA2.Branch.Name, commitA2.ID)

	//	_, err = c.FlushJob([]*pfs.Commit{commitA1, commitA2}, nil)
	//	require.NoError(t, err)

	//	// now run the pipeline with provenance from the same branch
	//	require.YesError(t, c.RunPipeline(pipeline, []*pfs.Commit{
	//		client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, branchA, commitA1.ID),
	//		client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, branchA, commitA2.ID),
	//	}, ""))
	//})
	//// Test on pipeline that should always fail
	//t.Run("RerunPipeline", func(t *testing.T) {
	//	dataRepo := tu.UniqueString("TestRerunPipeline_data")
	//	require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName,dataRepo))

	//	// jobs on this pipeline should always fail
	//	pipeline := tu.UniqueString("rerun-pipeline")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	//		pipeline,
	//		"",
	//		[]string{"bash"},
	//		[]string{"false"},
	//		nil,
	//		client.NewProjectPFSInputOpts("branch-a","", dataRepo, "branchA", "/*", "", "", false, false, nil),
	//		"",
	//		false,
	//	))

	//	commitA1, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, "branchA")
	//	require.NoError(t, err)
	//	require.NoError(t, c.PutFile(dataRepo, commitA1.Branch.Name, commitA1.ID, "/file", strings.NewReader("data A1\n"), client.WithAppendPutFile()))
	//	require.NoError(t, c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA1.Branch.Name, commitA1.ID))

	//	iter, err := c.FlushJob([]*pfs.Commit{commitA1}, nil)
	//	require.NoError(t, err)
	//	require.Equal(t, 2, len(commits))
	//	// now run the pipeline
	//	require.NoError(t, c.RunPipeline(pipeline, nil, ""))

	//	// running the pipeline should create a new job
	//	require.NoError(t, backoff.Retry(func() error {
	//		jobInfos, err := c.ListProjectJob(pfs.DefaultProjectName,pipeline, nil, nil, -1, true)
	//		require.NoError(t, err)
	//		if len(jobInfos) != 2 {
	//			return errors.Errorf("expected 2 jobs, got %d", len(jobInfos))
	//		}

	//		// but both of these jobs should fail
	//		for i, job := range jobInfos {
	//			if job.State.String() != "JOB_FAILURE" {
	//				return errors.Errorf("expected job %v to fail, but got %v", i, job.State.String())
	//			}
	//		}
	//		return nil
	//	}, backoff.NewTestingBackOff()))

	//	// Shouldn't error if you try to delete an already deleted pipeline
	//	require.NoError(t, c.DeleteProjectPipeline(pfs.DefaultProjectName,pipeline, false))
	//	require.NoError(t, c.DeleteProjectPipeline(pfs.DefaultProjectName,pipeline, false))
	//})
	//t.Run("RunPipelineStats", func(t *testing.T) {
	//	dataRepo := tu.UniqueString("TestRunPipeline_data")
	//	require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName,dataRepo))

	//	branchA := "branchA"

	//	pipeline := tu.UniqueString("stats-pipeline")
	//	_, err := c.PpsAPIClient.CreatePipeline(
	//		context.Background(),
	//		&pps.CreatePipelineRequest{
	//			Pipeline: client.NewProjectPipeline(pfs.DefaultProjectName,pipeline),
	//			Transform: &pps.Transform{
	//				Cmd: []string{"bash"},
	//				Stdin: []string{
	//					"cat /pfs/branch-a/file >> /pfs/out/file",

	//					"echo ran-pipeline",
	//				},
	//			},
	//			Input:       client.NewProjectPFSInputOpts("branch-a","", dataRepo, branchA, "/*", "", "", false, false, nil),
	//		})
	//	require.NoError(t, err)

	//	commitA, err := c.StartProjectCommit(pfs.DefaultProjectName,dataRepo, branchA)
	//	require.NoError(t, err)
	//	c.PutFile(dataRepo, commitA.Branch.Name, commitA.ID, "/file", strings.NewReader("data A\n", client.WithAppendPutFile()))
	//	c.FinishProjectCommit(pfs.DefaultProjectName,dataRepo, commitA.Branch.Name, commitA.ID)

	//	// wait for the commit to finish before calling RunPipeline
	//	_, err = c.WaitCommitSetAll([]*pfs.Commit{client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, commitA.ID)}, nil)
	//	require.NoError(t, err)

	//	// now run the pipeline
	//	require.NoError(t, backoff.Retry(func() error {
	//		return c.RunPipeline(pipeline, []*pfs.Commit{
	//			client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, branchA, commitA.ID),
	//		}, "")
	//	}, backoff.NewTestingBackOff()))

	//	// make sure the pipeline didn't crash
	//	commitInfos, err := c.WaitCommitSetAll([]*pfs.Commit{client.NewProjectCommit(pfs.DefaultProjectName,dataRepo, commitA.ID)}, nil)
	//	require.NoError(t, err)

	//	// we'll know it crashed if this causes it to hang
	//	require.NoErrorWithinTRetry(t, 80*time.Second, func() error {
	//		return nil
	//	})
	//})
}

func TestPipelineJobHasAuthToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	rc := tu.AuthenticateClient(t, c, auth.RootUser)

	dataRepo := tu.UniqueString("TestPipelineJobAuthToken_data")
	require.NoError(t, rc.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit, err := rc.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, rc.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, rc.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	port := c.GetAddress().Port
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, rc.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"echo $PACH_TOKEN | pachctl auth use-auth-token",
			fmt.Sprintf("pachctl config update context --pachd-address $(echo grpc://pachd.$PACH_NAMESPACE.svc.cluster.local:%d)", port),
			"echo $(pachctl auth whoami) >/pfs/out/file2",
			"pachctl list repo",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))
	var jobInfos []*pps.JobInfo
	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err = c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return errors.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		return nil
	}, backoff.NewTestingBackOff()))
	jobInfo, err := rc.WaitJob(pfs.DefaultProjectName, pipeline, jobInfos[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
	buffer := bytes.Buffer{}
	require.NoError(t, c.GetFile(jobInfo.OutputCommit, "file2", &buffer))
	require.Equal(t, fmt.Sprintf("You are \"pipeline:default/%s\"\n", jobInfo.Job.Pipeline.Name), buffer.String())
}

func TestPipelineFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineFailure_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"exit 1"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))
	var jobInfos []*pps.JobInfo
	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err = c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return errors.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		return nil
	}, backoff.NewTestingBackOff()))
	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, pipeline, jobInfos[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)
	require.True(t, strings.Contains(jobInfo.Reason, "datum"))
}

func TestPPSEgressURLOnly(t *testing.T) {
	if testing.Short() {
		t.Skipf("Skipping %s in short mode", t.Name())
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	repo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "file1", strings.NewReader("foo")))

	pipeline := tu.UniqueString("egress")
	pipelineReq := basicPipelineReq(pipeline, repo)
	pipelineReq.Egress = &pps.Egress{URL: fmt.Sprintf("test-minio://%s/%s/%s", minikubetestenv.MinioEndpoint, minikubetestenv.MinioBucket, pipeline)}

	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), pipelineReq)
	require.NoError(t, err)
	require.NoErrorWithinT(t, time.Minute, func() error {
		commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		if err != nil {
			return err
		}
		jobInfo, err := c.InspectJob(pfs.DefaultProjectName, pipeline, commitInfo.Commit.Id, false)
		if err != nil {
			return err
		}
		require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
		return nil
	})
}

func TestEgressFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	repo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "master", commit.Id))

	pipeline := tu.UniqueString("egress")
	_, err = c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Image: tu.DefaultTransformImage,
				Cmd:   []string{"bash"},
				Stdin: []string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repo)},
			},
			Input: &pps.Input{Pfs: &pps.PFSInput{
				Repo: repo,
				Glob: "/*",
			}},
			Egress: &pps.Egress{
				URL: "garbage-url",
			},
		},
	)
	require.NoError(t, err)
	var jobInfos []*pps.JobInfo
	require.NoErrorWithinTRetry(t, time.Minute, func() error {
		jobInfos, err = c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return errors.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		return nil
	})
	time.Sleep(10 * time.Second)
	jobInfo, err := c.InspectJob(pfs.DefaultProjectName, pipeline, jobInfos[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_EGRESSING, jobInfo.State)
	fileInfos, err := c.ListFileAll(commit, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))
}

func TestInputFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestInputFailure_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	pipeline1 := tu.UniqueString("pipeline1")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline1,
		"",
		[]string{"exit 1"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))
	var jobInfos []*pps.JobInfo
	require.NoErrorWithinTRetry(t, time.Minute, func() error {
		jobInfos, err = c.ListJob(pfs.DefaultProjectName, pipeline1, nil, -1, true)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return errors.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		return nil
	})
	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, pipeline1, jobInfos[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)
	require.True(t, strings.Contains(jobInfo.Reason, "datum"))
	pipeline2 := tu.UniqueString("pipeline2")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline2,
		"",
		[]string{"exit 0"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, pipeline1, "/*"),
		"",
		false,
	))
	require.NoErrorWithinTRetry(t, time.Minute, func() error {
		jobInfos, err = c.ListJob(pfs.DefaultProjectName, pipeline2, nil, -1, true)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return errors.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		return nil
	})
	jobInfo, err = c.WaitJob(pfs.DefaultProjectName, pipeline2, jobInfos[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_UNRUNNABLE, jobInfo.State)
	require.True(t, strings.Contains(jobInfo.Reason, "unrunnable because"))

	pipeline3 := tu.UniqueString("pipeline3")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline3,
		"",
		[]string{"exit 0"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, pipeline2, "/*"),
		"",
		false,
	))
	require.NoErrorWithinTRetry(t, time.Minute, func() error {
		jobInfos, err = c.ListJob(pfs.DefaultProjectName, pipeline3, nil, -1, true)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return errors.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		return nil
	})
	jobInfo, err = c.WaitJob(pfs.DefaultProjectName, pipeline3, jobInfos[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_UNRUNNABLE, jobInfo.State)
	// the fact that pipeline 2 failed should be noted in the message
	require.True(t, strings.Contains(jobInfo.Reason, pipeline2))
}

func TestPipelineErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	t.Run("ErrCmd", func(t *testing.T) {

		dataRepo := tu.UniqueString("TestPipelineErrorHandling_data")
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

		dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
		require.NoError(t, c.PutFile(dataCommit, "file1", strings.NewReader("foo\n"), client.WithAppendPutFile()))
		require.NoError(t, c.PutFile(dataCommit, "file2", strings.NewReader("bar\n"), client.WithAppendPutFile()))
		require.NoError(t, c.PutFile(dataCommit, "file3", strings.NewReader("bar\n"), client.WithAppendPutFile()))

		// In this pipeline, we'll have a command that fails for files 2 and 3, and an error handler that fails for file 2
		pipeline := tu.UniqueString("pipeline1")
		_, err := c.PpsAPIClient.CreatePipeline(
			context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:      []string{"bash"},
					Stdin:    []string{"if", fmt.Sprintf("[ -a pfs/%v/file1 ]", dataRepo), "then", "exit 0", "fi", "exit 1"},
					ErrCmd:   []string{"bash"},
					ErrStdin: []string{"if", fmt.Sprintf("[ -a pfs/%v/file3 ]", dataRepo), "then", "exit 0", "fi", "exit 1"},
				},
				Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			})
		require.NoError(t, err)

		commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)
		jobInfo, err := c.InspectJob(pfs.DefaultProjectName, pipeline, commitInfo.Commit.Id, false)
		require.NoError(t, err)

		// We expect the job to fail, and have 1 datum processed, recovered, and failed each
		require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)
		require.Equal(t, int64(1), jobInfo.DataProcessed)
		require.Equal(t, int64(1), jobInfo.DataRecovered)
		require.Equal(t, int64(1), jobInfo.DataFailed)

		// Now update this pipeline, we have the same command as before, but this time the error handling passes for all
		_, err = c.PpsAPIClient.CreatePipeline(
			context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:    []string{"bash"},
					Stdin:  []string{"if", fmt.Sprintf("[ -a pfs/%v/file1 ]", dataRepo), "then", "exit 0", "fi", "exit 1"},
					ErrCmd: []string{"true"},
				},
				Input:  client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
				Update: true,
			})
		require.NoError(t, err)

		commitInfo, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)
		jobInfo, err = c.InspectJob(pfs.DefaultProjectName, pipeline, commitInfo.Commit.Id, false)
		require.NoError(t, err)

		// so we expect the job to succeed, and to have recovered 2 datums
		require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
		require.Equal(t, int64(1), jobInfo.DataProcessed)
		require.Equal(t, int64(0), jobInfo.DataSkipped)
		require.Equal(t, int64(2), jobInfo.DataRecovered)
		require.Equal(t, int64(0), jobInfo.DataFailed)
	})
	t.Run("RecoveredDatums", func(t *testing.T) {
		dataRepo := tu.UniqueString("TestPipelineRecoveredDatums_data")
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

		dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
		require.NoError(t, c.PutFile(dataCommit, "foo", strings.NewReader("bar\n"), client.WithAppendPutFile()))

		// In this pipeline, we'll have a command that fails the datum, and then recovers it
		pipeline := tu.UniqueString("pipeline3")
		_, err := c.PpsAPIClient.CreatePipeline(
			context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:      []string{"bash"},
					Stdin:    []string{"false"},
					ErrCmd:   []string{"bash"},
					ErrStdin: []string{"true"},
				},
				Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			})
		require.NoError(t, err)

		commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)
		jobInfo, err := c.InspectJob(pfs.DefaultProjectName, pipeline, commitInfo.Commit.Id, false)
		require.NoError(t, err)

		// We expect there to be one recovered datum
		require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
		require.Equal(t, int64(0), jobInfo.DataProcessed)
		require.Equal(t, int64(1), jobInfo.DataRecovered)
		require.Equal(t, int64(0), jobInfo.DataFailed)

		// Update the pipeline so that datums will now successfully be processed
		_, err = c.PpsAPIClient.CreatePipeline(
			context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:   []string{"bash"},
					Stdin: []string{"true"},
				},
				Input:  client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
				Update: true,
			})
		require.NoError(t, err)

		commitInfo, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)
		jobInfo, err = c.InspectJob(pfs.DefaultProjectName, pipeline, commitInfo.Commit.Id, false)
		require.NoError(t, err)

		// Now the recovered datum should have been processed
		require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
		require.Equal(t, int64(1), jobInfo.DataProcessed)
		require.Equal(t, int64(0), jobInfo.DataRecovered)
		require.Equal(t, int64(0), jobInfo.DataFailed)
	})
}

func TestLazyPipelinePropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestLazyPipelinePropagation_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipelineA := tu.UniqueString("pipeline-A")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineA,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInputOpts("", pfs.DefaultProjectName, dataRepo, "", "/*", "", "", false, true, nil),
		"",
		false,
	))
	pipelineB := tu.UniqueString("pipeline-B")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineB,
		"",
		[]string{"cp", path.Join("/pfs", pipelineA, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInputOpts("", pfs.DefaultProjectName, pipelineA, "", "/*", "", "", false, true, nil),
		"",
		false,
	))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	_, err = c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)

	jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipelineA, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
	require.NotNil(t, jobInfos[0].Details.Input.Pfs)
	require.Equal(t, true, jobInfos[0].Details.Input.Pfs.Lazy)

	jobInfos, err = c.ListJob(pfs.DefaultProjectName, pipelineB, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
	require.NotNil(t, jobInfos[0].Details.Input.Pfs)
	require.Equal(t, true, jobInfos[0].Details.Input.Pfs.Lazy)
}

func TestLazyPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestLazyPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")

	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo: dataRepo,
					Glob: "/",
					Lazy: true,
				},
			},
		})
	require.NoError(t, err)

	// Do a commit
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	// We put 2 files, 1 of which will never be touched by the pipeline code.
	// This is an important part of the correctness of this test because the
	// job-shim sets up a goro for each pipe, pipes that are never opened will
	// leak but that shouldn't prevent the job from completing.
	require.NoError(t, c.PutFile(dataCommit, "file2", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))

	commitInfos, err := c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	buffer := bytes.Buffer{}
	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", commit.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

func TestEmptyFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestShufflePipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")

	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("if [ -s /pfs/%s/dir/file1 ]; then exit 1; fi", dataRepo),
					fmt.Sprintf("ln -s /pfs/%s/dir/file1 /pfs/out/file1", dataRepo),
					fmt.Sprintf("if [ -s /pfs/%s/dir/file2 ]; then exit 1; fi", dataRepo),
					fmt.Sprintf("ln -s /pfs/%s/dir/file2 /pfs/out/file2", dataRepo),
					fmt.Sprintf("if [ -s /pfs/%s/dir/file3 ]; then exit 1; fi", dataRepo),
					fmt.Sprintf("ln -s /pfs/%s/dir/file3 /pfs/out/file3", dataRepo),
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:       dataRepo,
					Glob:       "/*",
					EmptyFiles: true,
				},
			},
		})
	require.NoError(t, err)

	// Do a commit
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(dataCommit, "/dir/file1", strings.NewReader("foo\n")))
	require.NoError(t, c.PutFile(dataCommit, "/dir/file2", strings.NewReader("foo\n")))
	require.NoError(t, c.PutFile(dataCommit, "/dir/file3", strings.NewReader("foo\n")))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))

	commitInfos, err := c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	buffer := bytes.Buffer{}
	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", commit.Id)
	require.NoError(t, c.GetFile(outputCommit, "file1", &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(outputCommit, "file2", &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(outputCommit, "file3", &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

// TestProvenance creates a pipeline DAG that's not a transitive reduction
// It looks like this:
// A
// | \
// v  v
// B-->C
// When we commit to A we expect to see 1 commit on C rather than 2.
func TestProvenance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	aRepo := tu.UniqueString("A")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, aRepo))

	bPipeline := tu.UniqueString("B")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, aRepo, "/*"),
		"",
		false,
	))

	cPipeline := tu.UniqueString("C")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		cPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("diff %s %s >/pfs/out/file",
			path.Join("/pfs", aRepo, "file"), path.Join("/pfs", bPipeline, "file"))},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, aRepo, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, bPipeline, "/*"),
		),
		"",
		false,
	))

	// commit to aRepo
	commit1, err := c.StartCommit(pfs.DefaultProjectName, aRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, aRepo, "", commit1.Id))

	commit2, err := c.StartCommit(pfs.DefaultProjectName, aRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit2, "file", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, aRepo, "", commit2.Id))

	commitInfos, err := c.WaitCommitSetAll(commit2.Id)
	require.NoError(t, err)
	require.Equal(t, 7, len(commitInfos)) // input repo plus spec/output/meta for b and c pipelines

	for _, ci := range commitInfos {
		if ci.Commit.Repo.Name == cPipeline && ci.Commit.Repo.Type == pfs.UserRepoType {
			require.Equal(t, int64(0), ci.Details.SizeBytes)
		}
	}

	// We should only see three commits in aRepo (empty head, commit1, commit2)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, aRepo), client.NewCommit(pfs.DefaultProjectName, aRepo, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	// There are three commits in the bPipeline repo (bPipeline created, commit1, commit2)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, bPipeline), client.NewCommit(pfs.DefaultProjectName, bPipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	// There are three commits in the cPipeline repo (cPipeline created, commit1, commit2)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, cPipeline), client.NewCommit(pfs.DefaultProjectName, cPipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
}

/*
TestProvenance2 tests the following DAG:

	  A
	 / \
	B   C
	 \ /
	  D
*/
func TestProvenance2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	aRepo := tu.UniqueString("A")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, aRepo))

	bPipeline := tu.UniqueString("B")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "bfile"), "/pfs/out/bfile"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, aRepo, "/b*"),
		"",
		false,
	))

	cPipeline := tu.UniqueString("C")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		cPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "cfile"), "/pfs/out/cfile"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, aRepo, "/c*"),
		"",
		false,
	))

	dPipeline := tu.UniqueString("D")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		dPipeline,
		"",
		[]string{"sh"},
		[]string{
			fmt.Sprintf("diff /pfs/%s/bfile /pfs/%s/cfile >/pfs/out/file", bPipeline, cPipeline),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, bPipeline, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, cPipeline, "/*"),
		),
		"",
		false,
	))

	// commit to aRepo
	commit1, err := c.StartCommit(pfs.DefaultProjectName, aRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "bfile", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.PutFile(commit1, "cfile", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, aRepo, "", commit1.Id))

	commit2, err := c.StartCommit(pfs.DefaultProjectName, aRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit2, "bfile", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.PutFile(commit2, "cfile", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, aRepo, "", commit2.Id))

	_, err = c.WaitCommit(pfs.DefaultProjectName, dPipeline, "", commit2.Id)
	require.NoError(t, err)

	// We should see 3 commits in aRepo (empty head two user commits)
	commitInfos, err := c.ListCommit(client.NewRepo(pfs.DefaultProjectName, aRepo), client.NewCommit(pfs.DefaultProjectName, aRepo, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	// We should see 3 commits in bPipeline (bPipeline creation and from the two user commits)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, bPipeline), client.NewCommit(pfs.DefaultProjectName, bPipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	// We should see 3 commits in cPipeline (cPipeline creation and from the two user commits)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, cPipeline), client.NewCommit(pfs.DefaultProjectName, cPipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	// We should see 3 commits in dPipeline (dPipeline creation and from the two user commits)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, dPipeline), client.NewCommit(pfs.DefaultProjectName, dPipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	buffer := bytes.Buffer{}
	outputCommit := client.NewCommit(pfs.DefaultProjectName, dPipeline, "master", commit1.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buffer))
	require.Equal(t, "", buffer.String())

	buffer.Reset()
	outputCommit = client.NewCommit(pfs.DefaultProjectName, dPipeline, "master", commit2.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buffer))
	require.Equal(t, "", buffer.String())
}

// TestStopPipelineExtraCommit generates the following DAG:
// A -> B -> C
// and ensures that calling StopPipeline on B does not create an commit in C.
func TestStopPipelineExtraCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	aRepo := tu.UniqueString("A")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, aRepo))

	bPipeline := tu.UniqueString("B")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, aRepo, "/*"),
		"",
		false,
	))

	cPipeline := tu.UniqueString("C")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		cPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, bPipeline, "/*"),
		"",
		false,
	))

	// commit to aRepo
	commit1, err := c.StartCommit(pfs.DefaultProjectName, aRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, aRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 7, len(commitInfos))

	// We should see 2 commits in aRepo (empty head and the user commit)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, aRepo), client.NewCommit(pfs.DefaultProjectName, aRepo, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	// We should see 2 commits in bPipeline (bPipeline creation and the user commit)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, bPipeline), client.NewCommit(pfs.DefaultProjectName, bPipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	// We should see 2 commits in cPipeline (cPipeline creation and the user commit)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, cPipeline), client.NewCommit(pfs.DefaultProjectName, cPipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	require.NoError(t, c.StopPipeline(pfs.DefaultProjectName, bPipeline))
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, cPipeline), client.NewCommit(pfs.DefaultProjectName, cPipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
}

func TestWaitJobSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	prefix := tu.UniqueString("repo")
	makeRepoName := func(i int) string {
		return fmt.Sprintf("%s-%d", prefix, i)
	}

	sourceRepo := makeRepoName(0)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, sourceRepo))

	// Create a four-stage pipeline
	numStages := 4
	for i := 0; i < numStages; i++ {
		repo := makeRepoName(i)
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			makeRepoName(i+1),
			"",
			[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
			nil,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
			"",
			false,
		))
	}

	for i := 0; i < 5; i++ {
		commit, err := c.StartCommit(pfs.DefaultProjectName, sourceRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, sourceRepo, "", commit.Id))

		commitInfos, err := c.WaitCommitSetAll(commit.Id)
		require.NoError(t, err)
		require.Equal(t, numStages*3+1, len(commitInfos))

		jobInfos, err := c.WaitJobSetAll(commit.Id, false)
		require.NoError(t, err)
		require.Equal(t, numStages, len(jobInfos))
	}
}

func TestWaitJobSetFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestWaitJobSetFailures")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	prefix := tu.UniqueString("TestWaitJobSetFailures")
	pipelineName := func(i int) string { return prefix + fmt.Sprintf("%d", i) }

	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName(0),
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo)},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName(1),
		"",
		[]string{"sh"},
		[]string{
			fmt.Sprintf("if [ -f /pfs/%s/file1 ]; then exit 1; fi", pipelineName(0)),
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", pipelineName(0)),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, pipelineName(0), "/*"),
		"",
		false,
	))
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName(2),
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", pipelineName(1))},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, pipelineName(1), "/*"),
		"",
		false,
	))

	for i := 0; i < 2; i++ {
		commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(commit, fmt.Sprintf("file%d", i), strings.NewReader("foo\n"), client.WithAppendPutFile()))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
		jobInfos, err := c.WaitJobSetAll(commit.Id, false)
		require.NoError(t, err)
		require.Equal(t, 3, len(jobInfos))
		if i == 0 {
			for _, ji := range jobInfos {
				require.Equal(t, pps.JobState_JOB_SUCCESS.String(), ji.State.String())
			}
		} else {
			for _, ji := range jobInfos {
				switch ji.Job.Pipeline.Name {
				case pipelineName(1):
					require.Equal(t, pps.JobState_JOB_FAILURE.String(), ji.State.String())
				case pipelineName(2):
					require.Equal(t, pps.JobState_JOB_UNRUNNABLE.String(), ji.State.String())
				}
			}
		}
	}
}

func TestWaitCommitSetAfterCreatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	repo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))

	var commit *pfs.Commit
	var err error
	for i := 0; i < 10; i++ {
		commit, err = c.StartCommit(pfs.DefaultProjectName, repo, "dev")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(commit, "file", strings.NewReader(fmt.Sprintf("foo%d\n", i)), client.WithAppendPutFile()))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "", commit.Id))
	}
	require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, repo, "master", "dev", commit.Id, nil))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"",
		false,
	))
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
}

// TestRecreatePipeline tracks #432
func TestRecreatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	repo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "", commit.Id))
	pipeline := tu.UniqueString("pipeline")
	createPipeline := func() {
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
			nil,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
			"",
			false,
		))
		_, err := c.WaitCommitSetAll(commit.Id)
		require.NoError(t, err)
	}

	// Do it twice.  We expect jobs to be created on both runs.
	createPipeline()
	time.Sleep(5 * time.Second)
	require.NoError(t, c.DeletePipeline(pfs.DefaultProjectName, pipeline, false))
	time.Sleep(5 * time.Second)
	createPipeline()
}

func TestDeletePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	project := tu.UniqueString("prj")
	require.NoError(t, c.CreateProject(project))
	repo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, uuid.NewWithoutDashes(), strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "", commit.Id))
	pipelines := []string{tu.UniqueString("TestDeletePipeline1"), tu.UniqueString("TestDeletePipeline2")}
	createPipelines := func() {
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipelines[0],
			"",
			[]string{"sleep", "20"},
			nil,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
			"",
			false,
		))
		require.NoError(t, c.CreatePipeline(project,
			pipelines[1],
			"",
			[]string{"sleep", "20"},
			nil,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(pfs.DefaultProjectName, pipelines[0], "/*"),
			"",
			false,
		))
		time.Sleep(10 * time.Second)
		// Wait for the pipeline to start running
		require.NoErrorWithinTRetry(t, 90*time.Second, func() error {
			pipelineInfos, err := c.ListPipeline()
			if err != nil {
				return err
			}
			// Check number of pipelines
			names := make([]string, 0, len(pipelineInfos))
			for _, pi := range pipelineInfos {
				names = append(names, fmt.Sprintf("(%s, %s)", pi.Pipeline.Name, pi.State))
			}
			if len(pipelineInfos) != 2 {
				return errors.Errorf("Expected two pipelines, but got: %+v", names)
			}
			// make sure second pipeline is running
			pipelineInfo, err := c.InspectPipeline(project, pipelines[1], false)
			if err != nil {
				return err
			}
			if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
				return errors.Errorf("no running pipeline (only %+v)", names)
			}
			return nil
		})
	}

	createPipelines()

	deletePipeline := func(project, pipeline string) {
		require.NoError(t, c.DeletePipeline(project, pipeline, false))
		// Wait for the pipeline to disappear
		require.NoError(t, backoff.Retry(func() error {
			_, err := c.InspectPipeline(project, pipeline, false)
			if err == nil {
				return errors.Errorf("expected pipeline to be missing, but it's still present")
			}
			return nil
		}, backoff.NewTestingBackOff()))

	}
	// Can't delete a pipeline from the middle of the dag
	require.YesError(t, c.DeletePipeline(pfs.DefaultProjectName, pipelines[0], false))

	deletePipeline(project, pipelines[1])
	deletePipeline(pfs.DefaultProjectName, pipelines[0])

	// The jobs should be gone
	jobs, err := c.ListJob(pfs.DefaultProjectName, "", nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, len(jobs), 0)
	jobs, err = c.ListJob(project, "", nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, len(jobs), 0)

	// Listing jobs for a deleted pipeline should error
	_, err = c.ListJob(pfs.DefaultProjectName, pipelines[0], nil, -1, true)
	require.YesError(t, err)
	_, err = c.ListJob(project, pipelines[1], nil, -1, true)
	require.YesError(t, err)

	createPipelines()

	// Can force delete pipelines from the middle of the dag.
	require.NoError(t, c.DeletePipeline(pfs.DefaultProjectName, pipelines[0], true))
}

func TestPipelineState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	repo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"",
		false,
	))

	// Wait for pipeline to get picked up
	time.Sleep(15 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
			return errors.Errorf("pipeline should be in state running, not: %s", pipelineInfo.State.String())
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Stop pipeline and wait for the pipeline to pause
	require.NoError(t, c.StopPipeline(pfs.DefaultProjectName, pipeline))
	time.Sleep(5 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		if err != nil {
			return err
		}
		if !pipelineInfo.Stopped {
			return errors.Errorf("pipeline never paused, even though StopPipeline() was called, state: %s", pipelineInfo.State.String())
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Restart pipeline and wait for the pipeline to resume
	require.NoError(t, c.StartPipeline(pfs.DefaultProjectName, pipeline))
	time.Sleep(15 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
			return errors.Errorf("pipeline never restarted, even though StartPipeline() was called, state: %s", pipelineInfo.State.String())
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

// TestUpdatePipelineThatHasNoOutput tracks #1637
func TestUpdatePipelineThatHasNoOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestUpdatePipelineThatHasNoOutput")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"sh"},
		[]string{"exit 1"},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		"",
		false,
	))

	// Wait for job to spawn
	var jobInfos []*pps.JobInfo
	time.Sleep(10 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		var err error
		jobInfos, err = c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
		if err != nil {
			return err
		}
		if len(jobInfos) < 1 {
			return errors.Errorf("job not spawned")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, pipeline, jobInfos[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)

	// Now we update the pipeline
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"sh"},
		[]string{"exit 1"},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		"",
		true,
	))
}

func TestAcceptReturnCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestAcceptReturnCode")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipelineName := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd:              []string{"sh"},
				Stdin:            []string{"exit 1"},
				AcceptReturnCode: []int64{1},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		},
	)
	require.NoError(t, err)

	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	commitInfos, err := c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
	require.Equal(t, commit.Id, jobInfos[0].Job.Id)

	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, pipelineName, jobInfos[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
}

func TestPrettyPrinting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// create repos
	dataRepo := tu.UniqueString("TestPrettyPrinting_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)

	// Do a commit to repo
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	commitInfos, err := c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	repoInfo, err := c.InspectRepo(pfs.DefaultProjectName, dataRepo)
	require.NoError(t, err)
	require.NoError(t, pfspretty.PrintDetailedRepoInfo(pfspretty.NewPrintableRepoInfo(repoInfo)))
	for _, c := range commitInfos {
		require.NoError(t, pfspretty.PrintDetailedCommitInfo(os.Stdout, pfspretty.NewPrintableCommitInfo(c)))
	}

	fileInfo, err := c.InspectFile(commit, "file")
	require.NoError(t, err)
	require.NoError(t, pfspretty.PrintDetailedFileInfo(fileInfo))
	pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, true)
	require.NoError(t, err)
	require.NoError(t, ppspretty.PrintDetailedPipelineInfo(os.Stdout, ppspretty.NewPrintablePipelineInfo(pipelineInfo)))
	jobInfos, err := c.ListJob(pfs.DefaultProjectName, "", nil, -1, true)
	require.NoError(t, err)
	require.True(t, len(jobInfos) > 0)
	require.NoError(t, ppspretty.PrintDetailedJobInfo(os.Stdout, ppspretty.NewPrintableJobInfo(jobInfos[0], false)))
}

func TestAuthPrettyPrinting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	rc := tu.AuthenticateClient(t, c, auth.RootUser)

	// create repos
	dataRepo := tu.UniqueString("TestPrettyPrinting_data")
	require.NoError(t, rc.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	_, err := rc.PpsAPIClient.CreatePipeline(
		rc.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)

	// Do a commit to repo
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	commitInfos, err := c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	repoInfo, err := c.InspectRepo(pfs.DefaultProjectName, dataRepo)
	require.NoError(t, err)
	require.NoError(t, pfspretty.PrintDetailedRepoInfo(pfspretty.NewPrintableRepoInfo(repoInfo)))
	for _, c := range commitInfos {
		require.NoError(t, pfspretty.PrintDetailedCommitInfo(os.Stdout, pfspretty.NewPrintableCommitInfo(c)))
	}

	fileInfo, err := c.InspectFile(commit, "file")
	require.NoError(t, err)
	require.NoError(t, pfspretty.PrintDetailedFileInfo(fileInfo))
	pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, true)
	require.NoError(t, err)
	require.NoError(t, ppspretty.PrintDetailedPipelineInfo(os.Stdout, ppspretty.NewPrintablePipelineInfo(pipelineInfo)))
	jobInfos, err := c.ListJob(pfs.DefaultProjectName, "", nil, -1, true)
	require.NoError(t, err)
	require.True(t, len(jobInfos) > 0)
	require.NoError(t, ppspretty.PrintDetailedJobInfo(os.Stdout, ppspretty.NewPrintableJobInfo(jobInfos[0], false)))
}

func TestDeleteAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// this test cannot be run in parallel because it deletes everything
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestDeleteAll_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		"",
		false,
	))
	// Do commit to repo
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
	_, err = c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)
	require.NoError(t, c.DeleteAll(c.Ctx()))
	repoInfos, err := c.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 0, len(repoInfos))
	pipelineInfos, err := c.ListPipeline()
	require.NoError(t, err)
	require.Equal(t, 0, len(pipelineInfos))
	jobInfos, err := c.ListJob(pfs.DefaultProjectName, "", nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 0, len(jobInfos))
}

func TestRecursiveCp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestRecursiveCp_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	// create pipeline
	pipelineName := tu.UniqueString("TestRecursiveCp")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"sh"},
		[]string{
			fmt.Sprintf("cp -r /pfs/%s /pfs/out", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))
	// Do commit to repo
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		require.NoError(t, c.PutFile(
			commit,
			fmt.Sprintf("file%d", i),
			strings.NewReader(strings.Repeat("foo\n", 10000)),
		))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
	_, err = c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)
}

func TestPipelineUniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	repo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/"),
		"",
		false,
	))
	err := c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/"),
		"",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "pipeline .*? already exists", err.Error())
}

func TestUpdatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)
	// create repos and create the pipeline
	dataRepo := tu.UniqueString("TestUpdatePipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	pipelineName := tu.UniqueString("pipeline")
	pipeline := &pps.Pipeline{Project: &pfs.Project{Name: pfs.DefaultProjectName}, Name: pipelineName}
	pipelineCommit := client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", "")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"echo foo >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		true,
	))

	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("1"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))

	_, err = c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineCommit, "file", &buffer))
	require.Equal(t, "foo\n", buffer.String())

	// Update the pipeline
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"echo bar >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		true,
	))

	// Confirm that k8s resources have been updated (fix #4071)
	require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
		kc := tu.GetKubeClient(t)
		svcs, err := kc.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)
		var (
			newServiceSeen bool
			staleName      = ppsutil.PipelineRcName(&pps.PipelineInfo{Pipeline: pipeline, Version: 1})
			newName        = ppsutil.PipelineRcName(&pps.PipelineInfo{Pipeline: pipeline, Version: 2})
		)
		for _, svc := range svcs.Items {
			switch svc.ObjectMeta.Name {
			case staleName:
				return errors.Errorf("stale service encountered: %q", svc.ObjectMeta.Name)
			case newName:
				newServiceSeen = true
			}
		}
		if !newServiceSeen {
			return errors.Errorf("did not find new service: %q", newName)
		}
		rcs, err := kc.CoreV1().ReplicationControllers(ns).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)
		var newRCSeen bool
		for _, rc := range rcs.Items {
			switch rc.ObjectMeta.Name {
			case staleName:
				return errors.Errorf("stale RC encountered: %q", rc.ObjectMeta.Name)
			case newName:
				newRCSeen = true
			}
		}
		require.True(t, newRCSeen)
		if !newRCSeen {
			return errors.Errorf("did not find new RC: %q", newName)
		}
		return nil
	})

	commit, err = c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("2"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))
	_, err = c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)

	buffer.Reset()
	require.NoError(t, c.GetFile(pipelineCommit, "file", &buffer))
	require.Equal(t, "bar\n", buffer.String())

	// Inspect the first job to make sure it hasn't changed
	jis, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 4, len(jis))
	require.Equal(t, "echo bar >/pfs/out/file", jis[0].Details.Transform.Stdin[0])
	require.Equal(t, "echo bar >/pfs/out/file", jis[1].Details.Transform.Stdin[0])
	require.Equal(t, "echo foo >/pfs/out/file", jis[2].Details.Transform.Stdin[0])
	require.Equal(t, "echo foo >/pfs/out/file", jis[3].Details.Transform.Stdin[0])

	// Update the pipeline again, this time with Reprocess: true set. Now we
	// should see a different output file
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{"echo buzz >/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input:     client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			Update:    true,
			Reprocess: true,
		})
	require.NoError(t, err)

	// Confirm that k8s resources have been updated (fix #4071)
	require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
		kc := tu.GetKubeClient(t)
		svcs, err := kc.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)
		var (
			newServiceSeen bool
			staleName      = ppsutil.PipelineRcName(&pps.PipelineInfo{Pipeline: pipeline, Version: 1})
			newName        = ppsutil.PipelineRcName(&pps.PipelineInfo{Pipeline: pipeline, Version: 2})
		)
		for _, svc := range svcs.Items {
			switch svc.ObjectMeta.Name {
			case staleName:
				return errors.Errorf("stale service encountered: %q", svc.ObjectMeta.Name)
			case newName:
				newServiceSeen = true
			}
		}
		if !newServiceSeen {
			return errors.Errorf("did not find new service: %q", newName)
		}
		rcs, err := kc.CoreV1().ReplicationControllers(ns).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)
		var newRCSeen bool
		for _, rc := range rcs.Items {
			switch rc.ObjectMeta.Name {
			case staleName:
				return errors.Errorf("stale RC encountered: %q", rc.ObjectMeta.Name)
			case newName:
				newRCSeen = true
			}
		}
		require.True(t, newRCSeen)
		if !newRCSeen {
			return errors.Errorf("did not find new RC: %q", newName)
		}
		return nil
	})

	_, err = c.WaitCommit(pfs.DefaultProjectName, pipelineName, "master", "")
	require.NoError(t, err)
	buffer.Reset()
	require.NoError(t, c.GetFile(pipelineCommit, "file", &buffer))
	require.Equal(t, "buzz\n", buffer.String())
}

func TestUpdatePipelineWithInProgressCommitsAndStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c, _ := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestUpdatePipelineWithInProgressCommitsAndStats_data")
	c = tu.AuthenticatedPachClient(t, c, "pachtest")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	pipeline := tu.UniqueString("pipeline")
	createPipeline := func() {
		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:   []string{"bash"},
					Stdin: []string{"sleep 1"},
				},
				ParallelismSpec: &pps.ParallelismSpec{
					Constant: 1,
				},
				Autoscaling: true, // Autoscale for CORE-1972
				Input:       client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
				Update:      true,
			})
		require.NoError(t, err)
	}
	createPipeline()

	flushJob := func(commitNum int) {
		commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(commit, "file"+strconv.Itoa(commitNum), strings.NewReader("foo"), client.WithAppendPutFile()))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
		commitInfos, err := c.WaitCommitSetAll(commit.Id)
		require.NoError(t, err)
		require.Equal(t, 4, len(commitInfos))
	}
	// Create a new job that should succeed (both output and stats commits should be finished normally).
	flushJob(1)

	// Create multiple new commits.
	numCommits := 15
	for i := 1; i < numCommits; i++ {
		commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(commit, "file"+strconv.Itoa(i), strings.NewReader("foo"), client.WithAppendPutFile()))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
	}
	// wait for jobs to start before updating
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
			return errors.Errorf("pipeline waiting to run, state: %s", pipelineInfo.State.String())
		}
		return nil
	}, backoff.NewConstantBackOff(time.Millisecond*200)))
	// Force the in progress commits to be finished.
	createPipeline()
	// Create a new job that should succeed (should not get blocked on an unfinished stats commit).
	flushJob(numCommits)
}

func TestPipelineCheckStatusExceededExpectedDuration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c, _ := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestPipelineCheckStatusExceededExpectedDuration_data")
	c = tu.AuthenticatedPachClient(t, c, "pachtest")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	pipeline := tu.UniqueString("pipeline")
	createPipeline := func() {
		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:   []string{"bash"},
					Stdin: []string{"sleep infinity"},
				},
				ParallelismSpec: &pps.ParallelismSpec{
					Constant: 1,
				},
				Autoscaling:           true,
				Input:                 client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
				Update:                true,
				MaximumExpectedUptime: durationpb.New(time.Millisecond * 1),
			})
		require.NoError(t, err)
	}
	createPipeline()
	flushJob := func(commitNum int) {
		commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(commit, "file"+strconv.Itoa(commitNum), strings.NewReader("foo"), client.WithAppendPutFile()))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
	}
	flushJob(1)
	time.Sleep(10 * time.Second)
	require.NoErrorWithinTRetry(t, time.Second*30, func() error {
		checkStatusClient, err := c.PpsAPIClient.CheckStatus(
			c.Ctx(),
			&pps.CheckStatusRequest{
				Context: &pps.CheckStatusRequest_Project{
					Project: &pfs.Project{Name: pfs.DefaultProjectName},
				},
			},
		)
		if err != nil {
			return err
		}
		for {
			res, err := checkStatusClient.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
			}
			if err != nil {
				return err
			}
			if res == nil {
				return errors.Errorf("expected response, but got nil")
			}
			if len(res.GetAlerts()) != 1 {
				return errors.Errorf("expected 1 alert, but got %d", len(res.GetAlerts()))
			}
		}
	}, backoff.NewConstantBackOff(time.Millisecond*200))

	deletePipeline := func(project, pipeline string) {
		require.NoError(t, c.DeletePipeline(project, pipeline, false))
		require.NoError(t, backoff.Retry(func() error {
			_, err := c.InspectPipeline(project, pipeline, false)
			if err == nil {
				return errors.Errorf("expected pipeline to be missing, but it's still present")
			}
			return nil
		}, backoff.NewTestingBackOff()))

	}

	deletePipeline(pfs.DefaultProjectName, pipeline)
	time.Sleep(10 * time.Second)
	require.NoErrorWithinTRetry(t, time.Second*30, func() error {
		checkStatusClient, err := c.PpsAPIClient.CheckStatus(
			c.Ctx(),
			&pps.CheckStatusRequest{
				Context: &pps.CheckStatusRequest_Project{
					Project: &pfs.Project{Name: pfs.DefaultProjectName},
				},
			},
		)
		if err != nil {
			return err
		}
		for {
			res, err := checkStatusClient.Recv()
			// expect EOF as nothing should be returned
			if res != nil {
				return errors.Errorf("expected no response, but got: %v", res)
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil // success
				}
				return err
			}
		}
	}, backoff.NewConstantBackOff(time.Millisecond*200))
}

func TestUpdateFailedPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestUpdateFailedPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"imagethatdoesntexist",
		[]string{"bash"},
		[]string{"echo foo >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("1"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))

	// Wait for pod to try and pull the bad image
	require.NoErrorWithinTRetry(t, time.Second*30, func() error {
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, false)
		require.NoError(t, err)
		if pipelineInfo.State != pps.PipelineState_PIPELINE_CRASHING {
			return errors.Errorf("expected pipeline to be in CRASHING state but got: %v\n", pipelineInfo.State)
		}
		return nil
	})

	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"bash:4",
		[]string{"bash"},
		[]string{"echo bar >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		true,
	))
	require.NoErrorWithinTRetry(t, time.Second*30, func() error {
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, false)
		require.NoError(t, err)
		if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
			return errors.Errorf("expected pipeline to be in RUNNING state but got: %v\n", pipelineInfo.State)
		}
		return nil
	})

	// Sanity check run some actual data through the pipeline:
	commit, err = c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("2"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))
	_, err = c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", ""), "file", &buffer))
	require.Equal(t, "bar\n", buffer.String())
}

func TestUpdateStoppedPipeline(t *testing.T) {
	// Pipeline should be updated, but should not be restarted
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repo & pipeline
	project := tu.UniqueString("project")
	require.NoError(t, c.CreateProject(project))
	dataRepo := tu.UniqueString("TestUpdateStoppedPipeline_data")
	require.NoError(t, c.CreateRepo(project, dataRepo))
	dataCommit := client.NewCommit(project, dataRepo, "master", "")
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(project,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"cp /pfs/*/file /pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(project, dataRepo, "/*"),
		"",
		false,
	))

	commits, err := c.ListCommit(client.NewRepo(project, pipelineName), client.NewCommit(project, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(commits))

	// Add input data
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	commits, err = c.ListCommit(client.NewRepo(project, pipelineName), client.NewCommit(project, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commits))

	// Make sure the pipeline runs once (i.e. it's all the way up)
	commitInfos, err := c.WaitCommitSetAll(commits[0].Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))
	// Stop the pipeline (and confirm that it's stopped)
	require.NoError(t, c.StopPipeline(project, pipelineName))
	pipelineInfo, err := c.InspectPipeline(project, pipelineName, false)
	require.NoError(t, err)
	require.Equal(t, true, pipelineInfo.Stopped)
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err = c.InspectPipeline(project, pipelineName, false)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_PAUSED {
			return errors.Errorf("expected pipeline to be in state PAUSED, but was in %s",
				pipelineInfo.State)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	commits, err = c.ListCommit(client.NewRepo(project, pipelineName), client.NewCommit(project, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commits))

	// Update shouldn't restart it (wait for version to increment)
	require.NoError(t, c.CreatePipeline(project,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"cp /pfs/*/file /pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(project, dataRepo, "/*"),
		"",
		true,
	))
	time.Sleep(10 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err = c.InspectPipeline(project, pipelineName, false)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_PAUSED {
			return errors.Errorf("expected pipeline to be in state PAUSED, but was in %s",
				pipelineInfo.State)
		}
		if pipelineInfo.Version != 2 {
			return errors.Errorf("expected pipeline to be on v2, but was on v%d",
				pipelineInfo.Version)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	commits, err = c.ListCommit(client.NewRepo(project, pipelineName), client.NewCommit(project, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commits))

	// Create a commit (to give the pipeline pending work), then start the pipeline
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("bar"), client.WithAppendPutFile()))
	require.YesError(t, c.StartPipeline(pfs.DefaultProjectName, pipelineName)) // negative project testing
	require.NoError(t, c.StartPipeline(project, pipelineName))

	// Pipeline should start and create a job should succeed -- fix
	// https://github.com/pachyderm/pachyderm/v2/issues/3934)
	commitInfo, err := c.InspectCommit(project, pipelineName, "master", "")
	require.NoError(t, err)
	commitInfos, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))
	commits, err = c.ListCommit(client.NewRepo(project, pipelineName), client.NewCommit(project, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commits))

	var buf bytes.Buffer
	outputCommit := client.NewCommit(project, pipelineName, "master", commitInfo.Commit.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buf))
	require.Equal(t, "foobar", buf.String())
}

func TestUpdatePipelineRunningJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestUpdatePipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"sleep 1000"},
		&pps.ParallelismSpec{
			Constant: 2,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	numFiles := 50
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file-%d", i), strings.NewReader(""), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commit2, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit2, fmt.Sprintf("file-%d", i+numFiles), strings.NewReader(""), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit2.Id))

	b := backoff.NewTestingBackOff()
	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
		if err != nil {
			return err
		}
		if len(jobInfos) != 3 {
			return errors.Errorf("wrong number of jobs")
		}

		state := jobInfos[1].State
		if state != pps.JobState_JOB_RUNNING {
			return errors.Errorf("wrong state: %v for %s", state, jobInfos[1].Job.Id)
		}

		state = jobInfos[0].State
		if state != pps.JobState_JOB_RUNNING {
			return errors.Errorf("wrong state: %v for %s", state, jobInfos[0].Job.Id)
		}
		return nil
	}, b))

	// Update the pipeline. This will not create a new pipeline as reprocess
	// isn't set to true.
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"true"},
		&pps.ParallelismSpec{
			Constant: 2,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		true,
	))
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipelineName, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 4, len(jobInfos))
	require.Equal(t, pps.JobState_JOB_SUCCESS.String(), jobInfos[0].State.String())
	require.Equal(t, pps.JobState_JOB_KILLED.String(), jobInfos[1].State.String())
	require.Equal(t, pps.JobState_JOB_KILLED.String(), jobInfos[2].State.String())
	require.Equal(t, pps.JobState_JOB_SUCCESS.String(), jobInfos[3].State.String())
}

func TestManyFilesSingleCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestManyFilesSingleCommit_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")

	numFiles := 20000
	_, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.WithModifyFileClient(dataCommit, func(mfc client.ModifyFile) error {
		for i := 0; i < numFiles; i++ {
			require.NoError(t, mfc.PutFile(fmt.Sprintf("file-%d", i), strings.NewReader(""), client.WithAppendPutFile()))
		}
		return nil
	}))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))
	fileInfos, err := c.ListFileAll(dataCommit, "")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))
}

func TestManyFilesSingleOutputCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestManyFilesSingleOutputCommit_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	file := "file"

	// Setup pipeline.
	pipelineName := tu.UniqueString("TestManyFilesSingleOutputCommit")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd:   []string{"sh"},
				Stdin: []string{"while read line; do echo $line > /pfs/out/$line; done < " + path.Join("/pfs", dataRepo, file)},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		},
	)
	require.NoError(t, err)

	// Setup input.
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	numFiles := 20000
	var data string
	for i := 0; i < numFiles; i++ {
		data += strconv.Itoa(i) + "\n"
	}
	require.NoError(t, c.PutFile(commit, file, strings.NewReader(data), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", commit.Id))

	// Check results.
	jis, err := c.WaitJobSetAll(commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	fileInfos, err := c.ListFileAll(client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", ""), "")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))
}

// Checks that "time to first datum" for CreateDatum is faster than ListDatum
func BenchmarkCreateDatum(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}
	c := pachd.NewTestPachd(b)
	repo := tu.UniqueString("BenchmarkCreateDatum")
	require.NoError(b, c.CreateRepo(pfs.DefaultProjectName, repo))

	// Need multiple shards worth of files to see benefit of CreateDatum
	// compared to ListDatum
	numFiles := 5 * datum.ShardNumFiles
	commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(b, err)
	require.NoError(b, c.WithModifyFileClient(commit, func(mfc client.ModifyFile) error {
		for i := 0; i < numFiles; i++ {
			require.NoError(b, mfc.PutFile(fmt.Sprintf("file-%d", i), strings.NewReader(""), client.WithAppendPutFile()))
		}
		return nil
	}))
	require.NoError(b, c.FinishCommit(pfs.DefaultProjectName, repo, "master", ""))

	inputs := []struct {
		name  string
		input *pps.Input
	}{
		{"PFS", client.NewPFSInput(pfs.DefaultProjectName, repo, "/*")},
		{"Union", client.NewUnionInput(
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		)},
		{"Cross", client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/file-??"),
			client.NewPFSInput(pfs.DefaultProjectName, repo, "/file-???"),
		)},
		{"Join", client.NewJoinInput(
			client.NewPFSInputOpts(repo, pfs.DefaultProjectName, repo, "master", "/file-?*(??)0", "$1", "", false, false, nil),
			client.NewPFSInputOpts(repo, pfs.DefaultProjectName, repo, "master", "/file-?0(??)0", "$1", "", false, false, nil),
		)},
		// No entry for Group because CreateDatum's streaming can't improve
		// time to first datum
	}

	for _, input := range inputs {
		b.Run(input.name+"-ListDatum", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				req := &pps.ListDatumRequest{Input: input.input}
				client, err := c.PpsAPIClient.ListDatum(c.Ctx(), req)
				require.NoError(b, err)
				_, err = client.Recv()
				require.NoError(b, err)
			}
		})
	}

	for _, input := range inputs {
		b.Run(input.name+"-CreateDatum", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				client, err := c.PpsAPIClient.CreateDatum(c.Ctx())
				require.NoError(b, err)
				req := &pps.CreateDatumRequest{Body: &pps.CreateDatumRequest_Start{Start: &pps.StartCreateDatumRequest{Input: input.input}}}
				require.NoError(b, client.Send(req))
				_, err = client.Recv()
				require.NoError(b, err)
			}
		})
	}
}

func TestStopPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	// We should just have the initial output branch head commit
	commits, err := c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipelineName), client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, len(commits), 1)

	// Stop the pipeline, so it doesn't process incoming commits
	require.NoError(t, c.StopPipeline(pfs.DefaultProjectName, pipelineName))

	// Do first commit to repo
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	// wait for 10 seconds and check that no new commit has been outputted
	time.Sleep(10 * time.Second)
	commits, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipelineName), client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, len(commits), 1)

	// Restart pipeline, and make sure a new output commit is generated
	require.NoError(t, c.StartPipeline(pfs.DefaultProjectName, pipelineName))

	commits, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipelineName), client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, len(commits), 2)

	commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipelineName, "master", "")
	require.NoError(t, err)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(commitInfo.Commit, "file", &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

func TestAutoscalingStandby(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	t.Run("ChainOf10", func(t *testing.T) {
		require.NoError(t, c.DeleteAll(c.Ctx()))

		dataRepo := tu.UniqueString("TestAutoscalingStandby_data")
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
		dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")

		numPipelines := 10
		pipelines := make([]string, numPipelines)
		for i := 0; i < numPipelines; i++ {
			pipelines[i] = tu.UniqueString("TestAutoscalingStandby")
			input := dataRepo
			if i > 0 {
				input = pipelines[i-1]
			}
			_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
				&pps.CreatePipelineRequest{
					Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelines[i]),
					Transform: &pps.Transform{
						Cmd: []string{"cp", path.Join("/pfs", input, "file"), "/pfs/out/file"},
					},
					Input:       client.NewPFSInput(pfs.DefaultProjectName, input, "/*"),
					Autoscaling: true,
				},
			)
			require.NoError(t, err)
		}

		require.NoErrorWithinT(t, time.Second*90, func() error {
			_, err := c.WaitCommit(pfs.DefaultProjectName, pipelines[9], "master", "")
			return err
		})

		require.NoErrorWithinTRetry(t, time.Second*30, func() error {
			pis, err := c.ListPipeline()
			require.NoError(t, err)
			var standby int
			for _, pi := range pis {
				if pi.State == pps.PipelineState_PIPELINE_STANDBY {
					standby++
				}
			}
			if standby != numPipelines {
				return errors.Errorf("should have %d pipelines in standby, not %d", numPipelines, standby)
			}
			return nil
		})

		require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo")))
		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
		require.NoError(t, err)

		var eg errgroup.Group
		var finished bool
		eg.Go(func() error {
			_, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
			require.NoError(t, err)
			finished = true
			return nil
		})
		eg.Go(func() error {
			for !finished {
				pis, err := c.ListPipeline()
				require.NoError(t, err)
				var active int
				for _, pi := range pis {
					if pi.State != pps.PipelineState_PIPELINE_STANDBY {
						active++
					}
				}
				// We tolerate having 2 pipelines out of standby because there's
				// latency associated with entering and exiting standby.
				require.True(t, active <= 2, "active: %d", active)
			}
			return nil
		})
		require.NoError(t, eg.Wait())
	})
	t.Run("ManyCommits", func(t *testing.T) {
		require.NoError(t, c.DeleteAll(c.Ctx()))

		dataRepo := tu.UniqueString("TestAutoscalingStandby_data")
		dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
		pipeline := tu.UniqueString("TestAutoscalingStandby")
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
		_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:   []string{"sh"},
					Stdin: []string{"echo $PPS_POD_NAME >/pfs/out/pod"},
				},
				Input:       client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
				Autoscaling: true,
			},
		)
		require.NoError(t, err)
		numCommits := 100
		for i := 0; i < numCommits; i++ {
			require.NoError(t, c.PutFile(dataCommit, fmt.Sprintf("file-%d", i), strings.NewReader("foo")))
		}
		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
		require.NoError(t, err)
		commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
		require.NoError(t, err)
		nonAliasCommits := 0
		for _, v := range commitInfos {
			if v.Commit.Id != commitInfo.Commit.Id {
				nonAliasCommits++
			}
		}
		require.Equal(t, 1, nonAliasCommits)
		pod := ""
		commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline), client.NewCommit(pfs.DefaultProjectName, pipeline, "master", ""), nil, 0)
		require.NoError(t, err)
		for i, ci := range commitInfos {
			// the last commit accessed (the oldest commit) will not have any files in it
			// since it's created as a result of the spec commit
			if i == len(commitInfos)-1 {
				break
			}
			var buffer bytes.Buffer
			require.NoError(t, c.GetFile(ci.Commit, "pod", &buffer))
			if pod == "" {
				pod = buffer.String()
			} else {
				require.True(t, pod == buffer.String(), "multiple pods were used to process commits")
			}
		}
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		require.NoError(t, err)
		require.Equal(t, pps.PipelineState_PIPELINE_STANDBY.String(), pi.State.String())
	})
}

func TestStopStandbyPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString(t.Name() + "_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")

	pipeline := tu.UniqueString(t.Name())
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"/bin/bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out", dataRepo),
				},
			},
			Input:       client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			Autoscaling: true,
		},
	)
	require.NoError(t, err)

	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		require.NoError(t, err)
		if pi.State != pps.PipelineState_PIPELINE_STANDBY {
			return errors.Errorf("expected %q to be in STANDBY, but was in %s", pipeline, pi.State)
		}
		return nil
	})

	// Run the pipeline once under normal conditions. It should run and then go
	// back into standby
	require.NoError(t, c.PutFile(dataCommit, "/foo", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
		// Let pipeline run
		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
		require.NoError(t, err)
		_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
		require.NoError(t, err)
		// check ending state
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		require.NoError(t, err)
		if pi.State != pps.PipelineState_PIPELINE_STANDBY {
			return errors.Errorf("expected %q to be in STANDBY, but was in %s", pipeline, pi.State)
		}
		return nil
	})

	// Stop the pipeline...
	require.NoError(t, c.StopPipeline(pfs.DefaultProjectName, pipeline))
	require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		require.NoError(t, err)
		if pi.State != pps.PipelineState_PIPELINE_PAUSED {
			return errors.Errorf("expected %q to be in PAUSED, but was in %s", pipeline,
				pi.State)
		}
		return nil
	})
	// ...and then create several new input commits. Pipeline shouldn't run.
	for i := 0; i < 3; i++ {
		file := fmt.Sprintf("bar-%d", i)
		require.NoError(t, c.PutFile(dataCommit, "/"+file, strings.NewReader(file), client.WithAppendPutFile()))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	for ctx.Err() == nil {
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		require.NoError(t, err)
		require.NotEqual(t, pps.PipelineState_PIPELINE_RUNNING, pi.State)
	}
	cancel()

	// Start pipeline--it should run and then enter standby
	require.NoError(t, c.StartPipeline(pfs.DefaultProjectName, pipeline))
	require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
		// Let pipeline run
		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
		require.NoError(t, err)
		_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
		require.NoError(t, err)
		// check ending state
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		require.NoError(t, err)
		if pi.State != pps.PipelineState_PIPELINE_STANDBY {
			return errors.Errorf("expected %q to be in STANDBY, but was in %s", pipeline, pi.State)
		}
		return nil
	})

	// Finally, check that there's only three output commits
	commitInfos, err := c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline), client.NewCommit(pfs.DefaultProjectName, pipeline, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
}

func TestPipelineEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)

	// make a secret to reference
	k := tu.GetKubeClient(t)
	secretName := tu.UniqueString("test-secret")
	_, err := k.CoreV1().Secrets(ns).Create(
		context.Background(),
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
			},
			Data: map[string][]byte{
				"foo": []byte("foo\n"),
			},
		},
		metav1.CreateOptions{},
	)
	require.NoError(t, err)

	// create repos
	dataRepo := tu.UniqueString("TestPipelineEnv_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	input := client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*")
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"sh"},
				Stdin: []string{
					"ls /var/secret",
					"cat /var/secret/foo > /pfs/out/foo",
					"echo $bar> /pfs/out/bar",
					"echo $foo> /pfs/out/foo_env",
					fmt.Sprintf("echo $%s >/pfs/out/job_id", client.JobIDEnv),
					fmt.Sprintf("echo $%s >/pfs/out/output_commit_id", client.OutputCommitIDEnv),
					fmt.Sprintf("echo $%s >/pfs/out/input", dataRepo),
					fmt.Sprintf("echo $%s_COMMIT >/pfs/out/input_commit", dataRepo),
					fmt.Sprintf("echo $%s >/pfs/out/datum_id", client.DatumIDEnv),
				},
				Env: map[string]string{"bar": "bar"},
				Secrets: []*pps.SecretMount{
					{
						Name:      secretName,
						Key:       "foo",
						MountPath: "/var/secret",
						EnvVar:    "foo",
					},
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: input,
		})
	require.NoError(t, err)

	// Do first commit to repo
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)

	jis, err := c.WaitJobSetAll(commitInfo.Commit.Id, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(jis[0].OutputCommit, "foo", &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(jis[0].OutputCommit, "foo_env", &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(jis[0].OutputCommit, "bar", &buffer))
	require.Equal(t, "bar\n", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(jis[0].OutputCommit, "job_id", &buffer))
	require.Equal(t, fmt.Sprintf("%s\n", jis[0].Job.Id), buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(jis[0].OutputCommit, "output_commit_id", &buffer))
	require.Equal(t, fmt.Sprintf("%s\n", jis[0].OutputCommit.Id), buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(jis[0].OutputCommit, "input", &buffer))
	require.Equal(t, fmt.Sprintf("/pfs/%s/file\n", dataRepo), buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(jis[0].OutputCommit, "input_commit", &buffer))
	require.Equal(t, fmt.Sprintf("%s\n", jis[0].Details.Input.Pfs.Commit), buffer.String())
	datumInfos, err := c.ListDatumInputAll(input)
	require.NoError(t, err)
	buffer.Reset()
	require.NoError(t, c.GetFile(jis[0].OutputCommit, "datum_id", &buffer))
	require.Equal(t, fmt.Sprintf("%s\n", datumInfos[0].Datum.Id), buffer.String())
}

func TestPipelineWithFullObjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	// Do first commit to repo
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))
	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	var buffer bytes.Buffer
	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", commit1.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buffer))
	require.Equal(t, "foo\n", buffer.String())

	// Do second commit to repo
	commit2, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit2, "file", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit2.Id))
	commitInfos, err = c.WaitCommitSetAll(commit2.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	buffer = bytes.Buffer{}
	outputCommit = client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", commit2.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())
}

func TestPipelineWithExistingInputCommits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// Do first commit to repo
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	// Do second commit to repo
	commit2, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit2, "file", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit2.Id))

	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipelineName, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	buffer := bytes.Buffer{}
	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", commitInfo.Commit.Id)
	require.NoError(t, c.GetFile(outputCommit, "file", &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())

	// Check that one output commit is created (processing the inputs' head commits)
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipelineName), client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
}

func TestPipelineThatSymlinks(t *testing.T) {
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// create repos
	dataRepo := tu.UniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// Do first commit to repo
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "foo", strings.NewReader("foo")))
	require.NoError(t, c.PutFile(commit, "dir1/bar", strings.NewReader("bar")))
	require.NoError(t, c.PutFile(commit, "dir2/foo", strings.NewReader("foo")))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
	check := func(input *pps.Input) {
		pipelineName := tu.UniqueString("pipeline")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipelineName,
			"",
			[]string{"bash"},
			[]string{
				// Symlinks to input files
				fmt.Sprintf("ln -s /pfs/%s/foo /pfs/out/foo", dataRepo),
				fmt.Sprintf("ln -s /pfs/%s/dir1/bar /pfs/out/bar", dataRepo),
				"mkdir /pfs/out/dir",
				fmt.Sprintf("ln -s /pfs/%s/dir2 /pfs/out/dir/dir2", dataRepo),
				// Symlinks to external files
				"echo buzz > /tmp/buzz",
				"ln -s /tmp/buzz /pfs/out/buzz",
				"mkdir /tmp/dir3",
				"mkdir /tmp/dir3/dir4",
				"echo foobar > /tmp/dir3/dir4/foobar",
				"ln -s /tmp/dir3 /pfs/out/dir3",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			input,
			"",
			false,
		))
		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipelineName, "master", "")
		require.NoError(t, err)
		commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
		require.NoError(t, err)
		require.Equal(t, 4, len(commitInfos))
		// Check that the output files are identical to the input files.
		buffer := bytes.Buffer{}
		outputCommit := client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", commitInfo.Commit.Id)
		require.NoError(t, c.GetFile(outputCommit, "foo", &buffer))
		require.Equal(t, "foo", buffer.String())
		buffer.Reset()
		require.NoError(t, c.GetFile(outputCommit, "bar", &buffer))
		require.Equal(t, "bar", buffer.String())
		buffer.Reset()
		require.NoError(t, c.GetFile(outputCommit, "dir/dir2/foo", &buffer))
		require.Equal(t, "foo", buffer.String())
		buffer.Reset()
		require.NoError(t, c.GetFile(outputCommit, "buzz", &buffer))
		require.Equal(t, "buzz\n", buffer.String())
		buffer.Reset()
		require.NoError(t, c.GetFile(outputCommit, "dir3/dir4/foobar", &buffer))
		require.Equal(t, "foobar\n", buffer.String())
	}

	// Check normal pipeline.
	input := client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/")
	check(input)
	// Check pipeline with empty files.
	input = client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/")
	input.Pfs.EmptyFiles = true
	check(input)
	// Check pipeline with lazy files.
	input = client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/")
	input.Pfs.Lazy = true
	check(input)
}

// TestChainedPipelines tracks https://github.com/pachyderm/pachyderm/v2/issues/797
// DAG:
//
// A
// |
// B  D
// | /
// C
func TestChainedPipelines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	aRepo := tu.UniqueString("A")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, aRepo))

	dRepo := tu.UniqueString("D")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dRepo))

	aCommit, err := c.StartCommit(pfs.DefaultProjectName, aRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(aCommit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, aRepo, "master", ""))

	dCommit, err := c.StartCommit(pfs.DefaultProjectName, dRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(dCommit, "file", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dRepo, "master", ""))

	bPipeline := tu.UniqueString("B")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, aRepo, "/"),
		"",
		false,
	))

	cPipeline := tu.UniqueString("C")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		cPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/file /pfs/out/bFile", bPipeline),
			fmt.Sprintf("cp /pfs/%s/file /pfs/out/dFile", dRepo)},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, bPipeline, "/"),
			client.NewPFSInput(pfs.DefaultProjectName, dRepo, "/"),
		),
		"",
		false,
	))

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, cPipeline, "master", "")
	require.NoError(t, err)

	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 7, len(commitInfos))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfo.Commit, "bFile", &buf))
	require.Equal(t, "foo\n", buf.String())

	buf.Reset()
	require.NoError(t, c.GetFile(commitInfo.Commit, "dFile", &buf))
	require.Equal(t, "bar\n", buf.String())
}

// DAG:
//
// A
// |
// B  E
// | /
// C
// |
// D
func TestChainedPipelinesNoDelay(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	aRepo := tu.UniqueString("A")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, aRepo))

	eRepo := tu.UniqueString("E")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, eRepo))

	aCommit, err := c.StartCommit(pfs.DefaultProjectName, aRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(aCommit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, aRepo, "", "master"))

	eCommit, err := c.StartCommit(pfs.DefaultProjectName, eRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(eCommit, "file", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, eRepo, "master", ""))

	bPipeline := tu.UniqueString("B")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, aRepo, "/"),
		"",
		false,
	))

	cPipeline := tu.UniqueString("C")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		cPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/file /pfs/out/bFile", bPipeline),
			fmt.Sprintf("cp /pfs/%s/file /pfs/out/eFile", eRepo)},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, bPipeline, "/"),
			client.NewPFSInput(pfs.DefaultProjectName, eRepo, "/"),
		),
		"",
		false,
	))

	dPipeline := tu.UniqueString("D")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		dPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/bFile /pfs/out/bFile", cPipeline),
			fmt.Sprintf("cp /pfs/%s/eFile /pfs/out/eFile", cPipeline)},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, cPipeline, "/"),
		"",
		false,
	))

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dPipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 9, len(commitInfos))
	eCommit2, err := c.StartCommit(pfs.DefaultProjectName, eRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(eCommit2, "file", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, eRepo, "master", ""))

	commitInfos, err = c.WaitCommitSetAll(eCommit2.Id)
	require.NoError(t, err)
	// here we have one more commit for e.meta
	require.Equal(t, 10, len(commitInfos))

	// Get number of jobs triggered in pipeline D
	jobInfos, err := c.ListJob(pfs.DefaultProjectName, dPipeline, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
}

// TestMetaAlias tracks https://github.com/pachyderm/pachyderm/pull/8116
// This test is so we don't regress a problem; where a pipeline in a DAG is started,
// the pipeline should not get stuck due to a missing Meta Commit. with 8116 the meta
// repo should get the alias commit.
// DAG:
//
// A
// |
// B
// |
// C
func TestStartInternalPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	aRepo := tu.UniqueString("A")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, aRepo))
	aCommit, err := c.StartCommit(pfs.DefaultProjectName, aRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(aCommit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, aRepo, "master", ""))
	bPipeline := tu.UniqueString("B")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, aRepo, "/"),
		"",
		false,
	))
	cPipeline := tu.UniqueString("C")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		cPipeline,
		"",
		[]string{"cp", path.Join("/pfs", bPipeline, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, bPipeline, "/"),
		"",
		false,
	))
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, cPipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 6, len(commitInfos))
	// Stop and Start a pipeline was orginal trigger of bug so "reproduce" it here.
	require.NoError(t, c.StopPipeline(pfs.DefaultProjectName, bPipeline))
	require.NoError(t, c.StartPipeline(pfs.DefaultProjectName, bPipeline))
	// C's commit should be the same as b's meta commit
	cCommits, err := c.ListCommit(client.NewRepo(pfs.DefaultProjectName, cPipeline), nil, nil, 0)
	require.NoError(t, err)
	listClient, err := c.PfsAPIClient.ListCommit(c.Ctx(), &pfs.ListCommitRequest{
		Repo: client.NewSystemRepo(pfs.DefaultProjectName, bPipeline, pfs.MetaRepoType),
		All:  true,
	})
	require.NoError(t, err)
	allBmetaCommits, err := grpcutil.Collect[*pfs.CommitInfo](listClient, 1000)
	require.NoError(t, err)
	require.Equal(t, allBmetaCommits[0].Commit.Id, cCommits[0].Commit.Id)
}

func TestJobDeletion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	project := tu.UniqueString("project")
	require.NoError(t, c.CreateProject(project))
	// create repos
	dataRepo := tu.UniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(project,
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		"",
		false,
	))

	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	_, err = c.WaitCommitSetAll(commit.Id)
	require.NoError(t, err)

	err = c.DeleteJob(pfs.DefaultProjectName, pipelineName, commit.Id)
	require.YesError(t, err) // Wrong project should error
	err = c.DeleteJob(project, pipelineName, commit.Id)
	require.NoError(t, err)
}

func TestStopJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create project
	project := tu.UniqueString("project")
	require.NoError(t, c.CreateProject((project)))
	// create repos
	dataRepo := tu.UniqueString("TestStopJob")
	require.NoError(t, c.CreateRepo(project, dataRepo))

	// create pipeline
	pipelineName := tu.UniqueString("pipeline-stop-job")
	require.NoError(t, c.CreatePipeline(project,
		pipelineName,
		"",
		[]string{"sleep", "10"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(project, dataRepo, "/"),
		"",
		false,
	))

	// Create three input commits to trigger jobs.
	// We will stop the second job midway through, and assert that the
	// last job finishes.
	commit1, err := c.StartCommit(project, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(project, dataRepo, "", commit1.Id))

	commit2, err := c.StartCommit(project, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit2, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(project, dataRepo, "", commit2.Id))

	commit3, err := c.StartCommit(project, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit3, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(project, dataRepo, "", commit3.Id))

	jobInfos, err := c.ListJob(project, pipelineName, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 4, len(jobInfos))
	require.Equal(t, commit3.Id, jobInfos[0].Job.Id)
	require.Equal(t, commit2.Id, jobInfos[1].Job.Id)
	require.Equal(t, commit1.Id, jobInfos[2].Job.Id)

	require.NoError(t, backoff.Retry(func() error {
		jobInfo, err := c.InspectJob(project, pipelineName, commit1.Id, false)
		require.NoError(t, err)
		if jobInfo.State != pps.JobState_JOB_RUNNING {
			return errors.Errorf("first job has the wrong state %v expected %v", jobInfo.State, pps.JobState_JOB_RUNNING)
		}
		return nil
	}, backoff.NewConstantBackOff(2*time.Second).For(90*time.Second)))

	// Now stop the second job
	err = c.StopJob(project, pipelineName, commit2.Id)
	require.NoError(t, err)
	jobInfo, err := c.WaitJob(project, pipelineName, commit2.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_KILLED, jobInfo.State)

	// Check that the third job completes
	jobInfo, err = c.WaitJob(project, pipelineName, commit3.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
}

func testGetLogs(t *testing.T, useLoki bool) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	var opts []minikubetestenv.Option
	if !useLoki {
		opts = append(opts, minikubetestenv.SkipLokiOption)
	}
	c, _ := minikubetestenv.AcquireCluster(t, opts...)
	iter := c.GetLogs(pfs.DefaultProjectName, "", "", nil, "", false, false, 0)
	for iter.Next() {
	}
	require.NoError(t, iter.Err())
	// create repos
	dataRepo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"sh"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/file /pfs/out/file", dataRepo),
					"echo foo",
					"echo %s", // %s tests a formatting bug we had (#2729)
				},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 4,
			},
		})
	require.NoError(t, err)

	// Commit data to repo and flush commit
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))
	_, err = c.WaitJobSetAll(commit.Id, false)
	require.NoError(t, err)

	// Get logs from pipeline, using a pipeline that doesn't exist. There should
	// be an error
	iter = c.GetLogs(pfs.DefaultProjectName, "__DOES_NOT_EXIST__", "", nil, "", false, false, 0)
	require.False(t, iter.Next())
	require.YesError(t, iter.Err())
	require.Matches(t, "could not get", iter.Err().Error())

	// Get logs from pipeline, using a job that doesn't exist. There should
	// be an error
	iter = c.GetLogs(pfs.DefaultProjectName, pipelineName, "__DOES_NOT_EXIST__", nil, "", false, false, 0)
	require.False(t, iter.Next())
	require.YesError(t, iter.Err())
	require.Matches(t, "could not get", iter.Err().Error())

	// This is put in a backoff because there's the possibility that pod was
	// evicted from k8s and is being re-initialized, in which case `GetLogs`
	// will appropriately fail. With the loki logging backend enabled the
	// eviction worry goes away, but is replaced with there being a window when
	// Loki hasn't scraped the logs yet so they don't show up.
	require.NoError(t, backoff.Retry(func() error {
		// Get logs from pipeline, using pipeline
		iter = c.GetLogs(pfs.DefaultProjectName, pipelineName, "", nil, "", false, false, 0)
		var numLogs, totalLogs int
		for iter.Next() {
			totalLogs++
			if !iter.Message().User {
				continue
			}
			numLogs++
			require.True(t, iter.Message().Message != "")
			require.False(t, strings.Contains(iter.Message().Message, "MISSING"), iter.Message().Message)
		}
		if numLogs < 2 {
			return errors.Errorf("didn't get enough log lines (%d total logs, %d non-user logs)", totalLogs, numLogs)
		}
		if err := iter.Err(); err != nil {
			return err
		}

		// Get logs from pipeline, using job
		// (1) Get job ID, from pipeline that just ran
		jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
		if err != nil {
			return err
		}
		require.Equal(t, 2, len(jobInfos))
		// (2) Get logs using extracted job ID
		// wait for logs to be collected
		time.Sleep(10 * time.Second)
		iter = c.GetLogs(pfs.DefaultProjectName, pipelineName, jobInfos[0].Job.Id, nil, "", false, false, 0)
		numLogs = 0
		for iter.Next() {
			numLogs++
			require.True(t, iter.Message().Message != "")
		}
		// Make sure that we've seen some logs
		if err = iter.Err(); err != nil {
			return err
		}
		require.True(t, numLogs > 0)

		// Get logs for datums but don't specify pipeline or job. These should error
		iter = c.GetLogs(pfs.DefaultProjectName, "", "", []string{"/foo"}, "", false, false, 0)
		require.False(t, iter.Next())
		require.YesError(t, iter.Err())

		dis, err := c.ListDatumAll(pfs.DefaultProjectName, jobInfos[0].Job.Pipeline.Name, jobInfos[0].Job.Id)
		if err != nil {
			return err
		}
		require.True(t, len(dis) > 0)
		iter = c.GetLogs(pfs.DefaultProjectName, "", "", nil, dis[0].Datum.Id, false, false, 0)
		require.False(t, iter.Next())
		require.YesError(t, iter.Err())

		// Filter logs based on input (using file that exists). Get logs using file
		// path, hex hash, and base64 hash, and make sure you get the same log lines
		fileInfo, err := c.InspectFile(dataCommit, "/file")
		if err != nil {
			return err
		}

		pathLog := c.GetLogs(pfs.DefaultProjectName, pipelineName, jobInfos[0].Job.Id, []string{"/file"}, "", false, false, 0)

		base64Hash := "kstrTGrFE58QWlxEpCRBt3aT8NJPNY0rso6XK7a4+wM="
		require.Equal(t, base64Hash, base64.StdEncoding.EncodeToString(fileInfo.Hash))
		base64Log := c.GetLogs(pfs.DefaultProjectName, pipelineName, jobInfos[0].Job.Id, []string{base64Hash}, "", false, false, 0)

		numLogs = 0
		for {
			havePathLog, haveBase64Log := pathLog.Next(), base64Log.Next()
			if havePathLog != haveBase64Log {
				return errors.Errorf("Unequal log lengths")
			}
			if !havePathLog {
				break
			}
			numLogs++
			if pathLog.Message().Message != base64Log.Message().Message {
				return errors.Errorf(
					"unequal logs, pathLogs: \"%s\" base64Log: \"%s\"",
					pathLog.Message().Message,
					base64Log.Message().Message)
			}
		}
		for _, logsiter := range []*client.LogsIter{pathLog, base64Log} {
			if logsiter.Err() != nil {
				return logsiter.Err()
			}
		}
		if numLogs == 0 {
			return errors.Errorf("no logs found")
		}

		// Filter logs based on input (using file that doesn't exist). There should
		// be no logs
		iter = c.GetLogs(pfs.DefaultProjectName, pipelineName, jobInfos[0].Job.Id, []string{"__DOES_NOT_EXIST__"}, "", false, false, 0)
		require.False(t, iter.Next())
		if err = iter.Err(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		iter = c.WithCtx(ctx).GetLogs(pfs.DefaultProjectName, pipelineName, "", nil, "", false, false, 0)
		numLogs = 0
		for iter.Next() {
			numLogs++
			if numLogs == 8 {
				// Do another commit so there's logs to receive with follow
				_, err = c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
				if err != nil {
					return err
				}
				if err := c.PutFile(dataCommit, "file", strings.NewReader("bar\n"), client.WithAppendPutFile()); err != nil {
					return err
				}
				if err = c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""); err != nil {
					return err
				}
			}
			require.True(t, iter.Message().Message != "")
			if numLogs == 16 {
				break
			}
		}
		if err := iter.Err(); err != nil {
			return err
		}

		time.Sleep(time.Second * 30)

		numLogs = 0
		iter = c.WithCtx(ctx).GetLogs(pfs.DefaultProjectName, pipelineName, "", nil, "", false, false, 15*time.Second)
		for iter.Next() {
			numLogs++
		}
		if err := iter.Err(); err != nil {
			return err
		}
		if numLogs != 0 {
			return errors.Errorf("shouldn't return logs due to since time")
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestGetLogsWithoutLoki(t *testing.T) {
	testGetLogs(t, false)
}

func TestGetLogs(t *testing.T) {
	testGetLogs(t, true)
}

func TestManyLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	// create pipeline
	numLogs := 10000
	pipelineName := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"sh"},
				Stdin: []string{
					"i=0",
					fmt.Sprintf("while [ \"$i\" -lt %d ]", numLogs),
					"do",
					"	echo $i",
					"	i=`expr $i + 1`",
					"done",
				},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipelineName, "master", "")
	require.NoError(t, err)

	jis, err := c.WaitJobSetAll(commitInfo.Commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))

	require.NoErrorWithinTRetry(t, 2*time.Minute, func() error {
		iter := c.GetLogs(pfs.DefaultProjectName, pipelineName, jis[0].Job.Id, nil, "", false, false, 0)
		var logsReceived, totalLines int
		for iter.Next() {
			totalLines++
			if iter.Message().User {
				logsReceived++
			}
		}
		if iter.Err() != nil {
			return iter.Err()
		}
		if numLogs != logsReceived {
			return errors.Errorf("received: %d user log lines, expected: %d (%d total lines)", logsReceived, numLogs, totalLines)
		}
		return nil
	})
}

func TestLokiLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateEnterprise(t, c)
	// create repos
	dataRepo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	numFiles := 10
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(dataCommit, fmt.Sprintf("file-%d", i), strings.NewReader("foo\n"), client.WithAppendPutFile()))
	}

	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"echo", "foo"},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipelineName, "master", "")
	require.NoError(t, err)

	jis, err := c.WaitJobSetAll(commitInfo.Commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))

	// Follow the logs the make sure we get enough foos
	require.NoErrorWithinT(t, 2*time.Minute, func() error {
		iter := c.GetLogsLoki(pipelineName, jis[0].Job.Id, nil, "", false, true, 0)
		foundFoos := 0
		for iter.Next() {
			if strings.Contains(iter.Message().Message, "foo") {
				foundFoos++
				if foundFoos == numFiles {
					break
				}
			}
		}
		if iter.Err() != nil {
			return iter.Err()
		}
		return nil
	})

	time.Sleep(5 * time.Second)

	iter := c.GetLogsLoki(pipelineName, jis[0].Job.Id, nil, "", false, false, 0)
	foundFoos := 0
	for iter.Next() {
		if strings.Contains(iter.Message().Message, "foo") {
			foundFoos++
		}
	}
	require.NoError(t, iter.Err())
	require.Equal(t, numFiles, foundFoos, "didn't receive enough log lines containing foo")
}

func TestAllDatumsAreProcessed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo1 := tu.UniqueString("TestAllDatumsAreProcessed_data1")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo1))
	dataRepo2 := tu.UniqueString("TestAllDatumsAreProcessed_data2")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo2))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo1, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file1", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.PutFile(commit1, "file2", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo1, "master", ""))

	commit2, err := c.StartCommit(pfs.DefaultProjectName, dataRepo2, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit2, "file1", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.PutFile(commit2, "file2", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo2, "master", ""))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/%s/* /pfs/%s/* > /pfs/out/file", dataRepo1, dataRepo2),
		},
		nil,
		client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo1, "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo2, "/*"),
		),
		"",
		false,
	))

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 5, len(commitInfos))

	var buf bytes.Buffer
	rc, err := c.GetFileTAR(commitInfo.Commit, "file")
	require.NoError(t, err)
	defer rc.Close()
	require.NoError(t, tarutil.ConcatFileContent(&buf, rc))
	// should be 8 because each file gets copied twice due to cross product
	require.Equal(t, strings.Repeat("foo\n", 8), buf.String())
}

func TestDatumStatusRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	project := tu.UniqueString("PROJECT")
	require.NoError(t, c.CreateProject(project))

	dataRepo := tu.UniqueString("dataRepo")
	require.NoError(t, c.CreateRepo(project, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	// This pipeline sleeps for 30 secs per datum
	require.NoError(t, c.CreatePipeline(project,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"sleep 30",
		},
		nil,
		client.NewPFSInput(project, dataRepo, "/*"),
		"",
		false,
	))

	commit1, err := c.StartCommit(project, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(project, dataRepo, "", commit1.Id))

	// get the job status
	jobs, err := c.ListJob(project, pipeline, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))

	var datumStarted time.Time
	// checkStatus waits for 'pipeline' to start and makes sure that each time
	// it's called, the datum being processes was started at a new and later time
	// (than the last time checkStatus was called)
	checkStatus := func() {
		require.NoErrorWithinTRetryConstant(t, time.Minute, func() error {
			jobInfo, err := c.InspectJob(project, pipeline, commit1.Id, true)

			require.NoError(t, err)
			if len(jobInfo.Details.WorkerStatus) == 0 {
				return errors.Errorf("no worker statuses")
			}
			workerStatus := jobInfo.Details.WorkerStatus[0]
			if workerStatus.JobId == jobInfo.Job.Id {
				if workerStatus.DatumStatus == nil {
					return errors.Errorf("no datum status")
				}
				// The first time this function is called, datumStarted is zero
				// so `Before` is true for any non-zero time.
				started := workerStatus.DatumStatus.Started.AsTime()
				if !datumStarted.Before(started) {
					return errors.Errorf("%v ≮ %v", datumStarted, started)
				}
				datumStarted = started
				return nil
			}
			return errors.Errorf("worker status from wrong job")
		}, 500*time.Millisecond) // We need to check quickly. If we wait too long the datum might finish before we can restart it.
	}
	checkStatus()
	require.NoError(t, c.RestartDatum(project, pipeline, commit1.Id, []string{"/file"}))
	checkStatus()

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))
}

// TestSystemResourceRequest doesn't create any jobs or pipelines, it
// just makes sure that when pachyderm is deployed, we give pachd,
// and etcd default resource requests. This prevents them from overloading
// nodes and getting evicted, which can slow down or break a cluster.
func TestSystemResourceRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	_, ns := minikubetestenv.AcquireCluster(t)
	kubeClient := tu.GetKubeClient(t)

	// Expected resource requests for pachyderm system pods:
	defaultLocalMem := map[string]string{
		"pachd": "512M",
		"etcd":  "512M",
	}
	defaultLocalCPU := map[string]string{
		"pachd": "250m",
		"etcd":  "250m",
	}
	defaultCloudMem := map[string]string{
		"pachd": "3G",
		"etcd":  "2G",
	}
	defaultCloudCPU := map[string]string{
		"pachd": "1",
		"etcd":  "1",
	}
	// Get Pod info for 'app' from k8s
	var c v1.Container
	for _, app := range []string{"pachd", "etcd"} {
		err := backoff.Retry(func() error {
			podList, err := kubeClient.CoreV1().Pods(ns).List(
				context.Background(),
				metav1.ListOptions{
					LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
						map[string]string{"app": app, "suite": "pachyderm"},
					)),
				})
			if err != nil {
				return errors.EnsureStack(err)
			}
			if len(podList.Items) < 1 {
				return errors.Errorf("could not find pod for %s", app) // retry
			}
			c = podList.Items[0].Spec.Containers[0]
			return nil
		}, backoff.NewTestingBackOff())
		require.NoError(t, err)

		// Make sure the pod's container has resource requests
		cpu, ok := c.Resources.Requests[v1.ResourceCPU]
		require.True(t, ok, "could not get CPU request for "+app)
		require.True(t, cpu.String() == defaultLocalCPU[app] ||
			cpu.String() == defaultCloudCPU[app])
		mem, ok := c.Resources.Requests[v1.ResourceMemory]
		require.True(t, ok, "could not get memory request for "+app)
		require.True(t, mem.String() == defaultLocalMem[app] ||
			mem.String() == defaultCloudMem[app])
	}
}

// TestPipelineResourceRequest creates a pipeline with a resource request, and
// makes sure that's passed to k8s (by inspecting the pipeline's pods)
func TestPipelineResourceRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("repo")
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	// Resources are not yet in client.CreatePipeline() (we may add them later)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			ResourceRequests: &pps.ResourceSpec{
				Memory: "100M",
				Cpu:    0.5,
				Disk:   "10M",
			},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)

	// Get info about the pipeline pods from k8s & check for resources
	pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, false)
	require.NoError(t, err)

	var container v1.Container
	kubeClient := tu.GetKubeClient(t)
	require.NoError(t, backoff.Retry(func() error {
		podList, err := kubeClient.CoreV1().Pods(ns).List(
			context.Background(),
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{
						"app":             "pipeline",
						"pipelineName":    pipelineInfo.Pipeline.Name,
						"pipelineVersion": fmt.Sprint(pipelineInfo.Version),
					},
				)),
			})
		if err != nil {
			return errors.EnsureStack(err) // retry
		}
		if len(podList.Items) != 1 || len(podList.Items[0].Spec.Containers) == 0 {
			return errors.Errorf("could not find single container for pipeline %s", pipelineInfo.Pipeline.Name)
		}
		container = podList.Items[0].Spec.Containers[0]
		return nil // no more retries
	}, backoff.NewTestingBackOff()))
	// Make sure a CPU and Memory request are both set
	cpu, ok := container.Resources.Requests[v1.ResourceCPU]
	require.True(t, ok)
	require.Equal(t, "500m", cpu.String())
	mem, ok := container.Resources.Requests[v1.ResourceMemory]
	require.True(t, ok)
	require.Equal(t, "100M", mem.String())
	disk, ok := container.Resources.Requests[v1.ResourceEphemeralStorage]
	require.True(t, ok)
	require.Equal(t, "10M", disk.String())
}

func TestPipelineResourceLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPipelineResourceLimit")
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	// Resources are not yet in client.CreatePipeline() (we may add them later)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			ResourceLimits: &pps.ResourceSpec{
				Memory: "100M",
				Cpu:    0.5,
			},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)

	// Get info about the pipeline pods from k8s & check for resources
	pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, false)
	require.NoError(t, err)

	var container v1.Container
	kubeClient := tu.GetKubeClient(t)
	err = backoff.Retry(func() error {
		podList, err := kubeClient.CoreV1().Pods(ns).List(
			context.Background(),
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{
						"app":             "pipeline",
						"pipelineName":    pipelineInfo.Pipeline.Name,
						"pipelineVersion": fmt.Sprint(pipelineInfo.Version),
						"suite":           "pachyderm"},
				)),
			})
		if err != nil {
			return errors.EnsureStack(err) // retry
		}
		if len(podList.Items) != 1 || len(podList.Items[0].Spec.Containers) == 0 {
			return errors.Errorf("could not find single container for pipeline %s", pipelineInfo.Pipeline.Name)
		}
		container = podList.Items[0].Spec.Containers[0]
		return nil // no more retries
	}, backoff.NewTestingBackOff())
	require.NoError(t, err)
	// Make sure a CPU and Memory request are both set
	cpu, ok := container.Resources.Limits[v1.ResourceCPU]
	require.True(t, ok)
	require.Equal(t, "500m", cpu.String())
	mem, ok := container.Resources.Limits[v1.ResourceMemory]
	require.True(t, ok)
	require.Equal(t, "100M", mem.String())
}

func TestPipelineResourceLimitDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPipelineResourceLimit")
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	// Resources are not yet in client.CreatePipeline() (we may add them later)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)

	// Get info about the pipeline pods from k8s & check for resources
	pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, false)
	require.NoError(t, err)

	var container v1.Container
	kubeClient := tu.GetKubeClient(t)
	err = backoff.Retry(func() error {
		podList, err := kubeClient.CoreV1().Pods(ns).List(
			context.Background(),
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{
						"app":             "pipeline",
						"pipelineName":    pipelineInfo.Pipeline.Name,
						"pipelineVersion": fmt.Sprint(pipelineInfo.Version),
						"suite":           "pachyderm"},
				)),
			})
		if err != nil {
			return errors.EnsureStack(err) // retry
		}
		if len(podList.Items) != 1 || len(podList.Items[0].Spec.Containers) == 0 {
			return errors.Errorf("could not find single container for pipeline %s", pipelineInfo.Pipeline.Name)
		}
		container = podList.Items[0].Spec.Containers[0]
		return nil // no more retries
	}, backoff.NewTestingBackOff())
	require.NoError(t, err)
	require.Nil(t, container.Resources.Limits)
}

func TestPipelinePartialResourceRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPipelinePartialResourceRequest")
	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	// Resources are not yet in client.CreatePipeline() (we may add them later)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, fmt.Sprintf("%s-%d", pipelineName, 0)),
			Transform: &pps.Transform{
				Cmd: []string{"true"},
			},
			ResourceRequests: &pps.ResourceSpec{
				Cpu:    0.25,
				Memory: "100M",
			},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, fmt.Sprintf("%s-%d", pipelineName, 1)),
			Transform: &pps.Transform{
				Cmd: []string{"true"},
			},
			ResourceRequests: &pps.ResourceSpec{
				Memory: "100M",
			},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, fmt.Sprintf("%s-%d", pipelineName, 2)),
			Transform: &pps.Transform{
				Cmd: []string{"true"},
			},
			ResourceRequests: &pps.ResourceSpec{},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		for i := 0; i < 3; i++ {
			pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, fmt.Sprintf("%s-%d", pipelineName, i), false)
			require.NoError(t, err)
			if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
				return errors.Errorf("pipeline not in running state")
			}
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestPipelineCrashing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPipelineCrashing_data")
	pipelineName := tu.UniqueString("TestPipelineCrashing_pipeline")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	create := func(gpu bool) error {
		req := pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},

			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
			Update: true,
		}
		if gpu {
			req.ResourceLimits = &pps.ResourceSpec{
				Gpu: &pps.GPUSpec{
					Type:   "nvidia.com/gpu",
					Number: 1,
				},
			}
		}
		_, err := c.PpsAPIClient.CreatePipeline(context.Background(), &req)
		return errors.EnsureStack(err)
	}
	require.NoError(t, create(true))

	require.NoError(t, backoff.Retry(func() error {
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, false)
		require.NoError(t, err)
		if pi.State != pps.PipelineState_PIPELINE_CRASHING {
			return errors.Errorf("pipeline in wrong state: %s", pi.State.String())
		}
		require.True(t, pi.Reason != "")
		return nil
	}, backoff.NewTestingBackOff()))

	// recreate with no gpu
	require.NoError(t, create(false))
	// wait for pipeline to restart and enter running
	require.NoError(t, backoff.Retry(func() error {
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, false)
		require.NoError(t, err)
		if pi.State != pps.PipelineState_PIPELINE_RUNNING {
			return errors.Errorf("pipeline in wrong state: %s", pi.State.String())
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestPodOpts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPodSpecOpts_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	t.Run("Validation", func(t *testing.T) {
		pipelineName := tu.UniqueString("TestPodSpecOpts")
		_, err := c.PpsAPIClient.CreatePipeline(
			context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
				Transform: &pps.Transform{
					Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
				},
				Input: &pps.Input{
					Pfs: &pps.PFSInput{
						Repo:   dataRepo,
						Branch: "master",
						Glob:   "/*",
					},
				},
				PodSpec: "not-json",
			})
		require.YesError(t, err)
		_, err = c.PpsAPIClient.CreatePipeline(
			context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
				Transform: &pps.Transform{
					Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
				},
				Input: &pps.Input{
					Pfs: &pps.PFSInput{
						Repo:   dataRepo,
						Branch: "master",
						Glob:   "/*",
					},
				},
				PodPatch: "also-not-json",
			})
		require.YesError(t, err)
	})
	t.Run("Spec", func(t *testing.T) {
		pipelineName := tu.UniqueString("TestPodSpecOpts")
		_, err := c.PpsAPIClient.CreatePipeline(
			context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
				Transform: &pps.Transform{
					Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
				},
				ParallelismSpec: &pps.ParallelismSpec{
					Constant: 1,
				},
				Input: &pps.Input{
					Pfs: &pps.PFSInput{
						Repo:   dataRepo,
						Branch: "master",
						Glob:   "/*",
					},
				},
				SchedulingSpec: &pps.SchedulingSpec{
					// This NodeSelector will cause the worker pod to fail to
					// schedule, but the test can still pass because we just check
					// for values on the pod, it doesn't need to actually come up.
					NodeSelector: map[string]string{
						"foo": "bar",
					},
				},
				PodSpec: `{
				"hostname": "hostname"
			}`,
			})
		require.NoError(t, err)

		// Get info about the pipeline pods from k8s & check for resources
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, false)
		require.NoError(t, err)

		var pod v1.Pod
		kubeClient := tu.GetKubeClient(t)
		err = backoff.Retry(func() error {
			podList, err := kubeClient.CoreV1().Pods(ns).List(
				context.Background(),
				metav1.ListOptions{
					LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
						map[string]string{
							"app":             "pipeline",
							"pipelineName":    pipelineInfo.Pipeline.Name,
							"pipelineVersion": fmt.Sprint(pipelineInfo.Version),
							"suite":           "pachyderm"},
					)),
				})
			if err != nil {
				return errors.EnsureStack(err) // retry
			}
			if len(podList.Items) != 1 || len(podList.Items[0].Spec.Containers) == 0 {
				return errors.Errorf("could not find single container for pipeline %s", pipelineInfo.Pipeline.Name)
			}
			pod = podList.Items[0]
			return nil // no more retries
		}, backoff.NewTestingBackOff())
		require.NoError(t, err)
		// Make sure a CPU and Memory request are both set
		require.Equal(t, "bar", pod.Spec.NodeSelector["foo"])
		require.Equal(t, "hostname", pod.Spec.Hostname)
	})
	t.Run("Patch", func(t *testing.T) {
		pipelineName := tu.UniqueString("TestPodSpecOpts")
		_, err := c.PpsAPIClient.CreatePipeline(
			context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
				Transform: &pps.Transform{
					Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
				},
				ParallelismSpec: &pps.ParallelismSpec{
					Constant: 1,
				},
				Input: &pps.Input{
					Pfs: &pps.PFSInput{
						Repo:   dataRepo,
						Branch: "master",
						Glob:   "/*",
					},
				},
				SchedulingSpec: &pps.SchedulingSpec{
					// This NodeSelector will cause the worker pod to fail to
					// schedule, but the test can still pass because we just check
					// for values on the pod, it doesn't need to actually come up.
					NodeSelector: map[string]string{
						"foo": "bar",
					},
				},
				PodPatch: `[
					{ "op": "add", "path": "/hostname", "value": "hostname" }
			]`,
			})
		require.NoError(t, err)

		// Get info about the pipeline pods from k8s & check for resources
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipelineName, false)
		require.NoError(t, err)

		var pod v1.Pod
		kubeClient := tu.GetKubeClient(t)
		err = backoff.Retry(func() error {
			podList, err := kubeClient.CoreV1().Pods(ns).List(
				context.Background(),
				metav1.ListOptions{
					LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
						map[string]string{
							"app":             "pipeline",
							"pipelineName":    pipelineInfo.Pipeline.Name,
							"pipelineVersion": fmt.Sprint(pipelineInfo.Version),
							"suite":           "pachyderm"},
					)),
				})
			if err != nil {
				return errors.EnsureStack(err) // retry
			}
			if len(podList.Items) != 1 || len(podList.Items[0].Spec.Containers) == 0 {
				return errors.Errorf("could not find single container for pipeline %s", pipelineInfo.Pipeline.Name)
			}
			pod = podList.Items[0]
			return nil // no more retries
		}, backoff.NewTestingBackOff())
		require.NoError(t, err)
		// Make sure a CPU and Memory request are both set
		require.Equal(t, "bar", pod.Spec.NodeSelector["foo"])
		require.Equal(t, "hostname", pod.Spec.Hostname)
	})
}

func TestPipelineLargeOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineInputDataModification_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"for i in `seq 1 100`; do touch /pfs/out/$RANDOM; done",
		},
		&pps.ParallelismSpec{
			Constant: 4,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	numFiles := 100
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file-%d", i), strings.NewReader(""), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))
}

func TestJoinInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	var repos []string
	for i := 0; i < 2; i++ {
		repos = append(repos, tu.UniqueString(fmt.Sprintf("TestJoinInput%v", i)))
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repos[i]))
	}

	numFiles := 16
	for r, repo := range repos {
		commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		for i := 0; i < numFiles; i++ {
			require.NoError(t, c.PutFile(commit, fmt.Sprintf("file-%v.%4b", r, i), strings.NewReader(fmt.Sprintf("%d\n", i)), client.WithAppendPutFile()))
		}
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "master", ""))
	}

	pipeline := tu.UniqueString("join-pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("touch /pfs/out/$(echo $(ls -r /pfs/%s/)$(ls -r /pfs/%s/))", repos[0], repos[1]),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewJoinInput(
			client.NewPFSInputOpts("", pfs.DefaultProjectName, repos[0], "", "/file-?.(11*)", "$1", "", false, false, nil),
			client.NewPFSInputOpts("", pfs.DefaultProjectName, repos[1], "", "/file-?.(*0)", "$1", "", false, false, nil),
		),
		"",
		false,
	))

	commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	fileInfos, err := c.ListFileAll(commitInfo.Commit, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
	expectedNames := []string{"/file-0.1100file-1.1100", "/file-0.1110file-1.1110"}
	for i, fi := range fileInfos {
		// 1 byte per repo
		require.Equal(t, expectedNames[i], fi.File.Path)
	}

	dataRepo0 := "TestJoinInputDirectory-0"
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo0))
	masterCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo0, "master", "")
	require.NoError(t, c.PutFile(masterCommit, "/dir-01/foo", strings.NewReader("foo")))
	dataRepo1 := "TestJoinInputDirectory-1"
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo1))
	masterCommit = client.NewCommit(pfs.DefaultProjectName, dataRepo1, "master", "")
	require.NoError(t, c.PutFile(masterCommit, "/dir-10/bar", strings.NewReader("bar")))

	pipeline = tu.UniqueString("join-pipeline-directory")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp -r /pfs/%s/*/. /pfs/out", dataRepo0),
			fmt.Sprintf("cp -r /pfs/%s/*/. /pfs/out", dataRepo1),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewJoinInput(
			client.NewPFSInputOpts("", pfs.DefaultProjectName, dataRepo0, "", "/dir-(?)(?)", "$2", "", false, false, nil),
			client.NewPFSInputOpts("", pfs.DefaultProjectName, dataRepo1, "", "/dir-(?)(?)", "$1", "", false, false, nil),
		),
		"",
		false,
	))

	commitInfo, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	fileInfos, err = c.ListFileAll(commitInfo.Commit, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
	expectedNames = []string{"/bar", "/foo"}
	for i, fi := range fileInfos {
		require.Equal(t, expectedNames[i], fi.File.Path)
	}
}

func TestGroupInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	t.Run("Basic", func(t *testing.T) {
		repo := tu.UniqueString("TestGroupInput")
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
		numFiles := 16
		for i := 0; i < numFiles; i++ {
			require.NoError(t, c.PutFile(commit, fmt.Sprintf("file.%4b", i), strings.NewReader(fmt.Sprintf("%d\n", i)), client.WithAppendPutFile()))
		}

		pipeline := "group-pipeline"
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewGroupInput(
				client.NewPFSInputOpts("", pfs.DefaultProjectName, repo, "", "/file.(?)(?)(?)(?)", "", "$3", false, false, nil),
			),
			"",
			false,
		))

		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)

		jobs, err := c.WaitJobSetAll(commitInfo.Commit.Id, false)
		require.NoError(t, err)
		require.Equal(t, 1, len(jobs))

		// We're grouping by the third digit in the filename
		// for 0 and 1, this is just a space
		// then we should see the 8 files with a one there, and the 6 files with a zero there
		expected := [][]string{
			{"/file.   0",
				"/file.   1"},

			{"/file.  10",
				"/file.  11",
				"/file. 110",
				"/file. 111",
				"/file.1010",
				"/file.1011",
				"/file.1110",
				"/file.1111"},

			{"/file. 100",
				"/file. 101",
				"/file.1000",
				"/file.1001",
				"/file.1100",
				"/file.1101"}}
		actual := make([][]string, 0, 3)
		dis, err := c.ListDatumAll(pfs.DefaultProjectName, jobs[0].Job.Pipeline.Name, jobs[0].Job.Id)
		require.NoError(t, err)
		for _, di := range dis {
			sort.Slice(di.Data, func(i, j int) bool { return di.Data[i].File.Path < di.Data[j].File.Path })
			datumFiles := make([]string, 0)
			for _, fi := range di.Data {
				datumFiles = append(datumFiles, fi.File.Path)
			}
			actual = append(actual, datumFiles)
		}
		sort.Slice(actual, func(i, j int) bool {
			return actual[i][0] < actual[j][0]
		})
		require.Equal(t, expected, actual)
	})

	t.Run("MultiInput", func(t *testing.T) {
		var repos []string
		for i := 0; i < 2; i++ {
			repos = append(repos, tu.UniqueString(fmt.Sprintf("TestGroupInput%v", i)))
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repos[i]))
		}

		numFiles := 16
		for r, repo := range repos {
			for i := 0; i < numFiles; i++ {
				require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), fmt.Sprintf("file-%v.%4b", r, i), strings.NewReader(fmt.Sprintf("%d\n", i)), client.WithAppendPutFile()))
			}
		}

		pipeline := "group-pipeline-multi-input"
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewGroupInput(
				client.NewPFSInputOpts("", pfs.DefaultProjectName, repos[0], "", "/file-?.(?)(?)(?)(?)", "", "$3", false, false, nil),
				client.NewPFSInputOpts("", pfs.DefaultProjectName, repos[1], "", "/file-?.(?)(?)(?)(?)", "", "$2", false, false, nil),
			),
			"",
			false,
		))

		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)

		jobs, err := c.WaitJobSetAll(commitInfo.Commit.Id, false)
		require.NoError(t, err)
		require.Equal(t, 1, len(jobs))

		// this time, we are grouping by the third digit in the 0 repo, and the second digit in the 1 repo
		// so the first group should have all the things from the 0 repo with a space in the third digit
		// and all the things from the 1 repo with a space in the second digit
		//
		// similarly for the second and third groups
		expected := [][]string{
			{"/file-0.   0",
				"/file-0.   1",
				"/file-1.   0",
				"/file-1.   1",
				"/file-1.  10",
				"/file-1.  11"},

			{"/file-0.  10",
				"/file-0.  11",
				"/file-0. 110",
				"/file-0. 111",
				"/file-0.1010",
				"/file-0.1011",
				"/file-0.1110",
				"/file-0.1111",
				"/file-1. 100",
				"/file-1. 101",
				"/file-1. 110",
				"/file-1. 111",
				"/file-1.1100",
				"/file-1.1101",
				"/file-1.1110",
				"/file-1.1111"},

			{"/file-0. 100",
				"/file-0. 101",
				"/file-0.1000",
				"/file-0.1001",
				"/file-0.1100",
				"/file-0.1101",
				"/file-1.1000",
				"/file-1.1001",
				"/file-1.1010",
				"/file-1.1011"}}
		actual := make([][]string, 0, 3)
		dis, err := c.ListDatumAll(pfs.DefaultProjectName, jobs[0].Job.Pipeline.Name, jobs[0].Job.Id)
		require.NoError(t, err)
		for _, di := range dis {
			sort.Slice(di.Data, func(i, j int) bool { return di.Data[i].File.Path < di.Data[j].File.Path })
			datumFiles := make([]string, 0)
			for _, fi := range di.Data {
				datumFiles = append(datumFiles, fi.File.Path)
			}
			actual = append(actual, datumFiles)
		}
		sort.Slice(actual, func(i, j int) bool {
			return actual[i][0] < actual[j][0]
		})
		require.Equal(t, expected, actual)
	})

	t.Run("GroupJoinCombo", func(t *testing.T) {
		var repos []string
		for i := 0; i < 2; i++ {
			repos = append(repos, tu.UniqueString(fmt.Sprintf("TestGroupInput%v", i)))
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repos[i]))
		}

		numFiles := 16
		for r, repo := range repos {
			for i := 0; i < numFiles; i++ {
				require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), fmt.Sprintf("file-%v.%4b", r, i), strings.NewReader(fmt.Sprintf("%d\n", i)), client.WithAppendPutFile()))
			}
		}

		pipeline := "group-join-pipeline"
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewGroupInput(
				client.NewJoinInput(
					client.NewPFSInputOpts("", pfs.DefaultProjectName, repos[0], "", "/file-?.(?)(?)(?)(?)", "$1$2$3$4", "$3", false, false, nil),
					client.NewPFSInputOpts("", pfs.DefaultProjectName, repos[1], "", "/file-?.(?)(?)(?)(?)", "$4$3$2$1", "$2", false, false, nil),
				),
			),
			"",
			false,
		))

		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)

		jobs, err := c.WaitJobSetAll(commitInfo.Commit.Id, false)
		require.NoError(t, err)
		require.Equal(t, 1, len(jobs))

		// here, we're first doing a join to get pairs of files (one from each repo) that have the reverse numbers
		// we should see four pairs
		// then, we're grouping the files in these pairs by the third digit/second digit as before
		// this should regroup things into two groups of four
		expected := [][]string{
			{"/file-0.1001",
				"/file-0.1101",
				"/file-1.1001",
				"/file-1.1011"},
			{"/file-0.1011",
				"/file-0.1111",
				"/file-1.1101",
				"/file-1.1111"}}
		actual := make([][]string, 0, 2)
		dis, err := c.ListDatumAll(pfs.DefaultProjectName, jobs[0].Job.Pipeline.Name, jobs[0].Job.Id)
		require.NoError(t, err)
		for _, di := range dis {
			sort.Slice(di.Data, func(i, j int) bool { return di.Data[i].File.Path < di.Data[j].File.Path })
			datumFiles := make([]string, 0)
			for _, fi := range di.Data {
				datumFiles = append(datumFiles, fi.File.Path)
			}
			actual = append(actual, datumFiles)
		}
		sort.Slice(actual, func(i, j int) bool {
			return actual[i][0] < actual[j][0]
		})
		require.Equal(t, expected, actual)
	})
	t.Run("Symlink", func(t *testing.T) {
		// Fix for the bug exhibited here: https://github.com/pachyderm/pachyderm/v2/tree/example-groupby/examples/group
		repo := tu.UniqueString("TestGroupInputSymlink")
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))

		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		require.NoError(t, c.PutFile(commit, "/T1606331395-LIPID-PATID2-CLIA24D9871327.txt", strings.NewReader(""), client.WithAppendPutFile()))
		require.NoError(t, c.PutFile(commit, "/T1606707579-LIPID-PATID3-CLIA24D9871327.txt", strings.NewReader(""), client.WithAppendPutFile()))
		require.NoError(t, c.PutFile(commit, "/T1606707597-LIPID-PATID4-CLIA24D9871327.txt", strings.NewReader(""), client.WithAppendPutFile()))
		require.NoError(t, c.PutFile(commit, "/T1606707557-LIPID-PATID1-CLIA24D9871327.txt", strings.NewReader(""), client.WithAppendPutFile()))
		require.NoError(t, c.PutFile(commit, "/T1606707613-LIPID-PATID1-CLIA24D9871328.txt", strings.NewReader(""), client.WithAppendPutFile()))
		require.NoError(t, c.PutFile(commit, "/T1606707635-LIPID-PATID3-CLIA24D9871328.txt", strings.NewReader(""), client.WithAppendPutFile()))

		pipeline := "group-pipeline-symlink"
		_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"bash"},
					Stdin: []string{"PATTERN=.*-PATID\\(.*\\)-.*.txt",
						fmt.Sprintf("FILES=/pfs/%v/*", repo),
						"for f in $FILES",
						"do",
						"[[ $(basename $f) =~ $PATTERN ]]",
						"mkdir -p /pfs/out/${BASH_REMATCH[1]}/",
						"cp $f /pfs/out/${BASH_REMATCH[1]}/",
						"done"},
				},
				Input: client.NewGroupInput(
					client.NewPFSInputOpts("", pfs.DefaultProjectName, repo, "master", "/*-PATID(*)-*.txt", "", "$1", false, false, nil),
				),
				ParallelismSpec: &pps.ParallelismSpec{
					Constant: 1,
				},
			})
		require.NoError(t, err)

		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)

		jobs, err := c.WaitJobSetAll(commitInfo.Commit.Id, false)
		require.NoError(t, err)
		require.Equal(t, 1, len(jobs))

		require.Equal(t, "JOB_SUCCESS", jobs[0].State.String())

		expected := [][]string{
			{"/T1606331395-LIPID-PATID2-CLIA24D9871327.txt"},
			{"/T1606707557-LIPID-PATID1-CLIA24D9871327.txt", "/T1606707613-LIPID-PATID1-CLIA24D9871328.txt"},
			{"/T1606707579-LIPID-PATID3-CLIA24D9871327.txt", "/T1606707635-LIPID-PATID3-CLIA24D9871328.txt"},
			{"/T1606707597-LIPID-PATID4-CLIA24D9871327.txt"},
		}
		actual := make([][]string, 0, 3)
		dis, err := c.ListDatumAll(pfs.DefaultProjectName, jobs[0].Job.Pipeline.Name, jobs[0].Job.Id)
		require.NoError(t, err)
		// these don't come in a consistent order because group inputs use maps
		for _, di := range dis {
			sort.Slice(di.Data, func(i, j int) bool { return di.Data[i].File.Path < di.Data[j].File.Path })
			datumFiles := make([]string, 0)
			for _, fi := range di.Data {
				datumFiles = append(datumFiles, fi.File.Path)
			}
			actual = append(actual, datumFiles)
		}
		sort.Slice(actual, func(i, j int) bool {
			return actual[i][0] < actual[j][0]
		})
		require.Equal(t, expected, actual)
	})
}

func TestUnionInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	var repos []string
	for i := 0; i < 4; i++ {
		repos = append(repos, tu.UniqueString("TestUnionInput"))
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repos[i]))
	}

	numFiles := 2
	for _, repo := range repos {
		commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		for i := 0; i < numFiles; i++ {
			require.NoError(t, c.PutFile(commit, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)), client.WithAppendPutFile()))
		}
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "master", ""))
	}

	t.Run("union all", func(t *testing.T) {
		pipeline := tu.UniqueString("pipeline")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp /pfs/*/* /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewUnionInput(
				client.NewPFSInput(pfs.DefaultProjectName, repos[0], "/*"),
				client.NewPFSInput(pfs.DefaultProjectName, repos[1], "/*"),
				client.NewPFSInput(pfs.DefaultProjectName, repos[2], "/*"),
				client.NewPFSInput(pfs.DefaultProjectName, repos[3], "/*"),
			),
			"",
			false,
		))

		commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)
		fileInfos, err := c.ListFileAll(commitInfo.Commit, "")
		require.NoError(t, err)
		require.Equal(t, 8, len(fileInfos))
	})

	t.Run("union crosses", func(t *testing.T) {
		pipeline := tu.UniqueString("pipeline")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp -r -L /pfs/TestUnionInput* /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewUnionInput(
				client.NewCrossInput(
					client.NewPFSInput(pfs.DefaultProjectName, repos[0], "/*"),
					client.NewPFSInput(pfs.DefaultProjectName, repos[1], "/*"),
				),
				client.NewCrossInput(
					client.NewPFSInput(pfs.DefaultProjectName, repos[2], "/*"),
					client.NewPFSInput(pfs.DefaultProjectName, repos[3], "/*"),
				),
			),
			"",
			false,
		))

		commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)
		for _, repo := range repos {
			fileInfos, err := c.ListFileAll(commitInfo.Commit, repo)
			require.NoError(t, err)
			require.Equal(t, 4, len(fileInfos))
		}
	})

	t.Run("cross unions", func(t *testing.T) {
		pipeline := tu.UniqueString("pipeline")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp -r -L /pfs/TestUnionInput* /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewCrossInput(
				client.NewUnionInput(
					client.NewPFSInput(pfs.DefaultProjectName, repos[0], "/*"),
					client.NewPFSInput(pfs.DefaultProjectName, repos[1], "/*"),
				),
				client.NewUnionInput(
					client.NewPFSInput(pfs.DefaultProjectName, repos[2], "/*"),
					client.NewPFSInput(pfs.DefaultProjectName, repos[3], "/*"),
				),
			),
			"",
			false,
		))

		commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)
		for _, repo := range repos {
			fileInfos, err := c.ListFileAll(commitInfo.Commit, repo)
			require.NoError(t, err)
			require.Equal(t, 8, len(fileInfos))
		}
	})

	t.Run("union alias", func(t *testing.T) {
		pipeline := tu.UniqueString("pipeline")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp /pfs/in/* '/pfs/out/$RANDOM'",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewUnionInput(
				client.NewPFSInputOpts("in", pfs.DefaultProjectName, repos[0], "", "/*", "", "", false, false, nil),
				client.NewPFSInputOpts("in", pfs.DefaultProjectName, repos[1], "", "/*", "", "", false, false, nil),
				client.NewPFSInputOpts("in", pfs.DefaultProjectName, repos[2], "", "/*", "", "", false, false, nil),
				client.NewPFSInputOpts("in", pfs.DefaultProjectName, repos[3], "", "/*", "", "", false, false, nil),
			),
			"",
			false,
		))

		commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)
		fileInfos, err := c.ListFileAll(commitInfo.Commit, "")
		require.NoError(t, err)
		require.Equal(t, 8, len(fileInfos))
		dis, err := c.ListDatumAll(pfs.DefaultProjectName, pipeline, commitInfo.Commit.Id)
		require.NoError(t, err)
		require.Equal(t, 8, len(dis))
	})
}

func TestPipelineWithStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineWithStats_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	numFiles := 10
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file-%d", i), strings.NewReader(strings.Repeat("foo\n", 100))))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", commit1.Id))

	pipeline := tu.UniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 4,
			},
		})
	require.NoError(t, err)

	outputCommit, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)

	id := outputCommit.Commit.Id

	commitInfos, err := c.WaitCommitSetAll(id)
	require.NoError(t, err)
	// input (alias), output, spec, meta
	require.Equal(t, 4, len(commitInfos))

	jobs, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	resp, err := c.ListDatumAll(pfs.DefaultProjectName, pipeline, id)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(resp))
	require.Equal(t, 1, len(resp[0].Data))

	for _, datum := range resp {
		require.NoError(t, err)
		require.Equal(t, pps.DatumState_SUCCESS, datum.State)
	}

	// Make sure 'inspect datum' works
	datum, err := c.InspectDatum(pfs.DefaultProjectName, pipeline, id, resp[0].Datum.Id)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_SUCCESS, datum.State)
}

func TestPipelineWithStatsFailedDatums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineWithStatsFailedDatums_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	numFiles := 10
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file-%d", i), strings.NewReader(strings.Repeat("foo\n", 100))))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", commit1.Id))

	pipeline := tu.UniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("if [ -f /pfs/%s/file-5 ]; then exit 1; fi", dataRepo),
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 4,
			},
		})
	require.NoError(t, err)

	outputCommit, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	id := outputCommit.Commit.Id

	commitInfos, err := c.WaitCommitSetAll(id)
	require.NoError(t, err)
	// input (alias), output, spec, meta
	require.Equal(t, 4, len(commitInfos))

	jobs, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	resp, err := c.ListDatumAll(pfs.DefaultProjectName, pipeline, id)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(resp))

	var failedID string
	for _, d := range resp {
		if d.State == pps.DatumState_FAILED {
			// one entry should be failed (list datum isn't sorted)
			require.Equal(t, "", failedID)
			failedID = d.Datum.Id
		} else {
			// other entries should be success
			require.Equal(t, pps.DatumState_SUCCESS, d.State)
		}
	}

	// Make sure 'inspect datum' works for failed state
	datum, err := c.InspectDatum(pfs.DefaultProjectName, pipeline, id, failedID)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_FAILED, datum.State)
}

func TestPipelineWithStatsAcrossJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	project := tu.UniqueString("prj-")
	require.NoError(t, c.CreateProject(project))

	dataRepo := tu.UniqueString("TestPipelineWithStatsAcrossJobs_data")
	require.NoError(t, c.CreateRepo(project, dataRepo))

	numFiles := 10
	commit1, err := c.StartCommit(project, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("foo-%d", i), strings.NewReader(strings.Repeat("foo\n", 100)), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(project, dataRepo, "master", commit1.Id))

	pipeline := tu.UniqueString("StatsAcrossJobs")
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(project, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input: client.NewPFSInput(project, dataRepo, "/*"),
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
		})
	require.NoError(t, err)

	outputCommit, err := c.InspectCommit(project, pipeline, "master", "")
	require.NoError(t, err)
	id := outputCommit.Commit.Id

	commitInfos, err := c.WaitCommitSetAll(id)
	require.NoError(t, err)
	// input (alias), output, spec, meta
	require.Equal(t, 4, len(commitInfos))

	jobs, err := c.ListJob(project, pipeline, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	resp, err := c.ListDatumAll(project, pipeline, id)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(resp))

	datum, err := c.InspectDatum(project, pipeline, id, resp[0].Datum.Id)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_SUCCESS, datum.State)

	commit2, err := c.StartCommit(project, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit2, fmt.Sprintf("bar-%d", i), strings.NewReader(strings.Repeat("bar\n", 100)), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(project, dataRepo, "master", commit2.Id))

	outputCommit, err = c.InspectCommit(project, pipeline, "master", "")
	require.NoError(t, err)
	id = outputCommit.Commit.Id

	commitInfos, err = c.WaitCommitSetAll(id)
	require.NoError(t, err)
	// input (alias), output, spec, meta
	require.Equal(t, 4, len(commitInfos))

	jobs, err = c.ListJob(project, pipeline, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))

	_, err = c.ListDatumAll(pfs.DefaultProjectName, pipeline, id)
	require.YesError(t, err)
	resp, err = c.ListDatumAll(project, pipeline, id)
	require.NoError(t, err)
	// we should see all the datums from the first job (which should be skipped)
	// in addition to all the new datums processed in this job
	require.Equal(t, numFiles*2, len(resp))

	// test inspect datum
	var skippedCount int
	for _, info := range resp {
		if info.State == pps.DatumState_SKIPPED {
			skippedCount++
		}
		_, err := c.InspectDatum(pfs.DefaultProjectName, pipeline, id, info.Datum.Id)
		require.YesError(t, err)
		inspectedInfo, err := c.InspectDatum(project, pipeline, id, info.Datum.Id)
		require.NoError(t, err)
		require.Equal(t, info.State, inspectedInfo.State)
	}
	require.Equal(t, numFiles, skippedCount)
}

func TestPipelineOnStatsBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineOnStatsBranch_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	pipeline1, pipeline2 := tu.UniqueString("TestPipelineOnStatsBranch1"), tu.UniqueString("TestPipelineOnStatsBranch2")
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline1),
			Transform: &pps.Transform{
				Cmd: []string{"bash", "-c", "cp -r $(ls -d /pfs/*|grep -v /pfs/out) /pfs/out"},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline2),
			Transform: &pps.Transform{
				Cmd: []string{"bash", "-c", "cp -r $(ls -d /pfs/*|grep -v /pfs/out) /pfs/out"},
			},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:     pipeline1,
					RepoType: pfs.MetaRepoType,
					Branch:   "master",
					Glob:     "/*",
				},
			},
		})
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline2, "master", "")
	require.NoError(t, err)

	jobInfos, err := c.WaitJobSetAll(commitInfo.Commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	for _, ji := range jobInfos {
		require.Equal(t, ji.State.String(), pps.JobState_JOB_SUCCESS.String())
	}
}

func TestSkippedDatums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// create pipeline
	pipelineName := tu.UniqueString("pipeline")
	//	require.NoError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)

	// Do first commit to repo
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))
	jis, err := c.WaitJobSetAll(commit1.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	ji := jis[0]
	require.Equal(t, ji.State, pps.JobState_JOB_SUCCESS)
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(ji.OutputCommit, "file", &buffer))
	require.Equal(t, "foo\n", buffer.String())

	// Do second commit to repo
	commit2, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit2, "file2", strings.NewReader("bar\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit2.Id))
	jis, err = c.WaitJobSetAll(commit2.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	ji = jis[0]
	require.Equal(t, ji.State, pps.JobState_JOB_SUCCESS)

	jobs, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 3, len(jobs))

	job1 := jobs[1]
	datums, err := c.ListDatumAll(pfs.DefaultProjectName, pipelineName, job1.Job.Id)
	require.NoError(t, err)
	require.Equal(t, 1, len(datums))
	datum, err := c.InspectDatum(pfs.DefaultProjectName, pipelineName, job1.Job.Id, datums[0].Datum.Id)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_SUCCESS, datum.State)

	job2 := jobs[0]
	// check the successful datum from job1 is now skipped
	datum, err = c.InspectDatum(pfs.DefaultProjectName, pipelineName, job2.Job.Id, datums[0].Datum.Id)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_SKIPPED, datum.State)
	// load datums for job2
	datums, err = c.ListDatumAll(pfs.DefaultProjectName, pipelineName, job2.Job.Id)
	require.NoError(t, err)
	require.Equal(t, 2, len(datums))
	require.NoError(t, err)
	require.Equal(t, int64(1), job2.DataSkipped)
	require.Equal(t, pps.JobState_JOB_SUCCESS, job2.State)
}

func TestMetaRepoContents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	pipelineName := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)
	assertMetaContents := func(commitID string, inputFile string) {
		var datumID string
		require.NoError(t, c.ListDatum(pfs.DefaultProjectName, pipelineName, commitID, func(di *pps.DatumInfo) error {
			datumID = di.Datum.Id
			return nil
		}))
		expectedFiles := map[string]struct{}{
			fmt.Sprintf("/meta/%v/meta", datumID):                      {},
			fmt.Sprintf("/pfs/%v/.env", datumID):                       {},
			fmt.Sprintf("/pfs/%v/%v/%v", datumID, dataRepo, inputFile): {},
			fmt.Sprintf("/pfs/%v/out/%v", datumID, inputFile):          {},
		}
		metaCommit := &pfs.Commit{
			Branch: client.NewSystemRepo(pfs.DefaultProjectName, pipelineName, pfs.MetaRepoType).NewBranch("master"),
			Id:     commitID,
		}
		var files []string
		require.NoError(t, c.WalkFile(metaCommit, "", func(fi *pfs.FileInfo) error {
			if fi.FileType == pfs.FileType_FILE {
				files = append(files, fi.File.Path)
			}
			return nil
		}))
		require.Equal(t, len(expectedFiles), len(files))
		for _, f := range files {
			delete(expectedFiles, f)
		}
		require.Equal(t, 0, len(expectedFiles))
	}
	// Do first commit to repo
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "foo", strings.NewReader("foo\n")))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))
	jis, err := c.WaitJobSetAll(commit1.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	ji := jis[0]
	require.Equal(t, ji.State, pps.JobState_JOB_SUCCESS)
	assertMetaContents(commit1.Id, "foo")
	// Replace file in input repo
	commit2, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit2, "bar", strings.NewReader("bar\n")))
	require.NoError(t, c.DeleteFile(commit2, "/foo"))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit2.Id))
	_, err = c.WaitJobSetAll(commit2.Id, false)
	require.NoError(t, err)
	assertMetaContents(commit2.Id, "bar")
	fileCount := 0
	var fileName string
	require.NoError(t, c.ListFile(client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", commit2.Id), "", func(fi *pfs.FileInfo) error {
		fileCount++
		fileName = fi.File.Path
		return nil
	}))
	require.Equal(t, 1, fileCount)
	require.Equal(t, "/bar", fileName)
}

func TestCronPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	t.Run("SimpleCron", func(t *testing.T) {
		defer func() {
			require.NoError(t, c.DeleteAll(c.Ctx()))
		}()
		pipeline1 := tu.UniqueString("cron1-")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline1,
			"",
			[]string{"/bin/bash"},
			[]string{"cp /pfs/time/* /pfs/out/"},
			nil,
			client.NewCronInput("time", "@every 10s"),
			"",
			false,
		))
		pipeline2 := tu.UniqueString("cron2-")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline2,
			"",
			[]string{"/bin/bash"},
			[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline1) + " /pfs/out/"},
			nil,
			client.NewPFSInput(pfs.DefaultProjectName, pipeline1, "/*"),
			"",
			false,
		))

		// subscribe to the pipeline2 cron repo and wait for inputs
		repo := client.NewRepo(pfs.DefaultProjectName, pipeline2)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
		defer cancel() //cleanup resources
		// We'll look at four commits - with one initial head and one created in each tick
		// We expect the first commit to have 0 files, the second to have 1 file, etc...
		countBreakFunc := newCountBreakFunc(4)
		count := 0
		require.NoError(t, c.WithCtx(ctx).SubscribeCommit(repo, "master", "", pfs.CommitState_STARTED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				_, err := c.WaitCommitSetAll(ci.Commit.Id)
				require.NoError(t, err)

				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, count, len(files))
				count++

				return nil
			})
		}))
	})

	// Test a CronInput with the overwrite flag set to true
	t.Run("CronOverwrite", func(t *testing.T) {
		defer func() {
			require.NoError(t, c.DeleteAll(c.Ctx()))
		}()
		pipeline3 := tu.UniqueString("cron3-")
		overwriteInput := client.NewCronInput("time", "@every 10s")
		overwriteInput.Cron.Overwrite = true
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline3,
			"",
			[]string{"/bin/bash"},
			[]string{"cp /pfs/time/* /pfs/out/"},
			nil,
			overwriteInput,
			"",
			false,
		))
		repo := client.NewRepo(pfs.DefaultProjectName, fmt.Sprintf("%s_time", pipeline3))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
		defer cancel() //cleanup resources
		// We'll look at four commits - with one created in each tick
		// We expect each of the commits to have just a single file in this case
		// except the first, which is the default branch head.
		countBreakFunc := newCountBreakFunc(3)
		count := 0
		require.NoError(t, c.WithCtx(ctx).SubscribeCommit(repo, "master", "", pfs.CommitState_STARTED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				commitInfos, err := c.WaitCommitSetAll(ci.Commit.Id)
				require.NoError(t, err)
				require.Equal(t, 4, len(commitInfos))

				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, count, len(files))
				count = 1

				return nil
			})
		}))
	})

	// Create a non-cron input repo, and test a pipeline with a cross of cron and
	// non-cron inputs
	t.Run("CronPFSCross", func(t *testing.T) {
		defer func() {
			require.NoError(t, c.DeleteAll(c.Ctx()))
		}()
		dataRepo := tu.UniqueString("TestCronPipeline_data")
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
		pipeline4 := tu.UniqueString("cron4-")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline4,
			"",
			[]string{"bash"},
			[]string{
				"cp /pfs/time/time /pfs/out/time",
				fmt.Sprintf("cp /pfs/%s/file /pfs/out/file", dataRepo),
			},
			nil,
			client.NewCrossInput(
				client.NewCronInput("time", "@every 20s"),
				client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
			),
			"",
			false,
		))
		_, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "file", strings.NewReader("file"), client.WithAppendPutFile()))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))

		repo := client.NewRepo(pfs.DefaultProjectName, fmt.Sprintf("%s_time", pipeline4))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel() //cleanup resources
		countBreakFunc := newCountBreakFunc(2)
		require.NoError(t, c.WithCtx(ctx).SubscribeCommit(repo, "master", "", pfs.CommitState_STARTED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				commitInfos, err := c.WaitCommitSetAll(ci.Commit.Id)
				require.NoError(t, err)
				require.Equal(t, 5, len(commitInfos))

				return nil
			})
		}))
	})
	t.Run("RunCron", func(t *testing.T) {
		defer func() {
			require.NoError(t, c.DeleteAll(c.Ctx()))
		}()
		pipeline5 := tu.UniqueString("cron5-")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline5,
			"",
			[]string{"/bin/bash"},
			[]string{"cp /pfs/time/* /pfs/out/"},
			nil,
			client.NewCronInput("time", "@every 1h"),
			"",
			false,
		))
		pipeline6 := tu.UniqueString("cron6-")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline6,
			"",
			[]string{"/bin/bash"},
			[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline5) + " /pfs/out/"},
			nil,
			client.NewPFSInput(pfs.DefaultProjectName, pipeline5, "/*"),
			"",
			false,
		))

		_, err := c.PpsAPIClient.RunCron(context.Background(), &pps.RunCronRequest{Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline5)})
		require.NoError(t, err)
		_, err = c.PpsAPIClient.RunCron(context.Background(), &pps.RunCronRequest{Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline5)})
		require.NoError(t, err)
		_, err = c.PpsAPIClient.RunCron(context.Background(), &pps.RunCronRequest{Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline5)})
		require.NoError(t, err)

		// subscribe to the pipeline6 cron repo and wait for inputs
		repo := client.NewRepo(pfs.DefaultProjectName, pipeline6)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
		defer cancel() //cleanup resources
		countBreakFunc := newCountBreakFunc(4)
		require.NoError(t, c.WithCtx(ctx).SubscribeCommit(repo, "master", "", pfs.CommitState_STARTED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				_, err := c.WaitCommit(pfs.DefaultProjectName, pipeline6, "master", ci.Commit.Id)
				require.NoError(t, err)

				return nil
			})
		}))
	})
	t.Run("RunCronOverwrite", func(t *testing.T) {
		defer func() {
			require.NoError(t, c.DeleteAll(c.Ctx()))
		}()
		pipeline7 := tu.UniqueString("cron7-")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline7,
			"",
			[]string{"/bin/bash"},
			[]string{"cp /pfs/time/* /pfs/out/"},
			nil,
			client.NewCronInputOpts("time", "", "*/1 * * * *", true, nil), // every minute
			"",
			false,
		))
		pipeline8 := tu.UniqueString("cron8-")
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline8,
			"",
			[]string{"/bin/bash"},
			[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline7) + " /pfs/out/"},
			nil,
			client.NewPFSInput(pfs.DefaultProjectName, pipeline7, "/*"),
			"",
			false,
		))

		// subscribe to the pipeline1 cron repo and wait for inputs
		repo := fmt.Sprintf("%s_%s", pipeline7, "time")
		// if the runcron is run too soon, it will have the same timestamp and we won't hit the weird bug
		time.Sleep(2 * time.Second)

		nCronticks := 3
		for i := 0; i < nCronticks; i++ {
			_, err := c.PpsAPIClient.RunCron(context.Background(), &pps.RunCronRequest{Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline7)})
			require.NoError(t, err)
			_, err = c.WaitCommit(pfs.DefaultProjectName, repo, "master", "")
			require.NoError(t, err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
		defer cancel() //cleanup resources
		countBreakFunc := newCountBreakFunc(1)
		require.NoError(t, c.WithCtx(ctx).SubscribeCommit(client.NewRepo(pfs.DefaultProjectName, repo), "master", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				ctx, cancel = context.WithTimeout(context.Background(), time.Second*120)
				defer cancel() //cleanup resources
				// We expect to see four commits, despite the schedule being every minute, and the timeout 120 seconds
				// We expect each of the commits to have just a single file in this case
				// We check four so that we can make sure the scheduled cron is not messed up by the run crons
				countBreakFunc := newCountBreakFunc(nCronticks + 1)
				require.NoError(t, c.WithCtx(ctx).SubscribeCommit(client.NewRepo(pfs.DefaultProjectName, repo), "master", ci.Commit.Id, pfs.CommitState_STARTED, func(ci *pfs.CommitInfo) error {
					return countBreakFunc(func() error {
						_, err := c.WaitCommit(pfs.DefaultProjectName, repo, "", ci.Commit.Id)
						require.NoError(t, err)
						files, err := c.ListFileAll(ci.Commit, "/")
						require.NoError(t, err)
						require.Equal(t, 1, len(files))
						return nil
					})
				}))
				return nil
			})
		}))
	})
	t.Run("RunCronCross", func(t *testing.T) {
		defer func() {
			require.NoError(t, c.DeleteAll(c.Ctx()))
		}()
		pipeline9 := tu.UniqueString("cron9-")
		// schedule cron ticks so they definitely won't occur
		futureMonth := int(time.Now().AddDate(0, 2, 0).Month())
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline9,
			"",
			[]string{"/bin/bash"},
			[]string{"echo 'tick'"},
			nil,
			client.NewCrossInput(
				client.NewCronInput("time1", fmt.Sprintf("0 0 1 %d *", futureMonth)),
				client.NewCronInput("time2", fmt.Sprintf("0 0 15 %d *", futureMonth)),
			),
			"",
			false,
		))

		_, err := c.PpsAPIClient.RunCron(context.Background(), &pps.RunCronRequest{Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline9)})
		require.NoError(t, err)
		_, err = c.PpsAPIClient.RunCron(context.Background(), &pps.RunCronRequest{Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline9)})
		require.NoError(t, err)
		_, err = c.PpsAPIClient.RunCron(context.Background(), &pps.RunCronRequest{Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline9)})
		require.NoError(t, err)

		// We should see an initial empty commit, exactly six from our RunCron calls (two each), and nothing else
		commits, err := c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline9), nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 7, len(commits))
	})

	// Cron input with overwrite, and a start time in the past, so that the cron tick has to catch up.
	// Ensure that the deletion of the previous cron tick file completes before the new cron tick file is created.
	t.Run("CronOverwriteStartCatchup", func(t *testing.T) {
		defer func() {
			require.NoError(t, c.DeleteAll(c.Ctx()))
		}()
		pipeline := tu.UniqueString("testCron-")
		start := timestamppb.New(time.Now().Add(-10 * time.Hour))
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"/bin/bash"},
			[]string{"cp /pfs/time/* /pfs/out/"},
			nil,
			client.NewCronInputOpts("in", "", "@every 1h", true, start),
			"",
			false,
		))

		repo := client.NewRepo(pfs.DefaultProjectName, fmt.Sprintf("%s_in", pipeline))
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		count := 0
		countBreakFunc := newCountBreakFunc(11)
		require.NoError(t,
			c.WithCtx(ctx).SubscribeCommit(repo, "master", "", pfs.CommitState_STARTED, func(ci *pfs.CommitInfo) error {
				return countBreakFunc(func() error {
					commitInfos, err := c.WaitCommitSetAll(ci.Commit.Id)
					require.NoError(t, err)
					require.Equal(t, 4, len(commitInfos))

					files, err := c.ListFileAll(ci.Commit, "")
					require.NoError(t, err)
					require.Equal(t, count, len(files))
					count = 1
					return nil
				})
			}))
	})

	// test updating start time works for cron pipeline
	// create a cron pipeline without a start time so it starts immediately
	// let it run for couple ticks
	// update pipeline with a start time some ticks in future
	// let it run couple ticks past new start time
	// confirm pipeline did not run till the start time after pipeline was updated
	t.Run("CronUpdateStart", func(t *testing.T) {
		defer func() {
			require.NoError(t, c.DeleteAll(c.Ctx()))
		}()
		pipeline := tu.UniqueString("cron-upd1-")

		tick := time.Duration(10 * time.Second)
		cronExpr := fmt.Sprintf("@every %s", tick)

		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"/bin/bash"},
			[]string{"cp -v /pfs/sched/* /pfs/out/"},
			nil,
			client.NewCronInput("sched", cronExpr),
			"",
			false,
		))

		outputRepo := client.NewRepo(pfs.DefaultProjectName, pipeline)

		// sleep while pipeline runs couple ticks
		time.Sleep(2 * tick)

		numTicksToSkip := 3

		startTime := time.Now().Add(time.Duration(numTicksToSkip) * tick)
		startTimeStr := timestamppb.New(startTime)

		// update pipeline with start time few ticks into future
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"/bin/bash"},
			[]string{"cp -v /pfs/sched/* /pfs/out/"},
			nil,
			client.NewCronInputOpts("sched", "", cronExpr, false, startTimeStr),
			"",
			true,
		))

		// sleep till couple ticks past new start time
		time.Sleep(time.Duration(numTicksToSkip+2) * tick)

		// get most recent commit
		commits, err := c.ListCommit(outputRepo, nil, nil, 1)
		require.NoError(t, err)
		lastCommit := commits[0].Commit

		files, err := c.ListFileAll(lastCommit, "")
		require.NoError(t, err)
		// check intervals between cron timestamp files
		// all timestamps should be one tick apart except one that should be at least numTicksToSkip apart when the start time was updated
		prevFileTime, err := time.Parse(time.RFC3339, path.Base(files[0].File.Path))
		require.NoError(t, err)

		numSkips := 0
		numTicks := 0
		var timeAfterSkip time.Time
		for i, file := range files {
			if i == 0 {
				continue
			}
			fileTime, err := time.Parse(time.RFC3339, path.Base(file.File.Path))
			require.NoError(t, err)
			gap := fileTime.Sub(prevFileTime)
			prevFileTime = fileTime
			if gap >= time.Duration(numTicksToSkip)*tick {
				numSkips++
				timeAfterSkip = fileTime
				continue

			}
			if gap == tick {
				numTicks++
			}
		}
		require.Equal(t, numSkips, 1)
		require.False(t, timeAfterSkip.Before(startTime))
		require.Equal(t, numTicks, len(files)-2)
	})

}

func TestSelfReferentialPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	pipeline := tu.UniqueString("pipeline")
	require.YesError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"true"},
		nil,
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, pipeline, "/"),
		"",
		false,
	))
}

func TestPipelineBadImage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	pipeline1 := tu.UniqueString("bad_pipeline_1_")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline1,
		"BadImage",
		[]string{"true"},
		nil,
		nil,
		client.NewCronInput("time", "@every 20s"),
		"",
		false,
	))
	pipeline2 := tu.UniqueString("bad_pipeline_2_")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline2,
		"bs/badimage:vcrap",
		[]string{"true"},
		nil,
		nil,
		client.NewCronInput("time", "@every 20s"),
		"",
		false,
	))
	require.NoError(t, backoff.Retry(func() error {
		for _, pipeline := range []string{pipeline1, pipeline2} {
			pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
			if err != nil {
				return err
			}
			if pipelineInfo.State != pps.PipelineState_PIPELINE_CRASHING {
				return errors.Errorf("pipeline %s should be in crashing", pipeline)
			}
			require.True(t, pipelineInfo.Reason != "")
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestFixPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestFixPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("1"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))
	pipelineName := tu.UniqueString("TestFixPipeline_pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"exit 1"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return errors.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		jobInfo, err := c.WaitJob(pfs.DefaultProjectName, jobInfos[0].Job.Pipeline.Name, jobInfos[0].Job.Id, false)
		require.NoError(t, err)
		require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)
		return nil
	}, backoff.NewTestingBackOff()))

	// Update the pipeline, this will not create a new pipeline as reprocess
	// isn't set to true.
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"echo bar >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		true,
	))

	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
		require.NoError(t, err)
		if len(jobInfos) != 2 {
			return errors.Errorf("expected 2 jobs, got %d", len(jobInfos))
		}
		jobInfo, err := c.WaitJob(pfs.DefaultProjectName, jobInfos[0].Job.Pipeline.Name, jobInfos[0].Job.Id, false)
		require.NoError(t, err)
		require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestListJobTruncated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestListJobTruncated_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	liteJobInfos, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, 0, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(liteJobInfos))
	require.Equal(t, commit1.Id, liteJobInfos[0].Job.Id)
	fullJobInfos, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, commit1.Id, fullJobInfos[0].Job.Id)

	// Check that details are missing when not requested
	require.Nil(t, liteJobInfos[0].Details)
	require.Equal(t, pipeline, liteJobInfos[0].Job.Pipeline.Name)

	// Check that all fields are present when requested
	require.NotNil(t, fullJobInfos[0].Details)
	require.NotNil(t, fullJobInfos[0].Details.Transform)
	require.NotNil(t, fullJobInfos[0].Details.Input)
	require.Equal(t, pipeline, fullJobInfos[0].Job.Pipeline.Name)
}
func TestListJobSetPaged(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestListJobPaged_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline1Name := tu.UniqueString("pipeline1")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline1Name,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	pipeline2Name := tu.UniqueString("pipeline2")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline2Name,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", pipeline1Name),
		},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, pipeline1Name, "/*"),
		"",
		false,
	))

	numFiles := 7
	for i := 0; i < numFiles; i++ {
		commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(commit, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)), client.WithAppendPutFile()))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
		commitInfos, err := c.WaitCommitSetAll(commit.Id)
		require.NoError(t, err)
		require.Equal(t, 7, len(commitInfos))
	}
	jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipeline1Name, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 8, len(jobInfos))
	jobInfos, err = c.ListJob(pfs.DefaultProjectName, pipeline2Name, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 8, len(jobInfos))
	listJobSetRequest := &pps.ListJobSetRequest{}
	client, err := c.PpsAPIClient.ListJobSet(c.Ctx(), listJobSetRequest)
	require.NoError(t, err)
	allJobSetInfos, err := grpcutil.Collect[*pps.JobSetInfo](client, 1000)
	require.NoError(t, err)
	require.Equal(t, 9, len(allJobSetInfos))

	// Test pagination
	var pagedJSIs []*pps.JobSetInfo
	// get first page
	listJobSetRequest = &pps.ListJobSetRequest{Number: 3}
	client, err = c.PpsAPIClient.ListJobSet(c.Ctx(), listJobSetRequest)
	require.NoError(t, err)
	jsis, err := grpcutil.Collect[*pps.JobSetInfo](client, 1000)
	require.NoError(t, err)
	require.Equal(t, 3, len(jsis))
	pagedJSIs = append(pagedJSIs, jsis...)
	// get next two pages
	for i := 0; i < 2; i++ {
		listJobSetRequest = &pps.ListJobSetRequest{Number: 3, PaginationMarker: jsis[len(jsis)-1].Jobs[1].Created}
		client, err = c.PpsAPIClient.ListJobSet(c.Ctx(), listJobSetRequest)
		require.NoError(t, err)
		jsis, err = grpcutil.Collect[*pps.JobSetInfo](client, 1000)
		require.NoError(t, err)
		require.Equal(t, 3, len(jsis))
		pagedJSIs = append(pagedJSIs, jsis...)
	}
	require.Equal(t, len(allJobSetInfos), len(pagedJSIs))
	var reverseJIs []*pps.JobSetInfo
	// get last page
	listJobSetRequest = &pps.ListJobSetRequest{Number: 3, Reverse: true}
	client, err = c.PpsAPIClient.ListJobSet(c.Ctx(), listJobSetRequest)
	require.NoError(t, err)
	jsis, err = grpcutil.Collect[*pps.JobSetInfo](client, 1000)
	require.NoError(t, err)
	require.Equal(t, 3, len(jsis))
	reverseJIs = append(reverseJIs, jsis...)
	// get previous two pages
	for i := 0; i < 2; i++ {
		listJobSetRequest = &pps.ListJobSetRequest{Number: 3, Reverse: true, PaginationMarker: jsis[2].Jobs[0].Created}
		client, err = c.PpsAPIClient.ListJobSet(c.Ctx(), listJobSetRequest)
		require.NoError(t, err)
		jsis, err = grpcutil.Collect[*pps.JobSetInfo](client, 1000)
		require.NoError(t, err)
		require.Equal(t, 3, len(jsis))
		reverseJIs = append(reverseJIs, jsis...)
	}
	require.Equal(t, len(allJobSetInfos), len(reverseJIs))
}

func TestListJobSetWithProjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	client, _ := minikubetestenv.AcquireCluster(t)

	inputRepo := tu.UniqueString("repo")
	require.NoError(t, client.CreateRepo(pfs.DefaultProjectName, inputRepo))

	// pipeline1 is in default project and takes inputRepo as input
	// pipeline2 is in a non-default project, and takes pipeline1 as input
	project := tu.UniqueString("project-")
	pipeline1, pipeline2 := tu.UniqueString("pipeline1-"), tu.UniqueString("pipeline2-")
	require.NoError(t, client.CreateProject(project))
	require.NoError(t, client.CreatePipeline(
		pfs.DefaultProjectName,
		pipeline1,
		"", /* default image*/
		[]string{"cp", "-r", "/pfs/in", "/pfs/out"},
		nil, /* stdin */
		nil, /* spec */
		&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: inputRepo, Glob: "/*", Name: "in"}},
		"",   /* output */
		true, /* update */
	))
	require.NoError(t, client.CreatePipeline(
		project,
		pipeline2,
		"", /* default image*/
		[]string{"cp", "-r", "/pfs/in", "/pfs/out"},
		nil, /* stdin */
		nil, /* spec */
		&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: pipeline1, Glob: "/*", Name: "in"}},
		"",   /* output */
		true, /* update */
	))
	require.NoError(t, client.PutFile(
		&pfs.Commit{Branch: &pfs.Branch{
			Name: "master",
			Repo: &pfs.Repo{
				Name:    inputRepo,
				Type:    pfs.UserRepoType,
				Project: &pfs.Project{Name: pfs.DefaultProjectName},
			}}},
		"/file",
		strings.NewReader("test data"),
	))
	require.NoErrorWithinT(t, time.Minute, func() error {
		_, err := client.WaitCommit(project, pipeline2, "master", "")
		return err
	})

	// don't filter on any projects
	rpcClient, err := client.ListJobSet(client.Ctx(), &pps.ListJobSetRequest{})
	require.NoError(t, err)
	gotPipelines := []string{}
	require.NoError(t, grpcutil.ForEach[*pps.JobSetInfo](rpcClient, func(jobSetInfo *pps.JobSetInfo) error {
		for _, job := range jobSetInfo.Jobs {
			gotPipelines = append(gotPipelines, job.Job.Pipeline.Name)
		}
		return nil
	}))
	require.OneOfEquals(t, pipeline1, gotPipelines)
	require.OneOfEquals(t, pipeline2, gotPipelines)

	// filter on default project
	rpcClient, err = client.ListJobSet(client.Ctx(), &pps.ListJobSetRequest{Projects: []*pfs.Project{{Name: pfs.DefaultProjectName}}})
	require.NoError(t, err)
	gotPipelines = []string{}
	require.NoError(t, grpcutil.ForEach[*pps.JobSetInfo](rpcClient, func(jobSetInfo *pps.JobSetInfo) error {
		for _, job := range jobSetInfo.Jobs {
			gotPipelines = append(gotPipelines, job.Job.Pipeline.Name)
		}
		return nil
	}))
	require.OneOfEquals(t, pipeline1, gotPipelines)
	require.NoneEquals(t, pipeline2, gotPipelines)

	// filter on non-default project
	rpcClient, err = client.ListJobSet(client.Ctx(), &pps.ListJobSetRequest{Projects: []*pfs.Project{{Name: project}}})
	require.NoError(t, err)
	gotPipelines = []string{}
	require.NoError(t, grpcutil.ForEach[*pps.JobSetInfo](rpcClient, func(jobSetInfo *pps.JobSetInfo) error {
		for _, job := range jobSetInfo.Jobs {
			gotPipelines = append(gotPipelines, job.Job.Pipeline.Name)
		}
		return nil
	}))
	require.OneOfEquals(t, pipeline2, gotPipelines)
	require.NoneEquals(t, pipeline1, gotPipelines)

	// filter on a bogus project
	rpcClient, err = client.ListJobSet(client.Ctx(), &pps.ListJobSetRequest{Projects: []*pfs.Project{{Name: "bogus"}}})
	require.NoError(t, err)
	gotPipelines = []string{}
	require.NoError(t, grpcutil.ForEach[*pps.JobSetInfo](rpcClient, func(jobSetInfo *pps.JobSetInfo) error {
		for _, job := range jobSetInfo.Jobs {
			gotPipelines = append(gotPipelines, job.Job.Pipeline.Name)
		}
		return nil
	}))
	require.Len(t, gotPipelines, 0)

	// filter with jq
	rpcClient, err = client.ListJobSet(client.Ctx(), &pps.ListJobSetRequest{JqFilter: `.state == "JOB_SUCCESS"`})
	require.NoError(t, err)
	gotJobs := 0
	require.NoError(t, grpcutil.ForEach[*pps.JobSetInfo](rpcClient, func(jobSetInfo *pps.JobSetInfo) error {
		for _, job := range jobSetInfo.Jobs {
			require.Equal(t, pps.JobState_JOB_SUCCESS, job.State)
		}
		gotJobs += len(jobSetInfo.Jobs)
		return nil
	}))
	require.Equal(t, 4, gotJobs)

	rpcClient, err = client.ListJobSet(client.Ctx(), &pps.ListJobSetRequest{JqFilter: `.state == "JOB_FAILURE"`})
	require.NoError(t, err)
	gotJobs = 0
	require.NoError(t, grpcutil.ForEach[*pps.JobSetInfo](rpcClient, func(jobSetInfo *pps.JobSetInfo) error {
		gotJobs += len(jobSetInfo.Jobs)
		return nil
	}))
	require.Equal(t, 0, gotJobs)
}

func TestListJobPaged(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestListJobPaged_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipelineName := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	numFiles := 8
	for i := 0; i < numFiles; i++ {
		commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(commit, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)), client.WithAppendPutFile()))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
		commitInfos, err := c.WaitCommitSetAll(commit.Id)
		require.NoError(t, err)
		require.Equal(t, 4, len(commitInfos))
	}

	jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 9, len(jobInfos))

	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	listJobRequest := &pps.ListJobRequest{Pipeline: pipeline}
	client, err := c.PpsAPIClient.ListJob(c.Ctx(), listJobRequest)
	require.NoError(t, err)
	allJobInfos, err := grpcutil.Collect[*pps.JobInfo](client, 1000)
	require.NoError(t, err)
	require.Equal(t, 9, len(allJobInfos))

	// Test pagination
	var pagedJIs []*pps.JobInfo
	// get first page
	listJobRequest = &pps.ListJobRequest{Pipeline: pipeline, Number: 3}
	client, err = c.PpsAPIClient.ListJob(c.Ctx(), listJobRequest)
	require.NoError(t, err)
	jis, err := grpcutil.Collect[*pps.JobInfo](client, 1000)
	require.NoError(t, err)
	require.Equal(t, 3, len(jis))
	pagedJIs = append(pagedJIs, jis...)
	// get next two pages
	for i := 0; i < 2; i++ {
		listJobRequest = &pps.ListJobRequest{Pipeline: pipeline, Number: 3, PaginationMarker: jis[len(jis)-1].Created}
		client, err = c.PpsAPIClient.ListJob(c.Ctx(), listJobRequest)
		require.NoError(t, err)
		jis, err = grpcutil.Collect[*pps.JobInfo](client, 1000)
		require.NoError(t, err)
		require.Equal(t, 3, len(jis))
		pagedJIs = append(pagedJIs, jis...)
	}
	for i, ji := range allJobInfos {
		require.Equal(t, ji.Job.Id, pagedJIs[i].Job.Id)
	}
	var reverseJIs []*pps.JobInfo
	// get last page
	listJobRequest = &pps.ListJobRequest{Pipeline: pipeline, Number: 3, Reverse: true}
	client, err = c.PpsAPIClient.ListJob(c.Ctx(), listJobRequest)
	require.NoError(t, err)
	jis, err = grpcutil.Collect[*pps.JobInfo](client, 1000)
	require.NoError(t, err)
	require.Equal(t, 3, len(jis))
	reverseJIs = append(reverseJIs, jis...)
	// get previous two pages
	for i := 0; i < 2; i++ {
		listJobRequest = &pps.ListJobRequest{Pipeline: pipeline, Number: 3, Reverse: true, PaginationMarker: jis[2].Created}
		client, err = c.PpsAPIClient.ListJob(c.Ctx(), listJobRequest)
		require.NoError(t, err)
		jis, err = grpcutil.Collect[*pps.JobInfo](client, 1000)
		require.NoError(t, err)
		require.Equal(t, 3, len(jis))
		reverseJIs = append(reverseJIs, jis...)
	}
	for i, ji := range allJobInfos {
		require.Equal(t, ji.Job.Id, reverseJIs[len(reverseJIs)-1-i].Job.Id)
	}
}

func TestPipelineEnvVarAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineEnvVarAlias_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"env",
			fmt.Sprintf("cp $%s /pfs/out/", dataRepo),
		},
		nil,
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	numFiles := 10
	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit1.Id)

	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(outputCommit, fmt.Sprintf("file-%d", i), &buf))
		require.Equal(t, fmt.Sprintf("%d", i), buf.String())
	}
}

func TestPipelineEnvVarJoinOn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	repo1 := tu.UniqueString("TestPipelineEnvVarJoinOn_repo1")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo1))
	repo2 := tu.UniqueString("TestPipelineEnvVarJoinOn_repo2")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo2))

	input := client.NewJoinInput(
		client.NewPFSInput(pfs.DefaultProjectName, repo1, "/(*)"),
		client.NewPFSInput(pfs.DefaultProjectName, repo2, "/(*)"),
	)
	input.Join[0].Pfs.Name = "repo1"
	input.Join[1].Pfs.Name = "repo2"
	input.Join[0].Pfs.JoinOn = "$1"
	input.Join[1].Pfs.JoinOn = "$1"

	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo1, "master", ""), "a", strings.NewReader("foo")))
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo2, "master", ""), "a", strings.NewReader("bar")))

	// create pipeline
	pipeline := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"touch /pfs/out/repo1-$PACH_DATUM_repo1_JOIN_ON",
					"touch /pfs/out/repo2-$PACH_DATUM_repo2_JOIN_ON",
				},
			},
			Input:  input,
			Update: true,
		},
	)
	require.NoError(t, err)

	// wait for job and get its datums
	jobs, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, 0, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, pipeline, jobs[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)

	// check the value of JOIN_ON env variable
	require.NoError(t, c.GetFile(jobInfo.OutputCommit, "repo1-a", &bytes.Buffer{}))
	require.NoError(t, c.GetFile(jobInfo.OutputCommit, "repo2-a", &bytes.Buffer{}))
}

func TestPipelineEnvVarGroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// create repos
	repo := tu.UniqueString("TestPipelineEnvVarGroupBy_repo")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))

	input := client.NewGroupInput(
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/(*)-*"),
	)
	input.Group[0].Pfs.Name = "repo"
	input.Group[0].Pfs.GroupBy = "$1"

	commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "a-1", strings.NewReader("foo")))
	require.NoError(t, c.PutFile(commit, "a-2", strings.NewReader("foo")))
	require.NoError(t, c.PutFile(commit, "b-1", strings.NewReader("foo")))
	require.NoError(t, c.PutFile(commit, "b-2", strings.NewReader("foo")))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "", commit.Id))

	// create pipeline
	pipeline := tu.UniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"touch /pfs/out/$PACH_DATUM_repo_GROUP_BY",
				},
			},
			Input:  input,
			Update: true,
		},
	)
	require.NoError(t, err)

	// wait for job to finish
	jobs, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, 0, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, pipeline, jobs[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)

	// check the value of the GROUP_BY env var
	require.NoError(t, c.GetFile(jobInfo.OutputCommit, "a", &bytes.Buffer{}))
	require.NoError(t, c.GetFile(jobInfo.OutputCommit, "b", &bytes.Buffer{}))
}

func TestService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestService_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file1", strings.NewReader("foo")))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	annotations := map[string]string{"foo": "bar"}

	pipeline := tu.UniqueString("pipelineservice")
	// This pipeline sleeps for 10 secs per datum
	require.NoError(t, c.CreatePipelineService(pfs.DefaultProjectName,
		pipeline,
		"trinitronx/python-simplehttpserver",
		[]string{"sh"},
		[]string{
			"cd /pfs",
			// Note: a correct shell script would "exec python ..." here, but we want to
			// test that the server gets properly killed even if the user messes this
			// up.
			"python -m SimpleHTTPServer 8000",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		false,
		8000,
		31800,
		annotations,
	))
	time.Sleep(10 * time.Second)

	// Lookup the address for 'pipelineservice' (different inside vs outside k8s)
	serviceAddr := func() string {
		// Hack: detect if running inside the cluster by looking for this env var
		if _, ok := os.LookupEnv("KUBERNETES_PORT"); !ok {
			// Outside cluster: Re-use external IP and external port defined above
			host := c.GetAddress().Host
			return net.JoinHostPort(host, "31800")
		}
		// Get k8s service corresponding to pachyderm service above--must access
		// via internal cluster IP, but we don't know what that is
		var address string
		kubeClient := tu.GetKubeClient(t)
		backoff.Retry(func() error { //nolint:errcheck
			svcs, err := kubeClient.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
			require.NoError(t, err)
			for _, svc := range svcs.Items {
				// Pachyderm actually generates two services for pipelineservice: one
				// for pachyderm (a ClusterIP service) and one for the user container
				// (a NodePort service, which is the one we want)
				rightName := strings.Contains(svc.Name, "pipelineservice")
				rightType := svc.Spec.Type == v1.ServiceTypeNodePort
				if !rightName || !rightType {
					continue
				}
				host := svc.Spec.ClusterIP
				port := fmt.Sprintf("%d", svc.Spec.Ports[0].Port)
				address = net.JoinHostPort(host, port)

				actualAnnotations := svc.Annotations
				delete(actualAnnotations, "pipelineName")
				if !reflect.DeepEqual(actualAnnotations, annotations) {
					return errors.Errorf(
						"expected service annotations map %#v, got %#v",
						annotations,
						actualAnnotations,
					)
				}

				return nil
			}
			return errors.Errorf("no matching k8s service found")
		}, backoff.NewTestingBackOff())

		require.NotEqual(t, "", address)
		return address
	}()
	httpClient := &http.Client{
		Timeout: 3 * time.Second,
	}
	checkFile := func(expected string) {
		require.NoError(t, backoff.Retry(func() error {
			resp, err := httpClient.Get(fmt.Sprintf("http://%s/%s/file1", serviceAddr, dataRepo))
			if err != nil {
				return errors.EnsureStack(err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode != 200 {
				return errors.Errorf("GET returned %d", resp.StatusCode)
			}
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.EnsureStack(err)
			}
			if string(content) != expected {
				return errors.Errorf("wrong content for file1: expected %s, got %s", expected, string(content))
			}
			return nil
		}, backoff.NewTestingBackOff()))
	}
	checkFile("foo")

	// overwrite file, and check that we can access the new contents
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "file1", strings.NewReader("bar")))
	checkFile("bar")
}

func TestServiceEnvVars(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString(t.Name() + "-input")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "file1", strings.NewReader("foo"), client.WithAppendPutFile()))

	pipeline := tu.UniqueString("pipelineservice")
	_, err := c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Image: "trinitronx/python-simplehttpserver",
				Cmd:   []string{"sh"},
				Stdin: []string{
					"echo ${CUSTOM_ENV_VAR} >/pfs/custom_env_var",
					"cd /pfs",
					"exec python -m SimpleHTTPServer 8000",
				},
				Env: map[string]string{
					"CUSTOM_ENV_VAR": "custom-value",
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input:  client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
			Update: false,
			Service: &pps.Service{
				InternalPort: 8000,
				ExternalPort: 31801,
			},
		})
	require.NoError(t, err)

	// Lookup the address for 'pipelineservice' (different inside vs outside k8s)
	serviceAddr := func() string {
		// Hack: detect if running inside the cluster by looking for this env var
		if _, ok := os.LookupEnv("KUBERNETES_PORT"); !ok {
			// Outside cluster: Re-use external IP and external port defined above
			host := c.GetAddress().Host
			return net.JoinHostPort(host, "31801")
		}
		// Get k8s service corresponding to pachyderm service above--must access
		// via internal cluster IP, but we don't know what that is
		var address string
		kubeClient := tu.GetKubeClient(t)
		backoff.Retry(func() error { //nolint:errcheck
			svcs, err := kubeClient.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
			require.NoError(t, err)
			for _, svc := range svcs.Items {
				// Pachyderm actually generates two services for pipelineservice: one
				// for pachyderm (a ClusterIP service) and one for the user container
				// (a NodePort service, which is the one we want)
				rightName := strings.Contains(svc.Name, "pipelineservice")
				rightType := svc.Spec.Type == v1.ServiceTypeNodePort
				if !rightName || !rightType {
					continue
				}
				host := svc.Spec.ClusterIP
				port := fmt.Sprintf("%d", svc.Spec.Ports[0].Port)
				address = net.JoinHostPort(host, port)
				return nil
			}
			return errors.Errorf("no matching k8s service found")
		}, backoff.NewTestingBackOff())

		require.NotEqual(t, "", address)
		return address
	}()

	var envValue []byte
	require.NoErrorWithinTRetry(t, 2*time.Minute, func() error {
		httpC := http.Client{
			Timeout: 3 * time.Second, // fail fast
		}
		resp, err := httpC.Get(fmt.Sprintf("http://%s/custom_env_var", serviceAddr))
		if err != nil {
			// sleep => don't spam retries. Seems to make test less flaky
			time.Sleep(time.Second)
			return errors.EnsureStack(err)
		}
		if resp.StatusCode != 200 {
			return errors.Errorf("GET returned %d", resp.StatusCode)
		}
		envValue, err = io.ReadAll(resp.Body)
		if err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	})
	require.Equal(t, "custom-value", strings.TrimSpace(string(envValue)))
}

func TestDatumSetSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestDatumSetSpec_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	numFiles := 101
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file%d", i), strings.NewReader("foo"), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	t.Run("number", func(t *testing.T) {
		pipeline := tu.UniqueString("TestDatumSetSpec")
		_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"bash"},
					Stdin: []string{
						fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
					},
				},
				Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
				DatumSetSpec: &pps.DatumSetSpec{Number: 1},
			})
		require.NoError(t, err)

		commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)

		for i := 0; i < numFiles; i++ {
			var buf bytes.Buffer
			require.NoError(t, c.GetFile(commitInfo.Commit, fmt.Sprintf("file%d", i), &buf))
			require.Equal(t, "foo", buf.String())
		}
	})
	t.Run("size", func(t *testing.T) {
		pipeline := tu.UniqueString("TestDatumSetSpec")
		_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"bash"},
					Stdin: []string{
						fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
					},
				},
				Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
				DatumSetSpec: &pps.DatumSetSpec{SizeBytes: 5},
			})
		require.NoError(t, err)

		commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)

		for i := 0; i < numFiles; i++ {
			var buf bytes.Buffer
			require.NoError(t, c.GetFile(commitInfo.Commit, fmt.Sprintf("file%d", i), &buf))
			require.Equal(t, "foo", buf.String())
		}
	})
}

func TestLongDatums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestLongDatums_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("TestLongDatums")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"sleep 2s",
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 4,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	numFiles := 8
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file%d", i), strings.NewReader("foo"), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	outputCommit := client.NewCommit(pfs.DefaultProjectName, pipeline, "master", commit1.Id)

	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(outputCommit, fmt.Sprintf("file%d", i), &buf))
		require.Equal(t, "foo", buf.String())
	}
}

func TestPipelineWithDatumTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineWithDatumTimeout_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))
	timeout := 20
	pipeline := tu.UniqueString("pipeline")
	duration, err := time.ParseDuration(fmt.Sprintf("%vs", timeout))
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"while true; do sleep 1; date; done",
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			DatumTimeout: durationpb.New(duration),
		},
	)
	require.NoError(t, err)

	jobs, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	// Block on the job being complete before we call ListDatum
	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, jobs[0].Job.Pipeline.Name, jobs[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)

	// Now validate the datum timed out properly
	dis, err := c.ListDatumAll(pfs.DefaultProjectName, jobs[0].Job.Pipeline.Name, jobs[0].Job.Id)
	require.NoError(t, err)
	require.Equal(t, 1, len(dis))

	datum, err := c.InspectDatum(pfs.DefaultProjectName, jobs[0].Job.Pipeline.Name, jobs[0].Job.Id, dis[0].Datum.Id)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_FAILED, datum.State)
	// ProcessTime looks like "20 seconds"
	tokens := strings.Split(pretty.Duration(datum.Stats.ProcessTime), " ")
	require.Equal(t, 2, len(tokens))
	seconds, err := strconv.Atoi(tokens[0])
	require.NoError(t, err)
	require.Equal(t, timeout, seconds)
}

func TestListDatumDuringJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestListDatumDuringJob_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)

	fileCount := 10
	for i := 0; i < fileCount; i++ {
		fileName := "file" + strconv.Itoa(i)
		require.NoError(t, c.PutFile(commit1, fileName, strings.NewReader("foo")))
	}

	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))
	timeout := 20
	pipeline := tu.UniqueString("TestListDatumDuringJob_pipeline")
	duration, err := time.ParseDuration(fmt.Sprintf("%vs", timeout))
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"sleep 5;",
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			DatumTimeout: durationpb.New(duration),
			DatumSetSpec: &pps.DatumSetSpec{
				Number: 2, // since we set the DatumSetSpec number to 2, we expect our datums to be processed in 5 datum sets (10 files / 2 files per set)
			},
		},
	)
	require.NoError(t, err)

	var jobInfo *pps.JobInfo
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		return backoff.Retry(func() error {
			jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
			if err != nil {
				return err
			}
			if len(jobInfos) != 1 {
				return errors.Errorf("Expected one job, but got %d: %v", len(jobInfos), jobInfos)
			}
			jobInfo = jobInfos[0]
			return nil
		}, backoff.NewTestingBackOff())
	})

	// initially since no datum chunks have been processed, we receive 0 datums
	dis, err := c.ListDatumAll(pfs.DefaultProjectName, jobInfo.Job.Pipeline.Name, jobInfo.Job.Id)
	require.NoError(t, err)
	require.Equal(t, 0, len(dis))

	// test job progress by waiting until some datums are returned, and verify that it's not all of them
	require.NoErrorWithinT(t, 60*time.Second, func() error {
		return backoff.Retry(func() error {
			dis, err = c.ListDatumAll(pfs.DefaultProjectName, jobInfo.Job.Pipeline.Name, jobInfo.Job.Id)
			if err != nil {
				return err
			}
			if len(dis) == 0 {
				return errors.Errorf("expected some datums to be listed")
			}
			return nil
		}, backoff.NewTestingBackOff())
	})

	// fewer than all the datums have been processed
	require.True(t, len(dis) < fileCount)

	// wait until all datums are processed
	_, err = c.WaitCommitSetAll(jobInfo.Job.Id)
	require.NoError(t, err)

	dis, err = c.ListDatumAll(pfs.DefaultProjectName, jobInfo.Job.Pipeline.Name, jobInfo.Job.Id)
	require.NoError(t, err)
	require.Equal(t, fileCount, len(dis))
}

func TestPipelineWithDatumTimeoutControl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineWithDatumTimeoutControl_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	timeout := 20
	pipeline := tu.UniqueString("pipeline")
	duration, err := time.ParseDuration(fmt.Sprintf("%vs", timeout))
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("sleep %v", timeout-10),
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			DatumTimeout: durationpb.New(duration),
		},
	)
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	jobs, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	// Block on the job being complete before we call ListDatum
	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, jobs[0].Job.Pipeline.Name, jobs[0].Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
}

func TestPipelineWithJobTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineWithDatumTimeout_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	numFiles := 2
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(commit1, fmt.Sprintf("file-%v", i), strings.NewReader("foo"), client.WithAppendPutFile()))
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))
	timeout := 20
	pipeline := tu.UniqueString("pipeline")
	duration, err := time.ParseDuration(fmt.Sprintf("%vs", timeout))
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("sleep %v", timeout), // we have 2 datums, so the total exec time will more than double the timeout value
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:      client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			JobTimeout: durationpb.New(duration),
		},
	)
	require.NoError(t, err)

	// Wait for the job to get scheduled / appear in listjob
	var job *pps.JobInfo
	require.NoErrorWithinTRetry(t, 90*time.Second, func() error {
		jobs, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
		if err != nil {
			return fmt.Errorf("list job: %w", err) //nolint:wrapcheck

		}
		if got, want := len(jobs), 1; got != want {
			return fmt.Errorf("job count: got %v want %v (jobs: %v)", got, want, jobs) //nolint:wrapcheck
		}
		job = jobs[0]
		return nil
	}, "pipeline should appear in list jobs")

	// Block on the job being complete before we call ListDatum
	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, job.Job.Pipeline.Name, job.Job.Id, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_KILLED.String(), jobInfo.State.String())
	started := jobInfo.Started.AsTime()
	finished := jobInfo.Finished.AsTime()
	require.True(t, math.Abs((finished.Sub(started)-(time.Second*20)).Seconds()) <= 1.0)
}

func TestPipelineEmptyInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	dataRepo := tu.UniqueString("TestPipelineEmptyInput_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	commit, err := c.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
		Branch:      client.NewBranch(pfs.DefaultProjectName, dataRepo, "master"),
		Description: "test commit description in 'start commit'",
	})
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
	pipeline := tu.UniqueString("TestPipelineEmptyInput")
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{Cmd: []string{"true"}},
			Input:     client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		})
	require.NoError(t, err)
	// update pipeline to empty input
	_, err = c.PpsAPIClient.CreatePipeline(
		ctx,
		&pps.CreatePipelineRequest{
			Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{Cmd: []string{"true"}},
			Update:    true,
		})
	require.NoError(t, err)
	// restore pipeline by setting its input back
	_, err = c.PpsAPIClient.CreatePipeline(
		ctx,
		&pps.CreatePipelineRequest{
			Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{Cmd: []string{"true"}},
			Input:     client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
			Update:    true,
		})
	require.NoError(t, err)
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	jis, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, false)
	require.NoError(t, err)
	require.Equal(t, 3, len(jis))
	require.Equal(t, jis[0].State, pps.JobState_JOB_SUCCESS)
}

func TestCommitDescription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dataRepo := tu.UniqueString("TestCommitDescription")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// Test putting a message in StartCommit
	commit, err := c.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
		Branch:      client.NewBranch(pfs.DefaultProjectName, dataRepo, "master"),
		Description: "test commit description in 'start commit'",
	})
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id)
	require.NoError(t, err)
	require.Equal(t, "test commit description in 'start commit'", commitInfo.Description)
	require.NoError(t, pfspretty.PrintDetailedCommitInfo(os.Stdout, pfspretty.NewPrintableCommitInfo(commitInfo)))

	// Test putting a message in FinishCommit
	commit, err = c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
		Commit:      commit,
		Description: "test commit description in 'finish commit'",
	})
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", commit.Id)
	require.NoError(t, err)
	require.Equal(t, "test commit description in 'finish commit'", commitInfo.Description)
	require.NoError(t, pfspretty.PrintDetailedCommitInfo(os.Stdout, pfspretty.NewPrintableCommitInfo(commitInfo)))

	// Test overwriting a commit message
	commit, err = c.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
		Branch:      client.NewBranch(pfs.DefaultProjectName, dataRepo, "master"),
		Description: "test commit description in 'start commit'",
	})
	require.NoError(t, err)
	_, err = c.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
		Commit:      commit,
		Description: "test commit description in 'finish commit' that overwrites",
	})
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id)
	require.NoError(t, err)
	require.Equal(t, "test commit description in 'finish commit' that overwrites", commitInfo.Description)
	require.NoError(t, pfspretty.PrintDetailedCommitInfo(os.Stdout, pfspretty.NewPrintableCommitInfo(commitInfo)))
}

func TestPipelineDescription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineDescription_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	description := "pipeline description"
	pipeline := tu.UniqueString("TestPipelineDescription")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline:    client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform:   &pps.Transform{Cmd: []string{"true"}},
			Description: description,
			Input:       client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		})
	require.NoError(t, err)
	pi, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	require.Equal(t, description, pi.Details.Description)
}

// TestCancelJob creates a long-running job and then kills it, testing
// that the user process is killed.
func TestCancelJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// Create an input repo
	repo := tu.UniqueString("TestCancelJob")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))

	// Create an input commit
	commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "/time", strings.NewReader("600"), client.WithAppendPutFile()))
	require.NoError(t, c.PutFile(commit, "/data", strings.NewReader("commit data"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "", commit.Id))

	// Create sleep + copy pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"sleep `cat /pfs/*/time`",
			"cp /pfs/*/data /pfs/out/",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/"),
		"",
		false,
	))

	// Wait until PPS has started processing commit
	var jobInfo *pps.JobInfo
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		return backoff.Retry(func() error {
			jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
			if err != nil {
				return err
			}
			if len(jobInfos) != 1 {
				return errors.Errorf("Expected one job, but got %d: %v", len(jobInfos), jobInfos)
			}
			jobInfo = jobInfos[0]
			return nil
		}, backoff.NewTestingBackOff())
	})

	// stop the job
	require.NoError(t, c.StopJob(pfs.DefaultProjectName, jobInfo.Job.Pipeline.Name, jobInfo.Job.Id))

	// Wait until the job is cancelled
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		return backoff.Retry(func() error {
			updatedJobInfo, err := c.InspectJob(pfs.DefaultProjectName, jobInfo.Job.Pipeline.Name, jobInfo.Job.Id, false)
			if err != nil {
				return err
			}
			if updatedJobInfo.State != pps.JobState_JOB_KILLED {
				return errors.Errorf("job %s is still running, but should be KILLED", jobInfo.Job.Id)
			}
			return nil
		}, backoff.NewTestingBackOff())
	})

	// Create one more commit to make sure the pipeline can still process input
	// commits
	commit2, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.DeleteFile(commit2, "/time"))
	require.NoError(t, c.PutFile(commit2, "/time", strings.NewReader("1"), client.WithAppendPutFile()))
	require.NoError(t, c.DeleteFile(commit2, "/data"))
	require.NoError(t, c.PutFile(commit2, "/data", strings.NewReader("commit 2 data"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "", commit2.Id))

	// Flush commit2, and make sure the output is as expected
	commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", commit2.Id)
	require.NoError(t, err)

	buf := bytes.Buffer{}
	err = c.GetFile(commitInfo.Commit, "/data", &buf)
	require.NoError(t, err)
	require.Equal(t, "commit 2 data", buf.String())
}

// TestCancelManyJobs creates many jobs to test that the handling of many
// incoming job events is correct. Each job comes up (which tests that that
// cancelling job 'a' does not cancel subsequent job 'b'), must be the only job
// running (which tests that only one job can run at a time), and then is
// cancelled.
func TestCancelManyJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// Create an input repo
	repo := tu.UniqueString("TestCancelManyJobs")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))

	// Create sleep pipeline
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"sleep", "600"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		"",
		false,
	))

	// Create 10 input commits, to spawn 10 jobs
	var commits []*pfs.Commit
	for i := 0; i < 10; i++ {
		commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo")))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "", commit.Id))
		commits = append(commits, commit)
	}

	// For each expected job: watch to make sure the input job comes up, make
	// sure that it's the only job running, then cancel it
	for _, commit := range commits {
		// Wait until PPS has started processing commit
		var jobInfo *pps.JobInfo
		require.NoErrorWithinT(t, 30*time.Second, func() error {
			return backoff.Retry(func() error {
				jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipeline, []*pfs.Commit{commit}, -1, true)
				if err != nil {
					return err
				}
				if len(jobInfos) != 1 {
					return errors.Errorf("Expected one job, but got %d: %v", len(jobInfos), jobInfos)
				}
				jobInfo = jobInfos[0]
				return nil
			}, backoff.NewTestingBackOff())
		})

		// Stop the job
		require.NoError(t, c.StopJob(pfs.DefaultProjectName, jobInfo.Job.Pipeline.Name, jobInfo.Job.Id))

		// Check that the job is now killed
		require.NoErrorWithinT(t, 30*time.Second, func() error {
			return backoff.Retry(func() error {
				// TODO(msteffen): once github.com/pachyderm/pachyderm/v2/pull/2642 is
				// submitted, change ListJob here to filter on commit1 as the input commit,
				// rather than inspecting the input in the test
				updatedJobInfo, err := c.InspectJob(pfs.DefaultProjectName, jobInfo.Job.Pipeline.Name, jobInfo.Job.Id, false)
				if err != nil {
					return err
				}
				if updatedJobInfo.State != pps.JobState_JOB_KILLED {
					return errors.Errorf("job %s is still running, but should be KILLED", jobInfo.Job.Id)
				}
				return nil
			}, backoff.NewTestingBackOff())
		})
	}
}

// TestDropCommit tests that dropping a source commit results with the downstream job being stopped.
func TestDropCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	c = c.WithDefaultTransformUser("1000")
	ctx := pctx.TestContext(t)
	// Create a source repo
	repo := tu.UniqueString("TestDropCommit_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	masterCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, c.PutFile(masterCommit, "f", strings.NewReader("")))
	// Create a pipeline that creates jobs that sleep.
	pipeline := tu.UniqueString("TestDropCommit")
	require.NoError(t, c.CreatePipeline(
		pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"/bin/bash"},
		[]string{"sleep 3600"},
		&pps.ParallelismSpec{Constant: 1},
		client.NewPFSInput(pfs.DefaultProjectName, repo, "*"),
		"",
		false,
	))
	bi, err := c.InspectBranch(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	_, err = c.PfsAPIClient.DropCommit(ctx, &pfs.DropCommitRequest{
		Commit:    bi.Head,
		Recursive: true,
	})
	require.NoError(t, err)
	// Check that the downstream job was stopped.
	jobInfo, err := c.InspectJob(pfs.DefaultProjectName, pipeline, bi.Head.Id, false)
	// The job may be deleted by the worker master before it spins down.
	if !errutil.IsNotFoundError(err) {
		require.NoError(t, err)
		require.Equal(t, pps.JobState_JOB_KILLED, jobInfo.State)
	}
}

func TestDeleteSpecRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestDeleteSpecRepo_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("TestSimplePipeline")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"echo", "foo"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		"",
		false,
	))
	_, err := c.PfsAPIClient.DeleteRepo(
		c.Ctx(),
		&pfs.DeleteRepoRequest{
			Repo: client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.SpecRepoType),
		})
	require.YesError(t, err)
}

func TestDontReadStdin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestDontReadStdin_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("TestDontReadStdin")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"true"},
		[]string{"stdin that will never be read"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		"",
		false,
	))
	numCommits := 20
	for i := 0; i < numCommits; i++ {
		commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", ""))
		jobInfos, err := c.WaitJobSetAll(commit.Id, false)
		require.NoError(t, err)
		require.Equal(t, 1, len(jobInfos))
		require.Equal(t, jobInfos[0].State.String(), pps.JobState_JOB_SUCCESS.String())
	}
}

func TestStatsDeleteAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineWithStats_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"cp", fmt.Sprintf("/pfs/%s/file", dataRepo), "/pfs/out"},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
		})
	require.NoError(t, err)

	commit, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	jis, err := c.WaitJobSetAll(commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	require.Equal(t, pps.JobState_JOB_SUCCESS.String(), jis[0].State.String())
	require.NoError(t, c.DeleteAll(c.Ctx()))

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"cp", fmt.Sprintf("/pfs/%s/file", dataRepo), "/pfs/out"},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)

	commit, err = c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit.Id))

	jis, err = c.WaitJobSetAll(commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	require.Equal(t, pps.JobState_JOB_SUCCESS.String(), jis[0].State.String())
	require.NoError(t, c.DeleteAll(c.Ctx()))
}

func TestRapidUpdatePipelines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, namespace := minikubetestenv.AcquireCluster(t)
	pipeline := tu.UniqueString("pipeline-")
	cronInput := client.NewCronInput("time", "@every 20s")
	cronInput.Cron.Overwrite = true

	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"/bin/bash"},
		[]string{"cp /pfs/time/* /pfs/out/"},
		nil,
		cronInput,
		"",
		false,
	))
	err := waitForOnePodReady(t, context.Background(), namespace, fmt.Sprintf("pipelineName=%s", pipeline))
	require.NoError(t, err)

	for i := 0; i < 20; i++ {
		_, err := c.PpsAPIClient.CreatePipeline(
			context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:   []string{"/bin/bash"},
					Stdin: []string{"cp /pfs/time/* /pfs/out/"},
				},
				Input:     cronInput,
				Update:    true,
				Reprocess: true,
			})
		require.NoError(t, err)
	}
	// TODO ideally this test would not take 5 minutes (or even 3 minutes)
	require.NoErrorWithinTRetryConstant(t, 5*time.Minute, func() error {
		jis, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, -1, true)
		if err != nil {
			return err
		}
		if len(jis) < 6 {
			return errors.Errorf("should have more than 6 jobs in 5 minutes")
		}
		for i := 0; i < 6; i++ {
			if jis[i].Started == nil {
				return errors.Errorf("not enough jobs have been started yet")
			}
		}
		for i := 0; i < 5; i++ {
			difference := jis[i].Started.Seconds - jis[i+1].Started.Seconds
			if difference < 10 {
				return errors.Errorf("jobs too close together")
			} else if difference > 30 {
				return errors.Errorf("jobs too far apart")
			}
		}
		return nil
	}, time.Second*10)
}

func TestDatumTries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestDatumTries_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	tries := int64(5)
	pipeline := tu.UniqueString("TestSimplePipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"unknown"}, // Cmd fails because "unknown" isn't a known command.
			},
			Input:      client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/"),
			DatumTries: tries,
		})
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)

	jobInfos, err := c.WaitJobSetAll(commitInfo.Commit.Id, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))

	require.NoErrorWithinTRetry(t, 5*time.Minute, func() error {
		iter := c.GetLogs(pfs.DefaultProjectName, pipeline, jobInfos[0].Job.Id, nil, "", false, false, 0)
		var observedTries int64
		for iter.Next() {
			if strings.Contains(iter.Message().Message, "errored running user code") {
				observedTries++
			}
		}
		if tries != observedTries {
			return errors.Errorf("got %d but expected %d", observedTries, tries)
		}
		return nil
	})
}

func TestInspectJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	_, err := c.PpsAPIClient.InspectJob(context.Background(), &pps.InspectJobRequest{})
	require.YesError(t, err)
	require.True(t, strings.Contains(err.Error(), "must specify a job"))

	repo := tu.UniqueString("TestInspectJob")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	ci, err := c.InspectCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, err)

	_, err = c.InspectJob(pfs.DefaultProjectName, repo, ci.Commit.Id, false)
	require.YesError(t, err)
	require.True(t, errutil.IsNotFoundError(err))
}

func TestPipelineVersions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestPipelineVersions_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("TestPipelineVersions")
	nVersions := 5
	for i := 0; i < nVersions; i++ {
		require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{fmt.Sprintf("%d", i)}, // an obviously illegal command, but the pipeline will never run
			nil,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			"",
			i != 0,
		))
	}

	for i := 0; i < nVersions; i++ {
		pi, err := c.InspectPipeline(pfs.DefaultProjectName, ancestry.Add(pipeline, nVersions-1-i), true)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%d", i), pi.Details.Transform.Cmd[0])
	}
}

func TestDeferredProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestDeferredProcessing_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline1 := tu.UniqueString("TestDeferredProcessing1")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline1),
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo)},
			},
			Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			OutputBranch: "staging",
		})
	require.NoError(t, err)

	pipeline2 := tu.UniqueString("TestDeferredProcessing2")
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
		client.NewPFSInput(pfs.DefaultProjectName, pipeline1, "/*"),
		"",
		false,
	))

	commit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "staging", "")
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "staging", "")
	require.NoError(t, err)

	// The same commitset should be extended after each branch head move
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, dataRepo, "master", "staging", "", nil))

	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, pipeline1, "staging", "")
	require.NoError(t, err)
	commitInfos, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, pipeline1, "master", "staging", "", nil))

	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, pipeline2, "master", "")
	require.NoError(t, err)
	commitInfos, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 6, len(commitInfos))
}

func TestListPipelineAtCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create two pipelines and update 1 of them to represent several DAG states
	dataRepo := tu.UniqueString("TestListPipelineAtCommit_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	pipeline1 := tu.UniqueString("TestListPipelineAtCommit_pipeline1")
	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), basicPipelineReq(pipeline1, dataRepo))
	require.NoError(t, err)
	pipeline2 := tu.UniqueString("TestListPipelineAtCommit_pipeline2")
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), basicPipelineReq(pipeline2, pipeline1))
	require.NoError(t, err)
	updateReq := basicPipelineReq(pipeline1, dataRepo)
	updateReq.Update = true
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), updateReq)
	require.NoError(t, err)
	// assert correct pipeline versions are returned for each commit set ID
	ci, err := c.InspectCommit(pfs.DefaultProjectName, pipeline2, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(ci.Commit.Id)
	require.NoError(t, err)
	commitSets, err := c.ListCommitSet(c.Ctx(), &pfs.ListCommitSetRequest{})
	require.NoError(t, err)
	expected := []map[string]uint64{
		{pipeline1: 2, pipeline2: 1},
		{pipeline1: 1, pipeline2: 1},
		{pipeline1: 1},
	}
	i := 0
	require.NoError(t, grpcutil.ForEach[*pfs.CommitSetInfo](commitSets, func(csi *pfs.CommitSetInfo) error {
		pipelines, err := c.PpsAPIClient.ListPipeline(c.Ctx(), &pps.ListPipelineRequest{
			CommitSet: &pfs.CommitSet{Id: csi.CommitSet.Id},
		})
		require.NoError(t, err)
		count := 0
		require.NoError(t, grpcutil.ForEach[*pps.PipelineInfo](pipelines, func(pi *pps.PipelineInfo) error {
			count++
			require.Equal(t, expected[i][pi.Pipeline.Name], pi.Version)
			return nil
		}))
		require.Equal(t, len(expected[i]), count)
		i++
		return nil
	}))

}

func TestPipelineHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	// create repos
	dataRepo := tu.UniqueString("TestPipelineHistory_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	pipelineName := tu.UniqueString("TestPipelineHistory")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"echo foo >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		true,
	))

	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "file", strings.NewReader("1"), client.WithAppendPutFile()))
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	jis, err := c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jis))

	// Update the pipeline
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"echo bar >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		true,
	))

	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "file", strings.NewReader("2"), client.WithAppendPutFile()))
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	commitInfos, err := c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipelineName), client.NewCommit(pfs.DefaultProjectName, pipelineName, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(jis))
	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 1, true)
	require.NoError(t, err)
	require.Equal(t, 4, len(jis))
	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 4, len(jis))

	// Update the pipeline again
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"echo buzz >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		true,
	))
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 1, true)
	require.NoError(t, err)
	require.Equal(t, 3, len(jis))
	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 2, true)
	require.NoError(t, err)
	require.Equal(t, 5, len(jis))
	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 5, len(jis))

	// Add another pipeline, this shouldn't change the results of the above
	// commands.
	pipelineName2 := tu.UniqueString("TestPipelineHistory2")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName2,
		"",
		[]string{"bash"},
		[]string{"echo foo >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		true,
	))
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 0, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(jis))
	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 1, true)
	require.NoError(t, err)
	require.Equal(t, 3, len(jis))
	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, 2, true)
	require.NoError(t, err)
	require.Equal(t, 5, len(jis))
	jis, err = c.ListJob(pfs.DefaultProjectName, pipelineName, nil, -1, true)
	require.NoError(t, err)
	require.Equal(t, 5, len(jis))

	pipelineInfos, err := c.ListPipeline()
	require.NoError(t, err)
	require.Equal(t, 2, len(pipelineInfos))

	pipelineInfos, err = c.ListPipelineHistory(pfs.DefaultProjectName, "", -1)
	require.NoError(t, err)
	require.Equal(t, 4, len(pipelineInfos))

	pipelineInfos, err = c.ListPipelineHistory(pfs.DefaultProjectName, "", 1)
	require.NoError(t, err)
	require.Equal(t, 3, len(pipelineInfos))

	pipelineInfos, err = c.ListPipelineHistory(pfs.DefaultProjectName, pipelineName, -1)
	require.NoError(t, err)
	require.Equal(t, 3, len(pipelineInfos))

	pipelineInfos, err = c.ListPipelineHistory(pfs.DefaultProjectName, pipelineName2, -1)
	require.NoError(t, err)
	require.Equal(t, 1, len(pipelineInfos))
}

// TestCreatePipelineErrorNoTransform tests that sending a CreatePipeline
// requests to pachd with no 'pipeline' field doesn't kill pachd
func TestCreatePipelineErrorNoPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// Create input repo
	dataRepo := tu.UniqueString(t.Name() + "-data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// Create pipeline w/ no pipeline field--make sure we get a response
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: nil,
			Transform: &pps.Transform{
				Cmd:   []string{"/bin/bash"},
				Stdin: []string{`cat foo >/pfs/out/file`},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.YesError(t, err)
	require.Matches(t, "request.Pipeline", err.Error())
}

// TestCreatePipelineErrorNoTransform tests that sending a CreatePipeline
// requests to pachd with no 'transform' or 'pipeline' field doesn't kill pachd
func TestCreatePipelineError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// Create input repo
	dataRepo := tu.UniqueString(t.Name() + "-data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	// Create pipeline w/ no transform--make sure we get a response (& make sure
	// it explains the problem)
	pipeline := tu.UniqueString("no-transform-")
	_, err := c.PpsAPIClient.CreatePipelineV2(
		context.Background(),
		&pps.CreatePipelineV2Request{
			CreatePipelineRequestJson: fmt.Sprintf(`{"pipeline": {"project": {"name": %q}, "name": %q}, "transform": null, "input": {"pfs": {"project": %q, "repo": %q, "glob": %q}}}`, pfs.DefaultProjectName, pipeline, pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.YesError(t, err)
	require.Matches(t, "transform", err.Error())
}

// TestCreatePipelineErrorNoCmd tests that sending a CreatePipeline request to
// pachd with no 'transform.cmd' field doesn't kill pachd
func TestCreatePipelineErrorNoCmd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// Create input data
	dataRepo := tu.UniqueString(t.Name() + "-data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	// create pipeline
	pipeline := tu.UniqueString("no-cmd-")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd:   nil,
				Stdin: []string{`cat foo >/pfs/out/file`},
			},
			Input: client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		})
	require.NoError(t, err)
	time.Sleep(5 * time.Second) // give pipeline time to start

	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_FAILURE {
			return errors.Errorf("pipeline should be in state FAILURE, not: %s", pipelineInfo.State.String())
		}
		return nil
	})
}

// TestPodPatchUnmarshalling tests the fix for issues #3483, by adding a
// PodPatch to a pipeline spec and making sure it's applied correctly
func TestPodPatchUnmarshalling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)

	// Create input data
	dataRepo := tu.UniqueString(t.Name() + "-data-")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	// create pipeline
	pipeline := tu.UniqueString("pod-patch-")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{"cp /pfs/in/* /pfs/out"},
			},
			Input: &pps.Input{Pfs: &pps.PFSInput{
				Name: "in", Repo: dataRepo, Glob: "/*",
			}},
			PodPatch: `[
				{
				  "op": "add",
				  "path": "/volumes/0",
				  "value": {
				    "name": "vol0",
				    "hostPath": {
				      "path": "/volumePath"
				}}}]`,
		})
	require.NoError(t, err)

	commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfo.Commit, "file", &buf))
	require.Equal(t, "foo", buf.String())

	pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
	require.NoError(t, err)

	// make sure 'vol0' is correct in the pod spec
	var volumes []v1.Volume
	kubeClient := tu.GetKubeClient(t)
	require.NoError(t, backoff.Retry(func() error {
		podList, err := kubeClient.CoreV1().Pods(ns).List(
			context.Background(),
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{
						"app":             "pipeline",
						"pipelineName":    pipelineInfo.Pipeline.Name,
						"pipelineVersion": fmt.Sprint(pipelineInfo.Version),
					},
				)),
			})
		if err != nil {
			return errors.EnsureStack(err) // retry
		}
		if len(podList.Items) != 1 || len(podList.Items[0].Spec.Volumes) == 0 {
			return errors.Errorf("could not find volumes for pipeline %s", pipelineInfo.Pipeline.Name)
		}
		volumes = podList.Items[0].Spec.Volumes
		return nil // no more retries
	}, backoff.NewTestingBackOff()))
	// Make sure a CPU and Memory request are both set
	for _, vol := range volumes {
		require.True(t,
			vol.VolumeSource.HostPath == nil || vol.VolumeSource.EmptyDir == nil)
		if vol.Name == "vol0" {
			require.True(t, vol.VolumeSource.HostPath.Path == "/volumePath")
		}
	}
}

func TestSecrets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	b := []byte(
		`{
			"kind": "Secret",
			"apiVersion": "v1",
			"metadata": {
				"name": "test-secret",
				"creationTimestamp": null
			},
			"data": {
				"mykey": "bXktdmFsdWU="
			}
		}`)
	require.NoError(t, c.CreateSecret(b))

	secretInfo, err := c.InspectSecret("test-secret")
	secretInfo.CreationTimestamp = nil
	require.NoError(t, err)
	require.Equal(t, &pps.SecretInfo{
		Secret: &pps.Secret{
			Name: "test-secret",
		},
		Type:              "Opaque",
		CreationTimestamp: nil,
	}, secretInfo)

	secretInfos, err := c.ListSecret()
	require.NoError(t, err)
	initialLength := len(secretInfos)

	require.NoError(t, c.DeleteSecret("test-secret"))

	secretInfos, err = c.ListSecret()
	require.NoError(t, err)
	require.Equal(t, initialLength-1, len(secretInfos))

	_, err = c.InspectSecret("test-secret")
	require.YesError(t, err)
}

// Test that an unauthenticated user can't call secrets APIS
func TestSecretsUnauthenticated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	// Get an unauthenticated client
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	c.SetAuthToken("")

	b := []byte(
		`{
			"kind": "Secret",
			"apiVersion": "v1",
			"metadata": {
				"name": "test-secret",
				"creationTimestamp": null
			},
			"data": {
				"mykey": "bXktdmFsdWU="
			}
		}`)

	err := c.CreateSecret(b)
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())

	_, err = c.InspectSecret("test-secret")
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())

	_, err = c.ListSecret()
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())

	err = c.DeleteSecret("test-secret")
	require.YesError(t, err)
	require.Matches(t, "no authentication token", err.Error())
}

func TestCopyOutToIn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestCopyOutToIn_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	pipeline := tu.UniqueString("TestCopyOutToIn")
	pipelineCommit := client.NewCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp -R /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(dataCommit, "file2", pipelineCommit, "file", client.WithAppendCopyFile()))
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	require.NoError(t, c.PutFile(dataCommit, "file2", strings.NewReader("foo"), client.WithAppendPutFile()))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(pipelineCommit, "file2", &buf))
	require.Equal(t, "foo", buf.String())

	mfc, err := c.NewModifyFileClient(dataCommit)
	require.NoError(t, err)
	require.NoError(t, mfc.PutFile("dir/file3", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, mfc.PutFile("dir/file4", strings.NewReader("bar"), client.WithAppendPutFile()))
	require.NoError(t, mfc.Close())

	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	require.NoError(t, c.CopyFile(dataCommit, "dir2", pipelineCommit, "dir", client.WithAppendCopyFile()))

	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	buf.Reset()
	require.NoError(t, c.GetFile(pipelineCommit, "dir/file3", &buf))
	require.Equal(t, "foo", buf.String())
	buf.Reset()
	require.NoError(t, c.GetFile(pipelineCommit, "dir/file4", &buf))
	require.Equal(t, "bar", buf.String())
}

func TestKeepRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t, minikubetestenv.UseNewClusterOption)

	dataRepo := tu.UniqueString("TestKeepRepo_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	pipeline := tu.UniqueString("TestKeepRepo")
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	_, err = c.PpsAPIClient.DeletePipeline(c.Ctx(), &pps.DeletePipelineRequest{
		Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
		KeepRepo: true,
	})
	require.NoError(t, err)
	_, err = c.InspectRepo(pfs.DefaultProjectName, pipeline)
	require.NoError(t, err)

	_, err = c.PfsAPIClient.InspectRepo(c.Ctx(), &pfs.InspectRepoRequest{
		Repo: client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.SpecRepoType),
	})
	require.YesError(t, err)
	require.True(t, errutil.IsNotFoundError(err))

	_, err = c.PfsAPIClient.InspectRepo(c.Ctx(), &pfs.InspectRepoRequest{
		Repo: client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.MetaRepoType),
	})
	require.YesError(t, err)
	require.True(t, errutil.IsNotFoundError(err))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(client.NewCommit(pfs.DefaultProjectName, pipeline, "master", ""), "file", &buf))
	require.Equal(t, "foo", buf.String())
	require.YesError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
		"",
		false,
	))
}

// Regression test to make sure that pipeline creation doesn't crash pachd due to missing fields
func TestMalformedPipeline(t *testing.T) {
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	pipelineName := tu.UniqueString("MalformedPipeline")

	var err error
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{})
	require.YesError(t, err)
	require.Matches(t, "request.Pipeline cannot be nil", err.Error())

	_, err = c.PpsAPIClient.CreatePipelineV2(c.Ctx(), &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: fmt.Sprintf(`{"pipeline": {"project": {"name": %q}, "name": %q}, "transform": null}`, pfs.DefaultProjectName, pipelineName)},
	)
	require.YesError(t, err)
	require.Matches(t, "must specify a transform", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{},
	})
	require.YesError(t, err)
	require.Matches(t, "no input set", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Service: &pps.Service{
			Type: string(v1.ServiceTypeNodePort),
		},
		ParallelismSpec: &pps.ParallelismSpec{},
	})
	require.YesError(t, err)
	require.Matches(t, "services can only be run with a constant parallelism of 1", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:   client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform:  &pps.Transform{},
		SpecCommit: &pfs.Commit{},
	})
	require.YesError(t, err)
	require.Matches(t, "cannot resolve commit with no repo", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:   client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform:  &pps.Transform{},
		SpecCommit: &pfs.Commit{Branch: &pfs.Branch{}},
	})
	require.YesError(t, err)
	require.Matches(t, "cannot resolve commit with no repo", err.Error())

	dataRepo := tu.UniqueString("TestMalformedPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, c.PutFile(dataCommit, "file", strings.NewReader("foo"), client.WithAppendPutFile()))

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{}},
	})
	require.YesError(t, err)
	require.Matches(t, "input must specify a name", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{Name: "data"}},
	})
	require.YesError(t, err)
	require.Matches(t, "input must specify a repo", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{Repo: dataRepo}},
	})
	require.YesError(t, err)
	require.Matches(t, "input must specify a glob", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     client.NewPFSInput(pfs.DefaultProjectName, "out", "/*"),
	})
	require.YesError(t, err)
	require.Matches(t, "input cannot be named out", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{Name: "out", Repo: dataRepo, Glob: "/*"}},
	})
	require.YesError(t, err)
	require.Matches(t, "input cannot be named out", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Pfs: &pps.PFSInput{Name: "data", Repo: "dne", Glob: "/*"}},
	})
	require.YesError(t, err)
	require.Matches(t, "dne[^ ]* not found", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input: client.NewCrossInput(
			client.NewPFSInput(pfs.DefaultProjectName, "foo", "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, "foo", "/*"),
		),
	})
	require.YesError(t, err)
	require.Matches(t, "name \"foo\" was used more than once", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Cron: &pps.CronInput{}},
	})
	require.YesError(t, err)
	require.Matches(t, "input must specify a name", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Cron: &pps.CronInput{Name: "cron"}},
	})
	require.YesError(t, err)
	require.Matches(t, "Empty spec string", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Cross: []*pps.Input{}},
	})
	require.YesError(t, err)
	require.Matches(t, "no input set", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Union: []*pps.Input{}},
	})
	require.YesError(t, err)
	require.Matches(t, "no input set", err.Error())

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), &pps.CreatePipelineRequest{
		Pipeline:  client.NewPipeline(pfs.DefaultProjectName, pipelineName),
		Transform: &pps.Transform{},
		Input:     &pps.Input{Join: []*pps.Input{}},
	})
	require.YesError(t, err)
	require.Matches(t, "no input set", err.Error())
}

func TestTrigger(t *testing.T) {
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestTrigger_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, dataRepo, "master", "", "", nil))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	pipeline1 := tu.UniqueString("TestTrigger1")
	pipelineCommit1 := client.NewCommit(pfs.DefaultProjectName, pipeline1, "master", "")
	pipeline2 := tu.UniqueString("TestTrigger2")
	pipelineCommit2 := client.NewCommit(pfs.DefaultProjectName, pipeline2, "master", "")
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
			Branch: "master",
			Size:   "1K",
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
		client.NewPFSInputOpts(pipeline1, pfs.DefaultProjectName, pipeline1, "", "/*", "", "", false, false, &pfs.Trigger{
			Size: "2K",
		}),
		"",
		false,
	))
	// 10 100 byte files = 1K, so the last file should trigger pipeline1, but
	// not pipeline2.
	numFiles := 10
	fileBytes := 100
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(dataCommit, fmt.Sprintf("file%d", i), strings.NewReader(strings.Repeat("a", fileBytes)), client.WithAppendPutFile()))
	}
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	// This should have given us a job, flush to let it complete.
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipeline1, "master", "")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(pipelineCommit1, fmt.Sprintf("file%d", i), &buf))
		require.Equal(t, strings.Repeat("a", fileBytes), buf.String())
	}
	_, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline1), client.NewCommit(pfs.DefaultProjectName, pipeline1, "master", ""), nil, 0)
	require.NoError(t, err)
	// Another 10 100 byte files = 2K, so the last file should trigger both pipelines.
	for i := numFiles; i < 2*numFiles; i++ {
		require.NoError(t, c.PutFile(dataCommit, fmt.Sprintf("file%d", i), strings.NewReader(strings.Repeat("a", fileBytes)), client.WithAppendPutFile()))
		require.NoError(t, err)
	}
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	commitInfos, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipeline1, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipeline2, "master", "")
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	for i := 0; i < numFiles*2; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(pipelineCommit1, fmt.Sprintf("file%d", i), &buf))
		require.Equal(t, strings.Repeat("a", fileBytes), buf.String())
		buf.Reset()
		require.NoError(t, c.GetFile(pipelineCommit2, fmt.Sprintf("file%d", i), &buf))
		require.Equal(t, strings.Repeat("a", fileBytes), buf.String())
	}
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline1), client.NewCommit(pfs.DefaultProjectName, pipeline1, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline2), client.NewCommit(pfs.DefaultProjectName, pipeline2, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
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
		client.NewPFSInputOpts(pipeline1, pfs.DefaultProjectName, pipeline1, "", "/*", "", "", false, false, &pfs.Trigger{
			Size: "3K",
		}),
		"",
		true,
	))
	// Make sure that updating the pipeline reuses the previous branch name
	// rather than creating a new one.
	bis, err := c.ListBranch(pfs.DefaultProjectName, pipeline1)
	require.NoError(t, err)
	require.Equal(t, 2, len(bis))
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline2), client.NewCommit(pfs.DefaultProjectName, pipeline2, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	// Another 30 100 byte files = 3K, so the last file should trigger both pipelines.
	for i := 2 * numFiles; i < 5*numFiles; i++ {
		require.NoError(t, c.PutFile(dataCommit, fmt.Sprintf("file%d", i), strings.NewReader(strings.Repeat("a", fileBytes)), client.WithAppendPutFile()))
	}
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)
	commitInfos, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	commitInfos, err = c.ListCommit(client.NewRepo(pfs.DefaultProjectName, pipeline2), client.NewCommit(pfs.DefaultProjectName, pipeline2, "master", ""), nil, 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
}

func TestListDatum(t *testing.T) {
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	repo1 := tu.UniqueString("TestListDatum1")
	repo2 := tu.UniqueString("TestListDatum2")

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo1))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo2))

	numFiles := 5
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo1, "master", ""), fmt.Sprintf("file-%d", i), strings.NewReader("foo"), client.WithAppendPutFile()))
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo2, "master", ""), fmt.Sprintf("file-%d", i), strings.NewReader("foo"), client.WithAppendPutFile()))
	}

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, repo1, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, repo2, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	dis, err := c.ListDatumInputAll(&pps.Input{
		Cross: []*pps.Input{{
			Pfs: &pps.PFSInput{
				Repo: repo1,
				Glob: "/*",
			},
		}, {
			Pfs: &pps.PFSInput{
				Repo: repo2,
				Glob: "/*",
			},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, 25, len(dis))
}

func TestListDatumFilter(t *testing.T) {
	t.Parallel()
	var (
		ctx   = context.Background()
		c, _  = minikubetestenv.AcquireCluster(t)
		repo1 = tu.UniqueString("TestListDatum1")
		repo2 = tu.UniqueString("TestListDatum2")
	)

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo1))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo2))

	numFiles := 5
	for i := 0; i < numFiles; i++ {
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo1, "master", ""), fmt.Sprintf("file-%d", i), strings.NewReader("foo"), client.WithAppendPutFile()))
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo2, "master", ""), fmt.Sprintf("file-%d", i), strings.NewReader("foo"), client.WithAppendPutFile()))
	}

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, repo1, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, repo2, "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	// filtering for failed should yield zero datums
	s, err := c.PpsAPIClient.ListDatum(ctx, &pps.ListDatumRequest{
		Filter: &pps.ListDatumRequest_Filter{State: []pps.DatumState{pps.DatumState_FAILED}},
		Input: &pps.Input{
			Cross: []*pps.Input{{
				Pfs: &pps.PFSInput{
					Repo: repo1,
					Glob: "/*",
				},
			}, {
				Pfs: &pps.PFSInput{
					Repo: repo2,
					Glob: "/*",
				},
			}},
		},
	})
	require.NoError(t, err)

	var i int
	require.NoError(t, grpcutil.ForEach[*pps.DatumInfo](s, func(d *pps.DatumInfo) error {
		require.NotEqual(t, pps.DatumState_UNKNOWN, d.State)
		i++
		return nil
	}))
	require.Equal(t, 0, i)

	// filtering for only unknowns should yield 25 datums
	s, err = c.PpsAPIClient.ListDatum(ctx, &pps.ListDatumRequest{
		Filter: &pps.ListDatumRequest_Filter{State: []pps.DatumState{pps.DatumState_UNKNOWN}},
		Input: &pps.Input{
			Cross: []*pps.Input{{
				Pfs: &pps.PFSInput{
					Repo: repo1,
					Glob: "/*",
				},
			}, {
				Pfs: &pps.PFSInput{
					Repo: repo2,
					Glob: "/*",
				},
			}},
		},
	})
	require.NoError(t, err)

	require.NoError(t, grpcutil.ForEach[*pps.DatumInfo](s, func(d *pps.DatumInfo) error {
		require.Equal(t, pps.DatumState_UNKNOWN, d.State)
		i++
		return nil
	}))
	require.Equal(t, 25, i)
}

func testDebug(t *testing.T, c *client.APIClient, projectName, repoName string) {
	t.Helper()
	if projectName != "default" {
		require.NoError(t, c.CreateProject(projectName))
	}
	require.NoError(t, c.CreateRepo(projectName, repoName))

	expectedFiles, pipelines := tu.DebugFiles(t, projectName, repoName)

	for i, p := range pipelines {
		cmdStdin := []string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", repoName),
			"sleep 45",
		}
		if i == 0 {
			// We had a bug where generating a debug dump for failed pipelines/jobs would crash pachd.
			// Fail a pipeline on purpose to see if we can still generate debug dump.
			cmdStdin = append(cmdStdin, "exit -1")
		}
		require.NoError(t, c.CreatePipeline(projectName,
			p,
			"",
			[]string{"bash"},
			cmdStdin,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(projectName, repoName, "/*"),
			"",
			false,
		))
	}
	commit1, err := c.StartCommit(projectName, repoName, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(projectName, repoName, "", commit1.Id))

	commitInfos, err := c.WaitCommitSetAll(commit1.Id)
	require.NoError(t, err)
	require.Equal(t, 10, len(commitInfos))

	require.NoErrorWithinT(t, time.Minute, func() error {
		buf := &bytes.Buffer{}
		require.NoError(t, c.Dump(nil, 0, buf))
		gr, err := gzip.NewReader(buf)
		if err != nil {
			return err //nolint:wrapcheck
		}
		defer func() {
			require.NoError(t, gr.Close())
		}()
		// Check that all of the expected files were returned.
		var gotFiles []string
		tr := tar.NewReader(gr)
		for {
			hdr, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err //nolint:wrapcheck
			}
			gotFiles = append(gotFiles, hdr.Name)
			for pattern, g := range expectedFiles {
				if g.Match(hdr.Name) {
					delete(expectedFiles, pattern)
					break
				}
			}
		}
		if len(expectedFiles) > 0 {
			return errors.Errorf("Debug dump has produced %v of the expected files: %v", gotFiles, expectedFiles)
		}
		return nil
	})
}

func TestDebug(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	for _, projectName := range []string{pfs.DefaultProjectName, tu.UniqueString("project")} {
		t.Run(projectName, func(t *testing.T) {
			testDebug(t, c, projectName, tu.UniqueString("TestDebug_data"))
		})
	}
}

func TestUpdateMultiplePipelinesInTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	input := tu.UniqueString("in")
	commit := client.NewCommit(pfs.DefaultProjectName, input, "master", "")
	pipelineA := tu.UniqueString("A")
	pipelineB := tu.UniqueString("B")
	repoB := client.NewRepo(pfs.DefaultProjectName, pipelineB)

	createPipeline := func(c *client.APIClient, input, pipeline string, update bool) error {
		return c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", input)},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(pfs.DefaultProjectName, input, "/*"),
			"",
			update,
		)
	}

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, input))
	require.NoError(t, c.PutFile(commit, "foo", strings.NewReader("bar"), client.WithAppendPutFile()))

	_, err := c.ExecuteInTransaction(func(txnClient *client.APIClient) error {
		require.NoError(t, createPipeline(txnClient, input, pipelineA, false))
		require.NoError(t, createPipeline(txnClient, pipelineA, pipelineB, false))
		return nil
	})
	require.NoError(t, err)
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipelineB, "master", "")
	require.NoError(t, err)

	// now update both
	_, err = c.ExecuteInTransaction(func(txnClient *client.APIClient) error {
		require.NoError(t, createPipeline(txnClient, input, pipelineA, true))
		require.NoError(t, createPipeline(txnClient, pipelineA, pipelineB, true))
		return nil
	})
	require.NoError(t, err)

	_, err = c.WaitCommit(pfs.DefaultProjectName, pipelineB, "master", "")
	require.NoError(t, err)
	commits, err := c.ListCommitByRepo(repoB)
	require.NoError(t, err)
	require.Equal(t, 2, len(commits))

	jobInfos, err := c.ListJob(pfs.DefaultProjectName, pipelineB, nil, -1, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
}

func TestInterruptedUpdatePipelineInTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	inputA := tu.UniqueString("A")
	inputB := tu.UniqueString("B")
	inputC := tu.UniqueString("C")
	pipeline := tu.UniqueString("pipeline")

	createPipeline := func(c *client.APIClient, input string, update bool) error {
		return c.CreatePipeline(pfs.DefaultProjectName,
			pipeline,
			"",
			[]string{"bash"},
			[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", input)},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(pfs.DefaultProjectName, input, "/*"),
			"",
			update,
		)
	}

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, inputA))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, inputB))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, inputC))
	require.NoError(t, createPipeline(c, inputA, false))

	txn, err := c.StartTransaction()
	require.NoError(t, err)

	require.NoError(t, createPipeline(c.WithTransaction(txn), inputB, true))
	require.NoError(t, createPipeline(c, inputC, true))

	_, err = c.FinishTransaction(txn)
	require.NoError(t, err)
	// make sure the final pipeline is the third version, with the input from in the transaction
	pipelineInfo, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, true)
	require.NoError(t, err)
	require.Equal(t, uint64(3), pipelineInfo.Version)
	require.NotNil(t, pipelineInfo.Details.Input.Pfs)
	require.Equal(t, inputB, pipelineInfo.Details.Input.Pfs.Repo)
}

func TestPipelineAutoscaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)
	dataRepo := tu.UniqueString("TestPipelineAutoscaling_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
					"sleep 30",
				},
			},
			Input:           client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			ParallelismSpec: &pps.ParallelismSpec{Constant: 4},
			Autoscaling:     true,
		},
	)
	require.NoError(t, err)
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	time.Sleep(2 * time.Second) // CORE-2056
	fileIndex := 0
	commitNFiles := func(n int) {
		commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
		require.NoError(t, err)
		for i := fileIndex; i < fileIndex+n; i++ {
			err := c.PutFile(commit1, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)))
			require.NoError(t, err)
		}
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "master", commit1.Id))
		fileIndex += n
		replicas := n
		if replicas > 4 {
			replicas = 4
		}
		monitorReplicas(t, c, ns, pipeline, replicas)
	}
	commitNFiles(1)
	commitNFiles(3)
	commitNFiles(8)
}

func TestListDeletedDatums(t *testing.T) {
	// TODO(2.0 optional): Duplicate file paths from different datums no longer allowed.
	t.Skip("Duplicate file paths from different datums no longer allowed.")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	twoRepo := tu.UniqueString("TestListDeletedDatums_Two")
	threeRepo := tu.UniqueString("TestListDeletedDatums_Three")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, twoRepo))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, threeRepo))

	input := client.NewJoinInput(
		client.NewPFSInput(pfs.DefaultProjectName, twoRepo, "/(*)"),
		client.NewPFSInput(pfs.DefaultProjectName, threeRepo, "/(*)"),
	)
	input.Join[0].Pfs.JoinOn = "$1"
	input.Join[1].Pfs.JoinOn = "$1"

	// put files into the repo based on divisibility by 2 and 3
	twoCommit, err := c.StartCommit(pfs.DefaultProjectName, twoRepo, "master")
	require.NoError(t, err)
	threeCommit, err := c.StartCommit(pfs.DefaultProjectName, threeRepo, "master")
	require.NoError(t, err)

	const fileLimit = 12
	for i := 0; i < fileLimit; i++ {
		if i%2 == 0 {
			require.NoError(t, c.PutFile(twoCommit, strconv.Itoa(i), strings.NewReader("fizz")))
		}
		if i%3 == 0 {
			require.NoError(t, c.PutFile(threeCommit, strconv.Itoa(i), strings.NewReader("buzz")))
		}
	}
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, twoRepo, "master", twoCommit.Id))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, threeRepo, "master", threeCommit.Id))

	pipeline := tu.UniqueString("pipeline")
	createAndCheckJoin := func(twoOuter, threeOuter bool) {
		input.Join[0].Pfs.OuterJoin = twoOuter
		input.Join[1].Pfs.OuterJoin = threeOuter
		_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:   []string{"bash"},
					Stdin: []string{"touch pfs/out/ignore"},
				},
				Input:  input,
				Update: true,
			},
		)
		require.NoError(t, err)
		// wait for job and get its datums
		jobs, err := c.ListJob(pfs.DefaultProjectName, pipeline, nil, 0, false)
		require.NoError(t, err)
		require.Equal(t, 1, len(jobs))
		id := jobs[0].Job.Id
		info, err := c.WaitJob(pfs.DefaultProjectName, pipeline, id, false)
		require.NoError(t, err)
		require.Equal(t, pps.JobState_JOB_SUCCESS, info.State)
		datums, err := c.ListDatumAll(pfs.DefaultProjectName, pipeline, id)
		require.NoError(t, err)

		// find which file numbers are present in the datums
		numSet := make(map[int]struct{})
		for _, d := range datums {
			key, err := strconv.Atoi(path.Base(d.Data[0].File.Path))
			require.NoError(t, err)
			numSet[key] = struct{}{}
		}
		// we shouldn't have any duplicate keys across datums for any of these joins
		require.Equal(t, len(numSet), len(datums))

		// now check against what should be included
		for i := 0; i < fileLimit; i++ {
			expected := (i%6 == 0) || (i%2 == 0 && twoOuter) || (i%3 == 0 && threeOuter)
			_, ok := numSet[i]
			require.Equal(t, expected, ok)
		}
	}

	// first a completely outer join with all datums
	createAndCheckJoin(true, true)
	// then the half-outer joins, which should be smaller
	createAndCheckJoin(true, false)
	createAndCheckJoin(false, true)
	// finally the smallest, should only have datums divisible by 6
	createAndCheckJoin(false, false)
}

// TestNonrootPipeline tests that a pipeline can work when it's running as a UID that's not 0 (root).
// This test doesn't work in a typical local deployment because the storage sidecar needs to
// run as root to access the host path where the chunks are stored (at /pach).
func TestNonrootPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Only run this test in CI since we can control the permissions on the hostpath
	if os.Getenv("CIRCLE_REPOSITORY_URL") == "" {
		t.Skip("Skipping non-root test because hostpath permissions can't be verified")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestNonrootPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))

	pipeline := tu.UniqueString("TestNonrootPipeline")
	_, err = c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo)},
				User:  "65534", // This is "nobody" on the default ubuntu:20.04 container, but it works.
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			OutputBranch: "",
			Update:       false,
			PodPatch: `[
				{"op": "add",  "path": "/securityContext",  "value": {}},
				{"op": "add",  "path": "/securityContext/runAsUser",  "value": 65534}
			]`,
			Autoscaling: false,
		},
	)
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	// The commitset should have a commit in: data, spec, pipeline, meta
	// the last two are dependent upon the first two, so should come later
	// in topological ordering
	require.Equal(t, 4, len(commitInfos))
	var commitRepos []*pfs.Repo
	for _, info := range commitInfos {
		commitRepos = append(commitRepos, info.Commit.Repo)
	}
	require.EqualOneOf(t, commitRepos[:2], client.NewRepo(pfs.DefaultProjectName, dataRepo))
	require.EqualOneOf(t, commitRepos[:2], client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.SpecRepoType))
	require.EqualOneOf(t, commitRepos[2:], client.NewRepo(pfs.DefaultProjectName, pipeline))
	require.EqualOneOf(t, commitRepos[2:], client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.MetaRepoType))

	var buf bytes.Buffer
	for _, info := range commitInfos {
		if proto.Equal(info.Commit.Repo, client.NewRepo(pfs.DefaultProjectName, pipeline)) {
			require.NoError(t, c.GetFile(info.Commit, "file", &buf))
			require.Equal(t, "foo", buf.String())
		}
	}
}

func TestRewindCrossPipeline(t *testing.T) {
	// TODO(2.0 optional): Duplicate file paths from different datums no longer allowed.
	t.Skip("Duplicate file paths from different datums no longer allowed.")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestRewindCrossPipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	create := func(update bool, repos ...string) error {
		var indivInputs []*pps.Input
		var stdin []string
		for _, r := range repos {
			indivInputs = append(indivInputs, client.NewPFSInput(pfs.DefaultProjectName, r, "/*"))
			stdin = append(stdin, fmt.Sprintf("cp /pfs/%s/* /pfs/out/", r))
		}
		var input *pps.Input
		if len(repos) > 1 {
			input = client.NewCrossInput(indivInputs...)
		} else {
			input = indivInputs[0]
		}
		_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
				Transform: &pps.Transform{
					Cmd:   []string{"bash"},
					Stdin: stdin,
				},
				Input:  input,
				Update: update,
			})
		return errors.EnsureStack(err)
	}
	require.NoError(t, create(false, dataRepo))

	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "first", strings.NewReader("first")))
	_, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)

	// make a new repo, and update the pipeline to take it as input
	laterRepo := tu.UniqueString("LaterRepo")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, laterRepo))
	require.NoError(t, create(true, dataRepo, laterRepo))

	// save the current commit set ID
	oldCommit, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)

	// add new files to both repos
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", ""), "second", strings.NewReader("second")))
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, laterRepo, "master", ""), "later", strings.NewReader("later")))
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)

	// now, move dataRepo back to the saved commit
	require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, dataRepo, "master", "master", oldCommit.Commit.Id, nil))
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	info, err := c.InspectCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, err)

	// the new commit cannot reuse the old ID
	require.NotEqual(t, info.Commit.Id, oldCommit.Commit.Id)

	// laterRepo has the same contents
	files, err := c.ListFileAll(client.NewCommit(pfs.DefaultProjectName, laterRepo, "master", info.Commit.Id), "/")
	require.NoError(t, err)
	require.Equal(t, 1, len(files))
	require.Equal(t, "/later", files[0].File.Path)

	// which is reflected in the output
	files, err = c.ListFileAll(client.NewCommit(pfs.DefaultProjectName, pipeline, "master", info.Commit.Id), "/")
	require.NoError(t, err)
	require.Equal(t, 2, len(files))
	require.ElementsEqualUnderFn(t, []string{"/first", "/later"}, files, func(f interface{}) interface{} { return f.(*pfs.FileInfo).File.Path })
}

func TestPipelineInputTriggerSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, dataRepo, "master", "", "", nil))
	require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, dataRepo, "toMove", "master", "", nil))

	// Create a pipeline taking in a trigger branch (to be created at pipeline creation), which is triggered by the toMove branch.
	pipeline := tu.UniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{"cp /pfs/trigger/* /pfs/out/"},
			},
			Input: client.NewPFSInputOpts("trigger", pfs.DefaultProjectName, dataRepo, "trigger", "/*", "", "", false, false, &pfs.Trigger{
				Branch:  "toMove",
				Commits: 1,
			}),
		})
	require.NoError(t, err)

	// Make sure the trigger triggered
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, dataRepo, "toMove", ""), "foo", strings.NewReader("bar")))
	_, err = c.WaitCommit(pfs.DefaultProjectName, dataRepo, "toMove", "")
	require.NoError(t, err)
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	files, err := c.ListFileAll(client.NewCommit(pfs.DefaultProjectName, dataRepo, "trigger", ""), "/")
	require.NoError(t, err)
	require.Equal(t, 1, len(files))
	require.Equal(t, "/foo", files[0].File.Path)
}

func TestPipelineAncestry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	base := basicPipelineReq(pipeline, dataRepo)
	base.Autoscaling = true

	// create three versions of the pipeline differing only in the transform user field
	for i := 1; i <= 3; i++ {
		base.Transform.User = fmt.Sprintf("user:%d", i)
		base.Update = i != 1
		_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), base)
		require.NoError(t, err)
	}

	for i := 1; i <= 3; i++ {
		info, err := c.InspectPipeline(pfs.DefaultProjectName, fmt.Sprintf("%s^%d", pipeline, 3-i), true)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("user:%d", i), info.Details.Transform.User)
	}

	infos, err := c.ListPipeline()
	require.NoError(t, err)
	require.Equal(t, 1, len(infos))
	require.Equal(t, pipeline, infos[0].Pipeline.Name)
	require.Equal(t, fmt.Sprintf("user:%d", 3), infos[0].Details.Transform.User)

	checkInfos := func(infos []*pps.PipelineInfo) {
		// make sure versions are sorted and have the correct details
		for i, info := range infos {
			require.Equal(t, 3-i, int(info.Version))
			require.Equal(t, fmt.Sprintf("user:%d", 3-i), info.Details.Transform.User)
		}
	}

	// get all pipelines
	infos, err = c.ListPipelineHistory(pfs.DefaultProjectName, pipeline, -1)
	require.NoError(t, err)
	require.Equal(t, 3, len(infos))
	checkInfos(infos)

	// get all pipelines by asking for too many
	infos, err = c.ListPipelineHistory(pfs.DefaultProjectName, pipeline, 3)
	require.NoError(t, err)
	require.Equal(t, 3, len(infos))
	checkInfos(infos)

	// get only the later two pipelines
	infos, err = c.ListPipelineHistory(pfs.DefaultProjectName, pipeline, 1)
	require.NoError(t, err)
	require.Equal(t, 2, len(infos))
	checkInfos(infos)
}

func TestDatumSetCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)
	c = c.WithDefaultTransformUser("1000")
	dataRepo := tu.UniqueString("TestDatumSetCache_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	masterCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, c.WithModifyFileClient(masterCommit, func(mfc client.ModifyFile) error {
		for i := 0; i < 90; i++ {
			require.NoError(t, mfc.PutFile(strconv.Itoa(i), strings.NewReader("")))
		}
		return nil
	}))
	pipeline := tu.UniqueString("TestDatumSetCache")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
					"sleep 1",
				},
			},
			Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			DatumSetSpec: &pps.DatumSetSpec{Number: 1},
		})
	require.NoError(t, err)
	ctx, cancel := pctx.WithCancel(context.Background())
	defer cancel()
	go func() {
		ticker := time.NewTimer(50 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				deleteEtcd(t, ctx, ns)
				err := waitForOnePodReady(t, ctx, ns, "app=etcd")
				if err != nil {
					t.Logf("etcd pod did not become ready after deletion.") // if the test ends, or it's sloe this is OK for the test. Just log it and keep testing.
				}
				select {
				case <-ctx.Done():
					return
				default:
					ticker = time.NewTimer(50 * time.Second)
				}
			}
		}
	}()
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
		_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
		return err
	})
	for i := 0; i < 5; i++ {
		_, err := c.InspectFile(commitInfo.Commit, strconv.Itoa(i))
		require.NoError(t, err)
	}
}

func monitorReplicas(t testing.TB, c *client.APIClient, namespace, pipeline string, n int) {
	kc := tu.GetKubeClient(t)
	rcName := ppsutil.PipelineRcName(&pps.PipelineInfo{
		Pipeline: &pps.Pipeline{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    pipeline,
		},
		Version: 1,
	})
	enoughReplicas := false
	tooManyReplicas := false
	var maxSeen int
	require.NoErrorWithinTRetryConstant(t, 4*time.Minute, func() error {
		scale, err := kc.CoreV1().ReplicationControllers(namespace).GetScale(context.Background(), rcName, metav1.GetOptions{})
		if err != nil {
			return errors.EnsureStack(err)
		}
		replicas := int(scale.Spec.Replicas)
		if replicas >= n {
			enoughReplicas = true
		}
		if replicas > n {
			if replicas > maxSeen {
				maxSeen = replicas
			}
			tooManyReplicas = true
		}
		ci, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
		require.NoError(t, err)
		if ci.Finished != nil {
			return nil
		}
		return errors.Errorf("Commit not yet finished %v", ci.Commit.Id)
	}, 2*time.Second)
	require.True(t, enoughReplicas, "didn't get enough replicas, looking for: %d", n)
	require.False(t, tooManyReplicas, "got too many replicas (%d), looking for: %d", maxSeen, n)
}

func TestLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	resp, err := c.PpsAPIClient.RunLoadTestDefault(c.Ctx(), &emptypb.Empty{})
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, cmdutil.Encoder("", buf).EncodeProto(resp))
	require.Equal(t, "", resp.Error, buf.String())
}

func TestPutFileNoErrorOnErroredParentCommit(t *testing.T) {
	// Test whether PutFile still works on a repo where a parent commit errored
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "inA"))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "inB"))
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		"A",
		"",
		[]string{"bash"},
		[]string{"exit -1"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(pfs.DefaultProjectName, "inA", "/*"),
		"",
		false,
	))
	require.NoError(t, c.CreatePipeline(pfs.DefaultProjectName,
		"B",
		"",
		[]string{"bash"},
		[]string{"sleep 3"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewUnionInput(
			client.NewPFSInput(pfs.DefaultProjectName, "inB", "/*"),
			client.NewPFSInput(pfs.DefaultProjectName, "A", "/*"),
		),
		"",
		false,
	))
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, "inA", "master", ""), "file", strings.NewReader("foo")))
	commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, "A", "master", "")
	require.NoError(t, err)
	require.True(t, strings.Contains(commitInfo.Error, "failed"))
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, "inB", "master", ""), "file", strings.NewReader("foo")))
}

func TestTemporaryDuplicatedPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	repo := tu.UniqueString("repo")
	other := tu.UniqueString("other")
	pipeline := tu.UniqueString("pipeline")

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, other))

	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, other, "master", ""), "empty",
		strings.NewReader("")))

	// add an output file bigger than the sharding threshold so that two in a row
	// will fall on either side of a naive shard division
	bigSize := index.DefaultShardSizeThreshold * 5 / 4
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "a",
		strings.NewReader(strconv.Itoa(bigSize/units.MB))))

	// add some more files to push the fileset that deletes the old datums out of the first fan-in
	for i := 0; i < 10; i++ {
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), fmt.Sprintf("b-%d", i),
			strings.NewReader("1")))
	}

	req := basicPipelineReq(pipeline, repo)
	req.DatumSetSpec = &pps.DatumSetSpec{
		Number: 1,
	}
	req.Transform.Stdin = []string{
		fmt.Sprintf("for f in /pfs/%s/* ; do", repo),
		"dd if=/dev/zero of=/pfs/out/$(basename $f) bs=1M count=$(cat $f)",
		"done",
	}
	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), req)
	require.NoError(t, err)

	commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	require.Equal(t, "", commitInfo.Error)

	// update the pipeline to take new input, but produce identical output
	// the filesets in the output of the resulting job should look roughly like
	// [old output] [new output] [deletion of old output]
	// If a path is included in multiple shards, it would appear twice for both
	// the old and new datums, while only one copy would be deleted,
	// leading to either a validation error or a repeated path/datum pair
	req.Update = true
	req.Input = client.NewCrossInput(
		client.NewPFSInput(pfs.DefaultProjectName, repo, "/*"),
		client.NewPFSInput(pfs.DefaultProjectName, other, "/*"))
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), req)
	require.NoError(t, err)

	commitInfo, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	require.Equal(t, "", commitInfo.Error)
}

func TestValidationFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	repo := tu.UniqueString(t.Name())
	pipeline := tu.UniqueString("pipeline-" + t.Name())

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))

	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "foo", strings.NewReader("baz")))
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "bar", strings.NewReader("baz")))

	req := basicPipelineReq(pipeline, repo)
	req.Transform.Stdin = []string{
		fmt.Sprintf("cat /pfs/%s/* > /pfs/out/overlap", repo), // write datums to the same path
	}
	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), req)
	require.NoError(t, err)

	commitInfo, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	require.NotEqual(t, "", commitInfo.Error)

	// update the pipeline to fix things
	req.Update = true
	req.Transform.Stdin = []string{
		fmt.Sprintf("cp /pfs/%s/* /pfs/out", repo),
	}

	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), req)
	require.NoError(t, err)

	commitInfo, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	require.Equal(t, "", commitInfo.Error)
}

func TestMissingSecretFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)
	kc := tu.GetKubeClient(t).CoreV1().Secrets(ns)

	secret := tu.UniqueString(strings.ToLower(t.Name() + "-secret"))
	repo := tu.UniqueString(t.Name() + "-data")
	pipeline := tu.UniqueString(t.Name())
	_, err := kc.Create(c.Ctx(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secret},
		StringData: map[string]string{"foo": "bar"}}, metav1.CreateOptions{})
	require.NoError(t, err)

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	req := basicPipelineReq(pipeline, repo)
	req.Transform.Secrets = append(req.Transform.Secrets, &pps.SecretMount{
		Name:   secret,
		Key:    "foo",
		EnvVar: "MY_SECRET_ENV_VAR",
	})
	req.Autoscaling = true
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), req)
	require.NoError(t, err)

	// wait for system to go into standby
	require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
		_, err = c.WaitCommit(pfs.DefaultProjectName, repo, "master", "")
		if err != nil {
			return err
		}
		info, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		if err != nil {
			return err
		}
		if info.State != pps.PipelineState_PIPELINE_STANDBY {
			return errors.Errorf("pipeline in %s instead of standby", info.State)
		}
		return nil
	})

	require.NoError(t, kc.Delete(c.Ctx(), secret, metav1.DeleteOptions{}))
	// force pipeline to come back up, and check for failure
	require.NoError(t, c.PutFile(
		client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "foo", strings.NewReader("bar")))
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		info, err := c.InspectPipeline(pfs.DefaultProjectName, pipeline, false)
		if err != nil {
			return err
		}
		if info.State != pps.PipelineState_PIPELINE_CRASHING {
			return errors.Errorf("pipeline in %s instead of crashing", info.State)
		}
		return nil
	})
}
func TestDatabaseStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	type rowCountResults struct {
		NLiveTup   int    `json:"n_live_tup"`
		RelName    string `json:"relname"`
		SchemaName string `json:"schemaname"`
	}

	type commitResults struct {
		Key string `json:"key"`
	}

	repoName := tu.UniqueString("TestDatabaseStats-repo")
	branchName := "master"
	numCommits := 100
	buf := &bytes.Buffer{}
	filter := &debug.Filter{
		Filter: &debug.Filter_Database{Database: true},
	}

	c, _ := minikubetestenv.AcquireCluster(t)

	createTestCommits(t, repoName, branchName, numCommits, c)
	time.Sleep(5 * time.Second) // give some time for the stats collector to run.
	require.NoError(t, c.Dump(filter, 100, buf), "dumping database files should succeed")
	gr, err := gzip.NewReader(buf)
	require.NoError(t, err)

	var rows []commitResults
	foundCommitsJSON, foundRowCounts := false, false
	require.NoError(t, tarutil.Iterate(gr, func(f tarutil.File) error {
		fileContents := &bytes.Buffer{}
		if err := f.Content(fileContents); err != nil {
			return errors.EnsureStack(err)
		}

		hdr, err := f.Header()
		require.NoError(t, err, "getting database tar file header should succeed")
		require.NotMatch(t, "^[a-zA-Z0-9_\\-\\/]+\\.error$", hdr.Name)
		switch hdr.Name {
		case "database/row-counts.json":
			var rows []rowCountResults
			require.NoError(t, json.Unmarshal(fileContents.Bytes(), &rows),
				"unmarshalling row-counts.json should succeed")

			for _, row := range rows {
				if row.RelName == "commits" && row.SchemaName == "pfs" {
					require.NotEqual(t, 0, row.NLiveTup,
						"some commits from createTestCommits should be accounted for")
					foundRowCounts = true
				}
			}
		case "database/tables/pfs/commits.json":
			require.NoError(t, json.Unmarshal(fileContents.Bytes(), &rows),
				"unmarshalling commits.json should succeed")
			require.Equal(t, numCommits, len(rows), "number of commits should match number of rows")
			foundCommitsJSON = true
		}

		return nil
	}))

	require.Equal(t, true, foundRowCounts,
		"we should have an entry in row-counts.json for commits")
	require.Equal(t, true, foundCommitsJSON,
		"checks for commits.json should succeed")
}

func TestSimplePipelineNonRoot(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestSimplePipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))
	pipeline := tu.UniqueString("TestSimplePipeline")
	req := basicPipelineReq(pipeline, dataRepo)
	req.Transform.User = "65534"
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), req)
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	// The commitset should have a commit in: data, spec, pipeline, meta
	// the last two are dependent upon the first two, so should come later
	// in topological ordering
	require.Equal(t, 4, len(commitInfos))
	var commitRepos []*pfs.Repo
	for _, info := range commitInfos {
		commitRepos = append(commitRepos, info.Commit.Repo)
	}
	require.EqualOneOf(t, commitRepos[:2], client.NewRepo(pfs.DefaultProjectName, dataRepo))
	require.EqualOneOf(t, commitRepos[:2], client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.SpecRepoType))
	require.EqualOneOf(t, commitRepos[2:], client.NewRepo(pfs.DefaultProjectName, pipeline))
	require.EqualOneOf(t, commitRepos[2:], client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.MetaRepoType))

	var buf bytes.Buffer
	for _, info := range commitInfos {
		if proto.Equal(info.Commit.Repo, client.NewRepo(pfs.DefaultProjectName, pipeline)) {
			require.NoError(t, c.GetFile(info.Commit, "file", &buf))
			require.Equal(t, "foo", buf.String())
		}
	}
}

func TestSimplePipelinePodPatchNonRoot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	dataRepo := tu.UniqueString("TestSimplePipeline_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))

	commit1, err := c.StartCommit(pfs.DefaultProjectName, dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, dataRepo, "", commit1.Id))
	pipeline := tu.UniqueString("TestSimplePipeline")
	req := basicPipelineReq(pipeline, dataRepo)

	req.PodPatch = "[{\"op\":\"add\",\"path\":\"/securityContext\",\"value\":{}},{\"op\":\"add\",\"path\":\"/securityContext\",\"value\":{\"runAsGroup\":1000,\"runAsUser\":1000,\"fsGroup\":1000,\"runAsNonRoot\":true,\"seccompProfile\":{\"type\":\"RuntimeDefault\"}}},{\"op\":\"add\",\"path\":\"/containers/0/securityContext\",\"value\":{}},{\"op\":\"add\",\"path\":\"/containers/0/securityContext\",\"value\":{\"runAsGroup\":1000,\"runAsUser\":1000,\"allowPrivilegeEscalation\":false,\"capabilities\":{\"drop\":[\"all\"]},\"readOnlyRootFilesystem\":true}}]"
	_, err = c.PpsAPIClient.CreatePipeline(c.Ctx(), req)
	require.NoError(t, err)

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	// The commitset should have a commit in: data, spec, pipeline, meta
	// the last two are dependent upon the first two, so should come later
	// in topological ordering
	require.Equal(t, 4, len(commitInfos))
	var commitRepos []*pfs.Repo
	for _, info := range commitInfos {
		commitRepos = append(commitRepos, info.Commit.Repo)
	}
	require.EqualOneOf(t, commitRepos[:2], client.NewRepo(pfs.DefaultProjectName, dataRepo))
	require.EqualOneOf(t, commitRepos[:2], client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.SpecRepoType))
	require.EqualOneOf(t, commitRepos[2:], client.NewRepo(pfs.DefaultProjectName, pipeline))
	require.EqualOneOf(t, commitRepos[2:], client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.MetaRepoType))

	var buf bytes.Buffer
	for _, info := range commitInfos {
		if proto.Equal(info.Commit.Repo, client.NewRepo(pfs.DefaultProjectName, pipeline)) {
			require.NoError(t, c.GetFile(info.Commit, "file", &buf))
			require.Equal(t, "foo", buf.String())
		}
	}
}

func TestZombieCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	repo := tu.UniqueString(t.Name() + "-data")
	pipeline := tu.UniqueString(t.Name())
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))

	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""),
		"foo", strings.NewReader("baz")))

	req := basicPipelineReq(pipeline, repo)
	_, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), req)
	require.NoError(t, err)
	require.NoErrorWithinT(t, 30*time.Second, func() error {
		_, err := c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
		return err
	})
	// fsck should succeed (including zombie check, on by default)
	require.NoError(t, c.Fsck(false, func(response *pfs.FsckResponse) error {
		return errors.Errorf("got fsck error: %s", response.Error)
	}, client.WithZombieCheckAll()))

	// stop pipeline so we can modify
	require.NoError(t, c.StopPipeline(pfs.DefaultProjectName, pipeline))
	// create new commits on output and meta
	_, err = c.ExecuteInTransaction(func(c *client.APIClient) error {
		if _, err := c.StartCommit(pfs.DefaultProjectName, pipeline, "master"); err != nil {
			return errors.EnsureStack(err)
		}
		metaBranch := client.NewSystemRepo(pfs.DefaultProjectName, pipeline, pfs.MetaRepoType).NewBranch("master")
		if _, err := c.PfsAPIClient.StartCommit(c.Ctx(), &pfs.StartCommitRequest{Branch: metaBranch}); err != nil {
			return errors.EnsureStack(err)
		}
		_, err = c.PfsAPIClient.FinishCommit(c.Ctx(), &pfs.FinishCommitRequest{Commit: metaBranch.NewCommit("")})
		return errors.EnsureStack(err)
	})
	require.NoError(t, err)
	// add a file to the output with a datum that doesn't exist
	require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, pipeline, "master", ""),
		"zombie", strings.NewReader("zombie"), client.WithDatumPutFile("zombie")))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, pipeline, "master", ""))
	_, err = c.WaitCommit(pfs.DefaultProjectName, pipeline, "master", "")
	require.NoError(t, err)
	var messages []string
	// fsck should notice the zombie file
	require.NoError(t, c.Fsck(false, func(response *pfs.FsckResponse) error {
		messages = append(messages, response.Error)
		return nil
	}, client.WithZombieCheckAll()))
	require.Equal(t, 1, len(messages))
	require.Matches(t, "zombie", messages[0])
}

func TestJobPropagationOnlyOutputBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	c = c.WithDefaultTransformUser("1000")

	project := tu.UniqueString("project")
	require.NoError(t, c.CreateProject(project))
	dataRepo := tu.UniqueString("PropagationOnlyOutputBranch_data")
	require.NoError(t, c.CreateRepo(project, dataRepo))

	pipeline := tu.UniqueString("pipeline")
	outputBranch := client.NewBranch(project, pipeline, "output")
	require.NoError(t, c.CreatePipeline(project,
		pipeline,
		tu.DefaultTransformImage,
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(project, dataRepo, "/*"),
		outputBranch.Name,
		false,
	))

	require.NoError(t, c.CreateBranch(project, pipeline, "test", "", "", nil))

	commit := client.NewCommit(project, dataRepo, "master", "")
	for i := 0; i < 3; i++ {
		require.NoError(t, c.PutFile(commit, "file", strings.NewReader("")))
	}

	jobInfos, err := c.ListJob(project, pipeline, nil, -1, false)
	require.NoError(t, err)
	require.Equal(t, 4, len(jobInfos))
	for _, jobInfo := range jobInfos {
		require.Equal(t, outputBranch, jobInfo.OutputCommit.Branch)
	}
}

func TestDatumBatching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	c = c.WithDefaultTransformUser("1000")

	dataRepo := tu.UniqueString("DatumBatching_data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	numFiles := 15
	require.NoError(t, c.WithModifyFileClient(dataCommit, func(mfc client.ModifyFile) error {
		for i := 0; i < numFiles; i++ {
			require.NoError(t, mfc.PutFile(fmt.Sprintf("/file-%02d", i), strings.NewReader("")))
		}
		return nil
	}))

	createPipelineRequest := func(pipeline, script string) *pps.CreatePipelineRequest {
		return &pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd:           []string{"bash"},
				Stdin:         []string{script},
				DatumBatching: true,
			},
			Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			DatumSetSpec: &pps.DatumSetSpec{Number: 5},
		}
	}

	checkSuccess := func(request *pps.CreatePipelineRequest) {
		_, err := c.PpsAPIClient.CreatePipeline(context.Background(), request)
		require.NoError(t, err)
		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, request.Pipeline.Name, "master", "")
		require.NoError(t, err)
		jobInfo, err := c.WaitJob(pfs.DefaultProjectName, request.Pipeline.Name, commitInfo.Commit.Id, false)
		require.NoError(t, err)
		require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
		fileInfos, err := c.ListFileAll(jobInfo.OutputCommit, "")
		require.NoError(t, err)
		require.Equal(t, numFiles, len(fileInfos))
		for i := 0; i < numFiles; i++ {
			require.Equal(t, fmt.Sprintf("/file-%02d", i), fileInfos[i].File.Path)
		}
	}
	t.Run("Basic", func(t *testing.T) {
		script := fmt.Sprintf(`
			while true
			do
				pachctl next datum
				cp /pfs/%s/* /pfs/out/
			done
			`, dataRepo)
		pipeline := tu.UniqueString("DatumBatchingBasic")
		request := createPipelineRequest(pipeline, script)
		checkSuccess(request)
	})
	t.Run("Error", func(t *testing.T) {
		script := fmt.Sprintf(`
			while true
			do
				pachctl next datum
				if [ ! -f /tmp/exec ]
				then
					touch /tmp/exec
					pachctl next datum --error oops
				fi
				cp /pfs/%s/* /pfs/out/
			done
			`, dataRepo)
		pipeline := tu.UniqueString("DatumBatchingError")
		request := createPipelineRequest(pipeline, script)
		checkSuccess(request)
	})
	t.Run("Exit", func(t *testing.T) {
		script := fmt.Sprintf(`
			while true
			do
				pachctl next datum
				if [ ! -f /tmp/exec ]
				then
					touch /tmp/exec
					exit 1
				fi
				cp /pfs/%s/* /pfs/out/
			done
			`, dataRepo)
		pipeline := tu.UniqueString("DatumBatchingExit")
		request := createPipelineRequest(pipeline, script)
		checkSuccess(request)
	})
	t.Run("DatumTimeout", func(t *testing.T) {
		script := fmt.Sprintf(`
			while true
			do
				pachctl next datum
				if [ ! -f /tmp/exec ]
				then
					touch /tmp/exec
					sleep 5
				fi
				cp /pfs/%s/* /pfs/out/
			done
			`, dataRepo)
		pipeline := tu.UniqueString("DatumBatchingDatumTimeout")
		request := createPipelineRequest(pipeline, script)
		request.DatumTimeout = durationpb.New(3 * time.Second)
		checkSuccess(request)
	})

	checkState := func(request *pps.CreatePipelineRequest, state pps.JobState) {
		_, err := c.PpsAPIClient.CreatePipeline(context.Background(), request)
		require.NoError(t, err)
		commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, request.Pipeline.Name, "master", "")
		require.NoError(t, err)
		jobInfo, err := c.WaitJob(pfs.DefaultProjectName, request.Pipeline.Name, commitInfo.Commit.Id, false)
		require.NoError(t, err)
		require.Equal(t, state, jobInfo.State)
	}
	t.Run("JobFailure", func(t *testing.T) {
		script := `
			pachctl next datum
			exit 1
			`
		pipeline := tu.UniqueString("DatumBatchingJobFailure")
		request := createPipelineRequest(pipeline, script)
		checkState(request, pps.JobState_JOB_FAILURE)
	})
	t.Run("JobTimeout", func(t *testing.T) {
		script := `sleep 3600`
		pipeline := tu.UniqueString("DatumBatchingJobTimeout")
		request := createPipelineRequest(pipeline, script)
		request.JobTimeout = durationpb.New(10 * time.Second)
		checkState(request, pps.JobState_JOB_KILLED)
	})
	t.Run("PrematureExit", func(t *testing.T) {
		script := `
			pachctl next datum
			exit 0
			`
		pipeline := tu.UniqueString("DatumBatchingPrematureExit")
		request := createPipelineRequest(pipeline, script)
		checkState(request, pps.JobState_JOB_FAILURE)
	})
	t.Run("Env", func(t *testing.T) {
		// This test will error if the datum id environment variable isn't set and distinct for each datum.
		script := fmt.Sprintf(`
			set -a
			while true
			do
				pachctl next datum
				source /pfs/.env
				if [ -z $%s ]
				then
					exit 1
				fi
				touch /pfs/out/$%s
			done
			`, client.DatumIDEnv, client.DatumIDEnv)
		pipeline := tu.UniqueString("DatumBatchingEnv")
		request := createPipelineRequest(pipeline, script)
		checkState(request, pps.JobState_JOB_SUCCESS)
	})
}

func TestJQFilterInfiniteLoop(t *testing.T) {
	var ctx = pctx.Background("TestJQFilterDenialOfService")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	c = c.WithDefaultTransformUser("1000")

	dataRepo := tu.UniqueString("data")
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, dataRepo))
	dataCommit := client.NewCommit(pfs.DefaultProjectName, dataRepo, "master", "")
	require.NoError(t, c.WithModifyFileClient(dataCommit, func(mfc client.ModifyFile) error {
		require.NoError(t, mfc.PutFile("/test-file", strings.NewReader("")))
		return nil
	}))

	createPipelineRequest := func(pipeline, script string) *pps.CreatePipelineRequest {
		return &pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipeline),
			Transform: &pps.Transform{
				Cmd:           []string{"bash"},
				Stdin:         []string{script},
				DatumBatching: true,
			},
			Input:        client.NewPFSInput(pfs.DefaultProjectName, dataRepo, "/*"),
			DatumSetSpec: &pps.DatumSetSpec{Number: 5},
		}
	}

	script := fmt.Sprintf(`
			while true
			do
				pachctl next datum
				cp /pfs/%s/* /pfs/out/
			done
			`, dataRepo)
	pipeline := tu.UniqueString("pipeline")
	request := createPipelineRequest(pipeline, script)
	filter := `{ source: ., output: "" } | until(.source == "quux"; {"foo": "bar"}) | .output`

	_, err := c.PpsAPIClient.CreatePipeline(ctx, request)
	require.NoError(t, err, "could not create pipeline")

	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, request.Pipeline.Name, "master", "")
	require.NoError(t, err, "could not inspect commit")

	jobInfo, err := c.WaitJob(pfs.DefaultProjectName, request.Pipeline.Name, commitInfo.Commit.Id, false)
	require.NoError(t, err, "could not wait for job")
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State, "job not successful")

	t.Run("ListPipeline", func(t *testing.T) {
		resp, err := c.PpsAPIClient.ListPipeline(ctx, &pps.ListPipelineRequest{
			JqFilter: filter,
		})
		require.NoError(t, err, "could not list pipelines")
		for {
			_, err := resp.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if status.Code(err) == codes.DeadlineExceeded {
				// deadline exceeded means that the server handled it
				break
			}
			require.NoError(t, err, "list pipelines fails")
		}
	})

	t.Run("ListJob", func(t *testing.T) {
		resp, err := c.PpsAPIClient.ListJob(ctx, &pps.ListJobRequest{
			JqFilter: filter,
		})
		if status.Code(err) == codes.DeadlineExceeded {
			return
		}
		require.NoError(t, err, "could not list jobs")
		for {
			_, err := resp.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if status.Code(err) == codes.DeadlineExceeded {
				// deadline exceeded means that the server handled it
				break
			}
			require.NoError(t, err, "list jobs fails")
		}
	})
}

func TestPipelinesSummary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	repo := "input"
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	type PipState int
	const (
		Active PipState = iota
		Unhealthy
		Paused
		Failed
	)
	createPipeline := func(project, name string, state PipState) {
		cmd := []string{"cp", "-r", "/pfs/in", "/pfs/out"}
		switch state {
		case Unhealthy:
			cmd = []string{"exit 1"}
		case Failed:
			cmd = nil
		}
		require.NoError(t, c.CreatePipeline(
			project,
			name,
			"", /* default image*/
			cmd,
			nil, /* stdin */
			nil, /* spec */
			&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: repo, Glob: "/*", Name: "in"}},
			"",    /* output */
			false, /* update */
		))
		if state == Paused {
			require.NoError(t, c.StopPipeline(project, name))
		}
		_, err := c.WaitCommit(project, name, "master", "")
		require.NoError(t, err)
	}
	projects := []string{"a", "b"}
	for _, prj := range projects {
		require.NoError(t, c.CreateProject(prj))
	}
	pips := []string{"A", "B", "C", "D", "E"}
	for _, prj := range projects {
		for i, pip := range pips {
			var state PipState = Active
			if i == 0 {
				state = Unhealthy
			} else if i == 1 {
				state = Paused
			} else if i == 2 {
				state = Failed
			}
			createPipeline(prj, pip, state)
		}
	}
	commit, err := c.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFile(commit, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
	require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, repo, "master", ""))
	for _, tc := range [][]string{
		{},
		{"a"},
		{"b"},
		{"b", "a"},
	} {
		var pickers []*pfs.ProjectPicker
		for _, project := range tc {
			pickers = append(pickers, &pfs.ProjectPicker{
				Picker: &pfs.ProjectPicker_Name{Name: project},
			})
		}
		resp, err := c.PipelinesSummary(pctx.TestContext(t),
			&pps.PipelinesSummaryRequest{Projects: pickers})
		require.NoError(t, err)
		expectedProjects := tc
		if len(tc) == 0 {
			expectedProjects = projects
		}
		require.Len(t, resp.Summaries, len(expectedProjects))
		for i, project := range expectedProjects {
			require.Equal(t, project, resp.Summaries[i].Project.Name)
			require.Equal(t, int64(3), resp.Summaries[i].ActivePipelines) // unhealthy pipelines are still active
			require.Equal(t, int64(1), resp.Summaries[i].PausedPipelines)
			require.Equal(t, int64(1), resp.Summaries[i].FailedPipelines)
			require.Equal(t, int64(1), resp.Summaries[i].UnhealthyPipelines)
		}
	}
}
