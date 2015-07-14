package server

import "github.com/pachyderm/pachyderm/src/pfs"

// NewAPIServer returns a new ApiServer.
func NewAPIServer() pfs.ApiServer {
	return newAPIServer()
}
