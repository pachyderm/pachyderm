package fuse

import (
	"bytes"
	"context"
	"io"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type file struct {
	fs     *filesystem
	name   string
	c      *client.APIClient
	cancel func()
	// r is a reader for the file that we keep open, this allows us to optimize
	// the common case of reading the entire file from beginning to end.
	// Without r we would have to open a new connection every time someone
	// wanted to read a chunk of the file.
	r      io.Reader
	offset int64
	m      map[int64]chan struct{}
	mu     sync.Mutex
}

func newFile(fs *filesystem, name string) *file {
	ctx, cancel := context.WithCancel(fs.c.Ctx())
	c := fs.c.WithCtx(ctx)
	return &file{
		fs:     fs,
		name:   name,
		c:      c,
		cancel: cancel,
		m:      make(map[int64]chan struct{}),
	}
}

func (f *file) Write(data []byte, off int64) (written uint32, code fuse.Status) {
	return 0, fuse.EROFS
}

func (f *file) SetInode(*nodefs.Inode) {}

func (f *file) String() string {
	return f.name
}

func (f *file) InnerFile() nodefs.File {
	return nil
}

func (f *file) Read(dest []byte, offset int64) (fuse.ReadResult, fuse.Status) {
	repo, file, err := f.fs.parsePath(f.name)
	if err != nil {
		return nil, toStatus(err)
	}
	switch {
	case repo != nil:
		return nil, fuse.Status(syscall.EISDIR)
	case file != nil:
		return f.newReadResult(file, offset, len(dest)), fuse.OK
	default:
		return nil, fuse.Status(syscall.EISDIR)
	}
}

func (f *file) Flock(flags int) fuse.Status {
	return fuse.ENOSYS
}

func (f *file) Flush() fuse.Status {
	// For reasons I don't understand Flush gets called when reading files.
	return fuse.OK
}

func (f *file) Release() {}

func (f *file) Fsync(flags int) (code fuse.Status) {
	return fuse.EROFS
}

func (f *file) Truncate(size uint64) fuse.Status {
	return fuse.EROFS
}

func (f *file) GetAttr(out *fuse.Attr) fuse.Status {
	attr, status := f.fs.getAttr(f.name)
	*out = *attr
	return status
}

func (f *file) Chown(uid uint32, gid uint32) fuse.Status {
	return fuse.EROFS
}

func (f *file) Chmod(perms uint32) fuse.Status {
	return fuse.EROFS
}

func (f *file) Utimens(atime *time.Time, mtime *time.Time) fuse.Status {
	return fuse.EROFS
}

func (f *file) Allocate(off uint64, size uint64, mode uint32) fuse.Status {
	return fuse.EROFS
}

type readResult struct {
	f      *file
	file   *pfs.File
	offset int64
	size   int
}

func (f *file) newReadResult(file *pfs.File, offset int64, size int) *readResult {
	return &readResult{
		f:      f,
		file:   file,
		offset: offset,
		size:   size,
	}
}

// readFull is like read but always fills up the buffer
func readFull(r io.Reader, p []byte) error {
	n := 0
	for n < len(p) {
		_n, err := r.Read(p[n:])
		if err != nil {
			return err
		}
		n += _n
	}
	return nil
}

func (r *readResult) Bytes(buf []byte) ([]byte, fuse.Status) {
	size := r.size
	if len(buf) < size {
		size = len(buf)
	}
	if size < len(buf) {
		buf = buf[0:size]
	}

	res, err := func() ([]byte, error) {
		r.f.mu.Lock()
		defer r.f.mu.Unlock()
		if r.f.r == nil {
			reader, err := r.f.c.GetFileReader(r.file.Commit.Repo.Name, r.file.Commit.ID, r.file.Path, r.offset, 0)
			if err != nil {
				return nil, err
			}
			r.f.r = reader
			r.f.offset = r.offset
		}
		if r.offset > r.f.offset {
			ch := make(chan struct{})
			r.f.m[r.offset] = ch
			r.f.mu.Unlock()
			select {
			case <-ch:
			case <-time.After(time.Second):
			}
			r.f.mu.Lock()
			delete(r.f.m, r.offset)
		}
		if r.f.offset == r.offset {
			if err := readFull(r.f.r, buf); err != nil {
				r.f.r = nil
				if err == io.EOF {
					return nil, nil
				}
				return nil, err
			}
			r.f.offset += int64(size)
			if ch, ok := r.f.m[r.f.offset]; ok {
				close(ch)
			}
			return buf, nil
		}
		return nil, nil
	}()
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	if res != nil {
		return res, fuse.OK
	}

	buffer := bytes.NewBuffer(buf[:0])
	if err := r.f.c.GetFile(r.file.Commit.Repo.Name, r.file.Commit.ID, r.file.Path, r.offset, int64(size), buffer); err != nil {
		return nil, fuse.ToStatus(err)
	}
	return buffer.Bytes(), fuse.OK
}

func (r *readResult) Size() int {
	return r.size
}

func (r *readResult) Done() {}
