package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	authtesting "github.com/pachyderm/pachyderm/src/server/auth/testing"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	pfssync "github.com/pachyderm/pachyderm/src/server/pkg/sync"

	etcd "github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	servers = 2

	ALPHABET = "abcdefghijklmnopqrstuvwxyz"
)

var (
	port int32 = 30653
)

func CommitToID(commit interface{}) interface{} {
	return commit.(*pfs.Commit).ID
}

func CommitInfoToID(commit interface{}) interface{} {
	return commit.(*pfs.CommitInfo).Commit.ID
}

func RepoInfoToName(repoInfo interface{}) interface{} {
	return repoInfo.(*pfs.RepoInfo).Repo.Name
}

func RepoToName(repo interface{}) interface{} {
	return repo.(*pfs.Repo).Name
}

var testDBs []string

func TestInvalidRepo(t *testing.T) {
	client := getClient(t)
	require.YesError(t, client.CreateRepo("/repo"))

	require.NoError(t, client.CreateRepo("lenny"))
	require.NoError(t, client.CreateRepo("lenny123"))
	require.NoError(t, client.CreateRepo("lenny_123"))
	require.NoError(t, client.CreateRepo("lenny-123"))

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
		totalCount++
		if err == nil {
			successCount++
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
		totalCount++
		if err == nil {
			successCount++
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

func TestBranch(t *testing.T) {
	c := getClient(t)

	repo := "repo"
	require.NoError(t, c.CreateRepo(repo))
	_, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, "master"))
	commitInfo, err := c.InspectCommit(repo, "master")
	require.NoError(t, err)
	require.Nil(t, commitInfo.ParentCommit)

	_, err = c.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, "master"))
	commitInfo, err = c.InspectCommit(repo, "master")
	require.NoError(t, err)
	require.NotNil(t, commitInfo.ParentCommit)
}

func TestCreateAndInspectRepo(t *testing.T) {
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

func TestRepoSize(t *testing.T) {
	client := getClient(t)

	repo := "repo"
	require.NoError(t, client.CreateRepo(repo))

	repoInfo, err := client.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, 0, int(repoInfo.SizeBytes))

	fileContent1 := "foo"
	fileContent2 := "bar"
	fileContent3 := "buzz"
	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "bar", strings.NewReader(fileContent2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	repoInfo, err = client.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(fileContent1)+len(fileContent2), int(repoInfo.SizeBytes))

	commit, err = client.StartCommit(repo, "")
	require.NoError(t, err)
	// Deleting a file shouldn't affect the repo size, since the actual
	// data has not been removed from the storage system.
	require.NoError(t, client.DeleteFile(repo, commit.ID, "foo"))
	_, err = client.PutFile(repo, commit.ID, "buzz", strings.NewReader(fileContent3))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	repoInfo, err = client.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(fileContent1)+len(fileContent2)+len(fileContent3), int(repoInfo.SizeBytes))
}

func TestListRepo(t *testing.T) {
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
	require.ElementsEqualUnderFn(t, repoNames, repoInfos, RepoInfoToName)
}

// Make sure that commits of deleted repos do not resurface
func TestCreateDeletedRepo(t *testing.T) {
	client := getClient(t)

	repo := "repo"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	commitInfos, err := client.ListCommit(repo, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	require.NoError(t, client.DeleteRepo(repo, false))
	require.NoError(t, client.CreateRepo(repo))

	commitInfos, err = client.ListCommit(repo, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))
}

// The DAG looks like this before the update:
// prov1 prov2
//   \    /
//    repo
//   /    \
// d1      d2
//
// Looks like this after the update:
//
// prov2 prov3
//   \    /
//    repo
//   /    \
// d1      d2
func TestUpdateProvenance(t *testing.T) {
	client := getClient(t)

	prov1 := "prov1"
	require.NoError(t, client.CreateRepo(prov1))
	prov2 := "prov2"
	require.NoError(t, client.CreateRepo(prov2))
	prov3 := "prov3"
	require.NoError(t, client.CreateRepo(prov3))
	repo := "repo"
	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo: pclient.NewRepo("repo"),
		Provenance: []*pfs.Repo{
			pclient.NewRepo(prov1),
			pclient.NewRepo(prov2),
		},
	})
	require.NoError(t, err)
	downstream1 := "downstream1"
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo: pclient.NewRepo(downstream1),
		Provenance: []*pfs.Repo{
			pclient.NewRepo(repo),
		},
	})
	require.NoError(t, err)
	downstream2 := "downstream2"
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo: pclient.NewRepo(downstream2),
		Provenance: []*pfs.Repo{
			pclient.NewRepo(repo),
		},
	})
	require.NoError(t, err)

	// Without the Update flag it should fail
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo: pclient.NewRepo(repo),
		Provenance: []*pfs.Repo{
			pclient.NewRepo(prov2),
			pclient.NewRepo(prov3),
		},
	})
	require.YesError(t, err)

	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo: pclient.NewRepo(repo),
		Provenance: []*pfs.Repo{
			pclient.NewRepo(prov2),
			pclient.NewRepo(prov3),
		},
		Update: true,
	})
	require.NoError(t, err)

	// Check that the downstream repos have the correct provenance
	repoInfo, err := client.InspectRepo(repo)
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{prov2, prov3}, repoInfo.Provenance, RepoToName)

	repoInfo, err = client.InspectRepo(downstream1)
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{prov2, prov3, repo}, repoInfo.Provenance, RepoToName)

	repoInfo, err = client.InspectRepo(downstream2)
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{prov2, prov3, repo}, repoInfo.Provenance, RepoToName)

	// We should be able to delete prov1 since it's no longer the provenance
	// of other repos.
	require.NoError(t, client.DeleteRepo(prov1, false))

	// We shouldn't be able to delete prov3 since it's now a provenance
	// of other repos.
	require.YesError(t, client.DeleteRepo(prov3, false))
}

func TestPutFileIntoOpenCommit(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	_, err := client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.YesError(t, err)

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	_, err = client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.YesError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.YesError(t, err)

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	_, err = client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.YesError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.YesError(t, err)
}

func TestCreateInvalidBranchName(t *testing.T) {

	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	// Create a branch that's the same length as a commit ID
	_, err := client.StartCommit(repo, uuid.NewWithoutDashes())
	require.YesError(t, err)
}

func TestListRepoWithProvenance(t *testing.T) {
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
	require.ElementsEqualUnderFn(t, []string{"repo"}, repoInfos, RepoInfoToName)

	repoInfos, err = client.ListRepo([]string{"prov1", "prov2"})
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t, []string{"repo"}, repoInfos, RepoInfoToName)

	repoInfos, err = client.ListRepo([]string{"prov3"})
	require.NoError(t, err)
	require.Equal(t, 0, len(repoInfos))

	_, err = client.ListRepo([]string{"nonexistent"})
	require.YesError(t, err)
}

func TestDeleteRepo(t *testing.T) {
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
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	started := time.Now()
	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	tStarted, err := types.TimestampFromProto(commitInfo.Started)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.Nil(t, commitInfo.Finished)
	// PutFile does not update commit size; only FinishCommit does
	require.Equal(t, 0, int(commitInfo.SizeBytes))
	require.True(t, started.Before(tStarted))
	require.Nil(t, commitInfo.Finished)

	require.NoError(t, client.FinishCommit(repo, commit.ID))
	finished := time.Now()

	commitInfo, err = client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	tStarted, err = types.TimestampFromProto(commitInfo.Started)
	require.NoError(t, err)

	tFinished, err := types.TimestampFromProto(commitInfo.Finished)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.NotNil(t, commitInfo.Finished)
	require.Equal(t, len(fileContent), int(commitInfo.SizeBytes))
	require.True(t, started.Before(tStarted))
	require.True(t, finished.After(tFinished))
}

func TestInspectCommitBlock(t *testing.T) {
	client := getClient(t)

	repo := "TestInspectCommitBlock"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	go func() {
		time.Sleep(2 * time.Second)
		require.NoError(t, client.FinishCommit(repo, commit.ID))
	}()

	commitInfo, err := client.PfsAPIClient.InspectCommit(context.Background(),
		&pfs.InspectCommitRequest{
			Commit: commit,
			Block:  true,
		})
	require.NoError(t, err)
	require.NotNil(t, commitInfo.Finished)
}

func TestDeleteCommit(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, "master"))

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	require.NoError(t, client.DeleteCommit(repo, commit2.ID))

	commitInfo, err := client.InspectCommit(repo, commit2.ID)
	require.YesError(t, err)

	// Check that the head has been set to the parent
	commitInfo, err = client.InspectCommit(repo, "master")
	require.NoError(t, err)
	require.Equal(t, commit1.ID, commitInfo.Commit.ID)

	// Should error because the first commit has been finished
	require.YesError(t, client.DeleteCommit(repo, "master"))

	// Check that the branch still exists
	branches, err := client.ListBranch(repo)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))

	// Now start a new repo
	repo = "test2"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err = client.StartCommit(repo, "master")
	require.NoError(t, err)

	fileContent = "foo\n"
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, client.DeleteCommit(repo, "master"))

	// The branch has not been deleted, though it has no commits
	branches, err = client.ListBranch(repo)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	commits, err := client.ListCommit(repo, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 0, len(commits))

	// Check that repo size is back to 0
	repoInfo, err := client.InspectRepo(repo)
	require.Equal(t, 0, int(repoInfo.SizeBytes))
}

func TestCleanPath(t *testing.T) {
	c := getClient(t)
	repo := "TestCleanPath"
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit.ID, "./././file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	_, err = c.InspectFile(repo, commit.ID, "file")
	require.NoError(t, err)
}

func TestBasicFile(t *testing.T) {
	client := getClient(t)

	repo := "repo"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	file := "file"
	data := "data"
	_, err = client.PutFile(repo, commit.ID, file, strings.NewReader(data))
	require.NoError(t, err)
	var b bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit.ID, "file", 0, 0, &b))
	require.Equal(t, data, b.String())

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	b.Reset()
	require.NoError(t, client.GetFile(repo, commit.ID, "file", 0, 0, &b))
	require.Equal(t, data, b.String())
}

func TestSimpleFile(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit2.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	err = client.FinishCommit(repo, commit2.ID)
	require.NoError(t, err)

	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit2.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestStartCommitWithUnfinishedParent(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.StartCommit(repo, "master")
	// fails because the parent commit has not been finished
	require.YesError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
}

func TestAncestrySyntax(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "1", strings.NewReader("1"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "2", strings.NewReader("2"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	commit3, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "3", strings.NewReader("3"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	commitInfo, err := client.InspectCommit(repo, "master^")
	require.NoError(t, err)
	require.Equal(t, commit2, commitInfo.Commit)

	commitInfo, err = client.InspectCommit(repo, "master~")
	require.NoError(t, err)
	require.Equal(t, commit2, commitInfo.Commit)

	commitInfo, err = client.InspectCommit(repo, "master^1")
	require.NoError(t, err)
	require.Equal(t, commit2, commitInfo.Commit)

	commitInfo, err = client.InspectCommit(repo, "master~1")
	require.NoError(t, err)
	require.Equal(t, commit2, commitInfo.Commit)

	commitInfo, err = client.InspectCommit(repo, "master^^")
	require.NoError(t, err)
	require.Equal(t, commit1, commitInfo.Commit)

	commitInfo, err = client.InspectCommit(repo, "master~~")
	require.NoError(t, err)
	require.Equal(t, commit1, commitInfo.Commit)

	commitInfo, err = client.InspectCommit(repo, "master^2")
	require.NoError(t, err)
	require.Equal(t, commit1, commitInfo.Commit)

	commitInfo, err = client.InspectCommit(repo, "master~2")
	require.NoError(t, err)
	require.Equal(t, commit1, commitInfo.Commit)

	commitInfo, err = client.InspectCommit(repo, "master^^^")
	require.YesError(t, err)

	commitInfo, err = client.InspectCommit(repo, "master~~~")
	require.YesError(t, err)

	commitInfo, err = client.InspectCommit(repo, "master^3")
	require.YesError(t, err)

	commitInfo, err = client.InspectCommit(repo, "master~3")
	require.YesError(t, err)

	for i := 1; i <= 2; i++ {
		_, err := client.InspectFile(repo, fmt.Sprintf("%v^%v", commit3.ID, 3-i), fmt.Sprintf("%v", i))
		require.NoError(t, err)
	}
}

// TestProvenance implements the following DAG
//  A -> B -> C -> D
//           /
//  E - - - /

func TestProvenance(t *testing.T) {
	client := getClient(t)

	require.NoError(t, client.CreateRepo("A"))
	require.NoError(t, client.CreateRepo("B"))
	require.NoError(t, client.CreateRepo("C"))
	require.NoError(t, client.CreateRepo("D"))
	require.NoError(t, client.CreateRepo("E"))

	require.NoError(t, client.CreateBranch("B", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master")}))
	require.NoError(t, client.CreateBranch("C", "master", "", []*pfs.Branch{pclient.NewBranch("B", "master"), pclient.NewBranch("E", "master")}))
	require.NoError(t, client.CreateBranch("D", "master", "", []*pfs.Branch{pclient.NewBranch("C", "master")}))

	branchInfo, err := client.InspectBranch("B", "master")
	require.NoError(t, err)
	require.Equal(t, 1, len(branchInfo.Provenance))
	branchInfo, err = client.InspectBranch("C", "master")
	require.NoError(t, err)
	require.Equal(t, 3, len(branchInfo.Provenance))
	branchInfo, err = client.InspectBranch("D", "master")
	require.NoError(t, err)
	require.Equal(t, 4, len(branchInfo.Provenance))

	ACommit, err := client.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit.ID))
	ECommit, err := client.StartCommit("E", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("E", ECommit.ID))

	commitInfo, err := client.InspectCommit("B", "master")
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfo.Provenance))

	commitInfo, err = client.InspectCommit("C", "master")
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfo.Provenance))

	commitInfo, err = client.InspectCommit("D", "master")
	require.NoError(t, err)
	require.Equal(t, 4, len(commitInfo.Provenance))
}

func TestSimple(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	commitInfos, err := client.ListCommit(repo, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := client.StartCommit(repo, "master")
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

func TestBranch1(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, "master", "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	branches, err := client.ListBranch(repo)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0].Name)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = client.FinishCommit(repo, "master")
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, "master", "foo", 0, 0, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = client.ListBranch(repo)
	require.NoError(t, err)
	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0].Name)

	require.NoError(t, client.SetBranch(repo, commit.ID, "master2"))

	branches, err = client.ListBranch(repo)
	require.NoError(t, err)
	require.Equal(t, 2, len(branches))
	require.Equal(t, "master2", branches[0].Name)
	require.Equal(t, "master", branches[1].Name)
}

func TestPutFileBig(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	// Write a big blob that would normally not fit in a block
	fileSize := int(pfs.ChunkSize + 5*1024*1024)
	expectedOutputA := generateRandomString(fileSize)
	r := strings.NewReader(string(expectedOutputA))

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", r)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfo, err := client.InspectFile(repo, commit1.ID, "foo")
	require.NoError(t, err)
	require.Equal(t, fileSize, int(fileInfo.SizeBytes))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, string(expectedOutputA), buffer.String())
}

func TestPutFile(t *testing.T) {
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

	commit2, err := client.StartCommitParent(repo, "", commit1.ID)
	require.NoError(t, err)
	// file conflicts with the previous commit
	_, err = client.PutFile(repo, commit2.ID, "foo/bar", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "/bar", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.YesError(t, client.FinishCommit(repo, commit2.ID))

	commit2, err = client.StartCommitParent(repo, "", commit1.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "/bar", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	commit3, err := client.StartCommitParent(repo, "", commit2.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "dir1/foo", strings.NewReader("foo\n"))
	require.NoError(t, err) // because the directory dir does not exist
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	commit4, err := client.StartCommitParent(repo, "", commit3.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit4.ID, "dir2/bar", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit4.ID))

	buffer = bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit4.ID, "dir2/bar", 0, 0, &buffer))
	require.Equal(t, "bar\n", buffer.String())
	buffer = bytes.Buffer{}
	require.YesError(t, client.GetFile(repo, commit4.ID, "dir2", 0, 0, &buffer))
}

func TestPutFile2(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	expected := "foo\nbar\nbuzz\n"
	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit1.ID, "file", 0, 0, buffer))
	require.Equal(t, expected, buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "master", "file", 0, 0, buffer))
	require.Equal(t, expected, buffer.String())

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	expected = "foo\nbar\nbuzz\nfoo\nbar\nbuzz\n"
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit2.ID, "file", 0, 0, buffer))
	require.Equal(t, expected, buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "master", "file", 0, 0, buffer))
	require.Equal(t, expected, buffer.String())

	commit3, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit3.ID, "foo"))
	_, err = client.PutFile(repo, "foo", "file", strings.NewReader("foo\nbar\nbuzz\n"))
	require.NoError(t, client.FinishCommit(repo, "foo"))

	expected = "foo\nbar\nbuzz\nfoo\nbar\nbuzz\nfoo\nbar\nbuzz\n"
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "foo", "file", 0, 0, buffer))
	require.Equal(t, expected, buffer.String())
}

func TestPutFileLongName(t *testing.T) {
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
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	var eg errgroup.Group
	for i := 0; i < 3; i++ {
		eg.Go(func() error {
			_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader("foo\n"))
			return err
		})
	}
	require.NoError(t, eg.Wait())
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())
}

func TestInspectFile(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent1 := "foo\n"
	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)

	fileInfo, err := client.InspectFile(repo, commit1.ID, "foo")
	require.NoError(t, err)
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
	require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfo, err = client.InspectFile(repo, commit1.ID, "foo")
	require.NoError(t, err)
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
	require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))

	fileContent2 := "barbar\n"
	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo", strings.NewReader(fileContent2))
	require.NoError(t, err)

	fileInfo, err = client.InspectFile(repo, commit2.ID, "foo")
	require.NoError(t, err)
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
	require.Equal(t, len(fileContent1+fileContent2), int(fileInfo.SizeBytes))

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	fileInfo, err = client.InspectFile(repo, commit2.ID, "foo")
	require.NoError(t, err)
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
	require.Equal(t, len(fileContent1+fileContent2), int(fileInfo.SizeBytes))

	fileInfo, err = client.InspectFile(repo, commit2.ID, "foo")
	require.NoError(t, err)
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
	require.Equal(t, len(fileContent1)+len(fileContent2), int(fileInfo.SizeBytes))

	fileContent3 := "bar\n"
	commit3, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "bar", strings.NewReader(fileContent3))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	fileInfos, err := client.ListFile(repo, commit3.ID, "")
	require.NoError(t, err)
	require.Equal(t, len(fileInfos), 2)
}

func TestInspectFile2(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent1 := "foo\n"
	fileContent2 := "buzz\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader(fileContent1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfo, err := client.InspectFile(repo, "master", "/file")
	require.NoError(t, err)
	require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))
	require.Equal(t, "/file", fileInfo.File.Path)
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader(fileContent1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfo, err = client.InspectFile(repo, "master", "file")
	require.NoError(t, err)
	require.Equal(t, len(fileContent1)*2, int(fileInfo.SizeBytes))
	require.Equal(t, "file", fileInfo.File.Path)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master", "file")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader(fileContent2))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfo, err = client.InspectFile(repo, "master", "file")
	require.NoError(t, err)
	require.Equal(t, len(fileContent2), int(fileInfo.SizeBytes))
}

func TestInspectDir(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = client.PutFile(repo, commit1.ID, "dir/foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	fileInfo, err := client.InspectFile(repo, commit1.ID, "dir/foo")
	require.NoError(t, err)
	require.Equal(t, len(fileContent), int(fileInfo.SizeBytes))
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfo, err = client.InspectFile(repo, commit1.ID, "dir/foo")
	require.NoError(t, err)
	require.Equal(t, len(fileContent), int(fileInfo.SizeBytes))
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)

	fileInfo, err = client.InspectFile(repo, commit1.ID, "dir")
	require.NoError(t, err)
	require.Equal(t, len(fileContent), int(fileInfo.SizeBytes))
	require.Equal(t, pfs.FileType_DIR, fileInfo.FileType)

	_, err = client.InspectFile(repo, commit1.ID, "")
	require.NoError(t, err)
	require.Equal(t, len(fileContent), int(fileInfo.SizeBytes))
	require.Equal(t, pfs.FileType_DIR, fileInfo.FileType)
}

func TestInspectDir2(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/2", strings.NewReader(fileContent))
	require.NoError(t, err)

	fileInfo, err := client.InspectFile(repo, "master", "/dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfo.Children))
	require.Equal(t, "/dir", fileInfo.File.Path)
	require.Equal(t, pfs.FileType_DIR, fileInfo.FileType)

	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfo, err = client.InspectFile(repo, "master", "/dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfo.Children))
	require.Equal(t, "/dir", fileInfo.File.Path)
	require.Equal(t, pfs.FileType_DIR, fileInfo.FileType)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/3", strings.NewReader(fileContent))
	require.NoError(t, err)
	fileInfo, err = client.InspectFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfo.Children))

	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfo, err = client.InspectFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfo.Children))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master", "dir/2")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfo, err = client.InspectFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfo.Children))
}

func TestListFileTwoCommits(t *testing.T) {
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

	fileInfos, err := client.ListFile(repo, "master", "")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	for i := 0; i < numFiles; i++ {
		_, err = client.PutFile(repo, commit2.ID, fmt.Sprintf("file2-%d", i), strings.NewReader("foo\n"))
		require.NoError(t, err)
	}

	fileInfos, err = client.ListFile(repo, commit2.ID, "")
	require.NoError(t, err)
	require.Equal(t, 2*numFiles, len(fileInfos))

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	fileInfos, err = client.ListFile(repo, commit1.ID, "")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))

	fileInfos, err = client.ListFile(repo, commit2.ID, "")
	require.NoError(t, err)
	require.Equal(t, 2*numFiles, len(fileInfos))
}

func TestListFile(t *testing.T) {
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

	fileInfos, err := client.ListFile(repo, commit.ID, "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
	require.True(t, fileInfos[0].File.Path == "dir/foo" && fileInfos[1].File.Path == "dir/bar" || fileInfos[0].File.Path == "dir/bar" && fileInfos[1].File.Path == "dir/foo")
	require.True(t, fileInfos[0].SizeBytes == fileInfos[1].SizeBytes && fileInfos[0].SizeBytes == uint64(len(fileContent1)))

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	fileInfos, err = client.ListFile(repo, commit.ID, "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
	require.True(t, fileInfos[0].File.Path == "dir/foo" && fileInfos[1].File.Path == "dir/bar" || fileInfos[0].File.Path == "dir/bar" && fileInfos[1].File.Path == "dir/foo")
	require.True(t, fileInfos[0].SizeBytes == fileInfos[1].SizeBytes && fileInfos[0].SizeBytes == uint64(len(fileContent1)))

	fileInfos, err = client.ListFile(repo, commit.ID, "dir/foo")
	require.YesError(t, err)
}

func TestListFile2(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/2", strings.NewReader(fileContent))
	require.NoError(t, err)

	fileInfos, err := client.ListFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err = client.ListFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/3", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err = client.ListFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfos))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master", "dir/2")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err = client.ListFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
}

func TestListFile3(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	_, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/2", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err := client.ListFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/3/foo", strings.NewReader(fileContent))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/3/bar", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err = client.ListFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfos))
	require.Equal(t, int(fileInfos[2].SizeBytes), len(fileContent)*2)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, "master", "dir/3/bar")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err = client.ListFile(repo, "master", "dir")
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfos))
	require.Equal(t, int(fileInfos[2].SizeBytes), len(fileContent))

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "file", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfos, err = client.ListFile(repo, "master", "/")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))
}

func TestPutFileTypeConflict(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "dir", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.YesError(t, client.FinishCommit(repo, commit2.ID))
}

func TestRootDirectory(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	fileInfos, err := client.ListFile(repo, commit.ID, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	fileInfos, err = client.ListFile(repo, commit.ID, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))
}

func TestDeleteFile(t *testing.T) {
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

	_, err = client.InspectFile(repo, commit1.ID, "foo")
	require.YesError(t, err)

	fileInfos, err := client.ListFile(repo, commit1.ID, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	_, err = client.InspectFile(repo, commit1.ID, "foo")
	require.YesError(t, err)

	// Should see one file
	fileInfos, err = client.ListFile(repo, commit1.ID, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	// Deleting a file in a finished commit should result in an error
	require.YesError(t, client.DeleteFile(repo, commit1.ID, "bar"))

	// Empty commit
	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	// Should still see one files
	fileInfos, err = client.ListFile(repo, commit2.ID, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))

	// Delete bar
	commit3, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.DeleteFile(repo, commit3.ID, "bar"))

	// Should see no file
	fileInfos, err = client.ListFile(repo, commit3.ID, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	_, err = client.InspectFile(repo, commit3.ID, "bar")
	require.YesError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	// Should see no file
	fileInfos, err = client.ListFile(repo, commit3.ID, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	_, err = client.InspectFile(repo, commit3.ID, "bar")
	require.YesError(t, err)

	// Delete a nonexistent file; it should be no-op
	commit4, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.DeleteFile(repo, commit4.ID, "nonexistent"))
	require.NoError(t, client.FinishCommit(repo, commit4.ID))
}

func TestDeleteDir(t *testing.T) {
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

	fileInfos, err := client.ListFile(repo, commit1.ID, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfos, err = client.ListFile(repo, commit1.ID, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// dir should not exist
	_, err = client.InspectFile(repo, commit1.ID, "dir")
	require.YesError(t, err)

	// Commit 2: Delete the directory and add the same two files
	// The two files should reflect the new content
	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit2.ID, "dir/foo", strings.NewReader("foo2"))
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit2.ID, "dir/bar", strings.NewReader("bar2"))
	require.NoError(t, err)

	// Should see two files
	fileInfos, err = client.ListFile(repo, commit2.ID, "dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

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
	commit3, err := client.StartCommit(repo, "master")
	require.NoError(t, err)

	require.NoError(t, client.DeleteFile(repo, commit3.ID, "dir"))

	// Should see zero files
	fileInfos, err = client.ListFile(repo, commit3.ID, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	// Should see zero files
	fileInfos, err = client.ListFile(repo, commit3.ID, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(fileInfos))

	// TODO: test deleting "."
}

func TestDeleteFile2(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	err = client.DeleteFile(repo, commit2.ID, "file")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	expected := "bar\n"
	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, "master", "file", 0, 0, &buffer))
	require.Equal(t, expected, buffer.String())

	commit3, err := client.StartCommit(repo, "master")
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

func TestListCommit(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	numCommits := 10

	var midCommitID string
	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit(repo, "master")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, "master"))
		if i == numCommits/2 {
			midCommitID = commit.ID
		}
	}

	// list all commits
	commitInfos, err := client.ListCommit(repo, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, numCommits, len(commitInfos))

	// Test that commits are sorted in newest-first order
	for i := 0; i < len(commitInfos)-1; i++ {
		require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
	}

	// Now list all commits up to the last commit
	commitInfos, err = client.ListCommit(repo, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, numCommits, len(commitInfos))

	// Test that commits are sorted in newest-first order
	for i := 0; i < len(commitInfos)-1; i++ {
		require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
	}

	// Now list all commits up to the mid commit, excluding the mid commit
	// itself
	commitInfos, err = client.ListCommit(repo, "master", midCommitID, 0)
	require.NoError(t, err)
	require.Equal(t, numCommits-numCommits/2-1, len(commitInfos))

	// Test that commits are sorted in newest-first order
	for i := 0; i < len(commitInfos)-1; i++ {
		require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
	}

	// list commits by branch
	commitInfos, err = client.ListCommit(repo, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, numCommits, len(commitInfos))

	// Test that commits are sorted in newest-first order
	for i := 0; i < len(commitInfos)-1; i++ {
		require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
	}
}

func TestOffsetRead(t *testing.T) {
	client := getClient(t)

	repo := "TestOffsetRead"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	fileData := "foo\n"
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileData))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileData))
	require.NoError(t, err)

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit.ID, "foo", int64(len(fileData)*2)+1, 0, &buffer))
	require.Equal(t, "", buffer.String())

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit.ID, "foo", int64(len(fileData)*2)+1, 0, &buffer))
	require.Equal(t, "", buffer.String())
}

func TestBranch2(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	expectedBranches := []string{"branch1", "branch2", "branch3"}
	for _, branch := range expectedBranches {
		require.NoError(t, client.SetBranch(repo, commit.ID, branch))
	}

	branches, err := client.ListBranch(repo)
	require.Equal(t, len(expectedBranches), len(branches))
	for i, branch := range branches {
		// branches should return in newest-first order
		require.Equal(t, expectedBranches[len(branches)-i-1], branch.Name)
		require.Equal(t, commit, branch.Head)
	}

	commit2, err := client.StartCommit(repo, "branch1")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "branch1"))

	commit2Info, err := client.InspectCommit(repo, "branch1")
	require.NoError(t, err)
	require.Equal(t, commit, commit2Info.ParentCommit)

	// delete the last branch
	var lastBranch string
	lastBranch = expectedBranches[len(expectedBranches)-1]
	require.NoError(t, client.DeleteBranch(repo, lastBranch))
	branches, err = client.ListBranch(repo)
	require.Equal(t, 2, len(branches))
	require.Equal(t, "branch1", branches[0].Name)
	require.Equal(t, commit2, branches[0].Head)
	require.Equal(t, "branch2", branches[1].Name)
	require.Equal(t, commit, branches[1].Head)
}

func TestSubscribeCommit(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	numCommits := 10

	// create some commits that shouldn't affect the below SubscribeCommit call
	// reproduces #2469
	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit(repo, "master-v1")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, commit.ID))
	}

	var commits []*pfs.Commit
	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit(repo, "master")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, commit.ID))
		commits = append(commits, commit)
	}

	commitIter, err := client.SubscribeCommit(repo, "master", "")
	require.NoError(t, err)
	for i := 0; i < numCommits; i++ {
		commitInfo, err := commitIter.Next()
		require.NoError(t, err)
		require.Equal(t, commits[i], commitInfo.Commit)
	}

	// Create another batch of commits
	commits = nil
	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit(repo, "master")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, "master"))
		commits = append(commits, commit)
	}

	for i := 0; i < numCommits; i++ {
		commitInfo, err := commitIter.Next()
		require.NoError(t, err)
		require.Equal(t, commits[i], commitInfo.Commit)
	}

	commitIter.Close()
}

func TestInspectRepoSimple(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
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
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
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

func TestCreate(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	w, err := client.PutFileSplitWriter(repo, commit.ID, "foo", pfs.Delimiter_NONE, 0, 0, false)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.NoError(t, client.FinishCommit(repo, commit.ID))
	_, err = client.InspectFile(repo, commit.ID, "foo")
	require.NoError(t, err)
}

func TestGetFileInvalidCommit(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	err = client.GetFile(repo, "aninvalidcommitid", "file", 0, 0, &buffer)
	require.YesError(t, err)
}

func TestManyPutsSingleFileSingleCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long tests in short mode")
	}
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	rawMessage := `{
		"level":"debug",
		"message":{
			"thing":"foo"
		},
		"timing":[1,3,34,6,7]
	}`
	numObjs := 500
	numGoros := 10
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
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, string(expectedOutput), buffer.String())
}

func TestPutFileValidCharacters(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit.ID, "foo\x00bar", strings.NewReader("foobar\n"))
	// null characters error because when you `ls` files with null characters
	// they truncate things after the null character leading to strange results
	require.YesError(t, err)

	// Boundary tests for valid character range
	_, err = client.PutFile(repo, commit.ID, "\x1ffoobar", strings.NewReader("foobar\n"))
	require.YesError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo\x20bar", strings.NewReader("foobar\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foobar\x7e", strings.NewReader("foobar\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo\x7fbar", strings.NewReader("foobar\n"))
	require.YesError(t, err)

	// Random character tests outside and inside valid character range
	_, err = client.PutFile(repo, commit.ID, "foobar\x0b", strings.NewReader("foobar\n"))
	require.YesError(t, err)
	_, err = client.PutFile(repo, commit.ID, "\x41foobar", strings.NewReader("foobar\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo\x90bar", strings.NewReader("foobar\n"))
	require.YesError(t, err)
}

func TestPutFileURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getClient(t)

	repo := "TestPutFileURL"
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.PutFileURL(repo, commit.ID, "readme", "https://raw.githubusercontent.com/pachyderm/pachyderm/master/README.md", false, false))
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	fileInfo, err := c.InspectFile(repo, commit.ID, "readme")
	require.NoError(t, err)
	require.True(t, fileInfo.SizeBytes > 0)
}

func TestBigListFile(t *testing.T) {
	client := getClient(t)

	repo := "TestBigListFile"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	var eg errgroup.Group
	for i := 0; i < 25; i++ {
		for j := 0; j < 25; j++ {
			i := i
			j := j
			eg.Go(func() error {
				_, err = client.PutFile(repo, commit.ID, fmt.Sprintf("dir%d/file%d", i, j), strings.NewReader("foo\n"))
				return err
			})
		}
	}
	require.NoError(t, eg.Wait())
	require.NoError(t, client.FinishCommit(repo, commit.ID))
	for i := 0; i < 25; i++ {
		files, err := client.ListFile(repo, commit.ID, fmt.Sprintf("dir%d", i))
		require.NoError(t, err)
		require.Equal(t, 25, len(files))
	}
}

func TestStartCommitLatestOnBranch(t *testing.T) {
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

func TestSetBranchTwice(t *testing.T) {
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit1.ID, "master"))
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit2.ID, "master"))
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	branches, err := client.ListBranch(repo)
	require.NoError(t, err)

	require.Equal(t, 1, len(branches))
	require.Equal(t, "master", branches[0].Name)
	require.Equal(t, commit2.ID, branches[0].Head.ID)
}

func TestSyncPullPush(t *testing.T) {
	client := getClient(t)

	repo1 := "repo1"
	require.NoError(t, client.CreateRepo(repo1))

	commit1, err := client.StartCommit(repo1, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo1, commit1.ID, "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	_, err = client.PutFile(repo1, commit1.ID, "dir/bar", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo1, commit1.ID))

	tmpDir, err := ioutil.TempDir("/tmp", "pfs")
	require.NoError(t, err)

	puller := pfssync.NewPuller()
	require.NoError(t, puller.Pull(client, tmpDir, repo1, commit1.ID, "", false, false, 2, nil, ""))
	_, err = puller.CleanUp()
	require.NoError(t, err)

	repo2 := "repo2"
	require.NoError(t, client.CreateRepo(repo2))

	commit2, err := client.StartCommit(repo2, "master")
	require.NoError(t, err)

	require.NoError(t, pfssync.Push(client, tmpDir, commit2, false))
	require.NoError(t, client.FinishCommit(repo2, commit2.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo2, commit2.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo2, commit2.ID, "dir/bar", 0, 0, &buffer))
	require.Equal(t, "bar\n", buffer.String())

	fileInfos, err := client.ListFile(repo2, commit2.ID, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	commit3, err := client.StartCommit(repo2, "master")
	require.NoError(t, err)

	// Test the overwrite flag.
	// After this Push operation, all files should still look the same, since
	// the old files were overwritten.
	require.NoError(t, pfssync.Push(client, tmpDir, commit3, true))
	require.NoError(t, client.FinishCommit(repo2, commit3.ID))

	buffer.Reset()
	require.NoError(t, client.GetFile(repo2, commit3.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo2, commit3.ID, "dir/bar", 0, 0, &buffer))
	require.Equal(t, "bar\n", buffer.String())

	fileInfos, err = client.ListFile(repo2, commit3.ID, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfos))

	// Test Lazy files
	tmpDir2, err := ioutil.TempDir("/tmp", "pfs")
	require.NoError(t, err)

	puller = pfssync.NewPuller()
	require.NoError(t, puller.Pull(client, tmpDir2, repo1, "master", "", true, false, 2, nil, ""))

	data, err := ioutil.ReadFile(path.Join(tmpDir2, "dir/bar"))
	require.NoError(t, err)
	require.Equal(t, "bar\n", string(data))

	_, err = puller.CleanUp()
	require.NoError(t, err)
}

func TestSyncFile(t *testing.T) {
	client := getClient(t)

	repo := "repo"
	require.NoError(t, client.CreateRepo(repo))

	content1 := generateRandomString(int(pfs.ChunkSize))

	commit1, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, pfssync.PushFile(client, &pfs.File{
		Commit: commit1,
		Path:   "file",
	}, strings.NewReader(content1)))
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "file", 0, 0, &buffer))
	require.Equal(t, content1, buffer.String())

	content2 := generateRandomString(int(pfs.ChunkSize * 2))

	commit2, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, pfssync.PushFile(client, &pfs.File{
		Commit: commit2,
		Path:   "file",
	}, strings.NewReader(content2)))
	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit2.ID, "file", 0, 0, &buffer))
	require.Equal(t, content2, buffer.String())

	content3 := content2 + generateRandomString(int(pfs.ChunkSize))

	commit3, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, pfssync.PushFile(client, &pfs.File{
		Commit: commit3,
		Path:   "file",
	}, strings.NewReader(content3)))
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	buffer.Reset()
	require.NoError(t, client.GetFile(repo, commit3.ID, "file", 0, 0, &buffer))
	require.Equal(t, content3, buffer.String())
}

func TestSyncEmptyDir(t *testing.T) {
	client := getClient(t)

	repo := "repo"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	tmpDir, err := ioutil.TempDir("/tmp", "pfs")
	require.NoError(t, err)

	// We want to make sure that Pull creates an empty directory
	// when the path that we are cloning is empty.
	dir := filepath.Join(tmpDir, "tmp")

	puller := pfssync.NewPuller()
	require.NoError(t, puller.Pull(client, dir, repo, commit.ID, "", false, false, 0, nil, ""))
	_, err = os.Stat(dir)
	require.NoError(t, err)
	_, err = puller.CleanUp()
	require.NoError(t, err)
}

func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = ALPHABET[rand.Intn(len(ALPHABET))]
	}
	return string(b)
}

func runServers(t *testing.T, port int32, apiServer pfs.APIServer,
	blockAPIServer BlockAPIServer) {
	ready := make(chan bool)
	go func() {
		err := grpcutil.Serve(
			func(s *grpc.Server) {
				pfs.RegisterAPIServer(s, apiServer)
				pfs.RegisterObjectAPIServer(s, blockAPIServer)
				auth.RegisterAPIServer(s, &authtesting.InactiveAPIServer{}) // PFS server uses auth API
				close(ready)
			},
			grpcutil.ServeOptions{
				Version:    version.Version,
				MaxMsgSize: grpcutil.MaxMsgSize,
			},
			grpcutil.ServeEnv{GRPCPort: uint16(port)},
		)
		require.NoError(t, err)
	}()
	<-ready
}

var etcdOnce sync.Once

func getClient(t *testing.T) *pclient.APIClient {
	// src/server/pfs/server/driver.go expects an etcd server at "localhost:32379"
	// Try to establish a connection before proceeding with the test (which will
	// fail if the connection can't be established)
	etcdAddress := "localhost:32379"
	etcdOnce.Do(func() {
		require.NoError(t, backoff.Retry(func() error {
			_, err := etcd.New(etcd.Config{
				Endpoints:   []string{etcdAddress},
				DialOptions: pclient.EtcdDialOptions(),
			})
			if err != nil {
				return fmt.Errorf("could not connect to etcd: %s", err.Error())
			}
			return nil
		}, backoff.NewTestingBackOff()))
	})
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
		blockAPIServer, err := newLocalBlockAPIServer(root, 256*1024*1024, etcdAddress)
		require.NoError(t, err)
		apiServer, err := newLocalAPIServer(address, prefix)
		require.NoError(t, err)
		runServers(t, port, apiServer, blockAPIServer)
	}
	c, err := pclient.NewFromAddress(addresses[0])
	require.NoError(t, err)
	return c
}

func collectCommitInfos(commitInfoIter pclient.CommitInfoIterator) ([]*pfs.CommitInfo, error) {
	var commitInfos []*pfs.CommitInfo
	for {
		commitInfo, err := commitInfoIter.Next()
		if err == io.EOF {
			return commitInfos, nil
		}
		if err != nil {
			return nil, err
		}
		commitInfos = append(commitInfos, commitInfo)
	}
}

func TestFlush(t *testing.T) {
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	require.NoError(t, client.CreateRepo("B"))
	require.NoError(t, client.CreateBranch("B", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master")}))
	ACommit, err := client.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", "master"))
	require.NoError(t, client.FinishCommit("B", "master"))
	commitInfoIter, err := client.FlushCommit([]*pfs.Commit{pclient.NewCommit("A", ACommit.ID)}, nil)
	require.NoError(t, err)
	commitInfos, err := collectCommitInfos(commitInfoIter)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
}

// TestFlush2 implements the following DAG:
// A -> B -> C -> D
func TestFlush2(t *testing.T) {
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	require.NoError(t, client.CreateRepo("B"))
	require.NoError(t, client.CreateRepo("C"))
	require.NoError(t, client.CreateRepo("D"))
	require.NoError(t, client.CreateBranch("B", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master")}))
	require.NoError(t, client.CreateBranch("C", "master", "", []*pfs.Branch{pclient.NewBranch("B", "master")}))
	require.NoError(t, client.CreateBranch("D", "master", "", []*pfs.Branch{pclient.NewBranch("C", "master")}))
	ACommit, err := client.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", "master"))

	// do the other commits in a goro so we can block for them
	go func() {
		require.NoError(t, client.FinishCommit("B", "master"))
		require.NoError(t, client.FinishCommit("C", "master"))
		require.NoError(t, client.FinishCommit("D", "master"))
	}()

	// Flush ACommit
	commitInfoIter, err := client.FlushCommit([]*pfs.Commit{pclient.NewCommit("A", ACommit.ID)}, nil)
	require.NoError(t, err)
	commitInfos, err := collectCommitInfos(commitInfoIter)
	require.NoError(t, err)
	require.Equal(t, 3, len(commitInfos))

	commitInfoIter, err = client.FlushCommit(
		[]*pfs.Commit{pclient.NewCommit("A", ACommit.ID)},
		[]*pfs.Repo{pclient.NewRepo("C")},
	)
	require.NoError(t, err)
	commitInfos, err = collectCommitInfos(commitInfoIter)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
}

// A
//  \
//   C
//  /
// B
func TestFlush3(t *testing.T) {
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	require.NoError(t, client.CreateRepo("B"))
	require.NoError(t, client.CreateRepo("C"))

	require.NoError(t, client.CreateBranch("C", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master"), pclient.NewBranch("B", "master")}))

	ACommit, err := client.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit.ID))
	require.NoError(t, client.FinishCommit("C", "master"))
	BCommit, err := client.StartCommit("B", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("B", BCommit.ID))
	require.NoError(t, client.FinishCommit("C", "master"))

	BCommit, err = client.StartCommit("B", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("B", BCommit.ID))
	require.NoError(t, client.FinishCommit("C", "master"))

	commitIter, err := client.FlushCommit([]*pfs.Commit{pclient.NewCommit("B", BCommit.ID), pclient.NewCommit("A", ACommit.ID)}, nil)
	require.NoError(t, err)
	commitInfos, err := collectCommitInfos(commitIter)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	require.Equal(t, commitInfos[0].Commit.Repo.Name, "C")
}

func TestFlushCommitWithNoDownstreamRepos(t *testing.T) {
	c := getClient(t)
	repo := "test"
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{pclient.NewCommit(repo, commit.ID)}, nil)
	require.NoError(t, err)
	commitInfos, err := collectCommitInfos(commitIter)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))
}

func TestFlushOpenCommit(t *testing.T) {

	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	require.NoError(t, client.CreateRepo("B"))
	require.NoError(t, client.CreateBranch("B", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master")}))
	ACommit, err := client.StartCommit("A", "master")
	require.NoError(t, err)

	// do the other commits in a goro so we can block for them
	go func() {
		time.Sleep(5 * time.Second)
		require.NoError(t, client.FinishCommit("A", "master"))
		require.NoError(t, client.FinishCommit("B", "master"))
	}()

	// Flush ACommit
	commitIter, err := client.FlushCommit([]*pfs.Commit{pclient.NewCommit("A", ACommit.ID)}, nil)
	require.NoError(t, err)
	commitInfos, err := collectCommitInfos(commitIter)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
}

func TestEmptyFlush(t *testing.T) {

	client := getClient(t)
	commitIter, err := client.FlushCommit(nil, nil)
	require.NoError(t, err)
	_, err = collectCommitInfos(commitIter)
	require.YesError(t, err)
}

func TestFlushNonExistentCommit(t *testing.T) {

	c := getClient(t)
	iter, err := c.FlushCommit([]*pfs.Commit{pclient.NewCommit("fake-repo", "fake-commit")}, nil)
	require.NoError(t, err)
	_, err = collectCommitInfos(iter)
	require.YesError(t, err)
	repo := "FlushNonExistentCommit"
	require.NoError(t, c.CreateRepo(repo))
	_, err = c.FlushCommit([]*pfs.Commit{pclient.NewCommit(repo, "fake-commit")}, nil)
	require.NoError(t, err)
	_, err = collectCommitInfos(iter)
	require.YesError(t, err)
}

func TestPutFileSplit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getClient(t)
	// create repos
	repo := uniqueString("TestPutFileSplit")
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "none", pfs.Delimiter_NONE, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "line2", pfs.Delimiter_LINE, 2, 0, false, strings.NewReader("foo\nbar\nbuz\nfiz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "line3", pfs.Delimiter_LINE, 0, 8, false, strings.NewReader("foo\nbar\nbuz\nfiz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "json", pfs.Delimiter_JSON, 0, 0, false, strings.NewReader("{}{}{}{}{}{}{}{}{}{}"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "json", pfs.Delimiter_JSON, 0, 0, false, strings.NewReader("{}{}{}{}{}{}{}{}{}{}"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "json2", pfs.Delimiter_JSON, 2, 0, false, strings.NewReader("{}{}{}{}"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "json3", pfs.Delimiter_JSON, 0, 4, false, strings.NewReader("{}{}{}{}"))
	require.NoError(t, err)

	files, err := c.ListFile(repo, commit.ID, "line2")
	require.NoError(t, err)
	require.Equal(t, 2, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(8), fileInfo.SizeBytes)
	}

	require.NoError(t, c.FinishCommit(repo, commit.ID))
	commit2, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit2.ID, "line", pfs.Delimiter_LINE, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit2.ID, "json", pfs.Delimiter_JSON, 0, 0, false, strings.NewReader("{}{}{}{}{}{}{}{}{}{}"))
	require.NoError(t, err)

	files, err = c.ListFile(repo, commit2.ID, "line")
	require.NoError(t, err)
	require.Equal(t, 9, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(4), fileInfo.SizeBytes)
	}

	require.NoError(t, c.FinishCommit(repo, commit2.ID))
	fileInfo, err := c.InspectFile(repo, commit.ID, "none")
	require.NoError(t, err)
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
	files, err = c.ListFile(repo, commit.ID, "line")
	require.NoError(t, err)
	require.Equal(t, 6, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(4), fileInfo.SizeBytes)
	}
	files, err = c.ListFile(repo, commit2.ID, "line")
	require.NoError(t, err)
	require.Equal(t, 9, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(4), fileInfo.SizeBytes)
	}
	files, err = c.ListFile(repo, commit.ID, "line2")
	require.NoError(t, err)
	require.Equal(t, 2, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(8), fileInfo.SizeBytes)
	}
	files, err = c.ListFile(repo, commit.ID, "line3")
	require.NoError(t, err)
	require.Equal(t, 2, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(8), fileInfo.SizeBytes)
	}
	files, err = c.ListFile(repo, commit.ID, "json")
	require.NoError(t, err)
	require.Equal(t, 20, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(2), fileInfo.SizeBytes)
	}
	files, err = c.ListFile(repo, commit2.ID, "json")
	require.NoError(t, err)
	require.Equal(t, 30, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(2), fileInfo.SizeBytes)
	}
	files, err = c.ListFile(repo, commit.ID, "json2")
	require.NoError(t, err)
	require.Equal(t, 2, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(4), fileInfo.SizeBytes)
	}
	files, err = c.ListFile(repo, commit.ID, "json3")
	require.NoError(t, err)
	require.Equal(t, 2, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(4), fileInfo.SizeBytes)
	}
}

func TestPutFileSplitDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getClient(t)
	// create repos
	repo := uniqueString("TestPutFileSplitDelete")
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	require.NoError(t, c.DeleteFile(repo, commit.ID, fmt.Sprintf("line/%016x", 0)))
	_, err = c.PutFileSplit(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	require.NoError(t, c.DeleteFile(repo, commit.ID, fmt.Sprintf("line/%016x", 5)))
	_, err = c.PutFileSplit(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)

	files, err := c.ListFile(repo, commit.ID, "line")
	require.NoError(t, err)
	require.Equal(t, 7, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(4), fileInfo.SizeBytes)
	}

	require.NoError(t, c.FinishCommit(repo, commit.ID))

	files, err = c.ListFile(repo, commit.ID, "line")
	require.NoError(t, err)
	require.Equal(t, 7, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(4), fileInfo.SizeBytes)
	}
}

func TestPutFileSplitBig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getClient(t)
	// create repos
	repo := uniqueString("TestPutFileSplitBig")
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	w, err := c.PutFileSplitWriter(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, false)
	require.NoError(t, err)
	for i := 0; i < 1000; i++ {
		_, err = w.Write([]byte("foo\n"))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	files, err := c.ListFile(repo, commit.ID, "line")
	require.NoError(t, err)
	require.Equal(t, 1000, len(files))
	for _, fileInfo := range files {
		require.Equal(t, uint64(4), fileInfo.SizeBytes)
	}
}

func TestDiff(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getClient(t)
	repo := uniqueString("TestDiff")
	require.NoError(t, c.CreateRepo(repo))

	// Write foo
	_, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)

	newFiles, oldFiles, err := c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 1, len(newFiles))
	require.Equal(t, "foo", newFiles[0].File.Path)
	require.Equal(t, 0, len(oldFiles))

	require.NoError(t, c.FinishCommit(repo, "master"))

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 1, len(newFiles))
	require.Equal(t, "foo", newFiles[0].File.Path)
	require.Equal(t, 0, len(oldFiles))

	// Change the value of foo
	_, err = c.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.DeleteFile(repo, "master", "foo"))
	_, err = c.PutFile(repo, "master", "foo", strings.NewReader("not foo\n"))
	require.NoError(t, err)

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 1, len(newFiles))
	require.Equal(t, "foo", newFiles[0].File.Path)
	require.Equal(t, 1, len(oldFiles))
	require.Equal(t, "foo", oldFiles[0].File.Path)

	require.NoError(t, c.FinishCommit(repo, "master"))

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 1, len(newFiles))
	require.Equal(t, "foo", newFiles[0].File.Path)
	require.Equal(t, 1, len(oldFiles))
	require.Equal(t, "foo", oldFiles[0].File.Path)

	// Write bar
	_, err = c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, "master", "bar", strings.NewReader("bar\n"))
	require.NoError(t, err)

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 1, len(newFiles))
	require.Equal(t, "bar", newFiles[0].File.Path)
	require.Equal(t, 0, len(oldFiles))

	require.NoError(t, c.FinishCommit(repo, "master"))

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 1, len(newFiles))
	require.Equal(t, "bar", newFiles[0].File.Path)
	require.Equal(t, 0, len(oldFiles))

	// Delete bar
	_, err = c.StartCommit(repo, "master")
	require.NoError(t, err)
	require.NoError(t, c.DeleteFile(repo, "master", "bar"))

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 0, len(newFiles))
	require.Equal(t, 1, len(oldFiles))
	require.Equal(t, "bar", oldFiles[0].File.Path)

	require.NoError(t, c.FinishCommit(repo, "master"))

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 0, len(newFiles))
	require.Equal(t, 1, len(oldFiles))
	require.Equal(t, "bar", oldFiles[0].File.Path)

	// Write dir/fizz and dir/buzz
	_, err = c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, "master", "dir/fizz", strings.NewReader("fizz\n"))
	require.NoError(t, err)
	_, err = c.PutFile(repo, "master", "dir/buzz", strings.NewReader("buzz\n"))
	require.NoError(t, err)

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 2, len(newFiles))
	require.Equal(t, 0, len(oldFiles))

	require.NoError(t, c.FinishCommit(repo, "master"))

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 2, len(newFiles))
	require.Equal(t, 0, len(oldFiles))

	// Modify dir/fizz
	_, err = c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, "master", "dir/fizz", strings.NewReader("fizz\n"))
	require.NoError(t, err)

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 1, len(newFiles))
	require.Equal(t, "dir/fizz", newFiles[0].File.Path)
	require.Equal(t, 1, len(oldFiles))
	require.Equal(t, "dir/fizz", oldFiles[0].File.Path)

	require.NoError(t, c.FinishCommit(repo, "master"))

	newFiles, oldFiles, err = c.DiffFile(repo, "master", "", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, 1, len(newFiles))
	require.Equal(t, "dir/fizz", newFiles[0].File.Path)
	require.Equal(t, 1, len(oldFiles))
	require.Equal(t, "dir/fizz", oldFiles[0].File.Path)
}

func TestGlob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getClient(t)
	repo := uniqueString("TestGlob")
	require.NoError(t, c.CreateRepo(repo))

	// Write foo
	numFiles := 100
	_, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(repo, "master", fmt.Sprintf("file%d", i), strings.NewReader(""))
		require.NoError(t, err)
		_, err = c.PutFile(repo, "master", fmt.Sprintf("dir1/file%d", i), strings.NewReader(""))
		require.NoError(t, err)
		_, err = c.PutFile(repo, "master", fmt.Sprintf("dir2/dir3/file%d", i), strings.NewReader(""))
		require.NoError(t, err)
	}

	fileInfos, err := c.GlobFile(repo, "master", "*")
	require.NoError(t, err)
	require.Equal(t, numFiles+2, len(fileInfos))
	fileInfos, err = c.GlobFile(repo, "master", "file*")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))
	fileInfos, err = c.GlobFile(repo, "master", "dir1/*")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))
	fileInfos, err = c.GlobFile(repo, "master", "dir2/dir3/*")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))
	fileInfos, err = c.GlobFile(repo, "master", "*/*")
	require.NoError(t, err)
	require.Equal(t, numFiles+1, len(fileInfos))

	require.NoError(t, c.FinishCommit(repo, "master"))

	fileInfos, err = c.GlobFile(repo, "master", "*")
	require.NoError(t, err)
	require.Equal(t, numFiles+2, len(fileInfos))
	fileInfos, err = c.GlobFile(repo, "master", "file*")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))
	fileInfos, err = c.GlobFile(repo, "master", "dir1/*")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))
	fileInfos, err = c.GlobFile(repo, "master", "dir2/dir3/*")
	require.NoError(t, err)
	require.Equal(t, numFiles, len(fileInfos))
	fileInfos, err = c.GlobFile(repo, "master", "*/*")
	require.NoError(t, err)
	require.Equal(t, numFiles+1, len(fileInfos))
}

func TestOverwrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getClient(t)
	repo := uniqueString("TestGlob")
	require.NoError(t, c.CreateRepo(repo))

	// Write foo
	_, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(repo, "master", "file1", strings.NewReader("foo"))
	_, err = c.PutFileSplit(repo, "master", "file2", pfs.Delimiter_LINE, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, "master", "file3", pfs.Delimiter_LINE, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, "master"))
	_, err = c.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = c.PutFileOverwrite(repo, "master", "file1", strings.NewReader("bar"), 0)
	require.NoError(t, err)
	_, err = c.PutFileOverwrite(repo, "master", "file2", strings.NewReader("buzz"), 0)
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, "master", "file3", pfs.Delimiter_LINE, 0, 0, true, strings.NewReader("0\n1\n2\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(repo, "master", "file1", 0, 0, &buffer))
	require.Equal(t, "bar", buffer.String())
	buffer.Reset()
	require.NoError(t, c.GetFile(repo, "master", "file2", 0, 0, &buffer))
	require.Equal(t, "buzz", buffer.String())
	fileInfos, err := c.ListFile(repo, "master", "file3")
	require.NoError(t, err)
	require.Equal(t, 3, len(fileInfos))
	for i := 0; i < 3; i++ {
		buffer.Reset()
		require.NoError(t, c.GetFile(repo, "master", fmt.Sprintf("file3/%016x", i), 0, 0, &buffer))
		require.Equal(t, fmt.Sprintf("%d\n", i), buffer.String())
	}
}

func TestCopyFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getClient(t)
	repo := uniqueString("TestCopyFile")
	require.NoError(t, c.CreateRepo(repo))
	_, err := c.StartCommit(repo, "master")
	require.NoError(t, err)
	numFiles := 5
	for i := 0; i < numFiles; i++ {
		_, err = c.PutFile(repo, "master", fmt.Sprintf("files/%d", i), strings.NewReader(fmt.Sprintf("foo %d\n", i)))
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(repo, "master"))
	_, err = c.StartCommit(repo, "other")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(repo, "master", "files", repo, "other", "files", false))
	require.NoError(t, c.CopyFile(repo, "master", "files/0", repo, "other", "file0", false))
	require.NoError(t, c.FinishCommit(repo, "other"))
	for i := 0; i < numFiles; i++ {
		var b bytes.Buffer
		require.NoError(t, c.GetFile(repo, "other", fmt.Sprintf("files/%d", i), 0, 0, &b))
		require.Equal(t, fmt.Sprintf("foo %d\n", i), b.String())
	}
	var b bytes.Buffer
	require.NoError(t, c.GetFile(repo, "other", "file0", 0, 0, &b))
	require.Equal(t, "foo 0\n", b.String())
	_, err = c.StartCommit(repo, "other")
	require.NoError(t, c.CopyFile(repo, "other", "files/0", repo, "other", "files", true))
	require.NoError(t, c.FinishCommit(repo, "other"))
	b.Reset()
	require.NoError(t, c.GetFile(repo, "other", "files", 0, 0, &b))
	require.Equal(t, "foo 0\n", b.String())
}

func TestBuildCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getClient(t)
	repo := uniqueString("TestBuildCommit")
	require.NoError(t, c.CreateRepo(repo))

	tree1 := hashtree.NewHashTree()
	fooObj, fooSize, err := c.PutObject(strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, tree1.PutFile("foo", []*pfs.Object{fooObj}, fooSize))
	tree1Finish, err := tree1.Finish()
	require.NoError(t, err)
	serialized, err := hashtree.Serialize(tree1Finish)
	require.NoError(t, err)
	tree1Obj, _, err := c.PutObject(bytes.NewReader(serialized))
	_, err = c.BuildCommit(repo, "master", "", tree1Obj.Hash)
	require.NoError(t, err)
	repoInfo, err := c.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, uint64(fooSize), repoInfo.SizeBytes)
	commitInfo, err := c.InspectCommit(repo, "master")
	require.NoError(t, err)
	require.Equal(t, uint64(fooSize), commitInfo.SizeBytes)

	barObj, barSize, err := c.PutObject(strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, tree1.PutFile("bar", []*pfs.Object{barObj}, barSize))
	tree2Finish, err := tree1.Finish()
	require.NoError(t, err)
	serialized, err = hashtree.Serialize(tree2Finish)
	require.NoError(t, err)
	tree2Obj, _, err := c.PutObject(bytes.NewReader(serialized))
	_, err = c.BuildCommit(repo, "master", "", tree2Obj.Hash)
	require.NoError(t, err)
	repoInfo, err = c.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, uint64(fooSize+barSize), repoInfo.SizeBytes)
	commitInfo, err = c.InspectCommit(repo, "master")
	require.NoError(t, err)
	require.Equal(t, uint64(fooSize+barSize), commitInfo.SizeBytes)
}

func TestPropagateCommit(t *testing.T) {
	c := getClient(t)
	repo1 := uniqueString("TestPropagateCommit1")
	require.NoError(t, c.CreateRepo(repo1))
	repo2 := uniqueString("TestPropagateCommit2")
	require.NoError(t, c.CreateRepo(repo2))
	require.NoError(t, c.CreateBranch(repo2, "master", "", []*pfs.Branch{pclient.NewBranch(repo1, "master")}))
	commit, err := c.StartCommit(repo1, "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo1, commit.ID))
	commits, err := c.ListCommitByRepo(repo2)
	require.NoError(t, err)
	require.Equal(t, 1, len(commits))
}

// TestBackfillBranch implements the following DAG:
//
// A---->C
//  \   ^
//   \ /
//    X
//   / \
//  /   v
// B---->D
func TestBackfillBranch(t *testing.T) {
	c := getClient(t)
	require.NoError(t, c.CreateRepo("A"))
	require.NoError(t, c.CreateRepo("B"))
	require.NoError(t, c.CreateRepo("C"))
	require.NoError(t, c.CreateRepo("D"))
	require.NoError(t, c.CreateBranch("C", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master"), pclient.NewBranch("B", "master")}))
	_, err := c.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit("A", "master"))
	require.NoError(t, c.FinishCommit("C", "master"))
	_, err = c.StartCommit("B", "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit("B", "master"))
	require.NoError(t, c.FinishCommit("C", "master"))
	commits, err := c.ListCommitByRepo("C")
	require.NoError(t, err)
	require.Equal(t, 2, len(commits))

	// Create a branch in D, it should receive a single commit for the heads of `A` and `B`.
	require.NoError(t, c.CreateBranch("D", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master"), pclient.NewBranch("B", "master")}))
	commits, err = c.ListCommitByRepo("D")
	require.NoError(t, err)
	require.Equal(t, 1, len(commits))
}

// TestUpdateBranch tests the following DAG:
//
// A───>B───>C
//
// Then updates it to:
//
// A───>B───>C
//      ^
// D────┘
//
func TestUpdateBranch(t *testing.T) {
	c := getClient(t)
	require.NoError(t, c.CreateRepo("A"))
	require.NoError(t, c.CreateRepo("B"))
	require.NoError(t, c.CreateRepo("C"))
	require.NoError(t, c.CreateRepo("D"))
	require.NoError(t, c.CreateBranch("B", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master")}))
	require.NoError(t, c.CreateBranch("C", "master", "", []*pfs.Branch{pclient.NewBranch("B", "master")}))
	_, err := c.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit("A", "master"))
	require.NoError(t, c.FinishCommit("B", "master"))
	require.NoError(t, c.FinishCommit("C", "master"))

	_, err = c.StartCommit("D", "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit("D", "master"))

	require.NoError(t, c.CreateBranch("B", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master"), pclient.NewBranch("D", "master")}))
	require.NoError(t, c.FinishCommit("B", "master"))
	require.NoError(t, c.FinishCommit("C", "master"))
	cCommitInfo, err := c.InspectCommit("C", "master")
	require.NoError(t, err)
	require.Equal(t, 3, len(cCommitInfo.Provenance))
}

func TestBranchProvenance(t *testing.T) {
	c := getClient(t)
	tests := [][]struct {
		name       string
		directProv []string
		err        bool
		expectProv map[string][]string
		expectSubv map[string][]string
	}{{
		{name: "A"},
		{name: "B", directProv: []string{"A"}},
		{name: "C", directProv: []string{"B"}},
		{name: "D", directProv: []string{"C", "A"},
			expectProv: map[string][]string{"A": nil, "B": {"A"}, "C": {"B", "A"}, "D": {"A", "B", "C"}},
			expectSubv: map[string][]string{"A": {"B", "C", "D"}, "B": {"C", "D"}, "C": {"D"}, "D": {}}},
		// A ─> B ─> C ─> D
		// └──────────────⬏
		{name: "B",
			expectProv: map[string][]string{"A": {}, "B": {}, "C": {"B"}, "D": {"A", "B", "C"}},
			expectSubv: map[string][]string{"A": {"D"}, "B": {"C", "D"}, "C": {"D"}, "D": {}}},
		// A    B ─> C ─> D
		// └──────────────⬏
	}, {
		{name: "A"},
		{name: "B", directProv: []string{"A"}},
		{name: "C", directProv: []string{"A", "B"}},
		{name: "D", directProv: []string{"C"},
			expectProv: map[string][]string{"A": {}, "B": {"A"}, "C": {"A", "B"}, "D": {"A", "B", "C"}},
			expectSubv: map[string][]string{"A": {"B", "C", "D"}, "B": {"C", "D"}, "C": {"D"}, "D": {}}},
		// A ─> B ─> C ─> D
		// └─────────⬏
		{name: "C", directProv: []string{"B"},
			expectProv: map[string][]string{"A": {}, "B": {"A"}, "C": {"A", "B"}, "D": {"A", "B", "C"}},
			expectSubv: map[string][]string{"A": {"B", "C", "D"}, "B": {"C", "D"}, "C": {"D"}, "D": {}}},
		// A -> B -> C -> D
	}, {
		{name: "A"},
		{name: "B"},
		{name: "C", directProv: []string{"A", "B"}},
		{name: "D", directProv: []string{"C"}},
		{name: "E", directProv: []string{"A", "D"},
			expectProv: map[string][]string{"A": {}, "B": {}, "C": {"A", "B"}, "D": {"A", "B", "C"}, "E": {"A", "B", "C", "D"}},
			expectSubv: map[string][]string{"A": {"C", "D", "E"}, "B": {"C", "D", "E"}, "C": {"D", "E"}, "D": {"E"}, "E": {}}},
		// A    B ─> C ─> D ─> E
		// └─────────⬏         ^
		// └───────────────────┘
		{name: "C", directProv: []string{"B"},
			expectProv: map[string][]string{"A": {}, "B": {}, "C": {"B"}, "D": {"B", "C"}, "E": {"A", "B", "C", "D"}},
			expectSubv: map[string][]string{"A": {"E"}, "B": {"C", "D", "E"}, "C": {"D", "E"}, "D": {"E"}, "E": {}}},
		// A    B ─> C ─> D ─> E
		// └───────────────────⬏
	}, {
		{name: "A", directProv: []string{"A"}, err: true},
		{name: "A"},
		{name: "A", directProv: []string{"A"}, err: true},
		{name: "B", directProv: []string{"A"}},
		{name: "A", directProv: []string{"B"}, err: true},
	},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			repo := uniqueString("repo")
			require.NoError(t, c.CreateRepo(repo))
			for iStep, step := range test {
				var provenance []*pfs.Branch
				for _, branch := range step.directProv {
					provenance = append(provenance, pclient.NewBranch(repo, branch))
				}
				err := c.CreateBranch(repo, step.name, "", provenance)
				if step.err {
					require.YesError(t, err, "%d> CreateBranch(\"%s\", %v)", iStep, step.name, step.directProv)
				} else {
					require.NoError(t, err, "%d> CreateBranch(\"%s\", %v)", iStep, step.name, step.directProv)
				}
				for branch, expectedProv := range step.expectProv {
					bi, err := c.InspectBranch(repo, branch)
					require.NoError(t, err)
					sort.Strings(expectedProv)
					require.Equal(t, len(expectedProv), len(bi.Provenance))
					for _, b := range bi.Provenance {
						i := sort.SearchStrings(expectedProv, b.Name)
						if i >= len(expectedProv) || expectedProv[i] != b.Name {
							t.Fatalf("provenance for %s contains: %s, but should only contain: %v", branch, b.Name, expectedProv)
						}
					}
				}
				for branch, expectedSubv := range step.expectSubv {
					bi, err := c.InspectBranch(repo, branch)
					require.NoError(t, err)
					sort.Strings(expectedSubv)
					require.Equal(t, len(expectedSubv), len(bi.Subvenance))
					for _, b := range bi.Subvenance {
						i := sort.SearchStrings(expectedSubv, b.Name)
						if i >= len(expectedSubv) || expectedSubv[i] != b.Name {
							t.Fatalf("subvenance for %s contains: %s, but should only contain: %v", branch, b.Name, expectedSubv)
						}
					}
				}
			}
		})
	}
	// t.Run("1", func(t *testing.T) {
	// 	c := getClient(t)
	// 	require.NoError(t, c.CreateRepo("A"))
	// 	require.NoError(t, c.CreateRepo("B"))
	// 	require.NoError(t, c.CreateRepo("C"))
	// 	require.NoError(t, c.CreateRepo("D"))

	// 	require.NoError(t, c.CreateBranch("B", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master")}))
	// 	require.NoError(t, c.CreateBranch("C", "master", "", []*pfs.Branch{pclient.NewBranch("B", "master")}))
	// 	require.NoError(t, c.CreateBranch("D", "master", "", []*pfs.Branch{pclient.NewBranch("C", "master"), pclient.NewBranch("A", "master")}))

	// 	aMaster, err := c.InspectBranch("A", "master")
	// 	require.NoError(t, err)
	// 	require.Equal(t, 3, len(aMaster.Subvenance))

	// 	cMaster, err := c.InspectBranch("C", "master")
	// 	require.NoError(t, err)
	// 	require.Equal(t, 2, len(cMaster.Provenance))

	// 	dMaster, err := c.InspectBranch("D", "master")
	// 	require.NoError(t, err)
	// 	require.Equal(t, 3, len(dMaster.Provenance))

	// 	require.NoError(t, c.CreateBranch("B", "master", "", nil))

	// 	aMaster, err = c.InspectBranch("A", "master")
	// 	require.NoError(t, err)
	// 	require.Equal(t, 1, len(aMaster.Subvenance))

	// 	cMaster, err = c.InspectBranch("C", "master")
	// 	require.NoError(t, err)
	// 	require.Equal(t, 1, len(cMaster.Provenance))

	// 	dMaster, err = c.InspectBranch("D", "master")
	// 	require.NoError(t, err)
	// 	require.Equal(t, 3, len(dMaster.Provenance))
	// })
	// t.Run("2", func(t *testing.T) {
	// 	c := getClient(t)
	// 	require.NoError(t, c.CreateRepo("A"))
	// 	require.NoError(t, c.CreateRepo("B"))
	// 	require.NoError(t, c.CreateRepo("C"))
	// 	require.NoError(t, c.CreateRepo("D"))

	// 	require.NoError(t, c.CreateBranch("B", "master", "", []*pfs.Branch{pclient.NewBranch("A", "master")}))
	// 	require.NoError(t, c.CreateBranch("C", "master", "", []*pfs.Branch{pclient.NewBranch("B", "master"), pclient.NewBranch("A", "master")}))
	// 	require.NoError(t, c.CreateBranch("D", "master", "", []*pfs.Branch{pclient.NewBranch("C", "master")}))
	// })
}

func TestChildCommits(t *testing.T) {
	c := getClient(t)
	require.NoError(t, c.CreateRepo("A"))
	require.NoError(t, c.CreateBranch("A", "master", "", nil))

	// Small helper function wrapping c.InspectCommit, because it's called a lot
	inspect := func(repo, commit string) *pfs.CommitInfo {
		commitInfo, err := c.InspectCommit(repo, commit)
		require.NoError(t, err)
		return commitInfo
	}

	commit1, err := c.StartCommit("A", "master")
	require.NoError(t, err)
	commits, err := c.ListCommit("A", "master", "", 0)
	t.Logf("%v", commits)
	require.NoError(t, c.FinishCommit("A", "master"))

	commit2, err := c.StartCommit("A", "master")
	require.NoError(t, err)

	// Inspect commit 1 and 2
	commit1Info, commit2Info := inspect("A", commit1.ID), inspect("A", commit2.ID)
	require.Equal(t, commit1.ID, commit2Info.ParentCommit.ID)
	require.Equal(t, len(commit1Info.ChildCommits), 1)
	require.Equal(t, commit2.ID, commit1Info.ChildCommits[0].ID)

	// Delete commit 2 and make sure it's removed from commit1.ChildCommits
	require.NoError(t, c.DeleteCommit("A", commit2.ID))
	require.Equal(t, len(inspect("A", commit1.ID).ChildCommits), 0)

	// Re-create commit2, and create a third commit also extending from commit1.
	// Make sure both appear in commit1.children
	commit2, err = c.StartCommit("A", "master")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit("A", commit2.ID))
	commit3, err := c.PfsAPIClient.StartCommit(c.Ctx(), &pfs.StartCommitRequest{
		Parent: pclient.NewCommit("A", commit1.ID),
	})
	require.NoError(t, err)
	commit1Info = inspect("A", commit1.ID)
	require.Equal(t, len(commit1Info.ChildCommits), 2)
	require.OneOfEquals(t, commit2, commit1Info.ChildCommits)
	require.OneOfEquals(t, commit3, commit1Info.ChildCommits)

	// Delete commit3 and make sure commit1 has the right children
	require.NoError(t, c.DeleteCommit("A", commit3.ID))
	commit1Info = inspect("A", commit1.ID)
	require.Equal(t, len(commit1Info.ChildCommits), 1)
	require.Equal(t, commit2.ID, commit1Info.ChildCommits[0].ID)

	// Create a downstream branch in the same repo, then commit to "A" and make
	// sure the new HEAD commit is in teh parent's children (i.e. test
	// propagateCommit)
	require.NoError(t, c.CreateBranch("A", "out", "", []*pfs.Branch{
		pclient.NewBranch("A", "master"),
	}))
	outCommit1ID := inspect("A", "out").Commit.ID
	commit3, err = c.StartCommit("A", "master")
	require.NoError(t, err)
	c.FinishCommit("A", commit3.ID)
	// Re-inspect outCommit1, which has been updated by StartCommit
	outCommit1, outCommit2 := inspect("A", outCommit1ID), inspect("A", "out")
	require.Equal(t, outCommit1.Commit.ID, outCommit2.ParentCommit.ID)
	require.Equal(t, 1, len(outCommit1.ChildCommits))
	require.Equal(t, outCommit2.Commit.ID, outCommit1.ChildCommits[0].ID)

	// create a new branch in a different repo and do the same test again
	require.NoError(t, c.CreateRepo("B"))
	require.NoError(t, c.CreateBranch("B", "master", "", []*pfs.Branch{
		pclient.NewBranch("A", "master"),
	}))
	bCommit1ID := inspect("B", "master").Commit.ID
	commit3, err = c.StartCommit("A", "master")
	require.NoError(t, err)
	c.FinishCommit("A", commit3.ID)
	// Re-inspect bCommit1, which has been updated by StartCommit
	bCommit1, bCommit2 := inspect("B", bCommit1ID), inspect("B", "master")
	require.Equal(t, bCommit1.Commit.ID, bCommit2.ParentCommit.ID)
	require.Equal(t, 1, len(bCommit1.ChildCommits))
	require.Equal(t, bCommit2.Commit.ID, bCommit1.ChildCommits[0].ID)

	// create a new branch in a different repo, then update it so that two commits
	// are generated. Make sure the second commit is in the parent's children
	require.NoError(t, c.CreateRepo("C"))
	require.NoError(t, c.CreateBranch("C", "master", "", []*pfs.Branch{
		pclient.NewBranch("A", "master"),
	}))
	cCommit1ID := inspect("C", "master").Commit.ID // Get new commit's ID
	require.NoError(t, c.CreateBranch("C", "master", "master", []*pfs.Branch{
		pclient.NewBranch("A", "master"),
		pclient.NewBranch("B", "master"),
	}))
	// Re-inspect cCommit1, which has been updated by CreateBranch
	cCommit1, cCommit2 := inspect("C", cCommit1ID), inspect("C", "master")
	require.Equal(t, cCommit1.Commit.ID, cCommit2.ParentCommit.ID)
	require.Equal(t, 1, len(cCommit1.ChildCommits))
	require.Equal(t, cCommit2.Commit.ID, cCommit1.ChildCommits[0].ID)
}

func uniqueString(prefix string) string {
	return prefix + "-" + uuid.NewWithoutDashes()[0:12]
}
