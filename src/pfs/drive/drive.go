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
	RepoCreate(repo *pfs.Repo) error
	RepoInspect(repo *pfs.Repo, shard uint64) (*pfs.RepoInfo, error)
	RepoList(shard uint64) ([]*pfs.RepoInfo, error)
	RepoDelete(repo *pfs.Repo, shards map[uint64]bool) error
	CommitStart(parent *pfs.Commit, commit *pfs.Commit, shards map[uint64]bool) (*pfs.Commit, error)
	CommitFinish(commit *pfs.Commit, shards map[uint64]bool) error
	CommitInspect(commit *pfs.Commit, shard uint64) (*pfs.CommitInfo, error)
	CommitList(repo *pfs.Repo, shard uint64) ([]*pfs.CommitInfo, error)
	CommitDelete(commit *pfs.Commit, shards map[uint64]bool) error
	FilePut(file *pfs.File, shard uint64, offset int64, reader io.Reader) error
	MakeDirectory(file *pfs.File, shards map[uint64]bool) error
	FileGet(file *pfs.File, shard uint64) (ReaderAtCloser, error)
	FileInspect(file *pfs.File, shard uint64) (*pfs.FileInfo, error)
	FileList(file *pfs.File, shard uint64) ([]*pfs.FileInfo, error)
	FileDelete(file *pfs.File, shard uint64) error
	DiffPull(commit *pfs.Commit, shard uint64, diff io.Writer) error
	DiffPush(commit *pfs.Commit, diff io.Reader) error
}
