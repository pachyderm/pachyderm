package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfspretty "github.com/pachyderm/pachyderm/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/pretty"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
	ppspretty "github.com/pachyderm/pachyderm/src/server/pps/pretty"
	"github.com/pachyderm/pachyderm/src/server/pps/server/githook"

	"github.com/gogo/protobuf/types"
	apps "k8s.io/api/apps/v1beta2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// If this environment variable is set, then the tests are being run
	// in a real cluster in the cloud.
	InCloudEnv = "PACH_TEST_CLOUD"
)

func TestSimplePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestSimplePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("TestSimplePipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo", buf.String())
}

func TestPipelineWithParallelism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineWithParallelism_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 10000
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 4,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, fmt.Sprintf("file-%d", i), 0, 0, &buf))
		require.Equal(t, fmt.Sprintf("%d", i), buf.String())
	}
}

func TestPipelineWithLargeFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineInputDataModification_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	r := rand.New(rand.NewSource(99))
	numFiles := 10
	var fileContents []string

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		fileContent := workload.RandString(r, int(pfs.ChunkSize)+i*MB)
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i),
			strings.NewReader(fileContent))
		require.NoError(t, err)
		fileContents = append(fileContents, fileContent)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		nil,
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	commit := commitInfos[0].Commit

	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		fileName := fmt.Sprintf("file-%d", i)

		fileInfo, err := c.InspectFile(commit.Repo.Name, commit.ID, fileName)
		require.NoError(t, err)
		require.Equal(t, int(pfs.ChunkSize)+i*MB, int(fileInfo.SizeBytes))

		require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, fileName, 0, 0, &buf))
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

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestDatumDedup_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	// This pipeline sleeps for 10 secs per datum
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"sleep 10",
		},
		nil,
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))

	// Since we did not change the datum, the datum should not be processed
	// again, which means that the job should complete instantly.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	stream, err := c.PfsAPIClient.FlushCommit(
		ctx,
		&pfs.FlushCommitRequest{
			Commits: []*pfs.Commit{commit2},
		})
	require.NoError(t, err)
	_, err = stream.Recv()
	require.NoError(t, err)
}

func TestPipelineInputDataModification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineInputDataModification_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		nil,
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo", buf.String())

	// replace the contents of 'file' in dataRepo (from "foo" to "bar")
	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.DeleteFile(dataRepo, commit2.ID, "file"))
	_, err = c.PutFile(dataRepo, commit2.ID, "file", strings.NewReader("bar"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commit2}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	buf.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "bar", buf.String())

	// Add a file to dataRepo
	commit3, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.DeleteFile(dataRepo, commit3.ID, "file"))
	_, err = c.PutFile(dataRepo, commit3.ID, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit3.ID))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commit3}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	require.YesError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	buf.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file2", 0, 0, &buf))
	require.Equal(t, "foo", buf.String())

	commitInfos, err = c.ListCommit(pipeline, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
}

func TestMultipleInputsFromTheSameBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestMultipleInputsFromTheSameBranch_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "dirA/file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "dirB/file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"cat /pfs/out/file",
			fmt.Sprintf("cat /pfs/dirA/dirA/file >> /pfs/out/file"),
			fmt.Sprintf("cat /pfs/dirB/dirB/file >> /pfs/out/file"),
		},
		nil,
		client.NewCrossInput(
			client.NewAtomInputOpts("dirA", dataRepo, "", "/dirA/*", false),
			client.NewAtomInputOpts("dirB", dataRepo, "", "/dirB/*", false),
		),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\nfoo\n", buf.String())

	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit2.ID, "dirA/file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commit2}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	buf.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\nbar\nfoo\n", buf.String())

	commit3, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit3.ID, "dirB/file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit3.ID))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commit3}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	buf.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\nbar\nfoo\nbuzz\n", buf.String())

	commitInfos, err = c.ListCommit(pipeline, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
}

func TestMultipleInputsFromTheSameRepoDifferentBranches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestMultipleInputsFromTheSameRepoDifferentBranches_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	branchA := "branchA"
	branchB := "branchB"

	pipeline := uniqueString("pipeline")
	// Creating this pipeline should error, because the two inputs are
	// from the same repo but they don't specify different names.
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"cat /pfs/branch-a/file >> /pfs/out/file",
			"cat /pfs/branch-b/file >> /pfs/out/file",
		},
		nil,
		client.NewCrossInput(
			client.NewAtomInputOpts("branch-a", dataRepo, branchA, "/*", false),
			client.NewAtomInputOpts("branch-b", dataRepo, branchB, "/*", false),
		),
		"",
		false,
	))

	commitA, err := c.StartCommit(dataRepo, branchA)
	require.NoError(t, err)
	c.PutFile(dataRepo, commitA.ID, "/file", strings.NewReader("data A\n"))
	c.FinishCommit(dataRepo, commitA.ID)

	commitB, err := c.StartCommit(dataRepo, branchB)
	require.NoError(t, err)
	c.PutFile(dataRepo, commitB.ID, "/file", strings.NewReader("data B\n"))
	c.FinishCommit(dataRepo, commitB.ID)

	iter, err := c.FlushCommit([]*pfs.Commit{commitA, commitB}, nil)
	require.NoError(t, err)
	commits := collectCommitInfos(t, iter)
	require.Equal(t, 1, len(commits))
	buffer := bytes.Buffer{}
	require.NoError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "data A\ndata B\n", buffer.String())
}

func TestMultipleInputsFromTheSameRepoDifferentBranchesIncremental(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestMultipleInputsFromTheSameRepoDifferentBranchesIncremental_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	branchA := "branchA"
	branchB := "branchB"

	pipeline := uniqueString("pipeline")
	// Creating this pipeline should error, because the two inputs are
	// from the same repo but they don't specify different names.
	req := &pps.CreatePipelineRequest{
		Pipeline: &pps.Pipeline{Name: pipeline},
		Transform: &pps.Transform{
			Cmd: []string{"bash"},
			Stdin: []string{
				"ls /pfs/out/file-a && echo true >> /pfs/out/prev-a",
				"ls /pfs/out/file-b && echo true >> /pfs/out/prev-b",
				"ls /pfs/branch-a/file && echo true >> /pfs/out/file-a",
				"ls /pfs/branch-b/file && echo true >> /pfs/out/file-b",
			},
		},
		Input: client.NewCrossInput(
			client.NewAtomInputOpts("branch-a", dataRepo, branchA, "/*", false),
			client.NewAtomInputOpts("branch-b", dataRepo, branchB, "/*", false),
		),
		Incremental: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := c.PpsAPIClient.CreatePipeline(ctx, req)
	require.NoError(t, err)

	// Make four commits: branchA, branchB, branchA, branchB. We should see
	// 'prev-a' after the third commit, and 'prev-b' after the fourth
	commit, err := c.StartCommit(dataRepo, branchA)
	require.NoError(t, err)
	c.PutFile(dataRepo, commit.ID, "/file", strings.NewReader("data A\n"))
	c.FinishCommit(dataRepo, commit.ID)

	commit, err = c.StartCommit(dataRepo, branchB)
	require.NoError(t, err)
	c.PutFile(dataRepo, commit.ID, "/file", strings.NewReader("data B\n"))
	c.FinishCommit(dataRepo, commit.ID)
	iter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commits := collectCommitInfos(t, iter)
	require.Equal(t, 1, len(commits))
	buffer := bytes.Buffer{}
	require.YesError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.ID, "prev-a", 0, 0, &buffer))
	buffer.Reset()
	require.YesError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.ID, "prev-b", 0, 0, &buffer))

	commit, err = c.StartCommit(dataRepo, branchA)
	require.NoError(t, err)
	c.PutFile(dataRepo, commit.ID, "/file", strings.NewReader("data A\n"))
	c.FinishCommit(dataRepo, commit.ID)
	iter, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commits = collectCommitInfos(t, iter)
	require.Equal(t, 1, len(commits))
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.ID, "prev-a", 0, 0, &buffer))
	buffer.Reset()
	require.NoError(t, c.GetFile(commits[0].Commit.Repo.Name, commits[0].Commit.ID, "prev-b", 0, 0, &buffer))
}

func TestPipelineFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineFailure_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"exit 1"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))
	var jobInfos []*pps.JobInfo
	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err = c.ListJob(pipeline, nil, nil)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return fmt.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		return nil
	}, backoff.NewTestingBackOff()))
	jobInfo, err := c.PpsAPIClient.InspectJob(context.Background(), &pps.InspectJobRequest{
		Job:        jobInfos[0].Job,
		BlockState: true,
	})
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)
	require.True(t, strings.Contains(jobInfo.Reason, "datum"))
}

func TestEgressFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestEgressFailure_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	// This pipeline should fail because the egress URL is invalid
	pipeline := uniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			Input:  client.NewAtomInput(dataRepo, "/"),
			Egress: &pps.Egress{"invalid://blahblah"},
		})
	require.NoError(t, err)

	var jobInfos []*pps.JobInfo
	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err = c.ListJob(pipeline, nil, nil)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return fmt.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		return nil
	}, backoff.NewTestingBackOff()))
	jobInfo, err := c.PpsAPIClient.InspectJob(context.Background(), &pps.InspectJobRequest{
		Job:        jobInfos[0].Job,
		BlockState: true,
	})
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)
	require.True(t, strings.Contains(jobInfo.Reason, "egress"))
}

func TestLazyPipelinePropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	dataRepo := uniqueString("TestPipeline_datax")
	require.NoError(t, c.CreateRepo(dataRepo))
	pipelineA := uniqueString("pipelineA")
	require.NoError(t, c.CreatePipeline(
		pipelineA,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInputOpts("", dataRepo, "", "/*", true),
		"",
		false,
	))
	pipelineB := uniqueString("pipelineB")
	require.NoError(t, c.CreatePipeline(
		pipelineB,
		"",
		[]string{"cp", path.Join("/pfs", pipelineA, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInputOpts("", pipelineA, "", "/*", true),
		"",
		false,
	))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit1.ID)}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, commitIter)

	jobInfos, err := c.ListJob(pipelineA, nil, nil)
	require.NoError(t, err)
	// Two jobs -- one from creating the pipeline, one from the commit
	require.Equal(t, 2, len(jobInfos))
	require.NotNil(t, jobInfos[0].Input.Atom)
	require.Equal(t, true, jobInfos[0].Input.Atom.Lazy)
	jobInfos, err = c.ListJob(pipelineB, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
	require.NotNil(t, jobInfos[0].Input.Atom)
	require.Equal(t, true, jobInfos[0].Input.Atom.Lazy)
}

func TestLazyPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestLazyPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: &pps.Input{
				Atom: &pps.AtomInput{
					Repo: dataRepo,
					Glob: "/",
					Lazy: true,
				},
			},
		})
	require.NoError(t, err)
	// Do a commit
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, "master", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	// We put 2 files, 1 of which will never be touched by the pipeline code.
	// This is an important part of the correctness of this test because the
	// job-shim sets up a goro for each pipe, pipes that are never opened will
	// leak but that shouldn't prevent the job from completing.
	_, err = c.PutFile(dataRepo, "master", "file2", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, "master"))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	buffer := bytes.Buffer{}
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

// There's an issue where if you use cp with certain flags, it might copy
// special files without reading from them.  In our case, we use named pipes
// to simulate lazy files, so the pipes themselves might get copied into
// the output directory, blocking upload.
//
// We've updated the code such that we are able to detect if the files we
// are uploading are pipes, and make the job fail in that case.
func TestLazyPipelineCPPipes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestLazyPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipeline := uniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				// Using cp with the -r flag apparently just copes go
				Cmd: []string{"cp", "-r", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: &pps.Input{
				Atom: &pps.AtomInput{
					Repo: dataRepo,
					Glob: "/",
					Lazy: true,
				},
			},
		})
	require.NoError(t, err)
	// Do a commit
	_, err = c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, "master", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, "master"))

	// wait for job to spawn
	time.Sleep(15 * time.Second)
	var jobID string
	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob(pipeline, nil, nil)
		if err != nil {
			return err
		}
		if len(jobInfos) != 2 {
			return fmt.Errorf("len(jobInfos) should be 2 (an empty job from creating the pipeline and a real job from committing to the input repo)")
		}
		jobID = jobInfos[0].Job.ID
		jobInfo, err := c.PpsAPIClient.InspectJob(context.Background(), &pps.InspectJobRequest{
			Job:        client.NewJob(jobID),
			BlockState: true,
		})
		if err != nil {
			return err
		}
		if jobInfo.State != pps.JobState_JOB_FAILURE {
			return fmt.Errorf("job did not fail, even though it tried to copy " +
				"pipes, which should be disallowed by Pachyderm")
		}
		return nil
	}, backoff.NewTestingBackOff()))
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

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	aRepo := uniqueString("A")
	require.NoError(t, c.CreateRepo(aRepo))
	bPipeline := uniqueString("B")
	require.NoError(t, c.CreatePipeline(
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(aRepo, "/*"),
		"",
		false,
	))
	cPipeline := uniqueString("C")
	require.NoError(t, c.CreatePipeline(
		cPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("diff %s %s >/pfs/out/file",
			path.Join("/pfs", aRepo, "file"), path.Join("/pfs", bPipeline, "file"))},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewCrossInput(
			client.NewAtomInput(aRepo, "/*"),
			client.NewAtomInput(bPipeline, "/*"),
		),
		"",
		false,
	))
	// commit to aRepo
	commit1, err := c.StartCommit(aRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, commit1.ID))

	commit2, err := c.StartCommit(aRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, commit2.ID))

	aCommit := commit2
	commitIter, err := c.FlushCommit([]*pfs.Commit{aCommit}, []*pfs.Repo{{bPipeline}})
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	bCommit := commitInfos[0].Commit
	commitIter, err = c.FlushCommit([]*pfs.Commit{aCommit, bCommit}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	cCommitInfo := commitInfos[0]
	require.Equal(t, uint64(0), cCommitInfo.SizeBytes)

	// We should only see two commits in aRepo
	commitInfos, err = c.ListCommit(aRepo, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	// There are three commits in the pipeline repos (two from input commits, and
	// one from the CreatePipeline call that created each repo)
	commitInfos, err = c.ListCommit(bPipeline, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	commitInfos, err = c.ListCommit(cPipeline, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
}

// TestProvenance2 tests the following DAG:
//   A
//  / \
// B   C
//  \ /
//   D
func TestProvenance2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	aRepo := uniqueString("A")
	require.NoError(t, c.CreateRepo(aRepo))
	bPipeline := uniqueString("B")
	require.NoError(t, c.CreatePipeline(
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "bfile"), "/pfs/out/bfile"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(aRepo, "/b*"),
		"",
		false,
	))
	cPipeline := uniqueString("C")
	require.NoError(t, c.CreatePipeline(
		cPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "cfile"), "/pfs/out/cfile"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(aRepo, "/c*"),
		"",
		false,
	))
	dPipeline := uniqueString("D")
	require.NoError(t, c.CreatePipeline(
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
			client.NewAtomInput(bPipeline, "/*"),
			client.NewAtomInput(cPipeline, "/*"),
		),
		"",
		false,
	))
	// commit to aRepo
	commit1, err := c.StartCommit(aRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, commit1.ID, "bfile", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, commit1.ID, "cfile", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, commit1.ID))

	commit2, err := c.StartCommit(aRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, commit2.ID, "bfile", strings.NewReader("bar\n"))
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, commit2.ID, "cfile", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, commit2.ID))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit2}, []*pfs.Repo{{dPipeline}})
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	// We should only see two commits in each repo.
	commitInfos, err = c.ListCommit(bPipeline, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	commitInfos, err = c.ListCommit(cPipeline, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	commitInfos, err = c.ListCommit(dPipeline, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	for _, commitInfo := range commitInfos {
		commit := commitInfo.Commit
		buffer := bytes.Buffer{}
		require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, "file", 0, 0, &buffer))
		require.Equal(t, "", buffer.String())
	}
}

// TestFlushCommit
func TestFlushCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	prefix := uniqueString("repo")
	makeRepoName := func(i int) string {
		return fmt.Sprintf("%s-%d", prefix, i)
	}

	sourceRepo := makeRepoName(0)
	require.NoError(t, c.CreateRepo(sourceRepo))

	// Create a five-stage pipeline
	numStages := 5
	for i := 0; i < numStages; i++ {
		repo := makeRepoName(i)
		require.NoError(t, c.CreatePipeline(
			makeRepoName(i+1),
			"",
			[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
			nil,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewAtomInput(repo, "/*"),
			"",
			false,
		))
	}

	for i := 0; i < 10; i++ {
		commit, err := c.StartCommit(sourceRepo, "master")
		require.NoError(t, err)
		_, err = c.PutFile(sourceRepo, commit.ID, "file", strings.NewReader("foo\n"))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(sourceRepo, commit.ID))
		commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(sourceRepo, commit.ID)}, nil)
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, numStages, len(commitInfos))
	}
}

func TestFlushCommitAfterCreatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	repo := uniqueString("data")
	require.NoError(t, c.CreateRepo(repo))

	var commit *pfs.Commit
	var err error
	for i := 0; i < 10; i++ {
		commit, err = c.StartCommit(repo, "")
		require.NoError(t, err)
		_, err = c.PutFile(repo, commit.ID, "file", strings.NewReader(fmt.Sprintf("foo%d\n", i)))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(repo, commit.ID))
	}
	require.NoError(t, c.SetBranch(repo, commit.ID, "master"))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(repo, "/*"),
		"",
		false,
	))
	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(repo, "master")}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, commitIter)
}

// TestRecreatePipeline tracks #432
func TestRecreatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	repo := uniqueString("data")
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	pipeline := uniqueString("pipeline")
	createPipeline := func() {
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
			nil,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewAtomInput(repo, "/*"),
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
		require.NoError(t, err)
		require.Equal(t, 1, len(collectCommitInfos(t, commitIter)))
	}

	// Do it twice.  We expect jobs to be created on both runs.
	createPipeline()
	time.Sleep(5 * time.Second)
	require.NoError(t, c.DeletePipeline(pipeline, true))
	time.Sleep(5 * time.Second)
	createPipeline()
}

func TestDeletePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	repo := uniqueString("data")
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit.ID, uuid.NewWithoutDashes(), strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	pipeline := uniqueString("pipeline")
	createPipeline := func() {
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"sleep", "20"},
			nil,
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewAtomInput(repo, "/*"),
			"",
			false,
		))
	}

	createPipeline()
	time.Sleep(10 * time.Second)
	// Wait for the pipeline to start running
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err := c.InspectPipeline(pipeline)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
			return fmt.Errorf("no running pipeline")
		}
		return nil
	}, backoff.NewTestingBackOff()))
	require.NoError(t, c.DeletePipeline(pipeline, true))
	time.Sleep(5 * time.Second)
	// Wait for the pipeline to disappear
	require.NoError(t, backoff.Retry(func() error {
		_, err := c.InspectPipeline(pipeline)
		if err == nil {
			return fmt.Errorf("expected pipeline to be missing, but it's still present")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// The job should be gone
	jobs, err := c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, len(jobs), 0)

	createPipeline()
	// Wait for the pipeline to start running
	time.Sleep(10 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err := c.InspectPipeline(pipeline)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
			return fmt.Errorf("no running pipeline")
		}
		return nil
	}, backoff.NewTestingBackOff()))
	require.NoError(t, c.DeletePipeline(pipeline, false))
	// Wait for the pipeline to disappear
	time.Sleep(5 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		_, err := c.InspectPipeline(pipeline)
		if err == nil {
			return fmt.Errorf("expected pipeline to be missing, but it's still present")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// The job should still be there, and its state should be "KILLED"
	jobs, err = c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	require.Equal(t, pps.JobState_JOB_KILLED, jobs[0].State)
}

func TestPipelineState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	repo := uniqueString("data")
	require.NoError(t, c.CreateRepo(repo))
	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(repo, "/*"),
		"",
		false,
	))

	// Wait for pipeline to get picked up
	time.Sleep(15 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err := c.InspectPipeline(pipeline)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
			return fmt.Errorf("no running pipeline")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Stop pipeline and wait for the pipeline to pause
	require.NoError(t, c.StopPipeline(pipeline))
	time.Sleep(5 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err := c.InspectPipeline(pipeline)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_PAUSED {
			return fmt.Errorf("pipeline never paused, even though StopPipeline() was called")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Restart pipeline and wait for the pipeline to resume
	require.NoError(t, c.StartPipeline(pipeline))
	time.Sleep(15 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		pipelineInfo, err := c.InspectPipeline(pipeline)
		if err != nil {
			return err
		}
		if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
			return fmt.Errorf("pipeline never started, even though StartPipeline() was called")
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestPipelineJobCounts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	repo := uniqueString("data")
	require.NoError(t, c.CreateRepo(repo))
	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(repo, "/*"),
		"",
		false,
	))

	// Trigger a job by creating a commit
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, commitIter)
	jobInfos, err := c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	// There are two jobs even though we only made one commit: an empty one
	// corresponding to the commit triggered by the new pipeline, and a real one,
	// corresponding to the commit triggered by our input commit above
	require.Equal(t, 2, len(jobInfos))
	inspectJobRequest := &pps.InspectJobRequest{
		Job:        jobInfos[0].Job,
		BlockState: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	_, err = c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)

	// check that the job has been accounted for
	pipelineInfo, err := c.InspectPipeline(pipeline)
	require.NoError(t, err)
	require.Equal(t, int32(2), pipelineInfo.JobCounts[int32(pps.JobState_JOB_SUCCESS)])
}

// TODO(msteffen): This test breaks the suite when run against cloud providers,
// because killing the pachd pod breaks the connection with pachctl port-forward
func TestDeleteAfterMembershipChange(t *testing.T) {
	t.Skip("This is causing intermittent CI failures")

	test := func(up bool) {
		repo := uniqueString("TestDeleteAfterMembershipChange")
		c := getPachClient(t)
		defer require.NoError(t, c.DeleteAll())
		require.NoError(t, c.CreateRepo(repo))
		_, err := c.StartCommit(repo, "master")
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(repo, "master"))
		scalePachdRandom(t, up)
		c = getUsablePachClient(t)
		require.NoError(t, c.DeleteRepo(repo, false))
	}
	test(true)
	test(false)
}

// TODO(msteffen): This test breaks the suite when run against cloud providers,
// because killing the pachd pod breaks the connection with pachctl port-forward
func TestPachdRestartResumesRunningJobs(t *testing.T) {
	t.Skip("This is causing intermittent CI failures")
	// this test cannot be run in parallel because it restarts everything which breaks other tests.
	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPachdRestartPickUpRunningJobs")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{
			"sleep 10",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/"),
		"",
		false,
	))
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	time.Sleep(5 * time.Second)

	jobInfos, err := c.ListJob(pipelineName, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	require.Equal(t, pps.JobState_JOB_RUNNING, jobInfos[0].State)

	restartOne(t)
	// need a new client because the old one will have a defunct connection
	c = getUsablePachClient(t)

	jobInfo, err := c.InspectJob(jobInfos[0].Job.ID, true)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
}

// TestUpdatePipelineThatHasNoOutput tracks #1637
func TestUpdatePipelineThatHasNoOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestUpdatePipelineThatHasNoOutput")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"sh"},
		[]string{"exit 1"},
		nil,
		client.NewAtomInput(dataRepo, "/"),
		"",
		false,
	))

	// Wait for job to spawn
	var jobInfos []*pps.JobInfo
	time.Sleep(10 * time.Second)
	require.NoError(t, backoff.Retry(func() error {
		var err error
		jobInfos, err = c.ListJob(pipeline, nil, nil)
		if err != nil {
			return err
		}
		if len(jobInfos) < 1 {
			return fmt.Errorf("job not spawned")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	jobInfo, err := c.InspectJob(jobInfos[0].Job.ID, true)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)

	// Now we update the pipeline
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"sh"},
		[]string{"exit 1"},
		nil,
		client.NewAtomInput(dataRepo, "/"),
		"",
		true,
	))
}

func TestAcceptReturnCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestAcceptReturnCode")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	pipelineName := uniqueString("TestAcceptReturnCode")
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{pipelineName},
			Transform: &pps.Transform{
				Cmd:              []string{"sh"},
				Stdin:            []string{"exit 1"},
				AcceptReturnCode: []int64{1},
			},
			Input: client.NewAtomInput(dataRepo, "/*"),
		},
	)
	require.NoError(t, err)

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	jobInfos, err := c.ListJob(pipelineName, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))

	jobInfo, err := c.InspectJob(jobInfos[0].Job.ID, true)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
}

// TODO(msteffen): This test breaks the suite when run against cloud providers,
// because killing the pachd pod breaks the connection with pachctl port-forward
func TestRestartAll(t *testing.T) {
	t.Skip("This is causing intermittent CI failures")
	// this test cannot be run in parallel because it restarts everything which breaks other tests.
	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestRestartAll_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, commitIter)

	restartAll(t)

	// need a new client because the old one will have a defunct connection
	c = getUsablePachClient(t)

	// Wait a little for pipelines to restart
	time.Sleep(10 * time.Second)
	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_RUNNING, pipelineInfo.State)
	_, err = c.InspectRepo(dataRepo)
	require.NoError(t, err)
	_, err = c.InspectCommit(dataRepo, commit.ID)
	require.NoError(t, err)
}

// TODO(msteffen): This test breaks the suite when run against cloud providers,
// because killing the pachd pod breaks the connection with pachctl port-forward
func TestRestartOne(t *testing.T) {
	t.Skip("This is causing intermittent CI failures")
	// this test cannot be run in parallel because it restarts everything which breaks other tests.
	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestRestartOne_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/"),
		"",
		false,
	))
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, commitIter)

	restartOne(t)

	// need a new client because the old one will have a defunct connection
	c = getUsablePachClient(t)

	_, err = c.InspectPipeline(pipelineName)
	require.NoError(t, err)
	_, err = c.InspectRepo(dataRepo)
	require.NoError(t, err)
	_, err = c.InspectCommit(dataRepo, commit.ID)
	require.NoError(t, err)
}

func TestPrettyPrinting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPrettyPrinting_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{pipelineName},
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			ResourceRequests: &pps.ResourceSpec{
				Memory: "100M",
				Cpu:    0.5,
			},
			Input: client.NewAtomInput(dataRepo, "/*"),
		})
	require.NoError(t, err)
	// Do a commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	repoInfo, err := c.InspectRepo(dataRepo)
	require.NoError(t, err)
	require.NoError(t, pfspretty.PrintDetailedRepoInfo(repoInfo))
	for _, commitInfo := range commitInfos {
		require.NoError(t, pfspretty.PrintDetailedCommitInfo(commitInfo))
	}
	fileInfo, err := c.InspectFile(dataRepo, commit.ID, "file")
	require.NoError(t, err)
	require.NoError(t, pfspretty.PrintDetailedFileInfo(fileInfo))
	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)
	require.NoError(t, ppspretty.PrintDetailedPipelineInfo(pipelineInfo))
	jobInfos, err := c.ListJob("", nil, nil)
	require.NoError(t, err)
	require.True(t, len(jobInfos) > 0)
	require.NoError(t, ppspretty.PrintDetailedJobInfo(jobInfos[0]))
}

func TestDeleteAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// this test cannot be run in parallel because it deletes everything
	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestDeleteAll_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/"),
		"",
		false,
	))
	// Do commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(collectCommitInfos(t, commitIter)))
	require.NoError(t, c.DeleteAll())
	repoInfos, err := c.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(repoInfos))
	pipelineInfos, err := c.ListPipeline()
	require.NoError(t, err)
	require.Equal(t, 0, len(pipelineInfos))
	jobInfos, err := c.ListJob("", nil, nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(jobInfos))
}

func TestRecursiveCp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestRecursiveCp_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("TestRecursiveCp")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"sh"},
		[]string{
			fmt.Sprintf("cp -r /pfs/%s /pfs/out", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))
	// Do commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, err = c.PutFile(
			dataRepo,
			commit.ID,
			fmt.Sprintf("file%d", i),
			strings.NewReader(strings.Repeat("foo\n", 10000)),
		)
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(collectCommitInfos(t, commitIter)))
}

func TestPipelineUniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	repo := uniqueString("data")
	require.NoError(t, c.CreateRepo(repo))
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(repo, "/"),
		"",
		false,
	))
	err := c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(repo, "/"),
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

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestUpdatePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"echo foo >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	_, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, "master", "file", strings.NewReader("1"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, "master"))

	iter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, iter)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, "master", "file", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())

	// Update the pipeline, this will not create a new pipeline as reprocess
	// isn't set to true.
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"echo bar >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		true,
	))

	_, err = c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, "master", "file", strings.NewReader("2"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, "master"))
	iter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, iter)

	buffer.Reset()
	require.NoError(t, c.GetFile(pipelineName, "master", "file", 0, 0, &buffer))
	require.Equal(t, "bar\n", buffer.String())

	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{"echo buzz >/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input:     client.NewAtomInput(dataRepo, "/*"),
			Update:    true,
			Reprocess: true,
		})
	require.NoError(t, err)

	iter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, iter)
	buffer.Reset()
	require.NoError(t, c.GetFile(pipelineName, "master", "file", 0, 0, &buffer))
	require.Equal(t, "buzz\n", buffer.String())
}

func TestUpdatePipelineRunningJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestUpdatePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"sleep 1000"},
		&pps.ParallelismSpec{
			Constant: 2,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	numFiles := 50
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(""))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit2.ID, fmt.Sprintf("file-%d", i+numFiles), strings.NewReader(""))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))

	b := backoff.NewTestingBackOff()
	b.MaxElapsedTime = 30 * time.Second
	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob(pipelineName, nil, nil)
		if err != nil {
			return err
		}
		if len(jobInfos) != 1 {
			return fmt.Errorf("wrong number of jobs")
		}
		if pps.JobState_JOB_RUNNING != jobInfos[0].State {
			return fmt.Errorf("wrong state: %v for %s", jobInfos[0].State, jobInfos[0].Job.ID)
		}
		return nil
	}, b))

	// Update the pipeline. This will not create a new pipeline as reprocess
	// isn't set to true.
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"true"},
		&pps.ParallelismSpec{
			Constant: 2,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		true,
	))
	iter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, iter)

	// Currently, commits finish shortly before their respecive JobInfo documents
	// are updated (the pipeline master receives the commit update and then
	// updates the JobInfo document). Wait briefly for this to happen
	time.Sleep(10 * time.Second)

	jobInfos, err := c.ListJob(pipelineName, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
	require.Equal(t, pps.JobState_JOB_SUCCESS.String(), jobInfos[0].State.String())
	require.Equal(t, pps.JobState_JOB_KILLED.String(), jobInfos[1].State.String())
}

func TestStopPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	// Stop the pipeline, so it doesn't process incoming commits
	require.NoError(t, c.StopPipeline(pipelineName))

	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	// wait for 10 seconds and check that no commit has been outputted
	time.Sleep(10 * time.Second)
	commits, err := c.ListCommit(pipelineName, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, len(commits), 0)

	// Restart pipeline, and make sure old commit is processed
	require.NoError(t, c.StartPipeline(pipelineName))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, commitInfos[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

func TestPipelineAutoScaledown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipelineAutoScaleDown")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline-auto-scaledown")
	parallelism := 4
	scaleDownThreshold := time.Duration(10 * time.Second)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"sh"},
				Stdin: []string{
					"echo success",
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: uint64(parallelism),
			},
			ResourceRequests: &pps.ResourceSpec{
				Memory: "100M",
			},
			Input:              client.NewAtomInput(dataRepo, "/"),
			ScaleDownThreshold: types.DurationProto(scaleDownThreshold),
		})
	require.NoError(t, err)

	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)

	// Wait for the pipeline to scale down
	time.Sleep(10 * time.Second)
	b := backoff.NewTestingBackOff()
	b.MaxElapsedTime = scaleDownThreshold + 30*time.Second
	checkScaleDown := func() error {
		rc, err := pipelineRc(t, pipelineInfo)
		if err != nil {
			return err
		}
		if *rc.Spec.Replicas != 1 {
			return fmt.Errorf("rc.Spec.Replicas should be 1")
		}
		if !rc.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().IsZero() {
			return fmt.Errorf("resource requests should be zero when the pipeline is in scale-down mode")
		}
		return nil
	}
	require.NoError(t, backoff.Retry(checkScaleDown, b))

	// Trigger a job
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	rc, err := pipelineRc(t, pipelineInfo)
	require.NoError(t, err)
	require.Equal(t, int32(parallelism), int32(*rc.Spec.Replicas))
	// Check that the resource requests have been reset
	require.Equal(t, "100M", rc.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String())

	// Once the job finishes, the pipeline will scale down again
	b.Reset()
	require.NoError(t, backoff.Retry(checkScaleDown, b))
}

func TestPipelineEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// make a secret to reference
	k := getKubeClient(t)
	secretName := uniqueString("test-secret")
	_, err := k.CoreV1().Secrets(v1.NamespaceDefault).Create(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
			},
			Data: map[string][]byte{
				"foo": []byte("foo\n"),
			},
		},
	)
	require.NoError(t, err)
	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipelineEnv_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"sh"},
				Stdin: []string{
					"ls /var/secret",
					"cat /var/secret/foo > /pfs/out/foo",
					"echo $bar> /pfs/out/bar",
					"echo $foo> /pfs/out/foo_env",
				},
				Env: map[string]string{"bar": "bar"},
				Secrets: []*pps.Secret{
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
			Input: client.NewAtomInput(dataRepo, "/*"),
		})
	require.NoError(t, err)
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, commitInfos[0].Commit.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(pipelineName, commitInfos[0].Commit.ID, "foo_env", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(pipelineName, commitInfos[0].Commit.ID, "bar", 0, 0, &buffer))
	require.Equal(t, "bar\n", buffer.String())
}

func TestPipelineWithFullObjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	commitInfoIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit1.ID)}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitInfoIter)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	// Do second commit to repo
	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))
	commitInfoIter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitInfoIter)
	require.Equal(t, 1, len(commitInfos))
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())
}

func TestPipelineWithExistingInputCommits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	// Do second commit to repo
	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	commitInfoIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitInfoIter)
	require.Equal(t, 1, len(commitInfos))
	buffer := bytes.Buffer{}
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())

	// Check that one output commit is created (processing the inputs' head commits)
	commitInfos, err = c.ListCommit(pipelineName, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
}

func TestPipelineThatSymlinks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	// create repos
	dataRepo := uniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
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
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/"),
		"",
		false,
	))

	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "foo", strings.NewReader("foo"))
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "dir1/bar", strings.NewReader("bar"))
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "dir2/foo", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	commitInfoIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitInfoIter)
	require.Equal(t, 1, len(commitInfos))

	// Check that the output files are identical to the input files.
	buffer := bytes.Buffer{}
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "bar", 0, 0, &buffer))
	require.Equal(t, "bar", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "dir/dir2/foo", 0, 0, &buffer))
	require.Equal(t, "foo", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "buzz", 0, 0, &buffer))
	require.Equal(t, "buzz\n", buffer.String())

	// Make sure that we skipped the upload by checking that the input file
	// and the output file have the same object refs.
	inputFooFileInfo, err := c.InspectFile(dataRepo, commit.ID, "foo")
	require.NoError(t, err)
	outputFooFileInfo, err := c.InspectFile(pipelineName, commitInfos[0].Commit.ID, "foo")
	require.NoError(t, err)
	require.Equal(t, inputFooFileInfo.Objects, outputFooFileInfo.Objects)
	inputFooFileInfo, err = c.InspectFile(dataRepo, commit.ID, "dir1/bar")
	require.NoError(t, err)
	outputFooFileInfo, err = c.InspectFile(pipelineName, commitInfos[0].Commit.ID, "bar")
	require.NoError(t, err)
	require.Equal(t, inputFooFileInfo.Objects, outputFooFileInfo.Objects)
	inputFooFileInfo, err = c.InspectFile(dataRepo, commit.ID, "dir2/foo")
	require.NoError(t, err)
	outputFooFileInfo, err = c.InspectFile(pipelineName, commitInfos[0].Commit.ID, "dir/dir2/foo")
	require.NoError(t, err)
	require.Equal(t, inputFooFileInfo.Objects, outputFooFileInfo.Objects)
}

// TestChainedPipelines tracks https://github.com/pachyderm/pachyderm/issues/797
func TestChainedPipelines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	aRepo := uniqueString("A")
	require.NoError(t, c.CreateRepo(aRepo))

	dRepo := uniqueString("D")
	require.NoError(t, c.CreateRepo(dRepo))

	aCommit, err := c.StartCommit(aRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, "master", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, "master"))

	dCommit, err := c.StartCommit(dRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dRepo, "master", "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dRepo, "master"))

	bPipeline := uniqueString("B")
	require.NoError(t, c.CreatePipeline(
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(aRepo, "/"),
		"",
		false,
	))

	cPipeline := uniqueString("C")
	require.NoError(t, c.CreatePipeline(
		cPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/file /pfs/out/bFile", bPipeline),
			fmt.Sprintf("cp /pfs/%s/file /pfs/out/dFile", dRepo)},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewCrossInput(
			client.NewAtomInput(bPipeline, "/"),
			client.NewAtomInput(dRepo, "/"),
		),
		"",
		false,
	))
	resultIter, err := c.FlushCommit([]*pfs.Commit{aCommit, dCommit}, nil)
	require.NoError(t, err)
	results := collectCommitInfos(t, resultIter)
	require.Equal(t, 1, len(results))
	require.Equal(t, cPipeline, results[0].Commit.Repo.Name)
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(cPipeline, results[0].Commit.ID, "bFile", 0, 0, &buf))
	require.Equal(t, "foo\n", buf.String())
	buf.Reset()
	require.NoError(t, c.GetFile(cPipeline, results[0].Commit.ID, "dFile", 0, 0, &buf))
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

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	aRepo := uniqueString("A")
	require.NoError(t, c.CreateRepo(aRepo))

	eRepo := uniqueString("E")
	require.NoError(t, c.CreateRepo(eRepo))

	aCommit, err := c.StartCommit(aRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, "master", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, "master"))

	eCommit, err := c.StartCommit(eRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(eRepo, "master", "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(eRepo, "master"))

	bPipeline := uniqueString("B")
	require.NoError(t, c.CreatePipeline(
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(aRepo, "/"),
		"",
		false,
	))

	cPipeline := uniqueString("C")
	require.NoError(t, c.CreatePipeline(
		cPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/file /pfs/out/bFile", bPipeline),
			fmt.Sprintf("cp /pfs/%s/file /pfs/out/eFile", eRepo)},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewCrossInput(
			client.NewAtomInput(bPipeline, "/"),
			client.NewAtomInput(eRepo, "/"),
		),
		"",
		false,
	))

	dPipeline := uniqueString("D")
	require.NoError(t, c.CreatePipeline(
		dPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/bFile /pfs/out/bFile", cPipeline),
			fmt.Sprintf("cp /pfs/%s/eFile /pfs/out/eFile", cPipeline)},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(cPipeline, "/"),
		"",
		false,
	))

	resultsIter, err := c.FlushCommit([]*pfs.Commit{aCommit, eCommit}, nil)
	require.NoError(t, err)
	results := collectCommitInfos(t, resultsIter)
	require.Equal(t, 2, len(results))

	eCommit2, err := c.StartCommit(eRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(eRepo, "master", "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(eRepo, "master"))

	resultsIter, err = c.FlushCommit([]*pfs.Commit{eCommit2}, nil)
	require.NoError(t, err)
	results = collectCommitInfos(t, resultsIter)
	require.Equal(t, 2, len(results))

	// Get number of jobs triggered in pipeline D
	jobInfos, err := c.ListJob(dPipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
}

func collectCommitInfos(t testing.TB, commitInfoIter client.CommitInfoIterator) []*pfs.CommitInfo {
	var commitInfos []*pfs.CommitInfo
	for {
		commitInfo, err := commitInfoIter.Next()
		if err == io.EOF {
			return commitInfos
		}
		require.NoError(t, err)
		commitInfos = append(commitInfos, commitInfo)
	}
}

func TestParallelismSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	kubeclient := getKubeClient(t)
	nodes, err := kubeclient.CoreV1().Nodes().List(metav1.ListOptions{})
	numNodes := len(nodes.Items)

	// Test Constant strategy
	parellelism, err := ppsutil.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Constant: 7,
	})
	require.NoError(t, err)
	require.Equal(t, 7, parellelism)

	// Coefficient == 1 (basic test)
	// TODO(msteffen): This test can fail when run against cloud providers, if the
	// remote cluster has more than one node (in which case "Coefficient: 1" will
	// cause more than 1 worker to start)
	parellelism, err = ppsutil.GetExpectedNumWorkers(kubeclient, &pps.ParallelismSpec{
		Coefficient: 1,
	})
	require.NoError(t, err)
	require.Equal(t, numNodes, parellelism)

	// Coefficient > 1
	parellelism, err = ppsutil.GetExpectedNumWorkers(kubeclient, &pps.ParallelismSpec{
		Coefficient: 2,
	})
	require.NoError(t, err)
	require.Equal(t, 2*numNodes, parellelism)

	// Make sure we start at least one worker
	parellelism, err = ppsutil.GetExpectedNumWorkers(kubeclient, &pps.ParallelismSpec{
		Coefficient: 0.01,
	})
	require.NoError(t, err)
	require.Equal(t, 1, parellelism)

	// Test 0-initialized JobSpec
	parellelism, err = ppsutil.GetExpectedNumWorkers(kubeclient, &pps.ParallelismSpec{})
	require.NoError(t, err)
	require.Equal(t, 1, parellelism)

	// Test nil JobSpec
	parellelism, err = ppsutil.GetExpectedNumWorkers(kubeclient, nil)
	require.NoError(t, err)
	require.Equal(t, 1, parellelism)
}

func TestPipelineJobDeletion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/"),
		"",
		false,
	))

	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)

	_, err = commitIter.Next()
	require.NoError(t, err)

	// Now delete the corresponding job
	jobInfos, err := c.ListJob(pipelineName, nil, nil)
	require.NoError(t, err)
	// One empty job from the pipeline creation, and one real job
	require.Equal(t, 2, len(jobInfos))
	err = c.DeleteJob(jobInfos[0].Job.ID)
	require.NoError(t, err)
}

func TestStopJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestStopJob")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline-stop-job")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"sleep", "20"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/"),
		"",
		false,
	))

	// Create two input commits to trigger two jobs.
	// We will stop the first job midway through, and assert that the
	// second job finishes.
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit2.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))

	var jobID string
	b := backoff.NewTestingBackOff()
	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob(pipelineName, nil, nil)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return fmt.Errorf("len(jobInfos) should be 1")
		}
		jobID = jobInfos[0].Job.ID
		if pps.JobState_JOB_RUNNING != jobInfos[0].State {
			return fmt.Errorf("jobInfos[0] has the wrong state")
		}
		return nil
	}, b))

	// Now stop the first job
	err = c.StopJob(jobID)
	require.NoError(t, err)
	jobInfo, err := c.InspectJob(jobID, true)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_KILLED, jobInfo.State)

	b.Reset()
	// Check that the second job completes
	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob(pipelineName, nil, nil)
		require.NoError(t, err)
		if len(jobInfos) != 2 {
			return fmt.Errorf("len(jobInfos) should be 2")
		}
		jobID = jobInfos[0].Job.ID
		return nil
	}, b))
	jobInfo, err = c.InspectJob(jobID, true)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
}

func TestGetLogs(t *testing.T) {
	testGetLogs(t, false)
}

func TestGetLogsWithStats(t *testing.T) {
	testGetLogs(t, true)
}

func testGetLogs(t *testing.T, enableStats bool) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"sh"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/file /pfs/out/file", dataRepo),
					"echo foo",
					"echo foo",
				},
			},
			Input:       client.NewAtomInput(dataRepo, "/*"),
			EnableStats: enableStats,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 4,
			},
		})
	require.NoError(t, err)

	// Commit data to repo and flush commit
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, "master", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, "master"))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	// Get logs from pipeline, using pipeline
	iter := c.GetLogs(pipelineName, "", nil, "", false, false, 0)
	var numLogs int
	for iter.Next() {
		numLogs++
		require.True(t, iter.Message().Message != "")
	}
	require.Equal(t, 8, numLogs)
	require.NoError(t, iter.Err())

	// Get logs from pipeline, using pipeline
	iter = c.GetLogs(pipelineName, "", nil, "", false, false, 2)
	numLogs = 0
	for iter.Next() {
		numLogs++
		require.True(t, iter.Message().Message != "")
	}
	require.Equal(t, 2, numLogs)
	require.NoError(t, iter.Err())

	// Get logs from pipeline, using a pipeline that doesn't exist. There should
	// be an error
	iter = c.GetLogs("__DOES_NOT_EXIST__", "", nil, "", false, false, 0)
	require.False(t, iter.Next())
	require.YesError(t, iter.Err())
	require.Matches(t, "could not get", iter.Err().Error())

	// Get logs from pipeline, using job
	// (1) Get job ID, from pipeline that just ran
	jobInfos, err := c.ListJob(pipelineName, nil, nil)
	require.NoError(t, err)
	require.True(t, len(jobInfos) == 1)
	// (2) Get logs using extracted job ID
	// wait for logs to be collected
	time.Sleep(10 * time.Second)
	iter = c.GetLogs("", jobInfos[0].Job.ID, nil, "", false, false, 0)
	numLogs = 0
	for iter.Next() {
		numLogs++
		require.True(t, iter.Message().Message != "")
	}
	// Make sure that we've seen some logs
	require.True(t, numLogs > 0)
	require.NoError(t, iter.Err())

	// Get logs from pipeline, using a job that doesn't exist. There should
	// be an error
	iter = c.GetLogs("", "__DOES_NOT_EXIST__", nil, "", false, false, 0)
	require.False(t, iter.Next())
	require.YesError(t, iter.Err())
	require.Matches(t, "could not get", iter.Err().Error())

	// Filter logs based on input (using file that exists). Get logs using file
	// path, hex hash, and base64 hash, and make sure you get the same log lines
	fileInfo, err := c.InspectFile(dataRepo, commit.ID, "/file")
	require.NoError(t, err)

	// TODO(msteffen) This code shouldn't be wrapped in a backoff, but for some
	// reason GetLogs is not yet 100% consistent. This reduces flakes in testing.
	require.NoError(t, backoff.Retry(func() error {
		pathLog := c.GetLogs("", jobInfos[0].Job.ID, []string{"/file"}, "", false, false, 0)

		hexHash := "19fdf57bdf9eb5a9602bfa9c0e6dd7ed3835f8fd431d915003ea82747707be66"
		require.Equal(t, hexHash, hex.EncodeToString(fileInfo.Hash)) // sanity-check test
		hexLog := c.GetLogs("", jobInfos[0].Job.ID, []string{hexHash}, "", false, false, 0)

		base64Hash := "Gf31e9+etalgK/qcDm3X7Tg1+P1DHZFQA+qCdHcHvmY="
		require.Equal(t, base64Hash, base64.StdEncoding.EncodeToString(fileInfo.Hash))
		base64Log := c.GetLogs("", jobInfos[0].Job.ID, []string{base64Hash}, "", false, false, 0)

		numLogs = 0
		for {
			havePathLog, haveHexLog, haveBase64Log := pathLog.Next(), hexLog.Next(), base64Log.Next()
			if havePathLog != haveHexLog || haveHexLog != haveBase64Log {
				return fmt.Errorf("Unequal log lengths")
			}
			if !havePathLog {
				break
			}
			numLogs++
			if pathLog.Message().Message != hexLog.Message().Message ||
				hexLog.Message().Message != base64Log.Message().Message {
				return fmt.Errorf(
					"unequal logs, pathLogs: \"%s\" hexLog: \"%s\" base64Log: \"%s\"",
					pathLog.Message().Message,
					hexLog.Message().Message,
					base64Log.Message().Message)
			}
		}
		for _, logsiter := range []*client.LogsIter{pathLog, hexLog, base64Log} {
			if logsiter.Err() != nil {
				return logsiter.Err()
			}
		}
		if numLogs == 0 {
			return fmt.Errorf("no logs found")
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Filter logs based on input (using file that doesn't exist). There should
	// be no logs
	iter = c.GetLogs("", jobInfos[0].Job.ID, []string{"__DOES_NOT_EXIST__"}, "", false, false, 0)
	require.False(t, iter.Next())
	require.NoError(t, iter.Err())

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	iter = c.WithCtx(ctx).GetLogs(pipelineName, "", nil, "", false, false, 0)
	numLogs = 0
	for iter.Next() {
		numLogs++
		if numLogs == 8 {
			// Do another commit so there's logs to receive with follow
			_, err = c.StartCommit(dataRepo, "master")
			require.NoError(t, err)
			_, err = c.PutFile(dataRepo, "master", "file", strings.NewReader("bar\n"))
			require.NoError(t, err)
			require.NoError(t, c.FinishCommit(dataRepo, "master"))
		}
		require.True(t, iter.Message().Message != "")
		if numLogs == 16 {
			break
		}
	}
	require.NoError(t, iter.Err())
}

func TestPfsPutFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	repo1 := uniqueString("TestPfsPutFile1")
	require.NoError(t, c.CreateRepo(repo1))
	repo2 := uniqueString("TestPfsPutFile2")
	require.NoError(t, c.CreateRepo(repo2))

	commit1, err := c.StartCommit(repo1, "")
	require.NoError(t, err)
	_, err = c.PutFile(repo1, commit1.ID, "file1", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = c.PutFile(repo1, commit1.ID, "file2", strings.NewReader("bar\n"))
	require.NoError(t, err)
	_, err = c.PutFile(repo1, commit1.ID, "dir1/file3", strings.NewReader("fizz\n"))
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, err = c.PutFile(repo1, commit1.ID, fmt.Sprintf("dir1/dir2/file%d", i), strings.NewReader(fmt.Sprintf("content%d\n", i)))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(repo1, commit1.ID))

	commit2, err := c.StartCommit(repo2, "")
	require.NoError(t, err)
	err = c.PutFileURL(repo2, commit2.ID, "file", fmt.Sprintf("pfs://0.0.0.0:650/%s/%s/file1", repo1, commit1.ID), false, false)
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo2, commit2.ID))
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(repo2, commit2.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\n", buf.String())

	commit3, err := c.StartCommit(repo2, "")
	require.NoError(t, err)
	err = c.PutFileURL(repo2, commit3.ID, "", fmt.Sprintf("pfs://0.0.0.0:650/%s/%s", repo1, commit1.ID), true, false)
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo2, commit3.ID))
	buf = bytes.Buffer{}
	require.NoError(t, c.GetFile(repo2, commit3.ID, "file1", 0, 0, &buf))
	require.Equal(t, "foo\n", buf.String())
	buf = bytes.Buffer{}
	require.NoError(t, c.GetFile(repo2, commit3.ID, "file2", 0, 0, &buf))
	require.Equal(t, "bar\n", buf.String())
	buf = bytes.Buffer{}
	require.NoError(t, c.GetFile(repo2, commit3.ID, "dir1/file3", 0, 0, &buf))
	require.Equal(t, "fizz\n", buf.String())
	for i := 0; i < 100; i++ {
		buf = bytes.Buffer{}
		require.NoError(t, c.GetFile(repo2, commit3.ID, fmt.Sprintf("dir1/dir2/file%d", i), 0, 0, &buf))
		require.Equal(t, fmt.Sprintf("content%d\n", i), buf.String())
	}
}

func TestAllDatumsAreProcessed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo1 := uniqueString("TestAllDatumsAreProcessed_data1")
	require.NoError(t, c.CreateRepo(dataRepo1))
	dataRepo2 := uniqueString("TestAllDatumsAreProcessed_data2")
	require.NoError(t, c.CreateRepo(dataRepo2))

	commit1, err := c.StartCommit(dataRepo1, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo1, "master", "file1", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo1, "master", "file2", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo1, "master"))

	commit2, err := c.StartCommit(dataRepo2, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo2, "master", "file1", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo2, "master", "file2", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo2, "master"))

	require.NoError(t, c.CreatePipeline(
		uniqueString("TestAllDatumsAreProcessed_pipelines"),
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/%s/* /pfs/%s/* > /pfs/out/file", dataRepo1, dataRepo2),
		},
		nil,
		client.NewCrossInput(
			client.NewAtomInput(dataRepo1, "/*"),
			client.NewAtomInput(dataRepo2, "/*"),
		),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1, commit2}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	// should be 8 because each file gets copied twice due to cross product
	require.Equal(t, strings.Repeat("foo\n", 8), buf.String())
}

func TestDatumStatusRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestDatumDedup_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	// This pipeline sleeps for 20 secs per datum
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"sleep 20",
		},
		nil,
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))
	var jobID string
	var datumStarted time.Time
	// checkStatus waits for 'pipeline' to start and makes sure that each time
	// it's called, the datum being processes was started at a new and later time
	// (than the last time checkStatus was called)
	checkStatus := func() {
		require.NoError(t, backoff.Retry(func() error {
			// get the
			jobs, err := c.ListJob(pipeline, nil, nil)
			require.NoError(t, err)
			if len(jobs) == 0 {
				return fmt.Errorf("no jobs found")
			}

			jobID = jobs[0].Job.ID
			jobInfo, err := c.InspectJob(jobs[0].Job.ID, false)
			require.NoError(t, err)
			if len(jobInfo.WorkerStatus) == 0 {
				return fmt.Errorf("no worker statuses")
			}
			if jobInfo.WorkerStatus[0].JobID == jobInfo.Job.ID {
				// The first time this function is called, datumStarted is zero
				// so `Before` is true for any non-zero time.
				_datumStarted, err := types.TimestampFromProto(jobInfo.WorkerStatus[0].Started)
				require.NoError(t, err)
				require.True(t, datumStarted.Before(_datumStarted))
				datumStarted = _datumStarted
				return nil
			}
			return fmt.Errorf("worker status from wrong job")
		}, backoff.RetryEvery(time.Second).For(30*time.Second)))
	}
	checkStatus()
	require.NoError(t, c.RestartDatum(jobID, []string{"/file"}))
	checkStatus()

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
}

func TestUseMultipleWorkers(t *testing.T) {
	t.Skip("flaky")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestUseMultipleWorkers_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < 20; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file%d", i), strings.NewReader("foo"))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	// This pipeline sleeps for 10 secs per datum
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"sleep 10",
		},
		&pps.ParallelismSpec{
			Constant: 2,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))
	// Get job info 2x/sec for 20s until we confirm two workers for the current job
	require.NoError(t, backoff.Retry(func() error {
		jobs, err := c.ListJob(pipeline, nil, nil)
		if err != nil {
			return fmt.Errorf("could not list job: %s", err.Error())
		}
		if len(jobs) == 0 {
			return fmt.Errorf("failed to find job")
		}
		jobInfo, err := c.InspectJob(jobs[0].Job.ID, false)
		if err != nil {
			return fmt.Errorf("could not inspect job: %s", err.Error())
		}
		if len(jobInfo.WorkerStatus) != 2 {
			return fmt.Errorf("incorrect number of statuses: %v", len(jobInfo.WorkerStatus))
		}
		return nil
	}, backoff.RetryEvery(500*time.Millisecond).For(20*time.Second)))
}

// TestSystemResourceRequest doesn't create any jobs or pipelines, it
// just makes sure that when pachyderm is deployed, we give rethinkdb, pachd,
// and etcd default resource requests. This prevents them from overloading
// nodes and getting evicted, which can slow down or break a cluster.
func TestSystemResourceRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	kubeClient := getKubeClient(t)

	// Expected resource requests for pachyderm system pods:
	defaultLocalMem := map[string]string{
		"pachd": "512M",
		"etcd":  "256M",
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
			podList, err := kubeClient.CoreV1().Pods(v1.NamespaceDefault).List(
				metav1.ListOptions{
					LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
						map[string]string{"app": app, "suite": "pachyderm"},
					)),
				})
			if err != nil {
				return err
			}
			if len(podList.Items) < 1 {
				return fmt.Errorf("could not find pod for %s", app) // retry
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

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipelineResourceRequest")
	pipelineName := uniqueString("TestPipelineResourceRequest_Pipeline")
	require.NoError(t, c.CreateRepo(dataRepo))
	// Resources are not yet in client.CreatePipeline() (we may add them later)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{pipelineName},
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			ResourceRequests: &pps.ResourceSpec{
				Memory: "100M",
				Cpu:    0.5,
			},
			Input: &pps.Input{
				Atom: &pps.AtomInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)

	// Get info about the pipeline pods from k8s & check for resources
	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)

	var container v1.Container
	rcName := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	kubeClient := getKubeClient(t)
	err = backoff.Retry(func() error {
		podList, err := kubeClient.CoreV1().Pods(v1.NamespaceDefault).List(
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{"app": rcName},
				)),
			})
		if err != nil {
			return err // retry
		}
		if len(podList.Items) != 1 || len(podList.Items[0].Spec.Containers) == 0 {
			return fmt.Errorf("could not find single container for pipeline %s", pipelineInfo.Pipeline.Name)
		}
		container = podList.Items[0].Spec.Containers[0]
		return nil // no more retries
	}, backoff.NewTestingBackOff())
	require.NoError(t, err)
	// Make sure a CPU and Memory request are both set
	cpu, ok := container.Resources.Requests[v1.ResourceCPU]
	require.True(t, ok)
	require.Equal(t, "500m", cpu.String())
	mem, ok := container.Resources.Requests[v1.ResourceMemory]
	require.True(t, ok)
	require.Equal(t, "100M", mem.String())
	_, ok = container.Resources.Requests[v1.ResourceNvidiaGPU]
	require.False(t, ok)
}

func TestPipelineResourceLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipelineResourceLimit")
	pipelineName := uniqueString("TestPipelineResourceLimit_Pipeline")
	require.NoError(t, c.CreateRepo(dataRepo))
	// Resources are not yet in client.CreatePipeline() (we may add them later)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{pipelineName},
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
				Atom: &pps.AtomInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)

	// Get info about the pipeline pods from k8s & check for resources
	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)

	var container v1.Container
	rcName := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	kubeClient := getKubeClient(t)
	err = backoff.Retry(func() error {
		podList, err := kubeClient.CoreV1().Pods(v1.NamespaceDefault).List(metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
				map[string]string{"app": rcName, "suite": "pachyderm"},
			)),
		})
		if err != nil {
			return err // retry
		}
		if len(podList.Items) != 1 || len(podList.Items[0].Spec.Containers) == 0 {
			return fmt.Errorf("could not find single container for pipeline %s", pipelineInfo.Pipeline.Name)
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
	_, ok = container.Resources.Requests[v1.ResourceNvidiaGPU]
	require.False(t, ok)
}

func TestPipelineResourceLimitDefaults(t *testing.T) {
	// We need to make sure GPU is set to 0 for k8s 1.8
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipelineResourceLimit")
	pipelineName := uniqueString("TestPipelineResourceLimit_Pipeline")
	require.NoError(t, c.CreateRepo(dataRepo))
	// Resources are not yet in client.CreatePipeline() (we may add them later)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{pipelineName},
			Transform: &pps.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: &pps.Input{
				Atom: &pps.AtomInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)

	// Get info about the pipeline pods from k8s & check for resources
	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)

	var container v1.Container
	rcName := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	kubeClient := getKubeClient(t)
	err = backoff.Retry(func() error {
		podList, err := kubeClient.CoreV1().Pods(v1.NamespaceDefault).List(metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
				map[string]string{"app": rcName, "suite": "pachyderm"},
			)),
		})
		if err != nil {
			return err // retry
		}
		if len(podList.Items) != 1 || len(podList.Items[0].Spec.Containers) == 0 {
			return fmt.Errorf("could not find single container for pipeline %s", pipelineInfo.Pipeline.Name)
		}
		container = podList.Items[0].Spec.Containers[0]
		return nil // no more retries
	}, backoff.NewTestingBackOff())
	require.NoError(t, err)
	_, ok := container.Resources.Requests[v1.ResourceNvidiaGPU]
	require.False(t, ok)
}

func TestPipelinePartialResourceRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipelinePartialResourceRequest")
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreateRepo(dataRepo))
	// Resources are not yet in client.CreatePipeline() (we may add them later)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{fmt.Sprintf("%s-%d", pipelineName, 0)},
			Transform: &pps.Transform{
				Cmd: []string{"true"},
			},
			ResourceRequests: &pps.ResourceSpec{
				Cpu:    0.5,
				Memory: "100M",
			},
			Input: &pps.Input{
				Atom: &pps.AtomInput{
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
			Pipeline: &pps.Pipeline{fmt.Sprintf("%s-%d", pipelineName, 1)},
			Transform: &pps.Transform{
				Cmd: []string{"true"},
			},
			ResourceRequests: &pps.ResourceSpec{
				Memory: "100M",
			},
			Input: &pps.Input{
				Atom: &pps.AtomInput{
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
			Pipeline: &pps.Pipeline{fmt.Sprintf("%s-%d", pipelineName, 2)},
			Transform: &pps.Transform{
				Cmd: []string{"true"},
			},
			ResourceRequests: &pps.ResourceSpec{},
			Input: &pps.Input{
				Atom: &pps.AtomInput{
					Repo:   dataRepo,
					Branch: "master",
					Glob:   "/*",
				},
			},
		})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		for i := 0; i < 3; i++ {
			pipelineInfo, err := c.InspectPipeline(fmt.Sprintf("%s-%d", pipelineName, i))
			require.NoError(t, err)
			if pipelineInfo.State != pps.PipelineState_PIPELINE_RUNNING {
				return fmt.Errorf("pipeline not in running state")
			}
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestPipelineLargeOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineInputDataModification_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 100
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(""))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"for i in `seq 1 100`; do touch /pfs/out/$RANDOM; done",
		},
		&pps.ParallelismSpec{
			Constant: 4,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
}

func TestUnionInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	var repos []string
	for i := 0; i < 4; i++ {
		repos = append(repos, uniqueString("TestUnionInput"))
		require.NoError(t, c.CreateRepo(repos[i]))
	}

	numFiles := 2
	var commits []*pfs.Commit
	for _, repo := range repos {
		commit, err := c.StartCommit(repo, "master")
		require.NoError(t, err)
		commits = append(commits, commit)
		for i := 0; i < numFiles; i++ {
			_, err = c.PutFile(repo, "master", fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)))
		}
		require.NoError(t, c.FinishCommit(repo, "master"))
	}

	t.Run("union all", func(t *testing.T) {
		pipeline := uniqueString("pipeline")
		require.NoError(t, c.CreatePipeline(
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
				client.NewAtomInput(repos[0], "/*"),
				client.NewAtomInput(repos[1], "/*"),
				client.NewAtomInput(repos[2], "/*"),
				client.NewAtomInput(repos[3], "/*"),
			),
			"",
			false,
		))

		commitIter, err := c.FlushCommit(commits, []*pfs.Repo{client.NewRepo(pipeline)})
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))
		outCommit := commitInfos[0].Commit
		fileInfos, err := c.ListFile(outCommit.Repo.Name, outCommit.ID, "")
		require.NoError(t, err)
		require.Equal(t, 2, len(fileInfos))
		for _, fi := range fileInfos {
			// 1 byte per repo
			require.Equal(t, uint64(len(repos)), fi.SizeBytes)
		}
	})

	t.Run("union crosses", func(t *testing.T) {
		pipeline := uniqueString("pipeline")
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp -r /pfs/TestUnionInput* /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewUnionInput(
				client.NewCrossInput(
					client.NewAtomInput(repos[0], "/*"),
					client.NewAtomInput(repos[1], "/*"),
				),
				client.NewCrossInput(
					client.NewAtomInput(repos[2], "/*"),
					client.NewAtomInput(repos[3], "/*"),
				),
			),
			"",
			false,
		))

		commitIter, err := c.FlushCommit(commits, []*pfs.Repo{client.NewRepo(pipeline)})
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))
		outCommit := commitInfos[0].Commit
		for _, repo := range repos {
			fileInfos, err := c.ListFile(outCommit.Repo.Name, outCommit.ID, repo)
			require.NoError(t, err)
			require.Equal(t, 2, len(fileInfos))
			for _, fi := range fileInfos {
				// each file should be seen twice
				require.Equal(t, uint64(2), fi.SizeBytes)
			}
		}
	})

	t.Run("cross unions", func(t *testing.T) {
		pipeline := uniqueString("pipeline")
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp -r /pfs/TestUnionInput* /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewCrossInput(
				client.NewUnionInput(
					client.NewAtomInput(repos[0], "/*"),
					client.NewAtomInput(repos[1], "/*"),
				),
				client.NewUnionInput(
					client.NewAtomInput(repos[2], "/*"),
					client.NewAtomInput(repos[3], "/*"),
				),
			),
			"",
			false,
		))

		commitIter, err := c.FlushCommit(commits, []*pfs.Repo{client.NewRepo(pipeline)})
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))
		outCommit := commitInfos[0].Commit
		for _, repo := range repos {
			fileInfos, err := c.ListFile(outCommit.Repo.Name, outCommit.ID, repo)
			require.NoError(t, err)
			require.Equal(t, 2, len(fileInfos))
			for _, fi := range fileInfos {
				// each file should be seen twice
				require.Equal(t, uint64(4), fi.SizeBytes)
			}
		}
	})

	t.Run("union alias", func(t *testing.T) {
		pipeline := uniqueString("pipeline")
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp -r /pfs/in /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewUnionInput(
				client.NewAtomInputOpts("in", repos[0], "", "/*", false),
				client.NewAtomInputOpts("in", repos[1], "", "/*", false),
				client.NewAtomInputOpts("in", repos[2], "", "/*", false),
				client.NewAtomInputOpts("in", repos[3], "", "/*", false),
			),
			"",
			false,
		))

		commitIter, err := c.FlushCommit(commits, []*pfs.Repo{client.NewRepo(pipeline)})
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))
		outCommit := commitInfos[0].Commit
		fileInfos, err := c.ListFile(outCommit.Repo.Name, outCommit.ID, "in")
		require.NoError(t, err)
		require.Equal(t, 2, len(fileInfos))
		for _, fi := range fileInfos {
			require.Equal(t, uint64(4), fi.SizeBytes)
		}
	})

	t.Run("union cross alias", func(t *testing.T) {
		pipeline := uniqueString("pipeline")
		require.YesError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp -r /pfs/in* /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewUnionInput(
				client.NewCrossInput(
					client.NewAtomInputOpts("in1", repos[0], "", "/*", false),
					client.NewAtomInputOpts("in1", repos[1], "", "/*", false),
				),
				client.NewCrossInput(
					client.NewAtomInputOpts("in2", repos[2], "", "/*", false),
					client.NewAtomInputOpts("in2", repos[3], "", "/*", false),
				),
			),
			"",
			false,
		))
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp -r /pfs/in* /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewUnionInput(
				client.NewCrossInput(
					client.NewAtomInputOpts("in1", repos[0], "", "/*", false),
					client.NewAtomInputOpts("in2", repos[1], "", "/*", false),
				),
				client.NewCrossInput(
					client.NewAtomInputOpts("in1", repos[2], "", "/*", false),
					client.NewAtomInputOpts("in2", repos[3], "", "/*", false),
				),
			),
			"",
			false,
		))

		commitIter, err := c.FlushCommit(commits, []*pfs.Repo{client.NewRepo(pipeline)})
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))
		outCommit := commitInfos[0].Commit
		for _, dir := range []string{"in1", "in2"} {
			fileInfos, err := c.ListFile(outCommit.Repo.Name, outCommit.ID, dir)
			require.NoError(t, err)
			require.Equal(t, 2, len(fileInfos))
			for _, fi := range fileInfos {
				// each file should be seen twice
				require.Equal(t, uint64(4), fi.SizeBytes)
			}
		}
	})
	t.Run("cross union alias", func(t *testing.T) {
		pipeline := uniqueString("pipeline")
		require.YesError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp -r /pfs/in* /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewCrossInput(
				client.NewUnionInput(
					client.NewAtomInputOpts("in1", repos[0], "", "/*", false),
					client.NewAtomInputOpts("in2", repos[1], "", "/*", false),
				),
				client.NewUnionInput(
					client.NewAtomInputOpts("in1", repos[2], "", "/*", false),
					client.NewAtomInputOpts("in2", repos[3], "", "/*", false),
				),
			),
			"",
			false,
		))
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				"cp -r /pfs/in* /pfs/out",
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewCrossInput(
				client.NewUnionInput(
					client.NewAtomInputOpts("in1", repos[0], "", "/*", false),
					client.NewAtomInputOpts("in1", repos[1], "", "/*", false),
				),
				client.NewUnionInput(
					client.NewAtomInputOpts("in2", repos[2], "", "/*", false),
					client.NewAtomInputOpts("in2", repos[3], "", "/*", false),
				),
			),
			"",
			false,
		))

		commitIter, err := c.FlushCommit(commits, []*pfs.Repo{client.NewRepo(pipeline)})
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))
		outCommit := commitInfos[0].Commit
		for _, dir := range []string{"in1", "in2"} {
			fileInfos, err := c.ListFile(outCommit.Repo.Name, outCommit.ID, dir)
			require.NoError(t, err)
			require.Equal(t, 2, len(fileInfos))
			for _, fi := range fileInfos {
				// each file should be seen twice
				require.Equal(t, uint64(8), fi.SizeBytes)
			}
		}
	})
}

func TestIncrementalOverwritePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestIncrementalOverwritePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	pipeline := uniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"touch /pfs/out/sum",
					fmt.Sprintf("SUM=`cat /pfs/%s/data /pfs/out/sum | awk '{sum+=$1} END {print sum}'`", dataRepo),
					"echo $SUM > /pfs/out/sum",
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input:       client.NewAtomInput(dataRepo, "/"),
			Incremental: true,
		})
	require.NoError(t, err)
	expectedValue := 0
	for i := 0; i <= 150; i++ {
		_, err := c.StartCommit(dataRepo, "master")
		require.NoError(t, err)
		require.NoError(t, c.DeleteFile(dataRepo, "master", "data"))
		_, err = c.PutFile(dataRepo, "master", "data", strings.NewReader(fmt.Sprintf("%d\n", i)))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(dataRepo, "master"))
		expectedValue += i
	}

	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(pipeline, "master", "sum", 0, 0, &buf))
	require.Equal(t, fmt.Sprintf("%d\n", expectedValue), buf.String())
}

func TestIncrementalAppendPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestIncrementalAppendPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	pipeline := uniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"touch /pfs/out/sum",
					fmt.Sprintf("SUM=`cat /pfs/%s/data/* /pfs/out/sum | awk '{sum+=$1} END {print sum}'`", dataRepo),
					"echo $SUM > /pfs/out/sum",
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input:       client.NewAtomInput(dataRepo, "/"),
			Incremental: true,
		})
	require.NoError(t, err)
	expectedValue := 0
	for i := 0; i <= 150; i++ {
		_, err := c.StartCommit(dataRepo, "master")
		require.NoError(t, err)
		w, err := c.PutFileSplitWriter(dataRepo, "master", "data", pfs.Delimiter_LINE, 0, 0, false)
		require.NoError(t, err)
		_, err = w.Write([]byte(fmt.Sprintf("%d\n", i)))
		require.NoError(t, err)
		require.NoError(t, w.Close())
		require.NoError(t, c.FinishCommit(dataRepo, "master"))
		expectedValue += i
	}

	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(pipeline, "master", "sum", 0, 0, &buf))
	require.Equal(t, fmt.Sprintf("%d\n", expectedValue), buf.String())
}

func TestIncrementalOneFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestIncrementalOneFile")
	require.NoError(t, c.CreateRepo(dataRepo))

	pipeline := uniqueString("pipeline")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"find /pfs",
					fmt.Sprintf("cp /pfs/%s/dir/file /pfs/out/file", dataRepo),
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input:       client.NewAtomInput(dataRepo, "/dir/file"),
			Incremental: true,
		})
	require.NoError(t, err)
	_, err = c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, "master", "/dir/file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, "master"))
	_, err = c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, "master", "/dir/file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, "master"))

	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(pipeline, "master", "file", 0, 0, &buf))
	require.Equal(t, "foo\nbar\n", buf.String())
}

func TestGarbageCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	if os.Getenv(InCloudEnv) == "" {
		t.Skip("Skipping this test as it can only be run in the cloud.")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	// The objects/tags that are there originally.  We run GC
	// first so that later GC runs doesn't collect objects created
	// by other tests.
	require.NoError(t, c.GarbageCollect())
	originalObjects := getAllObjects(t, c)
	originalTags := getAllTags(t, c)

	dataRepo := uniqueString("TestGarbageCollection")
	pipeline := uniqueString("TestGarbageCollectionPipeline")

	var commit *pfs.Commit
	var err error
	createInputAndPipeline := func() {
		require.NoError(t, c.CreateRepo(dataRepo))

		commit, err = c.StartCommit(dataRepo, "master")
		require.NoError(t, err)
		_, err = c.PutFile(dataRepo, commit.ID, "foo", strings.NewReader("foo"))
		require.NoError(t, err)
		_, err = c.PutFile(dataRepo, commit.ID, "bar", strings.NewReader("bar"))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(dataRepo, "master"))

		// This pipeline copies foo and modifies bar
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				fmt.Sprintf("cp /pfs/%s/foo /pfs/out/foo", dataRepo),
				fmt.Sprintf("cp /pfs/%s/bar /pfs/out/bar", dataRepo),
				"echo bar >> /pfs/out/bar",
			},
			nil,
			client.NewAtomInput(dataRepo, "/"),
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))
	}
	createInputAndPipeline()

	objectsBefore := getAllObjects(t, c)
	tagsBefore := getAllTags(t, c)

	// Now delete the output repo and GC
	require.NoError(t, c.DeleteRepo(pipeline, false))
	require.NoError(t, c.GarbageCollect())

	// Check that data still exists in the input repo
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(dataRepo, commit.ID, "foo", 0, 0, &buf))
	require.Equal(t, "foo", buf.String())
	buf.Reset()
	require.NoError(t, c.GetFile(dataRepo, commit.ID, "bar", 0, 0, &buf))
	require.Equal(t, "bar", buf.String())

	// Check that the objects that should be removed have been removed.
	// We should've deleted precisely one object: the tree for the output
	// commit.
	objectsAfter := getAllObjects(t, c)
	tagsAfter := getAllTags(t, c)

	require.Equal(t, len(tagsBefore), len(tagsAfter))
	require.Equal(t, len(objectsBefore), len(objectsAfter)+1)
	objectsBefore = objectsAfter
	tagsBefore = tagsAfter

	// Now delete the pipeline and GC
	require.NoError(t, c.DeletePipeline(pipeline, false))
	require.NoError(t, c.GarbageCollect())

	// We should've deleted one tag since the pipeline has only processed
	// one datum.
	// We should've deleted two objects: one is the object referenced by
	// the tag, and another is the modified "bar" file.
	objectsAfter = getAllObjects(t, c)
	tagsAfter = getAllTags(t, c)

	require.Equal(t, len(tagsBefore), len(tagsAfter)+1)
	require.Equal(t, len(objectsBefore), len(objectsAfter)+2)

	// Now we delete the input repo.
	require.NoError(t, c.DeleteRepo(dataRepo, false))
	require.NoError(t, c.GarbageCollect())

	// Since we've now deleted everything that we created in this test,
	// the tag count and object count should be back to the originals.
	objectsAfter = getAllObjects(t, c)
	tagsAfter = getAllTags(t, c)
	require.Equal(t, len(originalTags), len(tagsAfter))
	require.Equal(t, len(originalObjects), len(objectsAfter))

	// Now we create the pipeline again and check that all data is
	// accessible.  This is important because there used to be a bug
	// where we failed to invalidate the cache such that the objects in
	// the cache were referencing blocks that had been GC-ed.
	createInputAndPipeline()
	buf.Reset()
	require.NoError(t, c.GetFile(dataRepo, commit.ID, "foo", 0, 0, &buf))
	require.Equal(t, "foo", buf.String())
	buf.Reset()
	require.NoError(t, c.GetFile(dataRepo, commit.ID, "bar", 0, 0, &buf))
	require.Equal(t, "bar", buf.String())
	buf.Reset()
	require.NoError(t, c.GetFile(pipeline, "master", "foo", 0, 0, &buf))
	require.Equal(t, "foo", buf.String())
	buf.Reset()
	require.NoError(t, c.GetFile(pipeline, "master", "bar", 0, 0, &buf))
	require.Equal(t, "barbar\n", buf.String())
}

func TestPipelineWithStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineWithStats_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 500
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(strings.Repeat("foo\n", 100)))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:       client.NewAtomInput(dataRepo, "/*"),
			EnableStats: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 4,
			},
		})

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	jobs, err := c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	// Check we can list datums before job completion
	resp, err := c.ListDatum(jobs[0].Job.ID, 0, 0)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(resp.DatumInfos))
	require.Equal(t, 1, len(resp.DatumInfos[0].Data))

	// Check we can list datums before job completion w pagination
	resp, err = c.ListDatum(jobs[0].Job.ID, 100, 0)
	require.NoError(t, err)
	require.Equal(t, 100, len(resp.DatumInfos))
	require.Equal(t, int64(numFiles/100), resp.TotalPages)
	require.Equal(t, int64(0), resp.Page)

	// Block on the job being complete before we call ListDatum again so we're
	// sure the datums have actually been processed.
	_, err = c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)

	resp, err = c.ListDatum(jobs[0].Job.ID, 0, 0)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(resp.DatumInfos))
	require.Equal(t, 1, len(resp.DatumInfos[0].Data))

	for _, datum := range resp.DatumInfos {
		require.NoError(t, err)
		require.Equal(t, pps.DatumState_SUCCESS, datum.State)
	}

	// Make sure inspect-datum works
	datum, err := c.InspectDatum(jobs[0].Job.ID, resp.DatumInfos[0].Datum.ID)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_SUCCESS, datum.State)
}

func TestPipelineWithStatsFailedDatums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineWithStatsFailedDatums_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 200
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(strings.Repeat("foo\n", 100)))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"if [ $RANDOM -gt 15000 ]; then exit 1; fi",
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:       client.NewAtomInput(dataRepo, "/*"),
			EnableStats: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 4,
			},
		})

	_, err = c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)

	// Without this sleep, I get no results from list-job
	// See issue: https://github.com/pachyderm/pachyderm/issues/2181
	time.Sleep(15 * time.Second)
	jobs, err := c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	// Block on the job being complete before we call ListDatum
	_, err = c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)

	resp, err := c.ListDatum(jobs[0].Job.ID, 0, 0)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(resp.DatumInfos))

	// First entry should be failed
	require.Equal(t, pps.DatumState_FAILED, resp.DatumInfos[0].State)
	// Last entry should be success
	require.Equal(t, pps.DatumState_SUCCESS, resp.DatumInfos[len(resp.DatumInfos)-1].State)

	// Make sure inspect-datum works for failed state
	datum, err := c.InspectDatum(jobs[0].Job.ID, resp.DatumInfos[0].Datum.ID)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_FAILED, datum.State)
}

func TestPipelineWithStatsPaginated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineWithStatsPaginated_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numPages := int64(2)
	pageSize := int64(100)
	numFiles := int(numPages * pageSize)
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(strings.Repeat("foo\n", 100)))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"if [ $RANDOM -gt 15000 ]; then exit 1; fi",
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:       client.NewAtomInput(dataRepo, "/*"),
			EnableStats: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 4,
			},
		})

	_, err = c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)

	var jobs []*pps.JobInfo
	require.NoError(t, backoff.Retry(func() error {
		jobs, err = c.ListJob(pipeline, nil, nil)
		require.NoError(t, err)
		if len(jobs) != 1 {
			return fmt.Errorf("expected 1 jobs, got %d", len(jobs))
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Block on the job being complete before we call ListDatum
	_, err = c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)

	resp, err := c.ListDatum(jobs[0].Job.ID, pageSize, 0)
	require.NoError(t, err)
	require.Equal(t, pageSize, int64(len(resp.DatumInfos)))
	require.Equal(t, int64(numFiles)/pageSize, resp.TotalPages)

	// First entry should be failed
	require.Equal(t, pps.DatumState_FAILED, resp.DatumInfos[0].State)

	resp, err = c.ListDatum(jobs[0].Job.ID, pageSize, int64(numPages-1))
	require.NoError(t, err)
	require.Equal(t, pageSize, int64(len(resp.DatumInfos)))
	require.Equal(t, int64(int64(numFiles)/pageSize-1), resp.Page)

	// Last entry should be success
	require.Equal(t, pps.DatumState_SUCCESS, resp.DatumInfos[len(resp.DatumInfos)-1].State)

	// Make sure we get error when requesting pages too high
	resp, err = c.ListDatum(jobs[0].Job.ID, pageSize, int64(numPages))
	require.YesError(t, err)
}

func TestPipelineWithStatsAcrossJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineWithStatsAcrossJobs_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 500
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("foo-%d", i), strings.NewReader(strings.Repeat("foo\n", 100)))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("StatsAcrossJobs")
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:       client.NewAtomInput(dataRepo, "/*"),
			EnableStats: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 4,
			},
		})

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	jobs, err := c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	// Block on the job being complete before we call ListDatum
	_, err = c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)

	resp, err := c.ListDatum(jobs[0].Job.ID, 0, 0)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(resp.DatumInfos))

	datum, err := c.InspectDatum(jobs[0].Job.ID, resp.DatumInfos[0].Datum.ID)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_SUCCESS, datum.State)

	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit2.ID, fmt.Sprintf("bar-%d", i), strings.NewReader(strings.Repeat("bar\n", 100)))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commit2}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	jobs, err = c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))

	// Block on the job being complete before we call ListDatum
	_, err = c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)

	resp, err = c.ListDatum(jobs[0].Job.ID, 0, 0)
	require.NoError(t, err)
	// we should see all the datums from the first job (which should be skipped)
	// in addition to all the new datums processed in this job
	require.Equal(t, numFiles*2, len(resp.DatumInfos))

	datum, err = c.InspectDatum(jobs[0].Job.ID, resp.DatumInfos[0].Datum.ID)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_SUCCESS, datum.State)
	// Test datums marked as skipped correctly
	// (also tests list datums are sorted by state)
	datum, err = c.InspectDatum(jobs[0].Job.ID, resp.DatumInfos[numFiles].Datum.ID)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_SKIPPED, datum.State)
}

func TestPipelineWithStatsSkippedEdgeCase(t *testing.T) {
	// If I add a file in commit1, delete it in commit2, add it again in commit 3 ...
	// the datum will be marked as success on the 3rd job, even though it should be marked as skipped
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineWithStatsSkippedEdgeCase_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 10
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(strings.Repeat("foo\n", 100)))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("StatsEdgeCase")
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:       client.NewAtomInput(dataRepo, "/*"),
			EnableStats: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 4,
			},
		})

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	jobs, err := c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	// Block on the job being complete before we call ListDatum
	_, err = c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)
	resp, err := c.ListDatum(jobs[0].Job.ID, 0, 0)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(resp.DatumInfos))

	for _, datum := range resp.DatumInfos {
		require.NoError(t, err)
		require.Equal(t, pps.DatumState_SUCCESS, datum.State)
	}

	// Make sure inspect-datum works
	datum, err := c.InspectDatum(jobs[0].Job.ID, resp.DatumInfos[0].Datum.ID)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_SUCCESS, datum.State)

	// Create a second commit that deletes a file in commit1
	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	err = c.DeleteFile(dataRepo, commit2.ID, "file-0")
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))

	// Create a third commit that re-adds the file removed in commit2
	commit3, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit3.ID, "file-0", strings.NewReader(strings.Repeat("foo\n", 100)))
	require.NoError(t, c.FinishCommit(dataRepo, commit3.ID))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commit3}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	jobs, err = c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(jobs))

	// Block on the job being complete before we call ListDatum
	_, err = c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)
	resp, err = c.ListDatum(jobs[0].Job.ID, 0, 0)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(resp.DatumInfos))

	var states []interface{}
	for _, datum := range resp.DatumInfos {
		require.Equal(t, pps.DatumState_SKIPPED, datum.State)
		states = append(states, datum.State)
	}
}

func TestIncrementalSharedProvenance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestIncrementalSharedProvenance_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	pipeline1 := uniqueString("pipeline1")
	require.NoError(t, c.CreatePipeline(
		pipeline1,
		"",
		[]string{"true"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/"),
		"",
		false,
	))
	pipeline2 := uniqueString("pipeline2")
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline2),
			Transform: &pps.Transform{
				Cmd: []string{"true"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: client.NewCrossInput(
				client.NewAtomInput(dataRepo, "/"),
				client.NewAtomInput(pipeline1, "/"),
			),
			Incremental: true,
		})
	require.YesError(t, err)
	pipeline3 := uniqueString("pipeline3")
	require.NoError(t, c.CreatePipeline(
		pipeline3,
		"",
		[]string{"true"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/"),
		"",
		false,
	))
	pipeline4 := uniqueString("pipeline4")
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline4),
			Transform: &pps.Transform{
				Cmd: []string{"true"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input: client.NewCrossInput(
				client.NewAtomInput(pipeline1, "/"),
				client.NewAtomInput(pipeline3, "/"),
			),
			Incremental: true,
		})
	require.YesError(t, err)
}

func TestSkippedDatums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	//	require.NoError(t, c.CreatePipeline(
	_, err := c.PpsAPIClient.CreatePipeline(context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input:       client.NewAtomInput(dataRepo, "/*"),
			EnableStats: true,
		})
	require.NoError(t, err)
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	commitInfoIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit1.ID)}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitInfoIter)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	// Do second commit to repo
	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit2.ID, "file2", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))
	commitInfoIter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitInfoIter)
	require.Equal(t, 1, len(commitInfos))
	/*
		jobs, err := c.ListJob(pipelineName, nil, nil)
		require.NoError(t, err)
		require.Equal(t, 2, len(jobs))

		datums, err := c.ListDatum(jobs[1].Job.ID)
		fmt.Printf("got datums: %v\n", datums)
		require.NoError(t, err)
		require.Equal(t, 2, len(datums))

		datum, err := c.InspectDatum(jobs[1].Job.ID, datums[0].ID)
		require.NoError(t, err)
		require.Equal(t, pps.DatumState_SUCCESS, datum.State)
	*/
}

func TestOpencvDemo(t *testing.T) {
	t.Skip("flaky")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	require.NoError(t, c.CreateRepo("images"))
	commit, err := c.StartCommit("images", "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFileURL("images", "master", "46Q8nDz.jpg", "http://imgur.com/46Q8nDz.jpg", false, false))
	require.NoError(t, c.FinishCommit("images", "master"))
	bytes, err := ioutil.ReadFile("../../doc/examples/opencv/edges.json")
	require.NoError(t, err)
	createPipelineRequest := &pps.CreatePipelineRequest{}
	require.NoError(t, json.Unmarshal(bytes, createPipelineRequest))
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(), createPipelineRequest)
	require.NoError(t, err)
	bytes, err = ioutil.ReadFile("../../doc/examples/opencv/montage.json")
	require.NoError(t, err)
	createPipelineRequest = &pps.CreatePipelineRequest{}
	require.NoError(t, json.Unmarshal(bytes, createPipelineRequest))
	_, err = c.PpsAPIClient.CreatePipeline(context.Background(), createPipelineRequest)
	require.NoError(t, err)
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 2, len(commitInfos))
}

func TestCronPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	pipeline1 := uniqueString("cron1-")
	require.NoError(t, c.CreatePipeline(
		pipeline1,
		"",
		[]string{"cp", "/pfs/time/time", "/pfs/out/time"},
		nil,
		nil,
		client.NewCronInput("time", "@every 20s"),
		"",
		false,
	))
	pipeline2 := uniqueString("cron2-")
	require.NoError(t, c.CreatePipeline(
		pipeline2,
		"",
		[]string{"cp", fmt.Sprintf("/pfs/%s/time", pipeline1), "/pfs/out/time"},
		nil,
		nil,
		client.NewAtomInput(pipeline1, "/*"),
		"",
		false,
	))

	// subscribe to the pipeline1 cron repo and wait for inputs
	repo := fmt.Sprintf("%s_%s", pipeline1, "time")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel() //cleanup resources
	iter, err := c.WithCtx(ctx).SubscribeCommit(repo, "master", "")
	require.NoError(t, err)
	commitInfo, err := iter.Next()
	require.NoError(t, err)

	commitIter, err := c.FlushCommit([]*pfs.Commit{commitInfo.Commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 2, len(commitInfos))

	// Create a non-cron input repo, and test a pipeline with a cross of cron and
	// non-cron inputs
	dataRepo := uniqueString("TestCronPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	pipeline3 := uniqueString("cron3-")
	require.NoError(t, c.CreatePipeline(
		pipeline3,
		"",
		[]string{"bash"},
		[]string{
			"cp /pfs/time/time /pfs/out/time",
			fmt.Sprintf("cp /pfs/%s/file /pfs/out/file", dataRepo),
		},
		nil,
		client.NewCrossInput(
			client.NewCronInput("time", "@every 20s"),
			client.NewAtomInput(dataRepo, "/"),
		),
		"",
		false,
	))
	dataCommit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, "master", "file", strings.NewReader("file"))
	require.NoError(t, c.FinishCommit(dataRepo, "master"))

	repo = fmt.Sprintf("%s_%s", pipeline3, "time")
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	iter, err = c.WithCtx(ctx).SubscribeCommit(repo, "master", "")
	require.NoError(t, err)
	commitInfo, err = iter.Next()
	require.NoError(t, err)

	commitIter, err = c.FlushCommit([]*pfs.Commit{dataCommit, commitInfo.Commit}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
}

func TestSelfReferentialPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	pipeline := uniqueString("pipeline")
	require.YesError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"true"},
		nil,
		nil,
		client.NewAtomInput(pipeline, "/"),
		"",
		false,
	))
}

func TestPipelineBadImage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	pipeline1 := uniqueString("bad_pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline1,
		"BadImage",
		[]string{"true"},
		nil,
		nil,
		client.NewCronInput("time", "@every 20s"),
		"",
		false,
	))
	pipeline2 := uniqueString("bad_pipeline")
	require.NoError(t, c.CreatePipeline(
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
			pipelineInfo, err := c.InspectPipeline(pipeline)
			if err != nil {
				return err
			}
			if pipelineInfo.State != pps.PipelineState_PIPELINE_FAILURE {
				return fmt.Errorf("pipeline %s should have failed", pipeline)
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

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	// create repos
	dataRepo := uniqueString("TestFixPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	_, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, "master", "file", strings.NewReader("1"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, "master"))
	pipelineName := uniqueString("TestFixPipeline_pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"exit 1"},
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob(pipelineName, nil, nil)
		require.NoError(t, err)
		if len(jobInfos) != 1 {
			return fmt.Errorf("expected 1 jobs, got %d", len(jobInfos))
		}
		jobInfo, err := c.InspectJob(jobInfos[0].Job.ID, true)
		require.NoError(t, err)
		require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)
		return nil
	}, backoff.NewTestingBackOff()))

	// Update the pipeline, this will not create a new pipeline as reprocess
	// isn't set to true.
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"echo bar >/pfs/out/file"},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		true,
	))

	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob(pipelineName, nil, nil)
		require.NoError(t, err)
		if len(jobInfos) != 2 {
			return fmt.Errorf("expected 2 jobs, got %d", len(jobInfos))
		}
		jobInfo, err := c.InspectJob(jobInfos[0].Job.ID, true)
		require.NoError(t, err)
		require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestListJobOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := getPachClient(t)

	dataRepo := uniqueString("TestListJobOutput_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 4,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	require.NoError(t, backoff.Retry(func() error {
		jobInfos, err := c.ListJob("", nil, commitInfos[0].Commit)
		if err != nil {
			return err
		}
		if len(jobInfos) != 1 {
			return fmt.Errorf("expected 1 job")
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestPipelineEnvVarAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineEnvVarAlias_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 10
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"env",
			fmt.Sprintf("cp $%s /pfs/out/", dataRepo),
		},
		nil,
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, fmt.Sprintf("file-%d", i), 0, 0, &buf))
		require.Equal(t, fmt.Sprintf("%d", i), buf.String())
	}
}

func TestMaxQueueSize(t *testing.T) {
	t.Skip("flaky")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := getPachClient(t)

	dataRepo := uniqueString("TestMaxQueueSize_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < 20; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file%d", i), strings.NewReader("foo"))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipeline")
	// This pipeline sleeps for 10 secs per datum
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"sleep 10",
				},
			},
			Input: client.NewAtomInput(dataRepo, "/*"),
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 2,
			},
			MaxQueueSize: 1,
		})
	require.NoError(t, err)
	// Get job info 2x/sec for 20s until we confirm two workers for the current job
	require.NoError(t, backoff.Retry(func() error {
		jobs, err := c.ListJob(pipeline, nil, nil)
		if err != nil {
			return fmt.Errorf("could not list job: %s", err.Error())
		}
		if len(jobs) == 0 {
			return fmt.Errorf("failed to find job")
		}
		jobInfo, err := c.InspectJob(jobs[0].Job.ID, false)
		if err != nil {
			return fmt.Errorf("could not inspect job: %s", err.Error())
		}
		if len(jobInfo.WorkerStatus) != 2 {
			return fmt.Errorf("incorrect number of statuses: %v", len(jobInfo.WorkerStatus))
		}
		for _, status := range jobInfo.WorkerStatus {
			if status.QueueSize > 1 {
				return fmt.Errorf("queue size too big")
			}
		}
		return nil
	}, backoff.RetryEvery(500*time.Millisecond).For(20*time.Second)))
}

func TestHTTPAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := getPachClient(t)

	clientAddr := c.GetAddress()
	host, _, err := net.SplitHostPort(clientAddr)
	port, ok := os.LookupEnv("PACHD_SERVICE_PORT_API_HTTP_PORT")
	if !ok {
		port = "30652" // default NodePort port for Pachd's HTTP API
	}
	httpAPIAddr := net.JoinHostPort(host, port)

	// Try to login
	token := "abbazabbadoo"
	form := url.Values{}
	form.Add("Token", token)
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/v1/auth/login", httpAPIAddr), strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, err)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 1, len(resp.Cookies()))
	require.Equal(t, auth.ContextTokenKey, resp.Cookies()[0].Name)
	require.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	require.Equal(t, token, resp.Cookies()[0].Value)

	// Try to logout
	req, err = http.NewRequest("POST", fmt.Sprintf("http://%s/v1/auth/logout", httpAPIAddr), nil)
	require.NoError(t, err)
	resp, err = httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 1, len(resp.Cookies()))
	require.Equal(t, auth.ContextTokenKey, resp.Cookies()[0].Name)
	require.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	// The cookie should be unset now
	require.Equal(t, "", resp.Cookies()[0].Value)

	// Make sure we get 404s for non existent routes
	req, err = http.NewRequest("POST", fmt.Sprintf("http://%s/v1/auth/logoutzz", httpAPIAddr), nil)
	require.NoError(t, err)
	resp, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
}

func TestHTTPGetFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := getPachClient(t)

	dataRepo := uniqueString("TestHTTPGetFile_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	f, err := os.Open("../../etc/testing/artifacts/giphy.gif")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "giphy.gif", f)
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	clientAddr := c.GetAddress()
	host, _, err := net.SplitHostPort(clientAddr)
	port, ok := os.LookupEnv("PACHD_SERVICE_PORT_API_HTTP_PORT")
	if !ok {
		port = "30652" // default NodePort port for Pachd's HTTP API
	}
	httpAPIAddr := net.JoinHostPort(host, port)

	// Try to get raw contents
	resp, err := http.Get(fmt.Sprintf("http://%s/v1/pfs/repos/%v/commits/%v/files/file", httpAPIAddr, dataRepo, commit1.ID))
	require.NoError(t, err)
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "foo", string(contents))
	contentDisposition := resp.Header.Get("Content-Disposition")
	require.Equal(t, "", contentDisposition)

	// Try to get file for downloading
	resp, err = http.Get(fmt.Sprintf("http://%s/v1/pfs/repos/%v/commits/%v/files/file?download=true", httpAPIAddr, dataRepo, commit1.ID))
	require.NoError(t, err)
	defer resp.Body.Close()
	contents, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "foo", string(contents))
	contentDisposition = resp.Header.Get("Content-Disposition")
	require.Equal(t, "attachment; filename=\"file\"", contentDisposition)

	// Make sure MIME type is set
	resp, err = http.Get(fmt.Sprintf("http://%s/v1/pfs/repos/%v/commits/%v/files/giphy.gif", httpAPIAddr, dataRepo, commit1.ID))
	require.NoError(t, err)
	defer resp.Body.Close()
	contentDisposition = resp.Header.Get("Content-Type")
	require.Equal(t, "image/gif", contentDisposition)
}

func TestService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := getPachClient(t)

	dataRepo := uniqueString("TestService_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	_, err = c.PutFile(dataRepo, commit1.ID, "file1", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("pipelineservice")
	// This pipeline sleeps for 10 secs per datum
	require.NoError(t, c.CreatePipelineService(
		pipeline,
		"trinitronx/python-simplehttpserver",
		[]string{"sh"},
		[]string{
			"cd /pfs",
			"exec python -m SimpleHTTPServer 8000",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/"),
		false,
		8000,
		31800,
	))
	time.Sleep(10 * time.Second)

	// Lookup the address for 'pipelineservice' (different inside vs outside k8s)
	serviceAddr := func() string {
		// Hack: detect if running inside the cluster by looking for this env var
		if _, ok := os.LookupEnv("KUBERNETES_PORT"); !ok {
			// Outside cluster: Re-use external IP and external port defined above
			clientAddr := c.GetAddress()
			host, _, err := net.SplitHostPort(clientAddr)
			require.NoError(t, err)
			return net.JoinHostPort(host, "31800")
		}
		// Get k8s service corresponding to pachyderm service above--must access
		// via internal cluster IP, but we don't know what that is
		var address string
		kubeClient := getKubeClient(t)
		backoff.Retry(func() error {
			svcs, err := kubeClient.CoreV1().Services("default").List(metav1.ListOptions{})
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
			return fmt.Errorf("no matching k8s service found")
		}, backoff.NewTestingBackOff())

		require.NotEqual(t, "", address)
		return address
	}()

	require.NoError(t, backoff.Retry(func() error {
		resp, err := http.Get(fmt.Sprintf("http://%s/%s/file1", serviceAddr, dataRepo))
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("GET returned %d", resp.StatusCode)
		}
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if string(content) != "foo" {
			return fmt.Errorf("wrong content for file1: expected foo, got %s", string(content))
		}
		return nil
	}, backoff.NewTestingBackOff()))

	commit2, err := c.StartCommit(dataRepo, "master")
	_, err = c.PutFile(dataRepo, commit2.ID, "file2", strings.NewReader("bar"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))

	require.NoError(t, backoff.Retry(func() error {
		resp, err := http.Get(fmt.Sprintf("http://%s/%s/file2", serviceAddr, dataRepo))
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("GET returned %d", resp.StatusCode)
		}
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if string(content) != "bar" {
			return fmt.Errorf("wrong content for file2: expected bar, got %s", string(content))
		}
		return nil
	}, backoff.NewTestingBackOff()))
}

func TestChunkSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestChunkSpec_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	numFiles := 101
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file%d", i), strings.NewReader("foo"))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	t.Run("number", func(t *testing.T) {
		pipeline := uniqueString("TestChunkSpec")
		c.PpsAPIClient.CreatePipeline(context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"bash"},
					Stdin: []string{
						fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
					},
				},
				Input:     client.NewAtomInput(dataRepo, "/*"),
				ChunkSpec: &pps.ChunkSpec{Number: 1},
			})

		commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, []*pfs.Repo{client.NewRepo(pipeline)})
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))

		for i := 0; i < numFiles; i++ {
			var buf bytes.Buffer
			require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, fmt.Sprintf("file%d", i), 0, 0, &buf))
			require.Equal(t, "foo", buf.String())
		}
	})
	t.Run("size", func(t *testing.T) {
		pipeline := uniqueString("TestChunkSpec")
		c.PpsAPIClient.CreatePipeline(context.Background(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"bash"},
					Stdin: []string{
						fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
					},
				},
				Input:     client.NewAtomInput(dataRepo, "/*"),
				ChunkSpec: &pps.ChunkSpec{SizeBytes: 5},
			})

		commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, []*pfs.Repo{client.NewRepo(pipeline)})
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))

		for i := 0; i < numFiles; i++ {
			var buf bytes.Buffer
			require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, fmt.Sprintf("file%d", i), 0, 0, &buf))
			require.Equal(t, "foo", buf.String())
		}
	})
}

func TestLongDatums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestLongDatums_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	numFiles := 8
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file%d", i), strings.NewReader("foo"))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := uniqueString("TestLongDatums")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"sleep 1m",
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 4,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, fmt.Sprintf("file%d", i), 0, 0, &buf))
		require.Equal(t, "foo", buf.String())
	}
}

func TestPipelineWithGitInputInvalidURLs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	outputFilename := "commitSHA"
	pipeline := uniqueString("github_pipeline")
	// Of the common git URL types (listed below), only the 'clone' url is supported RN
	// (for several reasons, one of which is that we can't assume we have SSH / an ssh env setup on the user container)
	//git_url: "git://github.com/sjezewski/testgithook.git",
	//ssh_url: "git@github.com:sjezewski/testgithook.git",
	//svn_url: "https://github.com/sjezewski/testgithook",
	//clone_url: "https://github.com/sjezewski/testgithook.git",
	require.YesError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/test-artifacts/.git/HEAD > /pfs/out/%v", outputFilename),
		},
		nil,
		&pps.Input{
			Git: &pps.GitInput{
				URL: "git://github.com/pachyderm/test-artifacts.git",
			},
		},
		"",
		false,
	))
	require.YesError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/test-artifacts/.git/HEAD > /pfs/out/%v", outputFilename),
		},
		nil,
		&pps.Input{
			Git: &pps.GitInput{
				URL: "git@github.com:pachyderm/test-artifacts.git",
			},
		},
		"",
		false,
	))
	require.YesError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/test-artifacts/.git/HEAD > /pfs/out/%v", outputFilename),
		},
		nil,
		&pps.Input{
			Git: &pps.GitInput{
				URL: "https://github.com:pachyderm/test-artifacts",
			},
		},
		"",
		false,
	))
}

func TestPipelineWithGitInputPrivateGHRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	outputFilename := "commitSHA"
	pipeline := uniqueString("github_pipeline")
	repoName := "pachyderm-dummy"
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/%v/.git/HEAD > /pfs/out/%v", repoName, outputFilename),
		},
		nil,
		&pps.Input{
			Git: &pps.GitInput{
				URL: fmt.Sprintf("https://github.com/pachyderm/%v.git", repoName),
			},
		},
		"",
		false,
	))
	// There should be a pachyderm repo created w no commits:
	repos, err := c.ListRepo(nil)
	require.NoError(t, err)
	found := false
	for _, repo := range repos {
		if repo.Repo.Name == repoName {
			found = true
		}
	}
	require.Equal(t, true, found)

	commits, err := c.ListCommit(repoName, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 0, len(commits))

	// To trigger the pipeline, we'll need to simulate the webhook by pushing a POST payload to the githook server
	simulateGitPush(t, "../../etc/testing/artifacts/githook-payloads/private.json")
	// Need to sleep since the webhook http handler is non blocking
	time.Sleep(2 * time.Second)

	// Now there should NOT be a new commit on the pachyderm repo
	branches, err := c.ListBranch(repoName)
	require.NoError(t, err)
	require.Equal(t, 0, len(branches))

	// We should see that the pipeline has failed
	pipelineInfo, err := c.InspectPipeline(pipeline)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_FAILURE, pipelineInfo.State)
	require.Equal(t, fmt.Sprintf("unable to clone private github repo (https://github.com/pachyderm/%v.git)", repoName), pipelineInfo.Reason)
}

func TestPipelineWithGitInputDuplicateNames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	outputFilename := "commitSHA"
	pipeline := uniqueString("github_pipeline")
	//Test same name on one pipeline
	require.YesError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/pachyderm/.git/HEAD > /pfs/out/%v", outputFilename),
		},
		nil,
		&pps.Input{
			Cross: []*pps.Input{
				&pps.Input{
					Git: &pps.GitInput{
						URL:  "https://github.com/pachyderm/test-artifacts.git",
						Name: "foo",
					},
				},
				&pps.Input{
					Git: &pps.GitInput{
						URL:  "https://github.com/pachyderm/test-artifacts.git",
						Name: "foo",
					},
				},
			},
		},
		"",
		false,
	))
	//Test same URL on one pipeline
	require.YesError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/pachyderm/.git/HEAD > /pfs/out/%v", outputFilename),
		},
		nil,
		&pps.Input{
			Cross: []*pps.Input{
				&pps.Input{
					Git: &pps.GitInput{
						URL: "https://github.com/pachyderm/test-artifacts.git",
					},
				},
				&pps.Input{
					Git: &pps.GitInput{
						URL: "https://github.com/pachyderm/test-artifacts.git",
					},
				},
			},
		},
		"",
		false,
	))
	// Test same URL but different names
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/pachyderm/.git/HEAD > /pfs/out/%v", outputFilename),
		},
		nil,
		&pps.Input{
			Cross: []*pps.Input{
				&pps.Input{
					Git: &pps.GitInput{
						URL:  "https://github.com/pachyderm/test-artifacts.git",
						Name: "foo",
					},
				},
				&pps.Input{
					Git: &pps.GitInput{
						URL: "https://github.com/pachyderm/test-artifacts.git",
					},
				},
			},
		},
		"",
		false,
	))
}

func TestPipelineWithGitInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	outputFilename := "commitSHA"
	pipeline := uniqueString("github_pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/test-artifacts/.git/HEAD > /pfs/out/%v", outputFilename),
		},
		nil,
		&pps.Input{
			Git: &pps.GitInput{
				URL: "https://github.com/pachyderm/test-artifacts.git",
			},
		},
		"",
		false,
	))
	// There should be a pachyderm repo created w no commits:
	repos, err := c.ListRepo(nil)
	require.NoError(t, err)
	found := false
	newRepoName := "test-artifacts"
	for _, repo := range repos {
		if repo.Repo.Name == newRepoName {
			found = true
		}
	}
	require.Equal(t, true, found)

	commits, err := c.ListCommit(newRepoName, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 0, len(commits))

	// To trigger the pipeline, we'll need to simulate the webhook by pushing a POST payload to the githook server
	simulateGitPush(t, "../../etc/testing/artifacts/githook-payloads/master.json")
	// Need to sleep since the webhook http handler is non blocking
	time.Sleep(2 * time.Second)

	// Now there should be a new commit on the pachyderm repo / master branch
	branches, err := c.ListBranch(newRepoName)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0].Name)
	commit := branches[0].Head

	// Now wait for the pipeline complete as normal
	outputRepo := &pfs.Repo{Name: pipeline}
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, []*pfs.Repo{outputRepo})
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	commit = commitInfos[0].Commit

	var buf bytes.Buffer

	require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, outputFilename, 0, 0, &buf))
	require.Equal(t, "9047fbfc251e7412ef3300868f743f2c24852539", strings.TrimSpace(buf.String()))
}

func TestPipelineWithGitInputSequentialPushes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	outputFilename := "commitSHA"
	pipeline := uniqueString("github_pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/test-artifacts/.git/HEAD > /pfs/out/%v", outputFilename),
		},
		nil,
		&pps.Input{
			Git: &pps.GitInput{
				URL: "https://github.com/pachyderm/test-artifacts.git",
			},
		},
		"",
		false,
	))
	// There should be a pachyderm repo created w no commits:
	repos, err := c.ListRepo(nil)
	require.NoError(t, err)
	found := false
	newRepoName := "test-artifacts"
	for _, repo := range repos {
		if repo.Repo.Name == newRepoName {
			found = true
		}
	}
	require.Equal(t, true, found)

	commits, err := c.ListCommit(newRepoName, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 0, len(commits))

	// To trigger the pipeline, we'll need to simulate the webhook by pushing a POST payload to the githook server
	simulateGitPush(t, "../../etc/testing/artifacts/githook-payloads/master.json")
	// Need to sleep since the webhook http handler is non blocking
	time.Sleep(2 * time.Second)

	// Now there should be a new commit on the pachyderm repo / master branch
	branches, err := c.ListBranch(newRepoName)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0].Name)
	commit := branches[0].Head

	// Now wait for the pipeline complete as normal
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	commit = commitInfos[0].Commit

	var buf bytes.Buffer

	require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, outputFilename, 0, 0, &buf))
	require.Equal(t, "9047fbfc251e7412ef3300868f743f2c24852539", strings.TrimSpace(buf.String()))

	// To trigger the pipeline, we'll need to simulate the webhook by pushing a POST payload to the githook server
	simulateGitPush(t, "../../etc/testing/artifacts/githook-payloads/master-2.json")
	// Need to sleep since the webhook http handler is non blocking
	time.Sleep(2 * time.Second)

	// Now there should be a new commit on the pachyderm repo / master branch
	branches, err = c.ListBranch(newRepoName)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0].Name)
	commit = branches[0].Head

	// Now wait for the pipeline complete as normal
	commitIter, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	commit = commitInfos[0].Commit

	buf.Reset()
	require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, outputFilename, 0, 0, &buf))
	require.Equal(t, "162963b4adf00cd378488abdedc085ba08e21674", strings.TrimSpace(buf.String()))
}

func TestPipelineWithGitInputCustomName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	outputFilename := "commitSHA"
	pipeline := uniqueString("github_pipeline")
	repoName := "foo"
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/%v/.git/HEAD > /pfs/out/%v", repoName, outputFilename),
		},
		nil,
		&pps.Input{
			Git: &pps.GitInput{
				URL:  "https://github.com/pachyderm/test-artifacts.git",
				Name: repoName,
			},
		},
		"",
		false,
	))
	// There should be a pachyderm repo created w no commits:
	repos, err := c.ListRepo(nil)
	require.NoError(t, err)
	found := false
	for _, repo := range repos {
		if repo.Repo.Name == repoName {
			found = true
		}
	}
	require.Equal(t, true, found)

	commits, err := c.ListCommit(repoName, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 0, len(commits))

	// To trigger the pipeline, we'll need to simulate the webhook by pushing a POST payload to the githook server
	simulateGitPush(t, "../../etc/testing/artifacts/githook-payloads/master.json")
	// Need to sleep since the webhook http handler is non blocking
	time.Sleep(2 * time.Second)

	// Now there should be a new commit on the pachyderm repo / master branch
	branches, err := c.ListBranch(repoName)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0].Name)
	commit := branches[0].Head

	// Now wait for the pipeline complete as normal
	outputRepo := &pfs.Repo{Name: pipeline}
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, []*pfs.Repo{outputRepo})
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	commit = commitInfos[0].Commit

	var buf bytes.Buffer

	require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, outputFilename, 0, 0, &buf))
	require.Equal(t, "9047fbfc251e7412ef3300868f743f2c24852539", strings.TrimSpace(buf.String()))
}

func TestPipelineWithGitInputMultiPipelineSeparateInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	outputFilename := "commitSHA"
	repos := []string{"pachyderm", "foo"}
	pipelines := []string{
		uniqueString("github_pipeline_a_"),
		uniqueString("github_pipeline_b_"),
	}
	for i, repoName := range repos {
		require.NoError(t, c.CreatePipeline(
			pipelines[i],
			"",
			[]string{"bash"},
			[]string{
				fmt.Sprintf("cat /pfs/%v/.git/HEAD > /pfs/out/%v", repoName, outputFilename),
			},
			nil,
			&pps.Input{
				Git: &pps.GitInput{
					URL:  "https://github.com/pachyderm/test-artifacts.git",
					Name: repoName,
				},
			},
			"",
			false,
		))
		// There should be a pachyderm repo created w no commits:
		repos, err := c.ListRepo(nil)
		require.NoError(t, err)
		found := false
		for _, repo := range repos {
			if repo.Repo.Name == repoName {
				found = true
			}
		}
		require.Equal(t, true, found)

		commits, err := c.ListCommit(repoName, "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 0, len(commits))
	}

	// To trigger the pipeline, we'll need to simulate the webhook by pushing a POST payload to the githook server
	simulateGitPush(t, "../../etc/testing/artifacts/githook-payloads/master.json")
	// Need to sleep since the webhook http handler is non blocking
	time.Sleep(2 * time.Second)

	for i, repoName := range repos {
		// Now there should be a new commit on the pachyderm repo / master branch
		branches, err := c.ListBranch(repoName)
		require.NoError(t, err)
		require.Equal(t, 1, len(branches))
		require.Equal(t, "master", branches[0].Name)
		commit := branches[0].Head

		// Now wait for the pipeline complete as normal
		outputRepo := &pfs.Repo{Name: pipelines[i]}
		commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, []*pfs.Repo{outputRepo})
		require.NoError(t, err)
		commitInfos := collectCommitInfos(t, commitIter)
		require.Equal(t, 1, len(commitInfos))

		commit = commitInfos[0].Commit

		var buf bytes.Buffer

		require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, outputFilename, 0, 0, &buf))
		require.Equal(t, "9047fbfc251e7412ef3300868f743f2c24852539", strings.TrimSpace(buf.String()))
	}
}

func TestPipelineWithGitInputMultiPipelineSameInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	outputFilename := "commitSHA"
	repos := []string{"test-artifacts", "test-artifacts"}
	pipelines := []string{
		uniqueString("github_pipeline_a_"),
		uniqueString("github_pipeline_b_"),
	}
	for i, repoName := range repos {
		require.NoError(t, c.CreatePipeline(
			pipelines[i],
			"",
			[]string{"bash"},
			[]string{
				fmt.Sprintf("cat /pfs/%v/.git/HEAD > /pfs/out/%v", repoName, outputFilename),
			},
			nil,
			&pps.Input{
				Git: &pps.GitInput{
					URL: "https://github.com/pachyderm/test-artifacts.git",
				},
			},
			"",
			false,
		))
		// There should be a pachyderm repo created w no commits:
		repos, err := c.ListRepo(nil)
		require.NoError(t, err)
		found := false
		for _, repo := range repos {
			if repo.Repo.Name == repoName {
				found = true
			}
		}
		require.Equal(t, true, found)

		commits, err := c.ListCommit(repoName, "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 0, len(commits))
	}

	// To trigger the pipeline, we'll need to simulate the webhook by pushing a POST payload to the githook server
	simulateGitPush(t, "../../etc/testing/artifacts/githook-payloads/master.json")
	// Need to sleep since the webhook http handler is non blocking
	time.Sleep(2 * time.Second)

	// Now there should be a new commit on the pachyderm repo / master branch
	branches, err := c.ListBranch(repos[0])
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0].Name)
	commit := branches[0].Head

	// Now wait for the pipeline complete as normal
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 2, len(commitInfos))

	commit = commitInfos[0].Commit

	for _, commitInfo := range commitInfos {
		commit = commitInfo.Commit
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, outputFilename, 0, 0, &buf))
		require.Equal(t, "9047fbfc251e7412ef3300868f743f2c24852539", strings.TrimSpace(buf.String()))
	}
}

func TestPipelineWithGitInputAndBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	branchName := "foo"
	outputFilename := "commitSHA"
	pipeline := uniqueString("github_pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/test-artifacts/.git/HEAD > /pfs/out/%v", outputFilename),
		},
		nil,
		&pps.Input{
			Git: &pps.GitInput{
				URL:    "https://github.com/pachyderm/test-artifacts.git",
				Branch: branchName,
			},
		},
		"",
		false,
	))
	// There should be a pachyderm repo created w no commits:
	repos, err := c.ListRepo(nil)
	require.NoError(t, err)
	found := false
	newRepoName := "test-artifacts"
	for _, repo := range repos {
		if repo.Repo.Name == newRepoName {
			found = true
		}
	}
	require.Equal(t, true, found)

	commits, err := c.ListCommit(newRepoName, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 0, len(commits))

	// Make sure a push to master does NOT trigger this pipeline
	simulateGitPush(t, "../../etc/testing/artifacts/githook-payloads/master.json")
	// Need to sleep since the webhook http handler is non blocking
	time.Sleep(5 * time.Second)
	// Now there should be a new commit on the pachyderm repo / master branch
	branches, err := c.ListBranch(newRepoName)
	require.NoError(t, err)
	require.Equal(t, 0, len(branches))

	// To trigger the pipeline, we'll need to simulate the webhook by pushing a POST payload to the githook server
	simulateGitPush(t, "../../etc/testing/artifacts/githook-payloads/branch.json")
	// Need to sleep since the webhook http handler is non blocking
	time.Sleep(2 * time.Second)
	// Now there should be a new commit on the pachyderm repo / master branch
	branches, err = c.ListBranch(newRepoName)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, branchName, branches[0].Name)
	commit := branches[0].Head

	// Now wait for the pipeline complete as normal
	outputRepo := &pfs.Repo{Name: pipeline}
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, []*pfs.Repo{outputRepo})
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	commit = commitInfos[0].Commit

	var buf bytes.Buffer

	require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, outputFilename, 0, 0, &buf))
	require.Equal(t, "81269575dcfc6ac2e2a463ad8016163f79c97f5c", strings.TrimSpace(buf.String()))
}

func TestPipelineWithDatumTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineWithDatumTimeout_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file",
		strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	timeout := 20
	pipeline := uniqueString("pipeline")
	duration, err := time.ParseDuration(fmt.Sprintf("%vs", timeout))
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					"while true; do sleep 1; date; done",
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:        client.NewAtomInput(dataRepo, "/*"),
			EnableStats:  true,
			DatumTimeout: types.DurationProto(duration),
		},
	)
	require.NoError(t, err)

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	jobs, err := c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	// Block on the job being complete before we call ListDatum
	jobInfo, err := c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)

	// Now validate the datum timed out properly
	resp, err := c.ListDatum(jobs[0].Job.ID, 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.DatumInfos))

	datum, err := c.InspectDatum(jobs[0].Job.ID, resp.DatumInfos[0].Datum.ID)
	require.NoError(t, err)
	require.Equal(t, pps.DatumState_FAILED, datum.State)
	// ProcessTime looks like "20 seconds"
	tokens := strings.Split(pretty.Duration(datum.Stats.ProcessTime), " ")
	require.Equal(t, 2, len(tokens))
	seconds, err := strconv.Atoi(tokens[0])
	require.NoError(t, err)
	require.Equal(t, timeout, seconds)
}

func TestPipelineWithDatumTimeoutControl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineWithDatumTimeoutControl_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file",
		strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	timeout := 20
	pipeline := uniqueString("pipeline")
	duration, err := time.ParseDuration(fmt.Sprintf("%vs", timeout))
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("sleep %v", timeout-10),
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:        client.NewAtomInput(dataRepo, "/*"),
			DatumTimeout: types.DurationProto(duration),
		},
	)
	require.NoError(t, err)

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	jobs, err := c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	// Block on the job being complete before we call ListDatum
	jobInfo, err := c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
}

func TestPipelineWithJobTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())

	dataRepo := uniqueString("TestPipelineWithDatumTimeout_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	numFiles := 2
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%v", i),
			strings.NewReader("foo"))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	timeout := 20
	pipeline := uniqueString("pipeline")
	duration, err := time.ParseDuration(fmt.Sprintf("%vs", timeout))
	require.NoError(t, err)
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					fmt.Sprintf("sleep %v", timeout), // we have 2 datums, so the total exec time will more than double the timeout value
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			Input:       client.NewAtomInput(dataRepo, "/*"),
			EnableStats: true,
			JobTimeout:  types.DurationProto(duration),
		},
	)
	require.NoError(t, err)

	// Wait for the job to get scheduled / appear in listjob
	// A sleep of 15s is insufficient
	time.Sleep(25 * time.Second)
	jobs, err := c.ListJob(pipeline, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))

	// Block on the job being complete before we call ListDatum
	jobInfo, err := c.InspectJob(jobs[0].Job.ID, true)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)
	durationString := pretty.TimeDifference(jobInfo.Started, jobInfo.Finished)
	// duration looks like "20 seconds"
	tokens := strings.Split(durationString, " ")
	require.Equal(t, 2, len(tokens))
	seconds, err := strconv.Atoi(tokens[0])
	require.NoError(t, err)
	epsilon := math.Abs(float64(seconds - timeout))
	require.True(t, epsilon <= 1.0)
}

func TestCommitDescription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dataRepo := uniqueString("TestCommitDescription")
	require.NoError(t, c.CreateRepo(dataRepo))

	// Test putting a message in StartCommit
	commit, err := c.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
		Branch:      "master",
		Parent:      client.NewCommit(dataRepo, ""),
		Description: "test commit description in start-commit",
	})
	require.NoError(t, err)
	c.FinishCommit(dataRepo, commit.ID)
	commitInfo, err := c.InspectCommit(dataRepo, commit.ID)
	require.NoError(t, err)
	require.Equal(t, "test commit description in start-commit", commitInfo.Description)
	require.NoError(t, pfspretty.PrintDetailedCommitInfo(commitInfo))

	// Test putting a message in FinishCommit
	commit, err = c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	c.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
		Commit:      commit,
		Description: "test commit description in finish-commit",
	})
	commitInfo, err = c.InspectCommit(dataRepo, commit.ID)
	require.NoError(t, err)
	require.Equal(t, "test commit description in finish-commit", commitInfo.Description)
	require.NoError(t, pfspretty.PrintDetailedCommitInfo(commitInfo))

	// Test overwriting a commit message
	commit, err = c.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
		Branch:      "master",
		Parent:      client.NewCommit(dataRepo, ""),
		Description: "test commit description in start-commit",
	})
	require.NoError(t, err)
	c.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
		Commit:      commit,
		Description: "test commit description in finish-commit that overwrites",
	})
	commitInfo, err = c.InspectCommit(dataRepo, commit.ID)
	require.NoError(t, err)
	require.Equal(t, "test commit description in finish-commit that overwrites", commitInfo.Description)
	require.NoError(t, pfspretty.PrintDetailedCommitInfo(commitInfo))
}

func TestGetFileWithEmptyCommits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	repoName := uniqueString("TestGetFileWithEmptyCommits")
	require.NoError(t, c.CreateRepo(repoName))

	// Create a real commit in repoName/master
	commit, err := c.StartCommit(repoName, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repoName, commit.ID, "/file", strings.NewReader("data contents"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repoName, commit.ID))

	// Create an empty commit in repoName/master
	commit, err = c.StartCommit(repoName, "master")
	c.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
		Commit: commit,
		Empty:  true,
	})

	// We get a "file not found" error when we try to get a file from repoName/master
	buf := bytes.Buffer{}
	err = c.GetFile(repoName, "master", "/file", 0, 0, &buf)
	require.YesError(t, err)
	require.True(t, strings.Contains(err.Error(), "not found"))
}

func getAllObjects(t testing.TB, c *client.APIClient) []*pfs.Object {
	objectsClient, err := c.ListObjects(context.Background(), &pfs.ListObjectsRequest{})
	require.NoError(t, err)
	var objects []*pfs.Object
	for object, err := objectsClient.Recv(); err != io.EOF; object, err = objectsClient.Recv() {
		require.NoError(t, err)
		objects = append(objects, object)
	}
	return objects
}

func getAllTags(t testing.TB, c *client.APIClient) []string {
	tagsClient, err := c.ListTags(context.Background(), &pfs.ListTagsRequest{})
	require.NoError(t, err)
	var tags []string
	for resp, err := tagsClient.Recv(); err != io.EOF; resp, err = tagsClient.Recv() {
		require.NoError(t, err)
		tags = append(tags, resp.Tag)
	}
	return tags
}

func restartAll(t *testing.T) {
	k := getKubeClient(t)
	podsInterface := k.CoreV1().Pods(v1.NamespaceDefault)
	podList, err := podsInterface.List(
		metav1.ListOptions{
			LabelSelector: "suite=pachyderm",
		})
	require.NoError(t, err)
	for _, pod := range podList.Items {
		require.NoError(t, podsInterface.Delete(pod.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		}))
	}
	waitForReadiness(t)
}

func restartOne(t *testing.T) {
	k := getKubeClient(t)
	podsInterface := k.CoreV1().Pods(v1.NamespaceDefault)
	podList, err := podsInterface.List(
		metav1.ListOptions{
			LabelSelector: "app=pachd",
		})
	require.NoError(t, err)
	require.NoError(t, podsInterface.Delete(
		podList.Items[rand.Intn(len(podList.Items))].Name,
		&metav1.DeleteOptions{GracePeriodSeconds: new(int64)}))
	waitForReadiness(t)
}

const (
	retries = 10
)

// getUsablePachClient is like getPachClient except it blocks until it gets a
// connection that actually works
func getUsablePachClient(t *testing.T) *client.APIClient {
	for i := 0; i < retries; i++ {
		client := getPachClient(t)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel() //cleanup resources
		_, err := client.PfsAPIClient.ListRepo(ctx, &pfs.ListRepoRequest{})
		if err == nil {
			return client
		}
	}
	t.Fatalf("failed to connect after %d tries", retries)
	return nil
}

func podRunningAndReady(e watch.Event) (bool, error) {
	if e.Type == watch.Deleted {
		return false, errors.New("received DELETE while watching pods")
	}
	pod, ok := e.Object.(*v1.Pod)
	if !ok {
	}
	return pod.Status.Phase == v1.PodRunning, nil
}

func waitForReadiness(t testing.TB) {
	k := getKubeClient(t)
	deployment := pachdDeployment(t)
	for {
		newDeployment, err := k.Apps().Deployments(v1.NamespaceDefault).Get(deployment.Name, metav1.GetOptions{})
		require.NoError(t, err)
		if newDeployment.Status.ObservedGeneration >= deployment.Generation && newDeployment.Status.Replicas == *newDeployment.Spec.Replicas {
			break
		}
		time.Sleep(time.Second * 5)
	}
	watch, err := k.CoreV1().Pods(v1.NamespaceDefault).Watch(metav1.ListOptions{
		LabelSelector: "app=pachd",
	})
	defer watch.Stop()
	require.NoError(t, err)
	readyPods := make(map[string]bool)
	for event := range watch.ResultChan() {
		ready, err := podRunningAndReady(event)
		require.NoError(t, err)
		if ready {
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				t.Fatal("event.Object should be an object")
			}
			readyPods[pod.Name] = true
			if len(readyPods) == int(*deployment.Spec.Replicas) {
				break
			}
		}
	}
}

func simulateGitPush(t *testing.T, pathToPayload string) {
	payload, err := ioutil.ReadFile(pathToPayload)
	require.NoError(t, err)
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://127.0.0.1:%v/v1/handle/push", githook.GitHookPort+30000),
		bytes.NewBuffer(payload),
	)
	req.Header.Set("X-Github-Delivery", "2984f5d0-c032-11e7-82d7-ed3ee54be25d")
	req.Header.Set("User-Agent", "GitHub-Hookshot/c1d08eb")
	req.Header.Set("X-Github-Event", "push")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
}

func pipelineRc(t testing.TB, pipelineInfo *pps.PipelineInfo) (*v1.ReplicationController, error) {
	k := getKubeClient(t)
	rc := k.CoreV1().ReplicationControllers(v1.NamespaceDefault)
	return rc.Get(
		ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
		metav1.GetOptions{})
}

func pachdDeployment(t testing.TB) *apps.Deployment {
	k := getKubeClient(t)
	result, err := k.Apps().Deployments(v1.NamespaceDefault).Get("pachd", metav1.GetOptions{})
	require.NoError(t, err)
	return result
}

// scalePachd scales the number of pachd nodes up or down.
// If up is true, then the number of nodes will be within (n, 2n]
// If up is false, then the number of nodes will be within [1, n)
func scalePachdRandom(t testing.TB, up bool) {
	pachdRc := pachdDeployment(t)
	originalReplicas := *pachdRc.Spec.Replicas
	for {
		if up {
			*pachdRc.Spec.Replicas = originalReplicas + int32(rand.Intn(int(originalReplicas))+1)
		} else {
			*pachdRc.Spec.Replicas = int32(rand.Intn(int(originalReplicas)-1) + 1)
		}

		if *pachdRc.Spec.Replicas != originalReplicas {
			break
		}
	}
	scalePachdN(t, int(*pachdRc.Spec.Replicas))
}

// scalePachdN scales the number of pachd nodes to N
func scalePachdN(t testing.TB, n int) {
	k := getKubeClient(t)
	// Modify the type metadata of the Deployment spec we read from k8s, so that
	// k8s will accept it if we're talking to a 1.7 cluster
	pachdDeployment := pachdDeployment(t)
	*pachdDeployment.Spec.Replicas = int32(n)
	pachdDeployment.TypeMeta.APIVersion = "apps/v1beta1"
	_, err := k.Apps().Deployments(v1.NamespaceDefault).Update(pachdDeployment)
	require.NoError(t, err)
	waitForReadiness(t)
	// Unfortunately, even when all pods are ready, the cluster membership
	// protocol might still be running, thus PFS API calls might fail.  So
	// we wait a little bit for membership to stablize.
	time.Sleep(15 * time.Second)
}

// scalePachd reads the number of pachd nodes from an env variable and
// scales pachd accordingly.
func scalePachd(t testing.TB) {
	nStr := os.Getenv("PACHD")
	if nStr == "" {
		return
	}
	n, err := strconv.Atoi(nStr)
	require.NoError(t, err)
	scalePachdN(t, n)
}

func getKubeClient(t testing.TB) *kube.Clientset {
	var config *rest.Config
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	if host != "" {
		var err error
		config, err = rest.InClusterConfig()
		require.NoError(t, err)
	} else {
		// Use kubectl binary to parse .kube/config and get address of current
		// cluster. Hopefully, once we upgrade to k8s.io/client-go, we will be able
		// to do this in-process with a library
		// First, figure out if we're talking to minikube or localhost
		cmd := exec.Command("kubectl", "config", "current-context")
		if context, err := cmd.Output(); err == nil {
			context = bytes.TrimSpace(context)
			// kubectl has a context -- not talking to localhost
			// Get cluster and user name from kubectl
			buf := &bytes.Buffer{}
			cmd := tu.BashCmd(strings.Join([]string{
				`kubectl config get-contexts "{{.context}}" | tail -n+2 | awk '{print $3}'`,
				`kubectl config get-contexts "{{.context}}" | tail -n+2 | awk '{print $4}'`,
			}, "\n"),
				"context", string(context))
			cmd.Stdout = buf
			require.NoError(t, cmd.Run(), "couldn't get kubernetes context info")
			lines := strings.Split(buf.String(), "\n")
			clustername, username := lines[0], lines[1]

			// Get user info
			buf.Reset()
			cmd = tu.BashCmd(strings.Join([]string{
				`cluster="$(kubectl config view -o json | jq -r '.users[] | select(.name == "{{.user}}") | .user' )"`,
				`echo "${cluster}" | jq -r '.["client-certificate"]'`,
				`echo "${cluster}" | jq -r '.["client-key"]'`,
			}, "\n"),
				"user", username)
			cmd.Stdout = buf
			require.NoError(t, cmd.Run(), "couldn't get kubernetes user info")
			lines = strings.Split(buf.String(), "\n")
			clientCert, clientKey := lines[0], lines[1]

			// Get cluster info
			buf.Reset()
			cmd = tu.BashCmd(strings.Join([]string{
				`cluster="$(kubectl config view -o json | jq -r '.clusters[] | select(.name == "{{.cluster}}") | .cluster')"`,
				`echo "${cluster}" | jq -r .server`,
				`echo "${cluster}" | jq -r '.["certificate-authority"]'`,
			}, "\n"),
				"cluster", clustername)
			cmd.Stdout = buf
			require.NoError(t, cmd.Run(), "couldn't get kubernetes cluster info: %s", buf.String())
			lines = strings.Split(buf.String(), "\n")
			address, CAKey := lines[0], lines[1]

			// Generate config
			config = &rest.Config{
				Host: address,
				TLSClientConfig: rest.TLSClientConfig{
					CertFile: clientCert,
					KeyFile:  clientKey,
					CAFile:   CAKey,
				},
			}
		} else {
			// no context -- talking to localhost
			config = &rest.Config{
				Host: "http://0.0.0.0:8080",
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: false,
				},
			}
		}
	}
	k, err := kube.NewForConfig(config)
	require.NoError(t, err)
	return k
}

var pachClient *client.APIClient
var getPachClientOnce sync.Once

func getPachClient(t testing.TB) *client.APIClient {
	getPachClientOnce.Do(func() {
		var err error
		if addr := os.Getenv("PACHD_PORT_650_TCP_ADDR"); addr != "" {
			pachClient, err = client.NewInCluster()
		} else {
			pachClient, err = client.NewOnUserMachine(false, "user")
		}
		require.NoError(t, err)
	})
	return pachClient
}

func uniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}
