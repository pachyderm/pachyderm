package fuse

import "github.com/hanwen/go-fuse/v2/fs"

// Options is for configuring fuse mounts. Any of the fields may be left nil
// and `nil` itself is a valid set of Options which uses the default for
// everything.
type Options struct {
	Fuse *fs.Options
	// commits is a map from repos to commits, if a repo is unspecified then
	// the master commit of the repo at the time the repo is first requested
	// will be used.
	Commits map[string]string

	Unmount chan struct{}
}

func (o *Options) getFuse() *fs.Options {
	if o == nil {
		return nil
	}
	return o.Fuse
}

func (o *Options) getCommits() map[string]string {
	if o == nil {
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
