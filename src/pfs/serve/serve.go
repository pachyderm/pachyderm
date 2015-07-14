package serve

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/route"
)

// NewAPIServer returns a new ApiServer.
func NewAPIServer(router route.Router) pfs.ApiServer {
	return newAPIServer(router)
}
