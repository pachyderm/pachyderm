package obj

import (
	"github.com/pachyderm/pachyderm/src/pfs/drive"
)

// NewDriver constructs a new Driver for obj.
func NewDriver(driveClient drive.APIClient) (drive.Driver, error) {
	return newDriver(driveClient)
}
