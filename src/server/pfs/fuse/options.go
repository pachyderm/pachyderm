package fuse

import (
	"github.com/hanwen/go-fuse/v2/fs"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

// Options is for configuring fuse mounts. Any of the fields may be left nil
// and `nil` itself is a valid set of Options which uses the default for
// everything.
type Options struct {
	Fuse *fs.Options

	// Write indicates that the pfs mount should allow writes.
	// Writes will be written back to the filesystem.
	Write bool

	// RepoOptions is a map from repo names to
	RepoOptions map[string]*RepoOptions

	// Unmount is a channel that will be closed when the filesystem has been
	// unmounted. It can be nil in which case it's ignored.
	Unmount chan struct{}
}

type RepoOptions struct {
	Branch string
	Write  bool
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
	for repo, opts := range o.RepoOptions {
		if opts.Branch != "" {
			result[repo] = opts.Branch
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
	for repo, opts := range o.RepoOptions {
		if opts.Write {
			if uuid.IsUUIDWithoutDashes(opts.Branch) {
				return errors.Errorf("can't mount commit %s@%s in Write mode (mount a branch instead)", repo, opts.Branch)
			}
			bi, err := c.InspectBranch(repo, opts.Branch)
			if err != nil && !errutil.IsNotFoundError(err) {
				return err
			}
			if bi != nil && len(bi.Provenance) > 0 {
				return errors.Errorf("can't mount branch %s@%s in Write mode because it's an output branch")
			}
		}
	}
	return nil
}
