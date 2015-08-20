/*
Package drive provides the definitions for the low-level pfs storage drivers.
*/
package drive

import (
	"io"

	"github.com/pachyderm/pachyderm/src/pfs"
)

const (
	// InitialCommitID is the initial id before any commits are made in a repository.
	InitialCommitID = "scratch"
)

var (
	// ReservedCommitIDs are the commit ids used by the system.
	ReservedCommitIDs = map[string]bool{
		InitialCommitID: true,
	}
)

// ReaderAtCloser is an interface that implements both io.ReaderAt and io.Closer.
type ReaderAtCloser interface {
	io.ReaderAt
	io.Closer
}

// Driver represents a low-level pfs storage driver.
type Driver interface {
	InitRepository(repository *pfs.Repository, shard map[int]bool) error
	GetFile(path *pfs.Path, shard int) (ReaderAtCloser, error)
	GetFileInfo(path *pfs.Path, shard int) (*pfs.FileInfo, bool, error)
	MakeDirectory(path *pfs.Path, shards map[int]bool) error
	PutFile(path *pfs.Path, shard int, offset int64, reader io.Reader) error
	ListFiles(path *pfs.Path, shard int) ([]*pfs.FileInfo, error)
	Branch(commit *pfs.Commit, newCommit *pfs.Commit, shards map[int]bool) (*pfs.Commit, error)
	Commit(commit *pfs.Commit, shards map[int]bool) error
	PullDiff(commit *pfs.Commit, shard int, diff io.Writer) error
	PushDiff(commit *pfs.Commit, diff io.Reader) error
	GetCommitInfo(commit *pfs.Commit, shard int) (*pfs.CommitInfo, bool, error)
	ListCommits(repository *pfs.Repository, shard int) ([]*pfs.CommitInfo, error)
}
