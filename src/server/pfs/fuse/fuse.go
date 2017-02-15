package fuse

import "github.com/pachyderm/pachyderm/src/client"

// A Mounter mounts fuse filesystems.
type Mounter interface {
	// Mount mounts a repository available as a fuse filesystem at mountPoint.
	// Mount blocks and will return once the volume is unmounted.
	MountAndCreate(
		mountPoint string,
		commitMounts []*CommitMount, // nil means mount all commits
		ready chan bool,
		debug bool,
		// if oneMount is true, mount only one CommitMount
		oneMount bool,
	) error

	Mount(
		mountPoint string,
		commitMounts []*CommitMount, // nil means mount all commits
		ready chan bool,
		debug bool,
		oneMount bool,
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
