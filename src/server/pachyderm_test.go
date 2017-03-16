package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfspretty "github.com/pachyderm/pachyderm/src/server/pfs/pretty"
	ppspretty "github.com/pachyderm/pachyderm/src/server/pps/pretty"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"

	"k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/restclient"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
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
	commit1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)))
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	require.NoError(t, c.SetBranch(dataRepo, commit1.ID, "master"))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 4,
		},
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/*",
		}},
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	for i := 0; i < 1000; i++ {
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

	commit1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	require.NoError(t, c.SetBranch(dataRepo, commit1.ID, "master"))

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
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/*",
		}},
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
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
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

	commit1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	require.NoError(t, c.SetBranch(dataRepo, commit1.ID, "master"))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		nil,
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/*",
		}},
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

	dirA := "dirA"
	dirB := "dirB"

	commit1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "dirA/file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "dirB/file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	require.NoError(t, c.SetBranch(dataRepo, commit1.ID, "master"))

	pipeline := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/%s/%s/file >> /pfs/out/file", dirA, dirA),
			fmt.Sprintf("cat /pfs/%s/%s/file >> /pfs/out/file", dirB, dirB),
		},
		nil,
		[]*pps.PipelineInput{{
			Name: dirA,
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/dirA/*",
		}, {
			Name: dirB,
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/dirB/*",
		}},
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

func TestMultipleInputsFromTheSameRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	dataRepo := uniqueString("TestMultipleInputsFromTheSameRepo_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	branchA := "branchA"
	branchB := "branchB"

	commitA1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commitA1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commitA1.ID))
	require.NoError(t, c.SetBranch(dataRepo, commitA1.ID, branchA))

	commitB1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commitB1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commitB1.ID))
	require.NoError(t, c.SetBranch(dataRepo, commitB1.ID, branchB))

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
		[]*pps.PipelineInput{{
			Repo:   &pfs.Repo{Name: dataRepo},
			Branch: branchA,
			Glob:   "/*",
		}, {
			Repo:   &pfs.Repo{Name: dataRepo},
			Branch: branchB,
			Glob:   "/*",
		}},
		"",
		false,
	))

	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/%s/file >> /pfs/out/file", branchA),
			fmt.Sprintf("cat /pfs/%s/file >> /pfs/out/file", branchB),
		},
		nil,
		[]*pps.PipelineInput{{
			Name:   branchA,
			Repo:   &pfs.Repo{Name: dataRepo},
			Branch: branchA,
			Glob:   "/*",
		}, {
			Name:   branchB,
			Repo:   &pfs.Repo{Name: dataRepo},
			Branch: branchB,
			Glob:   "/*",
		}},
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commitA1, commitB1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\nfoo\n", buf.String())

	commitA2, err := c.StartCommit(dataRepo, branchA)
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commitA2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commitA2.ID))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commitA2, commitB1}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	buf.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\nbar\nfoo\n", buf.String())

	commitB2, err := c.StartCommit(dataRepo, branchB)
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commitB2.ID, "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commitB2.ID))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commitA2, commitB2}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	buf.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\nbar\nfoo\nbuzz\n", buf.String())

	commitA3, err := c.StartCommit(dataRepo, branchA)
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commitA3.ID, "file", strings.NewReader("poo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commitA3.ID))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commitA3, commitB2}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	buf.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\nbar\npoo\nfoo\nbuzz\n", buf.String())

	commitInfos, err = c.ListCommit(pipeline, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))

	// Now we delete the pipeline and re-create it.  The pipeline should
	// only process the heads of the branches.
	require.NoError(t, c.DeletePipeline(pipeline))
	require.NoError(t, c.DeleteRepo(pipeline, false))

	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cat /pfs/%s/file >> /pfs/out/file", branchA),
			fmt.Sprintf("cat /pfs/%s/file >> /pfs/out/file", branchB),
		},
		nil,
		[]*pps.PipelineInput{{
			Name:   branchA,
			Repo:   &pfs.Repo{Name: dataRepo},
			Branch: branchA,
			Glob:   "/*",
		}, {
			Name:   branchB,
			Repo:   &pfs.Repo{Name: dataRepo},
			Branch: branchB,
			Glob:   "/*",
		}},
		"",
		false,
	))

	commitIter, err = c.FlushCommit([]*pfs.Commit{commitA3, commitB2}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	buf.Reset()
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo\nbar\npoo\nfoo\nbuzz\n", buf.String())
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
//commit, err := c.StartCommit(dataRepo, "")
//require.NoError(t, err)
//err = c.SetBranch(dataRepo, commit.ID, "master")
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
//Strategy: pps.ParallelismSpec_CONSTANT,
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

// This test fails if you updated some static assets (such as doc/deployment/pipeline_spec.md)
// that are used in code but forgot to run:
// $ make assets
func TestAssets(t *testing.T) {
	assetPaths := []string{"doc/deployment/pipeline_spec.md"}

	for _, path := range assetPaths {
		doc, err := ioutil.ReadFile(filepath.Join(os.Getenv("GOPATH"), "src/github.com/pachyderm/pachyderm/", path))
		if err != nil {
			t.Fatal(err)
		}

		asset, err := pachyderm.Asset(path)
		if err != nil {
			t.Fatal(err)
		}

		require.Equal(t, doc, asset)
	}
}

func TestPipelineFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	dataRepo := uniqueString("TestPipelineFailure_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit := new(pfs.Commit)
	var err error
	numCommits := 10
	for i := 0; i < numCommits; i++ {
		commit, err = c.StartCommit(dataRepo, commit.ID)
		require.NoError(t, err)
		_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	}
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))

	pipeline := uniqueString("pipeline")
	errMsg := "error message"
	// This pipeline fails half the times
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("echo '%s'", errMsg),
			"exit $(($RANDOM % 2))",
		},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/*",
		}},
		"",
		false,
	))

	// Wait for the jobs to spawn
	time.Sleep(20 * time.Second)

	jobInfos, err := c.ListJob(pipeline, nil)
	require.NoError(t, err)
	require.Equal(t, numCommits, len(jobInfos))

	for _, jobInfo := range jobInfos {
		// Wait for the job to finish
		jobInfo, err := c.InspectJob(jobInfo.Job.ID, true)
		require.NoError(t, err)

		require.EqualOneOf(t, []interface{}{pps.JobState_JOB_SUCCESS, pps.JobState_JOB_FAILURE}, jobInfo.State)
		if jobInfo.State == pps.JobState_JOB_FAILURE {
			require.Equal(t, errMsg+"\n", jobInfo.Error)
		}
	}
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Lazy: true,
			Glob: "/*",
		}},
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: pipelineA},
			Glob: "/*",
			Lazy: true,
		}},
		"",
		false,
	))

	commit1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit1.ID, "master"))
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit1.ID)}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, commitIter)

	jobInfos, err := c.ListJob(pipelineA, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	require.Equal(t, true, jobInfos[0].Inputs[0].Lazy)
	jobInfos, err = c.ListJob(pipelineB, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	require.Equal(t, true, jobInfos[0].Inputs[0].Lazy)
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
				Strategy: pps.ParallelismSpec_CONSTANT,
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
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(aRepo),
			Glob: "/*",
		}},
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(aRepo),
			Glob: "/*",
		}, {
			Repo: client.NewRepo(bPipeline),
			Glob: "/*",
		}},
		"",
		false,
	))
	// commit to aRepo
	commit1, err := c.StartCommit(aRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(aRepo, commit1.ID, "master"))
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
//Strategy: pps.ParallelismSpec_CONSTANT,
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
//Strategy: pps.ParallelismSpec_CONSTANT,
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
				Strategy: pps.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			[]*pps.PipelineInput{{
				Repo: client.NewRepo(repo),
				Glob: "/*",
			}},
			"",
			false,
		))
	}

	for i := 0; i < 10; i++ {
		var commit *pfs.Commit
		var err error
		if i == 0 {
			commit, err = c.StartCommit(sourceRepo, "")
			require.NoError(t, err)
			c.SetBranch(sourceRepo, commit.ID, "master")
		} else {
			commit, err = c.StartCommit(sourceRepo, "master")
			require.NoError(t, err)
		}
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(repo),
			Glob: "/*",
		}},
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
	commit, err := c.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(repo, commit.ID, "master"))
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
				Strategy: pps.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			[]*pps.PipelineInput{{
				Repo: client.NewRepo(repo),
				Glob: "/*",
			}},
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
	require.NoError(t, c.DeletePipeline(pipeline))
	time.Sleep(5 * time.Second)
	createPipeline()
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(repo),
			Glob: "/*",
		}},
		"",
		false,
	))

	pipelineInfo, err := c.InspectPipeline(pipeline)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_RUNNING, pipelineInfo.State)

	require.NoError(t, c.StopPipeline(pipeline))
	time.Sleep(5 * time.Second)

	pipelineInfo, err = c.InspectPipeline(pipeline)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_STOPPED, pipelineInfo.State)

	require.NoError(t, c.StartPipeline(pipeline))
	time.Sleep(5 * time.Second)

	pipelineInfo, err = c.InspectPipeline(pipeline)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineState_PIPELINE_RUNNING, pipelineInfo.State)
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(repo),
			Glob: "/*",
		}},
		"",
		false,
	))

	// Trigger a job by creating a commit
	commit, err := c.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(repo, commit.ID, "master"))
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

func TestDeleteAfterMembershipChange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	test := func(up bool) {
		repo := uniqueString("TestDeleteAfterMembershipChange")
		c := getPachClient(t)
		require.NoError(t, c.CreateRepo(repo))
		commit, err := c.StartCommit(repo, "")
		require.NoError(t, err)
		require.NoError(t, c.SetBranch(repo, commit.ID, "master"))
		require.NoError(t, c.FinishCommit(repo, "master"))
		scalePachd(t, up)
		c = getUsablePachClient(t)
		require.NoError(t, c.DeleteRepo(repo, false))
	}
	test(true)
	test(false)
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
//Strategy: pps.ParallelismSpec_CONSTANT,
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

func TestAcceptReturnCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	dataRepo := uniqueString("TestAcceptReturnCode")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	job, err := c.PpsAPIClient.CreateJob(
		context.Background(),
		&pps.CreateJobRequest{
			Transform: &pps.Transform{
				Cmd:              []string{"sh"},
				Stdin:            []string{"exit 1"},
				AcceptReturnCode: []int64{1},
			},
			Inputs: []*pps.JobInput{{
				Name:   dataRepo,
				Commit: commit,
				Glob:   "/*",
			}},
			OutputBranch: "master",
		},
	)
	require.NoError(t, err)
	inspectJobRequest := &pps.InspectJobRequest{
		Job:        job,
		BlockState: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS.String(), jobInfo.State.String())
}

func TestRestartAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/*",
		}},
		"",
		false,
	))
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
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

func TestRestartOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/",
		}},
		"",
		false,
	))
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
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
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/*",
		}},
		"",
		false,
	))
	// Do a commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/",
		}},
		"",
		false,
	))
	// Do commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(dataRepo),
			Glob: "/*",
		}},
		"",
		false,
	))
	// Do commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{
			{
				Repo: &pfs.Repo{Name: repo},
				Glob: "/",
			},
		},
		"",
		false,
	))
	err := c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{
			{
				Repo: &pfs.Repo{Name: repo},
				Glob: "/",
			},
		},
		"",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "pipeline .*? already exists", err.Error())
}

//func TestUpdatePipeline(t *testing.T) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}
//t.Parallel()
//c := getPachClient(t)
//fmt.Println("BP1")
//// create repos
//dataRepo := uniqueString("TestUpdatePipeline_data")
//require.NoError(t, c.CreateRepo(dataRepo))
//// create 2 pipelines
//pipelineName := uniqueString("pipeline")
//require.NoError(t, c.CreatePipeline(
//pipelineName,
//"",
//[]string{"bash"},
//[]string{fmt.Sprintf(`
//cat /pfs/%s/file1 >>/pfs/out/file
//`, dataRepo)},
//&pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*pps.PipelineInput{{
//Repo: client.NewRepo(dataRepo),
//Glob: "/*",
//}},
//"",
//false,
//))
//fmt.Println("BP2")
//pipeline2Name := uniqueString("pipeline2")
//require.NoError(t, c.CreatePipeline(
//pipeline2Name,
//"",
//[]string{"bash"},
//[]string{fmt.Sprintf(`
//cat /pfs/%s/file >>/pfs/out/file
//`, pipelineName)},
//&pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*pps.PipelineInput{{
//Repo: client.NewRepo(pipelineName),
//Glob: "/*",
//}},
//"",
//false,
//))
//fmt.Println("BP3")
//// Do first commit to repo
//var commit *pfs.Commit
//var err error
//for i := 0; i < 2; i++ {
//if i == 0 {
//commit, err = c.StartCommit(dataRepo, "")
//require.NoError(t, err)
//require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
//} else {
//commit, err = c.StartCommit(dataRepo, "master")
//require.NoError(t, err)
//}
//_, err = c.PutFile(dataRepo, commit.ID, "file1", strings.NewReader("file1\n"))
//_, err = c.PutFile(dataRepo, commit.ID, "file2", strings.NewReader("file2\n"))
//_, err = c.PutFile(dataRepo, commit.ID, "file3", strings.NewReader("file3\n"))
//require.NoError(t, err)
//require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
//}
//fmt.Println("BP4")
//commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
//require.NoError(t, err)
//commitInfos := collectCommitInfos(t, commitIter)
//require.Equal(t, 2, len(commitInfos))
//// only care about non-provenance commits
//commitInfos = commitInfos[1:]
//for _, commitInfo := range commitInfos {
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, &buffer))
//require.Equal(t, "file1\nfile1\n", buffer.String())
//}
//fmt.Println("BP5")

//outputRepoCommitInfos, err := c.ListCommit(pipelineName, "", "", 0)
//require.NoError(t, err)
//require.Equal(t, 2, len(outputRepoCommitInfos))

//// Update the pipeline to look at file2
//require.NoError(t, c.CreatePipeline(
//pipelineName,
//"",
//[]string{"bash"},
//[]string{fmt.Sprintf(`
//cat /pfs/%s/file2 >>/pfs/out/file
//`, dataRepo)},
//&pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*pps.PipelineInput{{
//Repo: &pfs.Repo{Name: dataRepo},
//Glob: "/*",
//}},
//"",
//true,
//))
//pipelineInfo, err := c.InspectPipeline(pipelineName)
//require.NoError(t, err)
//require.NotNil(t, pipelineInfo.CreatedAt)
//fmt.Println("BP6")
//commitIter, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
//require.NoError(t, err)
//commitInfos = collectCommitInfos(t, commitIter)
//require.Equal(t, 2, len(commitInfos))
//// only care about non-provenance commits
//commitInfos = commitInfos[1:]
//fmt.Println("BP7")
//for _, commitInfo := range commitInfos {
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, &buffer))
//require.Equal(t, "file2\nfile2\n", buffer.String())
//}
//outputRepoCommitInfos, err = c.ListCommit(pipelineName, "", "", 0)
//require.NoError(t, err)
//require.Equal(t, 3, len(outputRepoCommitInfos))

//// Update the pipeline to look at file3
//require.NoError(t, c.CreatePipeline(
//pipelineName,
//"",
//[]string{"bash"},
//[]string{fmt.Sprintf(`
//cat /pfs/%s/file3 >>/pfs/out/file
//`, dataRepo)},
//&pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*pps.PipelineInput{{
//Repo: &pfs.Repo{Name: dataRepo},
//Glob: "/*",
//}},
//"",
//true,
//))
//commitIter, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
//require.NoError(t, err)
//commitInfos = collectCommitInfos(t, commitIter)
//require.Equal(t, 3, len(commitInfos))
//// only care about non-provenance commits
//commitInfos = commitInfos[1:]
//for _, commitInfo := range commitInfos {
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, &buffer))
//require.Equal(t, "file3\nfile3\n", buffer.String())
//}
//outputRepoCommitInfos, err = c.ListCommit(pipelineName, "", "", 0)
//require.NoError(t, err)
//require.Equal(t, 12, len(outputRepoCommitInfos))
//// Expect real commits to still be 1
//outputRepoCommitInfos, err = c.ListCommit(pipelineName, "", "", 0)
//require.NoError(t, err)
//require.Equal(t, 2, len(outputRepoCommitInfos))

//commitInfos, _ = c.ListCommit(pipelineName, "", "", 0)
//// Do an update that shouldn't cause archiving
//_, err = c.PpsAPIClient.CreatePipeline(
//context.Background(),
//&pps.CreatePipelineRequest{
//Pipeline: client.NewPipeline(pipelineName),
//Transform: &pps.Transform{
//Cmd: []string{"bash"},
//Stdin: []string{fmt.Sprintf(`
//cat /pfs/%s/file3 >>/pfs/out/file
//`, dataRepo)},
//},
//ParallelismSpec: &pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 2,
//},
//Inputs: []*pps.PipelineInput{{
//Repo: &pfs.Repo{Name: dataRepo},
//Glob: "/*",
//}},
//OutputBranch: "",
//Update:       true,
//})
//require.NoError(t, err)
//commitIter, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
//require.NoError(t, err)
//commitInfos = collectCommitInfos(t, commitIter)
//require.Equal(t, 3, len(commitInfos))
//// only care about non-provenance commits
//commitInfos = commitInfos[1:]
//for _, commitInfo := range commitInfos {
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, &buffer))
//require.Equal(t, "file3\nfile3\n", buffer.String())
//}
//commitInfos, err = c.ListCommit(pipelineName, "", "", 0)
//require.NoError(t, err)
//require.Equal(t, 12, len(commitInfos))
//// Expect real commits to still be 1
//outputRepoCommitInfos, err = c.ListCommit(pipelineName, "", "", 0)
//require.NoError(t, err)
//require.Equal(t, 2, len(outputRepoCommitInfos))
//}

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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/*",
		}},
		"",
		false,
	))
	require.NoError(t, c.StopPipeline(pipelineName))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit1.ID, "master"))
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
				Strategy: pps.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			Inputs: []*pps.PipelineInput{{
				Repo: &pfs.Repo{Name: dataRepo},
				Glob: "/*",
			}},
		})
	require.NoError(t, err)
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{
			{
				Repo: client.NewRepo(dataRepo),
				Glob: "/*",
			},
		},
		"",
		false,
	))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit1.ID, "master"))
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

	aCommit, err := c.StartCommit(aRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(aRepo, aCommit.ID, "master"))
	_, err = c.PutFile(aRepo, "master", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, "master"))

	dCommit, err := c.StartCommit(dRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dRepo, dCommit.ID, "master"))
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(aRepo),
			Glob: "/",
		}},
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(bPipeline),
			Glob: "/",
		}, {
			Repo: client.NewRepo(dRepo),
			Glob: "/",
		}},
		"",
		false,
	))
	resultIter, err := c.FlushCommit([]*pfs.Commit{aCommit, dCommit}, nil)
	require.NoError(t, err)
	results := collectCommitInfos(t, resultIter)
	require.Equal(t, 1, len(results))
}

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

	aCommit, err := c.StartCommit(aRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(aRepo, aCommit.ID, "master"))
	_, err = c.PutFile(aRepo, "master", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, "master"))

	eCommit, err := c.StartCommit(eRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(eRepo, eCommit.ID, "master"))
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(aRepo),
			Glob: "/",
		}},
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(bPipeline),
			Glob: "/",
		}, {
			Repo: client.NewRepo(eRepo),
			Glob: "/",
		}},
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(cPipeline),
			Glob: "/",
		}},
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

func collectCommitInfos(t *testing.T, commitInfoIter client.CommitInfoIterator) []*pfs.CommitInfo {
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
	// Test Constant strategy
	parellelism, err := pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Strategy: pps.ParallelismSpec_CONSTANT,
		Constant: 7,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(7), parellelism)

	// Coefficient == 1 (basic test)
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Strategy:    pps.ParallelismSpec_COEFFICIENT,
		Coefficient: 1,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)

	// Coefficient > 1
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Strategy:    pps.ParallelismSpec_COEFFICIENT,
		Coefficient: 2,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(2), parellelism)

	// Make sure we start at least one worker
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Strategy:    pps.ParallelismSpec_COEFFICIENT,
		Coefficient: 0.1,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)

	// Test 0-initialized JobSpec
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{})
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)

	// Test nil JobSpec
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), nil)
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Name: dataRepo,
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/",
		}},
		"",
		false,
	))

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

	// Now delete the corresponding job
	jobInfos, err := c.ListJob(pipelineName, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	err = c.DeleteJob(jobInfos[0].Job.ID)
	require.NoError(t, err)
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

func waitForReadiness(t *testing.T) {
	k := getKubeClient(t)
	rc := pachdRc(t)
	for {
		// This code is taken from
		// k8s.io/kubernetes/pkg/client/unversioned.ControllerHasDesiredReplicas
		// It used to call that fun ction but an update to the k8s library
		// broke it due to a type error.  We should see if we can go back to
		// using that code but I(jdoliner) couldn't figure out how to fanagle
		// the types into compiling.
		newRc, err := k.ReplicationControllers(api.NamespaceDefault).Get(rc.Name)
		require.NoError(t, err)
		if newRc.Status.ObservedGeneration >= rc.Generation && newRc.Status.Replicas == newRc.Spec.Replicas {
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
			if len(readyPods) == int(rc.Spec.Replicas) {
				break
			}
		}
	}
}

func pachdRc(t *testing.T) *api.ReplicationController {
	k := getKubeClient(t)
	rc := k.ReplicationControllers(api.NamespaceDefault)
	result, err := rc.Get("pachd")
	require.NoError(t, err)
	return result
}

// scalePachd scales the number of pachd nodes up or down.
// If up is true, then the number of nodes will be within (n, 2n]
// If up is false, then the number of nodes will be within [1, n)
func scalePachd(t *testing.T, up bool) {
	k := getKubeClient(t)
	pachdRc := pachdRc(t)
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
	fmt.Printf("scaling pachd to %d replicas\n", pachdRc.Spec.Replicas)
	rc := k.ReplicationControllers(api.NamespaceDefault)
	_, err := rc.Update(pachdRc)
	require.NoError(t, err)
	waitForReadiness(t)
	// Unfortunately, even when all pods are ready, the cluster membership
	// protocol might still be running, thus PFS API calls might fail.  So
	// we wait a little bit for membership to stablize.
	time.Sleep(15 * time.Second)
}

func getKubeClient(t *testing.T) *kube.Client {
	config := &kube_client.Config{
		Host:     "http://0.0.0.0:8080",
		Insecure: false,
	}
	k, err := kube.New(config)
	require.NoError(t, err)
	return k
}

func getPachClient(t testing.TB) *client.APIClient {
	client, err := client.NewFromAddress("0.0.0.0:30650")
	require.NoError(t, err)
	return client
}

func uniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}
