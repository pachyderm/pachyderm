package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	"go.pedge.io/pkg/time"
	"go.pedge.io/proto/test"
	"go.pedge.io/protolog"

	"github.com/fsouza/go-dockerclient"
	"github.com/satori/go.uuid"
	"go.pachyderm.com/pachyderm/src/pfs"
	pfstesting "go.pachyderm.com/pachyderm/src/pfs/testing"
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pkg/require"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/store"
	"google.golang.org/grpc"
)

const (
	testNumServers = 1
)

func TestBasic(t *testing.T) {
	t.Skip()
	t.Parallel()
	runTest(t, testBasic)
}

func testBasic(t *testing.T, apiClient pps.ApiClient) {
	_ = os.RemoveAll("/tmp/pachyderm-test")
	pipelineSource, err := apiClient.CreatePipelineSource(
		context.Background(),
		&pps.CreatePipelineSourceRequest{
			PipelineSource: &pps.PipelineSource{
				TypedPipelineSource: &pps.PipelineSource_GithubPipelineSource{
					GithubPipelineSource: &pps.GithubPipelineSource{
						ContextDir: "src/pps/server/testdata/basic",
						User:       "pachyderm",
						Repository: "pachyderm",
						Branch:     "master",
					},
				},
			},
		},
	)
	require.NoError(t, err)
	pipeline, err := apiClient.CreateAndGetPipeline(
		context.Background(),
		&pps.CreateAndGetPipelineRequest{
			PipelineSourceId: pipelineSource.Id,
		},
	)
	require.NoError(t, err)
	pipelineRun, err := apiClient.CreatePipelineRun(
		context.Background(),
		&pps.CreatePipelineRunRequest{
			PipelineId: pipeline.Id,
		},
	)
	require.NoError(t, err)
	_, err = apiClient.StartPipelineRun(
		context.Background(),
		&pps.StartPipelineRunRequest{
			PipelineRunId: pipelineRun.Id,
		},
	)
	require.NoError(t, err)
	pipelineRunID := pipelineRun.Id
	pipelineRunStatus, err := getFinalPipelineRunStatus(apiClient, pipelineRunID)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS, pipelineRunStatus.PipelineRunStatusType)
	matches, err := filepath.Glob("/tmp/pachyderm-test/1-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pachyderm-test/1-out/1.txt",
			"/tmp/pachyderm-test/1-out/10.txt",
			"/tmp/pachyderm-test/1-out/2.txt",
			"/tmp/pachyderm-test/1-out/20.txt",
			"/tmp/pachyderm-test/1-out/3.txt",
			"/tmp/pachyderm-test/1-out/30.txt",
			"/tmp/pachyderm-test/1-out/4.txt",
			"/tmp/pachyderm-test/1-out/40.txt",
			"/tmp/pachyderm-test/1-out/5.txt",
			"/tmp/pachyderm-test/1-out/50.txt",
		},
		matches,
	)
	matches, err = filepath.Glob("/tmp/pachyderm-test/2-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pachyderm-test/2-out/1.txt.copy",
			"/tmp/pachyderm-test/2-out/10.txt.copy",
			"/tmp/pachyderm-test/2-out/2.txt.copy",
			"/tmp/pachyderm-test/2-out/20.txt.copy",
			"/tmp/pachyderm-test/2-out/3.txt.copy",
			"/tmp/pachyderm-test/2-out/30.txt.copy",
			"/tmp/pachyderm-test/2-out/4.txt.copy",
			"/tmp/pachyderm-test/2-out/40.txt.copy",
			"/tmp/pachyderm-test/2-out/5.txt.copy",
			"/tmp/pachyderm-test/2-out/50.txt.copy",
		},
		matches,
	)
	matches, err = filepath.Glob("/tmp/pachyderm-test/3-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pachyderm-test/3-out/1.txt.copy3",
			"/tmp/pachyderm-test/3-out/10.txt.copy3",
			"/tmp/pachyderm-test/3-out/2.txt.copy3",
			"/tmp/pachyderm-test/3-out/20.txt.copy3",
			"/tmp/pachyderm-test/3-out/3.txt.copy3",
			"/tmp/pachyderm-test/3-out/30.txt.copy3",
			"/tmp/pachyderm-test/3-out/4.txt.copy3",
			"/tmp/pachyderm-test/3-out/40.txt.copy3",
			"/tmp/pachyderm-test/3-out/5.txt.copy3",
			"/tmp/pachyderm-test/3-out/50.txt.copy3",
		},
		matches,
	)
	matches, err = filepath.Glob("/tmp/pachyderm-test/4-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pachyderm-test/4-out/1.txt.copy4",
			"/tmp/pachyderm-test/4-out/10.txt.copy4",
			"/tmp/pachyderm-test/4-out/2.txt.copy4",
			"/tmp/pachyderm-test/4-out/20.txt.copy4",
			"/tmp/pachyderm-test/4-out/3.txt.copy4",
			"/tmp/pachyderm-test/4-out/30.txt.copy4",
			"/tmp/pachyderm-test/4-out/4.txt.copy4",
			"/tmp/pachyderm-test/4-out/40.txt.copy4",
			"/tmp/pachyderm-test/4-out/5.txt.copy4",
			"/tmp/pachyderm-test/4-out/50.txt.copy4",
			"/tmp/pachyderm-test/4-out/build-file.txt4",
		},
		matches,
	)
	matches, err = filepath.Glob("/tmp/pachyderm-test/5-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pachyderm-test/5-out/1.txt.copy3",
			"/tmp/pachyderm-test/5-out/1.txt.copy4",
			"/tmp/pachyderm-test/5-out/10.txt.copy3",
			"/tmp/pachyderm-test/5-out/10.txt.copy4",
			"/tmp/pachyderm-test/5-out/2.txt.copy3",
			"/tmp/pachyderm-test/5-out/2.txt.copy4",
			"/tmp/pachyderm-test/5-out/20.txt.copy3",
			"/tmp/pachyderm-test/5-out/20.txt.copy4",
			"/tmp/pachyderm-test/5-out/3.txt.copy3",
			"/tmp/pachyderm-test/5-out/3.txt.copy4",
			"/tmp/pachyderm-test/5-out/30.txt.copy3",
			"/tmp/pachyderm-test/5-out/30.txt.copy4",
			"/tmp/pachyderm-test/5-out/4.txt.copy3",
			"/tmp/pachyderm-test/5-out/4.txt.copy4",
			"/tmp/pachyderm-test/5-out/40.txt.copy3",
			"/tmp/pachyderm-test/5-out/40.txt.copy4",
			"/tmp/pachyderm-test/5-out/5.txt.copy3",
			"/tmp/pachyderm-test/5-out/5.txt.copy4",
			"/tmp/pachyderm-test/5-out/50.txt.copy3",
			"/tmp/pachyderm-test/5-out/50.txt.copy4",
			"/tmp/pachyderm-test/5-out/build-file.txt4",
			"/tmp/pachyderm-test/5-out/build-file2.txt",
		},
		matches,
	)
}

func getFinalPipelineRunStatus(apiClient pps.ApiClient, pipelineRunID string) (*pps.PipelineRunStatus, error) {
	// TODO(pedge): not good
	ticker := time.NewTicker(time.Second)
	for i := 0; i < 20; i++ {
		<-ticker.C
		pipelineRunStatuses, err := apiClient.GetPipelineRunStatus(
			context.Background(),
			&pps.GetPipelineRunStatusRequest{
				PipelineRunId: pipelineRunID,
			},
		)
		if err != nil {
			return nil, err
		}
		pipelineRunStatus := pipelineRunStatuses.PipelineRunStatus[0]
		protolog.Printf("status at tick %d: %v\n", i, pipelineRunStatus)
		switch pipelineRunStatus.PipelineRunStatusType {
		case pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_ERROR:
			return pipelineRunStatus, nil
		case pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS:
			return pipelineRunStatus, nil
		}
	}
	return nil, fmt.Errorf("did not get final pipeline status for %s", pipelineRunID)
}

func runTest(
	t *testing.T,
	f func(t *testing.T, apiClient pps.ApiClient),
) {
	containerClient, err := getContainerClient()
	require.NoError(t, err)
	storeClient, err := getRethinkClient()
	require.NoError(t, err)
	pfstesting.RunTest(
		t,
		func(t *testing.T, apiClient pfs.ApiClient, internalApiClient pfs.InternalApiClient, cluster pfstesting.Cluster) {
			prototest.RunT(
				t,
				testNumServers,
				func(servers map[string]*grpc.Server) {
					for _, server := range servers {
						pps.RegisterApiServer(server, newAPIServer(apiClient, containerClient, storeClient, pkgtime.NewSystemTimer()))
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
						pps.NewApiClient(
							clientConn,
						),
					)
				},
			)
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

func getRethinkClient() (store.Client, error) {
	address, err := getRethinkAddress()
	if err != nil {
		return nil, err
	}
	databaseName := strings.Replace(uuid.NewV4().String(), "-", "", -1)
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
