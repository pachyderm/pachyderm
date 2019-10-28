package server

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestMerge(t *testing.T) {
	config := GetBasicConfig()
	config.NewStorageLayer = true
	config.StorageMemoryThreshold = 20
	config.StorageShardThreshold = 20
	c := GetPachClient(t, config)
	repo := "test"
	branch := "master"
	require.NoError(t, c.CreateRepo(repo))
	var commit *pfs.Commit
	var err error
	for i := 0; i < 10; i++ {
		commit, err = c.StartCommit(repo, branch)
		require.NoError(t, err)
		pfc, err := c.NewPutFileClient()
		require.NoError(t, err)
		for j := 0; j < 10; j++ {
			s := strconv.Itoa(i*10 + j)
			_, err := pfc.PutFile(repo, commit.ID, "/file"+s, strings.NewReader(s))
			require.NoError(t, err)
		}
		require.NoError(t, pfc.Close())
		require.NoError(t, c.FinishCommit(repo, commit.ID))
	}
	buf := &bytes.Buffer{}
	require.NoError(t, c.GetFile(repo, commit.ID, "/file0", 0, 0, buf))
	require.Equal(t, "0", string(buf.Bytes()))
	buf.Reset()
	require.NoError(t, c.GetFile(repo, commit.ID, "/file50", 0, 0, buf))
	require.Equal(t, "50", string(buf.Bytes()))
	buf.Reset()
	require.NoError(t, c.GetFile(repo, commit.ID, "/file99", 0, 0, buf))
	require.Equal(t, "99", string(buf.Bytes()))
}
