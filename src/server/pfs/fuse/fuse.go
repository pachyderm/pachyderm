package fuse

import (
	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
)

// A Mounter mounts fuse filesystems.
type Mounter interface {
	// Mount mounts a repository available as a fuse filesystem at mountPoint.
	// Mount blocks and will return once the volume is unmounted.
	MountAndCreate(
		mountPoint string,
		shard *pfsclient.Shard,
		commitMounts []*CommitMount, // nil means mount all commits
		ready chan bool,
		debug bool,
		allCommits bool,
	) error

	Mount(
		mountPoint string,
		shard *pfsclient.Shard,
		commitMounts []*CommitMount, // nil means mount all commits
		ready chan bool,
		debug bool,
		allCommits bool,
	) error
	// Unmount unmounts a mounted filesystem (duh).
	// There's nothing special about this unmount, it's just doing a syscall under the hood.
	Unmount(mountPoint string) error
}

// NewMounter creates a new Mounter.
// Address can be left blank, it's used only for aesthetic purposes.
func NewMounter(address string, apiClient *client.APIClient) Mounter {
	return newMounter(address, apiClient)
}
