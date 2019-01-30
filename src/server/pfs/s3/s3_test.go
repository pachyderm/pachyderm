package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	minio "github.com/minio/minio-go"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

var nextPort uint32 = 40000

func serve(t *testing.T) (*http.Server, uint16) {
	pc := server.GetPachClient(t)
	require.NoError(t, pc.CreateRepo("repo"))
	_, err := pc.PutFile("repo", "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	port := uint16(atomic.AddUint32(&nextPort, 1))
	srv := Server(pc, port)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Fatalf("http server returned an error: %v", err)
		}
	}()

	// Wait for the server to start
	c := &http.Client{}
	for i := 0; i < 100; i++ {
		res, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/repo/", port))
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
	srv, port := serve(t)
	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)

	fetchedContent, err := getObject(c, "repo", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)

	require.NoError(t, srv.Close())
}
