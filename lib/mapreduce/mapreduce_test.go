package mapreduce

import (
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
	inFile, err := bucket.GetReader("chess/file000000000")
	if err != nil {
		log.Print(err)
		return
	}
	inFile.Close()
}
