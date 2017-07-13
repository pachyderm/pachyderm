// Package sync provides utility functions similar to `git pull/push` for PFS
package sync

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"

	pachclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"

	"golang.org/x/sync/errgroup"
)

// Puller as a struct for managing a Pull operation.
type Puller struct {
	sync.Mutex
	// errCh contains an error from the pipe goros
	errCh chan error
	// pipes is a set containing all pipes that are currently blocking
	pipes map[string]bool
	// cleaned signals if the cleanup goroutine has been started
	cleaned bool
	// wg is used to wait for all goroutines associated with this Puller
	// to complete.
	wg sync.WaitGroup
	// size is the total amount this puller has pulled
	size int64
}

// NewPuller creates a new Puller struct.
func NewPuller() *Puller {
	return &Puller{
		errCh: make(chan error, 1),
		pipes: make(map[string]bool),
	}
}

type sizeWriter struct {
	w    io.Writer
	size int64
}

func (s *sizeWriter) Write(p []byte) (int, error) {
	n, err := s.w.Write(p)
	s.size += int64(n)
	return n, err
}

func (p *Puller) makePipe(path string, f func(io.Writer) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if err := syscall.Mkfifo(path, 0666); err != nil {
		return err
	}
	func() {
		p.Lock()
		defer p.Unlock()
		p.pipes[path] = true
	}()
	// This goro will block until the user's code opens the
	// fifo.  That means we need to "abandon" this goro so that
	// the function can return and the caller can execute the
	// user's code. Waiting for this goro to return would
	// produce a deadlock. This goro will exit (if it hasn't already)
	// when CleanUp is called.
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if err := func() (retErr error) {
			file, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
			if err != nil {
				return err
			}
			defer func() {
				if err := file.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			// If the CleanUp routine has already run, then there's
			// no point in downloading and sending the file, so we
			// exit early.
			if func() bool {
				p.Lock()
				defer p.Unlock()
				delete(p.pipes, path)
				return p.cleaned
			}() {
				return nil
			}
			w := &sizeWriter{w: file}
			if err := f(w); err != nil {
				return err
			}
			atomic.AddInt64(&p.size, w.size)
			return nil
		}(); err != nil {
			select {
			case p.errCh <- err:
			default:
			}
		}
	}()
	return nil
}

func (p *Puller) makeFile(path string, f func(io.Writer) error) (retErr error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	w := &sizeWriter{w: file}
	if err := f(w); err != nil {
		return err
	}
	atomic.AddInt64(&p.size, w.size)
	return nil
}

// Pull clones an entire repo at a certain commit.
// root is the local path you want to clone to.
// fileInfo is the file/dir we are puuling.
// pipes causes the function to create named pipes in place of files, thus
// lazily downloading the data as it's needed.
// tree is a hashtree to mirror the pulled content into (it may be left nil)
// treeRoot is the root the data is mirrored to within tree
func (p *Puller) Pull(client *pachclient.APIClient, root string, repo, commit, file string,
	pipes bool, concurrency int, tree hashtree.OpenHashTree, treeRoot string) error {
	limiter := limit.New(concurrency)
	var eg errgroup.Group
	if err := client.Walk(repo, commit, file, func(fileInfo *pfs.FileInfo) error {
		basepath, err := filepath.Rel(file, fileInfo.File.Path)
		if err != nil {
			return err
		}
		if tree != nil {
			treePath := path.Join(treeRoot, basepath)
			if err := tree.PutFile(treePath, fileInfo.Objects, int64(fileInfo.SizeBytes)); err != nil {
				return err
			}
		}
		path := filepath.Join(root, basepath)
		if fileInfo.FileType == pfs.FileType_DIR {
			return os.MkdirAll(path, 0700)
		}
		if pipes {
			return p.makePipe(path, func(w io.Writer) error {
				return client.GetFile(repo, commit, fileInfo.File.Path, 0, 0, w)
			})
		}
		eg.Go(func() (retErr error) {
			limiter.Acquire()
			defer limiter.Release()
			return p.makeFile(path, func(w io.Writer) error {
				return client.GetFile(repo, commit, fileInfo.File.Path, 0, 0, w)
			})
		})
		return nil
	}); err != nil {
		return err
	}
	return eg.Wait()
}

// PullDiff is like Pull except that it materializes a Diff of the content
// rather than a the actual content. If newOnly is true then only new files
// will be downloaded and they will be downloaded under root. Otherwise new and
// old files will be downloaded under root/new and root/old respectively.
func (p *Puller) PullDiff(client *pachclient.APIClient, root string, newRepo, newCommit, newPath, oldRepo, oldCommit, oldPath string,
	newOnly bool, pipes bool, concurrency int, tree hashtree.OpenHashTree, treeRoot string) error {
	limiter := limit.New(concurrency)
	var eg errgroup.Group
	newFiles, oldFiles, err := client.DiffFile(newRepo, newCommit, newPath, oldRepo, oldCommit, oldPath)
	if err != nil {
		return err
	}
	for _, newFile := range newFiles {
		basepath, err := filepath.Rel(newPath, newFile.File.Path)
		if err != nil {
			return err
		}
		if tree != nil {
			treePath := path.Join(treeRoot, "new", basepath)
			if newOnly {
				treePath = path.Join(treeRoot, basepath)
			}
			if err := tree.PutFile(treePath, newFile.Objects, int64(newFile.SizeBytes)); err != nil {
				return err
			}
		}
		path := filepath.Join(root, "new", basepath)
		if newOnly {
			path = filepath.Join(root, basepath)
		}
		if pipes {
			if err := p.makePipe(path, func(w io.Writer) error {
				return client.GetFile(newFile.File.Commit.Repo.Name, newFile.File.Commit.ID, newFile.File.Path, 0, 0, w)
			}); err != nil {
				return err
			}
		} else {
			newFile := newFile
			limiter.Acquire()
			eg.Go(func() error {
				defer limiter.Release()
				return p.makeFile(path, func(w io.Writer) error {
					return client.GetFile(newFile.File.Commit.Repo.Name, newFile.File.Commit.ID, newFile.File.Path, 0, 0, w)
				})
			})
		}
	}
	if !newOnly {
		for _, oldFile := range oldFiles {
			basepath, err := filepath.Rel(oldPath, oldFile.File.Path)
			if err != nil {
				return err
			}
			if tree != nil {
				treePath := path.Join(treeRoot, "old", basepath)
				if err := tree.PutFile(treePath, oldFile.Objects, int64(oldFile.SizeBytes)); err != nil {
					return err
				}
			}
			path := filepath.Join(root, "old", basepath)
			if pipes {
				if err := p.makePipe(path, func(w io.Writer) error {
					return client.GetFile(oldFile.File.Commit.Repo.Name, oldFile.File.Commit.ID, oldFile.File.Path, 0, 0, w)
				}); err != nil {
					return err
				}
			} else {
				oldFile := oldFile
				limiter.Acquire()
				eg.Go(func() error {
					defer limiter.Release()
					return p.makeFile(path, func(w io.Writer) error {
						return client.GetFile(oldFile.File.Commit.Repo.Name, oldFile.File.Commit.ID, oldFile.File.Path, 0, 0, w)
					})
				})
			}
		}
	}
	return eg.Wait()
}

// PullTree pulls from a raw HashTree rather than a repo.
func (p *Puller) PullTree(client *pachclient.APIClient, root string, tree hashtree.HashTree, pipes bool, concurrency int) error {
	limiter := limit.New(concurrency)
	var eg errgroup.Group
	if err := tree.Walk(func(path string, node *hashtree.NodeProto) error {
		if node.FileNode != nil {
			path := filepath.Join(root, path)
			var hashes []string
			for _, object := range node.FileNode.Objects {
				hashes = append(hashes, object.Hash)
			}
			if pipes {
				return p.makePipe(path, func(w io.Writer) error {
					return client.GetObjects(hashes, 0, 0, w)
				})
			}
			limiter.Acquire()
			eg.Go(func() (retErr error) {
				defer limiter.Release()
				return p.makeFile(path, func(w io.Writer) error {
					return client.GetObjects(hashes, 0, 0, w)
				})
			})
		}
		return nil
	}); err != nil {
		return err
	}
	return eg.Wait()
}

// CleanUp cleans up blocked syscalls for pipes that were never opened. And
// returns the total number of bytes that have been pulled/pushed. It also
// returns any errors that might have been encountered while trying to read
// data for the pipes. CleanUp should be called after all code that might
// access pipes has completed running, it should not be called concurrently.
func (p *Puller) CleanUp() (int64, error) {
	var result error
	select {
	case result = <-p.errCh:
	default:
	}

	// Open all the pipes to unblock the goros
	var pipes []io.Closer
	func() {
		p.Lock()
		defer p.Unlock()
		p.cleaned = true
		for path := range p.pipes {
			f, err := os.OpenFile(path, syscall.O_NONBLOCK+os.O_RDONLY, os.ModeNamedPipe)
			if err != nil && result == nil {
				result = err
			}
			pipes = append(pipes, f)
		}
		p.pipes = make(map[string]bool)
	}()

	// Wait for all goros to exit
	p.wg.Wait()

	// Close the pipes
	for _, pipe := range pipes {
		if err := pipe.Close(); err != nil && result == nil {
			result = err
		}
	}
	size := p.size
	p.size = 0
	return size, result
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
	sem := make(chan struct{}, 200)
	if err := pachClient.Walk(commit.Repo.Name, commit.ID, "", func(fileInfo *pfs.FileInfo) error {
		if fileInfo.FileType != pfs.FileType_FILE {
			return nil
		}
		eg.Go(func() (retErr error) {
			sem <- struct{}{}
			defer func() { <-sem }()
			w, err := objClient.Writer(filepath.Join(root, fileInfo.File.Path))
			if err != nil {
				return err
			}
			defer func() {
				if err := w.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			return pachClient.GetFile(commit.Repo.Name, commit.ID, fileInfo.File.Path, 0, 0, w)
		})
		return nil
	}); err != nil {
		return err
	}
	return eg.Wait()
}
