package fuse

import "github.com/pachyderm/pachyderm/v2/src/client"

// A long-running server which dynamically manages mounts in the given directory.

type ServerOptions struct {
	Daemonize            bool
	VersionedMountsDir   string
	UnversionedMountsDir string
	Socket               string
	LogFile              string
}

func Server(c *client.APIClient, opts *ServerOptions) (retErr error) {
	// TODO: respect opts.Daemonize

	return nil
}
