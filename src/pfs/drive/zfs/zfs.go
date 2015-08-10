// +build linux

package zfs

import "github.com/pachyderm/pachyderm/src/pfs/drive"

// NewDriver constructs a new Driver for zfs.
func NewDriver(rootDir string) drive.Driver {
	return newDriver(rootDir)
}
