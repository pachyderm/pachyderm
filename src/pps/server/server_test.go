package server

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"go.pedge.io/proto/test"

	"github.com/fsouza/go-dockerclient"
	"github.com/satori/go.uuid"
	"go.pachyderm.com/pachyderm/src/pfs"
	pfstesting "go.pachyderm.com/pachyderm/src/pfs/testing"
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pkg/require"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"google.golang.org/grpc"
)

const (
	testNumServers = 1
)

func TestBasic(t *testing.T) {
	runTest(t, testBasic)
}

func testBasic(t *testing.T, apiClient pps.APIClient) {
}

func runTest(
	t *testing.T,
	f func(t *testing.T, apiClient pps.APIClient),
) {
	containerClient, err := getTestContainerClient()
	require.NoError(t, err)
	persistAPIServer, err := getTestRethinkAPIServer()
	require.NoError(t, err)
	persistAPIClient := persist.NewLocalAPIClient(persistAPIServer)
	pfstesting.RunTest(
		t,
		func(t *testing.T, apiClient pfs.ApiClient, internalApiClient pfs.InternalApiClient, cluster pfstesting.Cluster) {
			prototest.RunT(
				t,
				testNumServers,
				func(servers map[string]*grpc.Server) {
					for _, server := range servers {
						pps.RegisterAPIServer(server, NewAPIServer(persistAPIClient, containerClient))
					}
				},
				func(t *testing.T, clientConns map[string]*grpc.ClientConn) {
					var clientConn *grpc.ClientConn
					for _, c := range clientConns {
						clientConn = c
						break
					}
					f(
						t,
						pps.NewAPIClient(
							clientConn,
						),
					)
				},
			)
		},
	)
}

func getTestContainerClient() (container.Client, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return container.NewDockerClient(client), nil
}

func getTestRethinkAPIServer() (persist.APIServer, error) {
	address, err := getTestRethinkAddress()
	if err != nil {
		return nil, err
	}
	databaseName := strings.Replace(uuid.NewV4().String(), "-", "", -1)
	if err := persist.InitDBs(address, databaseName); err != nil {
		return nil, err
	}
	return persist.NewRethinkAPIServer(address, databaseName)
}

func getTestRethinkAddress() (string, error) {
	rethinkAddr := os.Getenv("RETHINK_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINK_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
}
