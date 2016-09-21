package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	pfspretty "github.com/pachyderm/pachyderm/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	ppspretty "github.com/pachyderm/pachyderm/src/server/pps/pretty"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/restclient"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	kube_labels "k8s.io/kubernetes/pkg/labels"
)

const (
	NUMFILES = 25
	KB       = 1024
)

func TestJob(t *testing.T) {
	t.Parallel()
	testJob(t, 4)
}

func TestJobNoShard(t *testing.T) {
	t.Parallel()
	testJob(t, 0)
}

func testJob(t *testing.T, shards int) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := getPachClient(t)
	dataRepo := uniqueString("TestJob_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	fileContent := "foo\n"
	// We want to create lots of files so that each parallel job will be
	// started with some files
	numFiles := shards*100 + 100
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit.ID, fmt.Sprintf("file-%d", i), strings.NewReader(fileContent))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	job, err := c.CreateJob(
		"",
		[]string{"bash"},
		[]string{fmt.Sprintf("cp %s %s", path.Join("/pfs", dataRepo, "*"), "/pfs/out")},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: uint64(shards),
		},
		[]*ppsclient.JobInput{{
			Commit: commit,
			Method: client.ReduceMethod,
		}},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:        job,
		BlockState: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel() //cleanup resources
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS.String(), jobInfo.State.String())
	parellelism, err := pps_server.GetExpectedNumWorkers(getKubeClient(t), jobInfo.ParallelismSpec)
	require.NoError(t, err)
	require.True(t, parellelism > 0)
	commitInfo, err := c.InspectCommit(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	require.NoError(t, err)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	require.NotNil(t, jobInfo.Started)
	require.NotNil(t, jobInfo.Finished)
	require.True(t, prototime.TimestampToTime(jobInfo.Finished).After(prototime.TimestampToTime(jobInfo.Started)))
	for i := 0; i < numFiles; i++ {
		var buffer bytes.Buffer
		require.NoError(t, c.GetFile(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, fmt.Sprintf("file-%d", i), 0, 0, "", false, nil, &buffer))
		require.Equal(t, fileContent, buffer.String())
	}
}

func TestPachCommitIdEnvVarInJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	shards := 0
	c := getPachClient(t)
	repos := []string{
		uniqueString("TestJob_FriarTuck"),
		uniqueString("TestJob_RobinHood"),
	}

	var commits []*pfsclient.Commit

	for _, repo := range repos {
		require.NoError(t, c.CreateRepo(repo))
		commit, err := c.StartCommit(repo, "master")
		require.NoError(t, err)
		fileContent := "foo\n"

		_, err = c.PutFile(repo, commit.ID, "file", strings.NewReader(fileContent))
		require.NoError(t, err)

		require.NoError(t, c.FinishCommit(repo, commit.ID))
		commits = append(commits, commit)
	}

	job, err := c.CreateJob(
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("echo $PACH_%v_COMMIT_ID > /pfs/out/input-id-%v", pps_server.RepoNameToEnvString(repos[0]), repos[0]),
			fmt.Sprintf("echo $PACH_%v_COMMIT_ID > /pfs/out/input-id-%v", pps_server.RepoNameToEnvString(repos[1]), repos[1]),
		},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: uint64(shards),
		},
		[]*ppsclient.JobInput{
			{
				Commit: commits[0],
				Method: client.ReduceMethod,
			},
			{
				Commit: commits[1],
				Method: client.ReduceMethod,
			},
		},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:        job,
		BlockState: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS.String(), jobInfo.State.String())
	parallelism, err := pps_server.GetExpectedNumWorkers(getKubeClient(t), jobInfo.ParallelismSpec)
	require.NoError(t, err)
	require.True(t, parallelism > 0)
	commitInfo, err := c.InspectCommit(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	require.NoError(t, err)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, fmt.Sprintf("input-id-%v", repos[0]), 0, 0, "", false, nil, &buffer))
	require.Equal(t, jobInfo.Inputs[0].Commit.ID, strings.TrimSpace(buffer.String()))

	buffer.Reset()
	require.NoError(t, c.GetFile(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, fmt.Sprintf("input-id-%v", repos[1]), 0, 0, "", false, nil, &buffer))
	require.Equal(t, jobInfo.Inputs[1].Commit.ID, strings.TrimSpace(buffer.String()))
}

func TestDuplicatedJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	dataRepo := uniqueString("TestDuplicatedJob_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	pipelineName := uniqueString("TestDuplicatedJob_pipeline")
	_, err = c.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfsclient.CreateRepoRequest{
			Repo:       client.NewRepo(pipelineName),
			Provenance: []*pfsclient.Repo{client.NewRepo(dataRepo)},
		},
	)
	require.NoError(t, err)

	cmd := []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"}
	// Now we manually create the same job
	req := &ppsclient.CreateJobRequest{
		Transform: &ppsclient.Transform{
			Cmd: cmd,
		},
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
		Inputs: []*ppsclient.JobInput{{
			Commit: commit,
		}},
	}

	job1, err := c.PpsAPIClient.CreateJob(context.Background(), req)
	require.NoError(t, err)

	job2, err := c.PpsAPIClient.CreateJob(context.Background(), req)
	require.NoError(t, err)

	require.Equal(t, job1, job2)

	req.Force = true
	job3, err := c.PpsAPIClient.CreateJob(context.Background(), req)
	require.NoError(t, err)
	require.NotEqual(t, job1, job3)

	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:        job1,
		BlockState: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, fileContent, buffer.String())
}

func TestLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	job, err := c.CreateJob(
		"",
		[]string{"echo", "foo"},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 4,
		},
		[]*ppsclient.JobInput{},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:        job,
		BlockState: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	_, err = c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	// TODO we Sleep here because even though the job has completed kubernetes
	// might not have even noticed the container was created yet
	time.Sleep(10 * time.Second)
	var buffer bytes.Buffer
	require.NoError(t, c.GetLogs(job.ID, &buffer))
	require.Equal(t, "0 | foo\n1 | foo\n2 | foo\n3 | foo\n", buffer.String())

	// Should get an error if the job does not exist
	require.YesError(t, c.GetLogs("nonexistent", &buffer))
}

func TestGrep(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	dataRepo := uniqueString("TestGrep_data")
	c := getPachClient(t)
	require.NoError(t, c.CreateRepo(dataRepo))
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, err = c.PutFile(dataRepo, commit.ID, fmt.Sprintf("file%d", i), strings.NewReader("foo\nbar\nfizz\nbuzz\n"))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	job1, err := c.CreateJob(
		"",
		[]string{"bash"},
		[]string{fmt.Sprintf("grep foo /pfs/%s/* >/pfs/out/foo", dataRepo)},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.JobInput{{Commit: commit}},
		"",
	)
	require.NoError(t, err)
	job2, err := c.CreateJob(
		"",
		[]string{"bash"},
		[]string{fmt.Sprintf("grep foo /pfs/%s/* >/pfs/out/foo", dataRepo)},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 4,
		},
		[]*ppsclient.JobInput{{Commit: commit}},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:        job1,
		BlockState: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	job1Info, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	inspectJobRequest.Job = job2
	job2Info, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	repo1Info, err := c.InspectRepo(job1Info.OutputCommit.Repo.Name)
	require.NoError(t, err)
	repo2Info, err := c.InspectRepo(job2Info.OutputCommit.Repo.Name)
	require.NoError(t, err)
	require.Equal(t, repo1Info.SizeBytes, repo2Info.SizeBytes)
}

func TestJobLongOutputLine(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)
	job, err := c.CreateJob(
		"",
		[]string{"sh"},
		[]string{"yes | tr -d '\\n' | head -c 1000000 > /pfs/out/file"},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.JobInput{},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:        job,
		BlockState: true,
	}
	jobInfo, err := c.PpsAPIClient.InspectJob(context.Background(), inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS.String(), jobInfo.State.String())
}

func TestPipeline(t *testing.T) {

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
	outRepo := ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{
			Repo:   &pfsclient.Repo{Name: dataRepo},
			Method: client.MapMethod,
		}},
		false,
	))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{&pfsclient.Commit{
			Repo: outRepo,
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	// Do second commit to repo
	commit2, err := c.StartCommit(dataRepo, commit1.ID)
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))
	listCommitRequest = &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{outCommits[0].Commit},
		CommitType:  pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:       true,
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	require.NotNil(t, listCommitResponse.CommitInfo[0].ParentCommit)
	require.Equal(t, outCommits[0].Commit.ID, listCommitResponse.CommitInfo[0].ParentCommit.ID)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "bar\n", buffer.String())

	require.NoError(t, c.DeletePipeline(pipelineName))

	pipelineInfos, err := c.PpsAPIClient.ListPipeline(context.Background(), &ppsclient.ListPipelineRequest{})
	require.NoError(t, err)
	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		require.True(t, pipelineInfo.Pipeline.Name != pipelineName)
	}

	// Do third commit to repo; this time pipeline should not run since it's been deleted
	commit3, err := c.StartCommit(dataRepo, commit2.ID)
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit3.ID, "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit3.ID))

	// We will sleep a while to wait for the pipeline to actually get cancelled
	// Also if the pipeline didn't get cancelled (due to a bug), we sleep a while
	// to let the pipeline commit
	time.Sleep(5 * time.Second)
	listCommitRequest = &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{&pfsclient.Commit{
			Repo: outRepo,
		}},
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	// there should only be two commits in the pipeline
	require.Equal(t, 2, len(listCommitResponse.CommitInfo))
}

func TestPipelineWithEmptyInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
	// create repo
	dataRepo := uniqueString("data")
	require.NoError(t, c.CreateRepo(dataRepo))

	// create a pipeline that doesn't run with empty commits
	pipelineName := uniqueString("pipeline")
	outRepo := ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{
			"echo foo > /pfs/out/file",
		},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{
			Repo:     &pfsclient.Repo{Name: dataRepo},
			RunEmpty: false,
		}},
		false,
	))
	// Add first empty commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: outRepo,
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	require.Equal(t, 0, int(outCommits[0].SizeBytes))
	// An empty job should've been created
	jobInfos, err := c.ListJob(pipelineName, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	require.Equal(t, ppsclient.JobState_JOB_EMPTY, jobInfos[0].State)

	// Make another empty commit in the input repo
	// The output commit should have the previous output commit as its parent
	parentOutputCommit := outCommits[0].Commit
	commit2, err := c.StartCommit(dataRepo, commit1.ID)
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))
	listCommitRequest = &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{parentOutputCommit},
		CommitType:  pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:       true,
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	require.Equal(t, 0, int(outCommits[0].SizeBytes))
	require.Equal(t, parentOutputCommit.ID, outCommits[0].ParentCommit.ID)
	jobInfos, err = c.ListJob(pipelineName, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
	require.Equal(t, ppsclient.JobState_JOB_EMPTY, jobInfos[1].State)

	// create a pipeline that runs with empty commits
	dataRepo = uniqueString("data")
	require.NoError(t, c.CreateRepo(dataRepo))
	pipelineName = uniqueString("pipeline")

	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	outRepo = ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{
			"echo foo > /pfs/out/file",
		},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{
			Repo:     &pfsclient.Repo{Name: dataRepo},
			RunEmpty: true,
		}},
		false,
	))
	listCommitRequest = &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: outRepo,
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	require.Equal(t, len("foo\n"), int(outCommits[0].SizeBytes))
	jobInfos, err = c.ListJob(pipelineName, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
}

func TestPipelineWithTooMuchParallelism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestPipelineWithTooMuchParallelism_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	outRepo := ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	// This pipeline will fail if any pod sees empty input, since cp won't
	// be able to find the file.
	// We have parallelism set to 3 so that if we actually start 3 pods,
	// which would be a buggy behavior, some jobs don't see any files
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 3,
		},
		[]*ppsclient.PipelineInput{{
			Repo: &pfsclient.Repo{Name: dataRepo},
			// Use reduce method so only one pod gets the file
			Method: client.ReduceMethod,
		}},
		false,
	))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: outRepo,
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		ctx,
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	require.Equal(t, false, outCommits[0].Cancelled)
}

func TestPipelineWithNoInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
	// create pipeline
	pipelineName := uniqueString("pipeline")
	outRepo := ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"sh"},
		[]string{
			"NEW_UUID=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)",
			"echo foo > /pfs/out/$NEW_UUID",
		},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 3,
		},
		nil,
		false,
	))

	// Manually trigger the pipeline
	job, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})
	require.NoError(t, err)

	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:        job,
		BlockState: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS.String(), jobInfo.State.String())
	parallelism, err := pps_server.GetExpectedNumWorkers(getKubeClient(t), jobInfo.ParallelismSpec)
	require.NoError(t, err)
	require.Equal(t, 3, int(parallelism))

	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: outRepo,
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		ctx,
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	fileInfos, err := c.ListFile(outRepo.Name, outCommits[0].Commit.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfos))

	// Make sure that each job gets a different ID
	job2, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})
	require.NoError(t, err)
	require.True(t, job.ID != job2.ID)
}

func TestPipelineThatWritesToOneFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
	// create pipeline
	pipelineName := uniqueString("pipeline")
	outRepo := ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"sh"},
		[]string{
			"dd if=/dev/zero of=/pfs/out/file bs=10 count=1",
		},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 3,
		},
		nil,
		false,
	))

	// Manually trigger the pipeline
	_, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})
	require.NoError(t, err)

	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: outRepo,
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		ctx,
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, 30, buffer.Len())
}

func TestPipelineThatOverwritesFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
	// create pipeline
	pipelineName := uniqueString("pipeline")
	outRepo := ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"sh"},
		[]string{
			"echo foo > /pfs/out/file",
		},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 3,
		},
		nil,
		false,
	))

	// Manually trigger the pipeline
	job, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})
	require.NoError(t, err)

	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: outRepo,
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		ctx,
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())

	// Manually trigger the pipeline
	_, err = c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
		ParentJob: job,
	})
	require.NoError(t, err)

	listCommitRequest = &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{outCommits[0].Commit},
		CommitType:  pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:       true,
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		ctx,
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer2 bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer2))
	// we expect only 3 foos here because > _overwrites_ rather than appending.
	// Appending is done with >>.
	require.Equal(t, "foo\nfoo\nfoo\n", buffer2.String())
}

func TestPipelineThatAppendsToFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
	// create pipeline
	pipelineName := uniqueString("pipeline")
	outRepo := ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"sh"},
		[]string{
			"echo foo >> /pfs/out/file",
		},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 3,
		},
		nil,
		false,
	))

	// Manually trigger the pipeline
	job, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})
	require.NoError(t, err)

	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: outRepo,
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		ctx,
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())

	// Manually trigger the pipeline
	_, err = c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
		ParentJob: job,
	})
	require.NoError(t, err)

	listCommitRequest = &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{outCommits[0].Commit},
		CommitType:  pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:       true,
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		ctx,
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer2 bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer2))
	require.Equal(t, "foo\nfoo\nfoo\nfoo\nfoo\nfoo\n", buffer2.String())
}

func TestRemoveAndAppend(t *testing.T) {
	testParellelRemoveAndAppend(t, 1)
}

func TestParellelRemoveAndAppend(t *testing.T) {
	// This test does not pass on Travis which is why it's skipped right now As
	// soon as we have a hypothesis for why this fails on travis but not
	// locally we should un skip this test and try to fix it.
	testParellelRemoveAndAppend(t, 3)
}

func testParellelRemoveAndAppend(t *testing.T, parallelism int) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel() //cleanup resources

	job1, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Transform: &ppsclient.Transform{
			Cmd: []string{"sh"},
			Stdin: []string{
				"echo foo > /pfs/out/file",
			},
		},
		ParallelismSpec: &ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: uint64(parallelism),
		},
	})
	require.NoError(t, err)

	inspectJobRequest1 := &ppsclient.InspectJobRequest{
		Job:        job1,
		BlockState: true,
	}
	jobInfo1, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest1)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS, jobInfo1.State)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(jobInfo1.OutputCommit.Repo.Name, jobInfo1.OutputCommit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, strings.Repeat("foo\n", parallelism), buffer.String())

	job2, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Transform: &ppsclient.Transform{
			Cmd: []string{"sh"},
			Stdin: []string{
				"unlink /pfs/out/file && echo bar > /pfs/out/file",
			},
		},
		ParallelismSpec: &ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: uint64(parallelism),
		},
		ParentJob: job1,
	})
	require.NoError(t, err)

	inspectJobRequest2 := &ppsclient.InspectJobRequest{
		Job:        job2,
		BlockState: true,
	}
	jobInfo2, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest2)
	require.NoError(t, err)
	c.GetLogs(jobInfo2.Job.ID, os.Stdout)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS, jobInfo2.State)

	var buffer2 bytes.Buffer
	require.NoError(t, c.GetFile(jobInfo2.OutputCommit.Repo.Name, jobInfo2.OutputCommit.ID, "file", 0, 0, "", false, nil, &buffer2))
	require.Equal(t, strings.Repeat("bar\n", parallelism), buffer2.String())
}

func TestWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	seed := time.Now().UnixNano()
	require.NoError(t, workload.RunWorkload(c, rand.New(rand.NewSource(seed)), 100))
}

func TestSharding(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	repo := uniqueString("TestSharding")
	c := getPachClient(t)
	err := c.CreateRepo(repo)
	require.NoError(t, err)
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	var wg sync.WaitGroup
	for i := 0; i < NUMFILES; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			rand := rand.New(rand.NewSource(int64(i)))
			_, err = c.PutFile(repo, commit.ID, fmt.Sprintf("file%d", i), workload.NewReader(rand, KB))
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	err = c.FinishCommit(repo, commit.ID)
	require.NoError(t, err)
	wg = sync.WaitGroup{}
	for i := 0; i < NUMFILES; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			var buffer1Shard bytes.Buffer
			var buffer4Shard bytes.Buffer
			shard := &pfsclient.Shard{FileModulus: 1, BlockModulus: 1}
			err := c.GetFile(repo, commit.ID,
				fmt.Sprintf("file%d", i), 0, 0, "", false, shard, &buffer1Shard)
			require.NoError(t, err)
			shard.BlockModulus = 4
			for blockNumber := uint64(0); blockNumber < 4; blockNumber++ {
				shard.BlockNumber = blockNumber
				c.GetFile(repo, commit.ID, fmt.Sprintf("file%d", i), 0, 0, "", false, shard, &buffer4Shard)
			}
			require.Equal(t, buffer1Shard.Len(), buffer4Shard.Len())
		}()
	}
	wg.Wait()
}

func TestFromCommit(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	repo := uniqueString("TestFromCommit")
	c := getPachClient(t)
	seed := time.Now().UnixNano()
	rand := rand.New(rand.NewSource(seed))
	err := c.CreateRepo(repo)
	require.NoError(t, err)
	commit1, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit1.ID, "file", workload.NewReader(rand, KB))
	require.NoError(t, err)
	err = c.FinishCommit(repo, commit1.ID)
	require.NoError(t, err)
	commit2, err := c.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit2.ID, "file", workload.NewReader(rand, KB))
	require.NoError(t, err)
	err = c.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(repo, commit2.ID, "file", 0, 0, commit1.ID, false, nil, &buffer))
	require.Equal(t, buffer.Len(), KB)
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(repo, commit2.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, buffer.Len(), 2*KB)
}

func TestSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)
	repo := uniqueString("TestSimple")
	require.NoError(t, c.CreateRepo(repo))
	commit1, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit1.ID))
	commitInfos, err := c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{repo},
	}}, nil, client.CommitTypeNone, pfsclient.CommitStatus_NORMAL, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := c.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = c.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(repo, commit2.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestPipelineWithMultipleInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	inputRepo1 := uniqueString("inputRepo")
	require.NoError(t, c.CreateRepo(inputRepo1))
	inputRepo2 := uniqueString("inputRepo")
	require.NoError(t, c.CreateRepo(inputRepo2))

	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{fmt.Sprintf(`
repo1=%s
repo2=%s
echo $repo1
ls -1 /pfs/$repo1
echo $repo2
ls -1 /pfs/$repo2
for f1 in /pfs/$repo1/*
do
	for f2 in /pfs/$repo2/*
	do
		v1=$(<$f1)
		v2=$(<$f2)
		echo $v1$v2 >> /pfs/out/file
	done
done
`, inputRepo1, inputRepo2)},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 4,
		},
		[]*ppsclient.PipelineInput{
			{
				Repo:   &pfsclient.Repo{Name: inputRepo1},
				Method: client.IncrementalReduceMethod,
			},
			{
				Repo:   &pfsclient.Repo{Name: inputRepo2},
				Method: client.IncrementalReduceMethod,
			},
		},
		false,
	))

	content := "foo"
	numfiles := 10

	commit1, err := c.StartCommit(inputRepo1, "master")
	for i := 0; i < numfiles; i++ {
		_, err = c.PutFile(inputRepo1, commit1.ID, fmt.Sprintf("file%d", i), strings.NewReader(content))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(inputRepo1, commit1.ID))

	commit2, err := c.StartCommit(inputRepo2, "master")
	for i := 0; i < numfiles; i++ {
		_, err = c.PutFile(inputRepo2, commit2.ID, fmt.Sprintf("file%d", i), strings.NewReader(content))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(inputRepo2, commit2.ID))

	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: &pfsclient.Repo{pipelineName},
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))

	fileInfos, err := c.ListFile(pipelineName, outCommits[0].Commit.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	require.Equal(t, numfiles*numfiles, len(lines))
	for _, line := range lines {
		require.Equal(t, len(content)*2, len(line))
	}

	commit3, err := c.StartCommit(inputRepo1, commit1.ID)
	for i := 0; i < numfiles; i++ {
		_, err = c.PutFile(inputRepo1, commit3.ID, fmt.Sprintf("file2-%d", i), strings.NewReader(content))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(inputRepo1, commit3.ID))

	listCommitRequest.FromCommits[0] = outCommits[0].Commit
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))

	buffer.Reset()
	require.NoError(t, c.GetFile(pipelineName, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	lines = strings.Split(strings.TrimSpace(buffer.String()), "\n")
	require.Equal(t, 2*numfiles*numfiles, len(lines))
	for _, line := range lines {
		require.Equal(t, len(content)*2, len(line))
	}

	commit4, err := c.StartCommit(inputRepo2, commit2.ID)
	for i := 0; i < numfiles; i++ {
		_, err = c.PutFile(inputRepo2, commit4.ID, fmt.Sprintf("file2-%d", i), strings.NewReader(content))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(inputRepo2, commit4.ID))

	listCommitRequest.FromCommits[0] = outCommits[0].Commit
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))

	buffer.Reset()
	require.NoError(t, c.GetFile(pipelineName, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	lines = strings.Split(strings.TrimSpace(buffer.String()), "\n")
	require.Equal(t, 4*numfiles*numfiles, len(lines))
	for _, line := range lines {
		require.Equal(t, len(content)*2, len(line))
	}
}

func TestPipelineWithGlobalMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	globalRepo := uniqueString("inputRepo")
	require.NoError(t, c.CreateRepo(globalRepo))
	numfiles := 20

	pipelineName := uniqueString("pipeline")
	parallelism := 2
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		// this script simply outputs the number of files under the global repo
		[]string{fmt.Sprintf(`
numfiles=(/pfs/%s/*)
numfiles=${#numfiles[@]}
echo $numfiles > /pfs/out/file
`, globalRepo)},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: uint64(parallelism),
		},
		[]*ppsclient.PipelineInput{
			{
				Repo:   &pfsclient.Repo{Name: globalRepo},
				Method: client.GlobalMethod,
			},
		},
		false,
	))

	content := "foo"

	commit, err := c.StartCommit(globalRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numfiles; i++ {
		_, err = c.PutFile(globalRepo, commit.ID, fmt.Sprintf("file%d", i), strings.NewReader(content))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(globalRepo, commit.ID))

	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: &pfsclient.Repo{pipelineName},
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))

	fileInfos, err := c.ListFile(pipelineName, outCommits[0].Commit.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	require.Equal(t, parallelism, len(lines)) // each job outputs one line
	for _, line := range lines {
		require.Equal(t, fmt.Sprintf("%d", numfiles), line)
	}
}

func TestPipelineWithPrevRepoAndIncrementalReduceMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	repo := uniqueString("repo")
	require.NoError(t, c.CreateRepo(repo))

	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{fmt.Sprintf(`
cat /pfs/%s/file >>/pfs/out/file
if [ -d "/pfs/prev" ]; then
  cat /pfs/prev/file >>/pfs/out/file
fi
`, repo)},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{
			{
				Repo:   &pfsclient.Repo{Name: repo},
				Method: client.IncrementalReduceMethod,
			},
		},
		false,
	))

	commit1, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, c.FinishCommit(repo, commit1.ID))

	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: &pfsclient.Repo{pipelineName},
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err := c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	require.Equal(t, 1, len(lines))
	require.Equal(t, "foo", lines[0])

	commit2, err := c.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, c.FinishCommit(repo, commit2.ID))

	listCommitRequest = &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{outCommits[0].Commit},
		CommitType:  pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:       true,
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))

	var buffer2 bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, outCommits[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer2))
	lines = strings.Split(strings.TrimSpace(buffer2.String()), "\n")
	require.Equal(t, 3, len(lines))
	require.Equal(t, "foo", lines[0])
	require.Equal(t, "bar", lines[1])
	require.Equal(t, "foo", lines[2])
}

func TestPipelineThatUseNonexistentInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	pipelineName := uniqueString("pipeline")
	require.YesError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{
			{
				Repo: &pfsclient.Repo{Name: "nonexistent"},
			},
		},
		false,
	))
}

func TestPipelineWhoseInputsGetDeleted(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	repo := uniqueString("repo")
	require.NoError(t, c.CreateRepo(repo))

	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{"true"},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{
			{
				Repo: &pfsclient.Repo{Name: repo},
			},
		},
		false,
	))

	// Shouldn't be able to delete the input repo because it's the provenance
	// of the output repo.
	require.YesError(t, c.DeleteRepo(repo, false))

	// The correct flow to delete the input repo
	require.NoError(t, c.DeletePipeline(pipelineName))
	require.NoError(t, c.DeleteRepo(pipelineName, false))
	require.NoError(t, c.DeleteRepo(repo, false))
}

// This test fails if you updated some static assets (such as doc/development/pipeline_spec.md)
// that are used in code but forgot to run:
// $ make assets
func TestAssets(t *testing.T) {
	assetPaths := []string{"doc/development/pipeline_spec.md"}

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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: client.NewRepo(aRepo)}},
		false,
	))
	cPipeline := uniqueString("C")
	require.NoError(t, c.CreatePipeline(
		cPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("diff %s %s >/pfs/out/file",
			path.Join("/pfs", aRepo, "file"), path.Join("/pfs", bPipeline, "file"))},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{
			{Repo: client.NewRepo(aRepo)},
			{Repo: client.NewRepo(bPipeline)},
		},
		false,
	))
	// commit to aRepo
	commit1, err := c.StartCommit(aRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, commit1.ID))
	commitInfos, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit(aRepo, commit1.ID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	commit2, err := c.StartCommit(aRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(aRepo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, commit2.ID))
	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{client.NewCommit(aRepo, commit2.ID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	// There should only be 2 commits on cRepo
	commitInfos, err = c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{cPipeline},
	}}, nil, pfsclient.CommitType_COMMIT_TYPE_READ, pfsclient.CommitStatus_NORMAL, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	for _, commitInfo := range commitInfos {
		// C takes the diff of 2 files that should always be the same, so we
		// expect an empty file
		fileInfo, err := c.InspectFile(cPipeline, commitInfo.Commit.ID, "file", "", false, nil)
		require.NoError(t, err)
		require.Equal(t, uint64(0), fileInfo.SizeBytes)
	}
}

func TestDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel() //cleanup resources

	job1, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Transform: &ppsclient.Transform{
			Cmd: []string{"sh"},
			Stdin: []string{
				"mkdir /pfs/out/dir",
				"echo foo >> /pfs/out/dir/file",
			},
		},
		ParallelismSpec: &ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 3,
		},
	})
	require.NoError(t, err)
	inspectJobRequest1 := &ppsclient.InspectJobRequest{
		Job:        job1,
		BlockState: true,
	}
	jobInfo1, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest1)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS, jobInfo1.State)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(jobInfo1.OutputCommit.Repo.Name, jobInfo1.OutputCommit.ID, "dir/file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())

	job2, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Transform: &ppsclient.Transform{
			Cmd: []string{"sh"},
			Stdin: []string{
				"echo bar >> /pfs/out/dir/file",
			},
		},
		ParallelismSpec: &ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 3,
		},
		ParentJob: job1,
	})
	require.NoError(t, err)
	inspectJobRequest2 := &ppsclient.InspectJobRequest{
		Job:        job2,
		BlockState: true,
	}
	jobInfo2, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest2)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS, jobInfo2.State)

	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(jobInfo2.OutputCommit.Repo.Name, jobInfo2.OutputCommit.ID, "dir/file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\nbar\nbar\nbar\n", buffer.String())
}

func TestFailedJobReadData(t *testing.T) {
	// We want to enable users to be able to read data from cancelled commits for debugging purposes`
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	shards := 0
	c := getPachClient(t)
	repo := uniqueString("TestJob_Foo")

	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	fileContent := "foo\n"
	_, err = c.PutFile(repo, commit.ID, "file", strings.NewReader(fileContent))
	require.NoError(t, err)
	err = c.FinishCommit(repo, commit.ID)
	require.NoError(t, err)

	job, err := c.CreateJob(
		"",
		[]string{"bash"},
		[]string{
			"echo fubar > /pfs/out/file",
			"exit 1",
		},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: uint64(shards),
		},
		[]*ppsclient.JobInput{
			{
				Commit: commit,
				Method: client.ReduceMethod,
			},
		},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:        job,
		BlockState: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_FAILURE.String(), jobInfo.State.String())
	parallelism, err := pps_server.GetExpectedNumWorkers(getKubeClient(t), jobInfo.ParallelismSpec)
	require.NoError(t, err)
	require.True(t, parallelism > 0)
	c.GetLogs(jobInfo.Job.ID, os.Stdout)
	commitInfo, err := c.InspectCommit(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	require.NoError(t, err)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	require.Equal(t, true, commitInfo.Cancelled)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "fubar", strings.TrimSpace(buffer.String()))

}

// TestFlushCommit
func TestFlushCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)
	prefix := uniqueString("repo")
	makeRepoName := func(i int) string {
		return fmt.Sprintf("%s_%d", prefix, i)
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
			&ppsclient.ParallelismSpec{
				Strategy: ppsclient.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			[]*ppsclient.PipelineInput{{Repo: client.NewRepo(repo)}},
			false,
		))
	}

	test := func(parent string) string {
		commit, err := c.StartCommit(sourceRepo, parent)
		require.NoError(t, err)
		_, err = c.PutFile(sourceRepo, commit.ID, "file", strings.NewReader("foo\n"))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(sourceRepo, commit.ID))
		commitInfos, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit(sourceRepo, commit.ID)}, nil)
		require.NoError(t, err)
		require.Equal(t, numStages+1, len(commitInfos))
		return commit.ID
	}

	// Run the test twice, once on a orphan commit and another on
	// a commit with a parent
	commit := test(uuid.New())
	test(commit)
}

func TestFlushCommitAfterCreatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)
	repo := uniqueString("TestFlushCommitAfterCreatePipeline")
	require.NoError(t, c.CreateRepo(repo))

	for i := 0; i < 10; i++ {
		_, err := c.StartCommit(repo, "master")
		require.NoError(t, err)
		_, err = c.PutFile(repo, "master", "file", strings.NewReader(fmt.Sprintf("foo%d\n", i)))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(repo, "master"))
	}

	pipeline := uniqueString("TestFlushCommitAfterCreatePipelinePipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"cp", path.Join("/pfs", repo, "file"), "/pfs/out/file"},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: client.NewRepo(repo)}},
		false,
	))
	_, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit(repo, "master")}, nil)
	require.NoError(t, err)
}

// TestFlushCommitWithFailure is similar to TestFlushCommit except that
// the pipeline is designed to fail
func TestFlushCommitWithFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)
	prefix := uniqueString("repo")
	makeRepoName := func(i int) string {
		return fmt.Sprintf("%s_%d", prefix, i)
	}

	sourceRepo := makeRepoName(0)
	require.NoError(t, c.CreateRepo(sourceRepo))

	// Create a five-stage pipeline; the third stage is designed to fail
	numStages := 5
	for i := 0; i < numStages; i++ {
		fileName := "file"
		if i == 3 {
			fileName = "nonexistent"
		}
		repo := makeRepoName(i)
		require.NoError(t, c.CreatePipeline(
			makeRepoName(i+1),
			"",
			[]string{"cp", path.Join("/pfs", repo, fileName), "/pfs/out/file"},
			nil,
			&ppsclient.ParallelismSpec{
				Strategy: ppsclient.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			[]*ppsclient.PipelineInput{{Repo: client.NewRepo(repo)}},
			false,
		))
	}

	commit, err := c.StartCommit(sourceRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(sourceRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(sourceRepo, commit.ID))
	_, err = c.FlushCommit([]*pfsclient.Commit{client.NewCommit(sourceRepo, commit.ID)}, nil)
	require.YesError(t, err)
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
			&ppsclient.ParallelismSpec{
				Strategy: ppsclient.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			[]*ppsclient.PipelineInput{{Repo: client.NewRepo(repo)}},
			false,
		))
		listCommitRequest := &pfsclient.ListCommitRequest{
			FromCommits: []*pfsclient.Commit{{
				Repo: &pfsclient.Repo{pipeline},
			}},
			CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
			Block:      true,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		listCommitResponse, err := c.PfsAPIClient.ListCommit(
			ctx,
			listCommitRequest,
		)
		require.NoError(t, err)
		outCommits := listCommitResponse.CommitInfo
		require.Equal(t, 1, len(outCommits))
	}

	// Do it twice.  We expect jobs to be created on both runs.
	createPipeline()
	require.NoError(t, c.DeleteRepo(pipeline, false))
	require.NoError(t, c.DeletePipeline(pipeline))
	createPipeline()
}

func TestPipelineState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Skip("after the refactor, it's a little unclear how you'd introduce an error into a pipeline; see #762")

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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: client.NewRepo(repo)}},
		false,
	))

	time.Sleep(5 * time.Second) // wait for this pipeline to get picked up
	pipelineInfo, err := c.InspectPipeline(pipeline)
	require.NoError(t, err)
	require.Equal(t, ppsclient.PipelineState_PIPELINE_RUNNING, pipelineInfo.State)

	// Now we introduce an error to the pipeline by removing its output repo
	// and starting a job
	require.NoError(t, c.DeleteRepo(pipeline, false))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))

	// So the state of the pipeline will alternate between running and
	// restarting.  We just want to make sure that it has definitely restarted.
	var states []interface{}
	for i := 0; i < 20; i++ {
		time.Sleep(500 * time.Millisecond)
		pipelineInfo, err = c.InspectPipeline(pipeline)
		require.NoError(t, err)
		states = append(states, pipelineInfo.State)

	}
	require.EqualOneOf(t, states, ppsclient.PipelineState_PIPELINE_RESTARTING)
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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: client.NewRepo(repo)}},
		false,
	))

	// Trigger a job by creating a commit
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	_, err = c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)
	jobInfos, err := c.ListJob(pipeline, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	inspectJobRequest := &ppsclient.InspectJobRequest{
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
	require.Equal(t, int32(1), pipelineInfo.JobCounts[int32(ppsclient.JobState_JOB_SUCCESS)])
}

func TestJobState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)

	// This job uses a nonexistent image; it's supposed to stay in the
	// "pulling" state
	job, err := c.CreateJob(
		"nonexistent",
		[]string{"bash"},
		nil,
		&ppsclient.ParallelismSpec{},
		nil,
		"",
	)
	require.NoError(t, err)
	time.Sleep(10 * time.Second)
	jobInfo, err := c.InspectJob(job.ID, false)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_PULLING, jobInfo.State)

	// This job sleeps for 20 secs
	job, err = c.CreateJob(
		"",
		[]string{"bash"},
		[]string{"sleep 20"},
		&ppsclient.ParallelismSpec{},
		nil,
		"",
	)
	require.NoError(t, err)
	time.Sleep(10 * time.Second)
	jobInfo, err = c.InspectJob(job.ID, false)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_RUNNING, jobInfo.State)

	// Wait for the job to complete
	jobInfo, err = c.InspectJob(job.ID, true)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS, jobInfo.State)
}

func TestClusterFunctioningAfterMembershipChange(t *testing.T) {
	t.Skip("this test is flaky")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	scalePachd(t, true)
	testJob(t, 4)
	scalePachd(t, false)
	testJob(t, 4)
}

func TestDeleteAfterMembershipChange(t *testing.T) {
	t.Skip("this test is flaky")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	test := func(up bool) {
		repo := uniqueString("TestDeleteAfterMembershipChange")
		c := getPachClient(t)
		require.NoError(t, c.CreateRepo(repo))
		_, err := c.StartCommit(repo, "master")
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(repo, "master"))
		scalePachd(t, up)
		c = getUsablePachClient(t)
		require.NoError(t, c.DeleteRepo(repo, false))
	}
	test(true)
	test(false)
}

func TestScrubbedErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)

	_, err := c.InspectPipeline("blah")
	require.Equal(t, "PipelineInfos blah not found", err.Error())

	err = c.CreatePipeline(
		"lskdjf$#%^ERTYC",
		"",
		[]string{},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: "test"}}},
		false,
	)
	require.Equal(t, "repo test not found", err.Error())

	_, err = c.CreateJob("askjdfhgsdflkjh", []string{}, []string{}, &ppsclient.ParallelismSpec{}, []*ppsclient.JobInput{client.NewJobInput("bogusRepo", "bogusCommit", client.DefaultMethod)}, "")
	require.Matches(t, "could not create repo job_.*, not all provenance repos exist", err.Error())

	_, err = c.InspectJob("blah", true)
	require.Equal(t, "JobInfos blah not found", err.Error())

	home := os.Getenv("HOME")
	f, err := os.Create(filepath.Join(home, "/tmpfile"))
	defer func() {
		os.Remove(filepath.Join(home, "/tmpfile"))
	}()
	require.NoError(t, err)
	err = c.GetLogs("bogusJobId", f)
	require.Equal(t, "job bogusJobId not found", err.Error())
}

func TestLeakingRepo(t *testing.T) {
	// If CreateJob fails, it should also destroy the output repo it creates
	// If it doesn't, it can cause flush commit to fail, as a bogus repo will
	// be listed in the output repo's provenance

	// This test can't be run in parallel, since it requires using the repo counts as controls
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)

	repoInfos, err := c.ListRepo(nil)
	require.NoError(t, err)
	initialCount := len(repoInfos)

	_, err = c.CreateJob("bogusImage", []string{}, []string{}, &ppsclient.ParallelismSpec{}, []*ppsclient.JobInput{client.NewJobInput("bogusRepo", "bogusCommit", client.DefaultMethod)}, "")
	require.Matches(t, "could not create repo job_.*, not all provenance repos exist", err.Error())

	repoInfos, err = c.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, initialCount, len(repoInfos))
}

func TestAcceptReturnCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)
	job, err := c.PpsAPIClient.CreateJob(
		context.Background(),
		&ppsclient.CreateJobRequest{
			Transform: &ppsclient.Transform{
				Cmd:              []string{"sh"},
				Stdin:            []string{"exit 1"},
				AcceptReturnCode: []int64{1},
			},
			ParallelismSpec: &ppsclient.ParallelismSpec{},
		},
	)
	require.NoError(t, err)
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:        job,
		BlockState: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_SUCCESS.String(), jobInfo.State.String())
}

func TestRestartAll(t *testing.T) {
	t.Skip("this test is flaky")
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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
		false,
	))
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	_, err = c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)

	restartAll(t)

	// need a new client because the old one will have a defunct connection
	c = getUsablePachClient(t)

	// Wait a little for pipelines to restart
	time.Sleep(10 * time.Second)
	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)
	require.Equal(t, ppsclient.PipelineState_PIPELINE_RUNNING, pipelineInfo.State)
	_, err = c.InspectRepo(dataRepo)
	require.NoError(t, err)
	_, err = c.InspectCommit(dataRepo, commit.ID)
	require.NoError(t, err)
}

func TestRestartOne(t *testing.T) {
	t.Skip("this test is flaky")
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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
		false,
	))
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	_, err = c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)

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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
		false,
	))
	// Do a commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitInfos, err := c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)
	repoInfo, err := c.InspectRepo(dataRepo)
	require.NoError(t, err)
	require.NoError(t, pfspretty.PrintDetailedRepoInfo(repoInfo))
	for _, commitInfo := range commitInfos {
		require.NoError(t, pfspretty.PrintDetailedCommitInfo(commitInfo))
	}
	fileInfo, err := c.InspectFile(dataRepo, commit.ID, "file", "", false, nil)
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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
		false,
	))
	// Do commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	_, err = c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)
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
			fmt.Sprintf("mkdir /inputs"),
			fmt.Sprintf("cp -r /pfs/%s /inputs", dataRepo),
		},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{
			Repo:   client.NewRepo(dataRepo),
			Method: client.IncrementalReduceMethod,
		}},
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
	_, err = c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)
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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{
			{
				Repo: &pfsclient.Repo{Name: repo},
			},
		},
		false,
	))
	err := c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{
			{
				Repo: &pfsclient.Repo{Name: repo},
			},
		},
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "pipeline .*? already exists", err.Error())
}

func TestPipelineInfoDestroyedIfRepoCreationFails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	repo := uniqueString("data")
	require.NoError(t, c.CreateRepo(repo))
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreateRepo(pipelineName))
	err := c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{
			{
				Repo: &pfsclient.Repo{Name: repo},
			},
		},
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "repo .* exists", err.Error())
	_, err = c.InspectPipeline(pipelineName)
	require.YesError(t, err)
	require.Matches(t, "not found", err.Error())
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
	// create 2 pipelines
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file1"), "/pfs/out/file"},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: client.NewRepo(dataRepo)}},
		false,
	))
	pipeline2Name := uniqueString("pipeline2")
	require.NoError(t, c.CreatePipeline(
		pipeline2Name,
		"",
		[]string{"cp", path.Join("/pfs", pipelineName, "file"), "/pfs/out/file"},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: client.NewRepo(pipelineName)}},
		false,
	))
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file1", strings.NewReader("file1\n"))
	_, err = c.PutFile(dataRepo, commit.ID, "file2", strings.NewReader("file2\n"))
	_, err = c.PutFile(dataRepo, commit.ID, "file3", strings.NewReader("file3\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitInfos, err := c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	// only care about non-provenance commits
	commitInfos = commitInfos[1:]
	for _, commitInfo := range commitInfos {
		var buffer bytes.Buffer
		require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, "", false, nil, &buffer))
		require.Equal(t, "file1\n", buffer.String())
	}

	// We archive the temporary commits created per job/pod
	// So the total we see here is 2, but 'real' commits is just 1
	outputRepoCommitInfos, err := c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{pipelineName},
	}}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(outputRepoCommitInfos))

	outputRepoCommitInfos, err = c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{pipelineName},
	}}, nil, client.CommitTypeRead, client.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(outputRepoCommitInfos))

	// Update the pipeline to look at file2
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file2"), "/pfs/out/file"},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
		true,
	))
	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)
	require.NotNil(t, pipelineInfo.CreatedAt)
	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	// only care about non-provenance commits
	commitInfos = commitInfos[1:]
	for _, commitInfo := range commitInfos {
		var buffer bytes.Buffer
		require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, "", false, nil, &buffer))
		require.Equal(t, "file2\n", buffer.String())
	}
	outputRepoCommitInfos, err = c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{pipelineName},
	}}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
	require.NoError(t, err)
	require.Equal(t, 4, len(outputRepoCommitInfos))
	// Expect real commits to still be 1
	outputRepoCommitInfos, err = c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{pipelineName},
	}}, nil, client.CommitTypeRead, client.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(outputRepoCommitInfos))

	// Update the pipeline to look at file3
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file3"), "/pfs/out/file"},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
		true,
	))
	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	// only care about non-provenance commits
	commitInfos = commitInfos[1:]
	for _, commitInfo := range commitInfos {
		var buffer bytes.Buffer
		require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, "", false, nil, &buffer))
		require.Equal(t, "file3\n", buffer.String())
	}
	outputRepoCommitInfos, err = c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{pipelineName},
	}}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
	require.NoError(t, err)
	require.Equal(t, 6, len(outputRepoCommitInfos))
	// Expect real commits to still be 1
	outputRepoCommitInfos, err = c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{pipelineName},
	}}, nil, client.CommitTypeRead, client.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(outputRepoCommitInfos))

	commitInfos, _ = c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{pipelineName},
	}}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
	// Do an update that shouldn't cause archiving
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&ppsclient.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &ppsclient.Transform{
				Cmd: []string{"cp", path.Join("/pfs", dataRepo, "file3"), "/pfs/out/file"},
			},
			ParallelismSpec: &ppsclient.ParallelismSpec{
				Strategy: ppsclient.ParallelismSpec_CONSTANT,
				Constant: 2,
			},
			Inputs:    []*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
			Update:    true,
			NoArchive: true,
		})
	require.NoError(t, err)
	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	// only care about non-provenance commits
	commitInfos = commitInfos[1:]
	for _, commitInfo := range commitInfos {
		var buffer bytes.Buffer
		require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, "", false, nil, &buffer))
		require.Equal(t, "file3\n", buffer.String())
	}
	commitInfos, err = c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{pipelineName},
	}}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
	require.NoError(t, err)
	require.Equal(t, 6, len(commitInfos))
	// Expect real commits to still be 1
	outputRepoCommitInfos, err = c.ListCommit([]*pfsclient.Commit{{
		Repo: &pfsclient.Repo{pipelineName},
	}}, nil, client.CommitTypeRead, client.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(outputRepoCommitInfos))
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
	outRepo := ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
		false,
	))
	require.NoError(t, c.StopPipeline(pipelineName))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	// timeout because the Flush should never return since the pipeline is
	// stopped
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	_, err = c.PfsAPIClient.FlushCommit(
		ctx,
		&pfsclient.FlushCommitRequest{
			Commit: []*pfsclient.Commit{commit1},
		})
	require.YesError(t, err)
	require.NoError(t, c.StartPipeline(pipelineName))
	commitInfos, err := c.FlushCommit([]*pfsclient.Commit{commit1}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, commitInfos[1].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
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
		&ppsclient.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &ppsclient.Transform{
				Cmd: []string{"sh"},
				Stdin: []string{
					"ls /var/secret",
					"cat /var/secret/foo > /pfs/out/foo",
					"echo $bar> /pfs/out/bar",
				},
				Env: map[string]string{"bar": "bar"},
				Secrets: []*ppsclient.Secret{
					{
						Name:      secretName,
						MountPath: "/var/secret",
					},
				},
			},
			ParallelismSpec: &ppsclient.ParallelismSpec{
				Strategy: ppsclient.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			Inputs: []*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
		})
	require.NoError(t, err)
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitInfos, err := c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.Equal(t, 2, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, commitInfos[1].Commit.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(pipelineName, commitInfos[1].Commit.ID, "bar", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "bar\n", buffer.String())
}

func TestFlushNonExistantCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)
	_, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit("fake-repo", "fake-commit")}, nil)
	require.YesError(t, err)
	repo := uniqueString("FlushNonExistantCommit")
	require.NoError(t, c.CreateRepo(repo))
	_, err = c.FlushCommit([]*pfsclient.Commit{client.NewCommit(repo, "fake-commit")}, nil)
	require.YesError(t, err)
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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{
			{
				Repo: client.NewRepo(dataRepo),
				Method: &ppsclient.Method{
					Partition:   ppsclient.Partition_BLOCK,
					Incremental: ppsclient.Incremental_FULL,
				},
			},
		},
		false,
	))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	commitInfos, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit(dataRepo, commit1.ID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	// Do second commit to repo
	commit2, err := c.StartCommit(dataRepo, commit1.ID)
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))
	commitInfos, err = c.FlushCommit([]*pfsclient.Commit{client.NewCommit(dataRepo, commit2.ID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())
}

func TestArchiveAllWithPipelines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// This test cannot be run in parallel, since it archives all repos
	c := getUsablePachClient(t)
	dataRepo := uniqueString("TestUpdatePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	numPipelines := 10
	var outputRepos []*pfsclient.Repo
	for i := 0; i < numPipelines; i++ {
		pipelineName := uniqueString("pipeline")
		outputRepos = append(outputRepos, &pfsclient.Repo{Name: pipelineName})
		require.NoError(t, c.CreatePipeline(
			pipelineName,
			"",
			[]string{"cp", path.Join("/pfs", dataRepo, "file1"), "/pfs/out/file"},
			nil,
			&ppsclient.ParallelismSpec{
				Strategy: ppsclient.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			[]*ppsclient.PipelineInput{{Repo: client.NewRepo(dataRepo)}},
			false,
		))
	}
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file1", strings.NewReader("file1\n"))
	_, err = c.PutFile(dataRepo, commit.ID, "file2", strings.NewReader("file2\n"))
	_, err = c.PutFile(dataRepo, commit.ID, "file3", strings.NewReader("file3\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitInfos, err := c.FlushCommit([]*pfsclient.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, numPipelines+1, len(commitInfos))

	require.NoError(t, c.ArchiveAll())
	commitInfos, err = c.ListCommit(
		[]*pfsclient.Commit{{
			Repo: &pfsclient.Repo{dataRepo},
		}},
		nil,
		client.CommitTypeNone,
		client.CommitStatusNormal,
		false,
	)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))
}

// This test / failure pattern shouldn't be possible after
// the pfs-refactor branch lands
func TestListCommitReturnsBlankCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Skip("This test does a restart which seems to break other tests.")
	// this test cannot be run in parallel because it restarts everything which breaks other tests.
	c := getPachClient(t)

	// create repos
	dataRepo := uniqueString("TestListCommitReturnsBlankCommit")
	require.NoError(t, c.CreateRepo(dataRepo))
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	listCommitRequest := &pfsclient.ListCommitRequest{
		FromCommits: []*pfsclient.Commit{{
			Repo: &pfsclient.Repo{dataRepo},
		}},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	commitInfos, err := c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos.CommitInfo))
	restartAll(t)

	// need a new client because the old one will have a defunct connection
	c = getUsablePachClient(t)
	commitInfos, err = c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	// In the buggy behaviour, after restarting we'd see 2 commits, one of
	// which is the 'blank' commit that's created when creating a repo
	require.Equal(t, 1, len(commitInfos.CommitInfo))
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
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: client.NewRepo(aRepo)}},
		false,
	))

	cPipeline := uniqueString("C")
	require.NoError(t, c.CreatePipeline(
		cPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/file /pfs/out/bFile", bPipeline),
			fmt.Sprintf("cp /pfs/%s/file /pfs/out/dFile", dRepo)},
		&ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*ppsclient.PipelineInput{{Repo: client.NewRepo(bPipeline)},
			{Repo: client.NewRepo(dRepo)}},
		false,
	))
	results, err := c.FlushCommit([]*pfsclient.Commit{aCommit, dCommit}, nil)
	require.NoError(t, err)
	require.Equal(t, 4, len(results))
}

func TestParallelismSpec(t *testing.T) {
	// Test Constant strategy
	parellelism, err := pps_server.GetExpectedNumWorkers(getKubeClient(t), &ppsclient.ParallelismSpec{
		Strategy: ppsclient.ParallelismSpec_CONSTANT,
		Constant: 7,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(7), parellelism)

	// Coefficient == 1 (basic test)
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &ppsclient.ParallelismSpec{
		Strategy:    ppsclient.ParallelismSpec_COEFFICIENT,
		Coefficient: 1,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)

	// Coefficient > 1
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &ppsclient.ParallelismSpec{
		Strategy:    ppsclient.ParallelismSpec_COEFFICIENT,
		Coefficient: 2,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(2), parellelism)

	// Make sure we start at least one worker
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &ppsclient.ParallelismSpec{
		Strategy:    ppsclient.ParallelismSpec_COEFFICIENT,
		Coefficient: 0.1,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)

	// Test 0-initialized JobSpec
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &ppsclient.ParallelismSpec{})
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)

	// Test nil JobSpec
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), nil)
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)
}

func getPachClient(t testing.TB) *client.APIClient {
	client, err := client.NewFromAddress("0.0.0.0:30650")
	require.NoError(t, err)
	return client
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
		_, err := client.PfsAPIClient.ListRepo(ctx, &pfsclient.ListRepoRequest{})
		if err == nil {
			return client
		}
	}
	t.Fatalf("failed to connect after %d tries", retries)
	return nil
}

func getKubeClient(t *testing.T) *kube.Client {
	config := &kube_client.Config{
		Host:     "0.0.0.0:8080",
		Insecure: false,
	}
	k, err := kube.New(config)
	require.NoError(t, err)
	return k
}

func uniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
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

func scalePachdUp(t *testing.T) {
	scalePachd(t, true)
}

func scalePachdDown(t *testing.T) {
	scalePachd(t, false)
}

func waitForReadiness(t *testing.T) {
	k := getKubeClient(t)
	rc := pachdRc(t)
	for {
		has, err := kube.ControllerHasDesiredReplicas(k, rc)()
		require.NoError(t, err)
		if has {
			break
		}
		time.Sleep(time.Second * 5)
	}
	watch, err := k.Pods(api.NamespaceDefault).Watch(api.ListOptions{
		LabelSelector: kube_labels.SelectorFromSet(map[string]string{"app": "pachd"}),
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
