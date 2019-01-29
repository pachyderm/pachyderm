package fuse

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type file struct {
	name    string
	pfsFile *pfs.File
	ready   chan struct{}
}

func newFile(fs *filesystem, name string, flags int) (_ nodefs.File, retStatus fuse.Status) {
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
		file, err := os.OpenFile(f.name, flags, modeFile)
		if err != nil {
			return nil, toStatus(err)
		}
		return nodefs.NewLoopbackFile(file), fuse.OK
	}
	tmpF, err := ioutil.TempFile("", "pfs-fuse")
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	if err := fs.c.GetFile(pfsFile.Commit.Repo.Name, pfsFile.Commit.ID, pfsFile.Path, 0, 0, tmpF); err != nil {
		if !strings.Contains(err.Error(), "not found") || flags&os.O_CREATE == 0 {
			return nil, toStatus(err)
		}
	}
	f.name = tmpF.Name()
	close(f.ready)
	return nodefs.NewLoopbackFile(tmpF), fuse.OK
}

func (f *file) String() string {
	return fileString(f.pfsFile)
}

func fileString(f *pfs.File) string {
	return filepath.Join(f.Commit.Repo.Name, f.Path)
}
