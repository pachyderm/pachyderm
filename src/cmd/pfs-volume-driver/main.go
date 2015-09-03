package main

import (
	"strings"

	"go.pedge.io/dockervolume"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/volume"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"google.golang.org/grpc"
)

const (
	volumeDriverName  = "pfs"
	volumeDriverGroup = "root"
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS"`
	BaseMountpoint     string `env:"BASE_MOUNTPOINT,required"`
}

func main() {
	mainutil.Main(do, &appEnv{}, nil)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	address := appEnv.PachydermPfsd1Port
	if address == "" {
		address = appEnv.PfsAddress
	} else {
		address = strings.Replace(address, "tcp://", "", -1)
	}
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	return dockervolume.Serve(
		dockervolume.NewVolumeDriverHandler(
			volume.NewVolumeDriver(
				fuse.NewMounter(
					pfs.NewApiClient(
						clientConn,
					),
				),
				appEnv.BaseMountpoint,
			),
			dockervolume.VolumeDriverHandlerOptions{},
		),
		dockervolume.ProtocolUnix,
		volumeDriverName,
		volumeDriverGroup,
	)
}
