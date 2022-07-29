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

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
)

// Mount pfs to target, opts may be left nil.
func Mount(c *client.APIClient, target string, opts *Options) (retErr error) {
	if opts.RepoOptions == nil {
		opts.RepoOptions = make(map[string]*RepoOptions)
	}
	if len(opts.RepoOptions) == 0 {
		// Behavior of `pachctl mount` with no args is to mount everything. Make
		// that explicit here before we pass in the configuration to avoid
		// needing to special-case this deep within the FUSE implementation.
		// (`pachctl mount-server` does _not_ have the same behavior. It mounts
		// nothing to begin with.)
		ris, err := c.ListRepo()
		if err != nil {
			return err
		}
		for _, ri := range ris {
			// Behavior here is that we explicitly mount repos to mounts named
			// by the repo name. This is different to `pachctl mount-server`
			// which supports mounting different versions of the same repo at
			// different named paths.
			branch := "master"
			bi, err := c.InspectBranch(ri.Repo.Name, branch)
			if err != nil && !errutil.IsNotFoundError(err) {
				return err
			}
			isOutputBranch := bi != nil && len(bi.Provenance) > 0
			write := opts.Write
			if isOutputBranch {
				write = false
			}
			opts.RepoOptions[ri.Repo.Name] = &RepoOptions{
				// mount name is same as repo name, i.e. mount it at a directory
				// named the same as the repo itself
				Name:  ri.Repo.Name,
				File:  client.NewFile(ri.Repo.Name, branch, "", ""),
				Write: write,
			}
		}
	}
	if err := opts.validate(c); err != nil {
		return err
	}
	commits := make(map[string]string)
	if opts != nil {
		for repo, ropts := range opts.RepoOptions {
			if ropts.File.Commit.ID != "" && ropts.File.Commit.Branch.Name == "" {
				commits[repo] = ropts.File.Commit.ID
				cis, err := c.InspectCommitSet(ropts.File.Commit.ID)
				if err != nil {
					return err
				}
				branch := ""
				for _, ci := range cis {
					if ci.Commit.Branch.Repo.Name == repo {
						if branch != "" {
							return errors.Errorf("multiple branches (%s and %s) have commit %s, specify a branch", branch, ci.Commit.Branch.Name, ropts.File.Commit.ID)
						}
						branch = ci.Commit.Branch.Name
					}
				}
				ropts.File.Commit.Branch.Name = branch
			}
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
		server.Unmount() //nolint:errcheck
	}()
	server.Wait()
	mfcs := make(map[string]*client.ModifyFileClient)
	mfc := func(repo string) (*client.ModifyFileClient, error) {
		if mfc, ok := mfcs[repo]; ok {
			return mfc, nil
		}
		mfc, err := c.NewModifyFileClient(client.NewCommit(repo, root.branch(repo), ""))
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
	fmt.Println("Uploading files to Pachyderm...")
	// Rendering progress bars for thousands of files significantly slows down
	// throughput. Disabling progress bars takes throughput from 1MB/sec to
	// 200MB/sec on my system, when uploading 18K small files.
	progress.Disable()
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
	fmt.Println("Done!")
	return nil
}
