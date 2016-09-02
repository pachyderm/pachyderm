package server

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"go.pedge.io/proto/server"
	"google.golang.org/grpc"

	pclient "github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	persist "github.com/pachyderm/pachyderm/src/server/pfs/db"
)

const (
	shards  = 1
	servers = 1

	ALPHABET       = "abcdefghijklmnopqrstuvwxyz"
	RethinkAddress = "localhost:28015"
	RethinkTestDB  = "pachyderm_test"
)

var (
	port int32 = 30651
)

var testDBs []string

func TestMain(m *testing.M) {
	flag.Parse()
	code := m.Run()
	/*
		if code == 0 {
			for _, name := range testDBs {
				if err := persist.RemoveDB(RethinkAddress, name); err != nil {
					panic(err)
				}
			}
		}*/
	os.Exit(code)
}

func TestBlockRF(t *testing.T) {
	t.Parallel()
	blockClient := getBlockClient(t)
	_, err := blockClient.CreateDiff(
		context.Background(),
		&pfsclient.DiffInfo{
			Diff: pclient.NewDiff("foo", "", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfsclient.DiffInfo{
			Diff: pclient.NewDiff("foo", "c1", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfsclient.DiffInfo{
			Diff: pclient.NewDiff("foo", "c2", 0),
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

func TestInvalidRepoRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.YesError(t, client.CreateRepo("/repo"))

	require.NoError(t, client.CreateRepo("lenny"))
	require.NoError(t, client.CreateRepo("lenny123"))
	require.NoError(t, client.CreateRepo("lenny_123"))

	require.YesError(t, client.CreateRepo("lenny-123"))
	require.YesError(t, client.CreateRepo("lenny.123"))
	require.YesError(t, client.CreateRepo("lenny:"))
	require.YesError(t, client.CreateRepo("lenny,"))
	require.YesError(t, client.CreateRepo("lenny#"))
}

func TestCreateRepoNonexistantProvenanceRF(t *testing.T) {
	// This method of calling CreateRepo
	// is used within pps CreateJob()

	client := getClient(t)
	var provenance []*pfsclient.Repo
	provenance = append(provenance, pclient.NewRepo("bogusABC"))
	_, err := client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfsclient.CreateRepoRequest{
			Repo:       pclient.NewRepo("foo"),
			Provenance: provenance,
		},
	)
	require.YesError(t, err)
}

func TestSimpleRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	commitInfos, err := client.ListCommit([]string{repo}, nil, pclient.CommitTypeNone, false, pclient.CommitStatusNormal, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit2.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestBranchRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "", "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, "master", "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	branches, err := client.ListBranch(repo)
	require.NoError(t, err)
	require.Equal(t, commit1, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
	commit2, err := client.StartCommit(repo, "", "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, "master")
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, "master", "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = client.ListBranch(repo)
	require.NoError(t, err)
	require.Equal(t, commit2, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
}

func TestDisallowReadsDuringCommitRF(t *testing.T) {
	// OBSOLETE - we no longer accept file handles, and the default behavior is to
	// allow reads within a commit
	t.Skip()
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)

	// Make sure we can't get the file before the commit is finished
	var buffer bytes.Buffer
	require.YesError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "", buffer.String())

	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit2.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestInspectRepoSimpleRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	file1Content := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(file1Content))
	require.NoError(t, err)

	file2Content := "bar\n"
	_, err = client.PutFile(repo, commit.ID, "bar", strings.NewReader(file2Content))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	info, err := client.InspectRepo(repo)
	require.NoError(t, err)

	require.Equal(t, int(info.SizeBytes), len(file1Content)+len(file2Content))
}

func TestInspectRepoComplexRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
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

		_, err = client.PutFile(repo, commit.ID, fileName, strings.NewReader(fileContent))
		require.NoError(t, err)
	}

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	info, err := client.InspectRepo(repo)
	require.NoError(t, err)

	require.Equal(t, int(info.SizeBytes), totalSize)

	infos, err := client.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(infos))
	info = infos[0]

	require.Equal(t, int(info.SizeBytes), totalSize)
}

func TestListRepoRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	numRepos := 10
	var repoNames []string
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, client.CreateRepo(repo))
		repoNames = append(repoNames, repo)
	}

	repoInfos, err := client.ListRepo(nil)
	require.NoError(t, err)

	for i, repoInfo := range repoInfos {
		require.Equal(t, repoNames[i], repoInfo.Repo.Name)
	}

	require.Equal(t, len(repoInfos), numRepos)
}

func TestDeleteRepoRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	numRepos := 10
	repoNames := make(map[string]bool)
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, client.CreateRepo(repo))
		repoNames[repo] = true
	}

	reposToRemove := 5
	for i := 0; i < reposToRemove; i++ {
		// Pick one random element from repoNames
		for repoName := range repoNames {
			require.NoError(t, client.DeleteRepo(repoName, false))
			delete(repoNames, repoName)
			break
		}
	}

	repoInfos, err := client.ListRepo(nil)
	require.NoError(t, err)

	for _, repoInfo := range repoInfos {
		require.True(t, repoNames[repoInfo.Repo.Name])
	}

	require.Equal(t, len(repoInfos), numRepos-reposToRemove)
}

func TestDeleteProvenanceRepoRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	// Create two repos, one as another's provenance
	require.NoError(t, client.CreateRepo("A"))
	_, err := client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfsclient.CreateRepoRequest{
			Repo:       pclient.NewRepo("B"),
			Provenance: []*pfsclient.Repo{pclient.NewRepo("A")},
		},
	)
	require.NoError(t, err)

	// Delete the provenance repo; that should fail.
	require.YesError(t, client.DeleteRepo("A", false))

	// Delete the leaf repo, then the provenance repo; that should succeed
	require.NoError(t, client.DeleteRepo("B", false))
	require.NoError(t, client.DeleteRepo("A", false))

	repoInfos, err := client.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(repoInfos))

	// Create two repos again
	require.NoError(t, client.CreateRepo("A"))
	_, err = client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfsclient.CreateRepoRequest{
			Repo:       pclient.NewRepo("B"),
			Provenance: []*pfsclient.Repo{pclient.NewRepo("A")},
		},
	)
	require.NoError(t, err)

	// Force delete should succeed
	require.NoError(t, client.DeleteRepo("A", true))

	repoInfos, err = client.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfos))
}

func TestInspectCommitRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	started := time.Now()
	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_WRITE, commitInfo.CommitType)
	require.Equal(t, len(fileContent), int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.Nil(t, commitInfo.Finished)

	require.NoError(t, client.FinishCommit(repo, commit.ID))
	finished := time.Now()

	commitInfo, err = client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	require.Equal(t, len(fileContent), int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.True(t, finished.After(commitInfo.Finished.GoTime()))
}

func TestDeleteCommitFutureRF(t *testing.T) {
	// For when DeleteCommit gets implemented
	t.Skip()

	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	commitInfo, err := client.InspectCommit(repo, commit.ID)
	require.NotNil(t, commitInfo)

	require.NoError(t, client.DeleteCommit(repo, commit.ID))

	commitInfo, err = client.InspectCommit(repo, commit.ID)
	require.Nil(t, commitInfo)

	repoInfo, err := client.InspectRepo(repo)
	require.Equal(t, 0, repoInfo.SizeBytes)
}

func TestDeleteCommitRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	// Because DeleteCommit is not supported
	require.YesError(t, client.DeleteCommit(repo, commit.ID))
}

func TestPutFileRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo/bar", strings.NewReader("foo\n"))
	require.YesError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo/bar", strings.NewReader("foo\n"))
	require.YesError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "/bar", strings.NewReader("bar\n"))
	require.YesError(t, err) // because path starts with a slash
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	commit3, err := client.StartCommit(repo, commit2.ID, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "dir1/foo", strings.NewReader("foo\n"))
	require.NoError(t, err) // because the directory dir does not exist
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	commit4, err := client.StartCommit(repo, commit3.ID, "")
	require.NoError(t, err)
	require.NoError(t, client.MakeDirectory(repo, commit4.ID, "dir2"))
	_, err = client.PutFile(repo, commit4.ID, "dir2/bar", strings.NewReader("bar\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit4.ID, "dir1", strings.NewReader("foo\n"))
	require.YesError(t, err) // because dir1 is a directory
	require.NoError(t, client.FinishCommit(repo, commit4.ID))

	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit4.ID, "dir2/bar", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "bar\n", buffer.String())
	buffer = bytes.Buffer{}
	require.YesError(t, client.GetFile(repo, commit4.ID, "dir2", 0, 0, "", false, nil, &buffer))
}

func TestListFileTwoCommitsRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	numFiles := 5

	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	for i := 0; i < numFiles; i++ {
		_, err = client.PutFile(repo, commit1.ID, fmt.Sprintf("file%d", i), strings.NewReader("foo\n"))
		require.NoError(t, err)
	}

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)

	for i := 0; i < numFiles; i++ {
		_, err = client.PutFile(repo, commit2.ID, fmt.Sprintf("file2-%d", i), strings.NewReader("foo\n"))
		require.NoError(t, err)
	}

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	fileInfos, err := client.ListFile(repo, commit1.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))

	fileInfos, err = client.ListFile(repo, commit2.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 2*numFiles, len(fileInfos))
}

func TestPutSameFileInParallelRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader("foo\n"))
			require.NoError(t, err)
			wg.Done()
		}()
	}
	wg.Wait()
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())
}

func TestInspectFileRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent1 := "foo\n"
	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfo, err := client.InspectFile(repo, commit1.ID, "foo", "", false, nil)
	require.NoError(t, err)
	require.Equal(t, commit1, fileInfo.CommitModified)
	require.Equal(t, pfsclient.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)
	require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))

	// We inspect the file with two filter shards that have different block
	// numbers, so that only one of the filter shards should match the file
	// since the file only contains one block.
	_, err1 := client.InspectFile(repo, commit1.ID, "foo", "", false, &pfsclient.Shard{
		BlockNumber:  0,
		BlockModulus: 2,
	})
	_, err2 := client.InspectFile(repo, commit1.ID, "foo", "", false, &pfsclient.Shard{
		BlockNumber:  1,
		BlockModulus: 2,
	})
	require.True(t, (err1 == nil && err2 != nil) || (err1 != nil && err2 == nil))

	fileContent2 := "barbar\n"
	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo", strings.NewReader(fileContent2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	fileInfo, err = client.InspectFile(repo, commit2.ID, "foo", commit1.ID, false, nil)
	require.NoError(t, err)
	require.Equal(t, commit2, fileInfo.CommitModified)
	require.Equal(t, pfsclient.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)
	require.Equal(t, len(fileContent2), int(fileInfo.SizeBytes))

	fileInfo, err = client.InspectFile(repo, commit2.ID, "foo", "", false, nil)
	require.NoError(t, err)
	require.Equal(t, commit2, fileInfo.CommitModified)
	require.Equal(t, pfsclient.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)
	require.Equal(t, len(fileContent1)+len(fileContent2), int(fileInfo.SizeBytes))

	fileContent3 := "bar\n"
	commit3, err := client.StartCommit(repo, commit2.ID, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "bar", strings.NewReader(fileContent3))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	fileInfos, err := client.ListFile(repo, commit3.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, len(fileInfos), 2)
}

func TestListFileRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "dir/foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	fileContent2 := "bar\n"
	_, err = client.PutFile(repo, commit.ID, "dir/bar", strings.NewReader(fileContent2))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	fileInfos, err := client.ListFile(repo, commit.ID, "dir", "", false, nil, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
	require.True(t, fileInfos[0].File.Path == "/dir/foo" && fileInfos[1].File.Path == "/dir/bar" || fileInfos[0].File.Path == "/dir/bar" && fileInfos[1].File.Path == "/dir/foo")
	require.True(t, fileInfos[0].SizeBytes == fileInfos[1].SizeBytes && fileInfos[0].SizeBytes == uint64(len(fileContent1)))

	fileInfos, err = client.ListFile(repo, commit.ID, "dir/foo", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))
	require.True(t, fileInfos[0].File.Path == "/dir/foo")
}

func TestDeleteFileRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	// Commit 1: Add two files; delete one file within the commit
	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	fileContent2 := "bar\n"
	_, err = client.PutFile(repo, commit1.ID, "bar", strings.NewReader(fileContent2))
	require.NoError(t, err)

	require.NoError(t, client.DeleteFile(repo, commit1.ID, "foo", false, ""))

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	// foo should still be here because we can't remove a file that we are adding
	// in the same commit
	_, err = client.InspectFile(repo, commit1.ID, "foo", "", false, nil)
	require.YesError(t, err)

	// Should see one file
	fileInfos, err := client.ListFile(repo, commit1.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	// Empty commit
	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	// Should still see one files
	fileInfos, err = client.ListFile(repo, commit2.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	// Delete bar
	commit3, err := client.StartCommit(repo, commit2.ID, "")
	require.NoError(t, err)
	require.NoError(t, client.DeleteFile(repo, commit3.ID, "bar", false, ""))
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	// Should see no file
	fileInfos, err = client.ListFile(repo, commit3.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	_, err = client.InspectFile(repo, commit3.ID, "bar", "", false, nil)
	require.YesError(t, err)
}

func TestInspectDirRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = client.PutFile(repo, commit1.ID, "dir/foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	_, err = client.InspectFile(repo, commit1.ID, "dir/foo", "", false, nil)
	require.NoError(t, err)

	_, err = client.InspectFile(repo, commit1.ID, "dir", "", false, nil)
	require.NoError(t, err)

	// This is a limitation in our system: we cannot inspect .
	// . is assumed to be a directory
	// In order to be able to inspect the root directory, we have to have each
	// PutFile send a concurrent request to create an entry for ".", which is
	// a price we are not willing to pay.
	_, err = client.InspectFile(repo, commit1.ID, "", "", false, nil)
	require.YesError(t, err)
}

func TestDeleteDirRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	// Commit 1: Add two files into the same directory; delete the directory
	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit1.ID, "dir/foo", strings.NewReader("foo1"))
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit1.ID, "dir/bar", strings.NewReader("bar1"))
	require.NoError(t, err)

	require.NoError(t, client.DeleteFile(repo, commit1.ID, "dir", false, ""))

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfos, err := client.ListFile(repo, commit1.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// dir should not exist
	_, err = client.InspectFile(repo, commit1.ID, "dir", "", false, nil)
	require.YesError(t, err)

	// Commit 2: Delete the directory and add the same two files
	// The two files should reflect the new content
	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit2.ID, "dir/foo", strings.NewReader("foo2"))
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit2.ID, "dir/bar", strings.NewReader("bar2"))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	// Should see two files
	fileInfos, err = client.ListFile(repo, commit2.ID, "dir", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit2.ID, "dir/foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo2", buffer.String())

	var buffer2 bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit2.ID, "dir/bar", 0, 0, "", false, nil, &buffer2))
	require.Equal(t, "bar2", buffer2.String())

	// Commit 3: delete the directory
	commit3, err := client.StartCommit(repo, commit2.ID, "")
	require.NoError(t, err)

	require.NoError(t, client.DeleteFile(repo, commit3.ID, "dir", false, ""))

	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	// Should see zero files
	fileInfos, err = client.ListFile(repo, commit3.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// TODO: test deleting "."
}

func TestListCommitRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	commitInfos, err := client.ListCommit([]string{repo}, nil, pclient.CommitTypeNone, false, pclient.CommitStatusNormal, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	// test the block behaviour
	ch := make(chan bool)
	go func() {
		_, err = client.ListCommit([]string{repo}, []string{commit.ID}, pclient.CommitTypeNone, true, pclient.CommitStatusNormal, nil)
		close(ch)
	}()

	time.Sleep(time.Second)
	select {
	case <-ch:
		t.Fatal("ListCommit should not have returned")
	default:
	}

	commit2, err := client.StartCommit(repo, commit.ID, "")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	time.Sleep(5 * time.Second)
	select {
	case <-ch:
	default:
		t.Fatal("ListCommit should have returned")
	}

	// test that cancelled commits are not listed
	commit3, err := client.StartCommit(repo, commit2.ID, "")
	require.NoError(t, err)

	require.NoError(t, client.CancelCommit(repo, commit3.ID))
	commitInfos, err = client.ListCommit([]string{repo}, nil, pclient.CommitTypeNone, false, pclient.CommitStatusNormal, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	commitInfos, err = client.ListCommit([]string{repo}, nil, pclient.CommitTypeNone, false, pclient.CommitStatusAll, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	require.Equal(t, commit3, commitInfos[0].Commit)
}

func TestOffsetReadRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "TestOffsetRead"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "", "master")
	require.NoError(t, err)
	fileData := "foo\n"
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader(fileData))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader(fileData))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, "master", "foo", int64(len(fileData)*2)+1, 0, "", false, nil, &buffer))
	require.Equal(t, "", buffer.String())
}

func TestUnsafeOperationsRF(t *testing.T) {
	t.Parallel()
	// Unsafe is not a thing anymore
	t.Skip()
	client := getClient(t)
	repo := "TestUnsafeOperations"
	require.NoError(t, client.CreateRepo(repo))

	_, err := client.StartCommit(repo, "", "master")
	require.NoError(t, err)

	fileData := "foo"
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader(fileData))
	require.NoError(t, err)

	// A safe read should not be able to see the file
	var buffer bytes.Buffer
	require.YesError(t, client.GetFile(repo, "master", "foo", 0, 0, "", false, nil, &buffer))

	// An unsafe read should
	var buffer2 bytes.Buffer
	require.NoError(t, client.GetFileUnsafe(repo, "master", "foo", 0, 0, "", false, nil, "", &buffer2))
	require.Equal(t, "foo", buffer2.String())

	fileInfo, err := client.InspectFile(repo, "master", "foo", "", false, nil)
	require.YesError(t, err)

	fileInfo, err = client.InspectFileUnsafe(repo, "master", "foo", "", false, nil, "")
	require.NoError(t, err)
	require.Equal(t, 3, int(fileInfo.SizeBytes))

	fileInfos, err := client.ListFile(repo, "master", "", "", false, nil, true)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	fileInfos, err = client.ListFileUnsafe(repo, "master", "", "", false, nil, true, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	require.NoError(t, client.FinishCommit(repo, "master"))
}

// FinishCommit should block until the parent has been finished
func TestFinishCommitRF(t *testing.T) {
	t.Parallel()

	client := getClient(t)
	repo := "TestFinishCommit"

	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)

	ch := make(chan bool)
	go func() {
		require.NoError(t, client.FinishCommit(repo, commit2.ID))
		close(ch)
	}()

	time.Sleep(time.Second * 2)
	select {
	case <-ch:
		t.Fatalf("should block, since the parent commit has not been finished")
	default:
	}

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	time.Sleep(time.Second * 2)
	select {
	case <-ch:
	default:
		t.Fatalf("should not block, since the parent commit has been finished")
	}
}

func TestStartCommitWithNonexistentParentRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "TestStartCommitWithNonexistentParent"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "nonexistent", "")
	require.YesError(t, err)
}

// If a commit's parent has been cancelled, the commit should be cancelled too
func TestFinishCommitParentCancelledRF(t *testing.T) {
	t.Parallel()

	client := getClient(t)
	repo := "TestFinishCommitParentCancelled"

	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)

	ch := make(chan bool)
	go func() {
		require.NoError(t, client.FinishCommit(repo, commit2.ID))
		close(ch)
	}()

	require.NoError(t, client.CancelCommit(repo, commit1.ID))

	time.Sleep(time.Second * 2)
	select {
	case <-ch:
	default:
		t.Fatalf("should not block, since the parent commit has been finished")
	}

	commit2Info, err := client.InspectCommit(repo, commit2.ID)
	require.True(t, commit2Info.Cancelled)

	commit3, err := client.StartCommit(repo, commit2.ID, "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit3.ID))
	commit3Info, err := client.InspectCommit(repo, commit3.ID)
	require.True(t, commit3Info.Cancelled)
}

func TestHandleRaceRF(t *testing.T) {
	t.Parallel()
	// handle is not a thing anymore
	t.Skip()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	writer1, err := client.PutFileWriter(repo, commit.ID, "foo", pfsclient.Delimiter_LINE, "handle1")
	require.NoError(t, err)
	_, err = writer1.Write([]byte("foo"))
	require.NoError(t, err)
	writer2, err := client.PutFileWriter(repo, commit.ID, "foo", pfsclient.Delimiter_LINE, "handle2")
	require.NoError(t, err)
	_, err = writer2.Write([]byte("bar"))
	require.NoError(t, err)
	require.NoError(t, writer2.Close())
	_, err = writer1.Write([]byte("foo"))
	require.NoError(t, err)
	require.NoError(t, writer1.Close())
	require.NoError(t, client.FinishCommit(repo, commit.ID))
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.EqualOneOf(t, []interface{}{"foofoobar", "barfoofoo"}, buffer.String())
}

func Test0ModulusRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit.ID))
	zeroModulusShard := &pfsclient.Shard{}
	fileInfo, err := client.InspectFile(repo, commit.ID, "foo", "", false, zeroModulusShard)
	require.NoError(t, err)
	require.Equal(t, uint64(4), fileInfo.SizeBytes)
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit.ID, "foo", 0, 0, "", false, zeroModulusShard, &buffer))
	require.Equal(t, 4, buffer.Len())
	fileInfos, err := client.ListFile(repo, commit.ID, "", "", false, zeroModulusShard, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))
	require.Equal(t, uint64(4), fileInfos[0].SizeBytes)
}

func TestProvenanceRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfsclient.CreateRepoRequest{
		Repo:       pclient.NewRepo("B"),
		Provenance: []*pfsclient.Repo{pclient.NewRepo("A")},
	})
	require.NoError(t, err)
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfsclient.CreateRepoRequest{
		Repo:       pclient.NewRepo("C"),
		Provenance: []*pfsclient.Repo{pclient.NewRepo("B")},
	})
	require.NoError(t, err)
	repoInfo, err := client.InspectRepo("B")
	require.NoError(t, err)
	require.Equal(t, []*pfsclient.Repo{pclient.NewRepo("A")}, repoInfo.Provenance)
	repoInfo, err = client.InspectRepo("C")
	require.NoError(t, err)
	require.Equal(t, []*pfsclient.Repo{pclient.NewRepo("B"), pclient.NewRepo("A")}, repoInfo.Provenance)
	ACommit, err := client.StartCommit("A", "", "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit.ID))
	BCommit, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfsclient.StartCommitRequest{
			Repo:       pclient.NewRepo("B"),
			Provenance: []*pfsclient.Commit{ACommit},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("B", BCommit.ID))
	commitInfo, err := client.InspectCommit("B", BCommit.ID)
	require.NoError(t, err)
	fmt.Printf("provenance: %v\n", commitInfo.Provenance)
	require.Equal(t, []*pfsclient.Commit{ACommit}, commitInfo.Provenance)
	CCommit, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfsclient.StartCommitRequest{
			Repo:       pclient.NewRepo("C"),
			Provenance: []*pfsclient.Commit{BCommit},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("C", CCommit.ID))
	commitInfo, err = client.InspectCommit("C", CCommit.ID)
	require.NoError(t, err)
	require.Equal(t, []*pfsclient.Commit{BCommit, ACommit}, commitInfo.Provenance)

	// Test that we prevent provenant commits that aren't from provenant repos.
	_, err = client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfsclient.StartCommitRequest{
			Repo:       pclient.NewRepo("C"),
			Provenance: []*pfsclient.Commit{ACommit},
		},
	)
	require.YesError(t, err)

	// Test ListRepo using provenance filtering
	repoInfos, err := client.PfsAPIClient.ListRepo(
		context.Background(),
		&pfsclient.ListRepoRequest{
			Provenance: []*pfsclient.Repo{pclient.NewRepo("B")},
		},
	)
	require.NoError(t, err)
	var repos []*pfsclient.Repo
	for _, repoInfo := range repoInfos.RepoInfo {
		repos = append(repos, repoInfo.Repo)
	}
	require.Equal(t, []*pfsclient.Repo{pclient.NewRepo("C")}, repos)

	// Test ListRepo using provenance filtering
	repoInfos, err = client.PfsAPIClient.ListRepo(
		context.Background(),
		&pfsclient.ListRepoRequest{
			Provenance: []*pfsclient.Repo{pclient.NewRepo("A")},
		},
	)
	require.NoError(t, err)
	repos = nil
	for _, repoInfo := range repoInfos.RepoInfo {
		repos = append(repos, repoInfo.Repo)
	}
	require.EqualOneOf(t, []interface{}{
		[]*pfsclient.Repo{pclient.NewRepo("B"), pclient.NewRepo("C")},
		[]*pfsclient.Repo{pclient.NewRepo("C"), pclient.NewRepo("B")},
	}, repos)

	// Test ListCommit using provenance filtering
	commitInfos, err := client.PfsAPIClient.ListCommit(
		context.Background(),
		&pfsclient.ListCommitRequest{
			Repo:       []*pfsclient.Repo{pclient.NewRepo("C")},
			Provenance: []*pfsclient.Commit{ACommit},
		},
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos.CommitInfo))
	require.Equal(t, CCommit, commitInfos.CommitInfo[0].Commit)

	// Negative test ListCommit using provenance filtering
	commitInfos, err = client.PfsAPIClient.ListCommit(
		context.Background(),
		&pfsclient.ListCommitRequest{
			Repo:       []*pfsclient.Repo{pclient.NewRepo("A")},
			Provenance: []*pfsclient.Commit{BCommit},
		},
	)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos.CommitInfo))

	// Test Blocking ListCommit using provenance filtering
	ACommit2, err := client.StartCommit("A", "", "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit2.ID))
	commitInfosCh := make(chan *pfsclient.CommitInfos)
	go func() {
		commitInfos, err := client.PfsAPIClient.ListCommit(
			context.Background(),
			&pfsclient.ListCommitRequest{
				Repo:       []*pfsclient.Repo{pclient.NewRepo("B")},
				Provenance: []*pfsclient.Commit{ACommit2},
				CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
				Block:      true,
			},
		)
		require.NoError(t, err)
		commitInfosCh <- commitInfos
	}()
	BCommit2, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfsclient.StartCommitRequest{
			Repo:       pclient.NewRepo("B"),
			Provenance: []*pfsclient.Commit{ACommit2},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("B", BCommit2.ID))
	select {
	case <-time.After(5 * time.Second):
		t.Errorf("timeout waiting for commit")
	case commitInfos := <-commitInfosCh:
		require.Equal(t, 1, len(commitInfos.CommitInfo))
		require.Equal(t, BCommit2, commitInfos.CommitInfo[0].Commit)
	}
	go func() {
		commitInfos, err := client.PfsAPIClient.ListCommit(
			context.Background(),
			&pfsclient.ListCommitRequest{
				Repo:       []*pfsclient.Repo{pclient.NewRepo("C")},
				Provenance: []*pfsclient.Commit{ACommit2},
				CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
				Block:      true,
			},
		)
		require.NoError(t, err)
		commitInfosCh <- commitInfos
	}()
	CCommit2, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfsclient.StartCommitRequest{
			Repo:       pclient.NewRepo("C"),
			Provenance: []*pfsclient.Commit{BCommit2},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("C", CCommit2.ID))
	select {
	case <-time.After(5 * time.Second):
		t.Errorf("timeout waiting for commit")
	case commitInfos := <-commitInfosCh:
		require.Equal(t, 1, len(commitInfos.CommitInfo))
		require.Equal(t, CCommit2, commitInfos.CommitInfo[0].Commit)
	}
}

func TestFlushRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfsclient.CreateRepoRequest{
		Repo:       pclient.NewRepo("B"),
		Provenance: []*pfsclient.Repo{pclient.NewRepo("A")},
	})
	require.NoError(t, err)
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfsclient.CreateRepoRequest{
		Repo:       pclient.NewRepo("C"),
		Provenance: []*pfsclient.Repo{pclient.NewRepo("B")},
	})
	require.NoError(t, err)
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfsclient.CreateRepoRequest{
		Repo:       pclient.NewRepo("D"),
		Provenance: []*pfsclient.Repo{pclient.NewRepo("B")},
	})
	require.NoError(t, err)
	ACommit, err := client.StartCommit("A", "", "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit.ID))

	// do the other commits in a goro so we can block for them
	go func() {
		BCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfsclient.StartCommitRequest{
				Repo:       pclient.NewRepo("B"),
				Provenance: []*pfsclient.Commit{ACommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("B", BCommit.ID))
		CCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfsclient.StartCommitRequest{
				Repo:       pclient.NewRepo("C"),
				Provenance: []*pfsclient.Commit{BCommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("C", CCommit.ID))
		DCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfsclient.StartCommitRequest{
				Repo:       pclient.NewRepo("D"),
				Provenance: []*pfsclient.Commit{BCommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("D", DCommit.ID))
	}()

	// Flush ACommit
	commitInfos, err := client.FlushCommit([]*pfsclient.Commit{pclient.NewCommit("A", ACommit.ID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	commitInfos, err = client.FlushCommit(
		[]*pfsclient.Commit{pclient.NewCommit("A", ACommit.ID)},
		[]*pfsclient.Repo{pclient.NewRepo("C")},
	)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	// Now test what happens if one of the commits gets cancelled
	ACommit2, err := client.StartCommit("A", "", "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit2.ID))

	// do the other commits in a goro so we can block for them
	go func() {
		BCommit2, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfsclient.StartCommitRequest{
				Repo:       pclient.NewRepo("B"),
				Provenance: []*pfsclient.Commit{ACommit2},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.CancelCommit("B", BCommit2.ID))
	}()

	// Flush ACommit2
	_, err = client.FlushCommit(
		[]*pfsclient.Commit{pclient.NewCommit("A", ACommit2.ID)},
		[]*pfsclient.Repo{pclient.NewRepo("C")},
	)
	require.YesError(t, err)
}

func TestShardingInTopLevelRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	folders := 4
	filesPerFolder := 10

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	for i := 0; i < folders; i++ {
		for j := 0; j < filesPerFolder; j++ {
			_, err = client.PutFile(repo, commit.ID, fmt.Sprintf("dir%d/file%d", i, j), strings.NewReader("foo\n"))
			require.NoError(t, err)
		}
	}
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	totalFiles := 0
	for i := 0; i < folders; i++ {
		shard := &pfsclient.Shard{
			FileNumber:  uint64(i),
			FileModulus: uint64(folders),
		}
		for j := 0; j < folders; j++ {
			// You should either see all files in the folder (if the folder
			// matches the shard), or a NotFound error (if the folder does not
			// match the shard).
			fileInfos, err := client.ListFile(repo, commit.ID, fmt.Sprintf("dir%d", j), "", false, shard, false)
			require.True(t, err != nil || len(fileInfos) == filesPerFolder)
			totalFiles += len(fileInfos)
		}
	}
	require.Equal(t, folders*filesPerFolder, totalFiles)
}

func TestCreateRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	w, err := client.PutFileWriter(repo, commit.ID, "foo", pfsclient.Delimiter_LINE, "handle")
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.NoError(t, client.FinishCommit(repo, commit.ID))
	_, err = client.InspectFileUnsafe(repo, commit.ID, "foo", "", false, nil, "handle")
	require.NoError(t, err)
}

func TestGetFileInvalidCommitRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	err = client.GetFile(repo, "aninvalidcommitid", "file", 0, 0, "", false, nil, &buffer)
	require.YesError(t, err)

	require.Equal(t, fmt.Sprintf("commit %v not found in repo %v", "aninvalidcommitid", repo), err.Error())
}

func TestScrubbedErrorStringsRF(t *testing.T) {
	// REFACTOR todo (pfs-refactor): skipped because we didn't want to make an extra hop to validate repo existence for Put/Get file requests
	// post refactor, these APIs will be updated to only accept a commit, not a repo, and so we'll update the error messages then as well
	t.Skip()

	t.Parallel()
	client := getClient(t)

	err := client.CreateRepo("foo||@&#$TYX")
	require.Equal(t, "repo name (foo||@&#$TYX) invalid: only alphanumeric and underscore characters allowed", err.Error())

	_, err = client.StartCommit("zzzzz", "", "")
	require.Equal(t, "repo zzzzz not found", err.Error())

	_, err = client.ListCommit([]string{"zzzzz"}, []string{}, pclient.CommitTypeNone, false, pclient.CommitStatusNormal, nil)
	require.Equal(t, "repo zzzzz not found", err.Error())

	_, err = client.InspectRepo("bogusrepo")
	require.Equal(t, "repo bogusrepo not found", err.Error())

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	_, err = client.PutFile("sdf", commit1.ID, "file", strings.NewReader("foo\n"))
	require.Equal(t, "repo sdf not found", err.Error())

	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)

	err = client.FinishCommit("zzzzzz", commit1.ID)
	require.Equal(t, "repo zzzzzz not found", err.Error())

	err = client.FinishCommit(repo, "bogus")
	require.Equal(t, "commit bogus not found in repo test", err.Error())

	err = client.CancelCommit(repo, "bogus")
	require.Equal(t, "commit bogus not found in repo test", err.Error())

	_, err = client.InspectCommit(repo, "bogus")
	require.Equal(t, "commit bogus not found in repo test", err.Error())

	_, err = client.ListBranch("blah")
	require.Equal(t, "repo blah not found", err.Error())

	_, err = client.InspectFile(repo, commit1.ID, "file", "", false, nil)
	require.Equal(t, fmt.Sprintf("file file not found in repo %v at commit %v", repo, commit1.ID), err.Error())

	err = client.MakeDirectory(repo, "sdf", "foo")
	require.Equal(t, "commit sdf not found in repo test", err.Error())

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "file", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	err = client.GetFile(repo, "aninvalidcommitid", "file", 0, 0, "", false, nil, &buffer)
	require.YesError(t, err)

	require.Equal(t, fmt.Sprintf("commit %v not found in repo %v", "aninvalidcommitid", repo), err.Error())
}

func TestATonOfPutsRF(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long tests in short mode")
	}
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	rawMessage := `{
		"level":"debug",
		"message":{
			"thing":"foo"
		},
		"timing":[1,3,34,6,7]
	}`
	numObjs := 5000
	numGoros := 100
	var expectedOutput []byte
	var wg sync.WaitGroup
	for j := 0; j < numGoros; j++ {
		wg.Add(1)
		go func() {
			for i := 0; i < numObjs/numGoros; i++ {
				_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader(rawMessage))
				if err != nil {
					panic(err)
				}
			}
			wg.Done()
		}()
	}
	for i := 0; i < numObjs; i++ {
		expectedOutput = append(expectedOutput, []byte(rawMessage)...)
	}
	wg.Wait()

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, string(expectedOutput), buffer.String())
}

func TestPutFileWithJSONDelimiterRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	rawMessage := `{
		"level":"debug",
		"timestamp":"345",
		"message":{
			"thing":"foo"
		},
		"timing":[1,3,34,6,7]
	}`

	var expectedOutput []byte
	for !(len(expectedOutput) > 9*1024*1024) {
		expectedOutput = append(expectedOutput, []byte(rawMessage)...)
	}
	_, err = client.PutFileWithDelimiter(repo, commit1.ID, "foo", pfsclient.Delimiter_JSON, bytes.NewReader(expectedOutput))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	// Make sure all the content is there
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, len(expectedOutput), buffer.Len())
	require.Equal(t, string(expectedOutput), buffer.String())

	// Now verify that each block contains only valid JSON objects
	bigModulus := 10 // Make it big to make it less likely that I return both blocks together
	for b := 0; b < bigModulus; b++ {
		blockFilter := &pfsclient.Shard{
			BlockNumber:  uint64(b),
			BlockModulus: uint64(bigModulus),
		}

		buffer.Reset()
		if client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, blockFilter, &buffer) != nil {
			continue
		}

		// If any single block returns content of size equal to the total, we
		// got a block collision and we're not testing anything
		require.NotEqual(t, buffer.Len(), len(expectedOutput))

		var value json.RawMessage
		decoder := json.NewDecoder(&buffer)
		for {
			err = decoder.Decode(&value)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					require.NoError(t, err)
				}
			}
			require.Equal(t, rawMessage, string(value))
		}
	}
}

func TestPutFileWithNoDelimiterRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	rawMessage := "Some\ncontent\nthat\nshouldnt\nbe\nline\ndelimited.\n"

	// Write a big blob that would normally not fit in a block
	var expectedOutputA []byte
	for !(len(expectedOutputA) > 9*1024*1024) {
		expectedOutputA = append(expectedOutputA, []byte(rawMessage)...)
	}
	_, err = client.PutFileWithDelimiter(repo, commit1.ID, "foo", pfsclient.Delimiter_NONE, bytes.NewReader(expectedOutputA))
	require.NoError(t, err)

	// Write another big block
	var expectedOutputB []byte
	for !(len(expectedOutputB) > 18*1024*1024) {
		expectedOutputB = append(expectedOutputB, []byte(rawMessage)...)
	}
	_, err = client.PutFileWithDelimiter(repo, commit1.ID, "foo", pfsclient.Delimiter_NONE, bytes.NewReader(expectedOutputB))
	require.NoError(t, err)

	// Finish the commit so I can read the data
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	// Make sure all the content is there
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, len(expectedOutputA)+len(expectedOutputB), buffer.Len())
	require.Equal(t, string(append(expectedOutputA, expectedOutputB...)), buffer.String())

	// Now verify that each block only contains objects of the size we've written
	bigModulus := 10 // Make it big to make it less likely that I return both blocks together
	blockLengths := []interface{}{len(expectedOutputA), len(expectedOutputB)}
	for b := 0; b < bigModulus; b++ {
		blockFilter := &pfsclient.Shard{
			BlockNumber:  uint64(b),
			BlockModulus: uint64(bigModulus),
		}

		buffer.Reset()
		if client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, blockFilter, &buffer) != nil {
			continue
		}

		// If any single block returns content of size equal to the total, we
		// got a block collision and we're not testing anything
		require.NotEqual(t, buffer.Len(), len(expectedOutputA)+len(expectedOutputB))
		if buffer.Len() == 0 {
			continue
		}
		require.EqualOneOf(t, blockLengths, buffer.Len())
	}

}

func TestPutFileNullCharacterRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit.ID, "foo\x00bar", strings.NewReader("foobar\n"))
	// null characters error because when you `ls` files with null characters
	// they truncate things after the null character leading to strange results
	require.YesError(t, err)
}

func TestArchiveCommitRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo1 := "TestArchiveCommit1"
	repo2 := "TestArchiveCommit2"
	require.NoError(t, client.CreateRepo(repo1))
	_, err := client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfsclient.CreateRepoRequest{
			Repo:       &pfsclient.Repo{repo2},
			Provenance: []*pfsclient.Repo{{repo1}},
		})
	require.NoError(t, err)

	// create a commit on repo1
	commit1, err := client.StartCommit(repo1, "", "")
	require.NoError(t, err)
	_, err = client.PutFile(repo1, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo1, commit1.ID))
	// create a commit on repo2 with the previous commit as provenance
	commit2, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfsclient.StartCommitRequest{
			Repo:       &pfsclient.Repo{repo2},
			Provenance: []*pfsclient.Commit{commit1},
		})
	require.NoError(t, err)
	_, err = client.PutFile(repo2, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo2, commit2.ID))

	commitInfos, err := client.ListCommit([]string{repo1, repo2}, nil, pclient.CommitTypeNone, false, pclient.CommitStatusNormal, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	require.NoError(t, client.ArchiveCommit(repo1, commit1.ID))
	commitInfos, err = client.ListCommit([]string{repo1, repo2}, nil, pclient.CommitTypeNone, false, pclient.CommitStatusNormal, nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))

	// commits whose provenance has been archived should be archived on creation
	commit3, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfsclient.StartCommitRequest{
			Repo:       &pfsclient.Repo{repo2},
			Provenance: []*pfsclient.Commit{commit1},
		})
	require.NoError(t, err)
	_, err = client.PutFile(repo2, commit3.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo2, commit3.ID))
	// there should still be no commits to list
	commitInfos, err = client.ListCommit([]string{repo1, repo2}, nil, pclient.CommitTypeNone, false, pclient.CommitStatusNormal, nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))
}

func TestPutFileURLRF(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getClient(t)
	repo := "TestPutFileURL"
	require.NoError(t, c.CreateRepo(repo))
	_, err := c.StartCommit(repo, "", "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFileURL(repo, "master", "readme", "https://raw.githubusercontent.com/pachyderm/pachyderm/master/README.md"))
	require.NoError(t, c.FinishCommit(repo, "master"))
	fileInfo, err := c.InspectFile(repo, "master", "readme", "", false, nil)
	require.NoError(t, err)
	require.True(t, fileInfo.SizeBytes > 0)
}

func TestArchiveAllRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	numRepos := 10
	var repoNames []string
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, client.CreateRepo(repo))
		repoNames = append(repoNames, repo)

		commit1, err := client.StartCommit(repo, "", "")
		require.NoError(t, err)
		_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("aaa\n"))
		require.NoError(t, err)
		_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("bbb\n"))
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, commit1.ID))
	}

	commitInfos, err := client.ListCommit(repoNames, nil, pclient.CommitTypeNone, false, pclient.CommitStatusNormal, nil)
	require.NoError(t, err)
	require.Equal(t, numRepos, len(commitInfos))

	err = client.ArchiveAll()
	require.NoError(t, err)

	repoInfos, err := client.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, numRepos, len(repoInfos))

	commitInfos, err = client.ListCommit(repoNames, nil, pclient.CommitTypeNone, false, pclient.CommitStatusNormal, nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))
}

func TestBigListFileRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "TestBigListFile"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "", "master")
	require.NoError(t, err)
	var eg errgroup.Group
	for i := 0; i < 25; i++ {
		for j := 0; j < 25; j++ {
			i := i
			j := j
			eg.Go(func() error {
				_, err = client.PutFile(repo, "master", fmt.Sprintf("dir%d/file%d", i, j), strings.NewReader("foo\n"))
				return err
			})
		}
	}
	require.NoError(t, eg.Wait())
	require.NoError(t, client.FinishCommit(repo, "master"))
	for i := 0; i < 25; i++ {
		files, err := client.ListFile(repo, "master", fmt.Sprintf("dir%d", i), "", false, nil, false)
		require.NoError(t, err)
		require.Equal(t, 25, len(files))
	}
}

func TestFullFileRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "TestFullFile"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "", "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, "master", "file", 0, 0, "", true, nil, &buffer))
	require.Equal(t, "foo", buffer.String())

	_, err = client.StartCommit(repo, "", "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader("bar"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, "master")
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, "master", "file", 0, 0, "master/0", true, nil, &buffer))
	require.Equal(t, "foobar", buffer.String())
}

func TestFullFileDirRF(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "TestFullFile"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "", "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file0", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err := client.ListFile(repo, "master", "", "", true, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	_, err = client.StartCommit(repo, "", "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file1", strings.NewReader("bar"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, "master")
	require.NoError(t, err)

	fileInfos, err = client.ListFile(repo, "master", "", "master/0", true, nil, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
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
			protoserver.ServeOptions{Version: version.Version},
			protoserver.ServeEnv{GRPCPort: uint16(localPort)},
		)
		require.NoError(t, err)
	}()
	<-ready
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	return pfsclient.NewBlockAPIClient(clientConn)
}

func runServers(t *testing.T, port int32, apiServer pfsclient.APIServer,
	blockAPIServer pfsclient.BlockAPIServer) {
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfsclient.RegisterAPIServer(s, apiServer)
				pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: version.Version},
			protoserver.ServeEnv{GRPCPort: uint16(port)},
		)
		require.NoError(t, err)
	}()
	<-ready
}

func getClient(t *testing.T) pclient.APIClient {
	dbName := "pachyderm_test_" + uuid.NewWithoutDashes()[0:12]
	testDBs = append(testDBs, dbName)

	if err := persist.InitDB(RethinkAddress, dbName); err != nil {
		panic(err)
	}

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
	for i, port := range ports {
		address := addresses[i]
		driver, err := persist.NewDriver(address, RethinkAddress, dbName)
		require.NoError(t, err)
		blockAPIServer, err := NewLocalBlockAPIServer(root)
		require.NoError(t, err)
		apiServer := newAPIServer(driver)
		runServers(t, port, apiServer, blockAPIServer)
	}
	clientConn, err := grpc.Dial(addresses[0], grpc.WithInsecure())
	require.NoError(t, err)
	return pclient.APIClient{PfsAPIClient: pfsclient.NewAPIClient(clientConn)}
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
