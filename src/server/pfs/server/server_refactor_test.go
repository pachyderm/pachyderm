package server

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

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
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("bar"))
	require.NoError(t, err)
	_, err = client.PutFile(repo, commit1.ID, "file", strings.NewReader("buzz"))
	require.NoError(t, err)
	require.NoError(t, client.FinishCommit(repo, "master/0"))

	expected := "foo\nbar\nbuzz\n"
	buffer := &bytes.Buffer{}
	require.NoError(t, client.GetFile(repo, commit1.ID, "file", 0, 0, "", nil, buffer))
	require.Equal(t, expected, buffer.String())
	buffer.Reset()
	require.NoError(t, client.GetFile(repo, "master/0", "file", 0, 0, "", nil, buffer))
	require.Equal(t, expected, buffer.String())
}
