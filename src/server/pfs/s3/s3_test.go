package s3

// Tests for the PFS' S3 emulation API. Note that, in calls to
// `tu.UniqueString`, all lowercase characters are used, unlike in other
// tests. This is in order to generate repo names that are also valid bucket
// names. Otherwise minio complains that the bucket name is not valid.

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	minio "github.com/minio/minio-go"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func serve(t *testing.T, pc *client.APIClient) (*http.Server, uint16) {
	port := tu.UniquePort()
	srv := Server(pc, port)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Fatalf("http server returned an error: %v", err)
		}
	}()

	// Wait for the server to start
	require.NoError(t, backoff.Retry(func() error {
		c := &http.Client{}
		res, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/_ping", port))
		if err != nil {
			return err
		} else if res.StatusCode != 200 {
			return fmt.Errorf("Unexpected status code: %d", res.StatusCode)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	return srv, port
}

func getObject(c *minio.Client, repo, branch, file string) (string, error) {
	obj, err := c.GetObject(repo, fmt.Sprintf("%s/%s", branch, file))
	if err != nil {
		return "", err
	}
	defer func() { err = obj.Close() }()
	bytes, err := ioutil.ReadAll(obj)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}


func TestGetFile(t *testing.T) {
	repo := tu.UniqueString("testgetfile")
	pc := server.GetPachClient(t)
	require.NoError(t, pc.CreateRepo(repo))
	_, err := pc.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	srv, port := serve(t, pc)
	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)

	fetchedContent, err := getObject(c, repo, "master", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)

	require.NoError(t, srv.Close())
}

func TestGetFileInBranch(t *testing.T) {
	repo := tu.UniqueString("testgetfileinbranch")
	pc := server.GetPachClient(t)
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))
	_, err := pc.PutFile(repo, "branch", "file", strings.NewReader("content"))
	require.NoError(t, err)

	srv, port := serve(t, pc)
	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)

	fetchedContent, err := getObject(c, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)

	require.NoError(t, srv.Close())
}

func TestNonExistingBranch(t *testing.T) {
	repo := tu.UniqueString("testnonexistingbranch")
	pc := server.GetPachClient(t)
	require.NoError(t, pc.CreateRepo(repo))

	srv, port := serve(t, pc)
	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)

	_, err = getObject(c, repo, "branch", "file")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "The specified key does not exist.")
	require.NoError(t, srv.Close())
}

func TestNonExistingRepo(t *testing.T) {
	repo := tu.UniqueString("testnonexistingrepo")
	pc := server.GetPachClient(t)

	srv, port := serve(t, pc)
	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)

	_, err = getObject(c, repo, "master", "file")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "The specified bucket does not exist.")
	require.NoError(t, srv.Close())
}

func TestPutObject(t *testing.T) {
	repo := tu.UniqueString("testputobject")
	pc := server.GetPachClient(t)
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	srv, port := serve(t, pc)
	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)

	_, err = c.PutObject(repo, fmt.Sprintf("%s/%s", "branch", "file"), strings.NewReader("content"), "text/plain")
	require.NoError(t, err)

	fetchedContent, err := getObject(c, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)

	require.NoError(t, srv.Close())
}
