package mapreduce

import (
	"fmt"
	"log"
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
	for i := 0; i < 1000; i++ {
		go func(i int) {
			inFile, err := bucket.GetReader(fmt.Sprintf("chess/file%09d", i))
			if err != nil {
				log.Print(err)
				return
			}
			inFile.Close()
		}(i)
	}
}
