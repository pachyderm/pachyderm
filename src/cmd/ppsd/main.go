package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/context"

	"github.com/fsouza/go-dockerclient"

	"go.pedge.io/env"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/server"

	"go.pachyderm.com/pachyderm"
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	persistserver "go.pachyderm.com/pachyderm/src/pps/persist/server"
	"go.pachyderm.com/pachyderm/src/pps/server"
	"go.pachyderm.com/pachyderm/src/pps/watch"
	watchserver "go.pachyderm.com/pachyderm/src/pps/watch/server"
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
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS"`
	Address            string `env:"PPS_ADDRESS"`
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
	containerClient, err := getContainerClient()
	if err != nil {
		return err
	}
	rethinkAPIClient, err := getRethinkAPIClient(appEnv.DatabaseAddress, appEnv.DatabaseName)
	if err != nil {
		return err
	}
	pfsAddress := appEnv.PachydermPfsd1Port
	if pfsAddress == "" {
		pfsAddress = appEnv.PfsAddress
	} else {
		pfsAddress = strings.Replace(pfsAddress, "tcp://", "", -1)
	}
	clientConn, err := grpc.Dial(pfsAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	pfsAPIClient := pfs.NewAPIClient(clientConn)
	watchAPIServer := watchserver.NewAPIServer(pfsAPIClient, rethinkAPIClient)
	watchAPIClient := watch.NewLocalAPIClient(watchAPIServer)
	if _, err := watchAPIClient.Start(context.Background(), &google_protobuf.Empty{}); err != nil {
		return err
	}
	return protoserver.Serve(
		uint16(appEnv.Port),
		func(s *grpc.Server) {
			pps.RegisterAPIServer(s, server.NewAPIServer(rethinkAPIClient, watchAPIClient, containerClient))
		},
		protoserver.ServeOptions{
			DebugPort: uint16(appEnv.DebugPort),
			Version:   pachyderm.Version,
		},
	)
}

func getContainerClient() (container.Client, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return container.NewDockerClient(client), nil
}

func getRethinkAPIClient(address string, databaseName string) (persist.APIClient, error) {
	var err error
	if address == "" {
		address, err = getRethinkAddress()
		if err != nil {
			return nil, err
		}
	}
	if err := persistserver.InitDBs(address, databaseName); err != nil {
		return nil, err
	}
	rethinkAPIServer, err := persistserver.NewRethinkAPIServer(address, databaseName)
	if err != nil {
		return nil, err
	}
	return persist.NewLocalAPIClient(rethinkAPIServer), nil
}

func getRethinkAddress() (string, error) {
	rethinkAddr := os.Getenv("RETHINK_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINK_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
}
