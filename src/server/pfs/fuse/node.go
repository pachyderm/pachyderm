package fuse

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

type node struct {
	fs.Inode
	file *pfs.File
	m    *mount
}

func (n *node) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	out.FromStatfsT(&syscall.Statfs_t{})
	return 0
}

func (n *node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	var result staticDirEntries
	switch {
	case n.file.Commit.Repo.Name == "":
		ris, err := n.m.c.ListRepo()
		if err != nil {
			return nil, toErrno(err)
		}
		for _, ri := range ris {
			result = append(result, fuse.DirEntry{
				Mode: modeDir,
				Name: ri.Repo.Name,
			})
		}
	case n.file.Commit.ID == "":
		// headless branch, so we want to just return an empty result
	default:
		fis, err := n.m.c.ListFile(n.file.Commit.Repo.Name, n.file.Commit.ID, n.file.Path)
		if err != nil {
			return nil, toErrno(err)
		}
		for _, fi := range fis {
			mode := modeDir
			if fi.FileType == pfs.FileType_FILE {
				mode = modeFile
			}
			result = append(result, fuse.DirEntry{
				Name: path.Base(fi.File.Path),
				Mode: mode,
			})
		}
	}
	return &result, 0
}

func (n *node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	switch {
	case n.file.Commit.Repo.Name == "":
		ri, err := n.m.c.InspectRepo(name)
		if err != nil {
			return nil, toErrno(err)
		}
		commit, err := n.m.commit(name)
		if err != nil {
			return nil, toErrno(err)
		}
		return n.NewInode(ctx, &node{
			file: client.NewFile(ri.Repo.Name, commit, ""),
			m:    n.m,
		}, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	case n.file.Commit.ID == "":
		// headless branch, so we want to just return ENOENT
		return nil, syscall.ENOENT
	default:
		fi, err := n.m.c.InspectFile(n.file.Commit.Repo.Name, n.file.Commit.ID, path.Join(n.file.Path, name))
		if err != nil {
			return nil, toErrno(err)
		}
		var mode uint32 = syscall.S_IFDIR
		if fi.FileType == pfs.FileType_FILE {
			mode = syscall.S_IFREG
		}
		return n.NewInode(ctx, &node{
			file: client.NewFile(n.file.Commit.Repo.Name, n.file.Commit.ID, fi.File.Path),
			m:    n.m,
		}, fs.StableAttr{Mode: mode}), 0
	}
}

func allowsWrite(flags uint32) bool {
	return (int(flags) & (os.O_WRONLY | os.O_RDWR)) != 0
}

func (n *node) downloadFile() (_ string, retErr error) {
	f, err := ioutil.TempFile("", "pfs-fuse")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return f.Name(), n.m.c.GetFile(n.file.Commit.Repo.Name, n.file.Commit.ID, n.file.Path, 0, 0, f)
}

func (n *node) Open(ctx context.Context, openFlags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	if allowsWrite(openFlags) {
		return nil, 0, syscall.EROFS
	}
	key := fileKey(n.file)
	path := ""
	func() {
		n.m.mu.Lock()
		defer n.m.mu.Unlock()
		if file, ok := n.m.files[key]; ok {
			path = file.path
		}
	}()
	if path == "" {
		var err error
		path, err = n.downloadFile()
		if err != nil {
			return nil, 0, toErrno(err)
		}
		n.m.mu.Lock()
		defer n.m.mu.Unlock()
		// Check again if someone else added this file to the map while we were
		// downloading.
		if file, ok := n.m.files[key]; ok {
			// Someone did, so we remove our copy of the file and use theirs.
			os.Remove(path)
			path = file.path
		}
		n.m.files[fileKey(n.file)] = &file{
			pfs:  n.file,
			path: path,
		}
	}
	fd, err := syscall.Open(path, int(openFlags), 0)
	if err != nil {
		return nil, 0, toErrno(err)
	}
	return fs.NewLoopbackFile(fd), 0, 0
}

func toErrno(err error) syscall.Errno {
	if errutil.IsNotFoundError(err) {
		return syscall.ENOENT
	}
	return syscall.EIO
}

type staticDirEntries []fuse.DirEntry

func (d *staticDirEntries) HasNext() bool {
	return len(*d) > 0
}

func (d *staticDirEntries) Next() (fuse.DirEntry, syscall.Errno) {
	result := (*d)[0]
	*d = (*d)[1:]
	return result, 0
}

func (d *staticDirEntries) Close() {}

func fileKey(f *pfs.File) string {
	return fmt.Sprintf("%s@%s:%s", f.Commit.Repo.Name, f.Commit.ID, f.Path)
}
