package server

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
)

func TestSimplePipelineV2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestSimplePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := tu.UniqueString("TestSimplePipeline")
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
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 2, len(commitInfos))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo", buf.String())
}

func TestRepoSizeV2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	// create a data repo
	dataRepo := tu.UniqueString("TestRepoSize_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	// create a pipeline
	pipeline := tu.UniqueString("TestRepoSize")
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
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))

	// put a file without an open commit - should count towards repo size
	_, err := c.PutFile(dataRepo, "master", "file2", strings.NewReader("foo"))
	require.NoError(t, err)

	// put a file on another branch - should not count towards repo size
	_, err = c.PutFile(dataRepo, "develop", "file3", strings.NewReader("foo"))
	require.NoError(t, err)

	// put a file on an open commit - should count toward repo size
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file1", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	// wait for everything to be processed
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 2, len(commitInfos))

	// check data repo size
	repoInfo, err := c.InspectRepo(dataRepo)
	require.NoError(t, err)
	require.Equal(t, uint64(6), repoInfo.SizeBytes)

	// check pipeline repo size
	repoInfo, err = c.InspectRepo(pipeline)
	require.NoError(t, err)
	require.Equal(t, uint64(6), repoInfo.SizeBytes)

	// ensure size is updated when we delete a commit
	require.NoError(t, c.DeleteCommit(dataRepo, commit1.ID))
	repoInfo, err = c.InspectRepo(dataRepo)
	require.NoError(t, err)
	require.Equal(t, uint64(3), repoInfo.SizeBytes)
	repoInfo, err = c.InspectRepo(pipeline)
	require.NoError(t, err)
	require.Equal(t, uint64(3), repoInfo.SizeBytes)
}

func TestPipelineWithParallelismV2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestPipelineWithParallelism_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	numFiles := 1000
	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i), strings.NewReader(fmt.Sprintf("%d", i)))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := tu.UniqueString("pipeline")
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
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 2, len(commitInfos))

	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, fmt.Sprintf("file-%d", i), 0, 0, &buf))
		require.Equal(t, fmt.Sprintf("%d", i), buf.String())
	}
}

func TestPipelineWithLargeFilesV2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestPipelineWithLargeFiles_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	r := rand.New(rand.NewSource(99))
	numFiles := 10
	var fileContents []string

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	chunkSize := int(pfs.ChunkSize / 32) // We used to use a full ChunkSize, but it was increased which caused this test to take too long.
	for i := 0; i < numFiles; i++ {
		fileContent := workload.RandString(r, chunkSize+i*MB)
		_, err = c.PutFile(dataRepo, commit1.ID, fmt.Sprintf("file-%d", i),
			strings.NewReader(fileContent))
		require.NoError(t, err)
		fileContents = append(fileContents, fileContent)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	pipeline := tu.UniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		nil,
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 2, len(commitInfos))

	commit := commitInfos[0].Commit

	for i := 0; i < numFiles; i++ {
		var buf bytes.Buffer
		fileName := fmt.Sprintf("file-%d", i)

		fileInfo, err := c.InspectFile(commit.Repo.Name, commit.ID, fileName)
		require.NoError(t, err)
		require.Equal(t, chunkSize+i*MB, int(fileInfo.SizeBytes))

		require.NoError(t, c.GetFile(commit.Repo.Name, commit.ID, fileName, 0, 0, &buf))
		// we don't wanna use the `require` package here since it prints
		// the strings, which would clutter the output.
		if fileContents[i] != buf.String() {
			t.Fatalf("file content does not match")
		}
	}
}
