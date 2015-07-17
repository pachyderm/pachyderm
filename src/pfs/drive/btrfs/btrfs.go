package btrfs

import "github.com/pachyderm/pachyderm/src/pfs/drive"

func NewDriver(rootDir string) drive.Driver {
	return newDriver(rootDir)
}
