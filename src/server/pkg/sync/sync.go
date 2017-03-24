// Package sync provides utility functions similar to `git pull/push` for PFS
package sync

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	pachclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"

	"golang.org/x/sync/errgroup"
)

// Puller as a struct for managing a Pull operation.
type Puller struct {
	// errCh contains an error from the pipe goros
	errCh chan error
	// pipes is a set containing all pipes that are currently blocking
	pipes map[string]bool
	// pipesMu is a mutex to synchronize access to pipes
	pipesMu sync.Mutex
}

// NewPuller creates a new Puller struct.
func NewPuller() *Puller {
	return &Puller{
		errCh: make(chan error, 1),
		pipes: make(map[string]bool),
	}
}

// Pull clones an entire repo at a certain commit.
// root is the local path you want to clone to.
// fileInfo is the file/dir we are puuling.
// pipes causes the function to create named pipes in place of files, thus
// lazily downloading the data as it's needed.
func (p *Puller) Pull(ctx context.Context, client *pachclient.APIClient, root string, fileInfo *pfs.FileInfo, pipes bool, concurrency int) (retErr error) {
	limiter := limit.New(concurrency)
	commit := fileInfo.File.Commit
	switch fileInfo.FileType {
	case pfs.FileType_FILE:
		path := filepath.Join(root, fileInfo.File.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0666); err != nil {
			return err
		}
		if pipes {
			if err := syscall.Mkfifo(path, 0666); err != nil {
				return err
			}
			p.pipesMu.Lock()
			pipesMu[path] = true
			p.pipes.Unlock()
			// This goro will block until the user's code opens the
			// fifo.  That means we need to "abandon" this goro so that
			// the function can return and the caller can execute the
			// user's code. Waiting for this goro to return would
			// produce a deadlock.
			go func() {
				if err := func() (retErr error) {
					limiter.Acquire()
					defer limiter.Release()
					f, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
					p.pipesMu.Lock()
					delete(pipesMu, path)
					p.pipes.Unlock()
					if err != nil {
						return err
					}
					defer func() {
						if err := f.Close(); err != nil && retErr == nil {
							retErr = err
						}
					}()
					err = client.GetFile(commit.Repo.Name, commit.ID, fileInfo.File.Path, 0, 0, f)
					if err != nil {
						return err
					}
				}(); err != nil {
					select {
					case p.errCh <- err:
					default:
					}
				}
			}()
		} else {
			limiter.Acquire()
			defer limiter.Release()
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			return client.GetFile(commit.Repo.Name, commit.ID, fileInfo.File.Path, 0, 0, f)
		}
	case pfs.FileType_DIR:
		if err := os.MkdirAll(filepath.Join(root, fileInfo.File.Path), 0666); err != nil {
			return err
		}

		fileInfos, err := client.ListFile(
			commit.Repo.Name,
			commit.ID,
			fileInfo.File.Path,
		)
		if err != nil {
			return err
		}

		var g errgroup.Group
		for _, fileInfo := range fileInfos {
			fileInfo := fileInfo
			g.Go(func() error {
				return p.Pull(ctx, client, root, fileInfo, pipes, concurrency)
			})
		}

		return g.Wait()
	}
	return nil
}

// CleanUp cleans up blocked syscalls for pipes that were never opened. It also
// returns any errors that might have been encountered while trying to read
// data for the pipes. CleanUp should be called after all code that might
// access pipes has completed running, it should not be called concurrently.
func (p *Puller) CleanUp() error {
	var result error
	select {
	case result = <-p.errCh:
	default:
	}
	p.pipesMu.Lock()
	defer p.pipesMu.Unlock()
	for path := range p.pipes {
		f, err := os.OpenFile(path, syscall.O_NONBLOCK+os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	p.pipes = make(map[string]bool)
	return result
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
	if err := pachClient.Walk(commit.Repo.Name, commit.ID, "", func(fileInfo *pfs.FileInfo) error {
		if fileInfo.FileType != pfs.FileType_FILE {
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
			pachClient.GetFile(commit.Repo.Name, commit.ID, fileInfo.File.Path, 0, 0, w)
			return nil
		})
		return nil
	}); err != nil {
		return err
	}
	return eg.Wait()
}
