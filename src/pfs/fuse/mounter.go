package fuse

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/pachyderm/pachyderm/src/pfs"
)

const (
	namePrefix = "pfs://"
	subtype    = "pfs"
)

type mounter struct {
	ready chan bool
}

func newMounter() Mounter {
	return &mounter{make(chan bool)}
}

func (m *mounter) Mount(
	apiClient pfs.ApiClient,
	repositoryName string,
	commitID string,
	mountPoint string,
	shard uint64,
	modulus uint64,
) (retErr error) {
	if err := os.MkdirAll(mountPoint, 0777); err != nil {
		return err
	}
	conn, err := fuse.Mount(
		mountPoint,
		fuse.FSName(namePrefix+repositoryName),
		fuse.VolumeName(namePrefix+repositoryName),
		fuse.Subtype(subtype),
		fuse.WritebackCache(),
		fuse.MaxReadahead(1<<32-1),
	)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	close(m.ready)
	if err := fs.Serve(conn, newFilesystem(apiClient, repositoryName, commitID, shard, modulus)); err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-conn.Ready
	return conn.MountError
}

func (m *mounter) Unmount(mountPoint string) error {
	return fuse.Unmount(mountPoint)
}

func (m *mounter) Ready() {
	<-m.ready
}
