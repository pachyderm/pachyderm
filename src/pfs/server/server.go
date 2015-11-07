package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"golang.org/x/net/context"
)

type APIServer interface {
	pfs.APIServer
	pfs.ClusterAPIServer
	shard.Frontend
}

type InternalAPIServer interface {
	pfs.InternalAPIServer
	pfs.ReplicaAPIServer
	shard.Server
}

type ReplicaAPIServer interface {
	pfs.ReplicaAPIServer
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

func NewGoogleReplicaAPIServer(ctx context.Context, bucket string) ReplicaAPIServer {
	return newGoogleReplicaAPIServer(ctx, bucket)
}
