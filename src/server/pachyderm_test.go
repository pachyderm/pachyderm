package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfspretty "github.com/pachyderm/pachyderm/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	ppspretty "github.com/pachyderm/pachyderm/src/server/pps/pretty"

	"github.com/gogo/protobuf/types"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	kube_client "k8s.io/kubernetes/pkg/client/restclient"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
)

const (
	// If this environment variable is set, then the tests are being run
	// in a real cluster in the cloud.
	InCloudEnv = "PACH_TEST_CLOUD"
)

func TestPipelineWithParallelism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	dataRepo := uniqueString("TestPipelineInputDataModification_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 1000
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

func TestDatumDedup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

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
	t.Parallel()
	c := getPachClient(t)

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

	commitInfos, err = c.ListCommit(pipeline, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
}

func TestMultipleInputsFromTheSameBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

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
			client.NewAtomInputOpts("dirA", dataRepo, "", "/dirA/*", false, ""),
			client.NewAtomInputOpts("dirB", dataRepo, "", "/dirB/*", false, ""),
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

	commitInfos, err = c.ListCommit(pipeline, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
}

func TestMultipleInputsFromTheSameRepoDifferentBranches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	dataRepo := uniqueString("TestMultipleInputsFromTheSameRepo_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	branchA := "branchA"
	branchB := "branchB"

	pipeline := uniqueString("pipeline")
	// Creating this pipeline should error, because the two inputs are
	// from the same repo but they don't specify different names.
	require.YesError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/%s/file > /pfs/out/file", dataRepo),
			fmt.Sprintf("cat /pfs/%s/file > /pfs/out/file", dataRepo),
		},
		nil,
		client.NewCrossInput(
			client.NewAtomInputOpts("", dataRepo, branchA, "/*", false, ""),
			client.NewAtomInputOpts("", dataRepo, branchB, "/*", false, ""),
		),
		"",
		false,
	))

	require.YesError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/%s/file >> /pfs/out/file", branchA),
			fmt.Sprintf("cat /pfs/%s/file >> /pfs/out/file", branchB),
		},
		nil,
		client.NewCrossInput(
			client.NewAtomInputOpts(branchA, dataRepo, branchA, "/*", false, ""),
			client.NewAtomInputOpts(branchB, dataRepo, branchB, "/*", false, ""),
		),
		"",
		false,
	))
}

//func TestJob(t *testing.T) {
//t.Parallel()
//testJob(t, 4)
//}

//func TestJobNoShard(t *testing.T) {
//t.Parallel()
//testJob(t, 0)
//}

//func testJob(t *testing.T, shards int) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}
//c := getPachClient(t)

//// Create repo, commit, and branch
//dataRepo := uniqueString("TestJob_data")
//require.NoError(t, c.CreateRepo(dataRepo))
//commit, err := c.StartCommit(dataRepo, "master")
//require.NoError(t, err)

//fileContent := "foo\n"
//// We want to create lots of files so that each parallel job will be
//// started with some files
//numFiles := shards*100 + 100
//for i := 0; i < numFiles; i++ {
//fmt.Println("putting ", i)
//_, err = c.PutFile(dataRepo, commit.ID, fmt.Sprintf("file-%d", i), strings.NewReader(fileContent))
//require.NoError(t, err)
//}
//require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
//job, err := c.CreateJob(
//"",
//[]string{"bash"},
//[]string{fmt.Sprintf("cp %s %s", "/pfs/input/*", "/pfs/out")},
//&pps.ParallelismSpec{
//Constant: uint64(shards),
//},
//[]*pps.JobInput{{
//Name:   "input",
//Commit: commit,
//Glob:   "/*",
//}},
//0,
//0,
//)
//require.NoError(t, err)

//// Wait for job to finish and then inspect
//jobInfo, err := c.InspectJob(job.ID, true [> wait for job <])
//require.NoError(t, err)
//require.Equal(t, pps.JobState_JOB_SUCCESS.String(), jobInfo.State.String())
//require.NotNil(t, jobInfo.Started)
//require.NotNil(t, jobInfo.Finished)

//// Inspect job timestamps
//tFin, _ := types.TimestampFromProto(jobInfo.Finished)
//tStart, _ := types.TimestampFromProto(jobInfo.Started)
//require.True(t, tFin.After(tStart))

//// Inspect job parallelism
//parellelism, err := pps_server.GetExpectedNumWorkers(getKubeClient(t), jobInfo.ParallelismSpec)
//require.NoError(t, err)
//require.True(t, parellelism > 0)

//// Inspect output commit
//_, err = c.InspectCommit(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
//require.NoError(t, err)

//// Inspect output files
//for i := 0; i < numFiles; i++ {
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, fmt.Sprintf("file-%d", i), 0, 0, &buffer))
//require.Equal(t, fileContent, buffer.String())
//}
//}

// This test fails if you updated some static assets (such as doc/reference/pipeline_spec.md)
// that are used in code but forgot to run:
// $ make assets

func TestPipelineFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

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
		jobInfos, err = c.ListJob(pipeline, nil)
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
}

func TestLazyPipelinePropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)
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
		client.NewAtomInputOpts("", dataRepo, "", "/*", true, ""),
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
		client.NewAtomInputOpts("", pipelineA, "", "/*", true, ""),
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

	jobInfos, err := c.ListJob(pipelineA, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	require.NotNil(t, jobInfos[0].Input.Atom)
	require.Equal(t, true, jobInfos[0].Input.Atom.Lazy)
	jobInfos, err = c.ListJob(pipelineB, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	require.NotNil(t, jobInfos[0].Input.Atom)
	require.Equal(t, true, jobInfos[0].Input.Atom.Lazy)
}

func TestLazyPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
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
			Inputs: []*pps.PipelineInput{{
				Repo: &pfs.Repo{Name: dataRepo},
				Glob: "/",
				Lazy: true,
			}},
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
	t.Parallel()

	c := getPachClient(t)
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
			Inputs: []*pps.PipelineInput{{
				Repo: &pfs.Repo{Name: dataRepo},
				Glob: "/",
				Lazy: true,
			}},
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
		jobInfos, err := c.ListJob(pipeline, nil)
		if err != nil {
			return err
		}
		if len(jobInfos) != 1 {
			return fmt.Errorf("len(jobInfos) should be 1")
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

	t.Parallel()
	c := getPachClient(t)
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

	// We should only see two commits in each repo.
	commitInfos, err = c.ListCommit(aRepo, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	commitInfos, err = c.ListCommit(bPipeline, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	commitInfos, err = c.ListCommit(cPipeline, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
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

	t.Parallel()
	c := getPachClient(t)
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
	commitInfos, err = c.ListCommit(bPipeline, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	commitInfos, err = c.ListCommit(cPipeline, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	commitInfos, err = c.ListCommit(dPipeline, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	for _, commitInfo := range commitInfos {
		commit := commitInfo.Commit
		buffer := bytes.Buffer{}
		require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, "file", 0, 0, &buffer))
		require.Equal(t, "", buffer.String())
	}
}

//func TestDirectory(t *testing.T) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}
//t.Parallel()

//c := getPachClient(t)

//ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
//defer cancel() //cleanup resources

//job1, err := c.PpsAPIClient.CreateJob(context.Background(), &pps.CreateJobRequest{
//Transform: &pps.Transform{
//Cmd: []string{"sh"},
//Stdin: []string{
//"mkdir /pfs/out/dir",
//"echo foo >> /pfs/out/dir/file",
//},
//},
//ParallelismSpec: &pps.ParallelismSpec{
//Constant: 3,
//},
//})
//require.NoError(t, err)
//inspectJobRequest1 := &pps.InspectJobRequest{
//Job:        job1,
//BlockState: true,
//}
//jobInfo1, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest1)
//require.NoError(t, err)
//require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo1.State)

//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(jobInfo1.OutputCommit.Repo.Name, jobInfo1.OutputCommit.ID, "dir/file", 0, 0, "", false, nil, &buffer))
//require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())

//job2, err := c.PpsAPIClient.CreateJob(context.Background(), &pps.CreateJobRequest{
//Transform: &pps.Transform{
//Cmd: []string{"sh"},
//Stdin: []string{
//"mkdir /pfs/out/dir",
//"echo bar >> /pfs/out/dir/file",
//},
//},
//ParallelismSpec: &pps.ParallelismSpec{
//Constant: 3,
//},
//ParentJob: job1,
//})
//require.NoError(t, err)
//inspectJobRequest2 := &pps.InspectJobRequest{
//Job:        job2,
//BlockState: true,
//}
//jobInfo2, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest2)
//require.NoError(t, err)
//require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo2.State)

//buffer = bytes.Buffer{}
//require.NoError(t, c.GetFile(jobInfo2.OutputCommit.Repo.Name, jobInfo2.OutputCommit.ID, "dir/file", 0, 0, "", false, nil, &buffer))
//require.Equal(t, "foo\nfoo\nfoo\nbar\nbar\nbar\n", buffer.String())
//}

// TestFlushCommit
func TestFlushCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)
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

	t.Parallel()
	c := getPachClient(t)
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

	t.Parallel()
	c := getPachClient(t)
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
	require.NoError(t, c.DeleteRepo(pipeline, false))
	require.NoError(t, c.DeletePipeline(pipeline, true))
	time.Sleep(5 * time.Second)
	createPipeline()
}

func TestDeletePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)
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
	// Wait for the job to start running
	time.Sleep(15 * time.Second)
	require.NoError(t, c.DeleteRepo(pipeline, false))
	require.NoError(t, c.DeletePipeline(pipeline, true))
	time.Sleep(5 * time.Second)

	// The job should be gone
	jobs, err := c.ListJob(pipeline, nil)
	require.NoError(t, err)
	require.Equal(t, len(jobs), 0)

	createPipeline()
	// Wait for the job to start running
	time.Sleep(15 * time.Second)
	require.NoError(t, c.DeleteRepo(pipeline, false))
	require.NoError(t, c.DeletePipeline(pipeline, false))
	time.Sleep(5 * time.Second)

	// The job should still be there, and its state should be "KILLED"
	jobs, err = c.ListJob(pipeline, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	require.Equal(t, pps.JobState_JOB_KILLED, jobs[0].State)
}

func TestPipelineState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)
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

	t.Parallel()
	c := getPachClient(t)
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
	jobInfos, err := c.ListJob(pipeline, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
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
	require.Equal(t, int32(1), pipelineInfo.JobCounts[int32(pps.JobState_JOB_SUCCESS)])
}

//func TestJobState(t *testing.T) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}

//t.Parallel()
//c := getPachClient(t)

//// This job uses a nonexistent image; it's supposed to stay in the
//// "creating" state
//job, err := c.CreateJob(
//"nonexistent",
//[]string{"bash"},
//nil,
//&pps.ParallelismSpec{},
//nil,
//"",
//0,
//0,
//)
//require.NoError(t, err)
//time.Sleep(10 * time.Second)
//jobInfo, err := c.InspectJob(job.ID, false)
//require.NoError(t, err)
//require.Equal(t, pps.JobState_JOB_CREATING, jobInfo.State)

//// This job sleeps for 20 secs
//job, err = c.CreateJob(
//"",
//[]string{"bash"},
//[]string{"sleep 20"},
//&pps.ParallelismSpec{},
//nil,
//"",
//0,
//0,
//)
//require.NoError(t, err)
//time.Sleep(10 * time.Second)
//jobInfo, err = c.InspectJob(job.ID, false)
//require.NoError(t, err)
//require.Equal(t, pps.JobState_JOB_RUNNING, jobInfo.State)

//// Wait for the job to complete
//jobInfo, err = c.InspectJob(job.ID, true)
//require.NoError(t, err)
//require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.State)
//}

//func TestClusterFunctioningAfterMembershipChange(t *testing.T) {
//t.Skip("this test is flaky")
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}

//scalePachd(t, true)
//testJob(t, 4)
//scalePachd(t, false)
//testJob(t, 4)
//}

// TODO(msteffen): This test breaks the suite when run against cloud providers,
// because killing the pachd pod breaks the connection with pachctl port-forward
func TestDeleteAfterMembershipChange(t *testing.T) {
	t.Skip("This is causing intermittent CI failures")

	test := func(up bool) {
		repo := uniqueString("TestDeleteAfterMembershipChange")
		c := getPachClient(t)
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

	jobInfos, err := c.ListJob(pipelineName, nil)
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

//func TestScrubbedErrors(t *testing.T) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}

//t.Parallel()
//c := getPachClient(t)

//_, err := c.InspectPipeline("blah")
//require.Equal(t, "PipelineInfos blah not found", err.Error())

//err = c.CreatePipeline(
//"lskdjf$#%^ERTYC",
//"",
//[]string{},
//nil,
//&pps.ParallelismSpec{
//Constant: 1,
//},
//[]*pps.PipelineInput{{Repo: &pfs.Repo{Name: "test"}}},
//false,
//)
//require.Equal(t, "repo test not found", err.Error())

//_, err = c.CreateJob(
//"askjdfhgsdflkjh",
//[]string{},
//[]string{},
//&pps.ParallelismSpec{},
//[]*pps.JobInput{client.NewJobInput("bogusRepo", "bogusCommit", client.DefaultMethod)},
//"",
//0,
//0,
//)
//require.Matches(t, "could not create repo job_.*, not all provenance repos exist", err.Error())

//_, err = c.InspectJob("blah", true)
//require.Equal(t, "JobInfos blah not found", err.Error())

//home := os.Getenv("HOME")
//f, err := os.Create(filepath.Join(home, "/tmpfile"))
//defer func() {
//os.Remove(filepath.Join(home, "/tmpfile"))
//}()
//require.NoError(t, err)
//err = c.GetLogs("bogusJobId", f)
//require.Equal(t, "job bogusJobId not found", err.Error())
//}

//func TestLeakingRepo(t *testing.T) {
//// If CreateJob fails, it should also destroy the output repo it creates
//// If it doesn't, it can cause flush commit to fail, as a bogus repo will
//// be listed in the output repo's provenance

//// This test can't be run in parallel, since it requires using the repo counts as controls
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}

//c := getPachClient(t)

//repoInfos, err := c.ListRepo(nil)
//require.NoError(t, err)
//initialCount := len(repoInfos)

//_, err = c.CreateJob(
//"bogusImage",
//[]string{},
//[]string{},
//&pps.ParallelismSpec{},
//[]*pps.JobInput{client.NewJobInput("bogusRepo", "bogusCommit", client.DefaultMethod)},
//"",
//0,
//0,
//)
//require.Matches(t, "could not create repo job_.*, not all provenance repos exist", err.Error())

//repoInfos, err = c.ListRepo(nil)
//require.NoError(t, err)
//require.Equal(t, initialCount, len(repoInfos))
//}

// TestUpdatePipelineThatHasNoOutput tracks #1637
func TestUpdatePipelineThatHasNoOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

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
		jobInfos, err = c.ListJob(pipeline, nil)
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
	t.Parallel()
	c := getPachClient(t)

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
			Inputs: []*pps.PipelineInput{{
				Repo: &pfs.Repo{Name: dataRepo},
				Glob: "/*",
			}},
		},
	)
	require.NoError(t, err)

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	jobInfos, err := c.ListJob(pipelineName, nil)
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
	t.Parallel()

	c := getPachClient(t)
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
			ResourceSpec: &pps.ResourceSpec{
				Memory: "100M",
				Cpu:    0.5,
			},
			Inputs: []*pps.PipelineInput{{
				Repo: &pfs.Repo{Name: dataRepo},
				Glob: "/*",
			}},
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
	jobInfos, err := c.ListJob("", nil)
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
	jobInfos, err := c.ListJob("", nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(jobInfos))
}

func TestRecursiveCp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()

	c := getPachClient(t)
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
	t.Parallel()
	c := getPachClient(t)

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
	t.Parallel()
	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestUpdatePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	pipelineName := uniqueString("TestUpdatePipeline_pipeline")
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

	iter, err := c.SubscribeCommit(pipelineName, "master", "")
	require.NoError(t, err)

	_, err = iter.Next()
	require.NoError(t, err)
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

	_, err = iter.Next()
	require.NoError(t, err)
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

	_, err = iter.Next()
	require.NoError(t, err)
	buffer.Reset()
	require.NoError(t, c.GetFile(pipelineName, "master", "file", 0, 0, &buffer))
	require.Equal(t, "buzz\n", buffer.String())
}

func TestStopPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)
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
	require.NoError(t, c.StopPipeline(pipelineName))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	// wait for 10 seconds and check that no commit has been outputted
	time.Sleep(10 * time.Second)
	commits, err := c.ListCommit(pipelineName, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, len(commits), 0)
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
	t.Parallel()

	c := getPachClient(t)
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
			ResourceSpec: &pps.ResourceSpec{
				Memory: "100M",
			},
			Inputs: []*pps.PipelineInput{{
				Repo: &pfs.Repo{Name: dataRepo},
				Glob: "/",
			}},
			ScaleDownThreshold: types.DurationProto(scaleDownThreshold),
		})
	require.NoError(t, err)

	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)

	// Wait for the pipeline to scale down
	b := backoff.NewTestingBackOff()
	b.MaxElapsedTime = scaleDownThreshold + 30*time.Second
	checkScaleDown := func() error {
		rc, err := pipelineRc(t, pipelineInfo)
		if err != nil {
			return err
		}
		if rc.Spec.Replicas != 1 {
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
	require.Equal(t, parallelism, int(rc.Spec.Replicas))
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
	t.Parallel()

	// make a secret to reference
	k := getKubeClient(t)
	secretName := uniqueString("test-secret")
	_, err := k.Secrets(api.NamespaceDefault).Create(
		&api.Secret{
			ObjectMeta: api.ObjectMeta{
				Name: secretName,
			},
			Data: map[string][]byte{
				"foo": []byte("foo\n"),
			},
		},
	)
	require.NoError(t, err)
	c := getPachClient(t)
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
				},
				Env: map[string]string{"bar": "bar"},
				Secrets: []*pps.Secret{
					{
						Name:      secretName,
						MountPath: "/var/secret",
					},
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Inputs: []*pps.PipelineInput{{
				Repo: &pfs.Repo{Name: dataRepo},
				Glob: "/*",
			}},
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
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(pipelineName, commitInfos[0].Commit.ID, "bar", 0, 0, &buffer))
	require.Equal(t, "bar\n", buffer.String())
}

func TestPipelineWithFullObjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)
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
	t.Parallel()
	c := getPachClient(t)
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

	// Make sure that we got two output commits
	commitInfos, err = c.ListCommit(pipelineName, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
}

func TestPipelineThatSymlinks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

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
	t.Parallel()
	c := getPachClient(t)
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
	c := getPachClient(t)
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
	jobInfos, err := c.ListJob(dPipeline, nil)
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
	nodes, err := kubeclient.Nodes().List(api.ListOptions{})
	numNodes := len(nodes.Items)

	// Test Constant strategy
	parellelism, err := ppsserver.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Constant: 7,
	})
	require.NoError(t, err)
	require.Equal(t, 7, parellelism)

	// Coefficient == 1 (basic test)
	// TODO(msteffen): This test can fail when run against cloud providers, if the
	// remote cluster has more than one node (in which case "Coefficient: 1" will
	// cause more than 1 worker to start)
	parellelism, err = ppsserver.GetExpectedNumWorkers(kubeclient, &pps.ParallelismSpec{
		Coefficient: 1,
	})
	require.NoError(t, err)
	require.Equal(t, numNodes, parellelism)

	// Coefficient > 1
	parellelism, err = ppsserver.GetExpectedNumWorkers(kubeclient, &pps.ParallelismSpec{
		Coefficient: 2,
	})
	require.NoError(t, err)
	require.Equal(t, 2*numNodes, parellelism)

	// Make sure we start at least one worker
	parellelism, err = ppsserver.GetExpectedNumWorkers(kubeclient, &pps.ParallelismSpec{
		Coefficient: 0.01,
	})
	require.NoError(t, err)
	require.Equal(t, 1, parellelism)

	// Test 0-initialized JobSpec
	parellelism, err = ppsserver.GetExpectedNumWorkers(kubeclient, &pps.ParallelismSpec{})
	require.NoError(t, err)
	require.Equal(t, numNodes, parellelism)

	// Test nil JobSpec
	parellelism, err = ppsserver.GetExpectedNumWorkers(kubeclient, nil)
	require.NoError(t, err)
	require.Equal(t, numNodes, parellelism)
}

func TestPipelineJobDeletion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
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
	jobInfos, err := c.ListJob(pipelineName, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	err = c.DeleteJob(jobInfos[0].Job.ID)
	require.NoError(t, err)
}

func TestStopJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
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
		jobInfos, err := c.ListJob(pipelineName, nil)
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
		jobInfos, err := c.ListJob(pipelineName, nil)
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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"sh"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/file /pfs/out/file", dataRepo),
			"echo foo",
			"echo foo",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))

	// Commit data to repo and flush commit
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	_, err = commitIter.Next()
	require.NoError(t, err)

	// List output commits, to make sure one exists?
	commits, err := c.ListCommitByRepo(pipelineName)
	require.NoError(t, err)
	require.True(t, len(commits) == 1)

	// Get logs from pipeline, using pipeline
	iter := c.GetLogs(pipelineName, "", nil, false)
	var numLogs int
	for iter.Next() {
		numLogs++
		require.True(t, iter.Message().Message != "")
	}
	require.True(t, numLogs > 0)
	require.NoError(t, iter.Err())

	// Get logs from pipeline, using a pipeline that doesn't exist. There should
	// be an error
	iter = c.GetLogs("__DOES_NOT_EXIST__", "", nil, false)
	require.False(t, iter.Next())
	require.YesError(t, iter.Err())
	require.Matches(t, "could not get", iter.Err().Error())

	// Get logs from pipeline, using job
	// (1) Get job ID, from pipeline that just ran
	jobInfos, err := c.ListJob(pipelineName, nil)
	require.NoError(t, err)
	require.True(t, len(jobInfos) == 1)
	// (2) Get logs using extracted job ID
	// wait for logs to be collected
	time.Sleep(10 * time.Second)
	iter = c.GetLogs("", jobInfos[0].Job.ID, nil, false)
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
	iter = c.GetLogs("", "__DOES_NOT_EXIST__", nil, false)
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
		pathLog := c.GetLogs("", jobInfos[0].Job.ID, []string{"/file"}, false)

		hexHash := "19fdf57bdf9eb5a9602bfa9c0e6dd7ed3835f8fd431d915003ea82747707be66"
		require.Equal(t, hexHash, hex.EncodeToString(fileInfo.Hash)) // sanity-check test
		hexLog := c.GetLogs("", jobInfos[0].Job.ID, []string{hexHash}, false)

		base64Hash := "Gf31e9+etalgK/qcDm3X7Tg1+P1DHZFQA+qCdHcHvmY="
		require.Equal(t, base64Hash, base64.StdEncoding.EncodeToString(fileInfo.Hash))
		base64Log := c.GetLogs("", jobInfos[0].Job.ID, []string{base64Hash}, false)

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
	iter = c.GetLogs("", jobInfos[0].Job.ID, []string{"__DOES_NOT_EXIST__"}, false)
	require.False(t, iter.Next())
	require.NoError(t, iter.Err())
}

func TestPfsPutFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
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
	err = c.PutFileURL(repo2, commit2.ID, "file", fmt.Sprintf("pfs://0.0.0.0:650/%s/%s/file1", repo1, commit1.ID), false)
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo2, commit2.ID))
	var buf bytes.Buffer
	require.NoError(t, c.GetFile(repo2, commit2.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\n", buf.String())

	commit3, err := c.StartCommit(repo2, "")
	require.NoError(t, err)
	err = c.PutFileURL(repo2, commit3.ID, "", fmt.Sprintf("pfs://0.0.0.0:650/%s/%s", repo1, commit1.ID), true)
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
	t.Parallel()
	c := getPachClient(t)

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
	t.Parallel()
	c := getPachClient(t)

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
	checkStatus := func() {
		started := time.Now()
		for {
			time.Sleep(time.Second)
			if time.Since(started) > time.Second*60 {
				t.Fatalf("failed to find status in time")
			}
			jobs, err := c.ListJob(pipeline, nil)
			require.NoError(t, err)
			if len(jobs) == 0 {
				continue
			}
			jobID = jobs[0].Job.ID
			jobInfo, err := c.InspectJob(jobs[0].Job.ID, false)
			require.NoError(t, err)
			if len(jobInfo.WorkerStatus) == 0 {
				continue
			}
			if jobInfo.WorkerStatus[0].JobID == jobInfo.Job.ID {
				// This method is called before and after the datum is
				// restarted, this makes sure that the restart actually did
				// something.
				// The first time this function is called, datumStarted is zero
				// so `Before` is true for any non-zero time.
				_datumStarted, err := types.TimestampFromProto(jobInfo.WorkerStatus[0].Started)
				require.NoError(t, err)
				require.True(t, datumStarted.Before(_datumStarted))
				datumStarted = _datumStarted
				break
			}
		}
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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

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
	// This pipeline sleeps for 20 secs per datum
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			"sleep 5",
		},
		&pps.ParallelismSpec{
			Constant: 2,
		},
		client.NewAtomInput(dataRepo, "/*"),
		"",
		false,
	))
	err = backoff.Retry(func() error {
		jobs, err := c.ListJob(pipeline, nil)
		require.NoError(t, err)
		if len(jobs) == 0 {
			return fmt.Errorf("failed to find jobs")
		}
		jobInfo, err := c.InspectJob(jobs[0].Job.ID, false)
		require.NoError(t, err)
		if len(jobInfo.WorkerStatus) != 2 {
			return fmt.Errorf("incorrect number of statuses: %v", len(jobInfo.WorkerStatus))
		}
		return nil
	}, backoff.NewTestingBackOff())
	require.NoError(t, err)
}

// TestSystemResourceRequest doesn't create any jobs or pipelines, it
// just makes sure that when pachyderm is deployed, we give rethinkdb, pachd,
// and etcd default resource requests. This prevents them from overloading
// nodes and getting evicted, which can slow down or break a cluster.
func TestSystemResourceRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
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
	var c api.Container
	for _, app := range []string{"pachd", "etcd"} {
		err := backoff.Retry(func() error {
			podList, err := kubeClient.Pods(api.NamespaceDefault).List(api.ListOptions{
				LabelSelector: labels.SelectorFromSet(
					map[string]string{"app": app, "suite": "pachyderm"}),
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
		cpu, ok := c.Resources.Requests[api.ResourceCPU]
		require.True(t, ok, "could not get CPU request for "+app)
		require.True(t, cpu.String() == defaultLocalCPU[app] ||
			cpu.String() == defaultCloudCPU[app])
		mem, ok := c.Resources.Requests[api.ResourceMemory]
		require.True(t, ok, "could not get memory request for "+app)
		require.True(t, mem.String() == defaultLocalMem[app] ||
			mem.String() == defaultCloudMem[app])
	}
}

// TODO(msteffen) Refactor other tests to use this helper
func PutFileAndFlush(t *testing.T, repo, branch, filepath, contents string) *pfs.Commit {
	// This may be a bit wasteful, since the calling test likely has its own
	// client, but for a test the overhead seems acceptable (and the code is
	// shorter)
	c := getPachClient(t)

	commit, err := c.StartCommit(repo, branch)
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit.ID, filepath, strings.NewReader(contents))
	require.NoError(t, err)

	require.NoError(t, c.FinishCommit(repo, commit.ID))
	_, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	return commit
}

// TestPipelineResourceRequest creates a pipeline with a resource request, and
// makes sure that's passed to k8s (by inspecting the pipeline's pods)
func TestPipelineResourceRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
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
			ResourceSpec: &pps.ResourceSpec{
				Memory: "100M",
				Cpu:    0.5,
				Gpu:    1,
			},
			Inputs: []*pps.PipelineInput{{
				Repo:   &pfs.Repo{dataRepo},
				Branch: "master",
				Glob:   "/*",
			}},
		})
	require.NoError(t, err)
	PutFileAndFlush(t, dataRepo, "master", "file", "foo\n")

	// Get info about the pipeline pods from k8s & check for resources
	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)

	var container api.Container
	rcName := ppsserver.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	kubeClient := getKubeClient(t)
	err = backoff.Retry(func() error {
		podList, err := kubeClient.Pods(api.NamespaceDefault).List(api.ListOptions{
			LabelSelector: labels.SelectorFromSet(
				map[string]string{"app": rcName}),
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
	cpu, ok := container.Resources.Requests[api.ResourceCPU]
	require.True(t, ok)
	require.Equal(t, "500m", cpu.String())
	mem, ok := container.Resources.Requests[api.ResourceMemory]
	require.True(t, ok)
	require.Equal(t, "100M", mem.String())
	gpu, ok := container.Resources.Requests[api.ResourceNvidiaGPU]
	require.True(t, ok)
	require.Equal(t, "1", gpu.String())
}

func TestPipelineLargeOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

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
	t.Parallel()
	c := getPachClient(t)

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
			require.Equal(t, fi.SizeBytes, uint64(len(repos)))
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
				require.Equal(t, fi.SizeBytes, uint64(2))
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
				require.Equal(t, fi.SizeBytes, uint64(4))
			}
		}
	})
}

func TestIncrementalOverwritePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

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
	t.Parallel()
	c := getPachClient(t)

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
		w, err := c.PutFileSplitWriter(dataRepo, "master", "data", pfs.Delimiter_LINE, 0, 0)
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
	t.Parallel()
	c := getPachClient(t)

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
	t.Parallel()
	c := getPachClient(t)

	dataRepo := uniqueString("TestPipelineWithStats_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 100
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

	jobs, err := c.ListJob(pipeline, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
}

func TestIncrementalSharedProvenance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

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
	t.Parallel()
	c := getPachClient(t)
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
		jobs, err := c.ListJob(pipelineName, nil)
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
	podsInterface := k.Pods(api.NamespaceDefault)
	labelSelector, err := labels.Parse("suite=pachyderm")
	require.NoError(t, err)
	podList, err := podsInterface.List(
		api.ListOptions{
			LabelSelector: labelSelector,
		})
	require.NoError(t, err)
	for _, pod := range podList.Items {
		require.NoError(t, podsInterface.Delete(pod.Name, api.NewDeleteOptions(0)))
	}
	waitForReadiness(t)
}

func restartOne(t *testing.T) {
	k := getKubeClient(t)
	podsInterface := k.Pods(api.NamespaceDefault)
	labelSelector, err := labels.Parse("app=pachd")
	require.NoError(t, err)
	podList, err := podsInterface.List(
		api.ListOptions{
			LabelSelector: labelSelector,
		})
	require.NoError(t, err)
	require.NoError(t, podsInterface.Delete(podList.Items[rand.Intn(len(podList.Items))].Name, api.NewDeleteOptions(0)))
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

func waitForReadiness(t testing.TB) {
	k := getKubeClient(t)
	deployment := pachdDeployment(t)
	for {
		// This code is taken from
		// k8s.io/kubernetes/pkg/client/unversioned.ControllerHasDesiredReplicas
		// It used to call that fun ction but an update to the k8s library
		// broke it due to a type error.  We should see if we can go back to
		// using that code but I(jdoliner) couldn't figure out how to fanagle
		// the types into compiling.
		newDeployment, err := k.Extensions().Deployments(api.NamespaceDefault).Get(deployment.Name)
		require.NoError(t, err)
		if newDeployment.Status.ObservedGeneration >= deployment.Generation && newDeployment.Status.Replicas == newDeployment.Spec.Replicas {
			break
		}
		time.Sleep(time.Second * 5)
	}
	watch, err := k.Pods(api.NamespaceDefault).Watch(api.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": "pachd"}),
	})
	defer watch.Stop()
	require.NoError(t, err)
	readyPods := make(map[string]bool)
	for event := range watch.ResultChan() {
		ready, err := kube.PodRunningAndReady(event)
		require.NoError(t, err)
		if ready {
			pod, ok := event.Object.(*api.Pod)
			if !ok {
				t.Fatal("event.Object should be an object")
			}
			readyPods[pod.Name] = true
			if len(readyPods) == int(deployment.Spec.Replicas) {
				break
			}
		}
	}
}

func pipelineRc(t testing.TB, pipelineInfo *pps.PipelineInfo) (*api.ReplicationController, error) {
	k := getKubeClient(t)
	rc := k.ReplicationControllers(api.NamespaceDefault)
	return rc.Get(ppsserver.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version))
}

func pachdDeployment(t testing.TB) *extensions.Deployment {
	k := getKubeClient(t)
	result, err := k.Extensions().Deployments(api.NamespaceDefault).Get("pachd")
	require.NoError(t, err)
	return result
}

// scalePachd scales the number of pachd nodes up or down.
// If up is true, then the number of nodes will be within (n, 2n]
// If up is false, then the number of nodes will be within [1, n)
func scalePachdRandom(t testing.TB, up bool) {
	pachdRc := pachdDeployment(t)
	originalReplicas := pachdRc.Spec.Replicas
	for {
		if up {
			pachdRc.Spec.Replicas = originalReplicas + int32(rand.Intn(int(originalReplicas))+1)
		} else {
			pachdRc.Spec.Replicas = int32(rand.Intn(int(originalReplicas)-1) + 1)
		}

		if pachdRc.Spec.Replicas != originalReplicas {
			break
		}
	}
	scalePachdN(t, int(pachdRc.Spec.Replicas))
}

// scalePachdN scales the number of pachd nodes to N
func scalePachdN(t testing.TB, n int) {
	k := getKubeClient(t)
	pachdDeployment := pachdDeployment(t)
	pachdDeployment.Spec.Replicas = int32(n)
	_, err := k.Extensions().Deployments(api.NamespaceDefault).Update(pachdDeployment)
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

func getKubeClient(t testing.TB) *kube.Client {
	var config *kube_client.Config
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	if host != "" {
		var err error
		config, err = kube_client.InClusterConfig()
		require.NoError(t, err)
	} else {
		config = &kube_client.Config{
			Host:     "http://0.0.0.0:8080",
			Insecure: false,
		}
	}
	k, err := kube.New(config)
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
			pachClient, err = client.NewFromAddress("0.0.0.0:30650")
		}
		require.NoError(t, err)
	})
	return pachClient
}

func uniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}
