package obj

import (
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pkg/obj"
)

// NewDriver constructs a new Driver for btrfs.
func NewDriver(objClient obj.Client, cacheDir string, namespace string) (drive.Driver, error) {
	return newDriver(objClient, cacheDir, namespace)
}
