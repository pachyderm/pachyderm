package fuse

import (
	"fmt"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/pachyderm/pachyderm/src/client"
)

type filesystem struct {
	pathfs.FileSystem
	c *client.APIClient
}

func NewFileSystem(c *client.APIClient) pathfs.FileSystem {
	return &filesystem{
		FileSystem: pathfs.NewDefaultFileSystem(),
		c:          c,
	}
}

func Mount(mountPoint string, fs pathfs.FileSystem) error {
	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mountPoint, nfs.Root(), nil)
	if err != nil {
		return fmt.Errorf("nodefs.MountRoot: %v", err)
	}
	server.Serve()
	return nil
}
