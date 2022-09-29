//go:build unit_test

package fuse

import (
	"bytes"
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
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo",
					Repo:   "repo",
					Branch: "master",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err := put("_mount", b)
		require.NoError(t, err)

		defer resp.Body.Close()
		mountResp := &ListMountResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mountResp))
		require.Equal(t, "repo", (*mountResp).Mounted["repo"].Repo)
		require.Equal(t, "master", (*mountResp).Mounted["repo"].Branch)

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
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo",
					Repo:   "repo",
					Branch: "dev",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		_, err := put("_mount", b)
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
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "newname",
					Repo:   "repo",
					Branch: "master",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		_, err := put("_mount", b)
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
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo1",
					Repo:   "repo1",
					Branch: "master",
				},
				{
					Name:   "repo2",
					Repo:   "repo2",
					Branch: "master",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		_, err := put("_mount", b)
		require.NoError(t, err)

		repos, err := os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 2, len(repos))

		resp, err := put("_unmount_all", nil)
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
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "mount1",
					Repo:   "repo",
					Branch: "master",
				},
				{
					Name:   "mount2",
					Repo:   "repo",
					Branch: "master",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		_, err := put("_mount", b)
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

		ur := UnmountRequest{
			Mounts: []string{"mount2"},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(ur))
		_, err = put("_unmount", b)
		require.NoError(t, err)

		repos, err = os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		require.Equal(t, "mount1", filepath.Base(repos[0].Name()))

		data, err = os.ReadFile(filepath.Join(mountPoint, "mount1", "dir", "file"))
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))

		mr = MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "mount1",
					Repo:   "repo2",
					Branch: "master",
				},
			},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, _ := put("_mount", b)
		require.Equal(t, 500, resp.StatusCode)
	})
}

func TestMountNonexistentRepo(t *testing.T) {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo1",
					Repo:   "repo1",
					Branch: "master",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, _ := put("_mount", b)
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
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo",
					Repo:   "repo",
					Branch: "master",
					Mode:   "rw",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err := put("_mount", b)
		require.NoError(t, err)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// the commit created above isn't actually written until we unmount, so
		// we currently have 0 commits
		require.Equal(t, 0, len(commits))
		defer resp.Body.Close()
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file1"), []byte("hello"), 0644,
		)
		require.NoError(t, err)
		ur := UnmountRequest{
			Mounts: []string{"repo"},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(ur))
		_, err = put("_unmount", b)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 1, len(commits))
	})
}

func TestRwCommitCreatesCommit(t *testing.T) {
	// Commit operation creates a commit.
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	client.NewCommit("repo", "master", "")
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo",
					Repo:   "repo",
					Branch: "master",
					Mode:   "rw",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err := put("_mount", b)
		require.NoError(t, err)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// the commit created above isn't actually written until we unmount, so
		// we currently have 0 commits
		require.Equal(t, 0, len(commits))
		defer resp.Body.Close()
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file1"), []byte("hello"), 0644,
		)
		require.NoError(t, err)

		cr := CommitRequest{
			Mount: "repo",
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(cr))
		_, err = put("_commit", b)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 1, len(commits))
	})
}

func TestRwCommitTwiceCreatesTwoCommits(t *testing.T) {
	// Two sequential commit operations create two commits (and they contain the
	// correct files).
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	client.NewCommit("repo", "master", "")
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo",
					Repo:   "repo",
					Branch: "master",
					Mode:   "rw",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err := put("_mount", b)
		require.NoError(t, err)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// the commit created above isn't actually written until we unmount, so
		// we currently have 0 commits
		require.Equal(t, 0, len(commits))
		defer resp.Body.Close()
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file1"), []byte("hello"), 0644,
		)
		require.NoError(t, err)

		cr := CommitRequest{
			Mount: "repo",
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(cr))
		_, err = put("_commit", b)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 1, len(commits))

		// another file!
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file2"), []byte("hello"), 0644,
		)
		require.NoError(t, err)

		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(cr))
		_, err = put("_commit", b)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 2, len(commits))
	})
}

func TestRwCommitUnmountCreatesTwoCommits(t *testing.T) {
	// Commit and then unmount results in two commits, since unmounting creates
	// one too.
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	client.NewCommit("repo", "master", "")
	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo",
					Repo:   "repo",
					Branch: "master",
					Mode:   "rw",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err := put("_mount", b)
		require.NoError(t, err)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// the commit created above isn't actually written until we unmount, so
		// we currently have 0 commits
		require.Equal(t, 0, len(commits))
		defer resp.Body.Close()
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file1"), []byte("hello"), 0644,
		)
		require.NoError(t, err)

		cr := CommitRequest{
			Mount: "repo",
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(cr))
		_, err = put("_commit", b)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 1, len(commits))

		// another file!
		err = os.WriteFile(
			filepath.Join(mountPoint, "repo", "file2"), []byte("hello"), 0644,
		)
		require.NoError(t, err)
		ur := UnmountRequest{
			Mounts: []string{"repo"},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(ur))
		_, err = put("_unmount", b)
		require.NoError(t, err)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Name: "repo",
			Type: pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 2, len(commits))
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
