package fuse

import (
	"github.com/hanwen/go-fuse/v2/fs"
)

// Options is for configuring fuse mounts. Any of the fields may be left nil
// and `nil` itself is a valid set of Options which uses the default for
// everything.
type Options struct {
	Fuse *fs.Options

	// Write indicates that the pfs mount should allow writes.
	// Writes will be written back to the filesystem.
	Write bool

	// commits is a map from repos to commits, if a repo is unspecified then
	// the master commit of the repo at the time the repo is first requested
	// will be used.
	Commits map[string]string

	// Unmount is a channel that will be closed when the filesystem has been
	// unmounted. It can be nil in which case it's ignored.
	Unmount chan struct{}
}

func (o *Options) getFuse() *fs.Options {
	if o == nil || o.Fuse == nil {
		// We always return a struct here because otherwise the defaults that
		// fuse sets make files inaccessible.
		return &fs.Options{}
	}
	return o.Fuse
}

func (o *Options) getCommits() map[string]string {
	if o == nil || o.Commits == nil {
		return make(map[string]string)
	}
	return o.Commits
}

func (o *Options) getUnmount() chan struct{} {
	if o == nil {
		return nil
	}
	return o.Unmount
}
