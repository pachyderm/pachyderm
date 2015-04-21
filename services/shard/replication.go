package main

import (
	"io"
	"log"
	"path"

	"github.com/mitchellh/goamz/s3"
	"github.com/pachyderm/pfs/lib/btrfs"
	"github.com/pachyderm/pfs/lib/s3utils"
)

type S3Backup struct {
	bucket string
}

func (b S3Backup) Push(repo, from string) error {
	return btrfs.Pull(repo, from, func(commit string, r io.ReadCloser) error {
		bucket, err := s3utils.NewBucket(b.bucket)
		if err != nil {
			log.Print(err)
			return err
		}
		key := path.Join(repo, commit)
		// First check that we haven't already sent the chunk
		resp, err := bucket.Head(key)
		if err != nil {
			log.Print(err)
			return err
		}
		if resp.StatusCode == 200 {
			// The file exists, so we abort, because by assumption this means
			// the backup already has all previous diffs
			return fmt.Errorf("Diff exists.")
		}
		return s3utils.PutMulti(bucket, key, r, "application/octet-stream", s3.BucketOwnerFull)
	})
}
