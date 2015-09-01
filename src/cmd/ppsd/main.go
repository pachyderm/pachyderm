package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/ppsutil"
	"github.com/pachyderm/pachyderm/src/pps/server"
	"github.com/pachyderm/pachyderm/src/pps/store"
	"google.golang.org/grpc"
)

var (
	defaultEnv = map[string]string{
		"PPS_ADDRESS":       "0.0.0.0",
		"PPS_PORT":          "651",
		"PPS_DATABASE_NAME": "pachyderm",
	}
)

type appEnv struct {
	DockerHost      string `env:"DOCKER_HOST"`
	Address         string `env:"PPS_ADDRESS"`
	Port            int    `env:"PPS_PORT"`
	DatabaseAddress string `env:"PPS_DATABASE_ADDRESS"`
	DatabaseName    string `env:"PPS_DATABASE_NAME"`
	TracePort       int    `env:"PPS_TRACE_PORT"`
}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	containerClient, err := getContainerClient(appEnv.DockerHost)
	if err != nil {
		return err
	}
	rethinkClient, err := getRethinkClient(appEnv.DatabaseAddress, appEnv.DatabaseName)
	if err != nil {
		return err
	}
	discoveryClient, err := getEtcdClient()
	if err != nil {
		return err
	}
	address := fmt.Sprintf("%s:%d", appEnv.Address, appEnv.Port)
	ppsutil.NewPpsRegistry(discoveryClient).RegisterAddress(address)
	provider := pfsutil.NewPfsProvider(discoveryClient, grpcutil.NewDialer(grpc.WithInsecure()))
	clientConn, err := provider.GetClientConn()
	if err != nil {
		return err
	}
	return grpcutil.GrpcDo(
		appEnv.Port,
		appEnv.TracePort,
		pachyderm.Version,
		func(s *grpc.Server) {
			pps.RegisterApiServer(s, server.NewAPIServer(pfs.NewApiClient(clientConn), containerClient, rethinkClient, timing.NewSystemTimer()))
		},
	)
}

func getContainerClient(dockerHost string) (container.Client, error) {
	if dockerHost == "" {
		dockerHost = "unix:///var/run/docker.sock"
	}
	return container.NewDockerClient(
		container.DockerClientOptions{
			Host: dockerHost,
		},
	)
}

func getRethinkClient(address string, databaseName string) (store.Client, error) {
	var err error
	if address == "" {
		address, err = getRethinkAddress()
		if err != nil {
			return nil, err
		}
	}
	if err := store.InitDBs(address, databaseName); err != nil {
		return nil, err
	}
	return store.NewRethinkClient(address, databaseName)
}

func getRethinkAddress() (string, error) {
	rethinkAddr := os.Getenv("RETHINK_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINK_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
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
