package s3

import (
	"net"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"context"
	"time"

	minio "github.com/minio/minio-go"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func masterListBuckets(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
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

func masterListBucketsBranchless(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
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

func masterGetObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobject")
	require.NoError(t, pachClient.CreateRepo(repo))
	_, err := pachClient.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	fetchedContent, err := getObject(t, minioClient, repo, "master", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)
}

func masterGetObjectInBranch(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobjectinbranch")
	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))
	_, err := pachClient.PutFile(repo, "branch", "file", strings.NewReader("content"))
	require.NoError(t, err)

	fetchedContent, err := getObject(t, minioClient, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)
}

func masterStatObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
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

func masterPutObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
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

	fetchedContent, err := getObject(t, minioClient, repo, "branch", "file")
	require.NoError(t, err)
	require.Equal(t, "content2", fetchedContent)
}

func masterRemoveObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testremoveobject")
	require.NoError(t, pachClient.CreateRepo(repo))
	_, err := pachClient.PutFile(repo, "master", "file", strings.NewReader("content"))
	require.NoError(t, err)

	// as per PFS semantics, the second delete should be a no-op
	require.NoError(t, minioClient.RemoveObject(fmt.Sprintf("master.%s", repo), "file"))
	require.NoError(t, minioClient.RemoveObject(fmt.Sprintf("master.%s", repo), "file"))

	// make sure the object no longer exists
	_, err = getObject(t, minioClient, repo, "master", "file")
	keyNotFoundError(t, err)
}

// Tests inserting and getting files over 64mb in size
func masterLargeObjects(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
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

func masterGetObjectNoHead(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobjectnohead")
	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))

	_, err := getObject(t, minioClient, repo, "branch", "file")
	keyNotFoundError(t, err)
}

func masterGetObjectNoBranch(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobjectnobranch")
	require.NoError(t, pachClient.CreateRepo(repo))

	_, err := getObject(t, minioClient, repo, "branch", "file")
	bucketNotFoundError(t, err)
}

func masterGetObjectNoRepo(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobjectnorepo")
	_, err := getObject(t, minioClient, repo, "master", "file")
	bucketNotFoundError(t, err)
}

func masterMakeBucket(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucket")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), ""))

	repoInfo, err := pachClient.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "master")
}

func masterMakeBucketWithBranch(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucketwithbranch")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("branch.%s", repo), ""))

	repoInfo, err := pachClient.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "branch")
}

func masterMakeBucketWithRegion(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucketwithregion")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), "us-east-1"))
	_, err := pachClient.InspectRepo(repo)
	require.NoError(t, err)
}

func masterMakeBucketRedundant(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucketredundant")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), ""))
	err := minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), "")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "The bucket you tried to create already exists, and you own it.")
}

func masterMakeBucketDifferentBranches(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucketdifferentbranches")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), ""))
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("branch.%s", repo), ""))
}

func masterBucketExists(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
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

func masterRemoveBucket(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testremovebucket")

	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "master", "", nil))
	require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))

	require.NoError(t, minioClient.RemoveBucket(fmt.Sprintf("master.%s", repo)))
	require.NoError(t, minioClient.RemoveBucket(fmt.Sprintf("branch.%s", repo)))
}

func masterRemoveBucketBranchless(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testremovebucketbranchless")

	// should error out because the repo doesn't have a branch
	require.NoError(t, pachClient.CreateRepo(repo))
	bucketNotFoundError(t, minioClient.RemoveBucket(fmt.Sprintf("master.%s", repo)))
}

func masterListObjectsPaginated(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
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
		putListFileTestObject(t, pachClient, repo, commit.ID, "", i)
	}
	for i := 0; i < 10; i++ {
		putListFileTestObject(t, pachClient, repo, commit.ID, "dir/", i)
		require.NoError(t, err)
	}
	putListFileTestObject(t, pachClient, repo, "branch", "", 1001)
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

func masterListObjectsHeadlessBranch(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testlistobjectsheadlessbranch")
	require.NoError(t, pachClient.CreateRepo(repo))
	require.NoError(t, pachClient.CreateBranch(repo, "emptybranch", "", nil))

	// Request into branch that has no head
	ch := minioClient.ListObjects(fmt.Sprintf("emptybranch.%s", repo), "", false, make(chan struct{}))
	checkListObjects(t, ch, time.Now(), time.Now(), []string{}, []string{})
}

func masterListObjectsRecursive(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
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
	putListFileTestObject(t, pachClient, repo, commit.ID, "", 0)
	putListFileTestObject(t, pachClient, repo, commit.ID, "rootdir/", 1)
	putListFileTestObject(t, pachClient, repo, commit.ID, "rootdir/subdir/", 2)
	putListFileTestObject(t, pachClient, repo, "branch", "", 3)
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

func masterAuthV2(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
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

	pachClient, err := client.NewForTest()
	require.NoError(t, err)

	minioClient, err := minio.NewV4(fmt.Sprintf("127.0.0.1:%d", port), "", "", false)
	require.NoError(t, err)

	t.Run("master", func(t *testing.T) {
		t.Run("ListBuckets", func(t *testing.T) {
			masterListBuckets(t, pachClient, minioClient)
		})
		t.Run("ListBucketsBranchless", func(t *testing.T) {
			masterListBucketsBranchless(t, pachClient, minioClient)
		})
		t.Run("GetObject", func(t *testing.T) {
			masterGetObject(t, pachClient, minioClient)
		})
		t.Run("GetObjectInBranch", func(t *testing.T) {
			masterGetObjectInBranch(t, pachClient, minioClient)
		})
		t.Run("StatObject", func(t *testing.T) {
			masterStatObject(t, pachClient, minioClient)
		})
		t.Run("PutObject", func(t *testing.T) {
			masterPutObject(t, pachClient, minioClient)
		})
		t.Run("RemoveObject", func(t *testing.T) {
			masterRemoveObject(t, pachClient, minioClient)
		})
		t.Run("LargeObjects", func(t *testing.T) {
			masterLargeObjects(t, pachClient, minioClient)
		})
		t.Run("GetObjectNoHead", func(t *testing.T) {
			masterGetObjectNoHead(t, pachClient, minioClient)
		})
		t.Run("GetObjectNoBranch", func(t *testing.T) {
			masterGetObjectNoBranch(t, pachClient, minioClient)
		})
		t.Run("GetObjectNoRepo", func(t *testing.T) {
			masterGetObjectNoRepo(t, pachClient, minioClient)
		})
		t.Run("MakeBucket", func(t *testing.T) {
			masterMakeBucket(t, pachClient, minioClient)
		})
		t.Run("MakeBucketWithBranch", func(t *testing.T) {
			masterMakeBucketWithBranch(t, pachClient, minioClient)
		})
		t.Run("MakeBucketWithRegion", func(t *testing.T) {
			masterMakeBucketWithRegion(t, pachClient, minioClient)
		})
		t.Run("MakeBucketRedundant", func(t *testing.T) {
			masterMakeBucketRedundant(t, pachClient, minioClient)
		})
		t.Run("MakeBucketDifferentBranches", func(t *testing.T) {
			masterMakeBucketDifferentBranches(t, pachClient, minioClient)
		})
		t.Run("BucketExists", func(t *testing.T) {
			masterBucketExists(t, pachClient, minioClient)
		})
		t.Run("RemoveBucket", func(t *testing.T) {
			masterRemoveBucket(t, pachClient, minioClient)
		})
		t.Run("RemoveBucketBranchless", func(t *testing.T) {
			masterRemoveBucketBranchless(t, pachClient, minioClient)
		})
		t.Run("ListObjectsPaginated", func(t *testing.T) {
			masterListObjectsPaginated(t, pachClient, minioClient)
		})
		t.Run("ListObjectsHeadlessBranch", func(t *testing.T) {
			masterListObjectsHeadlessBranch(t, pachClient, minioClient)
		})
		t.Run("ListObjectsRecursive", func(t *testing.T) {
			masterListObjectsRecursive(t, pachClient, minioClient)
		})
		t.Run("AuthV2", func(t *testing.T) {
			masterAuthV2(t, pachClient, minioClient)
		})
	})
}
