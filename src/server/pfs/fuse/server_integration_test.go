//go:build k8s

package fuse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
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

//TODO(fahad): requires identity server to work with realEnv.
func TestUnauthenticatedCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
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

//TODO(fahad): requires identity server to work with realEnv.
func TestAuthLoginLogout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
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

		_, err = os.ReadDir(filepath.Join(mountPoint, "out")) // Loads "out" folder
		require.NoError(t, err)
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

func TestGetDatums(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo("repo"))
	commit := client.NewCommit("repo", "dev", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		resp, err := get("datums")
		require.NoError(t, err)

		dr := &DatumsResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(dr))
		require.Equal(t, 0, dr.NumDatums)

		input := []byte("{'input': {'pfs': {'repo': 'repo', 'glob': '/*', 'branch': 'dev'}}}")
		resp, err = put("_mount_datums", bytes.NewReader(input))
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

		resp, err = get("datums")
		require.NoError(t, err)

		dr = &DatumsResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(dr))
		require.Equal(t, 2, dr.NumDatums)
		require.Equal(t, "repo", dr.Input.Pfs.Repo)
		require.Equal(t, "dev", dr.Input.Pfs.Branch)
		require.Equal(t, "/*", dr.Input.Pfs.Glob)
		require.Equal(t, 0, dr.CurrIdx)

		resp, err = put("_show_datum?idx=1", nil)
		require.NoError(t, err)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 1, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)

		resp, err = get("datums")
		require.NoError(t, err)

		dr = &DatumsResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(dr))
		require.Equal(t, 1, dr.CurrIdx)

		resp, err = put(fmt.Sprintf("_show_datum?idx=1&id=%s", datum1Id), nil)
		require.NoError(t, err)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.Equal(t, datum1Id, mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
	})
}
