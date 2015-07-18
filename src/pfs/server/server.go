package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
)

type CombinedAPIServer interface {
	pfs.ApiServer
	pfs.InternalApiServer
}

// NewCombinedAPIServer returns a new CombinedAPIServer.
func NewCombinedAPIServer(
	sharder route.Sharder,
	router route.Router,
	driver drive.Driver,
) CombinedAPIServer {
	return newCombinedAPIServer(
		sharder,
		router,
		driver,
	)
}
