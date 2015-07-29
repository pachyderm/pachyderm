package fuse

import (
	"github.com/pachyderm/pachyderm/src/pfs"
)

type Mounter interface {
	// Mount mounts makes a repository available as a fuse filesystem at mountPoint
	// If it succeeds Mount will block.
	Mount(apiClient pfs.ApiClient, repositoryName string, mountPoint string, shard uint64, modulus uint64) error
	// Ready blocks until the filesysyem has been mounted
	Ready()
}

func NewMounter() Mounter {
	return newMounter()
}
