package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/role"
	"github.com/pachyderm/pachyderm/src/pfs/route"
)

const (
	// InitialCommitID is the initial id before any commits are made in a repository.
	InitialCommitID = "scratch"
)

var (
	// ReservedCommitIDs are the commit ids used by the system.
	ReservedCommitIDs = map[string]bool{
		InitialCommitID: true,
	}
)

type CombinedAPIServer interface {
	pfs.ApiServer
	pfs.InternalApiServer
	role.Server
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
