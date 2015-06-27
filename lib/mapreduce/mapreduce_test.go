package mapreduce

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"testing"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"

	"github.com/pachyderm/pfs/lib/btrfs"
	"github.com/pachyderm/pfs/lib/utils"
)

// TestS3 checks that the underlying S3 library is correct.
func TestS3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping S3 integration test")
	}
	auth, err := aws.EnvAuth()
	log.Printf("auth: %#v", auth)
	if err != nil {
		log.Print(err)
		return
	}
	client := s3.New(auth, aws.USWest)
	bucket := client.Bucket("pachyderm-data")
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			inFile, err := bucket.GetReader(fmt.Sprintf("chess/file%09d", i))
			if err != nil {
				log.Print(err)
				return
			}
			io.Copy(ioutil.Discard, inFile)
			inFile.Close()
		}(i)
	}
}

// Public types: Job
// Public functions: Materialize, Map, Reduce, PrepJob, WaitJob
// Could be made private: Map, Reduce, WaitJob
// Test: make two calls to Materialize and see what happens when they overlap
// Test: spin up / spin down / start containers. check its ip address, send it a command, spin it down.
// Test: positive test of 1000 jobs (Map/Reduce combinations)

func TestMapJob(t *testing.T) {
	inRepoName := "repo_TestMapJob_input"
	outRepoName := "repo_TestMapJob_output"
	check(btrfs.Init(inRepoName), t)
	check(btrfs.Init(outRepoName), t)

	// Create a file to map over:
	f, err := btrfs.CreateAll(fmt.Sprintf("%s/master/my_first_job/foo", inRepoName))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("hello world\n")
	f.Close()

	// Commit it:
	check(btrfs.Commit(inRepoName, "commit1", "master"), t)

	// Set up the job:
	j := Job{
		Type:  "map",
		Input: "my_first_job",
		Image: "jdoliner/hello-world",
		//Cmd:   []string{"wc -l"},
		Cmd: []string{"/go/bin/hello-world-mr"},
	}
	matInfo := materializeInfo{inRepoName, outRepoName, "master", "commit1"}

	// the following shard and mod combination will always be true,
	// because for all n, (n % 1) == 0:
	shard := uint64(0)
	mod := uint64(1)
	Map(j, "TestMapJob", matInfo, shard, mod)
	check(btrfs.Commit(outRepoName, "commit1", "master"), t)

	// Check that the output file exists and contains the expected output:
	output, err := btrfs.ReadFile(fmt.Sprintf("%s/commit1/TestMapJob/foo", outRepoName))
	check(err, t)
	if string(output) != "Hello World." {
		t.Fatalf("%s should be Hello World.", string(output))
	}
}

// Test: different input directory structures with Reduce (when do you use multiple reduce calls?)
// Test: Degenerate topologies (jobs that depend on each other (deadlock expected, use a timeout), jobs that are orphaned)
// Test: Degenerate jobs themselves (e.g. nonterminating, bad exit status)
// Test: WaitJob test that it returns when a Materialize call finishes.
// Test: WaitJob test that it never returns when a pipeline never finishes.

// TODO(rw): reify errors into their own types
