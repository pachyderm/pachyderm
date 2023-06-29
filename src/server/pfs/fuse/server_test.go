//go:build unit_test

package fuse

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
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
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "master", "")
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
		require.Equal(t, 200, resp.StatusCode)

		defer resp.Body.Close()
		mountResp := &ListMountResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mountResp))
		require.Equal(t, "repo", (*mountResp).Mounted[0].Repo)
		require.Equal(t, "master", (*mountResp).Mounted[0].Branch)

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
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "dev", "")
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
		resp, err := put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

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
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "master", "")
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
		resp, err := put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

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
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo1"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
	err := env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo2"))
	commit = client.NewCommit(pfs.DefaultProjectName, "repo2", "master", "")
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
		resp, err := put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		repos, err := os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 2, len(repos))

		resp, err = put("_unmount_all", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		defer resp.Body.Close()
		unmountResp := &ListMountResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(unmountResp))
		require.Equal(t, 2, len((*unmountResp).Unmounted))

		repos, err = os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 0, len(repos))
	})
}

func TestMultipleMount(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "master", "")
	err := env.PachClient.PutFile(commit, "dir/file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo2"))
	commit = client.NewCommit(pfs.DefaultProjectName, "repo2", "master", "")
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
		resp, err := put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

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
		resp, err = put("_unmount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

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
		resp, _ = put("_mount", b)
		require.Equal(t, 500, resp.StatusCode)
	})
}

func TestMountNonexistentRepo(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

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

func TestRwMountNonexistentBranch(t *testing.T) {
	// Mounting a nonexistent branch fails for read-only mounts and succeeds for read-write mounts
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo",
					Repo:   "repo",
					Branch: "master",
					Mode:   "ro",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err := put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)

		mr = MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo",
					Repo:   "repo",
					Branch: "master",
					Mode:   "rw",
				},
			},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err = put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})
}

func TestRwUnmountCreatesCommit(t *testing.T) {
	// Unmounting a mounted read-write filesystem which has had some data
	// written to it results in a new commit with that data in it.
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))

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
		require.Equal(t, 200, resp.StatusCode)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
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
		resp, err = put("_unmount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 1, len(commits))
	})
}

func TestRwCommitCreatesCommit(t *testing.T) {
	// Commit operation creates a commit.
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))

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
		require.Equal(t, 200, resp.StatusCode)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
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
		resp, err = put("_commit", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 1, len(commits))
	})
}

func TestRwCommitTwiceCreatesTwoCommits(t *testing.T) {
	// Two sequential commit operations create two commits (and they contain the
	// correct files).
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))

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
		require.Equal(t, 200, resp.StatusCode)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
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
		resp, err = put("_commit", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
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
		resp, err = put("_commit", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 2, len(commits))
	})
}

func TestRwCommitUnmountCreatesTwoCommits(t *testing.T) {
	// Commit and then unmount results in two commits, since unmounting creates
	// one too.
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))

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
		require.Equal(t, 200, resp.StatusCode)

		commits, err := env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
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
		resp, err = put("_commit", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
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
		resp, err = put("_unmount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		commits, err = env.PachClient.ListCommitByRepo(&pfs.Repo{
			Project: &pfs.Project{Name: pfs.DefaultProjectName},
			Name:    "repo",
			Type:    pfs.UserRepoType,
		})
		require.NoError(t, err)
		// we have one more commit than we did previously!
		require.Equal(t, 2, len(commits))
	})
}

func TestHealth(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	_, err := get("health")
	require.YesError(t, err)

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		resp, err := get("health")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})
}

// Requires 'use_allow_other' to be enabled in /etc/fuse.conf
func TestAuthLoginLogout(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnvWithIdentity(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, true))
	c = tu.UnauthenticatedPachClient(t, c)

	withServerMount(t, c, nil, func(mountPoint string) {
		authResp, err := put("auth/_login", nil)
		require.NoError(t, err)
		require.Equal(t, 200, authResp.StatusCode)
		defer authResp.Body.Close()

		type AuthLoginResp struct {
			AuthUrl   string `json:"auth_url"`
			OidcState string `json:"oidc_state"`
		}
		getAuthLogin := &AuthLoginResp{}
		require.NoError(t, json.NewDecoder(authResp.Body).Decode(getAuthLogin))

		tu.DoOAuthExchange(t, c, c, getAuthLogin.AuthUrl)
		time.Sleep(1 * time.Second)

		b := bytes.NewBufferString(getAuthLogin.OidcState)
		tokenResp, err := put("auth/_login_token", b)
		require.NoError(t, err)
		require.Equal(t, 200, tokenResp.StatusCode)
		defer tokenResp.Body.Close()

		_, err = c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.NoError(t, err)

		resp, err := put("auth/_logout", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		_, err = c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.ErrorIs(t, err, auth.ErrNotSignedIn)
	})
}

func TestRepoAccess(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnvWithIdentity(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	alice, bob := "robot"+tu.UniqueString("alice"), "robot"+tu.UniqueString("bob")
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	require.NoError(t, aliceClient.CreateRepo(pfs.DefaultProjectName, "repo1"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
	err := aliceClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, aliceClient, nil, func(mountPoint string) {
		resp, err := get("repos")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		reposResp := &ListRepoResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(reposResp))
		require.Equal(t, "write", (*reposResp)[0].Authorization)
	})

	withServerMount(t, bobClient, nil, func(mountPoint string) {
		resp, err := get("repos")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		reposResp := &ListRepoResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(reposResp))
		require.Equal(t, "none", (*reposResp)[0].Authorization)

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
		resp, _ = put("_mount", b)
		require.Equal(t, 500, resp.StatusCode)
	})
}

func TestUnauthenticatedCode(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnvWithIdentity(ctx, t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	withServerMount(t, c, nil, func(mountPoint string) {
		resp, _ := get("repos")
		require.Equal(t, 200, resp.StatusCode)
	})

	c = tu.UnauthenticatedPachClient(t, c)
	withServerMount(t, c, nil, func(mountPoint string) {
		resp, _ := get("repos")
		require.Equal(t, 401, resp.StatusCode)
	})

	c = tu.AuthenticateClient(t, c, "test")
	withServerMount(t, c, nil, func(mountPoint string) {
		resp, _ := get("repos")
		require.Equal(t, 200, resp.StatusCode)
	})
}

func TestDeletingMountedRepo(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))

	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "b1", "")
	err := env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	commit = client.NewCommit(pfs.DefaultProjectName, "repo", "b2", "")
	err = env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "repo_b1",
					Repo:   "repo",
					Branch: "b1",
					Mode:   "ro",
				},
				{
					Name:   "repo_b1_dup",
					Repo:   "repo",
					Branch: "b1",
					Mode:   "ro",
				},
				{
					Name:   "repo_b2",
					Repo:   "repo",
					Branch: "b2",
					Mode:   "ro",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err := put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, env.PachClient.DeleteBranch(pfs.DefaultProjectName, "repo", "b1", false))
		resp, err = get("mounts")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mountResp := &ListMountResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mountResp))
		require.Equal(t, 1, len((*mountResp).Mounted))
		require.Equal(t, "b2", (*mountResp).Mounted[0].Branch)
		require.Equal(t, 1, len((*mountResp).Unmounted))

		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, "repo", false))
		resp, err = get("mounts")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mountResp = &ListMountResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mountResp))
		require.Equal(t, 0, len((*mountResp).Mounted))
		require.Equal(t, 0, len((*mountResp).Unmounted))
	})
}

func TestMountWithProjects(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "b1", "")
	err := env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	commit = client.NewCommit(pfs.DefaultProjectName, "repo", "b2", "")
	err = env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	projectName := tu.UniqueString("p1")
	require.NoError(t, env.PachClient.CreateProject(projectName))
	require.NoError(t, env.PachClient.CreateRepo(projectName, "repo_p1"))
	commit = client.NewCommit(projectName, "repo_p1", "b1", "")
	err = env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	commit = client.NewCommit(projectName, "repo_p1", "b2", "")
	err = env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		mr := MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "default_repo_b1",
					Repo:   "repo",
					Branch: "b1",
				},
			},
		}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err := put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mr = MountRequest{
			Mounts: []*MountInfo{
				{
					Name:    "default_repo_b1_dup1",
					Project: "",
					Repo:    "repo",
					Branch:  "b1",
				},
			},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err = put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mr = MountRequest{
			Mounts: []*MountInfo{
				{
					Name:    "default_repo_b1_dup2",
					Project: "default",
					Repo:    "repo",
					Branch:  "b1",
				},
			},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err = put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mr = MountRequest{
			Mounts: []*MountInfo{
				{
					Name:    "invalid_repo_b1",
					Project: "invalid",
					Repo:    "repo",
					Branch:  "b1",
				},
			},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err = put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)

		mr = MountRequest{
			Mounts: []*MountInfo{
				{
					Name:   "p1_repo_p1_b2",
					Repo:   "repo_p1",
					Branch: "b2",
				},
			},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err = put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)

		mr = MountRequest{
			Mounts: []*MountInfo{
				{
					Name:    "p1_repo_p1_b2",
					Project: projectName,
					Repo:    "repo_p1",
					Branch:  "b2",
				},
			},
		}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(mr))
		resp, err = put("_mount", b)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mountsList := &ListMountResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mountsList))
		require.Equal(t, 4, len((*mountsList).Mounted))
	})
}

func TestProjects(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "b1", "")
	err := env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	commit = client.NewCommit(pfs.DefaultProjectName, "repo", "b2", "")
	err = env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	projectName := tu.UniqueString("p1")
	require.NoError(t, env.PachClient.CreateProject(projectName))
	require.NoError(t, env.PachClient.CreateRepo(projectName, "repo_p1"))
	commit = client.NewCommit(projectName, "repo_p1", "b1", "")
	err = env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	commit = client.NewCommit(projectName, "repo_p1", "b2", "")
	err = env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	emptyProjectName := tu.UniqueString("p2")
	require.NoError(t, env.PachClient.CreateProject(emptyProjectName))

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		type Project struct {
			Name string `json:"name"`
		}

		type ProjectAuth struct {
			Permissions []int    `json:"permissions"`
			Roles       []string `json:"roles"`
		}

		type ProjectResp struct {
			Project Project     `json:"project"`
			Auth    ProjectAuth `json:"auth_info"`
		}

		resp, err := get("projects")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		defer resp.Body.Close()
		projectData := []ProjectResp{}

		require.NoError(t, json.NewDecoder(resp.Body).Decode(&projectData))
		require.Equal(t, len(projectData), 3)
		require.Equal(t, projectData[0].Project.Name, emptyProjectName)
		require.Equal(t, projectData[1].Project.Name, projectName)
		require.Equal(t, projectData[2].Project.Name, pfs.DefaultProjectName)
	})
}
