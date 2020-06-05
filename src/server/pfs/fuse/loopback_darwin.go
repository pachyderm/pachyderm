// +build darwin

// Copyright 2019 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuse

import (
	"context"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/hanwen/go-fuse/v2/fs"
)

func (n *loopbackNode) Getxattr(ctx context.Context, attr string, dest []byte) (uint32, syscall.Errno) {
	sz, err := unix.Getxattr(n.path(), attr, dest)
	return uint32(sz), fs.ToErrno(err)
}

func (n *loopbackNode) Setxattr(ctx context.Context, attr string, data []byte, flags uint32) syscall.Errno {
	err := unix.Setxattr(n.path(), attr, data, int(flags))
	return fs.ToErrno(err)
}

func (n *loopbackNode) Removexattr(ctx context.Context, attr string) syscall.Errno {
	err := unix.Removexattr(n.path(), attr)
	return fs.ToErrno(err)
}

func (n *loopbackNode) Listxattr(ctx context.Context, dest []byte) (uint32, syscall.Errno) {
	sz, err := unix.Listxattr(n.path(), dest)
	return uint32(sz), fs.ToErrno(err)
}

func (n *loopbackNode) renameExchange(name string, newparent *loopbackNode, newName string) syscall.Errno {
	return syscall.ENOSYS
}

func (n *loopbackNode) CopyFileRange(ctx context.Context, fhIn fs.FileHandle,
	offIn uint64, out *fs.Inode, fhOut fs.FileHandle, offOut uint64,
	len uint64, flags uint64) (uint32, syscall.Errno) {
	return 0, syscall.ENOSYS
}
