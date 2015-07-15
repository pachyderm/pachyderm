package drive

import (
	"io"

	"github.com/pachyderm/pachyderm/src/pfs"
)

type Driver interface {
	Init() error
	DriverType() pfs.DriverType
	InitRepository(repository *pfs.Repository, shard int) error
	GetFile(path *pfs.Path, shard int) (io.Reader, error)
	PutFile(path *pfs.Path, shard int, reader io.Reader) error
	ListFiles(path *pfs.Path, shard int) ([]*pfs.Path, error)
	GetParent(commit *pfs.Commit) (*pfs.Commit, error)
	GetChildren(commit *pfs.Commit) (*pfs.Commit, error)
	Branch(commit *pfs.Commit) (*pfs.Commit, error)
	Commit(commit *pfs.Commit) error
	PullDiff(commit *pfs.Commit, shard int) (io.Reader, error)
	PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error
	GetCommitInfo(commit *pfs.Commit) (*pfs.CommitInfo, error)
}

func NewInMemoryDriver() Driver {
	return newInMemoryDriver()
}

func NewBtrfsDriver() Driver {
	return newBtrfsDriver()
}
