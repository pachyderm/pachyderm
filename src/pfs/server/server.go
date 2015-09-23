package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
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

type APIServer interface {
	pfs.ApiServer
}

type InternalAPIServer interface {
	pfs.InternalApiServer
	route.Server
}

// NewAPIServer returns a new APIServer.
func NewAPIServer(
	sharder route.Sharder,
	router route.Router,
) APIServer {
	return newAPIServer(
		sharder,
		router,
	)
}

// NewInternalAPIServer returns a new InternalAPIServer.
func NewInternalAPIServer(
	sharder route.Sharder,
	router route.Router,
	driver drive.Driver,
) InternalAPIServer {
	return newInternalAPIServer(
		sharder,
		router,
		driver,
	)
}
