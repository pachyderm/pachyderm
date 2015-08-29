package fuse

import (
	"github.com/pachyderm/pachyderm/src/pfs"
)

type Mounter interface {
	// Mount mounts a repository available as a fuse filesystem at mountPoint at the commitID.
	// commitID is optional - if not passed, all commits will be mounted.
	// Mount will not block and will return once mounted, or error otherwise.
	Mount(
		repositoryName string,
		commitID string,
		mountPoint string,
		shard uint64,
		modulus uint64,
	) error
	// Unmount unmounts a mounted filesystem (duh).
	// There's nothing special about this unmount, it's just doing a syscall under the hood.
	Unmount(mountPoint string) error
	// Wait waits for the mountPoint to either have errored or be unmounted.
	Wait(mountPoint string) error
}

func NewMounter(apiClient pfs.ApiClient) Mounter {
	return newMounter(apiClient)
}
