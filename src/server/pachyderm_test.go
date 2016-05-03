package server

import (
	"bytes"
	"fmt"
	"math/rand"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
)

const (
	NUMFILES = 25
	KB       = 1024 * 1024
)

func TestJob(t *testing.T) {
	testJob(t, 4)
}

func TestJobNoShard(t *testing.T) {
	testJob(t, 0)
}

func testJob(t *testing.T, shards int) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)
	dataRepo := uniqueString("TestJob.data")
	require.NoError(t, c.CreateRepo(dataRepo))
	commit, err := c.StartCommit(dataRepo, "", "")
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
		uint64(shards),
		[]*ppsclient.JobInput{{
			Commit: commit,
			Reduce: true,
		}},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:         job,
		BlockOutput: true,
		BlockState:  true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_STATE_SUCCESS.String(), jobInfo.State.String())
	require.True(t, jobInfo.Parallelism > 0)
	commitInfo, err := c.InspectCommit(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	require.NoError(t, err)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	for i := 0; i < numFiles; i++ {
		var buffer bytes.Buffer
		require.NoError(t, c.GetFile(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, fmt.Sprintf("file-%d", i), 0, 0, "", nil, &buffer))
		require.Equal(t, fileContent, buffer.String())
	}
}

func TestDuplicatedJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()

	c := getPachClient(t)

	dataRepo := uniqueString("TestDuplicatedJob.data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit, err := c.StartCommit(dataRepo, "", "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	pipelineName := uniqueString("TestDuplicatedJob.pipeline")
	require.NoError(t, c.CreateRepo(pipelineName))

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

	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:         job1,
		BlockOutput: true,
		BlockState:  true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel() //cleanup resources
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, "file", 0, 0, "", nil, &buffer))
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
		4,
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
}

func TestGrep(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	dataRepo := uniqueString("TestGrep.data")
	c := getPachClient(t)
	require.NoError(t, c.CreateRepo(dataRepo))
	commit, err := c.StartCommit(dataRepo, "", "")
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
		1,
		[]*ppsclient.JobInput{{Commit: commit}},
		"",
	)
	require.NoError(t, err)
	job2, err := c.CreateJob(
		"",
		[]string{"bash"},
		[]string{fmt.Sprintf("grep foo /pfs/%s/* >/pfs/out/foo", dataRepo)},
		4,
		[]*ppsclient.JobInput{{Commit: commit}},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:         job1,
		BlockOutput: true,
		BlockState:  true,
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
		1,
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
	require.Equal(t, ppsclient.JobState_JOB_STATE_SUCCESS.String(), jobInfo.State.String())
}

func TestPipeline(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestPipeline.data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	outRepo := ppsserver.PipelineRepo(client.NewPipeline(pipelineName))
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		1,
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
	))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "", "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	listCommitRequest := &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
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
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	// Do second commit to repo
	commit2, err := c.StartCommit(dataRepo, commit1.ID, "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))
	listCommitRequest = &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
		FromCommit: []*pfsclient.Commit{outCommits[0].Commit},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
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
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())

	require.NoError(t, c.DeletePipeline(pipelineName))

	pipelineInfos, err := c.PpsAPIClient.ListPipeline(context.Background(), &ppsclient.ListPipelineRequest{})
	require.NoError(t, err)
	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		require.True(t, pipelineInfo.Pipeline.Name != pipelineName)
	}

	// Do third commit to repo; this time pipeline should not run since it's been deleted
	commit3, err := c.StartCommit(dataRepo, commit2.ID, "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit3.ID, "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit3.ID))

	// We will sleep a while to wait for the pipeline to actually get cancelled
	// Also if the pipeline didn't get cancelled (due to a bug), we sleep a while
	// to let the pipeline commit
	time.Sleep(5 * time.Second)
	listCommitRequest = &pfsclient.ListCommitRequest{
		Repo: []*pfsclient.Repo{outRepo},
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	// there should only be two commits in the pipeline
	require.Equal(t, 2, len(listCommitResponse.CommitInfo))
}

func TestPipelineWithTooMuchParallelism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestPipelineWithTooMuchParallelism.data")
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
		3,
		[]*ppsclient.PipelineInput{{
			Repo:   &pfsclient.Repo{Name: dataRepo},
			Reduce: true, // setting reduce to true so only one pod gets the file
		}},
	))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "", "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	listCommitRequest := &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
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
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	require.Equal(t, false, outCommits[0].Cancelled)
}

func TestPipelineWithEmptyInputs(t *testing.T) {
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
		3,
		nil,
	))

	// Manually trigger the pipeline
	job, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})
	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:         job,
		BlockOutput: true,
		BlockState:  true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	jobInfo, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_STATE_SUCCESS.String(), jobInfo.State.String())
	require.Equal(t, 3, int(jobInfo.Parallelism))

	listCommitRequest := &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
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
	fileInfos, err := c.ListFile(outRepo.Name, outCommits[0].Commit.ID, "", "", nil, false)
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfos))

	// Make sure that each job gets a different ID
	job2, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})
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
		3,
		nil,
	))

	// Manually trigger the pipeline
	_, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})

	listCommitRequest := &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
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
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
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
		3,
		nil,
	))

	// Manually trigger the pipeline
	job, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})

	listCommitRequest := &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
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
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())

	// Manually trigger the pipeline
	_, err = c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
		ParentJob: job,
	})

	listCommitRequest = &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		FromCommit: []*pfsclient.Commit{outCommits[0].Commit},
		Block:      true,
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		ctx,
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer2 bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer2))
	require.Equal(t, "foo\nfoo\nfoo\nfoo\nfoo\nfoo\n", buffer2.String())
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
		3,
		nil,
	))

	// Manually trigger the pipeline
	job, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
	})

	listCommitRequest := &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
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
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())

	// Manually trigger the pipeline
	_, err = c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: pipelineName,
		},
		ParentJob: job,
	})

	listCommitRequest = &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
		FromCommit: []*pfsclient.Commit{outCommits[0].Commit},
	}
	listCommitResponse, err = c.PfsAPIClient.ListCommit(
		ctx,
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer2 bytes.Buffer
	require.NoError(t, c.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer2))
	require.Equal(t, "foo\nfoo\nfoo\nfoo\nfoo\nfoo\n", buffer2.String())
}

func TestRemoveAndAppend(t *testing.T) {
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
		Parallelism: 3,
	})
	require.NoError(t, err)

	inspectJobRequest1 := &ppsclient.InspectJobRequest{
		Job:         job1,
		BlockOutput: true,
		BlockState:  true,
	}
	jobInfo1, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest1)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_STATE_SUCCESS, jobInfo1.State)

	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(jobInfo1.OutputCommit.Repo.Name, jobInfo1.OutputCommit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())

	job2, err := c.PpsAPIClient.CreateJob(context.Background(), &ppsclient.CreateJobRequest{
		Transform: &ppsclient.Transform{
			Cmd: []string{"sh"},
			Stdin: []string{
				"unlink /pfs/out/file && echo bar > /pfs/out/file",
			},
		},
		Parallelism: 3,
		ParentJob:   job1,
	})
	require.NoError(t, err)

	inspectJobRequest2 := &ppsclient.InspectJobRequest{
		Job:         job2,
		BlockOutput: true,
		BlockState:  true,
	}
	jobInfo2, err := c.PpsAPIClient.InspectJob(ctx, inspectJobRequest2)
	require.NoError(t, err)
	require.Equal(t, ppsclient.JobState_JOB_STATE_SUCCESS, jobInfo2.State)

	var buffer2 bytes.Buffer
	require.NoError(t, c.GetFile(jobInfo2.OutputCommit.Repo.Name, jobInfo2.OutputCommit.ID, "file", 0, 0, "", nil, &buffer2))
	require.Equal(t, "bar\nbar\nbar\n", buffer2.String())
}

func TestWorkload(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
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
	commit, err := c.StartCommit(repo, "", "")
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
				fmt.Sprintf("file%d", i), 0, 0, "", shard, &buffer1Shard)
			require.NoError(t, err)
			shard.BlockModulus = 4
			for blockNumber := uint64(0); blockNumber < 4; blockNumber++ {
				shard.BlockNumber = blockNumber
				err := c.GetFile(repo, commit.ID,
					fmt.Sprintf("file%d", i), 0, 0, "", shard, &buffer4Shard)
				require.NoError(t, err)
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
	commit1, err := c.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit1.ID, "file", workload.NewReader(rand, KB))
	require.NoError(t, err)
	err = c.FinishCommit(repo, commit1.ID)
	require.NoError(t, err)
	commit2, err := c.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit2.ID, "file", workload.NewReader(rand, KB))
	require.NoError(t, err)
	err = c.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(repo, commit2.ID, "file", 0, 0, commit1.ID, nil, &buffer))
	require.Equal(t, buffer.Len(), KB)
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(repo, commit2.ID, "file", 0, 0, "", nil, &buffer))
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
	commit1, err := c.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit1.ID))
	commitInfos, err := c.ListCommit([]string{repo}, nil, client.NONE, false, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := c.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = c.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func getPachClient(t *testing.T) *client.APIClient {
	client, err := client.NewFromAddress("0.0.0.0:30650")
	require.NoError(t, err)
	return client
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
