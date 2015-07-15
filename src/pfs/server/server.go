package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/shard"
)

// NewAPIServer returns a new ApiServer.
func NewAPIServer(
	sharder shard.Sharder,
	router route.Router,
	driver drive.Driver,
) pfs.ApiServer {
	return newAPIServer(
		sharder,
		router,
		driver,
	)
}
