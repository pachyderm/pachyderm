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
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	persist "github.com/pachyderm/pachyderm/src/server/pfs/db"
)

const (
	servers = 2

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

func TestBlock(t *testing.T) {
	t.Parallel()
	blockClient := getBlockClient(t)
	_, err := blockClient.CreateDiff(
		context.Background(),
		&pfs.DiffInfo{
			Diff: pclient.NewDiff("foo", "", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfs.DiffInfo{
			Diff: pclient.NewDiff("foo", "c1", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfs.DiffInfo{
			Diff: pclient.NewDiff("foo", "c2", 0),
		})
	require.NoError(t, err)
	listDiffClient, err := blockClient.ListDiff(
		context.Background(),
		&pfs.ListDiffRequest{Shard: 0},
	)
	require.NoError(t, err)
	var diffInfos []*pfs.DiffInfo
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

func TestCreateRepoNonexistantProvenance(t *testing.T) {
	// This method of calling CreateRepo
	// is used within pps CreateJob()

	client := getClient(t)
	var provenance []*pfs.Repo
	provenance = append(provenance, pclient.NewRepo("bogusABC"))
	_, err := client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo:       pclient.NewRepo("foo"),
			Provenance: provenance,
		},
	)
	require.YesError(t, err)
}

func TestSimple(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	commitInfos, err := client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo, "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := client.StartCommit(repo, commit1.ID)
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

func TestBranch(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"

	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, "master", "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	branches, err := client.ListBranch(repo, pclient.CommitStatusNormal)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0])

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, "master")
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, "master", "foo", 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = client.ListBranch(repo, pclient.CommitStatusNormal)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0])

	_, err = client.StartCommit(repo, "master2")
	require.NoError(t, err)
	err = client.FinishCommit(repo, "master2")
	require.NoError(t, err)

	branches, err = client.ListBranch(repo, pclient.CommitStatusNormal)
	require.NoError(t, err)
	require.Equal(t, 2, len(branches))
	require.Equal(t, "master", branches[0])
	require.Equal(t, "master2", branches[1])
}

func TestInspectRepoSimple(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
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

func TestInspectRepoComplex(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
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

func TestListRepo(t *testing.T) {
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

func TestDeleteRepo(t *testing.T) {
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

// Make sure that commits of deleted repos do not resurface
func TestCreateDeletedRepo(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "repo"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	commitInfos, err := client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo, "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	require.NoError(t, client.DeleteRepo(repo, false))
	require.NoError(t, client.CreateRepo(repo))

	commitInfos, err = client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo, "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))
}

func TestDeleteProvenanceRepo(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	// Create two repos, one as another's provenance
	require.NoError(t, client.CreateRepo("A"))
	_, err := client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo:       pclient.NewRepo("B"),
			Provenance: []*pfs.Repo{pclient.NewRepo("A")},
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
		&pfs.CreateRepoRequest{
			Repo:       pclient.NewRepo("B"),
			Provenance: []*pfs.Repo{pclient.NewRepo("A")},
		},
	)
	require.NoError(t, err)

	// Force delete should succeed
	require.NoError(t, client.DeleteRepo("A", true))

	repoInfos, err = client.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfos))
}

func TestInspectCommit(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	started := time.Now()
	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_WRITE, commitInfo.CommitType)
	require.Equal(t, len(fileContent), int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.Nil(t, commitInfo.Finished)

	require.NoError(t, client.FinishCommit(repo, commit.ID))
	finished := time.Now()

	commitInfo, err = client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	require.Equal(t, len(fileContent), int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.True(t, finished.After(commitInfo.Finished.GoTime()))
}

func TestDeleteCommitFuture(t *testing.T) {
	// For when DeleteCommit gets implemented
	t.Skip()

	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
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

func TestDeleteCommit(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	// Because DeleteCommit is not supported
	require.YesError(t, client.DeleteCommit(repo, commit.ID))
}

func TestPutFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
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

	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo/bar", strings.NewReader("foo\n"))
	require.YesError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "/bar", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "dir1/foo", strings.NewReader("foo\n"))
	require.NoError(t, err) // because the directory dir does not exist
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	commit4, err := client.StartCommit(repo, commit3.ID)
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

func TestPutFileLongName(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileName := `oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)`

	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, fileName, strings.NewReader("foo\n"))
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit.ID, fileName, 0, 0, "", false, nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

func TestListFileTwoCommits(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	numFiles := 5

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	for i := 0; i < numFiles; i++ {
		_, err = client.PutFile(repo, commit1.ID, fmt.Sprintf("file%d", i), strings.NewReader("foo\n"))
		require.NoError(t, err)
	}

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, commit1.ID)
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

func TestPutSameFileInParallel(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
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

func TestInspectFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent1 := "foo\n"
	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfo, err := client.InspectFile(repo, commit1.ID, "foo", "", false, nil)
	require.NoError(t, err)
	require.Equal(t, commit1, fileInfo.CommitModified)
	require.Equal(t, pfs.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)
	require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))

	// We inspect the file with two filter shards that have different block
	// numbers, so that only one of the filter shards should match the file
	// since the file only contains one block.
	_, err1 := client.InspectFile(repo, commit1.ID, "foo", "", false, &pfs.Shard{
		BlockNumber:  0,
		BlockModulus: 2,
	})
	_, err2 := client.InspectFile(repo, commit1.ID, "foo", "", false, &pfs.Shard{
		BlockNumber:  1,
		BlockModulus: 2,
	})
	require.True(t, (err1 == nil && err2 != nil) || (err1 != nil && err2 == nil))

	fileContent2 := "barbar\n"
	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo", strings.NewReader(fileContent2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	fileInfo, err = client.InspectFile(repo, commit2.ID, "foo", commit1.ID, false, nil)
	require.NoError(t, err)
	require.Equal(t, commit2, fileInfo.CommitModified)
	require.Equal(t, pfs.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)
	require.Equal(t, len(fileContent2), int(fileInfo.SizeBytes))

	fileInfo, err = client.InspectFile(repo, commit2.ID, "foo", "", false, nil)
	require.NoError(t, err)
	require.Equal(t, commit2, fileInfo.CommitModified)
	require.Equal(t, pfs.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)
	require.Equal(t, len(fileContent1)+len(fileContent2), int(fileInfo.SizeBytes))

	fileContent3 := "bar\n"
	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "bar", strings.NewReader(fileContent3))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	fileInfos, err := client.ListFile(repo, commit3.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, len(fileInfos), 2)
}

func TestListFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
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

func TestDeleteFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	// Commit 1: Add two files; delete one file within the commit
	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	fileContent2 := "bar\n"
	_, err = client.PutFile(repo, commit1.ID, "bar", strings.NewReader(fileContent2))
	require.NoError(t, err)

	require.NoError(t, client.DeleteFile(repo, commit1.ID, "foo"))

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
	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	// Should still see one files
	fileInfos, err = client.ListFile(repo, commit2.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	// Delete bar
	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)
	require.NoError(t, client.DeleteFile(repo, commit3.ID, "bar"))
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	// Should see no file
	fileInfos, err = client.ListFile(repo, commit3.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	_, err = client.InspectFile(repo, commit3.ID, "bar", "", false, nil)
	require.YesError(t, err)
}

func TestInspectDir(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
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

func TestDeleteDir(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	// Commit 1: Add two files into the same directory; delete the directory
	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit1.ID, "dir/foo", strings.NewReader("foo1"))
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit1.ID, "dir/bar", strings.NewReader("bar1"))
	require.NoError(t, err)

	require.NoError(t, client.DeleteFile(repo, commit1.ID, "dir"))

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfos, err := client.ListFile(repo, commit1.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// dir should not exist
	_, err = client.InspectFile(repo, commit1.ID, "dir", "", false, nil)
	require.YesError(t, err)

	// Commit 2: Delete the directory and add the same two files
	// The two files should reflect the new content
	commit2, err := client.StartCommit(repo, commit1.ID)
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
	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)

	require.NoError(t, client.DeleteFile(repo, commit3.ID, "dir"))

	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	// Should see zero files
	fileInfos, err = client.ListFile(repo, commit3.ID, "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// TODO: test deleting "."
}

func TestListCommit(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	commitInfos, err := client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo, "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	// test the block behaviour
	ch := make(chan bool)
	go func() {
		_, err = client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo, commit.ID)}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, true)
		close(ch)
	}()

	time.Sleep(time.Second)
	select {
	case <-ch:
		t.Fatal("ListCommit should not have returned")
	default:
	}

	commit2, err := client.StartCommit(repo, commit.ID)
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	time.Sleep(5 * time.Second)
	select {
	case <-ch:
	default:
		t.Fatal("ListCommit should have returned")
	}

	// test that cancelled commits are not listed
	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)

	require.NoError(t, client.CancelCommit(repo, commit3.ID))
	commitInfos, err = client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo, "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	commitInfos, err = client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo, "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusAll, false)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))
	require.Equal(t, commit3, commitInfos[2].Commit)
}

func TestOffsetRead(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "TestOffsetRead"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "master")
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

// FinishCommit should block until the parent has been finished
func TestFinishCommit(t *testing.T) {
	t.Parallel()

	client := getClient(t)
	repo := "TestFinishCommit"

	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	commit2, err := client.StartCommit(repo, commit1.ID)
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

func TestStartCommitWithNonexistentParent(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "TestStartCommitWithNonexistentParent"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "nonexistent/3")
	require.YesError(t, err)
}

// If a commit's parent has been cancelled, the commit should be cancelled too
func TestFinishCommitParentCancelled(t *testing.T) {
	t.Parallel()

	client := getClient(t)
	repo := "TestFinishCommitParentCancelled"

	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	commit2, err := client.StartCommit(repo, commit1.ID)
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

	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit3.ID))
	commit3Info, err := client.InspectCommit(repo, commit3.ID)
	require.True(t, commit3Info.Cancelled)
}

func TestHandleRace(t *testing.T) {
	t.Parallel()
	// handle is not a thing anymore
	t.Skip()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	writer1, err := client.PutFileWriter(repo, commit.ID, "foo", pfs.Delimiter_LINE)
	require.NoError(t, err)
	_, err = writer1.Write([]byte("foo"))
	require.NoError(t, err)
	writer2, err := client.PutFileWriter(repo, commit.ID, "foo", pfs.Delimiter_LINE)
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

func Test0Modulus(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit.ID))
	zeroModulusShard := &pfs.Shard{}
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

func TestProvenance(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("B"),
		Provenance: []*pfs.Repo{pclient.NewRepo("A")},
	})
	require.NoError(t, err)
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("C"),
		Provenance: []*pfs.Repo{pclient.NewRepo("B")},
	})
	require.NoError(t, err)
	repoInfo, err := client.InspectRepo("B")
	require.NoError(t, err)
	require.Equal(t, []*pfs.Repo{pclient.NewRepo("A")}, repoInfo.Provenance)
	repoInfo, err = client.InspectRepo("C")
	require.NoError(t, err)
	require.Equal(t, []*pfs.Repo{pclient.NewRepo("A"), pclient.NewRepo("B")}, repoInfo.Provenance)
	ACommit, err := client.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit.ID))
	BCommit, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("B", "master"),
			Provenance: []*pfs.Commit{ACommit},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("B", BCommit.ID))
	commitInfo, err := client.InspectCommit("B", BCommit.ID)
	require.NoError(t, err)
	require.Equal(t, []*pfs.Commit{ACommit}, commitInfo.Provenance)
	CCommit, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("C", "master"),
			Provenance: []*pfs.Commit{BCommit},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("C", CCommit.ID))
	commitInfo, err = client.InspectCommit("C", CCommit.ID)
	require.NoError(t, err)
	require.Equal(t, []*pfs.Commit{BCommit, ACommit}, commitInfo.Provenance)

	// Test that we prevent provenant commits that aren't from provenant repos.
	_, err = client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("C", "master"),
			Provenance: []*pfs.Commit{ACommit},
		},
	)
	require.YesError(t, err)

	// Test ListRepo using provenance filtering
	repoInfos, err := client.PfsAPIClient.ListRepo(
		context.Background(),
		&pfs.ListRepoRequest{
			Provenance: []*pfs.Repo{pclient.NewRepo("B")},
		},
	)
	require.NoError(t, err)
	var repos []*pfs.Repo
	for _, repoInfo := range repoInfos.RepoInfo {
		repos = append(repos, repoInfo.Repo)
	}
	require.Equal(t, []*pfs.Repo{pclient.NewRepo("C")}, repos)

	// Test ListRepo using provenance filtering
	repoInfos, err = client.PfsAPIClient.ListRepo(
		context.Background(),
		&pfs.ListRepoRequest{
			Provenance: []*pfs.Repo{pclient.NewRepo("A")},
		},
	)
	require.NoError(t, err)
	repos = nil
	for _, repoInfo := range repoInfos.RepoInfo {
		repos = append(repos, repoInfo.Repo)
	}
	require.EqualOneOf(t, []interface{}{
		[]*pfs.Repo{pclient.NewRepo("B"), pclient.NewRepo("C")},
		[]*pfs.Repo{pclient.NewRepo("C"), pclient.NewRepo("B")},
	}, repos)

	// Test ListCommit using provenance filtering
	commitInfos, err := client.PfsAPIClient.ListCommit(
		context.Background(),
		&pfs.ListCommitRequest{
			FromCommits: []*pfs.Commit{pclient.NewCommit("C", "")},
			Provenance:  []*pfs.Commit{ACommit},
		},
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos.CommitInfo))
	require.Equal(t, CCommit, commitInfos.CommitInfo[0].Commit)

	// Negative test ListCommit using provenance filtering
	commitInfos, err = client.PfsAPIClient.ListCommit(
		context.Background(),
		&pfs.ListCommitRequest{
			FromCommits: []*pfs.Commit{pclient.NewCommit("A", "")},
			Provenance:  []*pfs.Commit{BCommit},
		},
	)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos.CommitInfo))

	// Test Blocking ListCommit using provenance filtering
	ACommit2, err := client.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit2.ID))
	commitInfosCh := make(chan *pfs.CommitInfos)
	go func() {
		commitInfos, err := client.PfsAPIClient.ListCommit(
			context.Background(),
			&pfs.ListCommitRequest{
				FromCommits: []*pfs.Commit{pclient.NewCommit("B", "")},
				Provenance:  []*pfs.Commit{ACommit2},
				CommitType:  pfs.CommitType_COMMIT_TYPE_READ,
				Block:       true,
			},
		)
		require.NoError(t, err)
		commitInfosCh <- commitInfos
	}()
	BCommit2, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("B", "master"),
			Provenance: []*pfs.Commit{ACommit2},
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
			&pfs.ListCommitRequest{
				FromCommits: []*pfs.Commit{pclient.NewCommit("C", "")},
				Provenance:  []*pfs.Commit{ACommit2},
				CommitType:  pfs.CommitType_COMMIT_TYPE_READ,
				Block:       true,
			},
		)
		require.NoError(t, err)
		commitInfosCh <- commitInfos
	}()
	CCommit2, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("C", "master"),
			Provenance: []*pfs.Commit{BCommit2},
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

func TestFlush(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("B"),
		Provenance: []*pfs.Repo{pclient.NewRepo("A")},
	})
	require.NoError(t, err)
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("C"),
		Provenance: []*pfs.Repo{pclient.NewRepo("B")},
	})
	require.NoError(t, err)
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("D"),
		Provenance: []*pfs.Repo{pclient.NewRepo("B")},
	})
	require.NoError(t, err)
	ACommit, err := client.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit.ID))

	// do the other commits in a goro so we can block for them
	go func() {
		BCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfs.StartCommitRequest{
				Parent:     pclient.NewCommit("B", "master"),
				Provenance: []*pfs.Commit{ACommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("B", BCommit.ID))
		CCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfs.StartCommitRequest{
				Parent:     pclient.NewCommit("C", "master"),
				Provenance: []*pfs.Commit{BCommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("C", CCommit.ID))
		DCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfs.StartCommitRequest{
				Parent:     pclient.NewCommit("D", "master"),
				Provenance: []*pfs.Commit{BCommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("D", DCommit.ID))
	}()

	// Flush ACommit
	commitInfos, err := client.FlushCommit([]*pfs.Commit{pclient.NewCommit("A", ACommit.ID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfos))
	commitInfos, err = client.FlushCommit(
		[]*pfs.Commit{pclient.NewCommit("A", ACommit.ID)},
		[]*pfs.Repo{pclient.NewRepo("C")},
	)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	// Now test what happens if one of the commits gets cancelled
	ACommit2, err := client.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit2.ID))

	// do the other commits in a goro so we can block for them
	go func() {
		BCommit2, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfs.StartCommitRequest{
				Parent:     pclient.NewCommit("B", "master"),
				Provenance: []*pfs.Commit{ACommit2},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.CancelCommit("B", BCommit2.ID))
	}()

	// Flush ACommit2
	_, err = client.FlushCommit(
		[]*pfs.Commit{pclient.NewCommit("A", ACommit2.ID)},
		[]*pfs.Repo{pclient.NewRepo("C")},
	)
	require.YesError(t, err)
}

func TestFlushCommitReturnsFromCommit(t *testing.T) {
	t.Parallel()
	c := getClient(t)
	repo := "TestFlushCommitReturnsFromCommit"
	require.NoError(t, c.CreateRepo(repo))
	_, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, "master"))
	commitInfos, err := c.FlushCommit([]*pfs.Commit{pclient.NewCommit(repo, "master")}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
}

func TestFlushOpenCommit(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("B"),
		Provenance: []*pfs.Repo{pclient.NewRepo("A")},
	})
	ACommit, err := client.StartCommit("A", "master")
	require.NoError(t, err)

	// do the other commits in a goro so we can block for them
	go func() {
		time.Sleep(5 * time.Second)
		require.NoError(t, client.FinishCommit("A", ACommit.ID))
		BCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfs.StartCommitRequest{
				Parent:     pclient.NewCommit("B", "master"),
				Provenance: []*pfs.Commit{ACommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("B", BCommit.ID))
	}()

	// Flush ACommit
	commitInfos, err := client.FlushCommit([]*pfs.Commit{pclient.NewCommit("A", ACommit.ID)}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
}

func TestEmptyFlush(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	_, err := client.FlushCommit(nil, nil)
	require.NoError(t, err)
}

func TestShardingInTopLevel(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	folders := 4
	filesPerFolder := 10

	commit, err := client.StartCommit(repo, "master")
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
		shard := &pfs.Shard{
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

func TestCreate(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	w, err := client.PutFileWriter(repo, commit.ID, "foo", pfs.Delimiter_LINE)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.NoError(t, client.FinishCommit(repo, commit.ID))
	_, err = client.InspectFile(repo, commit.ID, "foo", "", false, nil)
	require.NoError(t, err)
}

func TestGetFileInvalidCommit(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "master")
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

func TestScrubbedErrorStrings(t *testing.T) {
	// REFACTOR todo (pfs-refactor): skipped because we didn't want to make an extra hop to validate repo existence for Put/Get file requests
	// post refactor, these APIs will be updated to only accept a commit, not a repo, and so we'll update the error messages then as well
	t.Skip()

	t.Parallel()
	client := getClient(t)

	err := client.CreateRepo("foo||@&#$TYX")
	require.Equal(t, "repo name (foo||@&#$TYX) invalid: only alphanumeric and underscore characters allowed", err.Error())

	_, err = client.StartCommit("zzzzz", "master")
	require.Equal(t, "repo zzzzz not found", err.Error())

	_, err = client.ListCommit([]*pfs.Commit{pclient.NewCommit("zzzzz", "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.Equal(t, "repo zzzzz not found", err.Error())

	_, err = client.InspectRepo("bogusrepo")
	require.Equal(t, "repo bogusrepo not found", err.Error())

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "master")
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

	_, err = client.ListBranch("blah", pclient.CommitStatusNormal)
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

func TestATonOfPuts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long tests in short mode")
	}
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
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

func TestPutFileWithJSONDelimiter(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
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
	_, err = client.PutFileWithDelimiter(repo, commit1.ID, "foo", pfs.Delimiter_JSON, bytes.NewReader(expectedOutput))
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
		blockFilter := &pfs.Shard{
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

func TestPutFileWithNoDelimiter(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	rawMessage := "Some\ncontent\nthat\nshouldnt\nbe\nline\ndelimited.\n"

	// Write a big blob that would normally not fit in a block
	var expectedOutputA []byte
	for !(len(expectedOutputA) > 9*1024*1024) {
		expectedOutputA = append(expectedOutputA, []byte(rawMessage)...)
	}
	_, err = client.PutFileWithDelimiter(repo, commit1.ID, "foo", pfs.Delimiter_NONE, bytes.NewReader(expectedOutputA))
	require.NoError(t, err)

	// Write another big block
	var expectedOutputB []byte
	for !(len(expectedOutputB) > 18*1024*1024) {
		expectedOutputB = append(expectedOutputB, []byte(rawMessage)...)
	}
	_, err = client.PutFileWithDelimiter(repo, commit1.ID, "foo", pfs.Delimiter_NONE, bytes.NewReader(expectedOutputB))
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
		blockFilter := &pfs.Shard{
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

func TestPutFileNullCharacter(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit.ID, "foo\x00bar", strings.NewReader("foobar\n"))
	// null characters error because when you `ls` files with null characters
	// they truncate things after the null character leading to strange results
	require.YesError(t, err)
}

func TestArchiveCommit(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo1 := "TestArchiveCommit1"
	repo2 := "TestArchiveCommit2"
	require.NoError(t, client.CreateRepo(repo1))
	_, err := client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo: &pfs.Repo{
				Name: repo2,
			},
			Provenance: []*pfs.Repo{{
				Name: repo1,
			}},
		})
	require.NoError(t, err)

	// create a commit on repo1
	commit1, err := client.StartCommit(repo1, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo1, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo1, commit1.ID))
	// create a commit on repo2 with the previous commit as provenance
	commit2, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit(repo2, "master"),
			Provenance: []*pfs.Commit{commit1},
		})
	require.NoError(t, err)
	_, err = client.PutFile(repo2, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo2, commit2.ID))

	commitInfos, err := client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo1, ""), pclient.NewCommit(repo2, "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))
	require.NoError(t, client.ArchiveCommit(repo1, commit1.ID))
	commitInfos, err = client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo1, ""), pclient.NewCommit(repo2, "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))

	// commits whose provenance has been archived should be archived on creation
	commit3, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit(repo2, "master"),
			Provenance: []*pfs.Commit{commit1},
		})
	require.NoError(t, err)
	_, err = client.PutFile(repo2, commit3.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo2, commit3.ID))
	// there should still be no commits to list
	commitInfos, err = client.ListCommit([]*pfs.Commit{pclient.NewCommit(repo1, ""), pclient.NewCommit(repo2, "")}, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))
}

func TestPutFileURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getClient(t)
	repo := "TestPutFileURL"
	require.NoError(t, c.CreateRepo(repo))
	_, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFileURL(repo, "master", "readme", "https://raw.githubusercontent.com/pachyderm/pachyderm/master/README.md"))
	require.NoError(t, c.FinishCommit(repo, "master"))
	fileInfo, err := c.InspectFile(repo, "master", "readme", "", false, nil)
	require.NoError(t, err)
	require.True(t, fileInfo.SizeBytes > 0)
}

func TestArchiveAll(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	numRepos := 10
	var fromCommits []*pfs.Commit
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, client.CreateRepo(repo))
		fromCommits = append(fromCommits, pclient.NewCommit(repo, ""))

		commit1, err := client.StartCommit(repo, "master")
		require.NoError(t, err)
		_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("aaa\n"))
		require.NoError(t, err)
		_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("bbb\n"))
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, commit1.ID))
	}

	commitInfos, err := client.ListCommit(fromCommits, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, numRepos, len(commitInfos))

	err = client.ArchiveAll()
	require.NoError(t, err)

	repoInfos, err := client.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, numRepos, len(repoInfos))

	commitInfos, err = client.ListCommit(fromCommits, nil, pclient.CommitTypeNone, pclient.CommitStatusNormal, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))
}

func TestBigListFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "TestBigListFile"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "master")
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

func TestFullFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "TestFullFile"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, "master", "file", 0, 0, "", true, nil, &buffer))
	require.Equal(t, "foo", buffer.String())

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader("bar"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, "master")
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, "master", "file", 0, 0, "master/0", true, nil, &buffer))
	require.Equal(t, "foobar", buffer.String())
}

func TestFullFileDir(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "TestFullFile"
	require.NoError(t, client.CreateRepo(repo))
	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file0", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err := client.ListFile(repo, "master", "", "", true, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file1", strings.NewReader("bar"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, "master")
	require.NoError(t, err)

	fileInfos, err = client.ListFile(repo, "master", "", "master/0", true, nil, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
}

func TestBranchSimple(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "branchA")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))
	branches, err := client.ListBranch(repo, pclient.CommitStatusNormal)
	require.NoError(t, err)

	require.Equal(t, 1, len(branches))
	require.Equal(t, "branchA", branches[0])

}

func TestListBranch(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "branchA")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	// Don't specify a branch because we already should have it from parent
	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	// Specify branch, because branching off of commit1
	commit3, err := client.ForkCommit(repo, commit1.ID, "branchB")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit3.ID))
	branches, err := client.ListBranch(repo, pclient.CommitStatusNormal)
	require.NoError(t, err)

	require.Equal(t, 2, len(branches))
	branchNames := []interface{}{
		branches[0],
		branches[1],
	}

	require.EqualOneOf(t, branchNames, "branchA")
	require.EqualOneOf(t, branchNames, "branchB")
}

func TestListCommitBasic(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	require.NoError(t, client.CreateRepo("test"))
	numCommits := 10
	var commitIDs []string
	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit("test", "master")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("test", commit.ID))
		commitIDs = append(commitIDs, commit.ID)
	}

	commitInfos, err := client.ListCommit(
		[]*pfs.Commit{pclient.NewCommit("test", "")},
		nil,
		pclient.CommitTypeNone,
		pfs.CommitStatus_NORMAL,
		false,
	)
	require.NoError(t, err)

	require.Equal(t, len(commitInfos), numCommits)
	for i, commitInfo := range commitInfos {
		require.Equal(t, commitIDs[i], commitInfo.Commit.ID)
	}
}

func TestListCommitFromCommit(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	require.NoError(t, client.CreateRepo("test"))
	numCommits := 10
	var commitIDs []string
	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit("test", "master")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("test", commit.ID))
		commitIDs = append(commitIDs, commit.ID)
	}

	commitInfos, err := client.ListCommit(
		[]*pfs.Commit{pclient.NewCommit("test", "master/4")},
		nil,
		pclient.CommitTypeNone,
		pfs.CommitStatus_NORMAL,
		false,
	)
	require.NoError(t, err)

	require.Equal(t, len(commitInfos), 5)
	for i, commitInfo := range commitInfos {
		require.Equal(t, commitIDs[i+5], commitInfo.Commit.ID)
	}
}

func TestListCommitCorrectDescendents(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	require.NoError(t, client.CreateRepo("test"))
	numCommits := 10

	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit("test", "a")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("test", commit.ID))
	}

	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit("test", "b")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("test", commit.ID))
	}

	commitInfos, err := client.ListCommit(
		[]*pfs.Commit{pclient.NewCommit("test", "a/4")},
		nil,
		pclient.CommitTypeNone,
		pfs.CommitStatus_NORMAL,
		false,
	)
	require.NoError(t, err)

	require.Equal(t, len(commitInfos), 5)
	for _, commitInfo := range commitInfos {
		require.Equal(t, commitInfo.Branch, "a")
	}
}

func TestListCommitAll(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	numCommits := 10
	for i := 0; i < numCommits; i++ {
		repo := fmt.Sprintf("test%d", i)
		require.NoError(t, client.CreateRepo(repo))
		commit, err := client.StartCommit(repo, "master")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, commit.ID))
	}

	commitInfos, err := client.ListCommit(
		nil,
		nil,
		pclient.CommitTypeNone,
		pfs.CommitStatus_NORMAL,
		false,
	)
	require.NoError(t, err)
	require.Equal(t, len(commitInfos), numCommits)
}

func TestStartAndFinishCommit(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))
}

func TestInspectCommitBasic(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	started := time.Now()
	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_WRITE, commitInfo.CommitType)
	require.Equal(t, 0, int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.Nil(t, commitInfo.Finished)

	require.NoError(t, client.FinishCommit(repo, commit.ID))
	finished := time.Now()

	commitInfo, err = client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit.ID, commitInfo.Commit.ID)
	require.Equal(t, pfs.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	require.Equal(t, 0, int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.True(t, finished.After(commitInfo.Finished.GoTime()))
}

func TestStartCommitFromParentID(t *testing.T) {
	t.Parallel()

	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	branches, err := client.ListBranch(repo, pclient.CommitStatusNormal)
	require.NoError(t, err)

	require.Equal(t, 1, len(branches))

	// Should create commit off of parent on the same branch
	commit1, err := client.StartCommit(repo, commit.ID)
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	existingBranch := branches[0]
	branches, err = client.ListBranch(repo, pclient.CommitStatusNormal)
	require.NoError(t, err)

	require.Equal(t, 1, len(branches))

	// Should create commit off of parent on a new branch by name
	commit2, err := client.ForkCommit(repo, commit.ID, "foo")
	require.NoError(t, err)

	branches2, err := client.ListBranch(repo, pclient.CommitStatusNormal)
	require.NoError(t, err)

	uniqueBranches := make(map[string]bool)

	for _, thisBranch := range branches2 {
		uniqueBranches[thisBranch] = true
	}

	require.Equal(t, 2, len(uniqueBranches))
	delete(uniqueBranches, existingBranch)
	require.Equal(t, 1, len(uniqueBranches))

	require.NoError(t, client.FinishCommit(repo, commit2.ID))
}

func TestInspectRepoMostBasic(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	repoInfo, err := client.InspectRepo(repo)
	require.NoError(t, err)

	require.Equal(t, int(repoInfo.SizeBytes), 0)
}

func TestStartCommitLatestOnBranch(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	commit3, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	commitInfo, err := client.InspectCommit(repo, "master")
	require.Equal(t, commit3.ID, commitInfo.Commit.ID)
}

func TestListBranchRedundant(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "branchA")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	// Can't create branch if it exists
	_, err = client.ForkCommit(repo, commit1.ID, "branchA")
	require.YesError(t, err)

	branches, err := client.ListBranch(repo, pclient.CommitStatusNormal)
	require.NoError(t, err)

	require.Equal(t, 1, len(branches))
	require.Equal(t, "branchA", branches[0])
}

func TestStartCommitFromBranch(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(repo, "master/0")
	require.NoError(t, err)
	require.Equal(t, "master", commitInfo.Branch)
	require.Equal(t, "test", commitInfo.Commit.Repo.Name)

	require.NoError(t, client.FinishCommit(repo, "master/0"))
}

func TestStartCommitNewBranch(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/0"))

	_, err = client.ForkCommit(repo, commit1.ID, "foo")
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(repo, "foo/0")
	require.NoError(t, err)
	require.Equal(t, "foo", commitInfo.Branch)
	require.Equal(t, "test", commitInfo.Commit.Repo.Name)
}

func TestPutFile2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/0", "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/0"))

	expected := "foo\nbar\nbuzz\n"
	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit1.ID, "file", 0, 0, "", false, nil, buffer))
	require.Equal(t, expected, buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "master/0", "file", 0, 0, "", false, nil, buffer))
	require.Equal(t, expected, buffer.String())

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/1", "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/1"))

	expected = "foo\nbar\nbuzz\nfoo\nbar\nbuzz\n"
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit2.ID, "file", 0, 0, "", false, nil, buffer))
	require.Equal(t, expected, buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "master/1", "file", 0, 0, "", false, nil, buffer))
	require.Equal(t, expected, buffer.String())

	_, err = client.ForkCommit(repo, "master/1", "foo")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "foo/0", "file", strings.NewReader("foo\nbar\nbuzz\n"))
	require.NoError(t, client.FinishCommit(repo, "foo/0"))

	expected = "foo\nbar\nbuzz\nfoo\nbar\nbuzz\nfoo\nbar\nbuzz\n"
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "foo/0", "file", 0, 0, "", false, nil, buffer))
	require.Equal(t, expected, buffer.String())
}

func TestDeleteFile2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/0", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/0"))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master/1", "file")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/1", "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/1"))

	expected := "bar\n"
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, "master/1", "file", 0, 0, "master/0", false, nil, &buffer))
	require.Equal(t, expected, buffer.String())

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/2", "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master/2", "file")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/2", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/2"))

	expected = "foo\n"
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "master/2", "file", 0, 0, "master/0", false, nil, &buffer))
	require.Equal(t, expected, buffer.String())
}

func TestInspectFile2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent1 := "foo\n"
	fileContent2 := "buzz\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/0", "file", strings.NewReader(fileContent1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/0"))

	fileInfo, err := client.InspectFile(repo, "master/0", "file", "", false, nil)
	require.NoError(t, err)
	require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))
	require.Equal(t, "/file", fileInfo.File.Path)
	require.Equal(t, pfs.FileType_FILE_TYPE_REGULAR, fileInfo.FileType)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/1", "file", strings.NewReader(fileContent1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/1"))

	fileInfo, err = client.InspectFile(repo, "master/1", "file", "", false, nil)
	require.NoError(t, err)
	require.Equal(t, len(fileContent1)*2, int(fileInfo.SizeBytes))
	require.Equal(t, "/file", fileInfo.File.Path)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master/2", "file")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/2", "file", strings.NewReader(fileContent2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/2"))

	fileInfo, err = client.InspectFile(repo, "master/2", "file", "", false, nil)
	require.NoError(t, err)
	require.Equal(t, len(fileContent2), int(fileInfo.SizeBytes))
}

func TestInspectDirectory(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/0", "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/0", "dir/2", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/0"))

	fileInfo, err := client.InspectFile(repo, "master/0", "dir", "", false, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfo.Children))
	require.Equal(t, "/dir", fileInfo.File.Path)
	require.Equal(t, pfs.FileType_FILE_TYPE_DIR, fileInfo.FileType)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/1", "dir/3", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/1"))

	fileInfo, err = client.InspectFile(repo, "master/1", "dir", "", false, nil)
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfo.Children))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master/2", "dir/2")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/2"))

	fileInfo, err = client.InspectFile(repo, "master/2", "dir", "", false, nil)
	require.NoError(t, err)
	fmt.Printf("children: %+v", fileInfo.Children)
	require.Equal(t, 2, len(fileInfo.Children))
}

func TestListFile2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/0", "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/0", "dir/2", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/0"))

	fileInfos, err := client.ListFile(repo, "master/0", "dir", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/1", "dir/3", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/1"))

	fileInfos, err = client.ListFile(repo, "master/1", "dir", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfos))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master/2", "dir/2")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/2"))

	fileInfos, err = client.ListFile(repo, "master/2", "dir", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
}

func TestListFileRecurse(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/0", "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/0", "dir/2", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/0"))

	fileInfos, err := client.ListFile(repo, "master/0", "dir", "", false, nil, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/1", "dir/3/foo", strings.NewReader(fileContent))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master/1", "dir/3/bar", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/1"))

	fileInfos, err = client.ListFile(repo, "master/1", "dir", "", false, nil, true)
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfos))
	require.Equal(t, int(fileInfos[2].SizeBytes), len(fileContent)*2)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master/2", "dir/3/bar")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/2"))

	fileInfos, err = client.ListFile(repo, "master/2", "dir", "", false, nil, true)
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfos))
	require.Equal(t, int(fileInfos[2].SizeBytes), len(fileContent))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err = client.ListFile(repo, "master", "/", "", false, nil, true)
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
}

func TestPutFileTypeConflict(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir", strings.NewReader(fileContent))
	require.YesError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))
}

func TestRootDirectory(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err := client.ListFile(repo, "master", "", "", false, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))
}

func TestSquashMergeSameFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commitRoot, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	contentA1 := "foo1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentA2 := "foo2\n"
	_, err = client.StartCommit(repo, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentB1 := "bar1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file", strings.NewReader(contentB1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	contentB2 := "bar2\n"
	_, err = client.StartCommit(repo, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file", strings.NewReader(contentB2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	commit, err := squash(client, repo, []string{"A", "B"}, "master")
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit.ID, "file", 0, 0, "", false, nil, buffer))
	// The ordering of commits within the same branch should be preserved
	require.EqualOneOf(t, []interface{}{contentA1 + contentA2 + contentB1 + contentB2, contentB1 + contentB2 + contentA1 + contentA2}, buffer.String())
}

func TestReplayMergeSameFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commitRoot, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	contentA1 := "foo1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentA2 := "foo2\n"
	_, err = client.StartCommit(repo, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentB1 := "bar1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file", strings.NewReader(contentB1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	contentB2 := "bar2\n"
	_, err = client.StartCommit(repo, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file", strings.NewReader(contentB2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	mergedCommits, err := client.ReplayCommit(repo, []string{"A", "B"}, "master")
	require.NoError(t, err)
	require.Equal(t, 4, len(mergedCommits))

	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, mergedCommits[3].ID, "file", 0, 0, "", false, nil, buffer))
	// The ordering of commits within the same branch should be preserved
	require.EqualOneOf(t, []interface{}{contentA1 + contentA2 + contentB1 + contentB2, contentB1 + contentB2 + contentA1 + contentA2}, buffer.String())
}

func TestSquashMergeDiffOrdering(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commitRoot, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	contentA1 := "foo1\n"
	contentA2 := "foo2\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA1))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentB1 := "bar1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file", strings.NewReader(contentB1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	commit, err := squash(client, repo, []string{"A", "B"}, "master")
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit.ID, "file", 0, 0, "", false, nil, buffer))
	// The ordering of commits within the same branch should be preserved
	require.EqualOneOf(t, []interface{}{contentA1 + contentA2 + contentB1, contentB1 + contentA1 + contentA2}, buffer.String())
}

func TestReplayMergeDiffOrdering(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commitRoot, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	contentA1 := "foo1\n"
	contentA2 := "foo2\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA1))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentB1 := "bar1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file", strings.NewReader(contentB1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	mergedCommits, err := client.ReplayCommit(repo, []string{"A", "B"}, "master")
	require.NoError(t, err)
	require.Equal(t, 2, len(mergedCommits))

	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, mergedCommits[1].ID, "file", 0, 0, "", false, nil, buffer))
	// The ordering of commits within the same branch should be preserved
	require.EqualOneOf(t, []interface{}{contentA1 + contentA2 + contentB1, contentB1 + contentA1 + contentA2}, buffer.String())
}

func TestReplayMergeBranches(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commitRoot, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	contentA1 := "foo1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentA2 := "foo2\n"
	_, err = client.StartCommit(repo, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentA3 := "foo3\n"
	_, err = client.StartCommit(repo, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file", strings.NewReader(contentA3))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentB1 := "bar1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file", strings.NewReader(contentB1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	contentB2 := "bar2\n"
	_, err = client.StartCommit(repo, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file", strings.NewReader(contentB2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	mergedCommits, err := client.ReplayCommit(repo, []string{"A"}, "B")
	require.NoError(t, err)
	require.Equal(t, 3, len(mergedCommits))

	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, mergedCommits[2].ID, "file", 0, 0, "", false, nil, buffer))
	// The ordering of commits within the same branch should be preserved
	require.Equal(t, contentB1+contentB2+contentA1+contentA2+contentA3, buffer.String())
}

func TestReplayMergeMultipleFiles(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commitRoot, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	contentA1 := "foo1\n"
	contentA2 := "foo2\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file1", strings.NewReader(contentA1))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file2", strings.NewReader(contentA2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentB1 := "bar1\n"
	contentB2 := "bar2\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file1", strings.NewReader(contentB1))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file2", strings.NewReader(contentB2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	mergedCommits, err := client.ReplayCommit(repo, []string{"A", "B"}, "master")
	require.NoError(t, err)
	require.Equal(t, 2, len(mergedCommits))

	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, mergedCommits[1].ID, "file1", 0, 0, "", false, nil, buffer))
	require.EqualOneOf(t, []interface{}{contentB1 + contentA1, contentA1 + contentB1}, buffer.String())

	buffer.Reset()
	require.NoError(t, client.GetFile(repo, mergedCommits[1].ID, "file2", 0, 0, "", false, nil, buffer))
	require.EqualOneOf(t, []interface{}{contentB2 + contentA2, contentA2 + contentB2}, buffer.String())
}

func TestSquashMergeMultipleFiles(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commitRoot, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	contentA1 := "foo1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file1", strings.NewReader(contentA1))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file2", strings.NewReader(contentA1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentA2 := "foo2\n"
	_, err = client.StartCommit(repo, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file1", strings.NewReader(contentA2))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "file2", strings.NewReader(contentA2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	contentB1 := "bar1\n"
	_, err = client.ForkCommit(repo, commitRoot.ID, "B")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file1", strings.NewReader(contentB1))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "B", "file2", strings.NewReader(contentB1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "B"))

	commit, err := squash(client, repo, []string{"A", "B"}, "master")
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit.ID, "file1", 0, 0, "", false, nil, buffer))
	require.EqualOneOf(t, []interface{}{contentA1 + contentA2 + contentB1, contentB1 + contentA1 + contentA2}, buffer.String())
}

func TestLeadingSlashesBreakThis(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	contentA1 := "foo1\n"
	contentA2 := "foo2\n"
	commit1, err := client.StartCommit(repo, "A")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "dir/file1", strings.NewReader(contentA1))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "A", "dir/file2", strings.NewReader(contentA2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "A"))

	shard1 := &pfs.Shard{
		FileNumber:  0,
		FileModulus: 2,
	}
	fileInfos1, err := client.ListFile(repo, commit1.ID, "dir", "", false, shard1, false)
	shard2 := &pfs.Shard{
		FileNumber:  1,
		FileModulus: 2,
	}
	fileInfos2, err := client.ListFile(repo, commit1.ID, "dir", "", false, shard2, false)
	require.Equal(t, 2, len(fileInfos1)+len(fileInfos2))
}

func TestListFileWithFiltering(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	content := "foo\n"
	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	numFiles := 100
	for i := 0; i < numFiles; i++ {
		_, err = client.PutFile(repo, "master", fmt.Sprintf("file%d", i), strings.NewReader(content))
		require.NoError(t, err)
	}
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileShard1 := &pfs.Shard{
		FileNumber:  0,
		FileModulus: 2,
	}
	fileInfos1, err := client.ListFile(repo, commit.ID, "", "", false, fileShard1, false)
	require.NoError(t, err)
	fileShard2 := &pfs.Shard{
		FileNumber:  1,
		FileModulus: 2,
	}
	fileInfos2, err := client.ListFile(repo, commit.ID, "", "", false, fileShard2, false)
	require.NoError(t, err)
	getTotalSize := func(fileInfos []*pfs.FileInfo) int {
		var totalSize int
		for _, fileInfo := range fileInfos {
			totalSize += int(fileInfo.SizeBytes)
		}
		return totalSize
	}
	require.True(t, getTotalSize(fileInfos1) > 0)
	require.True(t, getTotalSize(fileInfos2) > 0)
	require.Equal(t, numFiles*len(content), getTotalSize(fileInfos1)+getTotalSize(fileInfos2))
	require.True(t, len(fileInfos1) > 0)
	require.True(t, len(fileInfos2) > 0)
	require.Equal(t, numFiles, len(fileInfos1)+len(fileInfos2))

	blockShard1 := &pfs.Shard{
		BlockNumber:  0,
		BlockModulus: 2,
	}
	fileInfos1, err = client.ListFile(repo, commit.ID, "", "", false, blockShard1, false)
	require.NoError(t, err)
	blockShard2 := &pfs.Shard{
		BlockNumber:  1,
		BlockModulus: 2,
	}
	fileInfos2, err = client.ListFile(repo, commit.ID, "", "", false, blockShard2, false)
	require.NoError(t, err)
	require.True(t, getTotalSize(fileInfos1) > 0)
	require.True(t, getTotalSize(fileInfos2) > 0)
	require.Equal(t, numFiles*len(content), getTotalSize(fileInfos1)+getTotalSize(fileInfos2))
	require.True(t, len(fileInfos1) > 0)
	require.True(t, len(fileInfos2) > 0)
	require.Equal(t, numFiles, len(fileInfos1)+len(fileInfos2))
}

func TestMergeProvenance(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo1 := "test1"
	require.NoError(t, client.CreateRepo(repo1))
	repo2 := "test2"
	_, err := client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo:       pclient.NewRepo(repo2),
			Provenance: []*pfs.Repo{pclient.NewRepo(repo1)},
		},
	)
	require.NoError(t, err)

	// Create two commits in repo1
	p1, err := client.StartCommit(repo1, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo1, "master"))
	p2, err := client.StartCommit(repo1, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo1, "master"))

	// Create two commits in repo 2, on different branches
	_, err = client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit(repo2, "A"),
			Provenance: []*pfs.Commit{p1},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo2, "A"))
	_, err = client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit(repo2, "B"),
			Provenance: []*pfs.Commit{p2},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo2, "B"))

	commit, err := squash(client, repo2, []string{"A", "B"}, "master")
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(commit.Repo.Name, commit.ID)
	require.Equal(t, 2, len(commitInfo.Provenance))
	require.Equal(t, p1.ID, commitInfo.Provenance[0].ID)
	require.Equal(t, p2.ID, commitInfo.Provenance[1].ID)
}

func TestSquashMergeDeletion(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commitRoot, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader("buzz"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	createThreeCommits := func(branch string) {
		_, err = client.ForkCommit(repo, commitRoot.ID, branch)
		require.NoError(t, err)
		_, err = client.PutFile(repo, branch, "file", strings.NewReader("foo"))
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, branch))

		_, err = client.StartCommit(repo, branch)
		require.NoError(t, err)
		require.NoError(t, client.DeleteFile(repo, branch, "file"))
		require.NoError(t, client.FinishCommit(repo, branch))

		_, err = client.StartCommit(repo, branch)
		require.NoError(t, err)
		_, err = client.PutFile(repo, branch, "file", strings.NewReader("bar"))
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, branch))
	}

	createThreeCommits("A")
	createThreeCommits("B")
	createThreeCommits("C")

	commit, err := squash(client, repo, []string{"A", "B", "C"}, "squash")
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit.ID, "file", 0, 0, "", false, nil, buffer))
	require.Equal(t, "barbarbar", buffer.String())

	mergedCommits, err := client.ReplayCommit(repo, []string{"A", "B", "C"}, "replay")
	require.NoError(t, err)
	// 3 commits on each branch plus 1 commit on master
	require.Equal(t, 10, len(mergedCommits))

	buffer = &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, mergedCommits[9].ID, "file", 0, 0, "", false, nil, buffer))
	require.Equal(t, "bar", buffer.String())
}

func TestCreateDirConflict(t *testing.T) {
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader("foo"))
	require.NoError(t, err)

	require.YesError(t, client.MakeDirectory(repo, commit.ID, "file"))

	require.NoError(t, client.FinishCommit(repo, "master"))
}

func TestListCommitOrder(t *testing.T) {
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	numCommits := 10

	for i := 0; i < numCommits; i++ {
		_, err := client.StartCommit(repo, "master")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, "master"))
	}

	var lastCommit *pfs.Commit
	var received int
	for {
		var fromCommits []*pfs.Commit
		if lastCommit != nil {
			fromCommits = append(fromCommits, pclient.NewCommit(repo, lastCommit.ID))
		} else {
			fromCommits = append(fromCommits, pclient.NewCommit(repo, ""))
		}
		commitInfos, err := client.ListCommit(fromCommits, nil, pclient.CommitTypeRead, pclient.CommitStatusNormal, true)
		require.NoError(t, err)
		for _, commitInfo := range commitInfos {
			received++
			require.Equal(t, lastCommit, commitInfo.ParentCommit)
			lastCommit = commitInfo.Commit
		}
		if received == numCommits {
			break
		}
	}
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ALPHABET[rand.Intn(len(ALPHABET))]
	}
	return string(b)
}

func getBlockClient(t *testing.T) pfs.BlockAPIClient {
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
				pfs.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: version.Version},
			protoserver.ServeEnv{GRPCPort: uint16(localPort)},
		)
		require.NoError(t, err)
	}()
	<-ready
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	return pfs.NewBlockAPIClient(clientConn)
}

func runServers(t *testing.T, port int32, apiServer pfs.APIServer,
	blockAPIServer pfs.BlockAPIServer) {
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfs.RegisterAPIServer(s, apiServer)
				pfs.RegisterBlockAPIServer(s, blockAPIServer)
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
	return pclient.APIClient{PfsAPIClient: pfs.NewAPIClient(clientConn)}
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}

func squash(client pclient.APIClient, repo string, fromCommits []string, branch string) (*pfs.Commit, error) {
	commit, err := client.StartCommit(repo, branch)
	if err != nil {
		return nil, err
	}
	if err := client.SquashCommit(repo, fromCommits, commit.ID); err != nil {
		return nil, err
	}
	if err := client.FinishCommit(repo, commit.ID); err != nil {
		return nil, err
	}
	return commit, nil
}
