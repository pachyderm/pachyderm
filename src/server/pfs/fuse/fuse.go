package fuse

import (
	"io/ioutil"
	"os"
	"os/signal"
	pathpkg "path"
	"path/filepath"
	"strings"

	"github.com/hanwen/go-fuse/v2/fs"

	"github.com/pachyderm/pachyderm/src/client"
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
	// branches := opts.getBranches()
	fuseTarget := target
	if opts.Write {
		var err error
		fuseTarget, err = ioutil.TempDir("", "pfs-fuse-lower")
		if err != nil {
			return err
		}
	}
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
		server.Unmount()
	}()
	server.Serve()
	pfc, err := c.NewPutFileClient()
	if err != nil {
		return err
	}
	defer func() {
		if err := pfc.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	for path, state := range root.files {
		if state != dirty {
			continue
		}
		f, err := os.Open(filepath.Join(root.rootPath, path))
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		parts := strings.Split(path, "/")
		if _, err := pfc.PutFileOverwrite(parts[0], root.branch(parts[0]),
			pathpkg.Join(parts[1:]...), f, 0); err != nil {
			return err
		}
	}
	return nil
}
