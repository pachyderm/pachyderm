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
	"github.com/pachyderm/pachyderm/src/pps/jobserver/run"
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
	containerClient, err := getTestContainerClient()
	require.NoError(t, err)
	pfstesting.RunTest(
		t,
		func(t *testing.T, pfsAddress string, pfsAPIClient pfs.APIClient, _ pfstesting.Cluster) {
			pfsMountDir, err := ioutil.TempDir("", "")
			require.NoError(t, err)
			mounter, err := fuse.NewMounter(pfsAddress)
			require.NoError(t, err)
			require.NoError(t, mounter.Mount(pfsMountDir, 0, 1))
			persistservertesting.RunTestWithRethinkAPIServer(
				t,
				func(t *testing.T, persistAPIServer persist.APIServer) {
					persistAPIClient := persist.NewLocalAPIClient(persistAPIServer)
					prototest.RunT(
						t,
						testNumServers,
						func(servers map[string]*grpc.Server) {
							jobAPIServer := jobserver.NewAPIServer(
								pfsAPIClient,
								persistAPIClient,
								containerClient,
								pfsMountDir,
								jobserverrun.JobRunnerOptions{
									RemoveContainers: true,
								},
							)
							jobAPIClient := pps.NewLocalJobAPIClient(jobAPIServer)
							pipelineAPIServer := pipelineserver.NewAPIServer(
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
						//func() {
						//_ = mounter.Unmount(pfsMountDir)
						//},
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
