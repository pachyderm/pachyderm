package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/trace"

	"go.pedge.io/env"
	"go.pedge.io/proto/server"
	"go.pedge.io/protolog/logrus"

	"go.pachyderm.com/pachyderm"
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pkg/timing"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/container"
	"go.pachyderm.com/pachyderm/src/pps/server"
	"go.pachyderm.com/pachyderm/src/pps/store"
	"google.golang.org/grpc"
)

var (
	defaultEnv = map[string]string{
		"PPS_ADDRESS":       "0.0.0.0",
		"PPS_PORT":          "651",
		"PPS_TRACE_PORT":    "1051",
		"PPS_DATABASE_NAME": "pachyderm",
	}
)

type appEnv struct {
	DockerHost         string `env:"DOCKER_HOST"`
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS"`
	PfsdPort           string `env:"PFSD_PORT_650_TCP"`
	Port               int    `env:"PPS_PORT"`
	DatabaseAddress    string `env:"PPS_DATABASE_ADDRESS"`
	DatabaseName       string `env:"PPS_DATABASE_NAME"`
	DebugPort          int    `env:"PPS_TRACE_PORT"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	logrus.Register()
	containerClient, err := getContainerClient(appEnv.DockerHost)
	if err != nil {
		return err
	}
	rethinkClient, err := getRethinkClient(appEnv.DatabaseAddress, appEnv.DatabaseName)
	if err != nil {
		return err
	}
	var pfsAddress string
	pfsAddress = appEnv.PachydermPfsd1Port
	if pfsAddress == "" {
		pfsAddress = appEnv.PfsAddress
	} else {
		pfsAddress = strings.Replace(pfsAddress, "tcp://", "", -1)
	}
	if pfsAddress == "" {
		pfsAddress = strings.Replace(appEnv.PfsdPort, "tcp://", "", -1)
	}
	clientConn, err := grpc.Dial(pfsAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	// TODO(pedge): no!
	trace.AuthRequest = func(_ *http.Request) (bool, bool) {
		return true, true
	}
	return protoserver.Serve(
		uint16(appEnv.Port),
		func(s *grpc.Server) {
			pps.RegisterApiServer(s, server.NewAPIServer(pfs.NewApiClient(clientConn), containerClient, rethinkClient, timing.NewSystemTimer()))
		},
		protoserver.ServeOptions{
			DebugPort: uint16(appEnv.DebugPort),
			Version:   pachyderm.Version,
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
	rethinkAddr := os.Getenv("RETHINKDB_DRIVER_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINKDB_DRIVER_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
}
