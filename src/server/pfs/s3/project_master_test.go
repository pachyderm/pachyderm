//go:build unit_test

package s3_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	minio "github.com/minio/minio-go/v6"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/s3"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func projectMasterListBuckets(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	repo := tu.UniqueString("testlistbuckets1")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", nil))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "branch", "", "", nil))

	hasMaster := false
	hasBranch := false

	buckets, err := minioClient.ListBuckets()
	require.NoError(t, err)

	for _, bucket := range buckets {
		if bucket.Name == fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName) {
			hasMaster = true
			require.True(t, startTime.Before(bucket.CreationDate))
			require.True(t, endTime.After(bucket.CreationDate))
		} else if bucket.Name == fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName) {
			hasBranch = true
			require.True(t, startTime.Before(bucket.CreationDate))
			require.True(t, endTime.After(bucket.CreationDate))
		}
	}

	require.True(t, hasMaster)
	require.True(t, hasBranch)
}

func projectMasterListBucketsBranchless(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo1 := tu.UniqueString("testlistbucketsbranchless1")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo1))
	repo2 := tu.UniqueString("testlistbucketsbranchless2")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo2))

	buckets, err := minioClient.ListBuckets()
	require.NoError(t, err)
	// there should be no buckets for repos without branches
	for _, bucket := range buckets {
		require.NotEqual(t, bucket.Name, fmt.Sprintf("%s.%s", repo1, pfs.DefaultProjectName))
		require.NotEqual(t, bucket.Name, fmt.Sprintf("%s.%s", repo2, pfs.DefaultProjectName))
	}
}

func projectMasterGetObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobject")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, pachClient.PutFile(commit, "file", strings.NewReader("content")))

	fetchedContent, err := getObject(t, minioClient, fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)
}

func projectMasterGetObjectInBranch(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobjectinbranch")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "branch", "", "", nil))
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "branch", "")
	require.NoError(t, pachClient.PutFile(commit, "file", strings.NewReader("content")))

	fetchedContent, err := getObject(t, minioClient, fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName), "file")
	require.NoError(t, err)
	require.Equal(t, "content", fetchedContent)
}

func projectMasterStatObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("teststatobject")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, pachClient.PutFile(commit, "file", strings.NewReader("content")))

	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	require.NoError(t, pachClient.PutFile(commit, "file", strings.NewReader("new-content")))
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	info, err := minioClient.StatObject(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "file", minio.StatObjectOptions{})
	require.NoError(t, err)
	require.True(t, startTime.Before(info.LastModified))
	require.True(t, endTime.After(info.LastModified))
	require.True(t, len(info.ETag) > 0)
	require.Equal(t, "text/plain; charset=utf-8", info.ContentType)
	require.Equal(t, int64(11), info.Size)
}

func projectMasterPutObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testputobject")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "branch", "", "", nil))

	r := strings.NewReader("content1")
	_, err := minioClient.PutObject(fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName), "file", r, int64(r.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
	require.NoError(t, err)

	// this should act as a PFS PutFile
	r2 := strings.NewReader("content2")
	_, err = minioClient.PutObject(fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName), "file", r2, int64(r2.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
	require.NoError(t, err)

	fetchedContent, err := getObject(t, minioClient, fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName), "file")
	require.NoError(t, err)
	require.Equal(t, "content2", fetchedContent)
}

func projectMasterRemoveObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testremoveobject")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
	require.NoError(t, pachClient.PutFile(commit, "file", strings.NewReader("content")))

	// as per PFS semantics, the second delete should be a no-op
	require.NoError(t, minioClient.RemoveObject(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "file"))
	require.NoError(t, minioClient.RemoveObject(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "file"))

	// make sure the object no longer exists
	_, err := getObject(t, minioClient, fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "file")
	keyNotFoundError(t, err)
}

// Tests inserting and getting files over 64mb in size
func projectMasterLargeObjects(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	// test repos: repo1 exists, repo2 does not
	repo1 := tu.UniqueString("testlargeobject1")
	repo2 := tu.UniqueString("testlargeobject2")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo1))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo1, "master", "", "", nil))

	// create a temporary file to put ~65mb of contents into it
	inputFile, err := os.CreateTemp("", "pachyderm-test-large-objects-input-*")
	require.NoError(t, err)
	defer os.Remove(inputFile.Name())
	n, err := inputFile.WriteString(strings.Repeat("no tv and no beer make homer something something.\n", 1363149))
	require.NoError(t, err)
	require.Equal(t, n, 68157450)
	require.NoError(t, inputFile.Sync())

	// first ensure that putting into a repo that doesn't exist triggers an
	// error
	_, err = minioClient.FPutObject(fmt.Sprintf("master.%s.%s", repo2, pfs.DefaultProjectName), "file", inputFile.Name(), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	bucketNotFoundError(t, err)

	// now try putting into a legit repo
	l, err := minioClient.FPutObject(fmt.Sprintf("master.%s.%s", repo1, pfs.DefaultProjectName), "file", inputFile.Name(), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	require.NoError(t, err)
	require.Equal(t, int(l), 68157450)

	// try getting an object that does not exist
	err = minioClient.FGetObject(fmt.Sprintf("master.%s.%s", repo2, pfs.DefaultProjectName), "file", "foo", minio.GetObjectOptions{})
	bucketNotFoundError(t, err)

	// get the file that does exist
	outputFile, err := os.CreateTemp("", "pachyderm-test-large-objects-output-*")
	require.NoError(t, err)
	defer os.Remove(outputFile.Name())
	err = minioClient.FGetObject(fmt.Sprintf("master.%s.%s", repo1, pfs.DefaultProjectName), "file", outputFile.Name(), minio.GetObjectOptions{})
	require.True(t, err == nil || errors.Is(err, io.EOF), fmt.Sprintf("unexpected error: %s", err))

	// compare the files and ensure they're the same
	// NOTE: Because minio's `FGetObject` does a rename from a buffer file
	// to the given filepath, `outputFile` will refer to an empty, overwritten
	// file. We can still use `outputFile.Name()` though.
	inputFileSize, inputFileHash := fileHash(t, inputFile.Name())
	outputFileSize, outputFileHash := fileHash(t, inputFile.Name())
	require.Equal(t, inputFileSize, outputFileSize)
	require.Equal(t, inputFileHash, outputFileHash)
}

func projectMasterGetObjectNoHead(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobjectnohead")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "branch", "", "", nil))

	_, err := getObject(t, minioClient, fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName), "file")
	keyNotFoundError(t, err)
}

func projectMasterGetObjectNoBranch(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobjectnobranch")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))

	_, err := getObject(t, minioClient, fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName), "file")
	bucketNotFoundError(t, err)
}

func projectMasterGetObjectNoRepo(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testgetobjectnorepo")
	_, err := getObject(t, minioClient, fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "file")
	bucketNotFoundError(t, err)
}

func projectMasterMakeBucket(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucket")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), ""))

	repoInfo, err := pachClient.InspectRepo(pfs.DefaultProjectName, repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "master")
}

func projectMasterMakeBucketWithBranch(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucketwithbranch")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName), ""))

	repoInfo, err := pachClient.InspectRepo(pfs.DefaultProjectName, repo)
	require.NoError(t, err)
	require.Equal(t, len(repoInfo.Branches), 1)
	require.Equal(t, repoInfo.Branches[0].Name, "branch")
}

func projectMasterMakeBucketWithRegion(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucketwithregion")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "us-east-1"))
	_, err := pachClient.InspectRepo(pfs.DefaultProjectName, repo)
	require.NoError(t, err)
}

func projectMasterMakeBucketRedundant(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucketredundant")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), ""))
	err := minioClient.MakeBucket(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "")
	require.YesError(t, err)
	require.Equal(t, err.Error(), "The bucket you tried to create already exists, and you own it.")
}

func projectMasterMakeBucketDifferentBranches(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testmakebucketdifferentbranches")
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), ""))
	require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName), ""))
}

func projectMasterBucketExists(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testbucketexists")

	exists, err := minioClient.BucketExists(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName))
	require.NoError(t, err)
	require.False(t, exists)

	// repo exists, but branch doesn't: should be false
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	exists, err = minioClient.BucketExists(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName))
	require.NoError(t, err)
	require.False(t, exists)

	// repo and branch exists: should be true
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", nil))
	exists, err = minioClient.BucketExists(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName))
	require.NoError(t, err)
	require.True(t, exists)

	// repo exists, but branch doesn't: should be false
	exists, err = minioClient.BucketExists(fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName))
	require.NoError(t, err)
	require.False(t, exists)

	// repo and branch exists: should be true
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "branch", "", "", nil))
	exists, err = minioClient.BucketExists(fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName))
	require.NoError(t, err)
	require.True(t, exists)
}

func projectMasterRemoveBucket(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	// TODO(required 2.0): removing bucket does a WalkFile which errors when no files match (on an empty bucket)
	t.Skip("broken in 2.0 - WalkFile errors on an empty bucket")
	repo := tu.UniqueString("testremovebucket")

	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", nil))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "branch", "", "", nil))

	require.NoError(t, minioClient.RemoveBucket(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName)))
	require.NoError(t, minioClient.RemoveBucket(fmt.Sprintf("branch.%s.%s", repo, pfs.DefaultProjectName)))
}

func projectMasterRemoveBucketBranchless(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testremovebucketbranchless")

	// should error out because the repo doesn't have a branch
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	bucketNotFoundError(t, minioClient.RemoveBucket(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName)))
}

func projectMasterListObjectsPaginated(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	// create a bunch of files - enough to require the use of paginated
	// requests when browsing all files. One file will be included on a
	// separate branch to ensure it's not returned when querying against the
	// master branch.
	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	// S3 client limits bucket name length to 63 chars, but we also want to query with commit
	// so we need to be conservative with the length of the repo name here
	repo := tu.UniqueString("LOP")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	commit, err := pachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
	require.NoError(t, err)

	require.NoError(t, pachClient.WithModifyFileClient(commit, func(mf client.ModifyFile) error {
		for i := 0; i <= 1000; i++ {
			putListFileTestObject(t, mf, "", i)
		}

		for i := 0; i < 10; i++ {
			putListFileTestObject(t, mf, "dir/", i)
		}
		return nil
	}))

	require.NoError(t, pachClient.WithModifyFileClient(client.NewCommit(pfs.DefaultProjectName, repo, "branch", ""), func(mf client.ModifyFile) error {
		putListFileTestObject(t, mf, "", 1001)
		return nil
	}))

	require.NoError(t, pachClient.FinishCommit(pfs.DefaultProjectName, repo, commit.Branch.Name, commit.Id))

	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	// Request that will list all files in master's root
	ch := minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "", false, make(chan struct{}))
	expectedFiles := []string{}
	for i := 0; i <= 1000; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("%d", i))
	}
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{"dir/"})

	// Query by commit.repo (only works in the default project)
	ch = minioClient.ListObjects(fmt.Sprintf("%s.%s", commit.Id, repo), "", false, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{"dir/"})

	// Query by commit.branch.repo
	ch = minioClient.ListObjects(fmt.Sprintf("%s.%s.%s.%s", commit.Id, commit.Branch.Name, repo, pfs.DefaultProjectName), "", false, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{"dir/"})

	// Request that will list all files in master starting with 1
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "1", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i <= 1000; i++ {
		file := fmt.Sprintf("%d", i)
		if strings.HasPrefix(file, "1") {
			expectedFiles = append(expectedFiles, file)
		}
	}
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})

	// Request that will list all files in a directory in master
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "dir/", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i < 10; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("dir/%d", i))
	}
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})
}

func projectMasterListObjectsHeadlessBranch(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("testlistobjectsheadlessbranch")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "emptybranch", "", "", nil))

	// Request into branch that has no head
	ch := minioClient.ListObjects(fmt.Sprintf("emptybranch.%s.%s", repo, pfs.DefaultProjectName), "", false, make(chan struct{}))
	checkListObjects(t, ch, nil, nil, []string{}, []string{})
}

func projectMasterListObjectsRecursive(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	// `startTime` and `endTime` will be used to ensure that an object's
	// `LastModified` date is correct. A few minutes are subtracted/added to
	// each to tolerate the node time not being the same as the host time.
	startTime := time.Now().Add(time.Duration(-5) * time.Minute)
	repo := tu.UniqueString("testlistobjectsrecursive")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "branch", "", "", nil))
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "emptybranch", "", "", nil))

	require.NoError(t, pachClient.WithModifyFileClient(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), func(mf client.ModifyFile) error {
		putListFileTestObject(t, mf, "", 0)
		putListFileTestObject(t, mf, "rootdir/", 1)
		putListFileTestObject(t, mf, "rootdir/subdir/", 2)
		return nil
	}))

	require.NoError(t, pachClient.WithModifyFileClient(client.NewCommit(pfs.DefaultProjectName, repo, "branch", ""), func(mf client.ModifyFile) error {
		putListFileTestObject(t, mf, "", 3)
		return nil
	}))
	endTime := time.Now().Add(time.Duration(5) * time.Minute)

	// Request that will list all files in master
	expectedFiles := []string{"0", "rootdir/1", "rootdir/subdir/2"}
	ch := minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "", true, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})

	// Requests that will list all files in rootdir
	expectedFiles = []string{"rootdir/1", "rootdir/subdir/2"}
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "r", true, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "rootdir", true, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "rootdir/", true, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})

	// Requests that will list all files in subdir
	expectedFiles = []string{"rootdir/subdir/2"}
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "rootdir/s", true, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "rootdir/subdir", true, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "rootdir/subdir/", true, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})
	ch = minioClient.ListObjects(fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName), "rootdir/subdir/2", true, make(chan struct{}))
	checkListObjects(t, ch, &startTime, &endTime, expectedFiles, []string{})
}

func projectMasterDoesNotListSystemRepoBuckets(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	repo := tu.UniqueString("listsystemrepo")
	require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
	specRepo := client.NewSystemRepo(pfs.DefaultProjectName, repo, pfs.SpecRepoType)
	_, err := pachClient.PfsAPIClient.CreateRepo(pachClient.Ctx(), &pfs.CreateRepoRequest{Repo: specRepo})
	require.NoError(t, err)
	metaRepo := client.NewSystemRepo(pfs.DefaultProjectName, repo, pfs.MetaRepoType)
	_, err = pachClient.PfsAPIClient.CreateRepo(pachClient.Ctx(), &pfs.CreateRepoRequest{Repo: metaRepo})
	require.NoError(t, err)

	// create a master branch on each
	require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", nil))
	_, err = pachClient.PfsAPIClient.CreateBranch(pachClient.Ctx(), &pfs.CreateBranchRequest{Branch: specRepo.NewBranch("master")})
	require.NoError(t, err)
	_, err = pachClient.PfsAPIClient.CreateBranch(pachClient.Ctx(), &pfs.CreateBranchRequest{Branch: metaRepo.NewBranch("master")})
	require.NoError(t, err)

	buckets, err := minioClient.ListBuckets()
	require.NoError(t, err)

	for _, bucket := range buckets {
		if !strings.Contains(bucket.Name, repo) {
			// sometimes this encounters stray buckets from other tests
			continue
		}
		require.Equal(t, bucket.Name, fmt.Sprintf("master.%s.%s", repo, pfs.DefaultProjectName))
	}
}

func TestProjectMasterDriver(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	projectTestRunner(env.Context, t, env.PachClient, "master", s3.NewMasterDriver(), func(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
		t.Run("ListBuckets", func(t *testing.T) {
			projectMasterListBuckets(t, pachClient, minioClient)
		})
		t.Run("ListBucketsBranchless", func(t *testing.T) {
			projectMasterListBucketsBranchless(t, pachClient, minioClient)
		})
		t.Run("GetObject", func(t *testing.T) {
			projectMasterGetObject(t, pachClient, minioClient)
		})
		t.Run("GetObjectInBranch", func(t *testing.T) {
			projectMasterGetObjectInBranch(t, pachClient, minioClient)
		})
		t.Run("StatObject", func(t *testing.T) {
			projectMasterStatObject(t, pachClient, minioClient)
		})
		t.Run("PutObject", func(t *testing.T) {
			projectMasterPutObject(t, pachClient, minioClient)
		})
		t.Run("RemoveObject", func(t *testing.T) {
			projectMasterRemoveObject(t, pachClient, minioClient)
		})
		t.Run("LargeObjects", func(t *testing.T) {
			projectMasterLargeObjects(t, pachClient, minioClient)
		})
		t.Run("GetObjectNoHead", func(t *testing.T) {
			projectMasterGetObjectNoHead(t, pachClient, minioClient)
		})
		t.Run("GetObjectNoBranch", func(t *testing.T) {
			projectMasterGetObjectNoBranch(t, pachClient, minioClient)
		})
		t.Run("GetObjectNoRepo", func(t *testing.T) {
			projectMasterGetObjectNoRepo(t, pachClient, minioClient)
		})
		t.Run("MakeBucket", func(t *testing.T) {
			projectMasterMakeBucket(t, pachClient, minioClient)
		})
		t.Run("MakeBucketWithBranch", func(t *testing.T) {
			projectMasterMakeBucketWithBranch(t, pachClient, minioClient)
		})
		t.Run("MakeBucketWithRegion", func(t *testing.T) {
			projectMasterMakeBucketWithRegion(t, pachClient, minioClient)
		})
		t.Run("MakeBucketRedundant", func(t *testing.T) {
			projectMasterMakeBucketRedundant(t, pachClient, minioClient)
		})
		t.Run("MakeBucketDifferentBranches", func(t *testing.T) {
			projectMasterMakeBucketDifferentBranches(t, pachClient, minioClient)
		})
		t.Run("BucketExists", func(t *testing.T) {
			projectMasterBucketExists(t, pachClient, minioClient)
		})
		t.Run("RemoveBucket", func(t *testing.T) {
			projectMasterRemoveBucket(t, pachClient, minioClient)
		})
		t.Run("RemoveBucketBranchless", func(t *testing.T) {
			projectMasterRemoveBucketBranchless(t, pachClient, minioClient)
		})
		t.Run("ListObjectsPaginated", func(t *testing.T) {
			projectMasterListObjectsPaginated(t, pachClient, minioClient)
		})
		t.Run("ListObjectsHeadlessBranch", func(t *testing.T) {
			projectMasterListObjectsHeadlessBranch(t, pachClient, minioClient)
		})
		t.Run("ListObjectsRecursive", func(t *testing.T) {
			projectMasterListObjectsRecursive(t, pachClient, minioClient)
		})
		t.Run("DoesNotListSystemRepoBucket", func(t *testing.T) {
			projectMasterDoesNotListSystemRepoBuckets(t, pachClient, minioClient)
		})
	})
}
