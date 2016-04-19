package server

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/context"

	"go.pedge.io/proto/server"
	"google.golang.org/grpc"

	pclient "github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
)

const (
	shards  = 32
	servers = 4

	ALPHABET = "abcdefghijklmnopqrstuvwxyz"
)

var (
	port int32 = 30651
)

func TestBlock(t *testing.T) {
	t.Parallel()
	blockClient := getBlockClient(t)
	_, err := blockClient.CreateDiff(
		context.Background(),
		&pfsclient.DiffInfo{
			Diff: pfsclient.NewDiff("foo", "", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfsclient.DiffInfo{
			Diff: pfsclient.NewDiff("foo", "c1", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfsclient.DiffInfo{
			Diff: pfsclient.NewDiff("foo", "c2", 0),
		})
	require.NoError(t, err)
	listDiffClient, err := blockClient.ListDiff(
		context.Background(),
		&pfsclient.ListDiffRequest{Shard: 0},
	)
	require.NoError(t, err)
	var diffInfos []*pfsclient.DiffInfo
	for {
		diffInfo, err := listDiffClient.Recv()
		if err == io.EOF {
			break
		} else {
			require.NoError(t, err)
		}
		diffInfos = append(diffInfos, diffInfo)
	}
	require.Equal(t, 3, len(diffInfos))
}

func TestInvalidRepo(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)
	require.YesError(t, pfsclient.CreateRepo(pfsClient, "/repo"))
}

func TestSimple(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))
	commitInfos, err := pfsclient.ListCommit(pfsClient, []string{repo}, nil, false, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pfsClient, repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestBranch(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	branches, err := pfsclient.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit1, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
	commit2, err := pfsclient.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pfsClient, repo, "master")
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = pfsclient.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit2, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = pfsclient.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit2, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
}

func TestDisallowReadsDuringCommit(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)

	// Make sure we can't get the file before the commit is finished
	var buffer bytes.Buffer
	require.YesError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "", buffer.String())

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pfsClient, repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestInspectRepoSimple(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	commit, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	file1Content := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit.ID, "foo", strings.NewReader(file1Content))
	require.NoError(t, err)

	file2Content := "bar\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit.ID, "bar", strings.NewReader(file2Content))
	require.NoError(t, err)

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit.ID))

	info, err := pfsclient.InspectRepo(pfsClient, repo)
	require.NoError(t, err)

	require.Equal(t, int(info.SizeBytes), len(file1Content)+len(file2Content))
}

func TestInspectRepoComplex(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	commit, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	numFiles := 100
	minFileSize := 1000
	maxFileSize := 2000
	totalSize := 0

	for i := 0; i < numFiles; i++ {
		fileContent := generateRandomString(rand.Intn(maxFileSize-minFileSize) + minFileSize)
		fileContent += "\n"
		fileName := fmt.Sprintf("file_%d", i)
		totalSize += len(fileContent)

		_, err = pfsclient.PutFile(pfsClient, repo, commit.ID, fileName, strings.NewReader(fileContent))
		require.NoError(t, err)
	}

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit.ID))

	info, err := pfsclient.InspectRepo(pfsClient, repo)
	require.NoError(t, err)

	require.Equal(t, int(info.SizeBytes), totalSize)

	infos, err := pfsclient.ListRepo(pfsClient)
	require.NoError(t, err)
	require.Equal(t, 1, len(infos))
	info = infos[0]

	require.Equal(t, int(info.SizeBytes), totalSize)
}

func TestListRepo(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)

	numRepos := 10
	var repoNames []string
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))
		repoNames = append(repoNames, repo)
	}

	test := func() {
		repoInfos, err := pfsclient.ListRepo(pfsClient)
		require.NoError(t, err)

		for i, repoInfo := range repoInfos {
			require.Equal(t, repoNames[len(repoNames)-i-1], repoInfo.Repo.Name)
		}

		require.Equal(t, len(repoInfos), numRepos)
	}

	test()

	restartServer(server, t)

	test()
}

func TestDeleteRepo(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	numRepos := 10
	repoNames := make(map[string]bool)
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))
		repoNames[repo] = true
	}

	reposToRemove := 5
	for i := 0; i < reposToRemove; i++ {
		// Pick one random element from repoNames
		for repoName := range repoNames {
			require.NoError(t, pfsclient.DeleteRepo(pfsClient, repoName))
			delete(repoNames, repoName)
			break
		}
	}

	repoInfos, err := pfsclient.ListRepo(pfsClient)
	require.NoError(t, err)

	for _, repoInfo := range repoInfos {
		require.True(t, repoNames[repoInfo.Repo.Name])
	}

	require.Equal(t, len(repoInfos), numRepos-reposToRemove)
}

func TestInspectCommit(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	started := time.Now()
	commit, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	commitInfo, err := pfsclient.InspectCommit(pfsClient, repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_WRITE, commitInfo.CommitType)
	require.Equal(t, len(fileContent), int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.Nil(t, commitInfo.Finished)

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit.ID))
	finished := time.Now()

	commitInfo, err = pfsclient.InspectCommit(pfsClient, repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	require.Equal(t, len(fileContent), int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.True(t, finished.After(commitInfo.Finished.GoTime()))
}

func TestDeleteCommitFuture(t *testing.T) {
	// For when DeleteCommit gets implemented
	t.Skip()

	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	commit, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit.ID))

	commitInfo, err := pfsclient.InspectCommit(pfsClient, repo, commit.ID)
	require.NotNil(t, commitInfo)

	require.NoError(t, pfsclient.DeleteCommit(pfsClient, repo, commit.ID))

	commitInfo, err = pfsclient.InspectCommit(pfsClient, repo, commit.ID)
	require.Nil(t, commitInfo)

	repoInfo, err := pfsclient.InspectRepo(pfsClient, repo)
	require.Equal(t, 0, repoInfo.SizeBytes)
}

func TestDeleteCommit(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	commit, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit.ID))

	// Because DeleteCommit is not supported
	require.YesError(t, pfsclient.DeleteCommit(pfsClient, repo, commit.ID))
}

func TestPutFile(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo/bar", strings.NewReader("foo\n"))
	require.YesError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))

	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())

	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "foo/bar", strings.NewReader("foo\n"))
	require.YesError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "/bar", strings.NewReader("bar\n"))
	require.YesError(t, err) // because path starts with a slash
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit2.ID))

	commit3, err := pfsclient.StartCommit(pfsClient, repo, commit2.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit3.ID, "dir1/foo", strings.NewReader("foo\n"))
	require.NoError(t, err) // because the directory dir does not exist
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit3.ID))

	commit4, err := pfsclient.StartCommit(pfsClient, repo, commit3.ID, "")
	require.NoError(t, err)
	require.NoError(t, pfsclient.MakeDirectory(pfsClient, repo, commit4.ID, "dir2"))
	_, err = pfsclient.PutFile(pfsClient, repo, commit4.ID, "dir2/bar", strings.NewReader("bar\n"))
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit4.ID, "dir1", strings.NewReader("foo\n"))
	require.YesError(t, err) // because dir1 is a directory
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit4.ID))

	var buffer2 bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit4.ID, "dir2/bar", 0, 0, "", nil, &buffer2))
	require.Equal(t, "bar\n", buffer2.String())
}

func TestInspectFile(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	fileContent1 := "foo\n"
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))

	fileInfo, err := pfsclient.InspectFile(pfsClient, repo, commit1.ID, "foo", "", nil)
	require.NoError(t, err)
	require.Equal(t, commit1, fileInfo.CommitModified)
	require.Equal(t, pfsclient.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)
	require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))

	fileContent2 := "barbar\n"
	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "foo", strings.NewReader(fileContent2))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit2.ID))

	fileInfo, err = pfsclient.InspectFile(pfsClient, repo, commit2.ID, "foo", commit1.ID, nil)
	require.NoError(t, err)
	require.Equal(t, commit2, fileInfo.CommitModified)
	require.Equal(t, pfsclient.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)
	require.Equal(t, len(fileContent2), int(fileInfo.SizeBytes))

	fileInfo, err = pfsclient.InspectFile(pfsClient, repo, commit2.ID, "foo", "", nil)
	require.NoError(t, err)
	require.Equal(t, commit2, fileInfo.CommitModified)
	require.Equal(t, pfsclient.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)
	require.Equal(t, len(fileContent1)+len(fileContent2), int(fileInfo.SizeBytes))

	fileContent3 := "bar\n"
	commit3, err := pfsclient.StartCommit(pfsClient, repo, commit2.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit3.ID, "bar", strings.NewReader(fileContent3))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit3.ID))

	fileInfos, err := pfsclient.ListFile(pfsClient, repo, commit3.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, len(fileInfos), 2)
}

func TestListFile(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	commit, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit.ID, "dir/foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	fileContent2 := "bar\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit.ID, "dir/bar", strings.NewReader(fileContent2))
	require.NoError(t, err)

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit.ID))

	fileInfos, err := pfsclient.ListFile(pfsClient, repo, commit.ID, "dir", "", nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
	require.True(t, fileInfos[0].File.Path == "dir/foo" && fileInfos[1].File.Path == "dir/bar" || fileInfos[0].File.Path == "dir/bar" && fileInfos[1].File.Path == "dir/foo")

	fileInfos, err = pfsclient.ListFile(pfsClient, repo, commit.ID, "dir/foo", "", nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))
	require.True(t, fileInfos[0].File.Path == "dir/foo")
}

func TestDeleteFile(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	// Commit 1: Add two files; delete one file within the commit
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	fileContent2 := "bar\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "bar", strings.NewReader(fileContent2))
	require.NoError(t, err)

	require.NoError(t, pfsclient.DeleteFile(pfsClient, repo, commit1.ID, "foo"))

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))

	// foo should not exist
	_, err = pfsclient.InspectFile(pfsClient, repo, commit1.ID, "foo", "", nil)
	require.YesError(t, err)

	// Should see one file
	fileInfos, err := pfsclient.ListFile(pfsClient, repo, commit1.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))
	require.Equal(t, fileInfos[0].File.Path, "bar")

	// Empty commit
	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit2.ID))

	// Should still see one file
	fileInfos, err = pfsclient.ListFile(pfsClient, repo, commit2.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))
	require.Equal(t, fileInfos[0].File.Path, "bar")

	// Delete the other file
	commit3, err := pfsclient.StartCommit(pfsClient, repo, commit2.ID, "")
	require.NoError(t, err)
	require.NoError(t, pfsclient.DeleteFile(pfsClient, repo, commit3.ID, "bar"))
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit3.ID))

	// Should see zero files
	fileInfos, err = pfsclient.ListFile(pfsClient, repo, commit3.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// The removed file should not exist
	_, err = pfsclient.InspectFile(pfsClient, repo, commit3.ID, "bar", "", nil)
	require.YesError(t, err)
}

func TestInspectDir(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	// Commit 1: Add two files into the same directory; delete the directory
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "dir/foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))

	_, err = pfsclient.InspectFile(pfsClient, repo, commit1.ID, "dir/foo", "", nil)
	require.NoError(t, err)

	_, err = pfsclient.InspectFile(pfsClient, repo, commit1.ID, "dir", "", nil)
	require.NoError(t, err)

	// This is a limitation in our system: we cannot inspect .
	// . is assumed to be a directory
	// In order to be able to inspect the root directory, we have to have each
	// PutFile send a concurrent request to create an entry for ".", which is
	// a price we are not willing to pay.
	_, err = pfsclient.InspectFile(pfsClient, repo, commit1.ID, "", "", nil)
	require.YesError(t, err)
}

func TestDeleteDir(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	// Commit 1: Add two files into the same directory; delete the directory
	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "dir/foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	fileContent2 := "bar\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "dir/bar", strings.NewReader(fileContent2))
	require.NoError(t, err)

	require.NoError(t, pfsclient.DeleteFile(pfsClient, repo, commit1.ID, "dir"))

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))

	// Should see zero files
	fileInfos, err := pfsclient.ListFile(pfsClient, repo, commit1.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// dir should not exist
	_, err = pfsclient.InspectFile(pfsClient, repo, commit1.ID, "dir", "", nil)
	require.YesError(t, err)

	// Commit 2: Add two files into the same directory
	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)

	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "dir/foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "dir/bar", strings.NewReader(fileContent2))
	require.NoError(t, err)

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit2.ID))

	// Should see two files
	fileInfos, err = pfsclient.ListFile(pfsClient, repo, commit2.ID, "dir", "", nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	// Commit 3: delete the directory
	commit3, err := pfsclient.StartCommit(pfsClient, repo, commit2.ID, "")
	require.NoError(t, err)

	require.NoError(t, pfsclient.DeleteFile(pfsClient, repo, commit3.ID, "dir"))

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit3.ID))

	// Should see zero files
	fileInfos, err = pfsclient.ListFile(pfsClient, repo, commit3.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// TODO: test deleting "."
}

func TestListCommit(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	commit, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, commit.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit.ID))

	commitInfos, err := pfsclient.ListCommit(pfsClient, []string{repo}, nil, false, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	// test the block behaviour
	ch := make(chan bool)
	go func() {
		_, err = pfsclient.ListCommit(pfsClient, []string{repo}, []string{commit.ID}, true, false)
		close(ch)
	}()

	time.Sleep(time.Second)
	select {
	case <-ch:
		t.Fatal("ListCommit should not have returned")
	default:
	}

	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit.ID, "")
	require.NoError(t, err)

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit2.ID))

	time.Sleep(time.Second)
	select {
	case <-ch:
	default:
		t.Fatal("ListCommit should have returned")
	}

	// test that cancelled commits are not listed
	commit3, err := pfsclient.StartCommit(pfsClient, repo, commit2.ID, "")
	require.NoError(t, err)

	require.NoError(t, pfsclient.CancelCommit(pfsClient, repo, commit3.ID))
	commitInfos, err = pfsclient.ListCommit(pfsClient, []string{repo}, nil, false, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	commitInfos, err = pfsclient.ListCommit(pfsClient, []string{repo}, nil, false, true)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	require.Equal(t, commit3, commitInfos[0].Commit)
}

func TestOffsetRead(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)
	repo := "TestOffsetRead"
	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))
	_, err := pfsclient.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	fileData := "foo\n"
	_, err = pfsclient.PutFile(pfsClient, repo, "master", "foo", strings.NewReader(fileData))
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, "master", "foo", strings.NewReader(fileData))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", int64(len(fileData)*2)+1, 0, "", nil, &buffer))
	require.Equal(t, "", buffer.String())
}

// FinishCommit should block until the parent has been finished
func TestFinishCommit(t *testing.T) {
	t.Parallel()

	pfsClient, _ := getClientAndServer(t)
	repo := "TestFinishCommit"

	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)

	ch := make(chan bool)
	go func() {
		require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit2.ID))
		close(ch)
	}()

	time.Sleep(time.Second * 2)
	select {
	case <-ch:
		t.Fatalf("should block, since the parent commit has not been finished")
	default:
	}

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))

	time.Sleep(time.Second * 2)
	select {
	case <-ch:
	default:
		t.Fatalf("should not block, since the parent commit has been finished")
	}
}

// If a commit's parent has been cancelled, the commit should be cancelled too
func TestFinishCommitParentCancelled(t *testing.T) {
	t.Parallel()

	pfsClient, _ := getClientAndServer(t)
	repo := "TestFinishCommitParentCancelled"

	require.NoError(t, pfsclient.CreateRepo(pfsClient, repo))

	commit1, err := pfsclient.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)

	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)

	ch := make(chan bool)
	go func() {
		require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit2.ID))
		close(ch)
	}()

	require.NoError(t, pfsclient.CancelCommit(pfsClient, repo, commit1.ID))

	time.Sleep(time.Second * 2)
	select {
	case <-ch:
	default:
		t.Fatalf("should not block, since the parent commit has been finished")
	}

	commit2Info, err := pfsclient.InspectCommit(pfsClient, repo, commit2.ID)
	require.True(t, commit2Info.Cancelled)

	commit3, err := pfsclient.StartCommit(pfsClient, repo, commit2.ID, "")
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit3.ID))
	commit3Info, err := pfsclient.InspectCommit(pfsClient, repo, commit3.ID)
	require.True(t, commit3Info.Cancelled)
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ALPHABET[rand.Intn(len(ALPHABET))]
	}
	return string(b)
}

func getBlockClient(t *testing.T) pfsclient.BlockAPIClient {
	localPort := atomic.AddInt32(&port, 1)
	address := fmt.Sprintf("localhost:%d", localPort)
	root := uniqueString("/tmp/pach_test/run")
	t.Logf("root %s", root)
	blockAPIServer, err := NewLocalBlockAPIServer(root)
	require.NoError(t, err)
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: pclient.Version},
			protoserver.ServeEnv{GRPCPort: uint16(localPort)},
		)
		require.NoError(t, err)
	}()
	<-ready
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	return pfsclient.NewBlockAPIClient(clientConn)
}

func runServers(t *testing.T, port int32, apiServer pfsclient.APIServer,
	internalAPIServer pfsclient.InternalAPIServer, blockAPIServer pfsclient.BlockAPIServer) {
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfsclient.RegisterAPIServer(s, apiServer)
				pfsclient.RegisterInternalAPIServer(s, internalAPIServer)
				pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: pclient.Version},
			protoserver.ServeEnv{GRPCPort: uint16(port)},
		)
		require.NoError(t, err)
	}()
	<-ready
}

func getClientAndServer(t *testing.T) (pfsclient.APIClient, []*internalAPIServer) {
	root := uniqueString("/tmp/pach_test/run")
	t.Logf("root %s", root)
	var ports []int32
	for i := 0; i < servers; i++ {
		ports = append(ports, atomic.AddInt32(&port, 1))
	}
	var addresses []string
	for _, port := range ports {
		addresses = append(addresses, fmt.Sprintf("localhost:%d", port))
	}
	sharder := shard.NewLocalSharder(addresses, shards)
	var internalAPIServers []*internalAPIServer
	for i, port := range ports {
		address := addresses[i]
		driver, err := drive.NewDriver(address)
		require.NoError(t, err)
		blockAPIServer, err := NewLocalBlockAPIServer(root)
		require.NoError(t, err)
		hasher := pfsserver.NewHasher(shards, 1)
		dialer := grpcutil.NewDialer(grpc.WithInsecure())
		apiServer := NewAPIServer(hasher, shard.NewRouter(sharder, dialer, address))
		internalAPIServer := newInternalAPIServer(hasher, shard.NewRouter(sharder, dialer, address), driver)
		internalAPIServers = append(internalAPIServers, internalAPIServer)
		runServers(t, port, apiServer, internalAPIServer, blockAPIServer)
		for i := 0; i < shards; i++ {
			require.NoError(t, internalAPIServer.AddShard(uint64(i)))
		}
	}
	clientConn, err := grpc.Dial(addresses[0], grpc.WithInsecure())
	require.NoError(t, err)
	return pfsclient.NewAPIClient(clientConn), internalAPIServers
}

func restartServer(servers []*internalAPIServer, t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, server := range servers {
		server := server
		for i := 0; i < shards; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				require.NoError(t, server.DeleteShard(uint64(i)))
				require.NoError(t, server.AddShard(uint64(i)))
			}()
		}
	}
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
