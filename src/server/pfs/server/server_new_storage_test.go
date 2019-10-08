package server

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
)

// (bryce) this is a very basic integration / distributed merge test
// that is hacky and will be updated as the integration of the new storage
// layer proceeds.
// the configuration setup will make it so that there will be multiple
// file set parts that will be merged across multiple merge shards.
func TestMerge(t *testing.T) {
	newStorageLayer = true
	memThreshold = 20
	fileset.ShardThreshold = 20
	c := GetPachClient(t)
	pfc, err := c.NewPutFileClient()
	require.NoError(t, err)
	repo := ""
	commit := ""
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		_, err := pfc.PutFile(repo, commit, "/file"+s, strings.NewReader(s))
		require.NoError(t, err)
	}
	require.NoError(t, pfc.Close())
	buf := &bytes.Buffer{}
	require.NoError(t, c.GetFile(repo, commit, "/file0", 0, 0, buf))
	require.Equal(t, "0", string(buf.Bytes()))
	buf.Reset()
	require.NoError(t, c.GetFile(repo, commit, "/file50", 0, 0, buf))
	require.Equal(t, "50", string(buf.Bytes()))
	buf.Reset()
	require.NoError(t, c.GetFile(repo, commit, "/file99", 0, 0, buf))
	require.Equal(t, "99", string(buf.Bytes()))
}
