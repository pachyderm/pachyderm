package server

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestPFSRefactorBranchSimpleRF(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "branchA")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))
	heads, err := client.ListBranch(repo)
	require.NoError(t, err)

	require.Equal(t, 1, len(heads))
	require.Equal(t, "branchA", heads[0].Branch)

}

func TestPFSRefactorListBranchRF(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "branchA")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	// Don't specify a branch because we already should have it from parent
	commit2, err := client.StartCommit(repo, commit1.ID, "")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	// Specify branch, because branching off of commit1
	commit3, err := client.StartCommit(repo, commit1.ID, "branchB")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit3.ID))
	branches, err := client.ListBranch(repo)
	require.NoError(t, err)

	require.Equal(t, 3, len(branches))
	branchNames := []interface{}{
		branches[0].Branch,
		branches[1].Branch,
		branches[2].Branch,
	}

	require.EqualOneOf(t, branchNames, "branchA")
	require.EqualOneOf(t, branchNames, "branchB")
}

func TestPFSRefactorListCommitBasicRF(t *testing.T) {
	t.Parallel()
	client, server := getClientAndServer(t)

	require.NoError(t, client.CreateRepo("test"))
	numCommits := 10
	var commitIDs []string
	for i := 0; i < numCommits; i++ {
		commit, err := client.StartCommit("test", "", "")
		require.NoError(t, err)
		require.NoError(t, client.FinishCommit("test", commit.ID))
		commitIDs = append(commitIDs, commit.ID)
	}

	test := func() {
		commitInfos, err := client.ListCommit(
			[]string{"test"},
			nil,
			pclient.CommitTypeNone,
			false,
			false,
			nil,
		)
		require.NoError(t, err)

		for i, commitInfo := range commitInfos {
			require.Equal(t, commitIDs[len(commitIDs)-i-1], commitInfo.Commit.ID)
		}

		require.Equal(t, len(commitInfos), numCommits)
	}

	test()

	restartServer(server, t)

	test()
}

func TestPFSRefactorStartAndFinishCommitRF(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))
}

func TestPFSRefactorInspectCommitBasicRF(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	started := time.Now()
	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit, commitInfo.Commit)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_WRITE, commitInfo.CommitType)
	require.Equal(t, 0, int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.Nil(t, commitInfo.Finished)

	require.NoError(t, client.FinishCommit(repo, commit.ID))
	finished := time.Now()

	commitInfo, err = client.InspectCommit(repo, commit.ID)
	require.NoError(t, err)

	require.Equal(t, commit.ID, commitInfo.Commit.ID)
	require.Equal(t, pfsclient.CommitType_COMMIT_TYPE_READ, commitInfo.CommitType)
	require.Equal(t, 0, int(commitInfo.SizeBytes))
	require.True(t, started.Before(commitInfo.Started.GoTime()))
	require.True(t, finished.After(commitInfo.Finished.GoTime()))
}

func TestPFSRefactorStartCommitFromParentIDRF(t *testing.T) {
	t.Parallel()

	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit, err := client.StartCommit(repo, "", "")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit.ID))

	branches, err := client.ListBranch(repo)
	require.NoError(t, err)

	require.Equal(t, 1, len(branches))

	// Should create commit off of parent on a new branch
	commit1, err := client.StartCommit(repo, commit.ID, "")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))
	// Now check to make sure that commit is on a new branch w a random name
	// This imitates the existing PFS behavior
	existingBranch := branches[0].Branch
	branches2, err := client.ListBranch(repo)
	require.NoError(t, err)

	uniqueBranches := make(map[string]bool)

	for _, thisBranch := range branches2 {
		uniqueBranches[thisBranch.Branch] = true
	}

	require.Equal(t, 2, len(uniqueBranches))
	delete(uniqueBranches, existingBranch)
	require.Equal(t, 1, len(uniqueBranches))
	var existingBranch2 string
	for name, _ := range uniqueBranches {
		existingBranch2 = name
	}

	// Should create commit off of parent on a new branch by name
	commit2, err := client.StartCommit(repo, commit.ID, "foo")
	require.NoError(t, err)

	branches3, err := client.ListBranch(repo)
	require.NoError(t, err)

	uniqueBranches = make(map[string]bool)

	for _, thisBranch := range branches3 {
		uniqueBranches[thisBranch.Branch] = true
	}

	require.Equal(t, 3, len(uniqueBranches))
	delete(uniqueBranches, existingBranch)
	require.Equal(t, 2, len(uniqueBranches))
	delete(uniqueBranches, existingBranch2)
	require.Equal(t, 1, len(uniqueBranches))

	require.NoError(t, client.FinishCommit(repo, commit2.ID))
}

func TestPFSRefactorInspectRepoMostBasicRF(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	repoInfo, err := client.InspectRepo(repo)
	require.NoError(t, err)

	require.Equal(t, int(repoInfo.SizeBytes), 0)
}

func TestPFSRefactorStartCommitLatestOnBranchRF(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "branchA")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	commit2, err := client.StartCommit(repo, "", "branchA")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit2.ID))

	commit3, err := client.StartCommit(repo, "", "branchA")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit3.ID))

	branches, err := client.ListBranch(repo)
	require.Equal(t, 1, len(branches))
	require.Equal(t, commit3.ID, branches[0].Commit.ID)
}

func TestPFSRefactorListBranchRedundantRF(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "branchA")
	require.NoError(t, err)

	require.NoError(t, client.FinishCommit(repo, commit1.ID))

	// Can't create branch if it exists
	_, err = client.StartCommit(repo, commit1.ID, "branchA")
	require.YesError(t, err)

	branches, err := client.ListBranch(repo)
	require.NoError(t, err)

	require.Equal(t, 1, len(branches))
	require.Equal(t, "branchA", branches[0].Branch)
}

func TestNEWAPIStartCommitFromBranchRF(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	_, err := client.StartCommit(repo, "", "master")
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(repo, "master/0")
	require.NoError(t, err)
	require.Equal(t, "master", commitInfo.Branch)
	require.Equal(t, "test", commitInfo.Commit.Repo.Name)

	require.NoError(t, client.FinishCommit(repo, "master/0"))
}

func TestNEWAPIStartCommitNewBranchRF(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)

	repo := "test"
	require.NoError(t, client.CreateRepo(repo))

	commit1, err := client.StartCommit(repo, "", "master")
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/0"))

	_, err = client.StartCommit(repo, commit1.ID, "foo")
	require.NoError(t, err)

	commitInfo, err := client.InspectCommit(repo, "foo/0")
	require.NoError(t, err)
	require.Equal(t, "foo", commitInfo.Branch)
	require.Equal(t, "test", commitInfo.Commit.Repo.Name)
}

func TestNEWAPIPutFile(t *testing.T) {
	t.Parallel()
	client, _ := getClientAndServer(t)
	repo := "test"
	require.NoError(t, client.CreateRepo(repo))
	commit1, err := client.StartCommit(repo, "", "master")
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
	require.NoError(t, client.GetFile(repo, commit1.ID, "file", 0, 0, "", nil, buffer))
	require.Equal(t, expected, buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "master/0", "file", 0, 0, "", nil, buffer))
	require.Equal(t, expected, buffer.String())

	commit2, err := client.StartCommit(repo, "", "master")
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
	require.NoError(t, client.GetFile(repo, commit2.ID, "file", 0, 0, "master/0", nil, buffer))
	require.Equal(t, expected, buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "master/1", "file", 0, 0, "master/0", nil, buffer))
	require.Equal(t, expected, buffer.String())

	_, err = client.StartCommit(repo, "master/1", "foo")
	require.NoError(t, err)
	_, err = client.PutFile(repo, "foo/0", "file", strings.NewReader("foo\nbar\nbuzz\n"))
	require.NoError(t, client.FinishCommit(repo, "foo/0"))

	expected = "foo\nbar\nbuzz\nfoo\nbar\nbuzz\nfoo\nbar\nbuzz\n"
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "foo/0", "file", 0, 0, "master/0", nil, buffer))
	require.Equal(t, expected, buffer.String())

}
