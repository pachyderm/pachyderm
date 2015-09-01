package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.pedge.io/dockervolume"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pfs/volume"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"google.golang.org/grpc"
)

const (
	volumeDriverName = "pfs"
)

type appEnv struct {
	EtcdAddress string `env:"ETCD_ADDRESS,required"`
}

func main() {
	mainutil.Main(do, &appEnv{}, nil)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)

	discoveryClient, err := getEtcdClient(appEnv.EtcdAddress)
	if err != nil {
		return err
	}
	provider := pfsutil.NewPfsProvider(discoveryClient, grpcutil.NewDialer(grpc.WithInsecure()))
	clientConn, err := provider.GetClientConn()
	if err != nil {
		return err
	}
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

func getEtcdClient(address string) (discovery.Client, error) {
	etcdAddress, err := getEtcdAddress(address)
	if err != nil {
		return nil, err
	}
	return discovery.NewEtcdClient(etcdAddress), nil
}

func getEtcdAddress(address string) (string, error) {
	if address != "" {
		return address, nil
	}
	etcdAddr := os.Getenv("ETCD_PORT_2379_TCP_ADDR")
	if etcdAddr == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("http://%s:2379", etcdAddr), nil
}
