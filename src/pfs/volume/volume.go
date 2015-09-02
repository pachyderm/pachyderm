package volume

import (
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"go.pedge.io/dockervolume"
)

func NewVolumeDriver(mounter fuse.Mounter, baseMountpoint string) dockervolume.VolumeDriver {
	return newVolumeDriver(mounter, baseMountpoint)
}
