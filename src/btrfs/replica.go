package btrfs

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path"
	"time"

	"github.com/mitchellh/goamz/s3"
	"github.com/pachyderm/pachyderm/src/s3utils"
	"github.com/pachyderm/pachyderm/src/shell"
)

// Pusher is an interface that wraps the Push method.

// Push applies diff to an underlying storage layer.
type Pusher interface {
	Push(diff io.Reader) error
}

// Puller is an interface that wraps the Pull method.
type Puller interface {
	// Pull produces binary diffs and passes them to p's Push method.
	Pull(from string, p Pusher) error
}

// Replica is the interface that groups the Puller and Pusher methods.
type Replica interface {
	Pusher
	Puller
}

// A LocalReplica implements the Replica interface using a btrfs repo.
type LocalReplica struct {
	repo string
}

func (r LocalReplica) Push(diff io.Reader) error {
	return recv(r.repo, diff)
}

func (r LocalReplica) Pull(from string, cb Pusher) error {
	return Pull(r.repo, from, cb)
}

func NewLocalReplica(repo string) *LocalReplica {
	return &LocalReplica{repo: repo}
}

// S3Replica implements the Replica interface using an s3 replica.
type S3Replica struct {
	uri   string
	count int // number of sent commits
}

func (r *S3Replica) Push(diff io.Reader) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%.10d", r.count)
	r.count++

	p, err := s3utils.GetPath(r.uri)
	if err != nil {
		return err
	}

	return s3utils.PutMulti(bucket, path.Join(p, key), diff, "application/octet-stream", s3.BucketOwnerFull)
}

func (r *S3Replica) Pull(from string, target Pusher) error {
	bucket, err := s3utils.NewBucket(r.uri)
	if err != nil {
		return err
	}
	err = s3utils.ForEachFile(r.uri, from, func(path string, modtime time.Time) (retErr error) {
		f, err := bucket.GetReader(path)
		if f == nil {
			return fmt.Errorf("Nil file returned.")
		}
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()

		err = target.Push(f)
		if err != nil {
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

// send produces a binary diff stream and passes it to cont
func send(repo, commit string, cont func(io.Reader) error) error {
	parent := GetMeta(path.Join(repo, commit), "parent")
	if parent == "" {
		return shell.CallCont(exec.Command("btrfs", "send", FilePath(path.Join(repo, commit))), cont)
	} else {
		return shell.CallCont(exec.Command("btrfs", "send", "-p",
			FilePath(path.Join(repo, parent)), FilePath(path.Join(repo, commit))), cont)
	}
}

// recv reads a binary stream from data and applies it to `repo`
func recv(repo string, data io.Reader) error {
	c := exec.Command("btrfs", "receive", FilePath(repo))
	stdin, err := c.StdinPipe()
	if err != nil {
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	err = c.Start()
	if err != nil {
		return err
	}
	_, err = io.Copy(stdin, data)
	if err != nil {
		return err
	}
	err = stdin.Close()
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	log.Print("Stderr:", buf)
	err = c.Wait()
	if err != nil {
		return err
	}
	createNewBranch(repo)
	return nil
}

// createNewBranch gets called after a new commit has been `Recv`ed it creates
// the branch that should be pointing to the newly made commit.
func createNewBranch(repo string) error {
	// TODO this is a kind of ugly way to get the most recent branch we should
	// make it better
	err := Commits(repo, "", Desc, func(name string) error {
		branch := GetMeta(path.Join(repo, name), "branch")
		err := subvolumeDeleteAll(path.Join(repo, branch))
		if err != nil {
			return err
		}
		err = Branch(repo, name, branch)
		if err != nil {
			return err
		}
		return ErrComplete
	})
	if err != nil && err != ErrComplete {
		return err
	}
	return nil
}
