package drive

import (
	"io"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/btrfs"
)

const (
	InitialCommitID = "scratch"
)

var (
	ReservedCommitIDs = map[string]bool{
		InitialCommitID: true,
	}
)

type Driver interface {
	Init() error
	InitRepository(repository *pfs.Repository, shard map[int]bool) error
	GetFile(path *pfs.Path, shard int) (io.ReadCloser, error)
	GetFileInfo(path *pfs.Path, shard int) (*pfs.FileInfo, bool)
	MakeDirectory(path *pfs.Path, shards map[int]bool) error
	PutFile(path *pfs.Path, shard int, reader io.Reader) error
	ListFiles(path *pfs.Path, shard int) ([]*pfs.FileInfo, error)
	Branch(commit *pfs.Commit, newCommit *pfs.Commit, shards map[int]bool) (*pfs.Commit, error)
	Commit(commit *pfs.Commit, shards map[int]bool) error
	PullDiff(commit *pfs.Commit, shard int) (io.Reader, error)
	PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error
	GetCommitInfo(commit *pfs.Commit, shard int) (*pfs.CommitInfo, error)
	ListCommits(repository *pfs.Repository, shard int) ([]*pfs.CommitInfo, error)
}

func NewBtrfsDriver(rootDir string, btrfsAPI btrfs.API) Driver {
	return newBtrfsDriver(rootDir, btrfsAPI)
}
