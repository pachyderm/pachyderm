package replication

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"

	"github.com/mitchellh/goamz/s3"
	"github.com/pachyderm/pfs/lib/s3utils"
)

type CommitBrancher interface {
	Commit(diff io.Reader) error
	Branch(base, name string) error
}

type Puller interface {
	// Pull pulls data from a replica and applies it to the target,
	// `from` is used to pickup where you left-off, passing `from=""` will start from the beginning
	// Pull returns the value that should be passed next as `from`
	Pull(from string, target *CommitBrancher) (string, error)
}

type Replica interface {
	CommitBrancher
	Puller
}

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
