/*
Package drive provides the definitions for the low-level pfs storage drivers.
*/
package drive

import (
	"io"

	"go.pedge.io/pb/go/google/protobuf"

	. "github.com/pachyderm/pachyderm/src/client/pfs"
)

// Driver represents a low-level pfs storage driver.
type Driver interface {
	CreateRepo(repo *Repo, created *google_protobuf.Timestamp, shards map[uint64]bool) error
	InspectRepo(repo *Repo, shards map[uint64]bool) (*RepoInfo, error)
	ListRepo(shards map[uint64]bool) ([]*RepoInfo, error)
	DeleteRepo(repo *Repo, shards map[uint64]bool) error
	StartCommit(repo *Repo, commitID string, parentID string, branch string, started *google_protobuf.Timestamp, shards map[uint64]bool) error
	FinishCommit(commit *Commit, finished *google_protobuf.Timestamp, shards map[uint64]bool) error
	InspectCommit(commit *Commit, shards map[uint64]bool) (*CommitInfo, error)
	ListCommit(repo []*Repo, fromCommit []*Commit, shards map[uint64]bool) ([]*CommitInfo, error)
	ListBranch(repo *Repo, shards map[uint64]bool) ([]*CommitInfo, error)
	DeleteCommit(commit *Commit, shards map[uint64]bool) error
	PutFile(file *File, shard uint64, offset int64, reader io.Reader) error
	MakeDirectory(file *File, shards map[uint64]bool) error
	GetFile(file *File, filterShard *Shard, offset int64, size int64, from *Commit, shard uint64) (io.ReadCloser, error)
	InspectFile(file *File, filterShard *Shard, from *Commit, shard uint64) (*FileInfo, error)
	ListFile(file *File, filterShard *Shard, from *Commit, shard uint64) ([]*FileInfo, error)
	DeleteFile(file *File, shard uint64) error
	AddShard(shard uint64) error
	DeleteShard(shard uint64) error
	Dump()
}

func NewDriver(blockAddress string) (Driver, error) {
	return newDriver(blockAddress)
}
