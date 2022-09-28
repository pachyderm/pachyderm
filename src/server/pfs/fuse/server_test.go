//go:build !k8s

package fuse

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

/*

Tests to write:

- read only is read only
- read write is read write
- write, unmount, files have gone from local mount
- write, unmount, remount, data comes back
- simple mount under different name works
- mount two versions of the same repo under different names works
	- there is no cache interference between them
- listing repos works and shows status
- when a repo is deleted in pachyderm it shows up as missing
- cp works (ff16dadac35bfd3f459d537fe68b788c8568db8a fixed it)

*/

func TestBasicServerSameNames(t *testing.T) {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	commit := client.NewCommit("repo", "master", "")
	err := env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = env.PachClient.PutFile(commit, "dir/file2", strings.NewReader("foo"))
	require.NoError(t, err)
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {

		resp, err := put("repos/repo/master/_mount?name=repo&mode=ro", nil)
		require.NoError(t, err)

		defer resp.Body.Close()
		repoResp := &ListRepoResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(repoResp))
		require.Equal(t, "repo", (*repoResp)["repo"].Name)
		require.Equal(t, "master", (*repoResp)["repo"].Branches["master"].Name)

		repos, err := os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		require.Equal(t, "repo", filepath.Base(repos[0].Name()))

		files, err := os.ReadDir(filepath.Join(mountPoint, "repo"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		require.Equal(t, "dir", filepath.Base(files[0].Name()))

		files, err = os.ReadDir(filepath.Join(mountPoint, "repo", "dir"))
		require.NoError(t, err)
		require.Equal(t, 2, len(files))
		require.Equal(t, "file1", filepath.Base(files[0].Name()))
		require.Equal(t, "file2", filepath.Base(files[1].Name()))

		data, err := os.ReadFile(filepath.Join(mountPoint, "repo", "dir", "file1"))
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))
	})
}

func TestBasicServerNonMasterBranch(t *testing.T) {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	commit := client.NewCommit("repo", "dev", "")
	err := env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = env.PachClient.PutFile(commit, "dir/file2", strings.NewReader("foo"))
	require.NoError(t, err)
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {

		_, err := put("repos/repo/dev/_mount?name=repo&mode=ro", nil)
		require.NoError(t, err)

		repos, err := os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		require.Equal(t, "repo", filepath.Base(repos[0].Name()))

		files, err := os.ReadDir(filepath.Join(mountPoint, "repo"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		require.Equal(t, "dir", filepath.Base(files[0].Name()))

		files, err = os.ReadDir(filepath.Join(mountPoint, "repo", "dir"))
		require.NoError(t, err)
		require.Equal(t, 2, len(files))
		require.Equal(t, "file1", filepath.Base(files[0].Name()))
		require.Equal(t, "file2", filepath.Base(files[1].Name()))

		data, err := os.ReadFile(filepath.Join(mountPoint, "repo", "dir", "file1"))
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))
	})
}

func TestBasicServerDifferingNames(t *testing.T) {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	commit := client.NewCommit("repo", "master", "")
	err := env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = env.PachClient.PutFile(commit, "dir/file2", strings.NewReader("foo"))
	require.NoError(t, err)
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {

		_, err := put("repos/repo/master/_mount?name=newname&mode=ro", nil)
		require.NoError(t, err)

		repos, err := os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		require.Equal(t, "newname", filepath.Base(repos[0].Name()))

		files, err := os.ReadDir(filepath.Join(mountPoint, "newname"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		require.Equal(t, "dir", filepath.Base(files[0].Name()))

		files, err = os.ReadDir(filepath.Join(mountPoint, "newname", "dir"))
		require.NoError(t, err)
		require.Equal(t, 2, len(files))
		require.Equal(t, "file1", filepath.Base(files[0].Name()))
		require.Equal(t, "file2", filepath.Base(files[1].Name()))

		data, err := os.ReadFile(filepath.Join(mountPoint, "newname", "dir", "file1"))
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))
	})
}

func TestUnmountAll(t *testing.T) {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

	require.NoError(t, env.PachClient.CreateRepo("repo1"))
	commit := client.NewCommit("repo1", "master", "")
	err := env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	require.NoError(t, env.PachClient.CreateRepo("repo2"))
	commit = client.NewCommit("repo2", "master", "")
	err = env.PachClient.PutFile(commit, "dir/file2", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		_, err := put("repos/repo1/master/_mount?name=repo1&mode=ro", nil)
		require.NoError(t, err)
		_, err = put("repos/repo2/master/_mount?name=repo2&mode=ro", nil)
		require.NoError(t, err)

		repos, err := os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 2, len(repos))

		resp, err := put("repos/_unmount", nil)
		require.NoError(t, err)

		defer resp.Body.Close()
		unmountResp := &ListRepoResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(unmountResp))
		require.Equal(t, 2, len(*unmountResp))

		repos, err = os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 0, len(repos))
	})
}

func TestMultipleMount(t *testing.T) {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	commit := client.NewCommit("repo", "master", "")
	err := env.PachClient.PutFile(commit, "dir/file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, env.PachClient.CreateRepo("repo2"))
	commit = client.NewCommit("repo2", "master", "")
	err = env.PachClient.PutFile(commit, "dir/file", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		_, err := put("repos/repo/master/_mount?name=mount1&mode=ro", nil)
		require.NoError(t, err)
		_, err = put("repos/repo/master/_mount?name=mount2&mode=ro", nil)
		require.NoError(t, err)

		repos, err := os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 2, len(repos))
		require.Equal(t, "mount1", filepath.Base(repos[0].Name()))
		require.Equal(t, "mount2", filepath.Base(repos[1].Name()))

		data, err := os.ReadFile(filepath.Join(mountPoint, "mount1", "dir", "file"))
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))
		data, err = os.ReadFile(filepath.Join(mountPoint, "mount2", "dir", "file"))
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))

		_, err = put("repos/repo/master/_unmount?name=mount2", nil)
		require.NoError(t, err)

		repos, err = os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		require.Equal(t, "mount1", filepath.Base(repos[0].Name()))

		data, err = os.ReadFile(filepath.Join(mountPoint, "mount1", "dir", "file"))
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))

		resp, _ := put("repos/repo2/master/_mount?name=mount1&mode=ro", nil)
		require.Equal(t, 500, resp.StatusCode)
	})
}

func TestMountNonexistentRepo(t *testing.T) {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		resp, _ := put("repos/repo1/master/_mount?name=repo1&mode=ro", nil)
		require.Equal(t, 400, resp.StatusCode)
	})
}

func TestRwUnmountCreatesCommit(t *testing.T) {
	// Unmounting a mounted read-write filesystem which has had some data
	// written to it results in a new commit with that data in it.
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	client.NewCommit("repo", "master", "")
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		resp, err := put("repos/repo/master/_mount?name=repo&mode=rw", nil)
		require.NoError(t, err)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// the commit created above isn't actually written until we unmount, so
		// we currently have 0 commits
		require.Equal(t, len(commits), 0)
		defer resp.Body.Close()
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file1"), []byte("hello"), 0644,
		)
		require.NoError(t, err)
		_, err = put("repos/repo/master/_unmount?name=repo", nil)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, len(commits), 1)
	})
}

func TestRwCommitCreatesCommit(t *testing.T) {
	// Commit operation creates a commit.
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	client.NewCommit("repo", "master", "")
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		resp, err := put("repos/repo/master/_mount?name=repo&mode=rw", nil)
		require.NoError(t, err)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// the commit created above isn't actually written until we unmount, so
		// we currently have 0 commits
		require.Equal(t, len(commits), 0)
		defer resp.Body.Close()
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file1"), []byte("hello"), 0644,
		)
		require.NoError(t, err)
		_, err = put("repos/repo/master/_commit?name=repo", nil)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, len(commits), 1)
	})
}

func TestRwCommitTwiceCreatesTwoCommits(t *testing.T) {
	// Two sequential commit operations create two commits (and they contain the
	// correct files).
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	client.NewCommit("repo", "master", "")
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		resp, err := put("repos/repo/master/_mount?name=repo&mode=rw", nil)
		require.NoError(t, err)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// the commit created above isn't actually written until we unmount, so
		// we currently have 0 commits
		require.Equal(t, len(commits), 0)
		defer resp.Body.Close()
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file1"), []byte("hello"), 0644,
		)
		require.NoError(t, err)
		_, err = put("repos/repo/master/_commit?name=repo", nil)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, len(commits), 1)

		// another file!
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file2"), []byte("hello"), 0644,
		)
		require.NoError(t, err)
		_, err = put("repos/repo/master/_commit?name=repo", nil)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, len(commits), 2)
	})
}

func TestRwCommitUnmountCreatesTwoCommits(t *testing.T) {
	// Commit and then unmount results in two commits, since unmounting creates
	// one too.
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	client.NewCommit("repo", "master", "")
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		resp, err := put("repos/repo/master/_mount?name=repo&mode=rw", nil)
		require.NoError(t, err)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// the commit created above isn't actually written until we unmount, so
		// we currently have 0 commits
		require.Equal(t, len(commits), 0)
		defer resp.Body.Close()
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file1"), []byte("hello"), 0644,
		)
		require.NoError(t, err)
		_, err = put("repos/repo/master/_commit?name=repo", nil)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, len(commits), 1)

		// another file!
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file2"), []byte("hello"), 0644,
		)
		require.NoError(t, err)
		_, err = put("repos/repo/master/_unmount?name=repo", nil)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, len(commits), 2)
	})
}

func TestHealth(t *testing.T) {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	_, err := get("health")
	require.YesError(t, err)

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		_, err = get("health")
		require.NoError(t, err)
	})
}
