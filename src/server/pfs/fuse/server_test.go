//go:build k8s

package fuse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func put(path string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://localhost:9002/%s", path), body)
	if err != nil {
		panic(err)
	}
	x, err := client.Do(req)
	return x, errors.EnsureStack(err)
}

func get(path string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:9002/%s", path), nil)
	if err != nil {
		panic(err)
	}
	x, err := client.Do(req)
	return x, errors.EnsureStack(err)
}

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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
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

func TestRepoAccess(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice, bob := "robot"+tu.UniqueString("alice"), "robot"+tu.UniqueString("bob")
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	require.NoError(t, aliceClient.CreateRepo("repo1"))
	commit := client.NewCommit("repo1", "master", "")
	err := aliceClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, aliceClient, nil, func(mountPoint string) {
		resp, err := get("repos")
		require.NoError(t, err)

		reposResp := &ListRepoResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(reposResp))
		require.Equal(t, "write", (*reposResp)["repo1"].Authorization)
	})

	withServerMount(t, bobClient, nil, func(mountPoint string) {
		resp, err := get("repos")
		require.NoError(t, err)

		reposResp := &ListRepoResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(reposResp))
		require.Equal(t, "none", (*reposResp)["repo1"].Authorization)

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

func TestUnmountAll(t *testing.T) {
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

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

func TestConfig(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	c = tu.AuthenticateClient(t, c, auth.RootUser)

	withServerMount(t, c, nil, func(mountPoint string) {
		type Config struct {
			ClusterStatus string `json:"cluster_status"`
			PachdAddress  string `json:"pachd_address"`
		}

		// PUT
		invalidCfg := &Config{ClusterStatus: "INVALID", PachdAddress: "bad_address"}
		m := map[string]string{"pachd_address": invalidCfg.PachdAddress}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(m))

		putResp, err := put("config", b)
		require.NoError(t, err)
		require.Equal(t, 500, putResp.StatusCode)

		cfg := &Config{ClusterStatus: "AUTH_ENABLED", PachdAddress: c.GetAddress().Qualified()}
		m = map[string]string{"pachd_address": cfg.PachdAddress}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(m))

		putResp, err = put("config", b)
		require.NoError(t, err)
		defer putResp.Body.Close()

		putConfig := &Config{}
		require.NoError(t, json.NewDecoder(putResp.Body).Decode(putConfig))

		cfgParsedPachdAddress, err := grpcutil.ParsePachdAddress(cfg.PachdAddress)
		require.NoError(t, err)

		require.Equal(t, cfg.ClusterStatus, putConfig.ClusterStatus)
		require.Equal(t, cfgParsedPachdAddress.Qualified(), putConfig.PachdAddress)
		require.Equal(t, cfgParsedPachdAddress.Qualified(), c.GetAddress().Qualified())

		// GET
		getResp, err := get("config")
		require.NoError(t, err)
		defer getResp.Body.Close()

		getConfig := &Config{}
		require.NoError(t, json.NewDecoder(getResp.Body).Decode(getConfig))

		require.Equal(t, cfg.ClusterStatus, getConfig.ClusterStatus)
		require.Equal(t, cfg.PachdAddress, getConfig.PachdAddress)
	})
}

func TestAuthLoginLogout(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c))
	c = tu.UnauthenticatedPachClient(t, c)

	withServerMount(t, c, nil, func(mountPoint string) {
		authResp, err := put("auth/_login", nil)
		require.NoError(t, err)
		defer authResp.Body.Close()

		type AuthLoginResp struct {
			AuthUrl string `json:"auth_url"`
		}
		getAuthLogin := &AuthLoginResp{}
		require.NoError(t, json.NewDecoder(authResp.Body).Decode(getAuthLogin))

		tu.DoOAuthExchange(t, c, c, getAuthLogin.AuthUrl)
		time.Sleep(1 * time.Second)

		_, err = c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.NoError(t, err)

		_, err = put("auth/_logout", nil)
		require.NoError(t, err)

		_, err = c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.ErrorIs(t, err, auth.ErrNotSignedIn)
	})
}

func TestUnauthenticatedCode(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
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

func TestMultipleMount(t *testing.T) {
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

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

// TODO: pass reference to the MountManager object to the test func, so that the
// test can call MountBranch, UnmountBranch etc directly for convenience
func withServerMount(tb testing.TB, c *client.APIClient, sopts *ServerOptions, f func(mountPoint string)) {
	dir := tb.TempDir()
	if sopts == nil {
		sopts = &ServerOptions{
			MountDir: dir,
		}
	}
	if sopts.Unmount == nil {
		sopts.Unmount = make(chan struct{})
	}
	unmounted := make(chan struct{})
	var mountErr error
	defer func() {
		close(sopts.Unmount)
		<-unmounted
		require.ErrorIs(tb, mountErr, http.ErrServerClosed)
	}()
	defer func() {
		// recover because panics leave the mount in a weird state that makes
		// it hard to rerun the tests, mostly relevent when you're iterating on
		// these tests, or the code they test.
		if r := recover(); r != nil {
			tb.Fatal(r)
		}
	}()
	go func() {
		mountErr = Server(sopts, c)
		close(unmounted)
	}()
	// Gotta give the fuse mount time to come up.
	time.Sleep(2 * time.Second)
	f(dir)
}

func TestRwUnmountCreatesCommit(t *testing.T) {
	// Unmounting a mounted read-write filesystem which has had some data
	// written to it results in a new commit with that data in it.
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	_, err := get("health")
	require.YesError(t, err)

	withServerMount(t, env.PachClient, nil, func(mountPoint string) {
		_, err = get("health")
		require.NoError(t, err)
	})
}

func TestMountDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo("repo"))
	commit := client.NewCommit("repo", "master", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte("{'input': {'pfs': {'repo': 'repo', 'glob': '/'}}}")
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 1, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint, "repo"))
		require.NoError(t, err)
		require.Equal(t, 2, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, "repo", "dir"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		_, err = put("_unmount_all", nil)
		require.NoError(t, err)

		input = []byte("{'input': {'pfs': {'repo': 'repo', 'glob': '/*'}}}")
		resp, err = put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)

		files, err = os.ReadDir(filepath.Join(mountPoint, "repo"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
	})
}

func TestCrossDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)

	require.NoError(t, c.CreateRepo("repo1"))
	commit := client.NewCommit("repo1", "master", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)

	require.NoError(t, c.CreateRepo("repo2"))
	commit = client.NewCommit("repo2", "dev", "")
	err = c.PutFile(commit, "dir/file3", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file4", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte("{'input': {'cross': [{'pfs': {'glob': '/', 'repo': 'repo1'}}, {'pfs': {'glob': '/*', 'repo': 'repo2', 'branch': 'dev'}}]}}}")
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint, "repo1"))
		require.NoError(t, err)
		require.Equal(t, 2, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, "repo2_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
	})
}

func TestUnionDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)

	require.NoError(t, c.CreateRepo("repo1"))
	commit := client.NewCommit("repo1", "master", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)

	require.NoError(t, c.CreateRepo("repo2"))
	commit = client.NewCommit("repo2", "dev", "")
	err = c.PutFile(commit, "dir/file3", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file4", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte("{'input': {'union': [{'pfs': {'glob': '/', 'repo': 'repo1'}}, {'pfs': {'glob': '/*', 'repo': 'repo2', 'branch': 'dev'}}]}}}")
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 3, mdr.NumDatums)
	})
}

func TestRepeatedBranchesDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo("repo1"))
	commit := client.NewCommit("repo1", "master", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	commit = client.NewCommit("repo1", "dev", "")
	err = c.PutFile(commit, "dir/file3", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file4", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte("{'input': {'cross': [{'pfs': {'glob': '/*', 'repo': 'repo1'}}, {'pfs': {'glob': '/*', 'repo': 'repo1'}}]}}")
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 4, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, "repo1"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		_, err = put("_unmount_all", nil)
		require.NoError(t, err)

		input = []byte("{'input': {'cross': [{'pfs': {'glob': '/*', 'repo': 'repo1'}}, {'pfs': {'glob': '/*', 'repo': 'repo1'}}, {'pfs': {'glob': '/*', 'repo': 'repo1', 'branch': 'dev'}}]}}")
		resp, err = put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 8, mdr.NumDatums)

		files, err = os.ReadDir(filepath.Join(mountPoint))
		require.NoError(t, err)
		require.Equal(t, 3, len(files)) // Need to account for "out" rw mount
		files, err = os.ReadDir(filepath.Join(mountPoint, "repo1_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
	})
}

func TestShowDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo("repo"))
	commit := client.NewCommit("repo", "dev", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte("{'input': {'pfs': {'repo': 'repo', 'glob': '/*', 'branch': 'dev'}}}")
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		datum1Id := mdr.Id
		require.Equal(t, 2, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint, "repo_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		resp, err = put("_show_datum?idx=1", nil)
		require.NoError(t, err)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 1, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)

		resp, err = put(fmt.Sprintf("_show_datum?idx=1&id=%s", datum1Id), nil)
		require.NoError(t, err)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.Equal(t, datum1Id, mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
	})
}
