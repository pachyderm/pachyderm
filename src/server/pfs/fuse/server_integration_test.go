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

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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
		require.Equal(t, 200, putResp.StatusCode)
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
		require.Equal(t, 200, getResp.StatusCode)
		defer getResp.Body.Close()

		getConfig := &Config{}
		require.NoError(t, json.NewDecoder(getResp.Body).Decode(getConfig))

		require.Equal(t, cfg.ClusterStatus, getConfig.ClusterStatus)
		require.Equal(t, cfg.PachdAddress, getConfig.PachdAddress)
	})
}

func TestMountDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "master", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/'}}}`,
			pfs.DefaultProjectName),
		)
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 1, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo"))
		require.NoError(t, err)
		require.Equal(t, 2, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo", "dir"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		resp, err = put("_unmount_all", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		input = []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/*'}}}`,
			pfs.DefaultProjectName),
		)
		resp, err = put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)

		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
	})
}

func TestCrossDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo1"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo1", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo2"))
	commit = client.NewCommit(pfs.DefaultProjectName, "repo2", "dev", "")
	err = c.PutFile(commit, "dir/file3", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file4", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, "repo2", "dev", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'cross': [{'pfs': {'glob': '/', 'project': '%s', 'repo': 'repo1'}},
			{'pfs': {'glob': '/*', 'repo': 'repo2', 'branch': 'dev'}}]}}}`,
			pfs.DefaultProjectName),
		)
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo1"))
		require.NoError(t, err)
		require.Equal(t, 2, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo2_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
	})
}

func TestUnionDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo1"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo1", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo2"))
	commit = client.NewCommit(pfs.DefaultProjectName, "repo2", "dev", "")
	err = c.PutFile(commit, "dir/file3", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file4", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, "repo2", "dev", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'union': [{'pfs': {'glob': '/', 'project': '%s', 'repo': 'repo1'}},
			{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo2', 'branch': 'dev'}}]}}}`,
			pfs.DefaultProjectName, pfs.DefaultProjectName),
		)
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 3, mdr.NumDatums)
	})
}

func TestRepeatedBranchesDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo1"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo1", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	commit = client.NewCommit(pfs.DefaultProjectName, "repo1", "dev", "")
	err = c.PutFile(commit, "dir/file3", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file4", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, "repo1", "dev", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'cross': [{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1'}},
			{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1'}}]}}`,
			pfs.DefaultProjectName, pfs.DefaultProjectName),
		)
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 4, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo1"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		resp, err = put("_unmount_all", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		input = []byte(fmt.Sprintf(
			`{'input': {'cross': [{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1'}},
			{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1'}},
			{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1', 'branch': 'dev'}}]}}`,
			pfs.DefaultProjectName, pfs.DefaultProjectName, pfs.DefaultProjectName),
		)
		resp, err = put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

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
		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo1_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
	})
}

func TestShowDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "dev", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo", "dev", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/*', 'branch': 'dev'}}}`,
			pfs.DefaultProjectName),
		)
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		datum1Id := mdr.Id
		require.Equal(t, 2, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		resp, err = put("_show_datum?idx=1", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 1, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)

		resp, err = put(fmt.Sprintf("_show_datum?idx=1&id=%s", datum1Id), nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.Equal(t, datum1Id, mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
	})
}

func TestGetDatums(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "dev", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo", "dev", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		resp, err := get("datums")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		dr := &DatumsResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(dr))
		require.Equal(t, 0, dr.NumDatums)

		input := []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/*', 'branch': 'dev'}}}`,
			pfs.DefaultProjectName),
		)
		resp, err = put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		datum1Id := mdr.Id
		require.Equal(t, 2, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		resp, err = get("datums")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		dr = &DatumsResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(dr))
		require.Equal(t, 2, dr.NumDatums)
		require.Equal(t, "repo", dr.Input.Pfs.Repo)
		require.Equal(t, "dev", dr.Input.Pfs.Branch)
		require.Equal(t, "/*", dr.Input.Pfs.Glob)
		require.Equal(t, 0, dr.CurrIdx)

		resp, err = put("_show_datum?idx=1", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 1, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)

		resp, err = get("datums")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		dr = &DatumsResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(dr))
		require.Equal(t, 1, dr.CurrIdx)

		resp, err = put(fmt.Sprintf("_show_datum?idx=1&id=%s", datum1Id), nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.Equal(t, datum1Id, mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
	})
}

func TestMountShowDatumsCrossProject(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo1"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
	err := c.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo1", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	commit = client.NewCommit(pfs.DefaultProjectName, "repo1", "dev", "")
	err = c.PutFile(commit, "dir/file3", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file4", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, "repo1", "dev", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	projectName := tu.UniqueString("p1")
	require.NoError(t, c.CreateProject(projectName))
	require.NoError(t, c.CreateRepo(projectName, "repo1"))
	commit = client.NewCommit(projectName, "repo1", "master", "")
	err = c.PutFile(commit, "dir/file3", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file4", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(projectName, "repo1", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	commit = client.NewCommit(projectName, "repo1", "dev", "")
	err = c.PutFile(commit, "dir/file3", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file4", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(projectName, "repo1", "dev", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(`{'input': {'cross': [{'pfs': {'glob': '/*', 'repo': 'repo1'}},
			{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1', 'branch': 'dev'}}]}}`,
			pfs.DefaultProjectName),
		)
		resp, err := put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 4, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo1"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo1_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		input = []byte(fmt.Sprintf(`{'input': {'cross': [{'pfs': {'glob': '/*', 'repo': 'repo1'}},
			{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1'}},
			{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1', 'branch': 'dev'}}]}}`,
			projectName, projectName),
		)
		resp, err = put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 8, mdr.NumDatums)

		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo1"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, projectName+"_repo1"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, projectName+"_repo1_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		resp, err = put("_show_datum?idx=1", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 1, mdr.Idx)
		require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 8, mdr.NumDatums)

		input = []byte("{'input': {'pfs': {'project': 'invalid', 'repo': 'repo1', 'glob': '/*'}}}")
		resp, err = put("_mount_datums", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})
}
