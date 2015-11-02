package fuse

type Mounter interface {
	// Mount mounts a repository available as a fuse filesystem at mountPoint at the commitID.
	// commitID is optional - if not passed, all commits will be mounted.
	// Mount will not block and will return once mounted, or error otherwise.
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

type MounterProvider interface {
	Get() (Mounter, error)
}

func NewMounterProvider(pfsAddress string) MounterProvider {
	return newMounterProvider(pfsAddress)
}
