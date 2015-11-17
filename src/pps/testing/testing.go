package testing

import (
	"io/ioutil"
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	pfstesting "github.com/pachyderm/pachyderm/src/pfs/testing"
	"github.com/pachyderm/pachyderm/src/pkg/container"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/jobserver"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	persistservertesting "github.com/pachyderm/pachyderm/src/pps/persist/server/testing"
	"github.com/pachyderm/pachyderm/src/pps/pipelineserver"
	"go.pedge.io/proto/test"
	"google.golang.org/grpc"
)

const (
	testNumServers = 1
)

func RunTest(
	t *testing.T,
	f func(*testing.T, pfs.APIClient, pps.JobAPIClient, pps.PipelineAPIClient),
) {
	pfstesting.RunTest(
		t,
		func(t *testing.T, pfsAPIClient pfs.APIClient, _ pfstesting.Cluster) {
			pfsMountDir, err := ioutil.TempDir("", "")
			require.NoError(t, err)
			mounter := fuse.NewMounter("localhost", pfsAPIClient)
			ready := make(chan bool)
			go func() {
				require.NoError(t, mounter.Mount(pfsMountDir, &pfs.Shard{0, 1}, nil, ready))
			}()
			<-ready
			persistservertesting.RunTestWithRethinkAPIServer(
				t,
				func(t *testing.T, persistAPIServer persist.APIServer) {
					persistAPIClient := persist.NewLocalAPIClient(persistAPIServer)
					var pipelineAPIServer pipelineserver.APIServer
					prototest.RunT(
						t,
						testNumServers,
						func(servers map[string]*grpc.Server) {
							jobAPIServer := jobserver.NewAPIServer(
								pfsAPIClient,
								persistAPIClient,
								nil,
							)
							jobAPIClient := pps.NewLocalJobAPIClient(jobAPIServer)
							pipelineAPIServer = pipelineserver.NewAPIServer(
								pfsAPIClient,
								jobAPIClient,
								persistAPIClient,
							)
							for _, server := range servers {
								pps.RegisterJobAPIServer(server, jobAPIServer)
								pps.RegisterPipelineAPIServer(server, pipelineAPIServer)
							}
						},
						func(t *testing.T, clientConns map[string]*grpc.ClientConn) {
							pipelineAPIServer.Start()
							var clientConn *grpc.ClientConn
							for _, c := range clientConns {
								clientConn = c
								break
							}
							f(
								t,
								pfsAPIClient,
								pps.NewJobAPIClient(
									clientConn,
								),
								pps.NewPipelineAPIClient(
									clientConn,
								),
							)
						},
						func() {
							_ = mounter.Unmount(pfsMountDir)
						},
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
