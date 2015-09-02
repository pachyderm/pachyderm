// +build linux

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"github.com/pachyderm/pachyderm/src/pkg/netutil"
	"google.golang.org/grpc"
)

var (
	defaultEnv = map[string]string{
		"PFS_NUM_SHARDS":  "16",
		"PFS_PORT":        "650",
		"PFS_DRIVER_TYPE": "btrfs",
	}
)

type appEnv struct {
	DriverRoot string `env:"PFS_DRIVER_ROOT,required"`
	DriverType string `env:"PFS_DRIVER_TYPE"`
	NumShards  int    `env:"PFS_NUM_SHARDS"`
	Address    string `env:"PFS_ADDRESS"`
	Port       int    `env:"PFS_PORT"`
	TracePort  int    `env:"PFS_TRACE_PORT"`
}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	discoveryClient, err := getEtcdClient()
	if err != nil {
		return err
	}
	address := appEnv.Address
	if address == "" {
		address, err = netutil.ExternalIP()
		if err != nil {
			return err
		}
	}
	address = fmt.Sprintf("%s:%d", address, appEnv.Port)
	pfsutil.NewPfsRegistry(discoveryClient).RegisterAddress(address)
	addresser := route.NewDiscoveryAddresser(
		discoveryClient,
		"namespace",
	)
	for i := 0; i < appEnv.NumShards; i++ {
		if _, err := addresser.SetMasterAddress(i, address); err != nil {
			return err
		}
	}
	var driver drive.Driver
	switch appEnv.DriverType {
	case "btrfs":
		driver, err = btrfs.NewDriver(appEnv.DriverRoot, "")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown value for PFS_DRIVER_TYPE: %s", appEnv.DriverType)
	}
	combinedAPIServer := server.NewCombinedAPIServer(
		route.NewSharder(
			appEnv.NumShards,
		),
		route.NewRouter(
			addresser,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
		driver,
	)
	return grpcutil.GrpcDo(
		appEnv.Port,
		appEnv.TracePort,
		pachyderm.Version,
		func(s *grpc.Server) {
			pfs.RegisterApiServer(s, combinedAPIServer)
			pfs.RegisterInternalApiServer(s, combinedAPIServer)
		},
	)
}

func getEtcdClient() (discovery.Client, error) {
	etcdAddress, err := getEtcdAddress()
	if err != nil {
		return nil, err
	}
	return discovery.NewEtcdClient(etcdAddress), nil
}

func getEtcdAddress() (string, error) {
	etcdAddr := os.Getenv("ETCD_PORT_2379_TCP_ADDR")
	if etcdAddr == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("http://%s:2379", etcdAddr), nil
}
