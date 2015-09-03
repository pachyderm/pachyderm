package volume

import (
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"go.pedge.io/dockervolume"
)

func NewVolumeDriver(
	mounterProvider func() (fuse.Mounter, error),
	baseMountpoint string,
) dockervolume.VolumeDriver {
	return newVolumeDriver(
		mounterProvider,
		baseMountpoint,
	)
}
