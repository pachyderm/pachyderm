package fuse

import (
	"fmt"
	"os"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"go.pachyderm.com/pachyderm/src/pfs"
)

const (
	namePrefix = "pfs://"
	subtype    = "pfs"
)

type mounter struct {
	apiClient pfs.APIClient

	mountpointToErrChan map[string]<-chan error
	lock                *sync.Mutex
}

func newMounter(apiClient pfs.APIClient) Mounter {
	return &mounter{
		apiClient,
		make(map[string]<-chan error),
		&sync.Mutex{},
	}
}

func (m *mounter) Mount(
	repositoryName string,
	commitID string,
	mountPoint string,
	shard uint64,
	modulus uint64,
) (retErr error) {
	// TODO(pedge): should we make the caller do this?
	if err := os.MkdirAll(mountPoint, 0777); err != nil {
		return err
	}
	name := namePrefix + repositoryName
	if commitID != "" {
		name = name + "-" + commitID
	}
	conn, err := fuse.Mount(
		mountPoint,
		fuse.FSName(name),
		fuse.VolumeName(name),
		fuse.Subtype(subtype),
		fuse.AllowOther(),
		fuse.WritebackCache(),
		fuse.MaxReadahead(1<<32-1),
	)
	if err != nil {
		return err
	}
	errChan := make(chan error, 1)
	m.lock.Lock()
	if _, ok := m.mountpointToErrChan[mountPoint]; ok {
		m.lock.Unlock()
		return fmt.Errorf("mountpoint %s already exists", mountPoint)
	}
	m.mountpointToErrChan[mountPoint] = errChan
	m.lock.Unlock()
	go func() {
		err := fs.Serve(conn, newFilesystem(m.apiClient, repositoryName, commitID, shard, modulus))
		closeErr := conn.Close()
		if err != nil {
			errChan <- err
		} else {
			errChan <- closeErr
		}
	}()
	<-conn.Ready
	return conn.MountError
}

func (m *mounter) Unmount(mountPoint string) error {
	return fuse.Unmount(mountPoint)
}

func (m *mounter) Wait(mountPoint string) error {
	m.lock.Lock()
	errChan, ok := m.mountpointToErrChan[mountPoint]
	if !ok {
		m.lock.Unlock()
		return fmt.Errorf("mountpoint %s does not exist", mountPoint)
	}
	delete(m.mountpointToErrChan, mountPoint)
	m.lock.Unlock()
	return <-errChan
}
