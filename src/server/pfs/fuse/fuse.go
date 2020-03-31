package fuse

import (
	"os"
	"os/signal"
	"sync"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	pfssync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
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
func Mount(c *client.APIClient, mountPoint string, opts *Options) (retErr error) {
	files := make(map[string]*file)
	commits := opts.getCommits()
	server, err := fs.Mount(mountPoint, &node{
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
	server.Serve()
	pfcs := make(map[string]client.PutFileClient)
	for _, file := range files {
		if file.dirty {
			var pfc client.PutFileClient
			if _pfc, ok := pfcs[file.pfs.Commit.Repo.Name]; ok {
				pfc = _pfc
			} else {
				pfc, err = c.NewPutFileClient()
				if err != nil {
					return err
				}
				defer func() {
					if err := pfc.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				pfcs[file.pfs.Commit.Repo.Name] = pfc
			}
			if err := func() error {
				f, err := os.Open(file.path)
				if err != nil {
					return err
				}
				if err := os.Remove(file.path); err != nil {
					return err
				}
				defer func() {
					if err := f.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				branch := "master"
				if commit, ok := commits[file.pfs.Commit.Repo.Name]; ok && !uuid.IsUUIDWithoutDashes(commit) {
					branch = commit
				}
				file.pfs.Commit.ID = branch
				return pfssync.PushFile(c, pfc, file.pfs, f)
			}(); err != nil {
				return err
			}
		}
	}
	return nil
}
