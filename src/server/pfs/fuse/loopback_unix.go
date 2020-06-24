// Copyright 2019 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuse

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"golang.org/x/sys/unix"
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
