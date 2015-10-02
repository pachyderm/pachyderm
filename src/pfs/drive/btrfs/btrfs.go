package btrfs //import "go.pachyderm.com/pachyderm/src/pfs/driver/btrfs"

import "go.pachyderm.com/pachyderm/src/pfs/drive"

// NewDriver constructs a new Driver for btrfs.
func NewDriver(rootDir string, namespace string) (drive.Driver, error) {
	return newDriver(rootDir, namespace)
}
