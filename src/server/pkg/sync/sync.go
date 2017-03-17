// Package sync provides utility functions similar to `git pull/push` for PFS
package sync

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	pachclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"

	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"golang.org/x/sync/errgroup"
)

// RetryError is used to log retry attempts.
type RetryError struct {
	Err               string
	TimeTillNextRetry string
	Repo              string
	CommitID          string
	File              string
}

// Pull clones an entire repo at a certain commit
//
// root is the local path you want to clone to
// commit is the commit you want to clone
// shard and diffMethod get passed to ListFile and GetFile. See documentations
// for those functions for details on these arguments.
// pipes causes the function to create named pipes in place of files, thus
// lazily downloading the data as it's needed
func Pull(client *pachclient.APIClient, root string, commit *pfs.Commit, diffMethod *pfs.DiffMethod, shard *pfs.Shard, pipes bool) error {
	return pullDir(client, root, commit, diffMethod, shard, "/", pipes)
}

func pullDir(client *pachclient.APIClient, root string, commit *pfs.Commit, diffMethod *pfs.DiffMethod, shard *pfs.Shard, dir string, pipes bool) error {
	fromCommit := ""
	fullFile := false
	if diffMethod != nil {
		if diffMethod.FromCommit != nil {
			fromCommit = diffMethod.FromCommit.ID
		}
		fullFile = diffMethod.FullFile
	}
	fileInfos, err := client.ListFile(
		commit.Repo.Name,
		commit.ID,
		dir,
		fromCommit,
		fullFile,
		shard,
		false,
	)
	if err != nil {
		return err
	}
	backoffConfig := obj.NewExponentialBackOffConfig()
	var g errgroup.Group
	sem := make(chan struct{}, 100)
	for _, fileInfo := range fileInfos {
		fileInfo := fileInfo
		sem <- struct{}{}
		g.Go(func() (retErr error) {
			defer func() { <-sem }()
			switch fileInfo.FileType {
			case pfs.FileType_FILE_TYPE_REGULAR:
				path := filepath.Join(root, fileInfo.File.Path)
				if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
					return err
				}
				if pipes {
					if err := syscall.Mkfifo(path, 0666); err != nil {
						return err
					}
					// This goro will block until the user's code opens the
					// fifo.  That means we need to "abandon" this goro so that
					// the function can return and the caller can execute the
					// user's code. Waiting for this goro to return would
					// produce a deadlock.
					go func() {
						f, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
						if err != nil {
							log.Printf("error opening %s: %s", path, err)
							return
						}
						defer func() {
							if err := f.Close(); err != nil {
								log.Printf("error closing %s: %s", path, err)
							}
						}()
						err = client.GetFile(commit.Repo.Name, commit.ID, fileInfo.File.Path, 0, 0, fromCommit, fullFile, shard, f)
						if err != nil {
							log.Printf("error from GetFile: %s", err)
							return
						}
					}()
				} else {
					f, err := os.Create(path)
					if err != nil {
						return err
					}
					defer func() {
						if err := f.Close(); err != nil && retErr == nil {
							retErr = err
						}
					}()

					backoff.RetryNotify(func() error {
						err = client.GetFile(commit.Repo.Name, commit.ID, fileInfo.File.Path, 0, 0, fromCommit, fullFile, shard, f)
						if err != nil && isRetryable(err) {
							return err
						}
						return nil
					}, backoffConfig, func(err error, d time.Duration) {
						log.Infof("Error getting file; retrying in %s: %#v", d, RetryError{
							Err:               err.Error(),
							TimeTillNextRetry: d.String(),
							Repo:              commit.Repo.Name,
							CommitID:          commit.ID,
							File:              fileInfo.File.Path,
						})
					})

					return err
				}
			case pfs.FileType_FILE_TYPE_DIR:
				return pullDir(client, root, commit, diffMethod, shard, fileInfo.File.Path, pipes)
			}
			return nil
		})
	}
	return g.Wait()
}

func isRetryable(err error) bool {
	if strings.Contains(err.Error(), "connection reset by peer") {
		return true
	}
	netErr, ok := err.(net.Error)
	return ok && netErr.Temporary()
}

// Push puts files under root into an open commit.
func Push(client *pachclient.APIClient, root string, commit *pfs.Commit, overwrite bool) error {
	var g errgroup.Group
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		g.Go(func() (retErr error) {
			if path == root || info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()

			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			if overwrite {
				if err := client.DeleteFile(commit.Repo.Name, commit.ID, relPath); err != nil {
					return err
				}
			}

			_, err = client.PutFile(commit.Repo.Name, commit.ID, relPath, f)
			return err
		})
		return nil
	}); err != nil {
		return err
	}

	return g.Wait()
}

// PushObj pushes data from commit to an object store.
func PushObj(pachClient pachclient.APIClient, commit *pfs.Commit, objClient obj.Client, root string) error {
	var eg errgroup.Group
	if err := pachClient.Walk(commit.Repo.Name, commit.ID, "", "", false, nil, func(fileInfo *pfs.FileInfo) error {
		if fileInfo.FileType != pfs.FileType_FILE_TYPE_REGULAR {
			return nil
		}
		eg.Go(func() (retErr error) {
			w, err := objClient.Writer(filepath.Join(root, fileInfo.File.Path))
			if err != nil {
				return err
			}
			defer func() {
				if err := w.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			pachClient.GetFile(commit.Repo.Name, commit.ID, fileInfo.File.Path, 0, 0, "", false, nil, w)
			return nil
		})
		return nil
	}); err != nil {
		return err
	}
	return eg.Wait()
}
