package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	minio "github.com/minio/minio-go"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func serve(t *testing.T, pc *client.APIClient, repo, branch string) (*http.Server, uint16) {
	port := tu.UniquePort()
	srv := Server(pc, port)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Fatalf("http server returned an error: %v", err)
		}
	}()

	// Wait for the server to start
	c := &http.Client{}
	for i := 0; i < 50; i++ {
		res, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/%s.%s/", port, branch, repo))
		if err == nil && res.StatusCode == 200 {
			return srv, port
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("server failed to start after a few seconds")
	return nil, 0
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


func TestGetFile(t *testing.T) {
	repo := tu.UniqueString("TestGetFile")
	pc := server.GetPachClient(t)
	require.NoError(t, pc.CreateRepo(repo))
	_, err := pc.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	srv, port := serve(t, pc, repo, "master")
	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)

	fetchedContent, err := getObject(c, repo, "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)

	require.NoError(t, srv.Close())
}

func TestGetFileInBranch(t *testing.T) {
	repo := tu.UniqueString("TestGetFileInBranch")
	pc := server.GetPachClient(t)
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))
	_, err := pc.PutFile(repo, "branch", "file", strings.NewReader("content"))
	require.NoError(t, err)

	srv, port := serve(t, pc, repo, "branch")
	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)

	fetchedContent, err := getObject(c, fmt.Sprintf("branch.%s", repo), "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)

	require.NoError(t, srv.Close())
}
