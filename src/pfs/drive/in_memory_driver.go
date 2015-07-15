package drive

import (
	"io"

	"github.com/pachyderm/pachyderm/src/pfs"
)

type inMemoryDriver struct{}

func newInMemoryDriver() *inMemoryDriver {
	return &inMemoryDriver{}
}

func (i *inMemoryDriver) Init() error {
	return nil
}

func (i *inMemoryDriver) DriverType() pfs.DriverType {
	return pfs.DriverType_DRIVER_TYPE_IN_MEMORY
}

func (i *inMemoryDriver) InitRepository(repository *pfs.Repository, shard int) error {
	return nil
}

func (i *inMemoryDriver) GetFile(path *pfs.Path, shard int) (io.ReadCloser, error) {
	return nil, nil
}

func (i *inMemoryDriver) PutFile(path *pfs.Path, shard int, reader io.Reader) error {
	return nil
}

func (i *inMemoryDriver) ListFiles(path *pfs.Path, shard int) ([]*pfs.Path, error) {
	return nil, nil
}

func (i *inMemoryDriver) GetParent(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (i *inMemoryDriver) GetChildren(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (i *inMemoryDriver) Branch(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (i *inMemoryDriver) Commit(commit *pfs.Commit) error {
	return nil
}

func (i *inMemoryDriver) PullDiff(commit *pfs.Commit, shard int) (io.Reader, error) {
	return nil, nil
}

func (i *inMemoryDriver) PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error {
	return nil
}

func (i *inMemoryDriver) GetCommitInfo(commit *pfs.Commit) (*pfs.CommitInfo, error) {
	return nil, nil
}
