package fuse

import (
	"context"

	"github.com/hanwen/go-fuse/v2/fs"
	"golang.org/x/sys/unix"
)

func (n *loopbackNode) renameExchange(name string, newparent *loopbackNode, newName string) unix.Errno {
	fd1, err := unix.Open(n.path(), unix.O_DIRECTORY, 0)
	if err != nil {
		return fs.ToErrno(err)
	}
	defer unix.Close(fd1)
	fd2, err := unix.Open(newparent.path(), unix.O_DIRECTORY, 0)
	defer unix.Close(fd2)
	if err != nil {
		return fs.ToErrno(err)
	}

	var st unix.Stat_t
	if err := unix.Fstat(fd1, &st); err != nil {
		return fs.ToErrno(err)
	}

	// Double check that nodes didn't change from under us.
	inode := &n.Inode
	if inode.Root() != inode && inode.StableAttr().Ino != n.root().idFromStat(&st).Ino {
		return unix.EBUSY
	}
	if err := unix.Fstat(fd2, &st); err != nil {
		return fs.ToErrno(err)
	}

	newinode := &newparent.Inode
	if newinode.Root() != newinode && newinode.StableAttr().Ino != n.root().idFromStat(&st).Ino {
		return unix.EBUSY
	}

	return fs.ToErrno(unix.Renameat2(fd1, name, fd2, newName, unix.RENAME_EXCHANGE))
}

func (n *loopbackNode) CopyFileRange(ctx context.Context, fhIn fs.FileHandle,
	offIn uint64, out *fs.Inode, fhOut fs.FileHandle, offOut uint64,
	len uint64, flags uint64) (uint32, unix.Errno) {
	lfIn, ok := fhIn.(*loopbackFile)
	if !ok {
		return 0, unix.ENOTSUP
	}
	lfOut, ok := fhOut.(*loopbackFile)
	if !ok {
		return 0, unix.ENOTSUP
	}

	signedOffIn := int64(offIn)
	signedOffOut := int64(offOut)
	count, err := unix.CopyFileRange(lfIn.fd, &signedOffIn, lfOut.fd, &signedOffOut, int(len), int(flags))
	return uint32(count), fs.ToErrno(err)
}
