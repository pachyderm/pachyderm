package main

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/docker/docker/client"
	billy "github.com/go-git/go-billy/v5"
	nfs "github.com/willscott/go-nfs"
)

type pachNFS struct {
	client.APIClient
}

func (p pachNFS) Mount(context.Context, net.Conn, nfs.MountRequest) (nfs.MountStatus, billy.Filesystem, []nfs.AuthFlavor) {
	return nfs.MountStatusErrServerFault, nil, nil
}

// Change can return 'nil' if filesystem is read-only
func (p pachNFS) Change(billy.Filesystem) billy.Change {
	return nil // PachNFS filesystems are currently all read-only
}

// Optional methods - generic helpers or trivial implementations can be sufficient depending on use case.

// Fill in information about a file system's free space.
func (p pachNFS) FSStat(context.Context, billy.Filesystem, *nfs.FSStat) error {
	return errors.New("NFS function FSStat not implemented")
}

// represent file objects as opaque references
// Can be safely implemented via helpers/cachinghandler.
// func (p pachNFS) ToHandle(fs billy.Filesystem, path []string) []byte {
// 	return nil
// }

func (p pachNFS) FromHandle(fh []byte) (billy.Filesystem, []string, error) {
	return nil, nil, errors.New("NFS function FromHandle not implemented")
}

// How many handles can be safely maintained by the handler.
func (p pachNFS) HandleLimit() int {
	return 100 // maybe I can skip this too? Return 0 or something?
}

func main() {
	fmt.Println("vim-go")
}
