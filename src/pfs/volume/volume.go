package volume

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"go.pedge.io/dockervolume"
)

func NewVolumeDriver(apiClient pfs.ApiClient) dockervolume.VolumeDriver {
	return newVolumeDriver(apiClient)
}
