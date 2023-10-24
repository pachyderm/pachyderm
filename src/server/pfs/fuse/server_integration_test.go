//go:build k8s

package fuse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

// TODO: Uncomment all checks below for datum ID when we transition fully to using ListDatum
// when getting datums. #ListDatumPagination

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
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 1, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

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
		resp, err = put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
	})
}

func TestMountDatumGlobFile(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "master", "")
	err := c.PutFile(commit, "file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "files/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "files/file2", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "otherfile", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/file*'}}}`,
			pfs.DefaultProjectName),
		)
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 3, mdr.NumDatums)

		files, err := os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		require.Equal(t, true, strings.HasPrefix(files[0].Name(), "file"))
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
			`{'input': {
				'cross': [
					{'pfs': {'glob': '/', 'project': '%s', 'repo': 'repo1'}},
					{'pfs': {'glob': '/*', 'repo': 'repo2', 'branch': 'dev'}}
				]
			}}`,
			pfs.DefaultProjectName),
		)
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		files, err := os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo1"))
		require.NoError(t, err)
		require.Equal(t, 2, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo2_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		input = []byte(fmt.Sprintf(
			`{'input': {
				'cross': [
					{
						'cross': [
							{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1'}},
							{'pfs': {'glob': '/*/*', 'project': '%s', 'repo': 'repo1'}}
						]
					},
					{
						'cross': [
							{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo2', 'branch': 'dev'}},
							{'pfs': {'glob': '/*/*', 'project': '%s', 'repo': 'repo2', 'branch': 'dev'}}
						]
					},
					{
						'cross': [
							{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1'}},
							{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo2', 'branch': 'dev'}}
						]
					}
				]
			}}`,
			pfs.DefaultProjectName, pfs.DefaultProjectName, pfs.DefaultProjectName, pfs.DefaultProjectName, pfs.DefaultProjectName, pfs.DefaultProjectName),
		)
		resp, err = put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 16, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)
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
			`{'input': {
				'union': [
					{'pfs': {'glob': '/', 'project': '%s', 'repo': 'repo1'}},
					{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo2', 'branch': 'dev'}}
				]
			}}`,
			pfs.DefaultProjectName, pfs.DefaultProjectName),
		)
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 3, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		input = []byte(fmt.Sprintf(
			`{'input': {
				'union': [
					{
						'union': [
							{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo1'}},
							{'pfs': {'glob': '/*/*', 'project': '%s', 'repo': 'repo1'}}
						]
					},
					{
						'union': [
							{'pfs': {'glob': '/*', 'project': '%s', 'repo': 'repo2', 'branch': 'dev'}},
							{'pfs': {'glob': '/*/*', 'project': '%s', 'repo': 'repo2', 'branch': 'dev'}}
						]
					}
				]
			}}`,
			pfs.DefaultProjectName, pfs.DefaultProjectName, pfs.DefaultProjectName, pfs.DefaultProjectName),
		)
		resp, err = put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 6, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		files, err := os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		resp, err = put("datums/_next", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		files, err = os.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
	})
}

func TestJoinDatum(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo1"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
	err := c.PutFile(commit, "file1.txt", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file2.txt", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "file3.txt", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo1", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo2"))
	commit = client.NewCommit(pfs.DefaultProjectName, "repo2", "master", "")
	err = c.PutFile(commit, "dir/file2.txt", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "dir/file3.txt", strings.NewReader("foo"))
	require.NoError(t, err)
	err = c.PutFile(commit, "dir/file4.txt", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, "repo2", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {
				'join': [
					{'pfs': {'glob': '/(*).txt', 'project': '%s', 'repo': 'repo1', 'join_on': '$1'}},
					{'pfs': {'glob': '/*/(*).txt', 'project': '%s', 'repo': 'repo2', 'join_on': '$1'}}
				]
			}}`,
			pfs.DefaultProjectName, pfs.DefaultProjectName),
		)
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

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
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 4, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

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
		resp, err = put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 8, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

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
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		// datum1Id := mdr.Id
		require.Equal(t, 2, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		files, err := os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		resp, err = put("datums/_next", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 1, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		resp, err = put("datums/_next", nil)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)

		resp, err = put("datums/_prev", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.Equal(t, datum1Id, mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		resp, err = put("datums/_prev", nil)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
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
		require.Equal(t, false, dr.AllDatumsReceived)

		input := []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/*', 'branch': 'dev'}}}`,
			pfs.DefaultProjectName),
		)
		resp, err = put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		// datum1Id := mdr.Id
		require.Equal(t, 2, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

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
		require.Equal(t, 0, dr.Idx)
		require.Equal(t, true, dr.AllDatumsReceived)

		resp, err = put("datums/_next", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 1, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		resp, err = get("datums")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		dr = &DatumsResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(dr))
		require.Equal(t, 1, dr.Idx)
		require.Equal(t, true, dr.AllDatumsReceived)

		resp, err = put("datums/_prev", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.Equal(t, datum1Id, mdr.Id)
		require.Equal(t, 2, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)
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
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 4, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

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
		resp, err = put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 8, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		files, err = os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo1"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, projectName+"_repo1"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		files, err = os.ReadDir(filepath.Join(mountPoint, projectName+"_repo1_dev"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))

		resp, err = put("datums/_next", nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 1, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, 8, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		input = []byte("{'input': {'pfs': {'project': 'invalid', 'repo': 'repo1', 'glob': '/*'}}}")
		resp, err = put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})
}

func TestMountDatumsBranchHeadFromOtherBranch(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo1"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
	err := c.PutFile(commit, "file1", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo1", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)
	commit = client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
	err = c.PutFile(commit, "file2", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo2, err := c.InspectCommit(pfs.DefaultProjectName, "repo1", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo2.Commit.Id)
	require.NoError(t, err)

	require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "repo1", "copy", "", commitInfo.Commit.Id, nil))

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo1', 'glob': '/*', 'branch': 'copy'}}}`,
			pfs.DefaultProjectName),
		)
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		files, err := os.ReadDir(filepath.Join(mountPoint, pfs.DefaultProjectName+"_repo1_copy"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		require.Equal(t, "file1", filepath.Base(files[0].Name()))

		resp, err = get("mounts")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mountResp := &ListMountResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mountResp))
		require.Equal(t, 1, len((*mountResp).Mounted))
		require.Equal(t, "master", (*mountResp).Mounted[0].Branch)
		require.Equal(t, commitInfo.Commit.Id, (*mountResp).Mounted[0].Commit)
	})
}

func TestMountDatumsSpecifyingCommit(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "master", "")
	err := c.PutFile(commit, "file1", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/*', 'commit': '%s'}}}`,
			pfs.DefaultProjectName, commitInfo.Commit.Id),
		)
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})
}

func TestMountDatumsPagination(t *testing.T) {
	t.Skip("Include test when ListDatum pagination is efficient #ListDatumPagination")
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo"))

	// Arbitrary number of files less than fuse.NumDatumsPerPage
	lessThanNumDatumsPerPage := 200
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", strconv.Itoa(lessThanNumDatumsPerPage), "")
	for i := 0; i < lessThanNumDatumsPerPage; i++ {
		err := c.PutFile(commit, fmt.Sprintf("file%d", i), strings.NewReader("foo"))
		require.NoError(t, err)
	}
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo", strconv.Itoa(lessThanNumDatumsPerPage), "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	numDatumsPerPage := NumDatumsPerPage
	commit = client.NewCommit(pfs.DefaultProjectName, "repo", strconv.Itoa(numDatumsPerPage), "")
	for i := 0; i < numDatumsPerPage; i++ {
		err := c.PutFile(commit, fmt.Sprintf("file%d", i), strings.NewReader("foo"))
		require.NoError(t, err)
	}
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, "repo", strconv.Itoa(numDatumsPerPage), "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	numDatumsPerPagePlusOne := numDatumsPerPage + 1
	commit = client.NewCommit(pfs.DefaultProjectName, "repo", strconv.Itoa(numDatumsPerPagePlusOne), "")
	for i := 0; i < numDatumsPerPagePlusOne; i++ {
		err := c.PutFile(commit, fmt.Sprintf("file%d", i), strings.NewReader("foo"))
		require.NoError(t, err)
	}
	commitInfo, err = c.InspectCommit(pfs.DefaultProjectName, "repo", strconv.Itoa(numDatumsPerPagePlusOne), "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/*', 'branch': '%d'}}}`,
			pfs.DefaultProjectName, lessThanNumDatumsPerPage),
		)
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		mdr := &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, lessThanNumDatumsPerPage, mdr.NumDatums)
		require.Equal(t, true, mdr.AllDatumsReceived)

		input = []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/*', 'branch': '%d'}}}`,
			pfs.DefaultProjectName, numDatumsPerPage),
		)
		resp, err = put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, numDatumsPerPage, mdr.NumDatums)
		require.Equal(t, false, mdr.AllDatumsReceived)

		// Cycle to last datum in page
		for i := 0; i < numDatumsPerPage-1; i++ {
			resp, err = put("datums/_next", nil)
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode)
		}
		resp, err = put("datums/_next", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)

		resp, err = get("datums")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		dr := &DatumsResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(dr))
		require.Equal(t, numDatumsPerPage, dr.NumDatums)
		require.Equal(t, numDatumsPerPage-1, dr.Idx)
		require.Equal(t, true, dr.AllDatumsReceived)

		input = []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/*', 'branch': '%d'}}}`,
			pfs.DefaultProjectName, numDatumsPerPagePlusOne),
		)
		resp, err = put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, 0, mdr.Idx)
		// require.NotEqual(t, "", mdr.Id)
		require.Equal(t, numDatumsPerPage, mdr.NumDatums)
		require.Equal(t, false, mdr.AllDatumsReceived)

		// Cycle to last datum in page
		for i := 0; i < numDatumsPerPage-1; i++ {
			resp, err = put("datums/_next", nil)
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode)
		}
		resp, err = put("datums/_next", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		mdr = &MountDatumResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(mdr))
		require.Equal(t, numDatumsPerPagePlusOne, mdr.NumDatums)
		require.Equal(t, numDatumsPerPagePlusOne-1, mdr.Idx)
		require.Equal(t, true, mdr.AllDatumsReceived)
	})
}

func TestMountDatumsNoGlobMatch(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo"))
	commit := client.NewCommit(pfs.DefaultProjectName, "repo", "master", "")
	err := c.PutFile(commit, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	commitInfo, err := c.InspectCommit(pfs.DefaultProjectName, "repo", "master", "")
	require.NoError(t, err)
	_, err = c.WaitCommitSetAll(commitInfo.Commit.Id)
	require.NoError(t, err)

	withServerMount(t, c, nil, func(mountPoint string) {
		input := []byte(fmt.Sprintf(
			`{'input': {'pfs': {'project': '%s', 'repo': 'repo', 'glob': '/file1*', 'commit': '%s'}}}`,
			pfs.DefaultProjectName, commitInfo.Commit.Id),
		)
		resp, err := put("datums/_mount", bytes.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})
}
