package main

import (
	"io"
	"log"

	"code.google.com/p/go-uuid/uuid"
	"github.com/mitchellh/goamz/s3"
	"github.com/pachyderm/pfs/lib/s3utils"
)

type S3Replica struct {
	bucket string
}

func (r S3Replica) Commit(diff io.Reader) error {
	bucket, err := s3utils.NewBucket(r.bucket)
	if err != nil {
		log.Print(err)
		return err
	}
	key := uuid.New()
	// First check that we haven't already sent the chunk
	resp, err := bucket.Head(key)
	if err != nil {
		log.Print(err)
		return err
	}
	if resp.StatusCode == 200 {
		// The file exists, so we don't need to send it
		return nil
	}
	return s3utils.PutMulti(bucket, key, diff, "application/octet-stream", s3.BucketOwnerFull)
}

func (r S3Replica) Branch(base, name string) error {
	panic("Not implemented.")
}
