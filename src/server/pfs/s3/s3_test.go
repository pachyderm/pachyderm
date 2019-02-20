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
	"time"

	// log "github.com/sirupsen/logrus"
	minio "github.com/minio/minio-go"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func serve(t *testing.T, pc *client.APIClient) (*http.Server, *minio.Client) {
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
		} else if res.StatusCode != 200 && res.StatusCode != 204 {
			return fmt.Errorf("Unexpected status code: %d", res.StatusCode)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)
	return srv, c
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

func TestListBuckets(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	startTime := time.Now()
	repo1 := tu.UniqueString("testlistbuckets1")
	require.NoError(t, pc.CreateRepo(repo1))
	repo2 := tu.UniqueString("testlistbuckets2")
	require.NoError(t, pc.CreateRepo(repo2))
	endTime := time.Now()

	buckets, err := c.ListBuckets()
	require.NoError(t, err)
	require.Equal(t, 2, len(buckets))

	for _, bucket := range buckets {
		require.EqualOneOf(t, []string{repo1, repo2}, bucket.Name)
		require.True(t, startTime.Before(bucket.CreationDate))
		require.True(t, endTime.After(bucket.CreationDate))
	}

	require.NoError(t, srv.Close())
}

func TestGetObject(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testgetobject")
	require.NoError(t, pc.CreateRepo(repo))
	_, err := pc.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	fetchedContent, err := getObject(c, repo, "master", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)

	require.NoError(t, srv.Close())
}

func TestGetObjectInBranch(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testgetobjectinbranch")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))
	_, err := pc.PutFile(repo, "branch", "file", strings.NewReader("content"))
	require.NoError(t, err)

	fetchedContent, err := getObject(c, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)

	require.NoError(t, srv.Close())
}

func TestStatObject(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("teststatobject")
	require.NoError(t, pc.CreateRepo(repo))
	_, err := pc.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	_, err = pc.PutFileOverwrite(repo, "master", "file", strings.NewReader("new-content"), 0)
	require.NoError(t, err)
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	info, err := c.StatObject(repo, "master/file")
	require.NoError(t, err)
	require.True(t, startTime.Before(info.LastModified))
	require.True(t, endTime.After(info.LastModified))
	require.Equal(t, "", info.ETag) //etags aren't returned by our API
	require.Equal(t, "text/plain; charset=utf-8", info.ContentType)
	require.Equal(t, int64(11), info.Size)

	require.NoError(t, srv.Close())
}

func TestPutObject(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testputobject")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	_, err := c.PutObject(repo, "branch/file", strings.NewReader("content"), "text/plain")
	require.NoError(t, err)

	fetchedContent, err := getObject(c, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)

	require.NoError(t, srv.Close())
}

func TestRemoveObject(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testremoveobject")
	require.NoError(t, pc.CreateRepo(repo))
	_, err := pc.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	// as per PFS semantics, the second delete should be a no-op
	require.NoError(t, c.RemoveObject(repo, "master/file"))
	require.NoError(t, c.RemoveObject(repo, "master/file"))

	require.NoError(t, srv.Close())
}

// Tries to get an object on a branch that does not have a head
func TestNonExistingHead(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testputobject2")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	_, err := getObject(c, repo, "branch", "file")
	require.YesError(t, err)

	require.NoError(t, srv.Close())
}

func TestNonExistingBranch(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testnonexistingbranch")
	require.NoError(t, pc.CreateRepo(repo))

	_, err := getObject(c, repo, "branch", "file")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "The specified key does not exist.")
	require.NoError(t, srv.Close())
}

func TestNonExistingRepo(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testnonexistingrepo")
	_, err := getObject(c, repo, "master", "file")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "The specified bucket does not exist.")
	require.NoError(t, srv.Close())
}

func TestMakeBucket(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo1 := tu.UniqueString("testmakebucket1")
	require.NoError(t, c.MakeBucket(repo1, ""))

	repo2 := tu.UniqueString("testmakebucket2")
	require.NoError(t, c.MakeBucket(repo2, "us-east-1"))

	_, err := pc.InspectRepo(repo1)
	require.NoError(t, err)

	_, err = pc.InspectRepo(repo2)
	require.NoError(t, err)

	require.NoError(t, srv.Close())
}

func TestBucketExists(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo1 := tu.UniqueString("testbucketexists1")
	require.NoError(t, pc.CreateRepo(repo1))
	exists, err := c.BucketExists(repo1)
	require.NoError(t, err)
	require.True(t, exists)

	repo2 := tu.UniqueString("testbucketexists1")
	exists, err = c.BucketExists(repo2)
	require.NoError(t, err)
	require.False(t, exists)
	
	require.NoError(t, srv.Close())
}

func TestRemoveBucket(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo1 := tu.UniqueString("testremovebucket1")
	require.NoError(t, pc.CreateRepo(repo1))
	require.NoError(t, c.RemoveBucket(repo1))

	repo2 := tu.UniqueString("testremovebucket2")
	require.YesError(t, c.RemoveBucket(repo2))
	
	require.NoError(t, srv.Close())
}
