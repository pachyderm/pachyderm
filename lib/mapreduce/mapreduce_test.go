package mapreduce

import (
	"testing"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/pachyderm/pfs/lib/btrfs"
)

func TestS3(t *testing.T) {
	auth, err := aws.EnvAuth()
	log.Printf("auth: %#v", auth)
	if err != nil {
		log.Print(err)
		return
	}
	client := s3.New(auth, aws.USWest)

	bucketName, err := getBucket(job.Input)
	if err != nil {
		log.Print(err)
		return
	}
	bucket := client.Bucket("pachyderm-data")
	inFile, err = bucket.GetReader("chess/file000000000")
	if err != nil {
		log.Print(err)
		return
	}
	defer inFile.Close()
}
