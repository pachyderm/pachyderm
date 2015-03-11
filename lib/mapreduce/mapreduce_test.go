package mapreduce

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

func TestS3(t *testing.T) {
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
	for i := 0; i < 5000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			inFile, err := bucket.GetReader(fmt.Sprintf("chess/file%09d", i))
			if err != nil {
				log.Print(err)
				return
			}
			io.Copy(os.Stdout, inFile)
			inFile.Close()
		}(i)
	}
}
