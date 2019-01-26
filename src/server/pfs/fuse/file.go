package fuse

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type file struct {
	file    *os.File
	pfsFile *pfs.File
	ready   chan struct{}
}

func newFile(fs *filesystem, name string) (nodefs.File, fuse.Status) {
	_, pfsFile, err := fs.parsePath(name)
	if err != nil {
		return nil, toStatus(err)
	}
	if pfsFile == nil {
		return nil, fuse.Status(syscall.EISDIR)
	}
	f := &file{
		ready:   make(chan struct{}),
		pfsFile: pfsFile,
	}
	resIf, ok := fs.files.LoadOrStore(f.String(), f)
	if ok {
		f = resIf.(*file)
		<-f.ready
		dup, err := dupFile(f.file)
		if err != nil {
			return nil, fuse.ToStatus(err)
		}
		return nodefs.NewLoopbackFile(dup), fuse.OK
	}
	tmpF, err := ioutil.TempFile("", "pfs-fuse")
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	if err := os.Remove(tmpF.Name()); err != nil {
		return nil, fuse.ToStatus(err)
	}
	if err := fs.c.GetFile(pfsFile.Commit.Repo.Name, pfsFile.Commit.ID, pfsFile.Path, 0, 0, tmpF); err != nil {
		return nil, toStatus(err)
	}
	f.file = tmpF
	close(f.ready)
	dup, err := dupFile(f.file)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	return nodefs.NewLoopbackFile(dup), fuse.OK
}

func (f *file) String() string {
	return filepath.Join(f.pfsFile.Commit.Repo.Name, f.pfsFile.Path)
}

func dupFile(f *os.File) (*os.File, error) {
	fd, err := syscall.Dup(int(f.Fd()))
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), f.Name()), nil
}
