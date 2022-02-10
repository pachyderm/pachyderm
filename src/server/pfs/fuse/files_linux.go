// Copyright 2019 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuse

import (
	"context"
	"time"
	"unsafe"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"golang.org/x/sys/unix"
)

func (f *loopbackFile) Allocate(ctx context.Context, off uint64, sz uint64, mode uint32) unix.Errno {
	f.mu.Lock()
	defer f.mu.Unlock()
	err := unix.Fallocate(f.fd, mode, int64(off), int64(sz))
	if err != nil {
		return fs.ToErrno(err)
	}
	return fs.OK
}

// Utimens - file handle based version of loopbackFileSystem.Utimens()
func (f *loopbackFile) utimens(a *time.Time, m *time.Time) unix.Errno {
	var ts [2]unix.Timespec
	ts[0] = fuse.UtimeToTimespec(a)
	ts[1] = fuse.UtimeToTimespec(m)
	err := futimens(int(f.fd), &ts)
	return fs.ToErrno(err)
}

// futimens - futimens(3) calls utimensat(2) with "pathname" set to null and
// "flags" set to zero
func futimens(fd int, times *[2]unix.Timespec) (err error) {
	_, _, e1 := unix.unix6(unix.SYS_UTIMENSAT, uintptr(fd), 0, uintptr(unsafe.Pointer(times)), uintptr(0), 0, 0)
	if e1 != 0 {
		err = unix.Errno(e1)
	}
	return
}
