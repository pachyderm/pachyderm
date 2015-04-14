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
// Test: different input directory structures with Reduce (when do you use multiple reduce calls?)
// Test: Degenerate topologies (jobs that depend on each other (deadlock expected, use a timeout), jobs that are orphaned)
// Test: Degenerate jobs themselves (e.g. nonterminating, bad exit status)
// Test: WaitJob test that it returns when a Materialize call finishes.
// Test: WaitJob test that it never returns when a pipeline never finishes.

// TODO(rw): reify errors into their own types
