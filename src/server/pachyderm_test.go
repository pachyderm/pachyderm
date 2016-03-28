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
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
)

const (
	NUMFILES = 25
	KB       = 1024 * 1024
)

func TestJob(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	dataRepo := uniqueString("TestJob.data")
	pachClient := getPachClient(t)
	require.NoError(t, pachClient.CreateRepo(dataRepo))
	commit, err := pachClient.StartCommit(dataRepo, "", "")
	require.NoError(t, err)
	_, err = pachClient.PutFile(dataRepo, commit.ID, "file", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pachClient.FinishCommit(dataRepo, commit.ID))
	job, err := ppsclient.CreateJob(
		pachClient,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		1,
		[]*ppsclient.JobInput{{Commit: commit}},
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
	require.Equal(t, ppsclient.JobState_JOB_STATE_SUCCESS.String(), jobInfo.State.String())
	commitInfo, err := pachClient.InspectCommit(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	require.NoError(t, err)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	var buffer bytes.Buffer
	require.NoError(t, pachClient.GetFile(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

func TestGrep(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	dataRepo := uniqueString("TestGrep.data")
	pachClient := getPachClient(t)
	require.NoError(t, pachClient.CreateRepo(dataRepo))
	commit, err := pachClient.StartCommit(dataRepo, "", "")
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, err = pachClient.PutFile(dataRepo, commit.ID, fmt.Sprintf("file%d", i), 0, strings.NewReader("foo\nbar\nfizz\nbuzz\n"))
		require.NoError(t, err)
	}
	require.NoError(t, pachClient.FinishCommit(dataRepo, commit.ID))
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
	repo1Info, err := pachClient.InspectRepo(job1Info.OutputCommit.Repo.Name)
	require.NoError(t, err)
	repo2Info, err := pachClient.InspectRepo(job2Info.OutputCommit.Repo.Name)
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
	require.NoError(t, pachClient.CreateRepo(dataRepo))
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
	commit1, err := pachClient.StartCommit(dataRepo, "", "")
	require.NoError(t, err)
	_, err = pachClient.PutFile(dataRepo, commit1.ID, "file", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pachClient.FinishCommit(dataRepo, commit1.ID))
/*	listCommitRequest := &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}*/
	outCommits, err := pachClient.ListCommit([]string{outRepo.Name})
	require.NoError(t, err)
	require.Equal(t, 1, len(outCommits))
	var buffer bytes.Buffer
	require.NoError(t, pachClient.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	// Do second commit to repo
	commit2, err := pachClient.StartCommit(dataRepo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pachClient.PutFile(dataRepo, commit2.ID, "file", 0, strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, pachClient.FinishCommit(dataRepo, commit2.ID))

/*	listCommitRequest = &pfsclient.ListCommitRequest{
		Repo:       []*pfsclient.Repo{outRepo},
		FromCommit: []*pfsclient.Commit{outCommits[0].Commit},
		CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err = pachClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)*/
	lastCommits := outCommits
	outCommits, err = pachClient.ListCommit([]string{outRepo.Name})
	require.NoError(t, err)
	require.NotNil(t, outCommits[0].ParentCommit)
	require.Equal(t, lastCommits[0].Commit.ID, outCommits[0].ParentCommit.ID)
	require.Equal(t, 1, len(outCommits))
	buffer = bytes.Buffer{}
	require.NoError(t, pachClient.GetFile(outRepo.Name, outCommits[0].Commit.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())
}

func TestWorkload(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	pachClient := getPachClient(t)
	seed := time.Now().UnixNano()
	require.NoError(t, workload.RunWorkload(pachClient.PfsAPIClient, pachClient, rand.New(rand.NewSource(seed)), 100))
}

func TestSharding(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()
	repo := uniqueString("TestSharding")
	pachClient := getPachClient(t)
	err := pachClient.CreateRepo(repo)
	require.NoError(t, err)
	commit, err := pachClient.StartCommit(repo, "", "")
	require.NoError(t, err)
	var wg sync.WaitGroup
	for i := 0; i < NUMFILES; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			rand := rand.New(rand.NewSource(int64(i)))
			_, err = pachClient.PutFile(repo, commit.ID, fmt.Sprintf("file%d", i), 0, workload.NewReader(rand, KB))
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	err = pachClient.FinishCommit(repo, commit.ID)
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
			err := pachClient.GetFile(repo, commit.ID,
				fmt.Sprintf("file%d", i), 0, 0, "", shard, &buffer1Shard)
			require.NoError(t, err)
			shard.BlockModulus = 4
			for blockNumber := uint64(0); blockNumber < 4; blockNumber++ {
				shard.BlockNumber = blockNumber
				err := pachClient.GetFile(repo, commit.ID,
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
	err := pachClient.CreateRepo(repo)
	require.NoError(t, err)
	commit1, err := pachClient.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = pachClient.PutFile(repo, commit1.ID, "file", 0, workload.NewReader(rand, KB))
	require.NoError(t, err)
	err = pachClient.FinishCommit(repo, commit1.ID)
	require.NoError(t, err)
	commit2, err := pachClient.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pachClient.PutFile(repo, commit2.ID, "file", 0, workload.NewReader(rand, KB))
	require.NoError(t, err)
	err = pachClient.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)
	var buffer bytes.Buffer
	require.NoError(t, pachClient.GetFile(repo, commit2.ID, "file", 0, 0, commit1.ID, nil, &buffer))
	require.Equal(t, buffer.Len(), KB)
	buffer = bytes.Buffer{}
	require.NoError(t, pachClient.GetFile(repo, commit2.ID, "file", 0, 0, "", nil, &buffer))
	require.Equal(t, buffer.Len(), 2*KB)
}

func TestSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	pachClient := getPachClient(t)
	repo := uniqueString("TestSimple")
	require.NoError(t, pachClient.CreateRepo(repo))
	commit1, err := pachClient.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = pachClient.PutFile(repo, commit1.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pachClient.FinishCommit(repo, commit1.ID))
	commitInfos, err := pachClient.ListCommit([]string{repo})
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, pachClient.GetFile(repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pachClient.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pachClient.PutFile(repo, commit2.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pachClient.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pachClient.GetFile(repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pachClient.GetFile(repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func getPachClient(t *testing.T) *client.APIClient {
	client, err := client.New()
	require.NoError(t, err)
	return client
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
