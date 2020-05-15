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
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/progress"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

// Mount pfs to target, opts may be left nil.
func Mount(c *client.APIClient, target string, opts *Options) (retErr error) {
	if err := opts.validate(c); err != nil {
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
	defer func() {
		if err := os.RemoveAll(rootDir); err != nil && retErr == nil {
			retErr = errors.WithStack(err)
		}
	}()
	root, err := newLoopbackRoot(rootDir, target, c, opts)
	if err != nil {
		return err
	}
	server, err := fs.Mount(target, root, opts.getFuse())
	if err != nil {
		return errors.WithStack(err)
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
	pfc := func(repo string) (client.PutFileClient, error) {
		if pfc, ok := pfcs[repo]; ok {
			return pfc, nil
		}
		pfc, err := c.NewPutFileClient()
		if err != nil {
			return nil, err
		}
		pfcs[repo] = pfc
		return pfc, nil
	}
	defer func() {
		for _, pfc := range pfcs {
			if err := pfc.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}
	}()
	for path, state := range root.files {
		if state != dirty {
			continue
		}
		parts := strings.Split(path, "/")
		pfc, err := pfc(parts[0])
		if err != nil {
			return err
		}
		if err := func() (retErr error) {
			f, err := progress.Open(filepath.Join(root.rootPath, path))
			if err != nil {
				if os.IsNotExist(err) {
					return pfc.DeleteFile(parts[0], root.branch(parts[0]), pathpkg.Join(parts[1:]...))
				}
				return errors.Wrap(err, "os.Open")
			}
			defer func() {
				if err := f.Close(); err != nil && retErr == nil {
					retErr = errors.Wrap(err, "f.Close")
				}
			}()
			if _, err := pfc.PutFileOverwrite(parts[0], root.branch(parts[0]),
				pathpkg.Join(parts[1:]...), f, 0); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil
}
