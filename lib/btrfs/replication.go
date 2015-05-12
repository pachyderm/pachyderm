package btrfs

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
	Pull(from string, target CommitBrancher) error
}

type Replica interface {
	CommitBrancher
	Puller
}

// A LocalReplica implements the CommitBrancher interface and replicates the
// commits to a local repo. It expects `repo` to already exist
type LocalReplica struct {
	repo string
}

func (r LocalReplica) Commit(diff io.Reader) error {
	return Recv(r.repo, diff)
}

func (r LocalReplica) Branch(base, name string) error {
	// We remove the old version of the branch if it exists here
	if err := SubvolumeDeleteAll(path.Join(r.repo, name)); err != nil {
		return err
	}
	return Branch(r.repo, base, name)
}

func (r LocalReplica) Pull(from string, cb CommitBrancher) error {
	return Pull(r.repo, from, cb)
}

func NewLocalReplica(repo string) *LocalReplica {
	return &LocalReplica{repo: repo}
}

type S3Replica struct {
	uri   string
	count int // number of sent commits
}

type BranchRecord struct {
	Base string `json:"base"`
	Name string `json:"name"`
}

func (r *S3Replica) Commit(diff io.Reader) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}
	key := fmt.Sprintf("%.10dC", r.count)
	r.count++

	p, err := s3utils.GetPath(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}

	return s3utils.PutMulti(bucket, path.Join(p, key), diff, "application/octet-stream", s3.BucketOwnerFull)
}

func (r *S3Replica) Branch(base, name string) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}
	data, err := json.Marshal(BranchRecord{Base: base, Name: name})
	if err != nil {
		log.Print(err)
		return err
	}
	key := fmt.Sprintf("%.10dB", r.count)
	r.count++

	p, err := s3utils.GetPath(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}

	return bucket.Put(path.Join(p, key), data, "application/octet-stream", s3.BucketOwnerFull)
}

func (r *S3Replica) Pull(from string, target CommitBrancher) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}
	_, err = s3utils.ForEachFile(r.uri, from, func(path string) error {
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
			b := BranchRecord{}
			if err = decoder.Decode(&b); err != nil {
				log.Print(err)
				return err
			}
			err := target.Branch(b.Base, b.Name)
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
	if err != nil {
		return err
	}
	return nil
}

func NewS3Replica(uri string) *S3Replica {
	return &S3Replica{uri: uri}
}
