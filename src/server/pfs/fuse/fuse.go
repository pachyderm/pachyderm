package fuse

import (
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/hanwen/go-fuse/v2/fs"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
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
	return root.sync(true)
}
