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
	pachClient := getPachClient(t)
	dataRepo := uniqueString("TestJob.data")
	require.NoError(t, pfsclient.CreateRepo(pachClient, dataRepo))
	commit, err := pfsclient.StartCommit(pachClient, dataRepo, "", "")
	require.NoError(t, err)
	fileContent := "foo\n"
	// We want to create lots of files so that each parallel job will be
	// started with some files
	numFiles := shards*100 + 100
	for i := 0; i < numFiles; i++ {
		_, err = pfsclient.PutFile(pachClient, dataRepo, commit.ID, fmt.Sprintf("file-%d", i), strings.NewReader(fileContent))
		require.NoError(t, err)
	}
	require.NoError(t, pfsclient.FinishCommit(pachClient, dataRepo, commit.ID))
	job, err := ppsclient.CreateJob(
		pachClient,
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
	jobInfo, err := pachClient.InspectJob(context.Background(), inspectJobRequest)
	require.NoError(t, err)
	t.Logf("jobInfo: %v", jobInfo)
	require.Equal(t, ppsclient.JobState_JOB_STATE_SUCCESS.String(), jobInfo.State.String())
	require.True(t, jobInfo.Shards > 0)
	commitInfo, err := pfsclient.InspectCommit(pachClient, jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	require.NoError(t, err)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	for i := 0; i < numFiles; i++ {
		var buffer bytes.Buffer
		require.NoError(t, pfsclient.GetFile(pachClient, jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, fmt.Sprintf("file-%d", i), 0, 0, "", nil, &buffer))
		require.Equal(t, fileContent, buffer.String())
	}
}

func TestDuplicatedJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()

	pachClient := getPachClient(t)

	dataRepo := uniqueString("TestDuplicatedJob.data")
	require.NoError(t, pfsclient.CreateRepo(pachClient, dataRepo))

	commit, err := pfsclient.StartCommit(pachClient, dataRepo, "", "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = pfsclient.PutFile(pachClient, dataRepo, commit.ID, "file", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, pfsclient.FinishCommit(pachClient, dataRepo, commit.ID))

	pipelineName := uniqueString("TestDuplicatedJob.pipeline")
	require.NoError(t, pfsclient.CreateRepo(pachClient, pipelineName))

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

	job1, err := pachClient.CreateJob(context.Background(), req)
	require.NoError(t, err)

	job2, err := pachClient.CreateJob(context.Background(), req)
	require.NoError(t, err)

	require.Equal(t, job1, job2)

	inspectJobRequest := &ppsclient.InspectJobRequest{
		Job:         job1,
		BlockOutput: true,
		BlockState:  true,
	}
	jobInfo, err := pachClient.InspectJob(context.Background(), inspectJobRequest)
	require.NoError(t, err)

	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pachClient, jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, fileContent, buffer.String())
}

func TestLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	pachClient := getPachClient(t)
	job, err := ppsclient.CreateJob(
		pachClient,
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
	_, err = pachClient.InspectJob(context.Background(), inspectJobRequest)
	require.NoError(t, err)
	// TODO we Sleep here because even though the job has completed kubernetes
	// might not have even noticed the container was created yet
	time.Sleep(10 * time.Second)
	var buffer bytes.Buffer
	require.NoError(t, ppsclient.GetLogs(pachClient, job.ID, &buffer))
	require.Equal(t, "0 | foo\n1 | foo\n2 | foo\n3 | foo\n", buffer.String())
}

func TestGrep(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	dataRepo := uniqueString("TestGrep.data")
	pachClient := getPachClient(t)
	require.NoError(t, pfsclient.CreateRepo(pachClient, dataRepo))
	commit, err := pfsclient.StartCommit(pachClient, dataRepo, "", "")
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, err = pfsclient.PutFile(pachClient, dataRepo, commit.ID, fmt.Sprintf("file%d", i), strings.NewReader("foo\nbar\nfizz\nbuzz\n"))
		require.NoError(t, err)
	}
	require.NoError(t, pfsclient.FinishCommit(pachClient, dataRepo, commit.ID))
	job1, err := ppsclient.CreateJob(
		pachClient,
		"",
		[]string{"bash"},
		[]string{fmt.Sprintf("grep foo /pfs/%s/* >/pfs/out/foo", dataRepo)},
		1,
		[]*ppsclient.JobInput{{Commit: commit}},
		"",
	)
	require.NoError(t, err)
	job2, err := ppsclient.CreateJob(
		pachClient,
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
	job1Info, err := pachClient.InspectJob(context.Background(), inspectJobRequest)
	require.NoError(t, err)
	inspectJobRequest.Job = job2
	job2Info, err := pachClient.InspectJob(context.Background(), inspectJobRequest)
	require.NoError(t, err)
	repo1Info, err := pfsclient.InspectRepo(pachClient, job1Info.OutputCommit.Repo.Name)
	require.NoError(t, err)
	repo2Info, err := pfsclient.InspectRepo(pachClient, job2Info.OutputCommit.Repo.Name)
	require.NoError(t, err)
	require.Equal(t, repo1Info.SizeBytes, repo2Info.SizeBytes)
}

func TestPipeline(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	pachClient := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestPipeline.data")
	require.NoError(t, pfsclient.CreateRepo(pachClient, dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	outRepo := ppsserver.PipelineRepo(ppsclient.NewPipeline(pipelineName))
	require.NoError(t, ppsclient.CreatePipeline(
		pachClient,
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		1,
		[]*ppsclient.PipelineInput{{Repo: &pfsclient.Repo{Name: dataRepo}}},
	))
	// Do first commit to repo
	commit1, err := pfsclient.StartCommit(pachClient, dataRepo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pachClient, dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pachClient, dataRepo, commit1.ID))
	listCommitRequest := &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err := pachClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pachClient, outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	// Do second commit to repo
	commit2, err := pfsclient.StartCommit(pachClient, dataRepo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pachClient, dataRepo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pachClient, dataRepo, commit2.ID))
	listCommitRequest = &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
		FromCommit: []*pfsclient.Commit{outCommits[0].Commit},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err = pachClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	require.NotNil(t, listCommitResponse.CommitInfo[0].ParentCommit)
	require.Equal(t, outCommits[0].Commit.ID, listCommitResponse.CommitInfo[0].ParentCommit.ID)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pachClient, outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())

	require.NoError(t, ppsclient.DeletePipeline(pachClient, pipelineName))

	pipelineInfos, err := pachClient.ListPipeline(context.Background(), &ppsclient.ListPipelineRequest{})
	require.NoError(t, err)
	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		require.True(t, pipelineInfo.Pipeline.Name != pipelineName)
	}

	// Do third commit to repo; this time pipeline should not run since it's been deleted
	commit3, err := pfsclient.StartCommit(pachClient, dataRepo, commit2.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pachClient, dataRepo, commit3.ID, "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pachClient, dataRepo, commit3.ID))

	// We will sleep a while to wait for the pipeline to actually get cancelled
	// Also if the pipeline didn't get cancelled (due to a bug), we sleep a while
	// to let the pipeline commit
	time.Sleep(5 * time.Second)
	listCommitRequest = &pfsclient.ListCommitRequest{
		Repo: []*pfsclient.Repo{outRepo},
	}
	listCommitResponse, err = pachClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	// there should only be two commits in the pipeline
	require.Equal(t, len(listCommitResponse.CommitInfo), 2)
}

func TestWorkload(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	pachClient := getPachClient(t)
	seed := time.Now().UnixNano()
	require.NoError(t, workload.RunWorkload(pachClient, pachClient, rand.New(rand.NewSource(seed)), 100))
}

func TestSharding(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	repo := uniqueString("TestSharding")
	pachClient := getPachClient(t)
	err := pfsclient.CreateRepo(pachClient, repo)
	require.NoError(t, err)
	commit, err := pfsclient.StartCommit(pachClient, repo, "", "")
	require.NoError(t, err)
	var wg sync.WaitGroup
	for i := 0; i < NUMFILES; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			rand := rand.New(rand.NewSource(int64(i)))
			_, err = pfsclient.PutFile(pachClient, repo, commit.ID, fmt.Sprintf("file%d", i), workload.NewReader(rand, KB))
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	err = pfsclient.FinishCommit(pachClient, repo, commit.ID)
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
			err := pfsclient.GetFile(pachClient, repo, commit.ID,
				fmt.Sprintf("file%d", i), 0, 0, "", shard, &buffer1Shard)
			require.NoError(t, err)
			shard.BlockModulus = 4
			for blockNumber := uint64(0); blockNumber < 4; blockNumber++ {
				shard.BlockNumber = blockNumber
				err := pfsclient.GetFile(pachClient, repo, commit.ID,
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
	pachClient := getPachClient(t)
	seed := time.Now().UnixNano()
	rand := rand.New(rand.NewSource(seed))
	err := pfsclient.CreateRepo(pachClient, repo)
	require.NoError(t, err)
	commit1, err := pfsclient.StartCommit(pachClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pachClient, repo, commit1.ID, "file", workload.NewReader(rand, KB))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pachClient, repo, commit1.ID)
	require.NoError(t, err)
	commit2, err := pfsclient.StartCommit(pachClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pachClient, repo, commit2.ID, "file", workload.NewReader(rand, KB))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pachClient, repo, commit2.ID)
	require.NoError(t, err)
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pachClient, repo, commit2.ID, "file", 0, 0, commit1.ID, nil, &buffer))
	require.Equal(t, buffer.Len(), KB)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pachClient, repo, commit2.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, buffer.Len(), 2*KB)
}

func TestSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	pachClient := getPachClient(t)
	repo := uniqueString("TestSimple")
	require.NoError(t, pfsclient.CreateRepo(pachClient, repo))
	commit1, err := pfsclient.StartCommit(pachClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pachClient, repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pachClient, repo, commit1.ID))
	commitInfos, err := pfsclient.ListCommit(pachClient, []string{repo}, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pachClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pfsclient.StartCommit(pachClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pachClient, repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pachClient, repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pachClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pachClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
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
