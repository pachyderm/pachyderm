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
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	minio "github.com/minio/minio-go"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

var (
	port int32 = 30800 // Initial port on which s3 gateway will serve

)

func clients(t testing.TB, view map[string]*pfs.Commit) (*client.APIClient, *minio.Client, func()) {
	t.Helper()

	pachClient := server.GetPachClient(t)
	s3Port := atomic.AddInt32(&port, 1)
	s := Server(pachClient, &Config{Port: uint16(s3Port), NoEnterpriseCheck: true, View: view})
	go s.ListenAndServe()

	minioClient, err := minio.New(fmt.Sprintf("127.0.0.1:%d", s3Port), "id", "secret", false)
	require.NoError(t, err)
	return pachClient, minioClient, func() { require.NoError(t, s.Close()) }
}

func bucket(repo, branch string) string {
	return fmt.Sprintf("%s.%s", branch, repo)
}

func getObject(t *testing.T, c *minio.Client, bucket, file string) (_ string, retErr error) {
	t.Helper()

	obj, err := c.GetObject(bucket, file)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := obj.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
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
		require.True(t, len(actualFile.ETag) > 0, fmt.Sprintf("unexpected empty etag for %s", expectedFilename))
		expectedLen := int64(len(filepath.Base(expectedFilename)) + 1)
		require.Equal(t, expectedLen, actualFile.Size, fmt.Sprintf("unexpected file length for %s", expectedFilename))
		require.True(t, startTime.Before(actualFile.LastModified), fmt.Sprintf("unexpected last modified for %s", expectedFilename))
		require.True(t, endTime.After(actualFile.LastModified), fmt.Sprintf("unexpected last modified for %s", expectedFilename))
	}

	for i, expectedDirname := range expectedDirs {
		actualDir := actualDirs[i]
		require.Equal(t, expectedDirname, actualDir.Key)
		require.True(t, len(actualDir.ETag) == 0, fmt.Sprintf("unexpected etag for %s: %s", expectedDirname, actualDir.ETag))
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
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	repo := tu.UniqueString("testlistbuckets1")
	require.NoError(t, pc.CreateRepo(repo))
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	require.NoError(t, pc.CreateBranch(repo, "master", "", nil))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	hasMaster := false
	hasBranch := false

	buckets, err := c.ListBuckets()
	require.NoError(t, err)

	for _, b := range buckets {
		if b.Name == bucket(repo, "master") {
			hasMaster = true
			require.True(t, startTime.Before(b.CreationDate))
			require.True(t, endTime.After(b.CreationDate))
		} else if b.Name == bucket(repo, "branch") {
			hasBranch = true
			require.True(t, startTime.Before(b.CreationDate))
			require.True(t, endTime.After(b.CreationDate))
		}
	}

	require.True(t, hasMaster)
	require.True(t, hasBranch)
}

func TestListBucketsBranchless(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

	repo1 := tu.UniqueString("testlistbucketsbranchless1")
	require.NoError(t, pc.CreateRepo(repo1))
	repo2 := tu.UniqueString("testlistbucketsbranchless2")
	require.NoError(t, pc.CreateRepo(repo2))

	// should be 0 since no branches have been made yet
	buckets, err := c.ListBuckets()
	require.NoError(t, err)
	for _, bucket := range buckets {
		require.NotEqual(t, bucket.Name, repo1)
		require.NotEqual(t, bucket.Name, repo2)
	}
}

func TestGetObject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

	repo := tu.UniqueString("testgetobject")
	require.NoError(t, pc.CreateRepo(repo))
	_, err := pc.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	fetchedContent, err := getObject(t, c, bucket(repo, "master"), "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)
}

func TestGetObjectInBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

	repo := tu.UniqueString("testgetobjectinbranch")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))
	_, err := pc.PutFile(repo, "branch", "file", strings.NewReader("content"))
	require.NoError(t, err)

	fetchedContent, err := getObject(t, c, bucket(repo, "branch"), "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)
}

func TestStatObject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

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

	info, err := c.StatObject(bucket(repo, "master"), "file")
	require.NoError(t, err)
	require.True(t, startTime.Before(info.LastModified))
	require.True(t, endTime.After(info.LastModified))
	require.True(t, len(info.ETag) > 0)
	require.Equal(t, "text/plain; charset=utf-8", info.ContentType)
	require.Equal(t, int64(11), info.Size)
}

func TestPutObject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

	repo := tu.UniqueString("testputobject")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	_, err := c.PutObject(bucket(repo, "branch"), "file", strings.NewReader("content1"), "text/plain")
	require.NoError(t, err)

	// this should act as a PFS PutFileOverwrite
	_, err = c.PutObject(bucket(repo, "branch"), "file", strings.NewReader("content2"), "text/plain")
	require.NoError(t, err)

	fetchedContent, err := getObject(t, c, bucket(repo, "branch"), "file")
	require.NoError(t, err)
	require.Equal(t, "content2", fetchedContent)
}

func TestRemoveObject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

	repo := tu.UniqueString("testremoveobject")
	require.NoError(t, pc.CreateRepo(repo))
	_, err := pc.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	// as per PFS semantics, the second delete should be a no-op
	require.NoError(t, c.RemoveObject(bucket(repo, "master"), "file"))
	require.NoError(t, c.RemoveObject(bucket(repo, "master"), "file"))

	// make sure the object no longer exists
	_, err = getObject(t, c, bucket(repo, "master"), "file")
	keyNotFoundError(t, err)
}

// Tests inserting and getting files over 64mb in size
func TestLargeObjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

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
	_, err = c.FPutObject(bucket(repo2, "master"), "file", inputFile.Name(), "text/plain")
	bucketNotFoundError(t, err)

	// now try putting into a legit repo
	// NOTE: `FPutObject` will try to use multipart uploads since the file
	// size is over 65mb. Because the s3gateway returns that
	// multipart-related endpoints are not implemented, minio gracefully
	// degrades to a standard put object operation. When a standard put
	// operation is performed, `FPutObject` will return no error assuming
	// everything went OK. If multipart were ever implemented, `FPutObject`
	// will default to using that instead, and will return `io.EOF` if
	// everything went OK instead.
	l, err := c.FPutObject(bucket(repo1, "master"), "file", inputFile.Name(), "text/plain")
	require.NoError(t, err)
	require.Equal(t, int(l), 68157450)

	// try getting an object that does not exist
	err = c.FGetObject(bucket(repo2, "master"), "file", "foo")
	bucketNotFoundError(t, err)

	// get the file that does exist
	outputFile, err := ioutil.TempFile("", "pachyderm-test-large-objects-output-*")
	require.NoError(t, err)
	defer os.Remove(outputFile.Name())
	err = c.FGetObject(bucket(repo1, "master"), "file", outputFile.Name())
	require.NoError(t, err)

	// compare the files and ensure they're the same
	// NOTE: Because minio's `FGetObject` does a rename from a buffer file
	// to the given filepath, `outputFile` will refer to an empty, overwritten
	// file. We can still use `outputFile.Name()` though.
	inputFileSize, inputFileHash := fileHash(t, inputFile.Name())
	outputFileSize, outputFileHash := fileHash(t, inputFile.Name())
	require.Equal(t, inputFileSize, outputFileSize)
	require.Equal(t, inputFileHash, outputFileHash)
}

func TestGetObjectNoHead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

	repo := tu.UniqueString("testgetobjectnohead")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	_, err := getObject(t, c, bucket(repo, "branch"), "file")
	keyNotFoundError(t, err)
}

func TestGetObjectNoBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

	repo := tu.UniqueString("testgetobjectnobranch")
	require.NoError(t, pc.CreateRepo(repo))

	_, err := getObject(t, c, bucket(repo, "branch"), "file")
	bucketNotFoundError(t, err)
}

func TestGetObjectNoRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	_, c, close := clients(t, nil)
	defer close()
	repo := tu.UniqueString("testgetobjectnorepo")
	_, err := getObject(t, c, bucket(repo, "master"), "file")
	bucketNotFoundError(t, err)
}

func TestMakeBucket(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()
	repo := tu.UniqueString("testmakebucket")
	require.NoError(t, c.MakeBucket(bucket(repo, "master"), ""))

	repoInfo, err := pc.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "master")
}

func TestMakeBucketWithBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()
	repo := tu.UniqueString("testmakebucketwithbranch")
	require.NoError(t, c.MakeBucket(bucket(repo, "branch"), ""))

	repoInfo, err := pc.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "branch")
}

func TestMakeBucketWithRegion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()
	repo := tu.UniqueString("testmakebucketwithregion")
	require.NoError(t, c.MakeBucket(bucket(repo, "master"), "us-east-1"))
	_, err := pc.InspectRepo(repo)
	require.NoError(t, err)
}

func TestMakeBucketRedundant(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	_, c, close := clients(t, nil)
	defer close()
	repo := tu.UniqueString("testmakebucketredundant")
	require.NoError(t, c.MakeBucket(bucket(repo, "master"), ""))
	err := c.MakeBucket(bucket(repo, "master"), "")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "The bucket you tried to create already exists, and you own it.")
}

func TestMakeBucketDifferentBranches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	_, c, close := clients(t, nil)
	defer close()
	repo := tu.UniqueString("testmakebucketdifferentbranches")
	require.NoError(t, c.MakeBucket(bucket(repo, "master"), ""))
	require.NoError(t, c.MakeBucket(bucket(repo, "branch"), ""))
}

func TestBucketExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()
	repo := tu.UniqueString("testbucketexists")

	exists, err := c.BucketExists(bucket(repo, "master"))
	require.NoError(t, err)
	require.False(t, exists)

	// repo exists, but branch doesn't: should be false
	require.NoError(t, pc.CreateRepo(repo))
	exists, err = c.BucketExists(bucket(repo, "master"))
	require.NoError(t, err)
	require.False(t, exists)

	// repo and branch exists: should be true
	require.NoError(t, pc.CreateBranch(repo, "master", "", nil))
	exists, err = c.BucketExists(bucket(repo, "master"))
	require.NoError(t, err)
	require.True(t, exists)

	// repo exists, but branch doesn't: should be false
	exists, err = c.BucketExists(bucket(repo, "branch"))
	require.NoError(t, err)
	require.False(t, exists)

	// repo and branch exists: should be true
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))
	exists, err = c.BucketExists(bucket(repo, "branch"))
	require.NoError(t, err)
	require.True(t, exists)
}

func TestRemoveBucket(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()
	repo := tu.UniqueString("testremovebucket")

	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "master", "", nil))
	require.NoError(t, pc.CreateBranch(repo, "branch", "", nil))

	require.NoError(t, c.RemoveBucket(bucket(repo, "master")))
	require.NoError(t, c.RemoveBucket(bucket(repo, "branch")))
}

func TestRemoveBucketBranchless(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()
	repo := tu.UniqueString("testremovebucketbranchless")

	// should error out because the repo doesn't have a branch
	require.NoError(t, pc.CreateRepo(repo))
	bucketNotFoundError(t, c.RemoveBucket(bucket(repo, "master")))
}

func TestListObjectsPaginated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

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
	ch := c.ListObjects(bucket(repo, "master"), "", false, make(chan struct{}))
	expectedFiles := []string{}
	for i := 0; i <= 1000; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("%d", i))
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{"dir/"})

	// Request that will list all files in master starting with 1
	ch = c.ListObjects(bucket(repo, "master"), "1", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i <= 1000; i++ {
		file := fmt.Sprintf("%d", i)
		if strings.HasPrefix(file, "1") {
			expectedFiles = append(expectedFiles, file)
		}
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	// Request that will list all files in a directory in master
	ch = c.ListObjects(bucket(repo, "master"), "dir/", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i < 10; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("dir/%d", i))
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
}

func TestListObjectsHeadlessBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

	repo := tu.UniqueString("testlistobjectsheadlessbranch")
	require.NoError(t, pc.CreateRepo(repo))
	require.NoError(t, pc.CreateBranch(repo, "emptybranch", "", nil))

	// Request into branch that has no head
	ch := c.ListObjects(bucket(repo, "emptybranch"), "", false, make(chan struct{}))
	checkListObjects(t, ch, time.Now(), time.Now(), []string{}, []string{})
}

func TestListObjectsRecursive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	pc, c, close := clients(t, nil)
	defer close()

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

	// Request that will list all files in master
	expectedFiles := []string{"0", "rootdir/1", "rootdir/subdir/2"}
	ch := c.ListObjects(bucket(repo, "master"), "", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	// Requests that will list all files in rootdir
	expectedFiles = []string{"rootdir/1", "rootdir/subdir/2"}
	ch = c.ListObjects(bucket(repo, "master"), "r", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = c.ListObjects(bucket(repo, "master"), "rootdir", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = c.ListObjects(bucket(repo, "master"), "rootdir/", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	// Requests that will list all files in subdir
	expectedFiles = []string{"rootdir/subdir/2"}
	ch = c.ListObjects(bucket(repo, "master"), "rootdir/s", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = c.ListObjects(bucket(repo, "master"), "rootdir/subdir", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = c.ListObjects(bucket(repo, "master"), "rootdir/subdir/", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = c.ListObjects(bucket(repo, "master"), "rootdir/subdir/2", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
}

func TestView(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	view := make(map[string]*pfs.Commit)
	pc, c, close := clients(t, view)
	defer close()
	repo := "view"
	require.NoError(t, pc.CreateRepo(repo))
	_, err := pc.PutFile(repo, "master", "file1", strings.NewReader("content1"))
	require.NoError(t, err)
	_, err = pc.PutFile(repo, "master", "file2", strings.NewReader("content2"))
	require.NoError(t, err)
	_, err = pc.PutFile(repo, "master", "file3", strings.NewReader("content3"))
	require.NoError(t, err)

	otherRepo := "other"
	require.NoError(t, pc.CreateRepo(otherRepo))
	_, err = pc.PutFile(otherRepo, "master", "file1", strings.NewReader("content4"))
	require.NoError(t, err)

	// Because the view is empty (not nil) we don't see any buckets, this case
	// is degenerate, but seems like the only reasonable behavior.
	buckets, err := c.ListBuckets()
	require.NoError(t, err)
	require.Equal(t, 0, len(buckets))

	view["bucket"] = client.NewCommit(repo, "master.1")

	buckets, err = c.ListBuckets()
	require.NoError(t, err)
	require.Equal(t, 1, len(buckets))
	require.Equal(t, "bucket", buckets[0].Name)

	ch := c.ListObjects("bucket", "", true, nil)
	var objects []string
	for object := range ch {
		objects = append(objects, object.Key)
	}
	require.Equal(t, 1, len(objects))

	content, err := getObject(t, c, "bucket", "file1")
	require.NoError(t, err)
	require.Equal(t, "content1", content)

	_, err = getObject(t, c, bucket(repo, "master"), "file1")
	require.YesError(t, err)
}
