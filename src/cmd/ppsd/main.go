package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/server"
	"github.com/pachyderm/pachyderm/src/pps/store"
	"google.golang.org/grpc"
)

var (
	defaultEnv = map[string]string{
		"PPS_API_PORT":      "651",
		"PPS_DATABASE_NAME": "pachyderm",
	}
)

type appEnv struct {
	DockerHost      string `env:"DOCKER_HOST"`
	PfsAddress      string `env:"PFS_ADDRESS"`
	APIPort         int    `env:"PPS_API_PORT"`
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
	apiClient, err := getPfsAPIClient(appEnv.PfsAddress)
	if err != nil {
		return err
	}
	return grpcutil.GrpcDo(
		appEnv.APIPort,
		appEnv.TracePort,
		pachyderm.Version,
		func(s *grpc.Server) {
			pps.RegisterApiServer(s, server.NewAPIServer(apiClient, containerClient, rethinkClient, timing.NewSystemTimer()))
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

func getPfsAPIClient(address string) (pfs.ApiClient, error) {
	var err error
	if address == "" {
		address, err = getPfsAddress()
		if err != nil {
			return nil, err
		}
	}
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pfs.NewApiClient(clientConn), nil
}

func getPfsAddress() (string, error) {
	rethinkAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("PFSD_PORT_650_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:650", rethinkAddr), nil
}
