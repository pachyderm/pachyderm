package obj

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
)

// NewDriver constructs a new Driver for obj.
func NewDriver(blockClient pfs.BlockAPIClient) (drive.Driver, error) {
	return newDriver(blockClient)
}
