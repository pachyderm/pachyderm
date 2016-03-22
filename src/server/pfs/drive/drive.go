/*
Package drive provides the definitions for the low-level pfs storage drivers.
*/
package drive

import (
	"io"

	"go.pedge.io/pb/go/google/protobuf"

	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
)

// Driver represents a low-level pfs storage driver.
type Driver interface {
	CreateRepo(repo *pfsserver.Repo, created *google_protobuf.Timestamp, shards map[uint64]bool) error
	InspectRepo(repo *pfsserver.Repo, shards map[uint64]bool) (*pfsserver.RepoInfo, error)
	ListRepo(shards map[uint64]bool) ([]*pfsserver.RepoInfo, error)
	DeleteRepo(repo *pfsserver.Repo, shards map[uint64]bool) error
	StartCommit(repo *pfsserver.Repo, commitID string, parentID string, branch string, started *google_protobuf.Timestamp, shards map[uint64]bool) error
	FinishCommit(commit *pfsserver.Commit, finished *google_protobuf.Timestamp, shards map[uint64]bool) error
	InspectCommit(commit *pfsserver.Commit, shards map[uint64]bool) (*pfsserver.CommitInfo, error)
	ListCommit(repo []*pfsserver.Repo, fromCommit []*pfsserver.Commit, shards map[uint64]bool) ([]*pfsserver.CommitInfo, error)
	ListBranch(repo *pfsserver.Repo, shards map[uint64]bool) ([]*pfsserver.CommitInfo, error)
	DeleteCommit(commit *pfsserver.Commit, shards map[uint64]bool) error
	PutFile(file *pfsserver.File, shard uint64, offset int64, reader io.Reader) error
	MakeDirectory(file *pfsserver.File, shards map[uint64]bool) error
	GetFile(file *pfsserver.File, filterShard *pfsserver.Shard, offset int64, size int64, from *pfsserver.Commit, shard uint64) (io.ReadCloser, error)
	InspectFile(file *pfsserver.File, filterShard *pfsserver.Shard, from *pfsserver.Commit, shard uint64) (*pfsserver.FileInfo, error)
	ListFile(file *pfsserver.File, filterShard *pfsserver.Shard, from *pfsserver.Commit, shard uint64) ([]*pfsserver.FileInfo, error)
	DeleteFile(file *pfsserver.File, shard uint64) error
	AddShard(shard uint64) error
	DeleteShard(shard uint64) error
	Dump()
}

func NewDriver(blockAddress string) (Driver, error) {
	return newDriver(blockAddress)
}
