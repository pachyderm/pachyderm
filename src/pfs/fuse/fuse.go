package fuse

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/pachyderm/pachyderm/src/pfs"
	"golang.org/x/net/context"
)

func Mount(apiClient pfs.ApiClient, repositoryName string, mountPoint string) error {
	if err := os.MkdirAll(mountPoint, 0777); err != nil {
		return err
	}
	c, err := fuse.Mount(
		mountPoint,
		fuse.FSName("pfs"),
		fuse.Subtype("hellofs"),
		fuse.VolumeName("pfs://"+repositoryName),
	)
	if err != nil {
		return err
	}
	defer c.Close()

	if err := fs.Serve(c, &filesystem{apiClient, repositoryName}); err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	return c.MountError
}

func Unmount(mountPoint string) error {
	return nil
}

type filesystem struct {
	apiClient      pfs.ApiClient
	repositoryName string
}

func (*filesystem) Root() (fs.Node, error) {
	return &directory{"/"}, nil
}

type directory struct {
	path string
}

func (*directory) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}
