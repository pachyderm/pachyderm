package fuse

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	pathpkg "path"
	"path/filepath"
	"strings"

	"github.com/hanwen/go-fuse/v2/fs"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
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
	rootDir, err := ioutil.TempDir("", "pfs")
	if err != nil {
		return err
	}
	fmt.Println("Root: ", rootDir)
	root, err := NewLoopbackRoot(rootDir, c, opts)
	if err != nil {
		return err
	}
	server, err := fs.Mount(target, root, opts.getFuse())
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
	//TODO this errors if it's empty
	pfc, err := c.NewPutFileClient()
	if err != nil {
		return errors.Wrapf(err, "NewPutFileClient")
	}
	defer func() {
		if err := pfc.Close(); err != nil && retErr == nil {
			retErr = errors.Wrapf(err, "pfc.Close")
		}
	}()
	for path, state := range root.files {
		fmt.Println(path, state >= dirty)
		if state < dirty {
			continue
		}
		fmt.Printf("Open(%s)\n", filepath.Join(root.rootPath, path))
		f, err := os.Open(filepath.Join(root.rootPath, path))
		if err != nil {
			return errors.Wrapf(err, "open")
		}
		defer func() {
			if err := f.Close(); err != nil && retErr == nil {
				retErr = errors.Wrapf(err, "close")
			}
		}()
		parts := strings.Split(path, "/")
		if _, err := pfc.PutFileOverwrite(parts[0], root.branch(parts[0]),
			pathpkg.Join(parts[1:]...), f, 0); err != nil {
			return errors.Wrapf(err, "PutFile")
		}
	}
	return nil
}
