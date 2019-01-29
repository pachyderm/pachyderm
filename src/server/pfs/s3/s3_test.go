package main

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

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
	c, err := minio.New("127.0.0.1:30655", "id", "secret", false)
	require.NoError(t, err)

	// Try to fetch the contents a few times, in case the s3 proxy is still
	// booting up
	var fetchedContent string
	for i := 0; i < 10; i++ {
		fetchedContent, err = getObject(c, repo, file)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	require.NoError(t, err)
	require.Equal(t, content, fetchedContent)
}

func getObject(c *minio.Client, repo, file string) (string, error) {
	obj, err := c.GetObject(repo, file)
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(obj)
	if err != nil {
		return "", err
	}
	if err = obj.Close(); err != nil {
		return "", err
	}
	return string(bytes), nil
}
