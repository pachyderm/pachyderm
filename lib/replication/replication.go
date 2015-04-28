package replication

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"strings"

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
	Pull(from string, target CommitBrancher) (string, error)
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

func (r *S3Replica) Commit(diff io.Reader) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}
	key := fmt.Sprintf("%.10dC", r.count)
	r.count++
	return s3utils.PutMulti(bucket, key, diff, "application/octet-stream", s3.BucketOwnerFull)
}

func (r *S3Replica) Branch(base, name string) error {
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
	key := fmt.Sprintf("%.10dB", r.count)
	r.count++
	return bucket.Put(path.Join(r.uri, key), data, "application/octet-stream", s3.BucketOwnerFull)
}

func (r *S3Replica) Pull(from string, target CommitBrancher) (string, error) {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		log.Print(err)
		return from, err
	}
	return s3utils.ForEachFile(r.uri, from, func(path string) error {
		f, err := bucket.GetReader(path)
		if f == nil {
			return fmt.Errorf("Nil file returned.")
		}
		if err != nil {
			log.Print(err)
			return err
		}
		defer f.Close()

		if strings.HasSuffix(path, "B") {
			decoder := json.NewDecoder(f)
			b := branchRecord{}
			if err = decoder.Decode(&b); err != nil {
				log.Print(err)
				return err
			}
			err := target.Branch(b.base, b.name)
			if err != nil {
				log.Print(err)
				return err
			}
		} else if strings.HasSuffix(path, "C") {
			err := target.Commit(f)
			if err != nil {
				log.Print(err)
				return err
			}
		} else {
			return fmt.Errorf("Unrecognized diff %s. (Last character should be \"B\" or \"C\".)", path)
		}
		return nil
	})
}
