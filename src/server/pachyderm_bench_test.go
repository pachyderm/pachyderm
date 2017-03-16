package server

import (
	"fmt"
	"math/rand"
	"path"
	"sync/atomic"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
	"golang.org/x/sync/errgroup"
)

const (
	KB               = 1024
	MB               = 1024 * KB
	GB               = 1024 * MB
	lettersAndSpaces = "abcdefghijklmnopqrstuvwxyz      "
)

// countWriter increments a counter by the number of bytes given
type countWriter struct {
	count int64
}

func (w *countWriter) Write(p []byte) (int, error) {
	atomic.AddInt64(&w.count, int64(len(p)))
	return len(p), nil
}

func getRand() *rand.Rand {
	source := rand.Int63()
	return rand.New(rand.NewSource(source))
}

func BenchmarkDailyManySmallFiles(b *testing.B) {
	benchmarkFiles(b, 1000000, 100, 10*KB, false)
}

func BenchmarkDailySomeLargeFiles(b *testing.B) {
	benchmarkFiles(b, 100, 100*MB, 30*GB, false)
}

func BenchmarkLocalSmallFiles(b *testing.B) {
	benchmarkFiles(b, 1000, 100, 10*KB, true)
}

// benchmarkFiles runs a benchmarks that uploads, downloads, and processes
// fileNum files, whose sizes (in bytes) are produced by a zipf
// distribution with the given min and max.
func benchmarkFiles(b *testing.B, fileNum int, minSize uint64, maxSize uint64, local bool) {
	repo := uniqueString("BenchmarkPachydermFiles")
	var c *client.APIClient
	var err error
	if local {
		c, err = client.NewFromAddress("localhost:30650")
	} else {
		c, err = client.NewInCluster()
	}
	require.NoError(b, err)
	require.NoError(b, c.CreateRepo(repo))

	commit, err := c.StartCommit(repo, "")
	require.NoError(b, err)
	if !b.Run(fmt.Sprintf("Put%dFiles", fileNum), func(b *testing.B) {
		var totalBytes int64
		var eg errgroup.Group
		r := getRand()
		zipf := rand.NewZipf(r, 1.05, 1, maxSize-minSize)
		for k := 0; k < fileNum; k++ {
			k := k
			fileSize := int64(zipf.Uint64() + minSize)
			totalBytes += fileSize
			eg.Go(func() error {
				var err error
				if k%10 == 0 {
					_, err = c.PutFile(repo, commit.ID, fmt.Sprintf("dir/file%d", k), workload.NewReader(getRand(), fileSize))
				} else {
					_, err = c.PutFile(repo, commit.ID, fmt.Sprintf("file%d", k), workload.NewReader(getRand(), fileSize))
				}
				return err
			})
		}
		require.NoError(b, eg.Wait())
		b.SetBytes(totalBytes)
	}) {
		return
	}
	require.NoError(b, c.FinishCommit(repo, commit.ID))
	require.NoError(b, c.SetBranch(repo, commit.ID, "master"))

	if !b.Run(fmt.Sprintf("Get%dFiles", fileNum), func(b *testing.B) {
		w := &countWriter{}
		var eg errgroup.Group
		for k := 0; k < fileNum; k++ {
			k := k
			eg.Go(func() error {
				if k%10 == 0 {
					return c.GetFile(repo, commit.ID, fmt.Sprintf("dir/file%d", k), 0, 0, w)
				} else {
					return c.GetFile(repo, commit.ID, fmt.Sprintf("file%d", k), 0, 0, w)
				}
			})
		}
		require.NoError(b, eg.Wait())
		defer func() { b.SetBytes(w.count) }()
	}) {
		return
	}

	if !b.Run(fmt.Sprintf("PipelineCopy%dFiles", fileNum), func(b *testing.B) {
		pipeline := uniqueString("BenchmarkPachydermPipeline")
		require.NoError(b, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{fmt.Sprintf("cp -R %s /pfs/out/", path.Join("/pfs", repo, "/*"))},
			&ppsclient.ParallelismSpec{
				Strategy: ppsclient.ParallelismSpec_CONSTANT,
				Constant: 4,
			},
			[]*ppsclient.PipelineInput{{
				Repo: client.NewRepo(repo),
				Glob: "/*",
			}},
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit(repo, commit.ID)}, nil)
		require.NoError(b, err)
		_, err = commitIter.Next()
		require.NoError(b, err)
		b.StopTimer()
		inputRepoInfo, err := c.InspectRepo(repo)
		require.NoError(b, err)
		outputRepoInfo, err := c.InspectRepo(pipeline)
		require.NoError(b, err)
		require.Equal(b, inputRepoInfo.SizeBytes, outputRepoInfo.SizeBytes)
		b.SetBytes(int64(inputRepoInfo.SizeBytes + outputRepoInfo.SizeBytes))
	}) {
		return
	}
}

func BenchmarkDailyPutLargeFileViaS3(b *testing.B) {
	repo := uniqueString("BenchmarkDailyPutLargeFileViaS3")
	c, err := client.NewInCluster()
	require.NoError(b, err)
	require.NoError(b, c.CreateRepo(repo))
	for i := 0; i < b.N; i++ {
		commit, err := c.StartCommit(repo, "")
		require.NoError(b, err)
		err = c.PutFileURL(repo, "master", "/", "s3://pachyderm-internal-benchmark/bigfiles/1gb.bytes", false)
		require.NoError(b, err)
		require.NoError(b, c.FinishCommit(repo, commit.ID))
		b.SetBytes(int64(1024 * MB))
	}
}

// BenchmarkDailyDataShuffle consists of the following steps:
//
// 1. Putting N tarballs into a Pachyderm cluster
// 2. Extracting M files from these tarballs, where M > N
// 3. Processing those M files
// 4. Compressing the resulting files into K tarballs
//
// It's essentially a data shuffle.
func BenchmarkDailyDataShuffle(b *testing.B) {
}
