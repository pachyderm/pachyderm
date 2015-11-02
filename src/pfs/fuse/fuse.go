package fuse

type Mounter interface {
	// Mount mounts a repository available as a fuse filesystem at mountPoint.
	// Mount blocks and will return once the volume is unmounted.
	Mount(
		mountPoint string,
		shard uint64,
		modulus uint64,
	) error
	// Unmount unmounts a mounted filesystem (duh).
	// There's nothing special about this unmount, it's just doing a syscall under the hood.
	Unmount(mountPoint string) error
}

func NewMounter(address string) (Mounter, error) {
	return newMounter(address)
}
