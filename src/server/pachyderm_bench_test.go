package server

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	pfs "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	pps "github.com/pachyderm/pachyderm/src/client/pps"
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
	return rand.New(rand.NewSource(time.Now().Unix()))
}

func BenchmarkDailySmallFiles(b *testing.B) {
	benchmarkFiles(b, 1000000, 100, 10*KB, false)
}

func BenchmarkDailyLargeFiles(b *testing.B) {
	benchmarkFiles(b, 100, 100*MB, 30*GB, false)
}

func BenchmarkLocalSmallFiles(b *testing.B) {
	benchmarkFiles(b, 1000, 100, 10*KB, true)
}

func BenchmarkLocalBigFiles(b *testing.B) {
	benchmarkFiles(b, 100, 1*MB, 10*MB, true)
}

// getClient returns a Pachyderm client that connects to either a
// local cluster or a remote one, depending on the LOCAL env variable.
func getClient() (*client.APIClient, error) {
	if os.Getenv("LOCAL") != "" {
		c, err := client.NewFromAddress("localhost:30650")
		return c, err
	} else {
		c, err := client.NewInCluster()
		return c, err
	}
}

// benchmarkFiles runs a benchmarks that uploads, downloads, and processes
// fileNum files, whose sizes (in bytes) are produced by a zipf
// distribution with the given min and max.
func benchmarkFiles(b *testing.B, fileNum int, minSize uint64, maxSize uint64, local bool) {
	scalePachd(b)

	repo := uniqueString("BenchmarkPachydermFiles")
	c, err := getClient()
	require.NoError(b, err)
	require.NoError(b, c.CreateRepo(repo))

	commit, err := c.StartCommit(repo, "")
	require.NoError(b, err)
	if !b.Run(fmt.Sprintf("Put%dFiles", fileNum), func(b *testing.B) {
		b.N = 1
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
		b.N = 1
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
		b.N = 1
		pipeline := uniqueString("BenchmarkFilesPipeline")
		require.NoError(b, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{fmt.Sprintf("cp -R %s /pfs/out/", path.Join("/pfs", repo, "/*"))},
			&pps.ParallelismSpec{
				Strategy: pps.ParallelismSpec_CONSTANT,
				Constant: 4,
			},
			[]*pps.PipelineInput{{
				Repo: client.NewRepo(repo),
				Glob: "/*",
			}},
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(repo, commit.ID)}, nil)
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

func BenchmarkDailyDataShuffle(b *testing.B) {
	// The following parameters, combined with the values given to the zipf
	// function (which we use to generate file sizes), give us a workload
	// with 1TB of data that consists of 20 tarballs, each of which has 10000
	// files, whose sizes are between 1KB and 100MB.
	benchmarkDataShuffle(b, 20, 1000, 1*KB, 100*MB, 10)
}

func BenchmarkLocalDataShuffle(b *testing.B) {
	benchmarkDataShuffle(b, 10, 100, 100, 10*MB, 1)
}

// benchmarkDailyDataShuffle consists of the following steps:
//
// 1. Putting N tarballs into a Pachyderm cluster
// 2. Extracting M files from these tarballs, where M > N
// 3. Processing those M files
// 4. Compressing the resulting files into K tarballs
//
// It's essentially a data shuffle.
//
// `round` specifies how many extra rounds the entire pipeline runs.
func benchmarkDataShuffle(b *testing.B, numTarballs int, numFilesPerTarball int, minFileSize uint64, maxFileSize uint64, round int) {
	scalePachd(b)

	numTotalFiles := numTarballs * numFilesPerTarball
	dataRepo := uniqueString("BenchmarkDataShuffle")
	c, err := getClient()
	require.NoError(b, err)
	require.NoError(b, c.CreateRepo(dataRepo))

	// addInputCommit adds an input commit to the data repo
	addInputCommit := func(b *testing.B) *pfs.Commit {
		b.N = 1
		commit, err := c.StartCommit(dataRepo, "")
		require.NoError(b, err)
		var totalSize int64
		// the group that generates data
		// the group that writes data to pachd
		var genEg errgroup.Group
		var writeEg errgroup.Group
		for i := 0; i < numTarballs; i++ {
			tarName := fmt.Sprintf("tar-%d.tar.gz", i)
			pr, pw := io.Pipe()
			genEg.Go(func() (retErr error) {
				defer func() {
					if err := pw.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				gw := gzip.NewWriter(pw)
				defer gw.Close()
				defer func() {
					if err := gw.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				tw := tar.NewWriter(gw)
				defer tw.Close()
				defer func() {
					if err := tw.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				r := getRand()
				zipf := rand.NewZipf(r, 1.01, 1, maxFileSize-minFileSize)
				for j := 0; j < numFilesPerTarball; j++ {
					fileName := workload.RandString(r, 10)
					fileSize := int64(zipf.Uint64() + minFileSize)
					atomic.AddInt64(&totalSize, fileSize)
					hdr := &tar.Header{
						Name: fileName,
						Mode: 0600,
						Size: fileSize,
					}
					if err := tw.WriteHeader(hdr); err != nil {
						return err
					}
					data, err := ioutil.ReadAll(workload.NewReader(r, fileSize))
					if err != nil {
						return err
					}
					if _, err := tw.Write(data); err != nil {
						return err
					}
				}
				return nil
			})
			writeEg.Go(func() error {
				c, err := getClient()
				if err != nil {
					return err
				}
				_, err = c.PutFile(dataRepo, commit.ID, tarName, pr)
				return err
			})
		}
		require.NoError(b, genEg.Wait())
		require.NoError(b, writeEg.Wait())
		require.NoError(b, c.FinishCommit(dataRepo, commit.ID))
		require.NoError(b, c.SetBranch(dataRepo, commit.ID, "master"))
		commitInfo, err := c.InspectCommit(dataRepo, commit.ID)
		require.NoError(b, err)
		b.SetBytes(int64(commitInfo.SizeBytes))
		return commit
	}

	var commit *pfs.Commit
	if !b.Run(fmt.Sprintf("Put%dTarballs", numTarballs), func(b *testing.B) {
		commit = addInputCommit(b)
	}) {
		return
	}

	pipelineOne := uniqueString("BenchmarkDataShuffleStageOne")
	if !b.Run(fmt.Sprintf("Extract%dTarballsInto%dFiles", numTarballs, numTotalFiles), func(b *testing.B) {
		b.N = 1
		require.NoError(b, c.CreatePipeline(
			pipelineOne,
			"",
			[]string{"bash"},
			[]string{
				fmt.Sprintf("for f in /pfs/%s/*.tar.gz; do", dataRepo),
				"  echo \"going to untar $f\"",
				"  tar xzf \"$f\" -C /pfs/out",
				"done",
			},
			&pps.ParallelismSpec{
				Strategy: pps.ParallelismSpec_CONSTANT,
				Constant: 4,
			},
			[]*pps.PipelineInput{{
				Repo: client.NewRepo(dataRepo),
				Glob: "/*",
			}},
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit.ID)}, nil)
		require.NoError(b, err)
		collectCommitInfos(b, commitIter)
		b.StopTimer()
		outputRepoInfo, err := c.InspectRepo(pipelineOne)
		b.SetBytes(int64(outputRepoInfo.SizeBytes))
	}) {
		return
	}

	pipelineTwo := uniqueString("BenchmarkDataShuffleStageTwo")
	if !b.Run(fmt.Sprintf("Processing%dFiles", numTotalFiles), func(b *testing.B) {
		b.N = 1
		require.NoError(b, c.CreatePipeline(
			pipelineTwo,
			"",
			[]string{"bash"},
			// This script computes the checksum for each file (to exercise
			// CPU) and groups files into directories named after the first
			// letter of the filename.
			[]string{
				"mkdir -p /pfs/out/{a..z}",
				fmt.Sprintf("for i in /pfs/%s/*; do", pipelineOne),
				"    cksum $i",
				"    FILE=$(basename \"$i\")",
				"    LTR=$(echo \"${FILE:0:1}\")",
				"    mv \"$i\" \"/pfs/out/$LTR/$FILE\"",
				"done",
			},
			&pps.ParallelismSpec{
				Strategy: pps.ParallelismSpec_CONSTANT,
				Constant: 4,
			},
			[]*pps.PipelineInput{{
				Repo: client.NewRepo(pipelineOne),
				Glob: "/*",
			}},
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit.ID)}, nil)
		require.NoError(b, err)
		collectCommitInfos(b, commitIter)
		b.StopTimer()
		outputRepoInfo, err := c.InspectRepo(pipelineTwo)
		b.SetBytes(int64(outputRepoInfo.SizeBytes))
	}) {
		return
	}

	pipelineThree := uniqueString("BenchmarkDataShuffleStageThree")
	if !b.Run(fmt.Sprintf("Compressing%dFilesInto%dTarballs", numTotalFiles, numTarballs), func(b *testing.B) {
		b.N = 1
		require.NoError(b, c.CreatePipeline(
			pipelineThree,
			"",
			[]string{"bash"},
			// This script computes the checksum for each file (to exercise
			// CPU) and groups files into directories named after the first
			// letter of the filename.
			[]string{
				fmt.Sprintf("for i in /pfs/%s/*; do", pipelineTwo),
				"    DIR=$(basename \"$i\")",
				"    tar czf /pfs/out/$DIR.tar.gz $i",
				"done",
			},
			&pps.ParallelismSpec{
				Strategy: pps.ParallelismSpec_CONSTANT,
				Constant: 4,
			},
			[]*pps.PipelineInput{{
				Repo: client.NewRepo(pipelineTwo),
				Glob: "/*",
			}},
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit.ID)}, nil)
		require.NoError(b, err)
		collectCommitInfos(b, commitIter)
		b.StopTimer()
		outputRepoInfo, err := c.InspectRepo(pipelineTwo)
		b.SetBytes(int64(outputRepoInfo.SizeBytes))
	}) {
		return
	}

	for i := 0; i < round; i++ {
		if !b.Run(fmt.Sprintf("RunPipelinesEndToEnd-Round%d", i), func(b *testing.B) {
			commit := addInputCommit(b)
			commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit.ID)}, nil)
			require.NoError(b, err)
			collectCommitInfos(b, commitIter)
		}) {
			return
		}
	}
}

func BenchmarkLocalMultipleInputs(b *testing.B) {

}
