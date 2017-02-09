package server

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	servers = 2

	ALPHABET = "abcdefghijklmnopqrstuvwxyz"
)

var (
	port int32 = 30651
)

var testDBs []string

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

func TestCreateSameRepoInParallel(t *testing.T) {
	client := getClient(t)

	numGoros := 1000
	errCh := make(chan error)
	for i := 0; i < numGoros; i++ {
		go func() {
			errCh <- client.CreateRepo("repo")
		}()
	}
	successCount := 0
	totalCount := 0
	for err := range errCh {
		totalCount += 1
		if err == nil {
			successCount += 1
		}
		if totalCount == numGoros {
			break
		}
	}
	// When creating the same repo, precisiely one attempt should succeed
	require.Equal(t, 1, successCount)
}

func TestCreateDifferentRepoInParallel(t *testing.T) {
	client := getClient(t)

	numGoros := 1000
	errCh := make(chan error)
	for i := 0; i < numGoros; i++ {
		i := i
		go func() {
			errCh <- client.CreateRepo(fmt.Sprintf("repo%d", i))
		}()
	}
	successCount := 0
	totalCount := 0
	for err := range errCh {
		totalCount += 1
		if err == nil {
			successCount += 1
		}
		if totalCount == numGoros {
			break
		}
	}
	require.Equal(t, numGoros, successCount)
}

func TestCreateRepoDeleteRepoRace(t *testing.T) {
	client := getClient(t)

	for i := 0; i < 1000; i++ {
		require.NoError(t, client.CreateRepo("foo"))
		errCh := make(chan error)
		go func() {
			errCh <- client.DeleteRepo("foo", false)
		}()
		go func() {
			_, err := client.PfsAPIClient.CreateRepo(
				context.Background(),
				&pfs.CreateRepoRequest{
					Repo:       pclient.NewRepo("bar"),
					Provenance: []*pfs.Repo{pclient.NewRepo("foo")},
				},
			)
			errCh <- err
		}()
		err1 := <-errCh
		err2 := <-errCh
		// these two operations should never race in such a way that they
		// both succeed, leaving us with a repo bar that has a nonexistent
		// provenance foo
		require.True(t, err1 != nil || err2 != nil)
		client.DeleteRepo("bar", false)
		client.DeleteRepo("foo", false)
	}
}

func TestCreateAndInspectRepo(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "repo"
	require.NoError(t, client.CreateRepo(repo))

	repoInfo, err := client.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, repo, repoInfo.Repo.Name)
	require.NotNil(t, repoInfo.Created)
	require.Equal(t, 0, int(repoInfo.SizeBytes))

	require.YesError(t, client.CreateRepo(repo))
	_, err = client.InspectRepo("nonexistent")
	require.YesError(t, err)

	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo: pclient.NewRepo("somerepo1"),
		Provenance: []*pfs.Repo{
			pclient.NewRepo(repo),
		},
	})
	require.NoError(t, err)

	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo: pclient.NewRepo("somerepo2"),
		Provenance: []*pfs.Repo{
			pclient.NewRepo("nonexistent"),
		},
	})
	require.YesError(t, err)
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

func TestListRepoWithProvenance(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	require.NoError(t, client.CreateRepo("prov1"))
	require.NoError(t, client.CreateRepo("prov2"))
	require.NoError(t, client.CreateRepo("prov3"))

	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo: pclient.NewRepo("repo"),
		Provenance: []*pfs.Repo{
			pclient.NewRepo("prov1"),
			pclient.NewRepo("prov2"),
		},
	})
	require.NoError(t, err)

	repoInfos, err := client.ListRepo([]string{"prov1"})
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfos))
	require.Equal(t, "repo", repoInfos[0].Repo.Name)

	repoInfos, err = client.ListRepo([]string{"prov1", "prov2"})
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfos))
	require.Equal(t, "repo", repoInfos[0].Repo.Name)

	repoInfos, err = client.ListRepo([]string{"prov3"})
	require.NoError(t, err)
	require.Equal(t, 0, len(repoInfos))

	_, err = client.ListRepo([]string{"nonexistent"})
	require.YesError(t, err)
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

func TestBasicFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "repo"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	file := "file"
	data := "data"
	_, err = client.PutFile(repo, commit.ID, file, strings.NewReader(data))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	var b bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit.ID, "file", 0, 0, &b))

	require.Equal(t, data, b.String())
}

func TestSimpleFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit2.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestPutFileBig(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	rawMessage := "Some\ncontent\nthat\nshouldnt\nbe\nline\ndelimited.\n"

	// Write a big blob that would normally not fit in a block
	var expectedOutputA []byte
	for !(len(expectedOutputA) > 5*1024*1024) {
		expectedOutputA = append(expectedOutputA, []byte(rawMessage)...)
	}
	r := strings.NewReader(string(expectedOutputA))

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", r)
	err = client.FinishCommit(repo, commit1.ID)
	require.NoError(t, err)

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, string(expectedOutputA), buffer.String())
}

func TestPutFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	// Detect file conflict
	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo/bar", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.YesError(t, client.FinishCommit(repo, commit1.ID))

	commit1, err = client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	// file conflicts with the previous commit
	_, err = client.PutFile(repo, commit2.ID, "foo/bar", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "/bar", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.YesError(t, client.FinishCommit(repo, commit2.ID))

	commit2, err = client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
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
	require.NoError(t, client.FinishCommit(repo, commit4.ID))

	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit4.ID, "dir2/bar", 0, 0, &buffer))
	require.Equal(t, "bar\n", buffer.String())
	buffer = bytes.Buffer{}
	require.YesError(t, client.GetFile(repo, commit4.ID, "dir2", 0, 0, &buffer))
}

func TestPutFileLongName(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileName := `oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)oaidhzoshd()&)(@*^$@(#)oandoancoasid1)(&@$)(@U)`

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, fileName, strings.NewReader("foo\n"))
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit.ID, fileName, 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

func TestPutSameFileInParallel(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
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
	require.NoError(t, client.GetFile(repo, commit.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())
}

func TestListFileTwoCommits(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	numFiles := 5

	commit1, err := client.StartCommit(repo, "")
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

	fileInfos, err := client.ListFile(repo, commit1.ID, "")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))

	fileInfos, err = client.ListFile(repo, commit2.ID, "")
	require.NoError(t, err)
	require.Equal(t, 2*numFiles, len(fileInfos))
}

func TestListFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	fileContent1 := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "dir/foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	fileContent2 := "bar\n"
	_, err = client.PutFile(repo, commit.ID, "dir/bar", strings.NewReader(fileContent2))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	fileInfos, err := client.ListFile(repo, commit.ID, "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
	require.True(t, fileInfos[0].File.Path == "dir/foo" && fileInfos[1].File.Path == "dir/bar" || fileInfos[0].File.Path == "dir/bar" && fileInfos[1].File.Path == "dir/foo")
	require.True(t, fileInfos[0].SizeBytes == fileInfos[1].SizeBytes && fileInfos[0].SizeBytes == uint64(len(fileContent1)))

	fileInfos, err = client.ListFile(repo, commit.ID, "dir/foo")
	require.YesError(t, err)
}

func TestDeleteFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	// Commit 1: Add two files; delete one file within the commit
	commit1, err := client.StartCommit(repo, "")
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
	_, err = client.InspectFile(repo, commit1.ID, "foo")
	require.YesError(t, err)

	// Should see one file
	fileInfos, err := client.ListFile(repo, commit1.ID, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	// Empty commit
	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	// Should still see one files
	fileInfos, err = client.ListFile(repo, commit2.ID, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	// Delete bar
	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)
	require.NoError(t, client.DeleteFile(repo, commit3.ID, "bar"))
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	// Should see no file
	fileInfos, err = client.ListFile(repo, commit3.ID, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	_, err = client.InspectFile(repo, commit3.ID, "bar")
	require.YesError(t, err)
}

func TestDeleteDir(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	// Commit 1: Add two files into the same directory; delete the directory
	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit1.ID, "dir/foo", strings.NewReader("foo1"))
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit1.ID, "dir/bar", strings.NewReader("bar1"))
	require.NoError(t, err)

	require.NoError(t, client.DeleteFile(repo, commit1.ID, "dir"))

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfos, err := client.ListFile(repo, commit1.ID, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// dir should not exist
	_, err = client.InspectFile(repo, commit1.ID, "dir")
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
	fileInfos, err = client.ListFile(repo, commit2.ID, "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit2.ID, "dir/foo", 0, 0, &buffer))
	require.Equal(t, "foo2", buffer.String())

	var buffer2 bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit2.ID, "dir/bar", 0, 0, &buffer2))
	require.Equal(t, "bar2", buffer2.String())

	// Commit 3: delete the directory
	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)

	require.NoError(t, client.DeleteFile(repo, commit3.ID, "dir"))

	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	// Should see zero files
	fileInfos, err = client.ListFile(repo, commit3.ID, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// TODO: test deleting "."
}

func TestDeleteFile2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	err = client.DeleteFile(repo, commit2.ID, "file")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	expected := "bar\n"
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, "master/1", "file", 0, 0, &buffer))
	require.Equal(t, expected, buffer.String())

	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	err = client.DeleteFile(repo, commit3.ID, "file")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	expected = "foo\n"
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit3.ID, "file", 0, 0, &buffer))
	require.Equal(t, expected, buffer.String())
}

func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
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
		err := grpcutil.Serve(
			func(s *grpc.Server) {
				pfs.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			grpcutil.ServeOptions{Version: version.Version},
			grpcutil.ServeEnv{GRPCPort: uint16(localPort)},
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
		err := grpcutil.Serve(
			func(s *grpc.Server) {
				pfs.RegisterAPIServer(s, apiServer)
				pfs.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			grpcutil.ServeOptions{
				Version:    version.Version,
				MaxMsgSize: MaxMsgSize,
			},
			grpcutil.ServeEnv{GRPCPort: uint16(port)},
		)
		require.NoError(t, err)
	}()
	<-ready
}

func getClient(t *testing.T) pclient.APIClient {
	dbName := "pachyderm_test_" + uuid.NewWithoutDashes()[0:12]
	testDBs = append(testDBs, dbName)

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
	prefix := generateRandomString(32)
	for i, port := range ports {
		address := addresses[i]
		driver, err := drive.NewLocalDriver(address, prefix)
		require.NoError(t, err)
		blockAPIServer, err := NewLocalBlockAPIServer(root)
		require.NoError(t, err)
		apiServer := newAPIServer(driver, nil)
		runServers(t, port, apiServer, blockAPIServer)
	}
	clientConn, err := grpc.Dial(addresses[0], grpc.WithInsecure())
	require.NoError(t, err)
	return pclient.APIClient{
		PfsAPIClient:   pfs.NewAPIClient(clientConn),
		BlockAPIClient: pfs.NewBlockAPIClient(clientConn),
	}
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
