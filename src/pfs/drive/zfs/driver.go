// +build linux

package zfs

import (
	"io"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
)

type driver struct {
	rootDir string
}

func newDriver(rootDir string) *driver {
	if 1 == 1 {
		panic("TODO!!!!")
	}
	return &driver{rootDir}
}

func (d *driver) InitRepository(repository *pfs.Repository, shards map[int]bool) error {
	return nil
}

func (d *driver) GetFile(path *pfs.Path, shard int) (drive.ReaderAtCloser, error) {
	return nil, nil
}

func (d *driver) GetFileInfo(path *pfs.Path, shard int) (_ *pfs.FileInfo, ok bool, _ error) {
	return nil, false, nil
}

func (d *driver) MakeDirectory(path *pfs.Path, shards map[int]bool) error {
	return nil
}

func (d *driver) PutFile(path *pfs.Path, shard int, offset int64, reader io.Reader) error {
	return nil
}

func (d *driver) ListFiles(path *pfs.Path, shard int) (_ []*pfs.FileInfo, retErr error) {
	return nil, nil
}

func (d *driver) stat(path *pfs.Path, shard int) (*pfs.FileInfo, error) {
	return nil, nil
}

func (d *driver) Branch(commit *pfs.Commit, newCommit *pfs.Commit, shards map[int]bool) (*pfs.Commit, error) {
	return nil, nil
}

func (d *driver) Commit(commit *pfs.Commit, shards map[int]bool) error {
	return nil
}

func (d *driver) PullDiff(commit *pfs.Commit, shard int) (io.Reader, error) {
	return nil, nil
}

func (d *driver) PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error {
	return nil
}

func (d *driver) GetCommitInfo(commit *pfs.Commit, shard int) (_ *pfs.CommitInfo, ok bool, _ error) {
	return nil, false, nil
}

func (d *driver) ListCommits(repository *pfs.Repository, shard int) (_ []*pfs.CommitInfo, retErr error) {
	return nil, nil
}
