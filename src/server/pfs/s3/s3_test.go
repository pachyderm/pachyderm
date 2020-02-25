package s3

// Tests for the PFS' S3 emulation API. Note that, in calls to
// `tu.UniqueString`, all lowercase characters are used, unlike in other
// tests. This is in order to generate repo names that are also valid bucket
// names. Otherwise minio complains that the bucket name is not valid.

import (
	"crypto/md5"
	"net"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"context"
	"time"

	minio "github.com/minio/minio-go"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

var (
	pachClient  *client.APIClient
	minioClient *minio.Client
)

type TestClientFactory struct {}

func (f *TestClientFactory) Client(authToken string) (*client.APIClient, error) {
	c, err := client.NewForTest()
	if err != nil {
		return nil, err
	}
	if authToken != "" {
		c.SetAuthToken(authToken)
	}
	return c, nil
}

func getObject(t *testing.T, repo, branch, file string) (string, error) {
	t.Helper()

	obj, err := minioClient.GetObject(fmt.Sprintf("%s.%s", branch, repo), file, minio.GetObjectOptions{})
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

func putListFileTestObject(t *testing.T, repo string, commitID string, dir string, i int) {
	t.Helper()
	_, err := pachClient.PutFile(
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

func testListBuckets(t *testing.T) {
	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	repo := tu.UniqueString("testlistbuckets1")
	require.NoError(t, pachClient.CreateRepo(repo))
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	require.NoError(t, pachClient.CreateBranch(repo, "master", "", nil))
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))

	hasMaster := false
	hasBranch := false

	buckets, err := minioClient.ListBuckets()
	require.NoError(t, err)

	for _, bucket := range buckets {
		if bucket.Name == fmt.Sprintf("master.%s", repo) {
			hasMaster = true
			require.True(t, startTime.Before(bucket.CreationDate))
			require.True(t, endTime.After(bucket.CreationDate))
		} else if bucket.Name == fmt.Sprintf("branch.%s", repo) {
			hasBranch = true
			require.True(t, startTime.Before(bucket.CreationDate))
			require.True(t, endTime.After(bucket.CreationDate))
		}
	}

	require.True(t, hasMaster)
	require.True(t, hasBranch)
}

func testListBucketsBranchless(t *testing.T) {
	repo1 := tu.UniqueString("testlistbucketsbranchless1")
	require.NoError(t, pachClient.CreateRepo(repo1))
	repo2 := tu.UniqueString("testlistbucketsbranchless2")
	require.NoError(t, pachClient.CreateRepo(repo2))

	// should be 0 since no branches have been made yet
	buckets, err := minioClient.ListBuckets()
	require.NoError(t, err)
	for _, bucket := range buckets {
		require.NotEqual(t, bucket.Name, repo1)
		require.NotEqual(t, bucket.Name, repo2)
	}
}

func testGetObject(t *testing.T) {
	repo := tu.UniqueString("testgetobject")
	require.NoError(t, pachClient.CreateRepo(repo))
	_, err := pachClient.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	fetchedContent, err := getObject(t, repo, "master", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)
}

func testGetObjectInBranch(t *testing.T) {
	repo := tu.UniqueString("testgetobjectinbranch")
	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))
	_, err := pachClient.PutFile(repo, "branch", "file", strings.NewReader("content"))
	require.NoError(t, err)

	fetchedContent, err := getObject(t, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)
}

func testStatObject(t *testing.T) {
	repo := tu.UniqueString("teststatobject")
	require.NoError(t, pachClient.CreateRepo(repo))
	_, err := pachClient.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	_, err = pachClient.PutFileOverwrite(repo, "master", "file", strings.NewReader("new-content"), 0)
	require.NoError(t, err)
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	info, err := minioClient.StatObject(fmt.Sprintf("master.%s", repo), "file", minio.StatObjectOptions{})
	require.NoError(t, err)
	require.True(t, startTime.Before(info.LastModified))
	require.True(t, endTime.After(info.LastModified))
	require.True(t, len(info.ETag) > 0)
	require.Equal(t, "text/plain; charset=utf-8", info.ContentType)
	require.Equal(t, int64(11), info.Size)
}

func testPutObject(t *testing.T) {
	repo := tu.UniqueString("testputobject")
	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))

	r := strings.NewReader("content1")
	_, err := minioClient.PutObject(fmt.Sprintf("branch.%s", repo), "file", r, int64(r.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
	require.NoError(t, err)

	// this should act as a PFS PutFileOverwrite
	r2 := strings.NewReader("content2")
	_, err = minioClient.PutObject(fmt.Sprintf("branch.%s", repo), "file", r2, int64(r2.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
	require.NoError(t, err)

	fetchedContent, err := getObject(t, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content2", fetchedContent)
}

func testRemoveObject(t *testing.T) {
	repo := tu.UniqueString("testremoveobject")
	require.NoError(t, pachClient.CreateRepo(repo))
	_, err := pachClient.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	// as per PFS semantics, the second delete should be a no-op
	require.NoError(t, minioClient.RemoveObject(fmt.Sprintf("master.%s", repo), "file"))
	require.NoError(t, minioClient.RemoveObject(fmt.Sprintf("master.%s", repo), "file"))

	// make sure the object no longer exists
	_, err = getObject(t, repo, "master", "file")
	keyNotFoundError(t, err)
}

// Tests inserting and getting files over 64mb in size
func testLargeObjects(t *testing.T) {
	// test repos: repo1 exists, repo2 does not
	repo1 := tu.UniqueString("testlargeobject1")
	repo2 := tu.UniqueString("testlargeobject2")
	require.NoError(t, pachClient.CreateRepo(repo1))
	require.NoError(t, pachClient.CreateBranch(repo1, "master", "", nil))

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
	_, err = minioClient.FPutObject(fmt.Sprintf("master.%s", repo2), "file", inputFile.Name(), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	bucketNotFoundError(t, err)

	// now try putting into a legit repo
	l, err := minioClient.FPutObject(fmt.Sprintf("master.%s", repo1), "file", inputFile.Name(), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	require.NoError(t, err)
	require.Equal(t, int(l), 68157450)

	// try getting an object that does not exist
	err = minioClient.FGetObject(fmt.Sprintf("master.%s", repo2), "file", "foo", minio.GetObjectOptions{})
	bucketNotFoundError(t, err)

	// get the file that does exist
	outputFile, err := ioutil.TempFile("", "pachyderm-test-large-objects-output-*")
	require.NoError(t, err)
	defer os.Remove(outputFile.Name())
	err = minioClient.FGetObject(fmt.Sprintf("master.%s", repo1), "file", outputFile.Name(), minio.GetObjectOptions{})
	require.True(t, err == nil || err == io.EOF, fmt.Sprintf("unexpected error: %s", err))

	// compare the files and ensure they're the same
	// NOTE: Because minio's `FGetObject` does a rename from a buffer file
	// to the given filepath, `outputFile` will refer to an empty, overwritten
	// file. We can still use `outputFile.Name()` though.
	inputFileSize, inputFileHash := fileHash(t, inputFile.Name())
	outputFileSize, outputFileHash := fileHash(t, inputFile.Name())
	require.Equal(t, inputFileSize, outputFileSize)
	require.Equal(t, inputFileHash, outputFileHash)
}

func testGetObjectNoHead(t *testing.T) {
	repo := tu.UniqueString("testgetobjectnohead")
	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))

	_, err := getObject(t, repo, "branch", "file")
	keyNotFoundError(t, err)
}

func testGetObjectNoBranch(t *testing.T) {
	repo := tu.UniqueString("testgetobjectnobranch")
	require.NoError(t, pachClient.CreateRepo(repo))

	_, err := getObject(t, repo, "branch", "file")
	bucketNotFoundError(t, err)
}

func testGetObjectNoRepo(t *testing.T) {
	repo := tu.UniqueString("testgetobjectnorepo")
	_, err := getObject(t, repo, "master", "file")
	bucketNotFoundError(t, err)
}

func testMakeBucket(t *testing.T) {
	repo := tu.UniqueString("testmakebucket")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), ""))

	repoInfo, err := pachClient.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "master")
}

func testMakeBucketWithBranch(t *testing.T) {
	repo := tu.UniqueString("testmakebucketwithbranch")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("branch.%s", repo), ""))

	repoInfo, err := pachClient.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "branch")
}

func testMakeBucketWithRegion(t *testing.T) {
	repo := tu.UniqueString("testmakebucketwithregion")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), "us-east-1"))
	_, err := pachClient.InspectRepo(repo)
	require.NoError(t, err)
}

func testMakeBucketRedundant(t *testing.T) {
	repo := tu.UniqueString("testmakebucketredundant")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), ""))
	err := minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), "")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "The bucket you tried to create already exists, and you own it.")
}

func testMakeBucketDifferentBranches(t *testing.T) {
	repo := tu.UniqueString("testmakebucketdifferentbranches")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), ""))
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("branch.%s", repo), ""))
}

func testBucketExists(t *testing.T) {
	repo := tu.UniqueString("testbucketexists")

	exists, err := minioClient.BucketExists(fmt.Sprintf("master.%s", repo))
	require.NoError(t, err)
	require.False(t, exists)

	// repo exists, but branch doesn't: should be false
	require.NoError(t, pachClient.CreateRepo(repo))
	exists, err = minioClient.BucketExists(fmt.Sprintf("master.%s", repo))
	require.NoError(t, err)
	require.False(t, exists)

	// repo and branch exists: should be true
	require.NoError(t, pachClient.CreateBranch(repo, "master", "", nil))
	exists, err = minioClient.BucketExists(fmt.Sprintf("master.%s", repo))
	require.NoError(t, err)
	require.True(t, exists)

	// repo exists, but branch doesn't: should be false
	exists, err = minioClient.BucketExists(fmt.Sprintf("branch.%s", repo))
	require.NoError(t, err)
	require.False(t, exists)

	// repo and branch exists: should be true
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))
	exists, err = minioClient.BucketExists(fmt.Sprintf("branch.%s", repo))
	require.NoError(t, err)
	require.True(t, exists)
}

func testRemoveBucket(t *testing.T) {
	repo := tu.UniqueString("testremovebucket")

	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "master", "", nil))
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))

	require.NoError(t, minioClient.RemoveBucket(fmt.Sprintf("master.%s", repo)))
	require.NoError(t, minioClient.RemoveBucket(fmt.Sprintf("branch.%s", repo)))
}

func testRemoveBucketBranchless(t *testing.T) {
	repo := tu.UniqueString("testremovebucketbranchless")

	// should error out because the repo doesn't have a branch
	require.NoError(t, pachClient.CreateRepo(repo))
	bucketNotFoundError(t, minioClient.RemoveBucket(fmt.Sprintf("master.%s", repo)))
}

func testListObjectsPaginated(t *testing.T) {
	// create a bunch of files - enough to require the use of paginated
	// requests when browsing all files. One file will be included on a
	// separate branch to ensure it's not returned when querying against the
	// master branch.
	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	repo := tu.UniqueString("testlistobjectspaginated")
	require.NoError(t, pachClient.CreateRepo(repo))
	commit, err := pachClient.StartCommit(repo, "master")
	require.NoError(t, err)
	for i := 0; i <= 1000; i++ {
		putListFileTestObject(t, repo, commit.ID, "", i)
	}
	for i := 0; i < 10; i++ {
		putListFileTestObject(t, repo, commit.ID, "dir/", i)
		require.NoError(t, err)
	}
	putListFileTestObject(t, repo, "branch", "", 1001)
	require.NoError(t, pachClient.FinishCommit(repo, commit.ID))
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	// Request that will list all files in master's root
	ch := minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "", false, make(chan struct{}))
	expectedFiles := []string{}
	for i := 0; i <= 1000; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("%d", i))
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{"dir/"})

	// Request that will list all files in master starting with 1
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "1", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i <= 1000; i++ {
		file := fmt.Sprintf("%d", i)
		if strings.HasPrefix(file, "1") {
			expectedFiles = append(expectedFiles, file)
		}
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	// Request that will list all files in a directory in master
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "dir/", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i < 10; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("dir/%d", i))
	}
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
}

func testListObjectsHeadlessBranch(t *testing.T) {
	repo := tu.UniqueString("testlistobjectsheadlessbranch")
	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "emptybranch", "", nil))

	// Request into branch that has no head
	ch := minioClient.ListObjects(fmt.Sprintf("emptybranch.%s", repo), "", false, make(chan struct{}))
	checkListObjects(t, ch, time.Now(), time.Now(), []string{}, []string{})
}

func testListObjectsRecursive(t *testing.T) {
	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	repo := tu.UniqueString("testlistobjectsrecursive")
	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))
	require.NoError(t, pachClient.CreateBranch(repo, "emptybranch", "", nil))
	commit, err := pachClient.StartCommit(repo, "master")
	require.NoError(t, err)
	putListFileTestObject(t, repo, commit.ID, "", 0)
	putListFileTestObject(t, repo, commit.ID, "rootdir/", 1)
	putListFileTestObject(t, repo, commit.ID, "rootdir/subdir/", 2)
	putListFileTestObject(t, repo, "branch", "", 3)
	require.NoError(t, pachClient.FinishCommit(repo, commit.ID))
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	// Request that will list all files in master
	expectedFiles := []string{"0", "rootdir/1", "rootdir/subdir/2"}
	ch := minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	// Requests that will list all files in rootdir
	expectedFiles = []string{"rootdir/1", "rootdir/subdir/2"}
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "r", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

	// Requests that will list all files in subdir
	expectedFiles = []string{"rootdir/subdir/2"}
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/s", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/subdir", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/subdir/", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/subdir/2", true, make(chan struct{}))
	checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
}

func testAuthV2(t *testing.T) {
	// The other tests use auth V4, versus this which checks auth V2
	minioClientV2, err := minio.NewV2("127.0.0.1:30600", "", "", false)
	require.NoError(t, err)
	_, err = minioClientV2.ListBuckets()
	require.NoError(t, err)
}

func TestS3Gateway(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	server, err := Server(0, NewMasterDriver(), &TestClientFactory{})
	require.NoError(t, err)
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		server.Serve(listener)
	}()
	defer func() {
		require.NoError(t, server.Shutdown(context.Background()))
	}()

	port := listener.Addr().(*net.TCPAddr).Port

	pachClient, err = client.NewForTest()
	require.NoError(t, err)

	minioClient, err = minio.NewV4(fmt.Sprintf("127.0.0.1:%d", port), "", "", false)
	require.NoError(t, err)

	t.Run("s3gateway", func(t *testing.T) {
		t.Run("testListBuckets", testListBuckets)
		t.Run("testListBucketsBranchless", testListBucketsBranchless)
		t.Run("testGetObject", testGetObject)
		t.Run("testGetObjectInBranch", testGetObjectInBranch)
		t.Run("testStatObject", testStatObject)
		t.Run("testPutObject", testPutObject)
		t.Run("testRemoveObject", testRemoveObject)
		t.Run("testLargeObjects", testLargeObjects)
		t.Run("testGetObjectNoHead", testGetObjectNoHead)
		t.Run("testGetObjectNoBranch", testGetObjectNoBranch)
		t.Run("testGetObjectNoRepo", testGetObjectNoRepo)
		t.Run("testMakeBucket", testMakeBucket)
		t.Run("testMakeBucketWithBranch", testMakeBucketWithBranch)
		t.Run("testMakeBucketWithRegion", testMakeBucketWithRegion)
		t.Run("testMakeBucketRedundant", testMakeBucketRedundant)
		t.Run("testMakeBucketDifferentBranches", testMakeBucketDifferentBranches)
		t.Run("testBucketExists", testBucketExists)
		t.Run("testRemoveBucket", testRemoveBucket)
		t.Run("testRemoveBucketBranchless", testRemoveBucketBranchless)
		t.Run("testListObjectsPaginated", testListObjectsPaginated)
		t.Run("testListObjectsHeadlessBranch", testListObjectsHeadlessBranch)
		t.Run("testListObjectsRecursive", testListObjectsRecursive)
		t.Run("testAuthV2", testAuthV2)
	})
}
