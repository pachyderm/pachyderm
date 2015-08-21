package role

import (
	"github.com/pachyderm/pachyderm/src/pfs/route"
)

// Roler is responsible for managing which roles the server fills
type Roler interface {
	Run() error
	Cancel() error
}

type Server interface {
	// Master tells the server that it is now the master for shard.
	// After this returns the Server is expected to service Master requests for shard.
	Master(shard int) error
	// Replica tells the server that it is now a replica for shard.
	// After this returns the Server is expected to service Replica requests for shard.
	Replica(shard int) error
	// Clear tells the server that it is no longer filling any role for shard.
	Clear(shard int) error
}

func NewRoler(addresser route.Addresser, sharder route.Sharder, server Server, localAddress string) Roler {
	return newRoler(addresser, sharder, server, localAddress)
}
