package fuse

import (
	"context"

	"github.com/hanwen/go-fuse/v2/fs"
	"golang.org/x/sys/unix"
)

func (n *loopbackNode) renameExchange(name string, newparent *loopbackNode, newName string) unix.Errno {
	return unix.ENOSYS
}

func (n *loopbackNode) CopyFileRange(ctx context.Context, fhIn fs.FileHandle,
	offIn uint64, out *fs.Inode, fhOut fs.FileHandle, offOut uint64,
	len uint64, flags uint64) (uint32, unix.Errno) {
	return 0, unix.ENOSYS
}
