package fuse

import "github.com/hanwen/go-fuse/fuse/nodefs"

type Options struct {
	Fuse *nodefs.Options
	// commits is a map from repos to commits, if a repo is unspecified then
	// the master commit of the repo at the time the repo is first requested
	// will be used.
	Commits map[string]string

	Unmount chan struct{}
}

func (o *Options) GetFuse() *nodefs.Options {
	if o == nil {
		return nil
	}
	return o.Fuse
}

func (o *Options) GetCommits() map[string]string {
	if o == nil {
		return nil
	}
	return o.Commits
}

func (o *Options) GetUnmount() chan struct{} {
	if o == nil {
		return nil
	}
	return o.Unmount
}
