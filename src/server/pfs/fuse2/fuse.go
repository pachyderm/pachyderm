package fuse

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

const (
	modeFile = fuse.S_IFREG | 0444 // everyone can read, no one can do anything else
	modeDir  = fuse.S_IFDIR | 0555 // everyone can read and execute, no one can do anything else (execute permission is required to list a dir)
)

// Mount pfs to mountPoint, opts may be left nil.
func Mount(c *client.APIClient, mountPoint string, opts *Options) error {
	root.c = c
	server, err := fs.Mount(mountPoint, root, opts.getFuse())
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
	return nil
}

type node struct {
	fs.Inode
	file *pfs.File
	c    *client.APIClient
}

var root = &node{
	file: client.NewFile("", "", ""),
}

func (n *node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	if n.file.Commit.Repo.Name == "" {
		ris, err := n.c.ListRepo()
		if err != nil {
			return nil, toErrno(err)
		}
		var result staticDirs
		for _, ri := range ris {
			result = append(result, fuse.DirEntry{
				Mode: modeDir,
				Name: ri.Repo.Name,
			})
		}
		return &result, 0
	}
	return nil, 0
}

func (n *node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if n.file.Commit.Repo.Name == "" {
		ri, err := n.c.InspectRepo(name)
		if err != nil {
			return nil, toErrno(err)
		}
		return n.NewInode(ctx, &node{
			file: client.NewFile(ri.Repo.Name, "", ""),
			c:    n.c,
		}, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	return nil, 0
}

func toErrno(err error) syscall.Errno {
	if strings.Contains(err.Error(), "not found") {
		return syscall.ENOENT
	}
	return syscall.EIO
}

type staticDirs []fuse.DirEntry

func (d *staticDirs) HasNext() bool {
	return len(*d) > 0
}

func (d *staticDirs) Next() (fuse.DirEntry, syscall.Errno) {
	result := (*d)[0]
	*d = (*d)[1:]
	return result, 0
}

func (d *staticDirs) Close() {}
