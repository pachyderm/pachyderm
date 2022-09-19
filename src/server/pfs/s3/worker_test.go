package s3_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	minio "github.com/minio/minio-go/v6"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/s3"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

type workerTestState struct {
	pachClient         *client.APIClient
	minioClient        *minio.Client
	inputRepo          string
	outputRepo         string
	inputMasterCommit  *pfs.Commit
	inputDevelopCommit *pfs.Commit
	outputBranch       string
}

func workerListBuckets(t *testing.T, s *workerTestState) {
	// create a repo - this should not show up list buckets with the worker
	// driver
	repo := tu.UniqueString("testlistbuckets1")
	require.NoError(t, s.pachClient.CreateRepo(repo))
	require.NoError(t, s.pachClient.CreateBranch(repo, "master", "", "", nil))

	buckets, err := s.minioClient.ListBuckets()
	require.NoError(t, err)

	actualBucketNames := []string{}
	for _, bucket := range buckets {
		actualBucketNames = append(actualBucketNames, bucket.Name)
	}

	require.ElementsEqual(t, []string{"in1", "in2", "out"}, actualBucketNames)
}

func workerGetObject(t *testing.T, s *workerTestState) {
	fetchedContent, err := getObject(t, s.minioClient, "in1", "0")
	require.NoError(t, err)
	require.Equal(t, "0\n", fetchedContent)
}

func workerGetObjectOutputRepo(t *testing.T, s *workerTestState) {
	_, err := getObject(t, s.minioClient, "out", "file")
	keyNotFoundError(t, err)
}

func workerStatObject(t *testing.T, s *workerTestState) {
	info, err := s.minioClient.StatObject("in1", "0", minio.StatObjectOptions{})
	require.NoError(t, err)
	require.True(t, len(info.ETag) > 0)
	require.Equal(t, "text/plain; charset=utf-8", info.ContentType)
	require.Equal(t, int64(2), info.Size)
}

func workerPutObject(t *testing.T, s *workerTestState) {
	r := strings.NewReader("content1")
	_, err := s.minioClient.PutObject("out", "file", r, int64(r.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
	require.NoError(t, err)

	// this should act as a PFS PutFile
	r2 := strings.NewReader("content2")
	_, err = s.minioClient.PutObject("out", "file", r2, int64(r2.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
	require.NoError(t, err)

	_, err = getObject(t, s.minioClient, "out", "file")
	keyNotFoundError(t, err)
}

func workerPutObjectInputRepo(t *testing.T, s *workerTestState) {
	r := strings.NewReader("content1")
	_, err := s.minioClient.PutObject("in1", "0", r, int64(r.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
	notImplementedError(t, err)
}

func workerRemoveObject(t *testing.T, s *workerTestState) {
	require.NoError(t, s.pachClient.PutFile(client.NewCommit(s.outputRepo, s.outputBranch, ""), "file", strings.NewReader("content")))

	// as per PFS semantics, the second delete should be a no-op
	require.NoError(t, s.minioClient.RemoveObject("out", "file"))
	require.NoError(t, s.minioClient.RemoveObject("out", "file"))
}

func workerRemoveObjectInputRepo(t *testing.T, s *workerTestState) {
	err := s.minioClient.RemoveObject("in1", "0")
	notImplementedError(t, err)
}

// Tests inserting and getting files over 64mb in size
func workerLargeObjects(t *testing.T, s *workerTestState) {
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
	_, err = s.minioClient.FPutObject("foobar", "file", inputFile.Name(), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	bucketNotFoundError(t, err)

	// now try putting into a legit repo
	l, err := s.minioClient.FPutObject("out", "file", inputFile.Name(), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	require.NoError(t, err)
	require.Equal(t, int(l), 68157450)

	// try getting an object that does not exist
	err = s.minioClient.FGetObject("foobar", "file", "foo", minio.GetObjectOptions{})
	bucketNotFoundError(t, err)

	// get the file that does exist, doesn't work because we're reading from
	// an output repo
	outputFile, err := os.CreateTemp("", "pachyderm-test-large-objects-output-*")
	require.NoError(t, err)
	defer os.Remove(outputFile.Name())
	err = s.minioClient.FGetObject("out", "file", outputFile.Name(), minio.GetObjectOptions{})
	keyNotFoundError(t, err)
}

func workerMakeBucket(t *testing.T, s *workerTestState) {
	repo := tu.UniqueString("testmakebucket")
	notImplementedError(t, s.minioClient.MakeBucket(repo, ""))
}

func workerBucketExists(t *testing.T, s *workerTestState) {
	exists, err := s.minioClient.BucketExists("in1")
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = s.minioClient.BucketExists("out")
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = s.minioClient.BucketExists("foobar")
	require.NoError(t, err)
	require.False(t, exists)
}

func workerRemoveBucket(t *testing.T, s *workerTestState) {
	notImplementedError(t, s.minioClient.RemoveBucket("in1"))
	notImplementedError(t, s.minioClient.RemoveBucket("out"))
}

func workerListObjectsPaginated(t *testing.T, s *workerTestState) {
	// Request that will list all files in root
	ch := s.minioClient.ListObjects("in2", "", false, make(chan struct{}))
	expectedFiles := []string{}
	for i := 0; i <= 100; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("%d", i))
	}
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{"dir/"})

	// Request that will list all files in with / as a prefix ("/" should mean
	// the same as "", e.g. rust-s3 client)
	ch = s.minioClient.ListObjects("in2", "/", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i <= 100; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("%d", i))
	}
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{"dir/"})

	// Request that will list all files starting with 1
	ch = s.minioClient.ListObjects("in2", "1", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i <= 100; i++ {
		file := fmt.Sprintf("%d", i)
		if strings.HasPrefix(file, "1") {
			expectedFiles = append(expectedFiles, file)
		}
	}
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})

	// Request that will list all files in a directory
	ch = s.minioClient.ListObjects("in2", "dir/", false, make(chan struct{}))
	expectedFiles = []string{}
	for i := 0; i < 10; i++ {
		expectedFiles = append(expectedFiles, fmt.Sprintf("dir/%d", i))
	}
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})
}

func workerListObjectsRecursive(t *testing.T, s *workerTestState) {
	// Request that will list all files in master
	expectedFiles := []string{"0", "rootdir/1", "rootdir/subdir/2"}
	ch := s.minioClient.ListObjects("in1", "", true, make(chan struct{}))
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})

	// Requests that will list all files in rootdir
	expectedFiles = []string{"rootdir/1", "rootdir/subdir/2"}
	ch = s.minioClient.ListObjects("in1", "r", true, make(chan struct{}))
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})
	ch = s.minioClient.ListObjects("in1", "rootdir", true, make(chan struct{}))
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})
	ch = s.minioClient.ListObjects("in1", "rootdir/", true, make(chan struct{}))
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})

	// Requests that will list all files in subdir
	expectedFiles = []string{"rootdir/subdir/2"}
	ch = s.minioClient.ListObjects("in1", "rootdir/s", true, make(chan struct{}))
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})
	ch = s.minioClient.ListObjects("in1", "rootdir/subdir", true, make(chan struct{}))
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})
	ch = s.minioClient.ListObjects("in1", "rootdir/subdir/", true, make(chan struct{}))
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})
	ch = s.minioClient.ListObjects("in1", "rootdir/subdir/2", true, make(chan struct{}))
	checkListObjects(t, ch, nil, nil, expectedFiles, []string{})
}

func TestWorkerDriver(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	pachClient := env.PachClient

	inputRepo := tu.UniqueString("testworkerdriverinput")
	require.NoError(t, pachClient.CreateRepo(inputRepo))
	outputRepo := tu.UniqueString("testworkerdriveroutput")
	require.NoError(t, pachClient.CreateRepo(outputRepo))

	// create a master branch on the input repo
	inputMasterCommit, err := pachClient.StartCommit(inputRepo, "master")
	require.NoError(t, err)

	require.NoError(t, pachClient.WithModifyFileClient(inputMasterCommit, func(mf client.ModifyFile) error {
		putListFileTestObject(t, mf, "", 0)
		putListFileTestObject(t, mf, "rootdir/", 1)
		putListFileTestObject(t, mf, "rootdir/subdir/", 2)
		return nil
	}))
	require.NoError(t, pachClient.FinishCommit(inputRepo, inputMasterCommit.Branch.Name, inputMasterCommit.ID))

	// create a develop branch on the input repo
	inputDevelopCommit, err := pachClient.StartCommit(inputRepo, "develop")
	require.NoError(t, err)

	require.NoError(t, pachClient.WithModifyFileClient(inputDevelopCommit, func(mf client.ModifyFile) error {
		for i := 0; i <= 100; i++ {
			putListFileTestObject(t, mf, "", i)
		}
		for i := 0; i < 10; i++ {
			putListFileTestObject(t, mf, "dir/", i)
		}
		return nil
	}))
	require.NoError(t, pachClient.FinishCommit(inputRepo, inputDevelopCommit.Branch.Name, inputDevelopCommit.ID))

	// create the output branch
	outputBranch := "master"
	require.NoError(t, pachClient.CreateBranch(outputRepo, outputBranch, "", "", nil))

	driver := s3.NewWorkerDriver(
		[]*s3.Bucket{
			{
				Commit: inputMasterCommit,
				Name:   "in1",
			},
			{
				Commit: inputDevelopCommit,
				Name:   "in2",
			},
		},
		&s3.Bucket{
			Commit: client.NewRepo(outputRepo).NewCommit(outputBranch, ""),
			Name:   "out",
		},
	)

	testRunner(t, pachClient, "worker", driver, func(t *testing.T, pachClient *client.APIClient, minioClient *minio.Client) {
		s := &workerTestState{
			pachClient:         pachClient,
			minioClient:        minioClient,
			inputRepo:          inputRepo,
			outputRepo:         outputRepo,
			inputMasterCommit:  inputMasterCommit,
			inputDevelopCommit: inputDevelopCommit,
			outputBranch:       outputBranch,
		}

		t.Run("ListBuckets", func(t *testing.T) {
			workerListBuckets(t, s)
		})
		t.Run("GetObject", func(t *testing.T) {
			workerGetObject(t, s)
		})
		t.Run("GetObjectOutputRepo", func(t *testing.T) {
			workerGetObjectOutputRepo(t, s)
		})
		t.Run("StatObject", func(t *testing.T) {
			workerStatObject(t, s)
		})
		t.Run("PutObject", func(t *testing.T) {
			workerPutObject(t, s)
		})
		t.Run("PutObjectInputRepo", func(t *testing.T) {
			workerPutObjectInputRepo(t, s)
		})
		t.Run("RemoveObject", func(t *testing.T) {
			workerRemoveObject(t, s)
		})
		t.Run("RemoveObjectInputRepo", func(t *testing.T) {
			workerRemoveObjectInputRepo(t, s)
		})
		t.Run("LargeObjects", func(t *testing.T) {
			workerLargeObjects(t, s)
		})
		t.Run("MakeBucket", func(t *testing.T) {
			workerMakeBucket(t, s)
		})
		t.Run("BucketExists", func(t *testing.T) {
			workerBucketExists(t, s)
		})
		t.Run("RemoveBucket", func(t *testing.T) {
			workerRemoveBucket(t, s)
		})
		t.Run("ListObjectsPaginated", func(t *testing.T) {
			workerListObjectsPaginated(t, s)
		})
		t.Run("ListObjectsRecursive", func(t *testing.T) {
			workerListObjectsRecursive(t, s)
		})
	})
}
