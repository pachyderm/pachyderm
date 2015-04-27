package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"

	"github.com/mitchellh/goamz/s3"
	"github.com/pachyderm/pfs/lib/s3utils"
)

type S3Replica struct {
	uri   string
	count int // number of sent commits
}

type branchRecord struct {
	base string `json:"base"`
	name string `json:"name"`
}

func (r S3Replica) Commit(diff io.Reader) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}
	key := fmt.Sprintf("%.10d", r.count)
	r.count++
	return s3utils.PutMulti(bucket, key, diff, "application/octet-stream", s3.BucketOwnerFull)
}

func (r S3Replica) Branch(base, name string) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}
	data, err := json.Marshal(branchRecord{base: base, name: name})
	if err != nil {
		log.Print(err)
		return err
	}
	key := fmt.Sprintf("%.10d", r.count)
	r.count++
	return bucket.Put(path.Join(r.uri, key), data, "application/octet-stream", s3.BucketOwnerFull)
}
