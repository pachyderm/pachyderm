package fuse

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

const (
	modeFile uint32 = fuse.S_IFREG | 0444 // everyone can read, no one can do anything else
	modeDir  uint32 = fuse.S_IFDIR | 0555 // everyone can read and execute, no one can do anything else (execute permission is required to list a dir)
)

type file struct {
	pfs   *pfs.File
	path  string
	dirty bool
}

type mount struct {
	c       *client.APIClient
	commits map[string]string
	files   map[string]*file
	mu      sync.Mutex
}

func (m *mount) commit(repo string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if commit, ok := m.commits[repo]; ok {
		return commit, nil
	}
	bi, err := m.c.InspectBranch(repo, "master")
	if errutil.IsNotFoundError(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if bi.Head == nil {
		m.commits[repo] = ""
		return "", nil
	}
	m.commits[repo] = bi.Head.ID
	return bi.Head.ID, nil
}

// Mount pfs to mountPoint, opts may be left nil.
func Mount(c *client.APIClient, mountPoint string, opts *Options) error {
	files := make(map[string]*file)
	server, err := fs.Mount(mountPoint, &node{
		file: client.NewFile("", "", ""),
		m: &mount{
			c:       c,
			commits: opts.getCommits(),
			files:   files,
		},
	}, opts.getFuse())
	if err != nil {
		return err
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		select {
		case <-sigChan:
		case <-opts.getUnmount():
		}
		server.Unmount()
	}()
	server.Serve()
	for _, file := range files {
		fmt.Printf("%+v\n", file)
	}
	return nil
}

type node struct {
	fs.Inode
	file *pfs.File
	m    *mount
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

func (n *node) Open(ctx context.Context, openFlags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	key := fileKey(n.file)
	path := ""
	func() {
		n.m.mu.Lock()
		defer n.m.mu.Unlock()
		if file, ok := n.m.files[key]; ok {
			path = file.path
			file.dirty = file.dirty || allowsWrite(openFlags)
		}
	}()
	var f *os.File
	if path != "" {
		var err error
		f, err = os.OpenFile(path, int(openFlags), 0755)
		if err != nil {
			return nil, 0, toErrno(err)
		}
	}
	if f == nil {
		var err error
		f, err = ioutil.TempFile("", "pfs-fuse")
		if err != nil {
			return nil, 0, toErrno(err)
		}
		if err := n.m.c.GetFile(n.file.Commit.Repo.Name, n.file.Commit.ID, n.file.Path, 0, 0, f); err != nil {
			return nil, 0, toErrno(err)
		}
		if _, err := f.Seek(0, 0); err != nil {
			return nil, 0, toErrno(err)
		}
		n.m.mu.Lock()
		defer n.m.mu.Unlock()
		n.m.files[fileKey(n.file)] = &file{
			pfs:   n.file,
			path:  f.Name(),
			dirty: allowsWrite(openFlags),
		}
	}
	return fs.NewLoopbackFile(int(f.Fd())), 0, 0
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
