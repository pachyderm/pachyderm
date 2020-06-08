// Copyright 2019 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuse

import (
	"context"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"golang.org/x/sys/unix"
)

// newLoopbackFile creates a FileHandle out of a file descriptor. All
// operations are implemented.
func newLoopbackFile(fd int, root *loopbackRoot) fs.FileHandle {
	return &loopbackFile{fd: fd, root: root}
}

type loopbackFile struct {
	mu   sync.Mutex
	fd   int
	root *loopbackRoot
}

var _ = (fs.FileHandle)((*loopbackFile)(nil))
var _ = (fs.FileReleaser)((*loopbackFile)(nil))
var _ = (fs.FileGetattrer)((*loopbackFile)(nil))
var _ = (fs.FileReader)((*loopbackFile)(nil))
var _ = (fs.FileWriter)((*loopbackFile)(nil))
var _ = (fs.FileGetlker)((*loopbackFile)(nil))
var _ = (fs.FileSetlker)((*loopbackFile)(nil))
var _ = (fs.FileSetlkwer)((*loopbackFile)(nil))
var _ = (fs.FileLseeker)((*loopbackFile)(nil))
var _ = (fs.FileFlusher)((*loopbackFile)(nil))
var _ = (fs.FileFsyncer)((*loopbackFile)(nil))
var _ = (fs.FileSetattrer)((*loopbackFile)(nil))
var _ = (fs.FileAllocater)((*loopbackFile)(nil))

func (f *loopbackFile) Read(ctx context.Context, buf []byte, off int64) (res fuse.ReadResult, errno syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r := fuse.ReadResultFd(uintptr(f.fd), off, len(buf))
	return r, fs.OK
}

func (f *loopbackFile) Write(ctx context.Context, data []byte, off int64) (uint32, syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, err := syscall.Pwrite(f.fd, data, off)
	return uint32(n), fs.ToErrno(err)
}

func (f *loopbackFile) Release(ctx context.Context) syscall.Errno {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fd != -1 {
		err := syscall.Close(f.fd)
		f.fd = -1
		return fs.ToErrno(err)
	}
	return syscall.EBADF
}

func (f *loopbackFile) Flush(ctx context.Context) syscall.Errno {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Since Flush() may be called for each dup'd fd, we don't
	// want to really close the file, we just want to flush. This
	// is achieved by closing a dup'd fd.
	newFd, err := syscall.Dup(f.fd)

	if err != nil {
		return fs.ToErrno(err)
	}
	err = syscall.Close(newFd)
	return fs.ToErrno(err)
}

func (f *loopbackFile) Fsync(ctx context.Context, flags uint32) (errno syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := syscall.Fsync(f.fd); err != nil {
		return fs.ToErrno(err)
	}
	if err := f.root.sync(false); err != nil {
		return fs.ToErrno(err)
	}
	return fs.OK
}

const (
	// GETLK is constant for F_OFD_GETLK
	GETLK = 36
	// SETLK is constant for F_OFD_SETLK
	SETLK = 37
	// SETLKW is constant for F_OFD_SETLKW
	SETLKW = 38
)

func (f *loopbackFile) Getlk(ctx context.Context, owner uint64, lk *fuse.FileLock, flags uint32, out *fuse.FileLock) (errno syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()
	flk := syscall.Flock_t{}
	lk.ToFlockT(&flk)
	errno = fs.ToErrno(syscall.FcntlFlock(uintptr(f.fd), GETLK, &flk))
	out.FromFlockT(&flk)
	return
}

func (f *loopbackFile) Setlk(ctx context.Context, owner uint64, lk *fuse.FileLock, flags uint32) (errno syscall.Errno) {
	return f.setLock(ctx, owner, lk, flags, false)
}

func (f *loopbackFile) Setlkw(ctx context.Context, owner uint64, lk *fuse.FileLock, flags uint32) (errno syscall.Errno) {
	return f.setLock(ctx, owner, lk, flags, true)
}

func (f *loopbackFile) setLock(ctx context.Context, owner uint64, lk *fuse.FileLock, flags uint32, blocking bool) (errno syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if (flags & fuse.FUSE_LK_FLOCK) != 0 {
		var op int
		switch lk.Typ {
		case syscall.F_RDLCK:
			op = syscall.LOCK_SH
		case syscall.F_WRLCK:
			op = syscall.LOCK_EX
		case syscall.F_UNLCK:
			op = syscall.LOCK_UN
		default:
			return syscall.EINVAL
		}
		if !blocking {
			op |= syscall.LOCK_NB
		}
		return fs.ToErrno(syscall.Flock(f.fd, op))
	}
	flk := syscall.Flock_t{}
	lk.ToFlockT(&flk)
	var op int
	if blocking {
		op = SETLKW
	} else {
		op = SETLK
	}
	return fs.ToErrno(syscall.FcntlFlock(uintptr(f.fd), op, &flk))
}

func (f *loopbackFile) Setattr(ctx context.Context, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	if errno := f.setAttr(ctx, in); errno != 0 {
		return errno
	}

	return f.Getattr(ctx, out)
}

func (f *loopbackFile) setAttr(ctx context.Context, in *fuse.SetAttrIn) syscall.Errno {
	f.mu.Lock()
	defer f.mu.Unlock()
	var errno syscall.Errno
	if mode, ok := in.GetMode(); ok {
		errno = fs.ToErrno(syscall.Fchmod(f.fd, mode))
		if errno != 0 {
			return errno
		}
	}

	uid32, uOk := in.GetUID()
	gid32, gOk := in.GetGID()
	if uOk || gOk {
		uid := -1
		gid := -1

		if uOk {
			uid = int(uid32)
		}
		if gOk {
			gid = int(gid32)
		}
		errno = fs.ToErrno(syscall.Fchown(f.fd, uid, gid))
		if errno != 0 {
			return errno
		}
	}

	mtime, mok := in.GetMTime()
	atime, aok := in.GetATime()

	if mok || aok {
		ap := &atime
		mp := &mtime
		if !aok {
			ap = nil
		}
		if !mok {
			mp = nil
		}
		errno = f.utimens(ap, mp)
		if errno != 0 {
			return errno
		}
	}

	if sz, ok := in.GetSize(); ok {
		errno = fs.ToErrno(syscall.Ftruncate(f.fd, int64(sz)))
		if errno != 0 {
			return errno
		}
	}
	return fs.OK
}

func (f *loopbackFile) Getattr(ctx context.Context, a *fuse.AttrOut) syscall.Errno {
	f.mu.Lock()
	defer f.mu.Unlock()
	st := syscall.Stat_t{}
	err := syscall.Fstat(f.fd, &st)
	if err != nil {
		return fs.ToErrno(err)
	}
	a.FromStat(&st)

	return fs.OK
}

func (f *loopbackFile) Lseek(ctx context.Context, off uint64, whence uint32) (uint64, syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, err := unix.Seek(f.fd, int64(off), int(whence))
	return uint64(n), fs.ToErrno(err)
}
