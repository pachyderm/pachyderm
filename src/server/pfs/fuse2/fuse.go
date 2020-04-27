package fuse

import (
	"io/ioutil"
	"os"
	"os/signal"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/hanwen/go-fuse/v2/fs"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

// Mount pfs to target, opts may be left nil.
func Mount(c *client.APIClient, target string, opts *Options) (retErr error) {
	if err := opts.validate(); err != nil {
		return err
	}
	commits := make(map[string]string)
	for repo, branch := range opts.getBranches() {
		if uuid.IsUUIDWithoutDashes(branch) {
			commits[repo] = branch
		}
	}
	files := make(map[string]*file)
	// branches := opts.getBranches()
	fuseTarget := target
	if opts.Write {
		var err error
		fuseTarget, err = ioutil.TempDir("", "pfs-fuse-lower")
		if err != nil {
			return err
		}
	}
	// mount := &mount{
	// 	c:        c,
	// 	branches: branches,
	// 	commits:  commits,
	// 	files:    files,
	// }
	rootDir, err := ioutil.TempDir("", "pfs")
	if err != nil {
		return err
	}
	root, err := NewLoopbackRoot(rootDir, c, opts)
	if err != nil {
		return err
	}
	server, err := fs.Mount(fuseTarget, root, opts.getFuse())
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
		// if opts.Write {
		// 	if err := unmount(target); err != nil {
		// 		fmt.Printf("error unmounting: %v\n", err)
		// 	}
		// }
		server.Unmount()
	}()
	// var overlayErr error
	// var upperdir string
	// if opts.Write {
	// 	upperdir, err = ioutil.TempDir("", "pfs-fuse-upper")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	workdir, err := ioutil.TempDir("", "pfs-fuse-work")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	go func() {
	// 		if err := overlay(fuseTarget, upperdir, workdir, target); err != nil {
	// 			overlayErr = errors.Wrap(err, "error creating overlay mount")
	// 			server.Unmount()
	// 		}
	// 	}()
	// }
	server.Serve()
	// if overlayErr != nil {
	// 	return overlayErr
	// }
	var eg errgroup.Group
	for _, file := range files {
		eg.Go(func() error {
			return os.Remove(file.path)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
	// pfc, err := c.NewPutFileClient()
	// if err != nil {
	// 	return err
	// }
	// defer func() {
	// 	if err := pfc.Close(); err != nil && retErr == nil {
	// 		retErr = err
	// 	}
	// }()
	// return filepath.Walk(upperdir, func(path string, info os.FileInfo, err error) (retErr error) {
	// 	if info.IsDir() {
	// 		return nil
	// 	}
	// 	f, err := os.Open(path)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer func() {
	// 		if err := f.Close(); err != nil && retErr == nil {
	// 			retErr = err
	// 		}
	// 	}()
	// 	split := strings.Split(strings.TrimPrefix(path, upperdir+"/"), "/")
	// 	repo := split[0]
	// 	file := filepath.Join(split[1:]...)
	// 	if _, err := pfc.PutFileOverwrite(repo, mount.branch(repo), file, f, 0); err != nil {
	// 		return err
	// 	}
	// 	return nil
	// })
}

type file struct {
	pfs  *pfs.File
	path string
}

type mount struct {
	c        *client.APIClient
	branches map[string]string
	commits  map[string]string
	files    map[string]*file
	mu       sync.Mutex
}

func (m *mount) branch(repo string) string {
	if branch, ok := m.branches[repo]; ok {
		return branch
	}
	return "master"
}

func (m *mount) commit(repo string) (string, error) {
	if commit, ok := m.commits[repo]; ok {
		return commit, nil
	}
	branch := m.branch(repo)
	bi, err := m.c.InspectBranch(repo, branch)
	if err != nil && !errutil.IsNotFoundError(err) {
		return "", err
	}
	// You can access branches that don't exist, which allows you to create
	// branches through the fuse mount.
	if errutil.IsNotFoundError(err) || bi.Head == nil {
		m.commits[repo] = ""
		return "", nil
	}
	m.commits[repo] = bi.Head.ID
	return bi.Head.ID, nil
}
