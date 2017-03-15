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

	"github.com/gogo/protobuf/types"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
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
		require.Equal(t, repoNames[len(repoNames)-i-1], repoInfo.Repo.Name)
	}

	require.Equal(t, len(repoInfos), numRepos)
}

// Make sure that commits of deleted repos do not resurface
func TestCreateDeletedRepo(t *testing.T) {
	t.Parallel()
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

func TestInspectCommit(t *testing.T) {
	t.Parallel()
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

func TestDeleteCommitFuture(t *testing.T) {
	// For when DeleteCommit gets implemented
	t.Skip()

	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
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

func TestCleanPath(t *testing.T) {
	t.Parallel()
	c := getClient(t)
	repo := "TestCleanPath"
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = c.PutFile(repo, commit.ID, "./././file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	_, err = c.InspectFile(repo, commit.ID, "file")
	require.NoError(t, err)
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

func TestStartCommitWithUnfinishedParent(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.StartCommit(repo, commit1.ID)
	// fails because the parent commit has not been finished
	require.YesError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	_, err = client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
}

func TestProvenance2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	require.NoError(t, client.CreateRepo("E"))
	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("B"),
		Provenance: []*pfs.Repo{pclient.NewRepo("A")},
	})
	require.NoError(t, err)
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("C"),
		Provenance: []*pfs.Repo{pclient.NewRepo("B"), pclient.NewRepo("E")},
	})
	_, err = client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("D"),
		Provenance: []*pfs.Repo{pclient.NewRepo("C")},
	})

	ACommit, err := client.StartCommit("A", "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit.ID))
	ECommit, err := client.StartCommit("E", "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("E", ECommit.ID))
	BCommit, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("B", ""),
			Provenance: []*pfs.Commit{ACommit},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("B", BCommit.ID))
	commitInfo, err := client.InspectCommit("B", BCommit.ID)
	require.NoError(t, err)

	CCommit, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("C", ""),
			Provenance: []*pfs.Commit{BCommit, ECommit},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("C", CCommit.ID))
	commitInfo, err = client.InspectCommit("C", CCommit.ID)
	require.NoError(t, err)

	DCommit, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("D", ""),
			Provenance: []*pfs.Commit{CCommit},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("D", DCommit.ID))

	commitInfo, err = client.InspectCommit("D", DCommit.ID)
	require.NoError(t, err)
	for _, commit := range commitInfo.Provenance {
		require.EqualOneOf(t, []interface{}{ACommit, ECommit, BCommit, CCommit}, commit)
	}
}

func TestSimple(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "")
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

func TestBranch1(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"

	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit.ID, "master"))
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

func TestPutFile2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit1.ID, "master"))
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

	commit3, err := client.StartCommit(repo, commit2.ID)
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

func TestInspectFile(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent1 := "foo\n"
	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "foo", strings.NewReader(fileContent1))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfo, err := client.InspectFile(repo, commit1.ID, "foo")
	require.NoError(t, err)
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
	require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))

	fileContent2 := "barbar\n"
	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "foo", strings.NewReader(fileContent2))
	require.NoError(t, err)
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
	commit3, err := client.StartCommit(repo, commit2.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit3.ID, "bar", strings.NewReader(fileContent3))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	fileInfos, err := client.ListFile(repo, commit3.ID, "")
	require.NoError(t, err)
	require.Equal(t, len(fileInfos), 2)
}

func TestInspectFile2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent1 := "foo\n"
	fileContent2 := "buzz\n"

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit.ID, "master"))
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
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	fileContent := "foo\n"
	_, err = client.PutFile(repo, commit1.ID, "dir/foo", strings.NewReader(fileContent))
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	fileInfo, err := client.InspectFile(repo, commit1.ID, "dir/foo")
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
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit.ID, "master"))
	_, err = client.PutFile(repo, "master", "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/2", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master"))

	fileInfo, err := client.InspectFile(repo, "master", "/dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(fileInfo.Children))
	require.Equal(t, "/dir", fileInfo.File.Path)
	require.Equal(t, pfs.FileType_DIR, fileInfo.FileType)

	_, err = client.StartCommit(repo, "master")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "master", "dir/3", strings.NewReader(fileContent))
	require.NoError(t, err)
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

func TestListFile2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit.ID, "master"))
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
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit.ID, "master"))
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
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "dir/1", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, commit1.ID)
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit2.ID, "dir", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.YesError(t, client.FinishCommit(repo, commit2.ID))
}

func TestRootDirectory(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	fileContent := "foo\n"

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit.ID, "foo", strings.NewReader(fileContent))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, commit.ID))

	fileInfos, err := client.ListFile(repo, commit.ID, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(fileInfos))
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
	require.NoError(t, client.SetBranch(repo, commit1.ID, "master"))

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

func TestListCommit(t *testing.T) {
	client := getClient(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	numCommits := 10

	var commit *pfs.Commit
	var err error
	var parentID string
	var midCommitID string
	for i := 0; i < numCommits; i++ {
		commit, err = client.StartCommit(repo, parentID)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, commit.ID))
		parentID = commit.ID
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
	commitInfos, err = client.ListCommit(repo, commit.ID, "", 0)
	require.NoError(t, err)
	require.Equal(t, numCommits, len(commitInfos))

	// Test that commits are sorted in newest-first order
	for i := 0; i < len(commitInfos)-1; i++ {
		require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
	}

	// Now list all commits up to the mid commit, excluding the mid commit
	// itself
	commitInfos, err = client.ListCommit(repo, commit.ID, midCommitID, 0)
	require.NoError(t, err)
	require.Equal(t, numCommits-numCommits/2-1, len(commitInfos))

	// Test that commits are sorted in newest-first order
	for i := 0; i < len(commitInfos)-1; i++ {
		require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
	}

	// list commits by branch
	require.NoError(t, client.SetBranch(repo, commit.ID, "master"))
	commitInfos, err = client.ListCommit(repo, "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, numCommits, len(commitInfos))

	// Test that commits are sorted in newest-first order
	for i := 0; i < len(commitInfos)-1; i++ {
		require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
	}
}

func TestOffsetRead(t *testing.T) {
	t.Parallel()
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
	require.NoError(t, client.FinishCommit(repo, commit.ID))
	var buffer bytes.Buffer
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

	var commits []*pfs.Commit
	var parentID string
	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit(repo, parentID)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit(repo, commit.ID))
		parentID = commit.ID
		commits = append(commits, commit)
	}

	client.SetBranch(repo, commits[len(commits)-1].ID, "master")

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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	w, err := client.PutFileSplitWriter(repo, commit.ID, "foo", pfs.Delimiter_NONE, 0, 0)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.NoError(t, client.FinishCommit(repo, commit.ID))
	_, err = client.InspectFile(repo, commit.ID, "foo")
	require.NoError(t, err)
}

func TestGetFileInvalidCommit(t *testing.T) {
	t.Parallel()
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

func TestATonOfPuts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long tests in short mode")
	}
	t.Parallel()
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
	numObjs := 5000
	numGoros := 100
	var expectedOutput []byte
	var wg sync.WaitGroup
	putFileStarted := time.Now()
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
	putFileFinished := time.Now()

	finishCommitStarted := time.Now()
	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	finishCommitFinished := time.Now()

	var buffer bytes.Buffer
	require.NoError(t, client.GetFile(repo, commit1.ID, "foo", 0, 0, &buffer))
	require.Equal(t, string(expectedOutput), buffer.String())
	fmt.Printf("PutFile took: %s\n", putFileFinished.Sub(putFileStarted))
	fmt.Printf("FinishCommit took: %s\n", finishCommitFinished.Sub(finishCommitStarted))
}

func TestPutFileNullCharacter(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "")
	require.NoError(t, err)

	_, err = client.PutFile(repo, commit.ID, "foo\x00bar", strings.NewReader("foobar\n"))
	// null characters error because when you `ls` files with null characters
	// they truncate things after the null character leading to strange results
	require.YesError(t, err)
}

func TestPutFileURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getClient(t)
	repo := "TestPutFileURL"
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, c.PutFileURL(repo, commit.ID, "readme", "https://raw.githubusercontent.com/pachyderm/pachyderm/master/README.md", false))
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	fileInfo, err := c.InspectFile(repo, commit.ID, "readme")
	require.NoError(t, err)
	require.True(t, fileInfo.SizeBytes > 0)
}

func TestBigListFile(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	client := getClient(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, client.SetBranch(repo, commit1.ID, "master"))
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
	t.Parallel()
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
				close(ready)
			},
			grpcutil.ServeOptions{
				Version:    version.Version,
				MaxMsgSize: pclient.MaxMsgSize,
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
	c, err := pclient.NewFromAddress(addresses[0])
	require.NoError(t, err)
	return *c
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
	ACommit, err := client.StartCommit("A", "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit.ID))

	// do the other commits in a goro so we can block for them
	go func() {
		BCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfs.StartCommitRequest{
				Parent:     pclient.NewCommit("B", ""),
				Provenance: []*pfs.Commit{ACommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("B", BCommit.ID))
		CCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfs.StartCommitRequest{
				Parent:     pclient.NewCommit("C", ""),
				Provenance: []*pfs.Commit{BCommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("C", CCommit.ID))
		DCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfs.StartCommitRequest{
				Parent:     pclient.NewCommit("D", ""),
				Provenance: []*pfs.Commit{BCommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("D", DCommit.ID))
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

func TestFlush2(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	require.NoError(t, client.CreateRepo("B"))
	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("C"),
		Provenance: []*pfs.Repo{pclient.NewRepo("A"), pclient.NewRepo("B")},
	})
	require.NoError(t, err)

	ACommit, err := client.StartCommit("A", "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("A", ACommit.ID))
	BCommit, err := client.StartCommit("B", "")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("B", BCommit.ID))
	CCommit, err := client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("C", ""),
			Provenance: []*pfs.Commit{ACommit, BCommit},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("C", CCommit.ID))

	BCommit, err = client.StartCommit("B", BCommit.ID)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("B", BCommit.ID))
	CCommit, err = client.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent:     pclient.NewCommit("C", CCommit.ID),
			Provenance: []*pfs.Commit{ACommit, BCommit},
		},
	)
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit("C", CCommit.ID))

	commitIter, err := client.FlushCommit([]*pfs.Commit{pclient.NewCommit("B", BCommit.ID), pclient.NewCommit("A", ACommit.ID)}, nil)
	require.NoError(t, err)
	commitInfos, err := collectCommitInfos(commitIter)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))

	require.Equal(t, commitInfos[0].Commit.Repo.Name, "C")
	require.Equal(t, commitInfos[0].Commit.ID, CCommit.ID)
}

func TestFlushCommitWithNoDownstreamRepos(t *testing.T) {
	t.Parallel()
	c := getClient(t)
	repo := "test"
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "")
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{pclient.NewCommit(repo, commit.ID)}, nil)
	require.NoError(t, err)
	commitInfos, err := collectCommitInfos(commitIter)
	require.NoError(t, err)
	require.Equal(t, 0, len(commitInfos))
}

func TestFlushOpenCommit(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.NoError(t, client.CreateRepo("A"))
	_, err := client.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
		Repo:       pclient.NewRepo("B"),
		Provenance: []*pfs.Repo{pclient.NewRepo("A")},
	})
	ACommit, err := client.StartCommit("A", "")
	require.NoError(t, err)

	// do the other commits in a goro so we can block for them
	go func() {
		time.Sleep(5 * time.Second)
		require.NoError(t, client.FinishCommit("A", ACommit.ID))
		BCommit, err := client.PfsAPIClient.StartCommit(
			context.Background(),
			&pfs.StartCommitRequest{
				Parent:     pclient.NewCommit("B", ""),
				Provenance: []*pfs.Commit{ACommit},
			},
		)
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("B", BCommit.ID))
	}()

	// Flush ACommit
	commitIter, err := client.FlushCommit([]*pfs.Commit{pclient.NewCommit("A", ACommit.ID)}, nil)
	require.NoError(t, err)
	commitInfos, err := collectCommitInfos(commitIter)
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
}

func TestEmptyFlush(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	commitIter, err := client.FlushCommit(nil, nil)
	require.NoError(t, err)
	_, err = collectCommitInfos(commitIter)
	require.YesError(t, err)
}

func TestFlushNonExistentCommit(t *testing.T) {
	t.Parallel()
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
	t.Parallel()

	c := getClient(t)
	// create repos
	repo := uniqueString("TestPutFileSplit")
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, "")
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "none", pfs.Delimiter_NONE, 0, 0, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "line2", pfs.Delimiter_LINE, 2, 0, strings.NewReader("foo\nbar\nbuz\nfiz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "line3", pfs.Delimiter_LINE, 0, 8, strings.NewReader("foo\nbar\nbuz\nfiz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "json", pfs.Delimiter_JSON, 0, 0, strings.NewReader("{}{}{}{}{}{}{}{}{}{}"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "json", pfs.Delimiter_JSON, 0, 0, strings.NewReader("{}{}{}{}{}{}{}{}{}{}"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "json2", pfs.Delimiter_JSON, 2, 0, strings.NewReader("{}{}{}{}"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit.ID, "json3", pfs.Delimiter_JSON, 0, 4, strings.NewReader("{}{}{}{}"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	commit2, err := c.StartCommit(repo, commit.ID)
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit2.ID, "line", pfs.Delimiter_LINE, 0, 0, strings.NewReader("foo\nbar\nbuz\n"))
	require.NoError(t, err)
	_, err = c.PutFileSplit(repo, commit2.ID, "json", pfs.Delimiter_JSON, 0, 0, strings.NewReader("{}{}{}{}{}{}{}{}{}{}"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(repo, commit2.ID))
	fileInfo, err := c.InspectFile(repo, commit.ID, "none")
	require.NoError(t, err)
	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
	files, err := c.ListFile(repo, commit.ID, "line")
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

func uniqueString(prefix string) string {
	return prefix + "-" + uuid.NewWithoutDashes()[0:12]
}
