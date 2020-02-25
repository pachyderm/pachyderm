package s3

import (
    // "fmt"
    // "io"
    // "io/ioutil"
    // "os"
    "strings"
    "testing"
    // "time"

    minio "github.com/minio/minio-go"
    "github.com/pachyderm/pachyderm/src/client"
    "github.com/pachyderm/pachyderm/src/client/pkg/require"
    tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func workerListBuckets(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
	// create a repo - this should not show up list buckets with the worker
	// driver
    repo := tu.UniqueString("testlistbuckets1")
    require.NoError(t, pachClient.CreateRepo(repo))
    require.NoError(t, pachClient.CreateBranch(repo, "master", "", nil))

    buckets, err := minioClient.ListBuckets()
    require.NoError(t, err)

    actualBucketNames := []string{}
    for _, bucket := range buckets {
    	actualBucketNames = append(actualBucketNames, bucket.Name)
    }

    require.ElementsEqual(t, []string{"in1", "in2", "out"}, actualBucketNames)
}

func workerGetObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
    fetchedContent, err := getObject(t, minioClient, "in1", "file")
    require.NoError(t, err)
    require.Equal(t, "foo", fetchedContent)
}

func workerGetObjectOutputRepo(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
    _, err := getObject(t, minioClient, "out", "file")
    keyNotFoundError(t, err)
}

func workerStatObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
    info, err := minioClient.StatObject("in1", "file", minio.StatObjectOptions{})
    require.NoError(t, err)
    require.True(t, len(info.ETag) > 0)
    require.Equal(t, "text/plain; charset=utf-8", info.ContentType)
    require.Equal(t, int64(3), info.Size)
}

// func workerPutObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
//     repo := tu.UniqueString("testputobject")
//     require.NoError(t, pachClient.CreateRepo(repo))
//     require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))

//     r := strings.NewReader("content1")
//     _, err := minioClient.PutObject(fmt.Sprintf("branch.%s", repo), "file", r, int64(r.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
//     require.NoError(t, err)

//     // this should act as a PFS PutFileOverwrite
//     r2 := strings.NewReader("content2")
//     _, err = minioClient.PutObject(fmt.Sprintf("branch.%s", repo), "file", r2, int64(r2.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
//     require.NoError(t, err)

//     fetchedContent, err := getObject(t, minioClient, repo, "branch", "file")
//     require.NoError(t, err)
//     require.Equal(t, "content2", fetchedContent)
// }

// func workerRemoveObject(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
//     repo := tu.UniqueString("testremoveobject")
//     require.NoError(t, pachClient.CreateRepo(repo))
//     _, err := pachClient.PutFile(repo, "master", "file", strings.NewReader("content"))
//     require.NoError(t, err)

//     // as per PFS semantics, the second delete should be a no-op
//     require.NoError(t, minioClient.RemoveObject(fmt.Sprintf("master.%s", repo), "file"))
//     require.NoError(t, minioClient.RemoveObject(fmt.Sprintf("master.%s", repo), "file"))

//     // make sure the object no longer exists
//     _, err = getObject(t, minioClient, repo, "master", "file")
//     keyNotFoundError(t, err)
// }

// // Tests inserting and getting files over 64mb in size
// func workerLargeObjects(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
//     // test repos: repo1 exists, repo2 does not
//     repo1 := tu.UniqueString("testlargeobject1")
//     repo2 := tu.UniqueString("testlargeobject2")
//     require.NoError(t, pachClient.CreateRepo(repo1))
//     require.NoError(t, pachClient.CreateBranch(repo1, "master", "", nil))

//     // create a temporary file to put ~65mb of contents into it
//     inputFile, err := ioutil.TempFile("", "pachyderm-test-large-objects-input-*")
//     require.NoError(t, err)
//     defer os.Remove(inputFile.Name())
//     n, err := inputFile.WriteString(strings.Repeat("no tv and no beer make homer something something.\n", 1363149))
//     require.NoError(t, err)
//     require.Equal(t, n, 68157450)
//     require.NoError(t, inputFile.Sync())

//     // first ensure that putting into a repo that doesn't exist triggers an
//     // error
//     _, err = minioClient.FPutObject(fmt.Sprintf("master.%s", repo2), "file", inputFile.Name(), minio.PutObjectOptions{
//         ContentType: "text/plain",
//     })
//     bucketNotFoundError(t, err)

//     // now try putting into a legit repo
//     l, err := minioClient.FPutObject(fmt.Sprintf("master.%s", repo1), "file", inputFile.Name(), minio.PutObjectOptions{
//         ContentType: "text/plain",
//     })
//     require.NoError(t, err)
//     require.Equal(t, int(l), 68157450)

//     // try getting an object that does not exist
//     err = minioClient.FGetObject(fmt.Sprintf("master.%s", repo2), "file", "foo", minio.GetObjectOptions{})
//     bucketNotFoundError(t, err)

//     // get the file that does exist
//     outputFile, err := ioutil.TempFile("", "pachyderm-test-large-objects-output-*")
//     require.NoError(t, err)
//     defer os.Remove(outputFile.Name())
//     err = minioClient.FGetObject(fmt.Sprintf("master.%s", repo1), "file", outputFile.Name(), minio.GetObjectOptions{})
//     require.True(t, err == nil || err == io.EOF, fmt.Sprintf("unexpected error: %s", err))

//     // compare the files and ensure they're the same
//     // NOTE: Because minio's `FGetObject` does a rename from a buffer file
//     // to the given filepath, `outputFile` will refer to an empty, overwritten
//     // file. We can still use `outputFile.Name()` though.
//     inputFileSize, inputFileHash := fileHash(t, inputFile.Name())
//     outputFileSize, outputFileHash := fileHash(t, inputFile.Name())
//     require.Equal(t, inputFileSize, outputFileSize)
//     require.Equal(t, inputFileHash, outputFileHash)
// }

// func workerGetObjectNoRepo(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
//     repo := tu.UniqueString("testgetobjectnorepo")
//     _, err := getObject(t, minioClient, repo, "master", "file")
//     bucketNotFoundError(t, err)
// }

// func workerMakeBucket(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
//     repo := tu.UniqueString("testmakebucket")
//     require.NoError(t, minioClient.MakeBucket(fmt.Sprintf("master.%s", repo), ""))

//     repoInfo, err := pachClient.InspectRepo(repo)
//     require.NoError(t, err)
//     require.Equal(t, len(repoInfo.Branches), 1)
//     require.Equal(t, repoInfo.Branches[0].Name, "master")
// }

// func workerBucketExists(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
//     repo := tu.UniqueString("testbucketexists")

//     exists, err := minioClient.BucketExists(fmt.Sprintf("master.%s", repo))
//     require.NoError(t, err)
//     require.False(t, exists)

//     // repo exists, but branch doesn't: should be false
//     require.NoError(t, pachClient.CreateRepo(repo))
//     exists, err = minioClient.BucketExists(fmt.Sprintf("master.%s", repo))
//     require.NoError(t, err)
//     require.False(t, exists)

//     // repo and branch exists: should be true
//     require.NoError(t, pachClient.CreateBranch(repo, "master", "", nil))
//     exists, err = minioClient.BucketExists(fmt.Sprintf("master.%s", repo))
//     require.NoError(t, err)
//     require.True(t, exists)

//     // repo exists, but branch doesn't: should be false
//     exists, err = minioClient.BucketExists(fmt.Sprintf("branch.%s", repo))
//     require.NoError(t, err)
//     require.False(t, exists)

//     // repo and branch exists: should be true
//     require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))
//     exists, err = minioClient.BucketExists(fmt.Sprintf("branch.%s", repo))
//     require.NoError(t, err)
//     require.True(t, exists)
// }

// func workerRemoveBucket(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
//     repo := tu.UniqueString("testremovebucket")

//     require.NoError(t, pachClient.CreateRepo(repo))
//     require.NoError(t, pachClient.CreateBranch(repo, "master", "", nil))
//     require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))

//     require.NoError(t, minioClient.RemoveBucket(fmt.Sprintf("master.%s", repo)))
//     require.NoError(t, minioClient.RemoveBucket(fmt.Sprintf("branch.%s", repo)))
// }

// func workerListObjectsPaginated(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
//     // create a bunch of files - enough to require the use of paginated
//     // requests when browsing all files. One file will be included on a
//     // separate branch to ensure it's not returned when querying against the
//     // master branch.
//     // `startTime` and `endTime` will be used to ensure that an object's
//     // `LastModified` date is correct. A few minutes are subtracted/added to
//     // each to tolerate the node time not being the same as the host time.
//     startTime := time.Now().Add(time.Duration(-5) * time.Minute)
//     repo := tu.UniqueString("testlistobjectspaginated")
//     require.NoError(t, pachClient.CreateRepo(repo))
//     commit, err := pachClient.StartCommit(repo, "master")
//     require.NoError(t, err)
//     for i := 0; i <= 1000; i++ {
//         putListFileTestObject(t, pachClient, repo, commit.ID, "", i)
//     }
//     for i := 0; i < 10; i++ {
//         putListFileTestObject(t, pachClient, repo, commit.ID, "dir/", i)
//         require.NoError(t, err)
//     }
//     putListFileTestObject(t, pachClient, repo, "branch", "", 1001)
//     require.NoError(t, pachClient.FinishCommit(repo, commit.ID))
//     endTime := time.Now().Add(time.Duration(5) * time.Minute)

//     // Request that will list all files in master's root
//     ch := minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "", false, make(chan struct{}))
//     expectedFiles := []string{}
//     for i := 0; i <= 1000; i++ {
//         expectedFiles = append(expectedFiles, fmt.Sprintf("%d", i))
//     }
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{"dir/"})

//     // Request that will list all files in master starting with 1
//     ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "1", false, make(chan struct{}))
//     expectedFiles = []string{}
//     for i := 0; i <= 1000; i++ {
//         file := fmt.Sprintf("%d", i)
//         if strings.HasPrefix(file, "1") {
//             expectedFiles = append(expectedFiles, file)
//         }
//     }
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

//     // Request that will list all files in a directory in master
//     ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "dir/", false, make(chan struct{}))
//     expectedFiles = []string{}
//     for i := 0; i < 10; i++ {
//         expectedFiles = append(expectedFiles, fmt.Sprintf("dir/%d", i))
//     }
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
// }

// func workerListObjectsRecursive(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
//     // `startTime` and `endTime` will be used to ensure that an object's
//     // `LastModified` date is correct. A few minutes are subtracted/added to
//     // each to tolerate the node time not being the same as the host time.
//     startTime := time.Now().Add(time.Duration(-5) * time.Minute)
//     repo := tu.UniqueString("testlistobjectsrecursive")
//     require.NoError(t, pachClient.CreateRepo(repo))
//     require.NoError(t, pachClient.CreateBranch(repo, "branch", "", nil))
//     require.NoError(t, pachClient.CreateBranch(repo, "emptybranch", "", nil))
//     commit, err := pachClient.StartCommit(repo, "master")
//     require.NoError(t, err)
//     putListFileTestObject(t, pachClient, repo, commit.ID, "", 0)
//     putListFileTestObject(t, pachClient, repo, commit.ID, "rootdir/", 1)
//     putListFileTestObject(t, pachClient, repo, commit.ID, "rootdir/subdir/", 2)
//     putListFileTestObject(t, pachClient, repo, "branch", "", 3)
//     require.NoError(t, pachClient.FinishCommit(repo, commit.ID))
//     endTime := time.Now().Add(time.Duration(5) * time.Minute)

//     // Request that will list all files in master
//     expectedFiles := []string{"0", "rootdir/1", "rootdir/subdir/2"}
//     ch := minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "", true, make(chan struct{}))
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

//     // Requests that will list all files in rootdir
//     expectedFiles = []string{"rootdir/1", "rootdir/subdir/2"}
//     ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "r", true, make(chan struct{}))
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
//     ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir", true, make(chan struct{}))
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
//     ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/", true, make(chan struct{}))
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})

//     // Requests that will list all files in subdir
//     expectedFiles = []string{"rootdir/subdir/2"}
//     ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/s", true, make(chan struct{}))
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
//     ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/subdir", true, make(chan struct{}))
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
//     ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/subdir/", true, make(chan struct{}))
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
//     ch = minioClient.ListObjects(fmt.Sprintf("master.%s", repo), "rootdir/subdir/2", true, make(chan struct{}))
//     checkListObjects(t, ch, startTime, endTime, expectedFiles, []string{})
// }

func TestWorkerDriver(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration tests in short mode")
    }

    pachClient, err := client.NewForTest()
    require.NoError(t, err)

    inputRepo := tu.UniqueString("testworkerdriverinput")
    require.NoError(t, pachClient.CreateRepo(inputRepo))
    outputRepo := tu.UniqueString("testworkerdriveroutput")
    require.NoError(t, pachClient.CreateRepo(outputRepo))

    // create a master branch on the input repo
    inputMasterCommit, err := pachClient.StartCommit(inputRepo, "master")
	require.NoError(t, err)
	_, err = pachClient.PutFile(inputRepo, inputMasterCommit.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, pachClient.FinishCommit(inputRepo, inputMasterCommit.ID))

	// create a develop branch on the input repo
    inputDevelopCommit, err := pachClient.StartCommit(inputRepo, "develop")
	require.NoError(t, err)
	_, err = pachClient.PutFile(inputRepo, inputDevelopCommit.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, pachClient.FinishCommit(inputRepo, inputDevelopCommit.ID))

	// create the output branch
    outputCommit, err := pachClient.StartCommit(outputRepo, "master")
	require.NoError(t, err)

    driver := NewWorkerDriver(
        []*Bucket{
            &Bucket{
                Repo: inputRepo,
                Commit: inputMasterCommit.ID,
                Name: "in1",
            },
            &Bucket{
                Repo: inputRepo,
                Commit: inputDevelopCommit.ID,
                Name: "in2",
            },
        },
        &Bucket{
            Repo: outputRepo,
            Commit: outputCommit.ID,
            Name: "out",
        },
    )

    testRunner(t, "worker", driver, func(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
        t.Run("ListBuckets", func(t *testing.T) {
            workerListBuckets(t, pachClient, minioClient)
        })
        t.Run("GetObject", func(t *testing.T) {
            workerGetObject(t, pachClient, minioClient)
        })
        t.Run("GetObjectOutputRepo", func(t *testing.T) {
            workerGetObjectOutputRepo(t, pachClient, minioClient)
        })
        t.Run("StatObject", func(t *testing.T) {
            workerStatObject(t, pachClient, minioClient)
        })
        // t.Run("PutObject", func(t *testing.T) {
        //     workerPutObject(t, pachClient, minioClient)
        // })
        // t.Run("RemoveObject", func(t *testing.T) {
        //     workerRemoveObject(t, pachClient, minioClient)
        // })
        // t.Run("LargeObjects", func(t *testing.T) {
        //     workerLargeObjects(t, pachClient, minioClient)
        // })
        // t.Run("GetObjectNoRepo", func(t *testing.T) {
        //     workerGetObjectNoRepo(t, pachClient, minioClient)
        // })
        // t.Run("MakeBucket", func(t *testing.T) {
        //     workerMakeBucket(t, pachClient, minioClient)
        // })
        // t.Run("BucketExists", func(t *testing.T) {
        //     workerBucketExists(t, pachClient, minioClient)
        // })
        // t.Run("RemoveBucket", func(t *testing.T) {
        //     workerRemoveBucket(t, pachClient, minioClient)
        // })
        // t.Run("ListObjectsPaginated", func(t *testing.T) {
        //     workerListObjectsPaginated(t, pachClient, minioClient)
        // })
        // t.Run("ListObjectsRecursive", func(t *testing.T) {
        //     workerListObjectsRecursive(t, pachClient, minioClient)
        // })
    })
}
