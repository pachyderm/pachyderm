package fuse

import (
	"github.com/hanwen/go-fuse/v2/fs"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// Options is for configuring fuse mounts. Any of the fields may be left nil
// and `nil` itself is a valid set of Options which uses the default for
// everything.
type Options struct {
	Fuse *fs.Options

	// Write indicates that the pfs mount should allow writes.
	// Writes will be written back to the filesystem.
	Write bool

	// RepoOptions is a map from mount names to options associated with them.
	RepoOptions map[string]*RepoOptions

	// Unmount is a channel that will be closed when the filesystem has been
	// unmounted. It can be nil in which case it's ignored.
	Unmount chan struct{}
}

// RepoOptions are the options associated with a mounted repo.
type RepoOptions struct {
	// Name is the name _of the mount_. This is needed because the mount might
	// have a different name to the repo, to support mounting multiple versions
	// of the same repo at the same time.
	Name string

	File     *pfs.File
	Subpaths []string
	// Repo is the name of the repo to mount
	// Repo string
	// Branch is the branch of the repo to mount
	// Branch string
	// Commit is a specific commit on the branch to mount
	// Commit string
	// Write indicates that the repo should be mounted for writing.
	Write bool
}

func (o *Options) getFuse() *fs.Options {
	if o == nil || o.Fuse == nil {
		// We always return a struct here because otherwise the defaults that
		// fuse sets make files inaccessible.
		return &fs.Options{}
	}
	return o.Fuse
}

func (o *Options) getRepoOpts() map[string]*RepoOptions {
	if o == nil {
		return make(map[string]*RepoOptions)
	}
	return o.RepoOptions
}

func (o *Options) getBranches() map[string]string {
	result := make(map[string]string)
	if o == nil {
		return result
	}
	for name, opts := range o.RepoOptions {
		if opts.File.Commit.Branch.Name != "" {
			result[name] = opts.File.Commit.Branch.Name
		}
	}
	return result
}

func (o *Options) getWrite() bool {
	if o == nil {
		return false
	}
	return o.Write
}

func (o *Options) getUnmount() chan struct{} {
	if o == nil {
		return nil
	}
	return o.Unmount
}

func (o *Options) validate(c *client.APIClient) error {
	if o == nil {
		return nil
	}
	for _, opts := range o.RepoOptions {
		if opts.Write {
			if uuid.IsUUIDWithoutDashes(opts.File.Commit.Branch.Name) {
				return errors.Errorf("can't mount commit %s@%s as %s in Write mode (mount a branch instead)", opts.File.Commit.Branch.Repo.Name, opts.File.Commit.Branch.Name, opts.Name)
			}
			bi, err := c.InspectBranch(opts.File.Commit.Branch.Repo.Name, opts.File.Commit.Branch.Name)
			if err != nil && !errutil.IsNotFoundError(err) {
				return err
			}
			if bi != nil && len(bi.Provenance) > 0 {
				return errors.Errorf("can't mount branch %s@%s as %s in Write mode because it's an output branch", opts.File.Commit.Branch.Repo.Name, opts.File.Commit.Branch.Name, opts.Name)
			}
		}
	}
	return nil
}
