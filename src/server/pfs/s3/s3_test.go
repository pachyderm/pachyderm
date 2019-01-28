package main

import (
	"io/ioutil"
	"strings"
	"testing"

	minio "github.com/minio/minio-go"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestSimple(t *testing.T) {
	pc := server.GetPachClient(t)
	repo := "repo"
	require.NoError(t, pc.CreateRepo(repo))
	file := "file"
	content := "content"
	_, err := pc.PutFile(repo, "master", file, strings.NewReader(content))
	require.NoError(t, err)
	go func() { Serve(pc, 30655) }()
	c, err := minio.NewWithRegion("127.0.0.1:30655", "id", "secret", false, "region")
	require.NoError(t, err)
	obj, err := c.GetObject(repo, file)
	require.NoError(t, err)
	bytes, err := ioutil.ReadAll(obj)
	require.NoError(t, err)
	require.Equal(t, content, string(bytes))
	require.NoError(t, obj.Close())
}
