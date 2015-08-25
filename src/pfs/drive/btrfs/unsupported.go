// +build !linux

package btrfs

import "github.com/pachyderm/pachyderm/src/pfs/drive"

// NewDriver constructs a new Driver for btrfs.
func NewDriver(rootDir string, namespace string) (drive.Driver, error) {
	return nil
}
