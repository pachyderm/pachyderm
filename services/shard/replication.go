package main

import (
	"io"
	"log"

	"github.com/pachyderm/pfs/lib/btrfs"
	"github.com/pachyderm/pfs/lib/s3utils"
)

type S3Backup struct {
	bucket string
}

func (b S3Backup) Push(repo, from string) error {
	return btrfs.Pull(repo, from, func(commit string, r io.ReadCloser) error {
		_, err := s3utils.NewBucket(b.bucket)
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	})
}
