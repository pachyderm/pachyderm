package drive

import (
	"io"

	"github.com/pachyderm/pachyderm/src/pfs"
)

const (
	SystemRootCommitID = "__root__"
	InitialCommitID    = "scratch"
)

var (
	ReservedCommitIDs = map[string]bool{
		SystemRootCommitID: true,
	}
)

type Driver interface {
	Init() error
	DriverType() pfs.DriverType
	InitRepository(repository *pfs.Repository, shard int) error
	GetFile(path *pfs.Path, shard int) (io.ReadCloser, error)
	MakeDirectory(path *pfs.Path, shard int) error
	PutFile(path *pfs.Path, shard int, reader io.Reader) error
	ListFiles(path *pfs.Path, shard int) ([]*pfs.Path, error)
	GetParent(commit *pfs.Commit) (*pfs.Commit, error)
	Branch(commit *pfs.Commit) (*pfs.Commit, error)
	Commit(commit *pfs.Commit) error
	PullDiff(commit *pfs.Commit, shard int) (io.Reader, error)
	PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error
	GetCommitInfo(commit *pfs.Commit) (*pfs.CommitInfo, error)
}

func NewBtrfsDriver(rootDir string) Driver {
	return newBtrfsDriver(rootDir)
}
