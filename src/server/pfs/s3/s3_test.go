package s3

// Tests for the PFS' S3 emulation API. Note that, in calls to
// `tu.UniqueString`, all lowercase characters are used, unlike in other
// tests. This is in order to generate repo names that are also valid bucket
// names. Otherwise minio complains that the bucket name is not valid.

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	"github.com/sirupsen/logrus"
)

func serve(t *testing.T, multipartDir string) (*http.Server, *client.APIClient, *minio.Client) {
	t.Helper()

	port := tu.UniquePort()
	pc := server.GetPachClient(t)

	// note: this logrus writer is not closed. This shouldn't be a problem for
	// tests.
	srv := Server(pc, port, logrus.StandardLogger().Writer(), multipartDir)

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
	return srv, pc, c
}

func getObject(t *testing.T, c *minio.Client, repo, branch, file string) (string, error) {
	t.Helper()

	obj, err := c.GetObject(fmt.Sprintf("%s-%s", repo, branch), file)
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

func bucketNotFoundError(t *testing.T, err error) {
	t.Helper()
	require.YesError(t, err)
	require.Equal(t, "The specified bucket does not exist.", err.Error())
}

func keyNotFoundError(t *testing.T, err error) {
	t.Helper()
	require.YesError(t, err)
	require.Equal(t, "The specified key does not exist.", err.Error())
}

func fileHash(t *testing.T, name string) (int64, []byte) {
	t.Helper()

	f, err := os.Open(name)
	require.NoError(t, err)

	fi, err := f.Stat()
	require.NoError(t, err)

	hash := md5.New()
	_, err = io.Copy(hash, f)
	require.NoError(t, err)

	// make sure it's not the hash of an empty string
	hashSum := hash.Sum(nil)
	require.NotEqual(t, hashSum, []byte{0xd4, 0x1d, 0x8c, 0xd9, 0x8f, 0x0, 0xb2, 0x4, 0xe9, 0x80, 0x9, 0x98, 0xec, 0xf8, 0x42, 0x7e})

	return fi.Size(), hashSum
}

func TestListBuckets(t *testing.T) {
	srv, pc, c := serve(t, "")

	startTime := time.Now()
	repo1 := tu.UniqueString("testlistbuckets1")
	require.NoError(t, pc.CreateRepo(repo1))
	repo2 := tu.UniqueString("testlistbuckets2")
	require.NoError(t, pc.CreateRepo(repo2))
	endTime := time.Now()

	require.NoError(t, pc.CreateBranch(repo1, "master", "", nil))
	require.NoError(t, pc.CreateBranch(repo1, "branch", "", nil))
	require.NoError(t, pc.CreateBranch(repo2, "branch", "", nil))
	expectedBranches := []string{repo1, fmt.Sprintf("%s-branch", repo1), fmt.Sprintf("%s-branch", repo2)}
	buckets, err := c.ListBuckets()
	require.NoError(t, err)
	require.Equal(t, 3, len(buckets))
	for _, bucket := range buckets {
		require.EqualOneOf(t, expectedBranches, bucket.Name)
		require.True(t, startTime.Before(bucket.CreationDate))
		require.True(t, endTime.After(bucket.CreationDate))
	}

	require.NoError(t, srv.Close())
}

func TestListBucketsBranchless(t *testing.T) {
	srv, pc, c := serve(t, "")

	repo1 := tu.UniqueString("testlistbucketsbranchless1")
	require.NoError(t, pc.CreateRepo(repo1))
	repo2 := tu.UniqueString("testlistbucketsbranchless2")
	require.NoError(t, pc.CreateRepo(repo2))

	// should be 0 since no branches have been made yet
	buckets, err := c.ListBuckets()
	require.NoError(t, err)
	require.Equal(t, 0, len(buckets))

	require.NoError(t, srv.Close())
}

func TestGetObject(t *testing.T) {
	srv, pc, c := serve(t, "")

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
	srv, pc, c := serve(t, "")

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
	srv, pc, c := serve(t, "")

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

	info, err := c.StatObject(repo, "file")
	require.NoError(t, err)
	require.True(t, startTime.Before(info.LastModified))
	require.True(t, endTime.After(info.LastModified))
	require.Equal(t, "", info.ETag) //etags aren't returned by our API
	require.Equal(t, "text/plain; charset=utf-8", info.ContentType)
	require.Equal(t, int64(11), info.Size)

	require.NoError(t, srv.Close())
}

func TestPutObject(t *testing.T) {
	srv, pc, c := serve(t, "")

	repo := tu.UniqueString("testputobject")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	_, err := c.PutObject(fmt.Sprintf("%s-branch", repo), "file", strings.NewReader("content1"), "text/plain")
	require.NoError(t, err)

	// this should act as a PFS PutFileOverwrite
	_, err = c.PutObject(fmt.Sprintf("%s-branch", repo), "file", strings.NewReader("content2"), "text/plain")
	require.NoError(t, err)

	fetchedContent, err := getObject(t, c, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content2", fetchedContent)

	require.NoError(t, srv.Close())
}

func TestRemoveObject(t *testing.T) {
	srv, pc, c := serve(t, "")

	repo := tu.UniqueString("testremoveobject")
	require.NoError(t, pc.CreateRepo(repo))
	_, err := pc.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	// as per PFS semantics, the second delete should be a no-op
	require.NoError(t, c.RemoveObject(repo, "file"))
	require.NoError(t, c.RemoveObject(repo, "file"))

	// make sure the object no longer exists
	_, err = getObject(t, c, repo, "master", "file")
	keyNotFoundError(t, err)

	require.NoError(t, srv.Close())
}

// Tests inserting and getting files over 64mb in size
func TestLargeObjects(t *testing.T) {
	multipartDir, err := ioutil.TempDir("", "pachyderm-test-s3gateway-multipart")
	require.NoError(t, err)
	defer os.RemoveAll(multipartDir)
	srv, pc, c := serve(t, multipartDir)

	// test repos: repo1 exists, repo2 does not
	repo1 := tu.UniqueString("testlargeobject1")
	repo2 := tu.UniqueString("testlargeobject2")
	require.NoError(t, pc.CreateRepo(repo1))
	require.NoError(t, pc.CreateBranch(repo1, "master", "", nil))

	// create a temporary file to put ~65mb of contents into it
	inputFile, err := ioutil.TempFile("", "pachyderm-test-large-objects-input-*")
	require.NoError(t, err)
	defer os.Remove(inputFile.Name())
	n, err := inputFile.WriteString(strings.Repeat("no tv and no beer make homer something something.\n", 1363149))
	require.NoError(t, err)
	require.Equal(t, n, 68157450)
	require.NoError(t, inputFile.Sync())

	// first ensure that putting into a repo that doesn't exist triggers an
	// error
	_, err = c.FPutObject(repo2, "file", inputFile.Name(), "text/plain")
	bucketNotFoundError(t, err)

	// now try putting into a legit repo
	l, err := c.FPutObject(repo1, "file", inputFile.Name(), "text/plain")
	require.Equal(t, err, io.EOF)
	require.Equal(t, int(l), 68157450)

	// try getting an object that does not exist
	err = c.FGetObject(repo2, "file", "foo")
	bucketNotFoundError(t, err)

	// get the file that does exist
	outputFile, err := ioutil.TempFile("", "pachyderm-test-large-objects-output-*")
	require.NoError(t, err)
	defer os.Remove(outputFile.Name())
	err = c.FGetObject(repo1, "file", outputFile.Name())
	require.NoError(t, err)

	// compare the files and ensure they're the same
	// NOTE: Because minio's `FGetObject` does a rename from a buffer file
	// to the given filepath, `outputFile` will refer to an empty, overwritten
	// file. We can still use `outputFile.Name()` though.
	inputFileSize, inputFileHash := fileHash(t, inputFile.Name())
	outputFileSize, outputFileHash := fileHash(t, inputFile.Name())
	require.Equal(t, inputFileSize, outputFileSize)
	require.Equal(t, inputFileHash, outputFileHash)

	require.NoError(t, srv.Close())
}

func TestGetObjectNoHead(t *testing.T) {
	srv, pc, c := serve(t, "")

	repo := tu.UniqueString("testgetobjectnohead")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	_, err := getObject(t, c, repo, "branch", "file")
	keyNotFoundError(t, err)

	require.NoError(t, srv.Close())
}

func TestGetObjectNoBranch(t *testing.T) {
	srv, pc, c := serve(t, "")

	repo := tu.UniqueString("testgetobjectnobranch")
	require.NoError(t, pc.CreateRepo(repo))

	_, err := getObject(t, c, repo, "branch", "file")
	bucketNotFoundError(t, err)
	require.NoError(t, srv.Close())
}

func TestGetObjectNoRepo(t *testing.T) {
	srv, _, c := serve(t, "")

	repo := tu.UniqueString("testgetobjectnorepo")
	_, err := getObject(t, c, repo, "master", "file")
	bucketNotFoundError(t, err)
	require.NoError(t, srv.Close())
}

func TestMakeBucket(t *testing.T) {
	srv, pc, c := serve(t, "")
	repo := tu.UniqueString("testmakebucket")
	require.NoError(t, c.MakeBucket(repo, ""))

	repoInfo, err := pc.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "master")

	require.NoError(t, srv.Close())
}

func TestMakeBucketWithBranch(t *testing.T) {
	srv, pc, c := serve(t, "")
	repo := tu.UniqueString("testmakebucketwithbranch")
	require.NoError(t, c.MakeBucket(fmt.Sprintf("%s-branch", repo), ""))

	repoInfo, err := pc.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "branch")

	require.NoError(t, srv.Close())
}

func TestMakeBucketWithRegion(t *testing.T) {
	srv, pc, c := serve(t, "")
	repo := tu.UniqueString("testmakebucketwithregion")
	require.NoError(t, c.MakeBucket(repo, "us-east-1"))
	_, err := pc.InspectRepo(repo)
	require.NoError(t, err)
	require.NoError(t, srv.Close())
}

func TestMakeBucketRedundant(t *testing.T) {
	srv, _, c := serve(t, "")
	repo := tu.UniqueString("testmakebucketredundant")
	require.NoError(t, c.MakeBucket(repo, ""))
	err := c.MakeBucket(repo, "")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "There is already a repo with that name.")
	require.NoError(t, srv.Close())
}

func TestMakeBucketDifferentBranches(t *testing.T) {
	srv, _, c := serve(t, "")
	repo := tu.UniqueString("testmakebucketdifferentbranches")

	require.NoError(t, c.MakeBucket(repo, ""))

	// this should error because the last bucket creation implicitly made the
	// master branch
	err := c.MakeBucket(fmt.Sprintf("%s-master", repo), "")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "There is already a repo with that name.")

	// this should not error because it's a separate branch
	err = c.MakeBucket(fmt.Sprintf("%s-branch", repo), "")
	require.NoError(t, err)

	require.NoError(t, srv.Close())
}

func TestBucketExists(t *testing.T) {
	srv, pc, c := serve(t, "")
	repo := tu.UniqueString("testbucketexists")

	exists, err := c.BucketExists(repo)
	require.NoError(t, err)
	require.False(t, exists)

	// repo exists, but branch doesn't: should be false
	require.NoError(t, pc.CreateRepo(repo))
	exists, err = c.BucketExists(repo)
	require.NoError(t, err)
	require.False(t, exists)

	// repo and branch exists: should be true
	require.NoError(t, pc.CreateBranch(repo, "master", "", nil))
	exists, err = c.BucketExists(repo)
	require.NoError(t, err)
	require.True(t, exists)

	// repo exists, but branch doesn't: should be false
	exists, err = c.BucketExists(fmt.Sprintf("%s-branch", repo))
	require.NoError(t, err)
	require.False(t, exists)

	// repo and branch exists: should be true
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))
	exists, err = c.BucketExists(fmt.Sprintf("%s-branch", repo))
	require.NoError(t, err)
	require.True(t, exists)

	require.NoError(t, srv.Close())
}

func TestRemoveBucket(t *testing.T) {
	srv, pc, c := serve(t, "")
	repo := tu.UniqueString("testremovebucket")

	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "master", "", nil))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	require.NoError(t, c.RemoveBucket(repo))
	require.NoError(t, c.RemoveBucket(fmt.Sprintf("%s-branch", repo)))

	require.NoError(t, srv.Close())
}

func TestRemoveBucketBranchless(t *testing.T) {
	srv, pc, c := serve(t, "")
	repo := tu.UniqueString("testremovebucketbranchless")

	// should error out because the repo doesn't have a branch
	require.NoError(t, pc.CreateRepo(repo))
	bucketNotFoundError(t, c.RemoveBucket(repo))
	bucketNotFoundError(t, c.RemoveBucket(fmt.Sprintf("%s-master", repo)))

	require.NoError(t, srv.Close())
}

func TestListObjectsPaginated(t *testing.T) {
	srv, pc, c := serve(t, "")

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
	ch := c.ListObjects(repo, "", false, make(chan struct{}))
	expectedFiles := []string{}
	for i := 0; i <= 1000; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("%d", i))
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{"dir/"})

	// Request that will list all files in master starting with 1
	ch = c.ListObjects(repo, "1", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i <= 1000; i++ {
		file := fmt.Sprintf("%d", i)
		if strings.HasPrefix(file, "1") {
			expectedFiles = append(expectedFiles, file)
		}
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	// Request that will list all files in a directory in master
	ch = c.ListObjects(repo, "dir/", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i < 10; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("dir/%d", i))
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	require.NoError(t, srv.Close())
}

func TestListObjectsHeadlessBranch(t *testing.T) {
	srv, pc, c := serve(t, "")

	repo := tu.UniqueString("testlistobjectsheadlessbranch")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "emptybranch", "", nil))

	// Request into branch that has no head
	ch := c.ListObjects(fmt.Sprintf("%s-emptybranch", repo), "", false, make(chan struct{}))
	checkListObjects(t, ch, time.Now(), time.Now(), []string{}, []string{})

	require.NoError(t, srv.Close())
}

func TestListObjectsRecursive(t *testing.T) {
	srv, pc, c := serve(t, "")

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

	// Request that will list all files and dirs in master
	expectedFiles := []string{"0", "rootdir/1", "rootdir/subdir/2"}
	expectedDirs := []string{"rootdir/", "rootdir/subdir/"}
	ch := c.ListObjects(repo, "", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Requests that will list all files in rootdir, including rootdir itself
	expectedFiles = []string{"rootdir/1", "rootdir/subdir/2"}
	ch = c.ListObjects(repo, "r", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)
	ch = c.ListObjects(repo, "rootdir", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Request that will list all files in rootdir, excluding rootdir itself
	expectedDirs = []string{"rootdir/subdir/"}
	ch = c.ListObjects(repo, "rootdir/", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Requests that will list all files in subdir, including subdir itself
	expectedFiles = []string{"rootdir/subdir/2"}
	ch = c.ListObjects(repo, "rootdir/s", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)
	ch = c.ListObjects(repo, "rootdir/subdir", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	// Request that will list all files in subdir, excluding subdir itself
	expectedDirs = []string{}
	ch = c.ListObjects(repo, "rootdir/subdir/", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)
	ch = c.ListObjects(repo, "rootdir/subdir/2", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, expectedDirs)

	require.NoError(t, srv.Close())
}
