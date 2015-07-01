package btrfs

import (
	"fmt"
	"io"
	"log"
	"path"
	"time"

	"github.com/mitchellh/goamz/s3"
	"github.com/pachyderm/pachyderm/src/s3utils"
)

type Pusher interface {
	Push(diff io.Reader) error
}

type Puller interface {
	// Pull pulls data from a replica and applies it to the target,
	// `from` is used to pickup where you left-off, passing `from=""` will start from the beginning
	// Pull returns the value that should be passed next as `from`
	Pull(from string, target Pusher) error
}

type Replica interface {
	Pusher
	Puller
}

// A LocalReplica implements the Replica interface and replicates the
// commits to a local repo. It expects `repo` to already exist
type LocalReplica struct {
	repo string
}

func (r LocalReplica) Push(diff io.Reader) error {
	return Recv(r.repo, diff)
}

func (r LocalReplica) Pull(from string, cb Pusher) error {
	return Pull(r.repo, from, cb)
}

func NewLocalReplica(repo string) *LocalReplica {
	return &LocalReplica{repo: repo}
}

type S3Replica struct {
	uri   string
	count int // number of sent commits
}

func (r *S3Replica) Push(diff io.Reader) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}
	key := fmt.Sprintf("%.10d", r.count)
	r.count++

	p, err := s3utils.GetPath(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}

	return s3utils.PutMulti(bucket, path.Join(p, key), diff, "application/octet-stream", s3.BucketOwnerFull)
}

func (r *S3Replica) Pull(from string, target Pusher) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		log.Print(err)
		return err
	}
	err = s3utils.ForEachFile(r.uri, from, func(path string, modtime time.Time) error {
		f, err := bucket.GetReader(path)
		if f == nil {
			return fmt.Errorf("Nil file returned.")
		}
		if err != nil {
			log.Print(err)
			return err
		}
		defer f.Close()

		err = target.Push(f)
		if err != nil {
			log.Print(err)
			return err
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
