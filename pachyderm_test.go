package pachyderm

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"testing"

	"go.pedge.io/protolog"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/pkg/workload"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/ppsutil"
	"google.golang.org/grpc"
)

func TestJob(t *testing.T) {
	t.Parallel()
	dataRepo := uniqueString("TestJob.data")
	pfsClient := getPfsClient(t)
	require.NoError(t, pfsutil.CreateRepo(pfsClient, dataRepo))
	commit, err := pfsutil.StartCommit(pfsClient, dataRepo, "")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, dataRepo, commit.Id, "file", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsutil.FinishCommit(pfsClient, dataRepo, commit.Id))
	ppsClient := getPpsClient(t)
	job, err := ppsutil.CreateJob(
		ppsClient,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		"",
		1,
		[]*pps.JobInput{{Commit: commit}},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &pps.InspectJobRequest{
		Job:         job,
		BlockOutput: true,
		BlockState:  true,
	}
	jobInfo, err := ppsClient.InspectJob(context.Background(), inspectJobRequest)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_STATE_SUCCESS.String(), jobInfo.State.String())
	commitInfo, err := pfsutil.InspectCommit(pfsClient, jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.Id)
	require.NoError(t, err)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	var buffer bytes.Buffer
	require.NoError(t, pfsutil.GetFile(pfsClient, jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.Id, "file", 0, 0, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

func TestGrep(t *testing.T) {
	t.Parallel()
	dataRepo := uniqueString("pachyderm.TestGrep.data")
	pfsClient := getPfsClient(t)
	require.NoError(t, pfsutil.CreateRepo(pfsClient, dataRepo))
	commit, err := pfsutil.StartCommit(pfsClient, dataRepo, "")
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, err = pfsutil.PutFile(pfsClient, dataRepo, commit.Id, fmt.Sprintf("file%d", i), 0, strings.NewReader("foo\nbar\nfizz\nbuzz\n"))
		require.NoError(t, err)
	}
	require.NoError(t, pfsutil.FinishCommit(pfsClient, dataRepo, commit.Id))
	ppsClient := getPpsClient(t)
	job1, err := ppsutil.CreateJob(
		ppsClient,
		"",
		[]string{"bash"},
		fmt.Sprintf("grep foo /pfs/%s/* >/pfs/out/foo", dataRepo),
		1,
		[]*pps.JobInput{{Commit: commit}},
		"",
	)
	require.NoError(t, err)
	job2, err := ppsutil.CreateJob(
		ppsClient,
		"",
		[]string{"bash"},
		fmt.Sprintf("grep foo /pfs/%s/* >/pfs/out/foo", dataRepo),
		4,
		[]*pps.JobInput{{Commit: commit}},
		"",
	)
	require.NoError(t, err)
	inspectJobRequest := &pps.InspectJobRequest{
		Job:         job1,
		BlockOutput: true,
		BlockState:  true,
	}
	job1Info, err := ppsClient.InspectJob(context.Background(), inspectJobRequest)
	require.NoError(t, err)
	inspectJobRequest.Job = job2
	job2Info, err := ppsClient.InspectJob(context.Background(), inspectJobRequest)
	require.NoError(t, err)
	repo1Info, err := pfsutil.InspectRepo(pfsClient, job1Info.OutputCommit.Repo.Name)
	require.NoError(t, err)
	repo2Info, err := pfsutil.InspectRepo(pfsClient, job2Info.OutputCommit.Repo.Name)
	require.NoError(t, err)
	require.Equal(t, repo1Info.SizeBytes, repo2Info.SizeBytes)
}

func TestPipeline(t *testing.T) {
	t.Parallel()
	pfsClient := getPfsClient(t)
	ppsClient := getPpsClient(t)
	// create repos
	dataRepo := uniqueString("TestPipeline.data")
	require.NoError(t, pfsutil.CreateRepo(pfsClient, dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	outRepo := pps.PipelineRepo(ppsutil.NewPipeline(pipelineName))
	require.NoError(t, ppsutil.CreatePipeline(
		ppsClient,
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		"",
		1,
		[]*pps.PipelineInput{{Repo: &pfs.Repo{Name: dataRepo}}},
	))
	// Do first commit to repo
	commit1, err := pfsutil.StartCommit(pfsClient, dataRepo, "")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, dataRepo, commit1.Id, "file", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsutil.FinishCommit(pfsClient, dataRepo, commit1.Id))
	listCommitRequest := &pfs.ListCommitRequest{
		Repo:       []*pfs.Repo{outRepo},
		CommitType: pfs.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err := pfsClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits := listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	var buffer bytes.Buffer
	require.NoError(t, pfsutil.GetFile(pfsClient, outRepo.Name, outCommits[0].Commit.Id, "file", 0, 0, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	// Do second commit to repo
	commit2, err := pfsutil.StartCommit(pfsClient, dataRepo, commit1.Id)
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, dataRepo, commit2.Id, "file", 0, strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, pfsutil.FinishCommit(pfsClient, dataRepo, commit2.Id))
	listCommitRequest = &pfs.ListCommitRequest{
		Repo:       []*pfs.Repo{outRepo},
		FromCommit: []*pfs.Commit{outCommits[0].Commit},
		CommitType: pfs.CommitType_COMMIT_TYPE_READ,
		Block:      true,
	}
	listCommitResponse, err = pfsClient.ListCommit(
		context.Background(),
		listCommitRequest,
	)
	require.NoError(t, err)
	outCommits = listCommitResponse.CommitInfo
	require.Equal(t, 1, len(outCommits))
	buffer = bytes.Buffer{}
	require.NoError(t, pfsutil.GetFile(pfsClient, outRepo.Name, outCommits[0].Commit.Id, "file", 0, 0, nil, &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())
}

func TestWorkload(t *testing.T) {
	t.Parallel()
	pfsClient := getPfsClient(t)
	ppsClient := getPpsClient(t)
	//seed := time.Now().UnixNano()
	seed := int64(7)
	require.NoError(t, workload.RunWorkload(pfsClient, ppsClient, rand.New(rand.NewSource(seed)), 100))
}

func TestBigWrite(t *testing.T) {
	t.Parallel()
	protolog.SetLevel(protolog.Level_LEVEL_DEBUG)
	repo := uniqueString("TestBigWrite")
	pfsClient := getPfsClient(t)
	err := pfsutil.CreateRepo(pfsClient, repo)
	require.NoError(t, err)
	commit, err := pfsutil.StartCommit(pfsClient, repo, "")
	require.NoError(t, err)
	rand := rand.New(rand.NewSource(5))
	_, err = pfsutil.PutFile(pfsClient, repo, commit.Id, "file", 0, workload.NewReader(rand, 10000))
	require.NoError(t, err)
	err = pfsutil.FinishCommit(pfsClient, repo, commit.Id)
	require.NoError(t, err)
	var buffer bytes.Buffer
	err = pfsutil.GetFile(pfsClient, repo, commit.Id, "file", 0, 0, nil, &buffer)
	require.NoError(t, err)
}

func getPfsClient(t *testing.T) pfs.APIClient {
	pfsdAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR")
	if pfsdAddr == "" {
		t.Error("PFSD_PORT_650_TCP_ADDR not set")
	}
	clientConn, err := grpc.Dial(fmt.Sprintf("%s:650", pfsdAddr), grpc.WithInsecure())
	require.NoError(t, err)
	return pfs.NewAPIClient(clientConn)
}

func getPpsClient(t *testing.T) pps.APIClient {
	ppsdAddr := os.Getenv("PPSD_PORT_651_TCP_ADDR")
	if ppsdAddr == "" {
		t.Error("PPSD_PORT_651_TCP_ADDR not set")
	}
	clientConn, err := grpc.Dial(fmt.Sprintf("%s:651", ppsdAddr), grpc.WithInsecure())
	require.NoError(t, err)
	return pps.NewAPIClient(clientConn)
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
