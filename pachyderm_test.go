package pachyderm

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/ppsutil"
	"google.golang.org/grpc"
)

func TestJob(t *testing.T) {
	dataRepo := uniqueString("TestJob.data")
	pfsClient := getPfsClient(t)
	require.NoError(t, pfsutil.CreateRepo(pfsClient, dataRepo))
	commit, err := pfsutil.StartCommit(pfsClient, dataRepo, "")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, dataRepo, commit.Id, "file", 0, strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, pfsutil.FinishCommit(pfsClient, dataRepo, commit.Id))
	ppsClient := getPpsClient(t)
	job, err := ppsutil.CreateJob(
		ppsClient,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		"",
		1,
		[]*pfs.Commit{commit},
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
	var buffer bytes.Buffer
	require.NoError(t, pfsutil.GetFile(pfsClient, jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.Id, "file", 0, 0, nil, &buffer))
	require.Equal(t, "foo", buffer.String())
}

func TestGrep(t *testing.T) {
	t.Skip()
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
	_, err = ppsutil.CreateJob(
		ppsClient,
		"",
		[]string{"sh"},
		fmt.Sprintf("grep foo /pfs/%s/* >/pfs/out/foo", dataRepo),
		1,
		[]*pfs.Commit{commit},
		"",
	)
	require.NoError(t, err)
}

func TestPipeline(t *testing.T) {
	t.Skip()
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
		[]*pfs.Repo{&pfs.Repo{Name: dataRepo}},
	))
	// Do first commit to repo
	log.Printf("Do first commit.")
	commit1, err := pfsutil.StartCommit(pfsClient, dataRepo, "")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, dataRepo, commit1.Id, "file", 0, strings.NewReader("foo"))
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
	log.Printf("First commit: %+v", outCommits[0])
	require.NoError(t, pfsutil.GetFile(pfsClient, outRepo.Name, outCommits[0].Commit.Id, "file", 0, 0, nil, &buffer))
	require.Equal(t, "foo", buffer.String())
	// Do second commit to repo
	log.Printf("Do second commit.")
	commit2, err := pfsutil.StartCommit(pfsClient, dataRepo, commit1.Id)
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, dataRepo, commit2.Id, "file", 0, strings.NewReader("bar"))
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
	log.Printf("Second commit: %+v", outCommits[0])
	require.NoError(t, pfsutil.GetFile(pfsClient, outRepo.Name, outCommits[0].Commit.Id, "file", 0, 0, nil, &buffer))
	require.Equal(t, "foobar", buffer.String())
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
