package s3utils

import (
	"log"
	"path"
	"strings"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

// An s3 input looks like: s3://bucket/dir
// Where dir can be a path

// getBucket extracts the bucket from an s3 input
func GetBucket(input string) (string, error) {
	return strings.Split(strings.TrimPrefix(input, "s3://"), "/")[0], nil
}

// getPath extracts the path from an s3 input
func GetPath(input string) (string, error) {
	return path.Join(strings.Split(strings.TrimPrefix(input, "s3://"), "/")[1:]...), nil
}

func NewBucket(bucketName string) (*s3.Bucket, error) {
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	client := s3.New(auth, aws.USWest)

	return client.Bucket(bucketName), nil
}
