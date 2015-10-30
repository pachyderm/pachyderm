package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.pedge.io/env"
	"go.pedge.io/proto/server"
	"google.golang.org/grpc"

	"github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/container"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/jobserver"
	"github.com/pachyderm/pachyderm/src/pps/jobserver/run"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	persistserver "github.com/pachyderm/pachyderm/src/pps/persist/server"
	"github.com/pachyderm/pachyderm/src/pps/pipelineserver"
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
	PfsMountDir        string `env:"PFS_MOUNT_DIR"`
	Address            string `env:"PPS_ADDRESS"`
	Port               int    `env:"PPS_PORT"`
	DatabaseAddress    string `env:"PPS_DATABASE_ADDRESS"`
	DatabaseName       string `env:"PPS_DATABASE_NAME"`
	DebugPort          int    `env:"PPS_TRACE_PORT"`
	RemoveContainers   bool   `env:"PPS_REMOVE_CONTAINERS"`
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
	jobAPIServer := jobserver.NewAPIServer(
		pfsAPIClient,
		rethinkAPIClient,
		containerClient,
		appEnv.PfsMountDir,
		jobserverrun.JobRunnerOptions{
			RemoveContainers: appEnv.RemoveContainers,
		},
	)
	jobAPIClient := pps.NewLocalJobAPIClient(jobAPIServer)
	pipelineAPIServer := pipelineserver.NewAPIServer(pfsAPIClient, jobAPIClient, rethinkAPIClient)
	errC := make(chan error)
	go func() {
		errC <- protoserver.Serve(
			uint16(appEnv.Port),
			func(s *grpc.Server) {
				pps.RegisterJobAPIServer(s, jobAPIServer)
				pps.RegisterPipelineAPIServer(s, pipelineAPIServer)
			},
			protoserver.ServeOptions{
				DebugPort: uint16(appEnv.DebugPort),
				Version:   pachyderm.Version,
			},
		)
	}()
	// TODO: pretty sure without this there is a problem, this is bad, we would
	// prefer a callback for when the server is ready to accept requests
	time.Sleep(1 * time.Second)
	if err := pipelineAPIServer.Start(); err != nil {
		return err
	}
	return <-errC
}

func getContainerClient() (container.Client, error) {
	// TODO: this will just connect to the local instance of docker
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
