package volume

import (
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"go.pedge.io/dockervolume"
)

func NewVolumeDriver(mounter fuse.Mounter) dockervolume.VolumeDriver {
	return newVolumeDriver(mounter)
}
