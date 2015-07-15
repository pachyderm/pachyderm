package drive

import (
	"io"

	"github.com/pachyderm/pachyderm/src/pfs"
)

type btrfsDriver struct{}

func newBtrfsDriver() *btrfsDriver {
	return &btrfsDriver{}
}

func (b *btrfsDriver) Init() error {
	return nil
}

func (b *btrfsDriver) DriverType() pfs.DriverType {
	return pfs.DriverType_DRIVER_TYPE_BTRFS
}

func (b *btrfsDriver) InitRepository(repository *pfs.Repository, shard int) error {
	return nil
}

func (b *btrfsDriver) GetFile(path *pfs.Path, shard int) (io.Reader, error) {
	return nil, nil
}

func (b *btrfsDriver) PutFile(path *pfs.Path, shard int, reader io.Reader) error {
	return nil
}

func (b *btrfsDriver) ListFiles(path *pfs.Path, shard int) ([]*pfs.Path, error) {
	return nil, nil
}

func (b *btrfsDriver) GetParent(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (b *btrfsDriver) GetChildren(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (b *btrfsDriver) Branch(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (b *btrfsDriver) Commit(commit *pfs.Commit) error {
	return nil
}

func (b *btrfsDriver) PullDiff(commit *pfs.Commit, shard int) (io.Reader, error) {
	return nil, nil
}

func (b *btrfsDriver) PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error {
	return nil
}

func (b *btrfsDriver) GetCommitInfo(commit *pfs.Commit) (*pfs.CommitInfo, error) {
	return nil, nil
}
