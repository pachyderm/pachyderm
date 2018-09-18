package fuse

import (
	"context"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type file struct {
	name    string
	attr    *fuse.Attr
	cancel  func()
	pfsFile *pfs.File
	file    *os.File
	counter *counter
	err     error
}

func newFile(fs *filesystem, name string) (*file, fuse.Status) {
	attr, status := fs.getAttr(name)
	if status != fuse.OK {
		return nil, status
	}
	ctx, cancel := context.WithCancel(fs.c.Ctx())
	c := fs.c.WithCtx(ctx)
	f, err := ioutil.TempFile("", "pfs-fuse")
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		return nil, fuse.ToStatus(err)
	}
	_, pfsFile, err := fs.parsePath(name)
	if err != nil {
		return nil, toStatus(err)
	}
	if pfsFile == nil {
		return nil, fuse.Status(syscall.EISDIR)
	}
	counter := newCounter()
	// Argument order is important here because it means that writes to w must
	// complete writing to f before being written to counter. Thus counter can
	// tell us conclusively at least (but not at most) a certain number of
	// bytes has been written to f.
	w := io.MultiWriter(f, counter)
	result := &file{
		attr:    attr,
		cancel:  cancel,
		pfsFile: pfsFile,
		file:    f,
		counter: counter,
	}
	go func() {
		if err := c.GetFile(pfsFile.Commit.Repo.Name, pfsFile.Commit.ID, pfsFile.Path, 0, 0, w); err != nil {
			result.err = err
			counter.cancel()
		}
	}()
	return result, fuse.OK
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
	waitn := offset + int64(len(dest))
	if waitn > int64(f.attr.Size) {
		waitn = int64(f.attr.Size)
	}
	f.counter.wait(waitn)
	// check if there was an error reading the file
	if f.err != nil {
		return nil, toStatus(f.err)
	}
	if err := f.file.Sync(); err != nil {
		return nil, toStatus(err)
	}
	return fuse.ReadResultFd(f.file.Fd(), offset, len(dest)), fuse.OK
}

func (f *file) Flock(flags int) fuse.Status {
	return fuse.ENOSYS
}

func (f *file) Flush() fuse.Status {
	// For reasons I don't understand Flush gets called when reading files.
	return fuse.OK
}

func (f *file) Release() {
	f.cancel()
}

func (f *file) Fsync(flags int) (code fuse.Status) {
	return fuse.EROFS
}

func (f *file) Truncate(size uint64) fuse.Status {
	return fuse.EROFS
}

func (f *file) GetAttr(out *fuse.Attr) fuse.Status {
	*out = *f.attr
	return fuse.OK
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

type counter struct {
	n    int64
	mu   sync.Mutex
	cond *sync.Cond
}

func newCounter() *counter {
	result := &counter{}
	result.cond = sync.NewCond(&result.mu)
	return result
}

func (c *counter) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n += int64(len(p))
	c.cond.Broadcast()
	return len(p), nil
}

// wait until more than n bytes have been written
func (c *counter) wait(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for c.n < n {
		c.cond.Wait()
	}
}

// cancel indicates that an error has occurred which will prevent any further
// calls to Write, it causes all calls to wait() to return
func (c *counter) cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n = math.MaxInt64
}
