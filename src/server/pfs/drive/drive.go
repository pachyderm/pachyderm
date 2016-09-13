/*
Package drive provides the definitions for the low-level pfs storage drivers.
*/
package drive

import (
	"io"
	"strings"

	"go.pedge.io/pb/go/google/protobuf"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// IsPermissionError returns true if a given error is a permission error.
func IsPermissionError(err error) bool {
	return strings.Contains(err.Error(), "has already been finished")
}

// Driver represents a low-level pfs storage driver.
type Driver interface {
	CreateRepo(repo *pfs.Repo, provenance []*pfs.Repo) error
	InspectRepo(repo *pfs.Repo) (*pfs.RepoInfo, error)
	ListRepo(provenance []*pfs.Repo) ([]*pfs.RepoInfo, error)
	DeleteRepo(repo *pfs.Repo, force bool) error
	Fork(parent *pfs.Commit, branch string, provenance []*pfs.Commit) (*pfs.Commit, error)
	StartCommit(parent *pfs.Commit, provenance []*pfs.Commit) (*pfs.Commit, error)
	FinishCommit(commit *pfs.Commit, finished *google_protobuf.Timestamp, cancel bool) error
	ArchiveCommit(commit []*pfs.Commit) error
	InspectCommit(commit *pfs.Commit) (*pfs.CommitInfo, error)
	ListCommit(repo []*pfs.Repo, commitType pfs.CommitType, fromCommit []*pfs.Commit,
		provenance []*pfs.Commit, status pfs.CommitStatus, block bool) ([]*pfs.CommitInfo, error)
	FlushCommit(fromCommits []*pfs.Commit, toRepos []*pfs.Repo) ([]*pfs.CommitInfo, error)
	ListBranch(repo *pfs.Repo) ([]*pfs.CommitInfo, error)
	DeleteCommit(commit *pfs.Commit) error
	PutFile(file *pfs.File, delimiter pfs.Delimiter, reader io.Reader) error
	MakeDirectory(file *pfs.File) error
	GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64,
		size int64, diffMethod *pfs.DiffMethod) (io.ReadCloser, error)
	InspectFile(file *pfs.File, filterShard *pfs.Shard, diffMethod *pfs.DiffMethod) (*pfs.FileInfo, error)
	ListFile(file *pfs.File, filterShard *pfs.Shard, diffMethod *pfs.DiffMethod, recurse bool) ([]*pfs.FileInfo, error)
	DeleteFile(file *pfs.File) error
	DeleteAll() error
	ArchiveAll() error
	Dump()
	// If strategy is SQUASH, to is the ID of an open commit
	// If strategy is REPLAY, to is a branch name
	Merge(repo *pfs.Repo, commits []*pfs.Commit, to string, strategy pfs.MergeStrategy, cancel bool) (*pfs.Commits, error)
}

// CommitID is an alias for string
type CommitID string // master/0
type Repo string

// NewDriver
type NewDriver interface {
	CreateRepo(repo, provenance []*pfs.Repo) error
	InspectRepo(repo *pfs.Repo) (*pfs.RepoInfo, error)
	ListRepo(provenance []*pfs.Repo) ([]*pfs.RepoInfo, error)
	DeleteRepo(repo *pfs.Repo, force bool) error
	Fork(parent *pfs.Commit, branch string, provenance []*pfs.Commit) (*pfs.Commit, error)
	StartCommit(parent *pfs.Commit, provenance []*pfs.Commit) (*pfs.Commit, error)
	FinishCommit(commit *pfs.Commit, cancel bool) error
	ArchiveCommits(commit []*pfs.Commit) error
	InspectCommit(commit *pfs.Commit) (*pfs.CommitInfo, error)
	ListCommit(fromCommits []*pfs.Commit, provenance []*pfs.Commit, commitType pfs.CommitType, status pfs.CommitStatus, block bool) ([]*pfs.CommitInfo, error)
	FlushCommit(fromCommits []*pfs.Commit, toRepos []*pfs.Repo) ([]*pfs.CommitInfo, error)
	ListBranch(repo *pfs.Repo) ([]string, error)
	DeleteCommit(commit *pfs.Commit) error
	PutFile(file *pfs.File, delimiter pfs.Delimiter, reader io.Reader) error
	MakeDirectory(file *pfs.File) error
	GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64,
		size int64, diffMethod *pfs.DiffMethod) (io.ReadCloser, error)
	InspectFile(file *pfs.File, filterShard *pfs.Shard, diffMethod *pfs.DiffMethod) (*pfs.FileInfo, error)
	ListFile(file *pfs.File, filterShard *pfs.Shard, diffMethod *pfs.DiffMethod, recurse bool) ([]*pfs.FileInfo, error)
	DeleteFile(file *pfs.File) error
	DeleteAll() error
	ArchiveAll() error
	Dump()
	Squash(fromCommits []*pfs.Commit, toCommit *pfs.Commit) error
	Replay(fromCommits []*pfs.Commit, toBranch string) ([]*pfs.Commit, error)
}
