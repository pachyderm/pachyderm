package fuse

import (
	"io/ioutil"
	"os"
	"os/signal"
	pathpkg "path"
	"path/filepath"
	"strings"

	"github.com/hanwen/go-fuse/v2/fs"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
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
	rootDir, err := ioutil.TempDir("", "pfs")
	if err != nil {
		return errors.WithStack(err)
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
	mfcs := make(map[string]*client.ModifyFileClient)
	mfc := func(repo string) (*client.ModifyFileClient, error) {
		if mfc, ok := mfcs[repo]; ok {
			return mfc, nil
		}
		mfc, err := c.NewModifyFileClient(repo, root.branch(repo))
		if err != nil {
			return nil, err
		}
		mfcs[repo] = mfc
		return mfc, nil
	}
	defer func() {
		for _, mfc := range mfcs {
			if err := mfc.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}
	}()
	for path, state := range root.files {
		if state != dirty {
			continue
		}
		parts := strings.Split(path, "/")
		mfc, err := mfc(parts[0])
		if err != nil {
			return err
		}
		if err := func() (retErr error) {
			f, err := progress.Open(filepath.Join(root.rootPath, path))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return mfc.DeleteFile(pathpkg.Join(parts[1:]...))
				}
				return errors.WithStack(err)
			}
			defer func() {
				if err := f.Close(); err != nil && retErr == nil {
					retErr = errors.WithStack(err)
				}
			}()
			return mfc.PutFile(pathpkg.Join(parts[1:]...), f)
		}(); err != nil {
			return err
		}
	}
	return nil
}
