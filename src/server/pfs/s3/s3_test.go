package s3

// Tests for the PFS' S3 emulation API. Note that, in calls to
// `tu.UniqueString`, all lowercase characters are used, unlike in other
// tests. This is in order to generate repo names that are also valid bucket
// names. Otherwise minio complains that the bucket name is not valid.

import (
	"fmt"
	// "io"
	"io/ioutil"
	"net/http"
	// "os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	minio "github.com/minio/minio-go"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func serve(t *testing.T, pc *client.APIClient) (*http.Server, *minio.Client) {
	t.Helper()

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
		res, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
		if err != nil {
			return err
		} else if res.StatusCode != 200 {
			return fmt.Errorf("Unexpected status code: %d", res.StatusCode)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	c, err := minio.New(fmt.Sprintf("127.0.0.1:%d", port), "id", "secret", false)
	require.NoError(t, err)
	return srv, c
}

func getObject(t *testing.T, c *minio.Client, repo, branch, file string) (string, error) {
	t.Helper()

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

func checkListObjects(t *testing.T, ch <-chan minio.ObjectInfo, startTime time.Time, endTime time.Time, expectedFiles []string, expectedDirs []string) {
	t.Helper()

	// sort expected files/dirs, as the S3 gateway should always return
	// results in sorted order
	sort.Strings(expectedFiles)
	sort.Strings(expectedDirs)

	actualFiles := []minio.ObjectInfo{}
	actualDirs := []minio.ObjectInfo{}
	for obj := range ch {
		require.NoError(t, obj.Err)

		if strings.HasSuffix(obj.Key, "/") {
			actualDirs = append(actualDirs, obj)
		} else {
			actualFiles = append(actualFiles, obj)
		}
	}

	require.Equal(t, len(expectedFiles), len(actualFiles), "unexpected number of files")
	require.Equal(t, len(expectedDirs), len(actualDirs), "unexpected number of dirs")

	for i, expectedFilename := range expectedFiles {
		actualFile := actualFiles[i]
		require.Equal(t, expectedFilename, actualFile.Key)
		require.Equal(t, "", actualFile.ETag, fmt.Sprintf("unexpected etag for %s", expectedFilename))
		expectedLen := int64(len(filepath.Base(expectedFilename)) + 1)
		require.Equal(t, expectedLen, actualFile.Size, fmt.Sprintf("unexpected file length for %s", expectedFilename))
		require.True(t, startTime.Before(actualFile.LastModified), fmt.Sprintf("unexpected last modified for %s", expectedFilename))
		require.True(t, endTime.After(actualFile.LastModified), fmt.Sprintf("unexpected last modified for %s", expectedFilename))
	}

	for i, expectedDirname := range expectedDirs {
		actualDir := actualDirs[i]
		require.Equal(t, expectedDirname, actualDir.Key)
		require.Equal(t, "", actualDir.ETag, fmt.Sprintf("unexpected etag for %s", expectedDirname))
		require.Equal(t, int64(0), actualDir.Size)
		require.True(t, actualDir.LastModified.IsZero(), fmt.Sprintf("unexpected last modified for %s", expectedDirname))
	}
}

func putListFileTestObject(t *testing.T, pc *client.APIClient, repo string, commitID string, dir string, i int) {
	t.Helper()
	_, err := pc.PutFile(
		repo,
		commitID,
		fmt.Sprintf("%s%d", dir, i),
		strings.NewReader(fmt.Sprintf("%d\n", i)),
	)
	require.NoError(t, err)
}

func nonServerError(t *testing.T, err error) {
	t.Helper()
	require.YesError(t, err)
	require.NotEqual(t, "500 Internal Server Error", err.Error(), "expected a non-500 error")
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

	fetchedContent, err := getObject(t, c, repo, "master", "file")
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

	fetchedContent, err := getObject(t, c, repo, "branch", "file")
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

	_, err := c.PutObject(repo, "branch/file", strings.NewReader("content1"), "text/plain")
	require.NoError(t, err)

	// this should act as a PFS PutFileOverwrite
	_, err = c.PutObject(repo, "branch/file", strings.NewReader("content2"), "text/plain")
	require.NoError(t, err)

	fetchedContent, err := getObject(t, c, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content2", fetchedContent)

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

// // Tests inserting and getting files over 64mb in size
// func TestLargeObjects(t *testing.T) {
// 	pc := server.GetPachClient(t)
// 	srv, c := serve(t, pc)

// 	// test repos: repo1 exists, repo2 does not
// 	repo1 := tu.UniqueString("testlargeobject1")
// 	repo2 := tu.UniqueString("testlargeobject2")
// 	require.NoError(t, pc.CreateRepo(repo1))

// 	// create a temporary file to put ~65mb of contents into it
// 	// TODO: see how much faster it is to just hold everything in memory
// 	bytesWritten := 0
// 	inputFile, err := ioutil.TempFile("", "pachyderm-test-large-objects-input-*")
// 	require.NoError(t, err)
// 	defer os.Remove(inputFile.Name())
// 	for i := 0; i < 1363149; i++ {
// 		n, err := inputFile.WriteString("no tv and no beer make homer something something.\n")
// 		require.NoError(t, err)
// 		bytesWritten += n
// 	}

// 	// make sure we wrote ~65mb
// 	if bytesWritten < 68157450 {
// 		t.Errorf("too few bytes written to %s: %d", inputFile.Name(), bytesWritten)
// 	}

// 	// first ensure that putting into a repo that doesn't exist triggers an
// 	// error
// 	_, err = c.FPutObject(repo2, "file", inputFile.Name(), "text/plain")
// 	nonServerError(t, err)

// 	// now try putting into a legit repo
// 	l, err := c.FPutObject(repo1, "file", inputFile.Name(), "text/plain")
// 	require.NoError(t, err)
// 	require.Equal(t, bytesWritten, l)

// 	// create a file to write the results back to
// 	outputFile, err := ioutil.TempFile("", "pachyderm-test-large-objects-output-*")
// 	require.NoError(t, err)
// 	defer os.Remove(outputFile.Name())

// 	// try getting an object that does not exist
// 	err = c.FGetObject(repo2, "file", outputFile.Name())
// 	nonServerError(t, err)
// 	bytes, err := ioutil.ReadFile(outputFile.Name())
// 	require.NoError(t, err)
// 	require.Equal(t, 0, len(bytes))

// 	// get the file that does exist
// 	err = c.FGetObject(repo1, "file", outputFile.Name())
// 	require.NoError(t, err)

// 	// compare the input file and output file to ensure they're the same
// 	b1 := make([]byte, 65536)
// 	b2 := make([]byte, 65536)
// 	inputFile.Seek(0, 0)
// 	outputFile.Seek(0, 0)
// 	for {
// 		n1, err1 := inputFile.Read(b1)
// 		n2, err2 := outputFile.Read(b2)

// 		require.Equal(t, n1, n2)

// 		if err1 == io.EOF && err2 == io.EOF {
// 			break
// 		}

// 		require.NoError(t, err1)
// 		require.NoError(t, err2)
// 		require.Equal(t, b1, b2)
// 	}

// 	require.NoError(t, srv.Close())
// }

func TestGetObjectNoHead(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testgetobjectnohead")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	_, err := getObject(t, c, repo, "branch", "file")
	nonServerError(t, err)

	require.NoError(t, srv.Close())
}

func TestGetObjectNoBranch(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testgetobjectnobranch")
	require.NoError(t, pc.CreateRepo(repo))

	_, err := getObject(t, c, repo, "branch", "file")
	nonServerError(t, err)
	require.Equal(t, err.Error(), "The specified key does not exist.")
	require.NoError(t, srv.Close())
}

func TestGetObjectNoRepo(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testgetobjectnorepo")
	_, err := getObject(t, c, repo, "master", "file")
	nonServerError(t, err)
	require.Equal(t, err.Error(), "The specified bucket does not exist.")
	require.NoError(t, srv.Close())
}

func TestMakeBucket(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)
	repo := tu.UniqueString("testmakebucket")
	require.NoError(t, c.MakeBucket(repo, ""))
	_, err := pc.InspectRepo(repo)
	require.NoError(t, err)
	require.NoError(t, srv.Close())
}

func TestMakeBucketWithRegion(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)
	repo := tu.UniqueString("testmakebucketwithregion")
	require.NoError(t, c.MakeBucket(repo, "us-east-1"))
	_, err := pc.InspectRepo(repo)
	require.NoError(t, err)
	require.NoError(t, srv.Close())
}

func TestMakeBucketRedundant(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)
	repo := tu.UniqueString("testmakebucketredundant")
	require.NoError(t, c.MakeBucket(repo, ""))
	nonServerError(t, c.MakeBucket(repo, ""))
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
	nonServerError(t, c.RemoveBucket(repo2))

	require.NoError(t, srv.Close())
}

func TestListObjectsPaginated(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	// create a bunch of files - enough to require the use of paginated
	// requests when browsing all files. One file will be included on a
	// separate branch to ensure it's not returned when querying against the
	// master branch.
	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	repo := tu.UniqueString("testlistobjectspaginated")
	require.NoError(t, pc.CreateRepo(repo))
	commit, err := pc.StartCommit(repo, "master")
	require.NoError(t, err)
	for i := 0; i <= 1000; i++ {
		putListFileTestObject(t, pc, repo, commit.ID, "", i)
	}
	for i := 0; i < 10; i++ {
		putListFileTestObject(t, pc, repo, commit.ID, "dir/", i)
		require.NoError(t, err)
	}
	putListFileTestObject(t, pc, repo, "branch", "", 1001)
	require.NoError(t, pc.FinishCommit(repo, commit.ID))
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	// Request that will list all files in master's root
	ch := c.ListObjects(repo, "master/", false, make(chan struct{}))
	expectedFiles := []string{}
	for i := 0; i <= 1000; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("master/%d", i))
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{"master/dir/"})

	// Request that will list all files in master starting with 1
	ch = c.ListObjects(repo, "master/1", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i <= 1000; i++ {
		file := fmt.Sprintf("master/%d", i)
		if strings.HasPrefix(file, "master/1") {
			expectedFiles = append(expectedFiles, file)
		}
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	// Request that will list all files in a directory in master
	ch = c.ListObjects(repo, "master/dir/", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i < 10; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("master/dir/%d", i))
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	require.NoError(t, srv.Close())
}

func TestListObjectsBranches(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testlistobjectsbranches")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))
	require.NoError(t, pc.CreateBranch(repo, "emptybranch", "", nil))

	// put files on `master` and `branch`, but not on `emptybranch`. As a
	// result, `emptybranch` will have no head commit.
	putListFileTestObject(t, pc, repo, "master", "", 1)
	putListFileTestObject(t, pc, repo, "branch", "", 2)

	// Request that will list branches as common prefixes
	ch := c.ListObjects(repo, "", false, make(chan struct{}))
	checkListObjects(t, ch, time.Time{}, time.Time{}, []string{}, []string{"branch/", "master/"})
	ch = c.ListObjects(repo, "master", false, make(chan struct{}))
	checkListObjects(t, ch, time.Time{}, time.Time{}, []string{}, []string{"master/"})

	require.NoError(t, srv.Close())
}

func TestListObjectsHeadlessBranch(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	repo := tu.UniqueString("testlistobjectsheadlessbranch")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "emptybranch", "", nil))

	// Request into branch that has no head
	ch := c.ListObjects(repo, "emptybranch/", false, make(chan struct{}))
	checkListObjects(t, ch, time.Time{}, time.Time{}, []string{}, []string{})

	require.NoError(t, srv.Close())
}

func TestListObjectsRecursive(t *testing.T) {
	pc := server.GetPachClient(t)
	srv, c := serve(t, pc)

	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	repo := tu.UniqueString("testlistobjectsrecursive")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))
	require.NoError(t, pc.CreateBranch(repo, "emptybranch", "", nil))
	commit, err := pc.StartCommit(repo, "master")
	require.NoError(t, err)
	putListFileTestObject(t, pc, repo, commit.ID, "", 0)
	putListFileTestObject(t, pc, repo, commit.ID, "rootdir/", 1)
	putListFileTestObject(t, pc, repo, commit.ID, "rootdir/subdir/", 2)
	putListFileTestObject(t, pc, repo, "branch", "", 3)
	require.NoError(t, pc.FinishCommit(repo, commit.ID))
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	// Request that will list all files across all branches
	expectedFiles := []string{"master/0", "master/rootdir/1", "master/rootdir/subdir/2", "branch/3"}
	expectedDirs := []string{"master/", "master/rootdir/", "master/rootdir/subdir/", "branch/"}
	ch := c.ListObjects(repo, "", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Requests that will list all files and dirs in master, including master itself
	expectedFiles = []string{"master/0", "master/rootdir/1", "master/rootdir/subdir/2"}
	expectedDirs = []string{"master/", "master/rootdir/", "master/rootdir/subdir/"}
	ch = c.ListObjects(repo, "mas", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)
	ch = c.ListObjects(repo, "master", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Request that will list all files and dirs in master, excluding master itself
	expectedDirs = []string{"master/rootdir/", "master/rootdir/subdir/"}
	ch = c.ListObjects(repo, "master/", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Requests that will list all files in rootdir, including rootdir itself
	expectedFiles = []string{"master/rootdir/1", "master/rootdir/subdir/2"}
	ch = c.ListObjects(repo, "master/r", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)
	ch = c.ListObjects(repo, "master/rootdir", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Request that will list all files in rootdir, excluding rootdir itself
	expectedDirs = []string{"master/rootdir/subdir/"}
	ch = c.ListObjects(repo, "master/rootdir/", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Requests that will list all files in subdir, including subdir itself
	expectedFiles = []string{"master/rootdir/subdir/2"}
	ch = c.ListObjects(repo, "master/rootdir/s", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)
	ch = c.ListObjects(repo, "master/rootdir/subdir", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Request that will list all files in subdir, excluding subdir itself
	expectedDirs = []string{}
	ch = c.ListObjects(repo, "master/rootdir/subdir/", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)
	ch = c.ListObjects(repo, "master/rootdir/subdir/2", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	require.NoError(t, srv.Close())
}
