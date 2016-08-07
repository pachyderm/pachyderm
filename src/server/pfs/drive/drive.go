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
	CreateRepo(repo *pfs.Repo, created *google_protobuf.Timestamp, provenance []*pfs.Repo, shards map[uint64]bool) error
	InspectRepo(repo *pfs.Repo, shards map[uint64]bool) (*pfs.RepoInfo, error)
	ListRepo(provenance []*pfs.Repo, shards map[uint64]bool) ([]*pfs.RepoInfo, error)
	DeleteRepo(repo *pfs.Repo, shards map[uint64]bool, force bool) error
	StartCommit(repo *pfs.Repo, commitID string, parentID string, branch string, started *google_protobuf.Timestamp, provenance []*pfs.Commit, shards map[uint64]bool) error
	FinishCommit(commit *pfs.Commit, finished *google_protobuf.Timestamp, cancel bool, shards map[uint64]bool) error
	InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error)
	ListCommit(repo []*pfs.Repo, commitType pfs.CommitType, fromCommit []*pfs.Commit,
		provenance []*pfs.Commit, all bool, shards map[uint64]bool, block bool) ([]*pfs.CommitInfo, error)
	ListBranch(repo *pfs.Repo, shards map[uint64]bool) ([]*pfs.CommitInfo, error)
	DeleteCommit(commit *pfs.Commit, shards map[uint64]bool) error
	PutFile(file *pfs.File, handle string, delimiter pfs.Delimiter, shard uint64, reader io.Reader) error
	MakeDirectory(file *pfs.File, shard uint64) error
	GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64,
		size int64, from *pfs.Commit, shard uint64, unsafe bool, handle string) (io.ReadCloser, error)
	InspectFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64, unsafe bool, handle string) (*pfs.FileInfo, error)
	ListFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64, recurse bool, unsafe bool, handle string) ([]*pfs.FileInfo, error)
	DeleteFile(file *pfs.File, shard uint64, unsafe bool, handle string) error
	DeleteAll(shards map[uint64]bool) error
	AddShard(shard uint64) error
	DeleteShard(shard uint64) error
	Dump()
}

type CommitID string // master/0

// Driver represents a low-level pfs storage driver.
type PfsRefactorDriver interface {
	CreateRepo(repo *pfs.Repo, provenance []*pfs.Repo) error
	InspectRepo(repo *pfs.Repo) (*pfs.RepoInfo, error)
	ListRepo(provenance []*pfs.Repo) ([]*pfs.RepoInfo, error)
	DeleteRepo(repo *pfs.Repo, force bool) error
	StartCommit(repo *pfs.Repo, branch string, provenance []*pfs.Commit) (CommitID, error)
	StartCommitNewBranch(repo *pfs.Repo, parentID string, branch string, provenance []*pfs.Commit) (CommitID, error)
	FinishCommit(commit CommitID, cancel bool) error
	InspectCommit(commit CommitID) (*pfs.CommitInfo, error)
	ListCommit(repo *pfs.Repo, branch string, commitType pfs.CommitType, fromCommit *pfs.Commit, provenance []*pfs.Commit, all bool) ([]*pfs.CommitInfo, error)
	ListBranch(repo *pfs.Repo) ([]string, error)
	DeleteCommit(commit CommitID) error
	PutFile(file *pfs.File, delimiter pfs.Delimiter, reader io.Reader) error
	MakeDirectory(file *pfs.File) error
	GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64,
		size int64, from CommitID, unsafe bool) (io.ReadCloser, error)
	InspectFile(file *pfs.File, filterShard *pfs.Shard, from CommitID, unsafe bool) (*pfs.FileInfo, error)
	ListFile(file *pfs.File, filterShard *pfs.Shard, from CommitID, recurse bool, unsafe bool) ([]*pfs.FileInfo, error)
	DeleteFile(file *pfs.File, unsafe bool) error
	DeleteAll() error
	AddShard() error
	DeleteShard() error
	Squash(from []*pfs.Commit, to *pfs.Commit) error
	Merge(from []*pfs.Commit, parent *pfs.Commit, strategy pfs.MergeStrategy) error
	Dump()
}

// NewDriver creates a Driver.
func NewDriver(blockAddress string) (Driver, error) {
	return newDriver(blockAddress)
}
