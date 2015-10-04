/*
Package drive provides the definitions for the low-level pfs storage drivers.
*/
package drive //import "go.pachyderm.com/pachyderm/src/pfs/drive"

import (
	"io"

	"go.pachyderm.com/pachyderm/src/pfs"
)

// ReaderAtCloser is an interface that implements both io.ReaderAt and io.Closer.
type ReaderAtCloser interface {
	io.Reader
	io.ReaderAt
	io.Closer
}

// Driver represents a low-level pfs storage driver.
type Driver interface {
	CreateRepo(repo *pfs.Repo) error
	InspectRepo(repo *pfs.Repo, shard uint64) (*pfs.RepoInfo, error)
	ListRepo(shard uint64) ([]*pfs.RepoInfo, error)
	DeleteRepo(repo *pfs.Repo, shards map[uint64]bool) error
	StartCommit(parent *pfs.Commit, commit *pfs.Commit, shards map[uint64]bool) error
	FinishCommit(commit *pfs.Commit, shards map[uint64]bool) error
	InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error)
	ListCommit(repo *pfs.Repo, from *pfs.Commit, shards map[uint64]bool) ([]*pfs.CommitInfo, error)
	DeleteCommit(commit *pfs.Commit, shards map[uint64]bool) error
	PutBlock(file *pfs.File, block *pfs.Block, shard uint64, reader io.Reader) error
	GetBlock(block *pfs.Block, shard uint64) (ReaderAtCloser, error)
	InspectBlock(block *pfs.Block, shard uint64) (*pfs.BlockInfo, error)
	ListBlock(shard uint64) ([]*pfs.BlockInfo, error)
	PutFile(file *pfs.File, shard uint64, offset int64, reader io.Reader) error
	MakeDirectory(file *pfs.File, shards map[uint64]bool) error
	GetFile(file *pfs.File, shard uint64) (ReaderAtCloser, error)
	InspectFile(file *pfs.File, shard uint64) (*pfs.FileInfo, error)
	ListFile(file *pfs.File, shard uint64) ([]*pfs.FileInfo, error)
	ListChange(file *pfs.File, from *pfs.Commit, shard uint64) ([]*pfs.Change, error)
	DeleteFile(file *pfs.File, shard uint64) error
	PullDiff(commit *pfs.Commit, shard uint64, diff io.Writer) error
	PushDiff(commit *pfs.Commit, shard uint64, diff io.Reader) error
}
