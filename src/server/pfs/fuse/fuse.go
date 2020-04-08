package fuse

import (
	"io/ioutil"
	"os"
	"os/signal"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/pkg/errors"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

const (
	modeFile uint32 = fuse.S_IFREG | 0444 // everyone can read, no one can do anything else
	modeDir  uint32 = fuse.S_IFDIR | 0555 // everyone can read and execute, no one can do anything else (execute permission is required to list a dir)
)

type file struct {
	pfs  *pfs.File
	path string
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

// Mount pfs to target, opts may be left nil.
func Mount(c *client.APIClient, target string, opts *Options) error {
	files := make(map[string]*file)
	commits := opts.getCommits()
	fuseTarget := target
	if opts.Write {
		var err error
		fuseTarget, err = ioutil.TempDir("", "pfs-fuse-lower")
		if err != nil {
			return err
		}
	}
	server, err := fs.Mount(fuseTarget, &node{
		file: client.NewFile("", "", ""),
		m: &mount{
			c:       c,
			commits: commits,
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
	var overlayErr error
	if opts.Write {
		upperdir, err := ioutil.TempDir("", "pfs-fuse-upper")
		if err != nil {
			return err
		}
		workdir, err := ioutil.TempDir("", "pfs-fuse-work")
		if err != nil {
			return err
		}
		go func() {
			if err := overlay(fuseTarget, upperdir, workdir, target); err != nil {
				overlayErr = errors.Wrap(err, "error creating overlay mount")
				server.Unmount()
			}
		}()
	}
	server.Serve()
	var eg errgroup.Group
	for _, file := range files {
		eg.Go(func() error {
			return os.Remove(file.path)
		})
	}
	return eg.Wait()
}
