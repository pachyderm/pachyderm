package main

import (
	"strings"

	"go.pedge.io/dockervolume"
	"go.pedge.io/protolog/logrus"

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

var (
	defaultEnv = map[string]string{
		"PFS_ADDRESS":     "0.0.0.0:650",
		"BASE_MOUNTPOINT": "/tmp/pfs-volume-driver",
	}
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS"`
	BaseMountpoint     string `env:"BASE_MOUNTPOINT"`
}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	logrus.Register()
	address := appEnv.PachydermPfsd1Port
	if address == "" {
		address = appEnv.PfsAddress
	} else {
		address = strings.Replace(address, "tcp://", "", -1)
	}
	return dockervolume.Serve(
		dockervolume.NewVolumeDriverHandler(
			volume.NewVolumeDriver(
				func() (fuse.Mounter, error) {
					clientConn, err := grpc.Dial(address, grpc.WithInsecure())
					if err != nil {
						return nil, err
					}
					return fuse.NewMounter(
						pfs.NewApiClient(
							clientConn,
						),
					), nil
				},
				appEnv.BaseMountpoint,
			),
			dockervolume.VolumeDriverHandlerOptions{},
		),
		dockervolume.ProtocolUnix,
		volumeDriverName,
		volumeDriverGroup,
	)
}
