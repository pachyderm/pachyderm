package main

import (
	"net/http"
	"strings"

	"go.pedge.io/dockervolume"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/volume"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"google.golang.org/grpc"
)

const (
	volumeDriverName = "pfs"
)

var (
	defaultEnv = map[string]string{
		"PFS_ADDRESS": "0.0.0.0:650",
	}
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	Address            string `env:"PFS_ADDRESS"`
}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)

	address := appEnv.PachydermPfsd1Port
	if address == "" {
		address = appEnv.Address
	} else {
		address = strings.Replace(address, "tcp://", "", -1)
	}
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}

	volumeDriverHandler := dockervolume.NewVolumeDriverHandler(
		volume.NewVolumeDriver(
			fuse.NewMounter(
				pfs.NewApiClient(
					clientConn,
				),
			),
		),
	)
	start := make(chan struct{})
	listener, err := dockervolume.NewUnixListener(
		volumeDriverName,
		"root",
		start,
	)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:    volumeDriverName,
		Handler: volumeDriverHandler,
	}
	close(start)
	return server.Serve(listener)
}
