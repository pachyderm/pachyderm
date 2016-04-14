/*
Package drive provides the definitions for the low-level pfs storage drivers.
*/
package drive

import (
	"io"

	"go.pedge.io/pb/go/google/protobuf"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// Driver represents a low-level pfs storage driver.
type Driver interface {
	CreateRepo(repo *pfs.Repo, created *google_protobuf.Timestamp, shards map[uint64]bool) error
	InspectRepo(repo *pfs.Repo, shards map[uint64]bool) (*pfs.RepoInfo, error)
	ListRepo(shards map[uint64]bool) ([]*pfs.RepoInfo, error)
	DeleteRepo(repo *pfs.Repo, shards map[uint64]bool) error
	StartCommit(repo *pfs.Repo, commitID string, parentID string, branch string, started *google_protobuf.Timestamp, shards map[uint64]bool) error
	FinishCommit(commit *pfs.Commit, finished *google_protobuf.Timestamp, shards map[uint64]bool) error
	InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error)
	ListCommit(repo []*pfs.Repo, fromCommit []*pfs.Commit, shards map[uint64]bool) ([]*pfs.CommitInfo, error)
	ListBranch(repo *pfs.Repo, shards map[uint64]bool) ([]*pfs.CommitInfo, error)
	DeleteCommit(commit *pfs.Commit, shards map[uint64]bool) error
	PutFile(file *pfs.File, shard uint64, reader io.Reader) error
	MakeDirectory(file *pfs.File, shard uint64) error
	GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64, size int64, from *pfs.Commit, shard uint64) (io.ReadCloser, error)
	InspectFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64) (*pfs.FileInfo, error)
	ListFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64) ([]*pfs.FileInfo, error)
	DeleteFile(file *pfs.File, shard uint64) error
	AddShard(shard uint64) error
	DeleteShard(shard uint64) error
	Dump()
}

func NewDriver(blockAddress string) (Driver, error) {
	return newDriver(blockAddress)
}
