// main implements a load test that effectively builds a reverse index: records
// are randomly generated and written to a pachyderm input repo (in batches of
// size --records-per-file) such that every record is associated with one of N
// unique keys (N = --total-unique-keys). A pipeline reads each record out of
// its input file and writes it to an output file corresponding to its key.
//
// This primarily exercises Pachyderm because many input files write to any
// given output file, so merging is O(--unique-keys-per-file * --input-files),
// making the merge essentially quadratic and the load test's performance
// highly sensitive to the performance of Pachyderm's merge algorithm.

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	keySz       = 10  // the size of each row's key, in bytes
	separator   = '|' // the key/value separator
	separatorSz = 1   // the size of the key/value separator ('|'), for readability
	minValueSz  = 24  // see NewInputFile for an explanation
)

var (
	// flags:
	recordSz            int64  // size of each record
	recordsPerFile      int64  // records per file
	filesPerCommit      int    // files per commit
	numCommits          int    // number of commits (and therefore jobs) to create
	uniqueKeysPerFile   int64  // unique keys/file (ie size of each datum output)
	totalUniqueKeys     int64  // total number of output files
	pipelineConcurrency uint64 // parallelism of split pipeline
	hashtreeShards      uint64 // number of output hashtree shards for the loadtest pipeline
	putFileConcurrency  int64  // number of allowed concurrent 'put file's

	// commitTimes[i] is the amount of time that it took to start and finish
	// commit number 'i' (read by main() and PrintDurations())
	commitTimes []time.Duration

	// jobTimes[i] is the amount of time that it took to start and finish job
	// number 'i' (read by main() and PrintDurations())
	jobTimes []time.Duration
)

func init() {
	flag.Int64Var(&recordSz, "record-size", 100, "size of each record "+
		"generated and written to an input file")
	flag.Int64Var(&recordsPerFile, "records-per-file", 100, "number of records "+
		"written to each input file (total size of the file is "+
		"--record-size * --records-per-file")
	flag.IntVar(&filesPerCommit, "files-per-commit", 100, "total number of "+
		"input files that the load test writes in each commit (processed in each "+
		"job). The total size of an input commit is "+
		"--record-size * --records-per-file * --files-per-commit")
	flag.IntVar(&numCommits, "num-commits", 10, "total number of commits that "+
		"the load test creates (each containing --files-per-commit files and "+
		"spawning one job).")
	flag.Int64Var(&uniqueKeysPerFile, "unique-keys-per-file", 10, "number of "+
		"unique keys per file. This determines the difficulty of the load test: "+
		"higher --unique-keys-per-file => more metadata => bigger merge")
	flag.Int64Var(&totalUniqueKeys, "total-unique-keys", 1000, "number of "+
		"total unique keys. This determines the shape of the output. High "+
		"--total-unique-keys = many small output files. Low --total-unique-keys = "+
		"few large output files.")
	flag.Uint64Var(&pipelineConcurrency, "pipeline-concurrency", 5, "the "+
		"parallelism of the split pipeline")
	flag.Uint64Var(&hashtreeShards, "hashtree-shards", 3, "the "+
		"number of output hashtree shards for the split pipeline")
	flag.Int64Var(&putFileConcurrency, "put-file-concurrency", 3, "the number "+
		"of concurrent 'put file' RPCs that the load test will make while loading "+
		"input data")
}

// PrintFlags just prints the flag values, set above, to stdout. Useful for
// comparing benchmark runs
// TODO(msteffen): could this be eliminated with some kind of reflection?
func PrintFlags() {
	fmt.Printf("record-size: %v\n", recordSz)
	fmt.Printf("records-per-file: %v\n", recordsPerFile)
	fmt.Printf("files-per-commit: %v\n", filesPerCommit)
	fmt.Printf("num-commits: %v\n", numCommits)
	fmt.Printf("unique-keys-per-file: %v\n", uniqueKeysPerFile)
	fmt.Printf("total-unique-keys: %v\n", totalUniqueKeys)
	fmt.Printf("pipeline-concurrency: %v\n", pipelineConcurrency)
	fmt.Printf("hashtree-shards: %v\n", hashtreeShards)
	fmt.Printf("put-file-concurrency: %v\n", putFileConcurrency)
}

// PrintDurations prints the duration of all commits and jobs finished so far
func PrintDurations() {
	fmt.Print(" Job  Commit Time    Job Time\n")
	for i := 0; i < numCommits; i++ {
		fmt.Printf(" %3d: ", i)
		if i < len(commitTimes) {
			fmt.Printf("%11.3f", commitTimes[i].Seconds())
		} else {
			fmt.Print("        ---")
		}
		fmt.Print(" ")
		if i < len(jobTimes) {
			fmt.Printf("%11.3f", jobTimes[i].Seconds())
		} else {
			fmt.Print("        ---")
		}
		fmt.Print("\n")
	}
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	// validate flags
	if uniqueKeysPerFile > recordsPerFile {
		log.Fatalf("--unique-keys-per-file(%d) > --records-per-file(%d), but "+
			"files cannot have more unique keys than total records (each record "+
			"has one key", uniqueKeysPerFile, recordsPerFile)
	}
	if uniqueKeysPerFile > totalUniqueKeys {
		log.Fatalf("--unique-keys-per-file(%d) > --total-unique-keys(%d), but "+
			"there cannot be more unique keys within a file than there are total",
			uniqueKeysPerFile, totalUniqueKeys)
	}
	if recordSz < (keySz + separatorSz + minValueSz) {
		log.Fatalf("records must be at least %d bytes, as they start with a "+
			"%d-byte key and a separator, and values must be at least %d bytes",
			keySz+separatorSz+minValueSz, keySz, minValueSz)
	}
	PrintFlags()

	// Connect to pachyderm cluster
	log.Printf("starting to initialize pachyderm client")
	log.Printf("pachd address: \"%s:%s\"", os.Getenv("PACHD_SERVICE_HOST"),
		os.Getenv("PACHD_SERVICE_PORT"))
	c, err := client.NewInCluster()
	if err != nil {
		log.Fatalf("could not initialize Pachyderm client: %v", err)
	}

	// Make sure cluster is empty
	if ris, err := c.ListRepo(); err != nil || len(ris) > 0 {
		log.Fatalf("cluster must be empty before running the \"split\" loadtest")
	}

	// Create input repo
	log.Printf("creating input repo and pipeline")
	repo, branch := "input", "master"
	if err := c.CreateRepo(repo); err != nil {
		log.Fatalf("could not create input repo: %v", err)
	}

	// Create pipeline (using BENCH_VERSION env var, which ties the supervisor
	// container version to the pipeline container version)
	image := "pachyderm/split-loadtest-pipeline"
	if tag, ok := os.LookupEnv("BENCH_VERSION"); ok && tag != "" {
		image = "pachyderm/split-loadtest-pipeline:" + tag
	}
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{Name: "split"},
			Transform: &pps.Transform{
				Image: image,
				Cmd:   []string{"/pipeline", fmt.Sprintf("--key-size=%d", keySz)},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: pipelineConcurrency},
			HashtreeSpec:    &pps.HashtreeSpec{Constant: hashtreeShards},
			ResourceRequests: &pps.ResourceSpec{
				Memory: "1G",
				Cpu:    1,
			},
			Input: &pps.Input{
				Pfs: &pps.PFSInput{
					Repo:   repo,
					Branch: branch,
					Glob:   "/*",
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("could not create load test pipeline: %v", err)
	}

	// start timing load test
	var start = time.Now()
	defer func() {
		fmt.Printf("Benchmark complete. Total time: %.3f\n", time.Now().Sub(start).Seconds())
	}()

	// Start creating input files
	for i := 0; i < numCommits; i++ {
		commitStart := time.Now()
		// Start commit
		commit, err := c.StartCommit(repo, branch)
		if err != nil {
			log.Fatalf("could not start commit: %v", err)
		}
		log.Printf("starting commit %d (%s)", i, commit.ID) // log every 10%

		// Generate input files (a few at a time) and write them to pachyderm
		var eg errgroup.Group
		var sem = semaphore.NewWeighted(putFileConcurrency)
		for j := 0; j < filesPerCommit; j++ {
			i, j := i, j
			eg.Go(func() error {
				// if any 'put file' fails, the load test panics, so don't need a context
				sem.Acquire(context.Background(), 1)
				defer sem.Release(1)
				// log progress every 10% of the way through ingressing data
				if filesPerCommit < 10 || j%(filesPerCommit/10) == 0 {
					log.Printf("starting 'put file' (input-%d), (number %d in commit %d)", i*filesPerCommit+j, j, i)
				}
				fileNo := i*filesPerCommit + j
				name := fmt.Sprintf("input-%010x", fileNo)
				_, err := c.PutFile(repo, commit.ID, name, NewInputFile(fileNo))
				return err
			})
		}
		if err := eg.Wait(); err != nil {
			log.Fatalf("could not put file: %v", err)
		}
		if err := c.FinishCommit(repo, commit.ID); err != nil {
			log.Fatalf("could not finish commit: %v", err)
		}
		jobStart := time.Now()
		commitTimes = append(commitTimes, jobStart.Sub(commitStart))
		log.Printf("commit %d (%s) finished", i, commit.ID)
		PrintDurations()

		log.Printf("watching job %d (commit %s)", i, commit.ID)
		iter, err := c.FlushCommit([]*pfs.Commit{commit}, []*pfs.Repo{client.NewRepo("split")})
		if err != nil {
			log.Fatalf("could not flush commit %d: %v", i, err)
		}
		if _, err = iter.Next(); err != nil {
			log.Fatalf("could not get commit info after flushing commit %d: %v", i, err)
		}
		jobTimes = append(jobTimes, time.Now().Sub(jobStart))
		log.Printf("job %d (commit %s) finished", i, commit.ID)
		PrintDurations()
	}

	// Validate output
	var totalOutputSize, numFiles int64
	var expectedTotalOutputSize = (int64(filesPerCommit) * int64(numCommits) * recordsPerFile * recordSz)
	var expectedOutputFileSize = expectedTotalOutputSize / totalUniqueKeys
	fis, err := c.ListFile("split", "master", "/")
	if err != nil {
		log.Printf("[ Warning ] could not list output files to verify output: %v", err)
	} else {
		// Validate total # of output files
		if int64(len(fis)) != totalUniqueKeys {
			log.Printf("[ Warning ] expected %d output files, but saw %d", totalUniqueKeys, len(fis))
		} else {
			log.Printf("[  Pass   ] pipeline produced %d output files, as expected", totalUniqueKeys)
		}

		// Validate approximate size of each output file
		var warningsIssued bool
		for _, fi := range fis {
			fileSz := int64(fi.SizeBytes)
			totalOutputSize += fileSz
			numFiles++
			if fileSz%recordSz != 0 {
				log.Printf("[ Warning ] output file %s appears to have fragmented records", fi.File.Path)
				warningsIssued = true
			}
			if fileSz < (expectedOutputFileSize / 2) {
				log.Printf("[ Warning ] output file %s seems too small", fi.File.Path)
				warningsIssued = true
			}
			if fileSz > (expectedOutputFileSize * 2) {
				log.Printf("[ Warning ] output file %s seems too large", fi.File.Path)
				warningsIssued = true
			}
		}
		if !warningsIssued {
			log.Printf("[  Pass   ] All output file sizes appear to be correct")
		}

		// Validate total size of all output
		if totalOutputSize != expectedTotalOutputSize {
			log.Printf("[ Warning ] output data is not the same size as input data (%d input vs %d output)",
				expectedTotalOutputSize, totalOutputSize)
		} else {
			log.Printf("[  Pass   ] Total output size appears to be correct")
		}
	}
}

// InputFile is a synthetic file that generates test data for reading into
// Pachyderm
type InputFile struct {
	written  int64
	keyStart int64
	value    string // all keys in a given input file have the same value
}

// NewInputFile constructs a new InputFile reader
func NewInputFile(fileNo int) *InputFile {
	// 'value' is a confusing expression, but the goal is simply to pretty-print
	// the current file and line number as each line's value, so that the merge
	// results are easy to verify visually. On margin size:
	//   - File and line number take up 20 bytes
	//   - the '[', ':', ']', and '\n' characters take up four bytes.
	//   - therefore minValueSz is 24 bytes
	//   - This leaves (valueSz-minValueSz) bytes to be taken up by space
	//   - In case (valueSz-minValueSz) is odd, we make the right margin size
	//     round up so that (leftMargin + rightMargin) == (valueSz - minValueSz)
	//     holds.
	//   - Leave one formatting directive in the string as a literal, so that it
	//     can be replaced with the line number
	valueSz := recordSz - keySz - separatorSz
	leftMargin, rightMargin := (valueSz-minValueSz)/2, (valueSz-minValueSz+1)/2
	value := fmt.Sprintf("[%*s%010d:%%010d%*s]\n", leftMargin, "", fileNo, rightMargin, "")
	result := &InputFile{
		keyStart: ((int64(fileNo) % totalUniqueKeys) * uniqueKeysPerFile) % totalUniqueKeys,
		value:    value,
	}
	return result
}

// Read implements the io.Reader interface for InputFile
func (t *InputFile) Read(b []byte) (int, error) {
	fileSz := recordSz * recordsPerFile
	// sanity check state of 't'
	if t.written > fileSz {
		log.Fatalf("testFile exceeded file size (wrote %d bytes of %d)", t.written, fileSz)
	}
	if t.written == fileSz {
		return 0, io.EOF
	}

	initial := t.written
	var dn int
	for len(b) > 0 && t.written < fileSz {
		// figure out line & column based on # of bytes written
		line, c := t.written/recordSz, t.written%recordSz
		switch {
		case c < keySz:
			idx := (line % uniqueKeysPerFile)
			keyNo := (t.keyStart + idx) % totalUniqueKeys
			key := fmt.Sprintf("%0*d", keySz, keyNo)
			dn = copy(b, key[c:])
		case c == keySz:
			b[0] = separator
			dn = 1
		default:
			value := fmt.Sprintf(t.value, line) // replace formatting directive w/ line
			dn = copy(b, value[c-keySz-separatorSz:])
		}
		b = b[dn:]
		t.written += int64(dn)
	}
	return int(t.written - initial), nil
}
