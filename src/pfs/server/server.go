package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/role"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"go.pedge.io/google-protobuf"
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
	emptyInstance = &google_protobuf.Empty{}
)

type ApiServer interface {
	pfs.ApiServer
	role.Server
}

type InternalApiServer interface {
	pfs.InternalApiServer
	role.Server
}

// NewApiServer returns a new ApiServer.
func NewApiServer(
	sharder route.Sharder,
	router route.Router,
	driver drive.Driver,
) ApiServer {
	return newApiServer(
		sharder,
		router,
		driver,
	)
}

// NewApiServer returns a new ApiServer.
func NewInternalApiServer(
	sharder route.Sharder,
	router route.Router,
	driver drive.Driver,
) InternalApiServer {
	return newInternalApiServer(
		sharder,
		router,
		driver,
	)
}
