package fuse

import (
	"bytes"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type file struct {
	c    *client.APIClient
	name string
}

func newFile(c *client.APIClient, name string) *file {
	return &file{
		c:    c,
		name: name,
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
	repo, file := parsePath(f.name)
	switch {
	case repo != nil:
		return nil, fuse.Status(syscall.EISDIR)
	case file != nil:
		return newReadResult(f.c, file, offset, len(dest)), fuse.OK
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
	attr, status := getAttr(f.c, f.name)
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
	c      *client.APIClient
	file   *pfs.File
	offset int64
	size   int
}

func newReadResult(c *client.APIClient, file *pfs.File, offset int64, size int) *readResult {
	return &readResult{
		c:      c,
		file:   file,
		offset: offset,
		size:   size,
	}
}

func (r *readResult) Bytes(buf []byte) (_ []byte, status fuse.Status) {
	size := r.size
	if len(buf) < size {
		size = len(buf)
	}

	buffer := bytes.NewBuffer(buf[:0])
	if err := r.c.GetFile(r.file.Commit.Repo.Name, r.file.Commit.ID, r.file.Path, r.offset, int64(size), buffer); err != nil {
		return nil, fuse.ToStatus(err)
	}
	return buffer.Bytes(), fuse.OK
}

func (r *readResult) Size() int {
	return r.size
}

func (r *readResult) Done() {}
