package server

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"path"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	pfs "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	pps "github.com/pachyderm/pachyderm/src/client/pps"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
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
	seed := time.Now().Unix()
	fmt.Println("Producing rand.Rand with seed: ", seed)
	return rand.New(rand.NewSource(seed))
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

// benchmarkFiles runs a benchmarks that uploads, downloads, and processes
// fileNum files, whose sizes (in bytes) are produced by a zipf
// distribution with the given min and max.
func benchmarkFiles(b *testing.B, fileNum int, minSize uint64, maxSize uint64, local bool) {
	scalePachd(b)
	r := getRand()

	repo := tu.UniqueString("BenchmarkPachydermFiles")
	c := getPachClient(b)
	require.NoError(b, c.CreateRepo(repo))

	commit, err := c.StartCommit(repo, "master")
	require.NoError(b, err)
	if !b.Run(fmt.Sprintf("Put%dFiles", fileNum), func(b *testing.B) {
		var totalBytes int64
		var eg errgroup.Group
		zipf := rand.NewZipf(r, 1.05, 1, maxSize-minSize)
		for k := 0; k < fileNum; k++ {
			k := k
			fileSize := int64(zipf.Uint64() + minSize)
			totalBytes += fileSize
			eg.Go(func() error {
				var err error
				if k%10 == 0 {
					_, err = c.PutFile(repo, commit.ID, fmt.Sprintf("dir/file%d", k), workload.NewReader(r, fileSize))
				} else {
					_, err = c.PutFile(repo, commit.ID, fmt.Sprintf("file%d", k), workload.NewReader(r, fileSize))
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

	if !b.Run(fmt.Sprintf("Get%dFiles", fileNum), func(b *testing.B) {
		b.N = 1
		w := &countWriter{}
		var eg errgroup.Group
		for k := 0; k < fileNum; k++ {
			k := k
			eg.Go(func() error {
				if k%10 == 0 {
					return c.GetFile(repo, commit.ID, fmt.Sprintf("dir/file%d", k), 0, 0, w)
				}
				return c.GetFile(repo, commit.ID, fmt.Sprintf("file%d", k), 0, 0, w)
			})
		}
		require.NoError(b, eg.Wait())
		defer func() { b.SetBytes(w.count) }()
	}) {
		return
	}

	if !b.Run(fmt.Sprintf("PipelineCopy%dFiles", fileNum), func(b *testing.B) {
		pipeline := tu.UniqueString("BenchmarkFilesPipeline")
		require.NoError(b, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{fmt.Sprintf("cp -R %s /pfs/out/", path.Join("/pfs", repo, "/*"))},
			&pps.ParallelismSpec{
				Constant: 4,
			},
			client.NewPFSInput(repo, "/*"),
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

// TODO(msteffen): Run this only in S3
// func BenchmarkDailyPutLargeFileViaS3(b *testing.B) {
// 	repo := tu.UniqueString("BenchmarkDailyPutLargeFileViaS3")
// 	c, err := client.NewInCluster()
// 	require.NoError(b, err)
// 	require.NoError(b, c.CreateRepo(repo))
// 	for i := 0; i < b.N; i++ {
// 		commit, err := c.StartCommit(repo, "master")
// 		require.NoError(b, err)
// 		err = c.PutFileURL(repo, "master", "/", "s3://pachyderm-internal-benchmark/bigfiles/1gb.bytes", false)
// 		require.NoError(b, err)
// 		require.NoError(b, c.FinishCommit(repo, commit.ID))
// 		b.SetBytes(int64(1024 * MB))
// 	}
// }

func BenchmarkDailyDataShuffle(b *testing.B) {
	// The following workload consists of roughly 500GB of data
	//benchmarkDataShuffle(b, 20, 1000, 1*KB, 100*MB, 10)
	//benchmarkDataShuffle(b, 10, 10000, 100, 10*MB, 10)
	benchmarkDataShuffle(b, 1, 10000, 100, 10*MB, 9)
}

func BenchmarkLocalDataShuffle(b *testing.B) {
	benchmarkDataShuffle(b, 1, 100, 1*KB, 100*MB, 3)
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
	r := getRand()

	numTotalFiles := numTarballs * numFilesPerTarball
	dataRepo := tu.UniqueString("BenchmarkDataShuffle")
	c := getPachClient(b)
	require.NoError(b, c.CreateRepo(dataRepo))

	var lastIndex int
	// addInputCommit adds an input commit to the data repo
	addInputCommit := func(b *testing.B) *pfs.Commit {
		commit, err := c.StartCommit(dataRepo, "master")
		require.NoError(b, err)
		var totalSize int64
		// the group that generates data
		// the group that writes data to pachd
		var genEg errgroup.Group
		var writeEg errgroup.Group
		for i := lastIndex; i < lastIndex+numTarballs; i++ {
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
				c := getPachClient(b)
				_, err = c.PutFile(dataRepo, commit.ID, tarName, pr)
				fmt.Printf("Finished putting file %s\n", tarName)
				return err
			})
		}
		lastIndex += numTarballs
		require.NoError(b, genEg.Wait())
		require.NoError(b, writeEg.Wait())
		fmt.Println("Finished putting all files")
		require.NoError(b, c.FinishCommit(dataRepo, commit.ID))
		fmt.Println("Finished commit")
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

	pipelineOne := tu.UniqueString("BenchmarkDataShuffleStageOne")
	if !b.Run(fmt.Sprintf("Extract%dTarballsInto%dFiles", numTarballs, numTotalFiles), func(b *testing.B) {
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
				Constant: 4,
			},
			client.NewPFSInput(dataRepo, "/*"),
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit.ID)}, nil)
		require.NoError(b, err)
		collectCommitInfos(b, commitIter)
		b.StopTimer()
		outputRepoInfo, err := c.InspectRepo(pipelineOne)
		require.NoError(b, err)
		b.SetBytes(int64(outputRepoInfo.SizeBytes))
	}) {
		return
	}

	pipelineTwo := tu.UniqueString("BenchmarkDataShuffleStageTwo")
	if !b.Run(fmt.Sprintf("Processing%dFiles", numTotalFiles), func(b *testing.B) {
		require.NoError(b, c.CreatePipeline(
			pipelineTwo,
			"",
			[]string{"bash"},
			// This script computes the checksum for each file (to exercise
			// CPU) and groups files into directories named after the first
			// letter of the filename.
			[]string{
				fmt.Sprintf("for i in /pfs/%s/*; do", pipelineOne),
				"    cksum $i",
				"    FILE=$(basename \"$i\")",
				"    LTR=$(echo \"${FILE:0:1}\")",
				"	 mkdir -p /pfs/out/$LTR",
				"    mv \"$i\" \"/pfs/out/$LTR/$FILE\"",
				"done",
			},
			&pps.ParallelismSpec{
				Constant: 4,
			},
			client.NewPFSInput(pipelineOne, "/*"),
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit.ID)}, nil)
		require.NoError(b, err)
		collectCommitInfos(b, commitIter)
		b.StopTimer()
		outputRepoInfo, err := c.InspectRepo(pipelineTwo)
		require.NoError(b, err)
		b.SetBytes(int64(outputRepoInfo.SizeBytes))
	}) {
		return
	}

	pipelineThree := tu.UniqueString("BenchmarkDataShuffleStageThree")
	if !b.Run(fmt.Sprintf("Compressing26DirectoriesWith%dFilesInto26Tarballs", numTotalFiles), func(b *testing.B) {
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
				Constant: 4,
			},
			client.NewPFSInput(pipelineTwo, "/*"),
			"",
			false,
		))
		commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit.ID)}, nil)
		require.NoError(b, err)
		collectCommitInfos(b, commitIter)
		b.StopTimer()
		outputRepoInfo, err := c.InspectRepo(pipelineTwo)
		require.NoError(b, err)
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

var numFiles = 1000

// These benchmarks can take a while, so I recommend running like:
// $go test -v ./src/server/ -bench=BenchmarkPutManyFiles -run=BenchmarkPutManyFilesSingleCommit -timeout=3000s -benchtime 100ms

func BenchmarkPutManyFilesSingleCommitFinishCommit1(b *testing.B) {
	benchmarkPutManyFilesSingleCommitFinishCommit(numFiles, b)
}
func BenchmarkPutManyFilesSingleCommitFinishCommit10(b *testing.B) {
	benchmarkPutManyFilesSingleCommitFinishCommit(numFiles*10, b)
}
func BenchmarkPutManyFilesSingleCommitFinishCommit100(b *testing.B) {
	benchmarkPutManyFilesSingleCommitFinishCommit(numFiles*100, b)
}

func benchmarkPutManyFilesSingleCommitFinishCommit(numFiles int, b *testing.B) {
	// Issue #2575
	// At first we couldn't support this at all (correctness), but this has been fixed
	// This bench is to see how fast we can handle this

	c := getPachClient(b)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		require.NoError(b, c.DeleteAll())
		// create repos
		dataRepo := tu.UniqueString("TestManyFilesSingleCommit_data")
		require.NoError(b, c.CreateRepo(dataRepo))

		_, err := c.StartCommit(dataRepo, "master")
		require.NoError(b, err)
		for i := 0; i < numFiles; i++ {
			_, err = c.PutFile(dataRepo, "master", fmt.Sprintf("file-%d", i), strings.NewReader(""))
			require.NoError(b, err)
		}
		// We're actually benchmarking the finish commit
		b.StartTimer()
		require.NoError(b, c.FinishCommit(dataRepo, "master"))
	}
}

func BenchmarkPutManyFilesSingleCommit1(b *testing.B) {
	benchmarkPutManyFilesSingleCommit(numFiles, b)
}
func BenchmarkPutManyFilesSingleCommit10(b *testing.B) {
	benchmarkPutManyFilesSingleCommit(numFiles*10, b)
}
func BenchmarkPutManyFilesSingleCommit100(b *testing.B) {
	benchmarkPutManyFilesSingleCommit(numFiles*100, b)
}
func benchmarkPutManyFilesSingleCommit(numFiles int, b *testing.B) {
	// Issue #2575
	// At first we couldn't support this at all (correctness), but this has been fixed
	// This bench is to see how fast we can handle this

	c := getPachClient(b)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		require.NoError(b, c.DeleteAll())
		// create repos
		dataRepo := tu.UniqueString("TestManyFilesSingleCommit_data")
		require.NoError(b, c.CreateRepo(dataRepo))

		_, err := c.StartCommit(dataRepo, "master")
		require.NoError(b, err)
		b.StartTimer()
		for i := 0; i < numFiles; i++ {
			_, err = c.PutFile(dataRepo, "master", fmt.Sprintf("file-%d", i), strings.NewReader(""))
			require.NoError(b, err)
		}
	}
}
