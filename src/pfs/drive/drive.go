/*
Package drive provides the definitions for the low-level pfs storage drivers.
*/
package drive

import (
	"io"

	"github.com/pachyderm/pachyderm/src/pfs"
)

// ReaderAtCloser is an interface that implements both io.ReaderAt and io.Closer.
type ReaderAtCloser interface {
	io.ReaderAt
	io.Closer
}

// Driver represents a low-level pfs storage driver.
type Driver interface {
	RepoCreate(repo *pfs.Repo, shard map[int]bool) error
	RepoInspect(repo *pfs.Repo, shard int) (*pfs.RepoInfo, error)
	RepoList(shard int) ([]*pfs.RepoInfo, error)
	RepoDelete(repo *pfs.Repo, shard map[int]bool) error
	CommitStart(parent *pfs.Commit, commit *pfs.Commit, shard map[int]bool) (*pfs.Commit, error)
	CommitFinish(commit *pfs.Commit, shard map[int]bool) error
	CommitInspect(commit *pfs.Commit, shard int) (*pfs.CommitInfo, error)
	CommitList(repo *pfs.Repo, shard int) ([]*pfs.CommitInfo, error)
	CommitDelete(commit *pfs.Commit, shard map[int]bool) error
	FilePut(file *pfs.File, shard int, offset int64, reader io.Reader) error
	MakeDirectory(file *pfs.File, shard map[int]bool) error
	FileGet(file *pfs.File, shard int) (ReaderAtCloser, error)
	FileInspect(file *pfs.File, shard int) (*pfs.FileInfo, error)
	FileList(file *pfs.File, shard int) ([]*pfs.FileInfo, error)
	FileDelete(file *pfs.File, shard int) error
	DiffPull(commit *pfs.Commit, shard int, diff io.Writer) error
	DiffPush(commit *pfs.Commit, diff io.Reader) error
}
