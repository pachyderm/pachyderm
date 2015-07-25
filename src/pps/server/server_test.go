package server

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/pkg/executil"
	"github.com/pachyderm/pachyderm/src/pkg/grpctest"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/ppsutil"
	"github.com/pachyderm/pachyderm/src/pps/store"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

const (
	testNumServers = 1
)

func init() {
	// TODO(pedge): needed in tests? will not be needed for golang 1.5 for sure
	runtime.GOMAXPROCS(runtime.NumCPU())
	executil.SetDebug(true)
}

func TestBasic(t *testing.T) {
	runTest(t, testBasic)
}

func testBasic(t *testing.T, apiClient pps.ApiClient) {
	_ = os.RemoveAll("/tmp/pps-server-test")
	startPipelineRunResponse, err := ppsutil.StartPipelineRunGithub(
		apiClient,
		"src/pps/server/testdata/basic",
		"pachyderm",
		"pachyderm",
		"master",
		"",
	)
	require.NoError(t, err)
	pipelineRunID := startPipelineRunResponse.PipelineRunId
	pipelineRunStatus, err := getFinalPipelineRunStatus(apiClient, pipelineRunID)
	require.NoError(t, err)
	require.Equal(t, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS, pipelineRunStatus.PipelineRunStatusType)
	matches, err := filepath.Glob("/tmp/pps-server-test/1-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pps-server-test/1-out/1.txt",
			"/tmp/pps-server-test/1-out/10.txt",
			"/tmp/pps-server-test/1-out/2.txt",
			"/tmp/pps-server-test/1-out/20.txt",
			"/tmp/pps-server-test/1-out/3.txt",
			"/tmp/pps-server-test/1-out/30.txt",
			"/tmp/pps-server-test/1-out/4.txt",
			"/tmp/pps-server-test/1-out/40.txt",
			"/tmp/pps-server-test/1-out/5.txt",
			"/tmp/pps-server-test/1-out/50.txt",
		},
		matches,
	)
	matches, err = filepath.Glob("/tmp/pps-server-test/2-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pps-server-test/2-out/1.txt.copy",
			"/tmp/pps-server-test/2-out/10.txt.copy",
			"/tmp/pps-server-test/2-out/2.txt.copy",
			"/tmp/pps-server-test/2-out/20.txt.copy",
			"/tmp/pps-server-test/2-out/3.txt.copy",
			"/tmp/pps-server-test/2-out/30.txt.copy",
			"/tmp/pps-server-test/2-out/4.txt.copy",
			"/tmp/pps-server-test/2-out/40.txt.copy",
			"/tmp/pps-server-test/2-out/5.txt.copy",
			"/tmp/pps-server-test/2-out/50.txt.copy",
		},
		matches,
	)
	matches, err = filepath.Glob("/tmp/pps-server-test/3-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pps-server-test/3-out/1.txt.copy3",
			"/tmp/pps-server-test/3-out/10.txt.copy3",
			"/tmp/pps-server-test/3-out/2.txt.copy3",
			"/tmp/pps-server-test/3-out/20.txt.copy3",
			"/tmp/pps-server-test/3-out/3.txt.copy3",
			"/tmp/pps-server-test/3-out/30.txt.copy3",
			"/tmp/pps-server-test/3-out/4.txt.copy3",
			"/tmp/pps-server-test/3-out/40.txt.copy3",
			"/tmp/pps-server-test/3-out/5.txt.copy3",
			"/tmp/pps-server-test/3-out/50.txt.copy3",
		},
		matches,
	)
	matches, err = filepath.Glob("/tmp/pps-server-test/4-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pps-server-test/4-out/1.txt.copy4",
			"/tmp/pps-server-test/4-out/10.txt.copy4",
			"/tmp/pps-server-test/4-out/2.txt.copy4",
			"/tmp/pps-server-test/4-out/20.txt.copy4",
			"/tmp/pps-server-test/4-out/3.txt.copy4",
			"/tmp/pps-server-test/4-out/30.txt.copy4",
			"/tmp/pps-server-test/4-out/4.txt.copy4",
			"/tmp/pps-server-test/4-out/40.txt.copy4",
			"/tmp/pps-server-test/4-out/5.txt.copy4",
			"/tmp/pps-server-test/4-out/50.txt.copy4",
		},
		matches,
	)
	matches, err = filepath.Glob("/tmp/pps-server-test/5-out/*")
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"/tmp/pps-server-test/5-out/1.txt.copy3",
			"/tmp/pps-server-test/5-out/10.txt.copy3",
			"/tmp/pps-server-test/5-out/2.txt.copy3",
			"/tmp/pps-server-test/5-out/20.txt.copy3",
			"/tmp/pps-server-test/5-out/3.txt.copy3",
			"/tmp/pps-server-test/5-out/30.txt.copy3",
			"/tmp/pps-server-test/5-out/4.txt.copy3",
			"/tmp/pps-server-test/5-out/40.txt.copy3",
			"/tmp/pps-server-test/5-out/5.txt.copy3",
			"/tmp/pps-server-test/5-out/50.txt.copy3",
			"/tmp/pps-server-test/5-out/1.txt.copy",
			"/tmp/pps-server-test/5-out/10.txt.copy4",
			"/tmp/pps-server-test/5-out/2.txt.copy4",
			"/tmp/pps-server-test/5-out/20.txt.copy4",
			"/tmp/pps-server-test/5-out/3.txt.copy4",
			"/tmp/pps-server-test/5-out/30.txt.copy4",
			"/tmp/pps-server-test/5-out/4.txt.copy4",
			"/tmp/pps-server-test/5-out/40.txt.copy4",
			"/tmp/pps-server-test/5-out/5.txt.copy4",
			"/tmp/pps-server-test/5-out/50.txt.copy4",
		},
		matches,
	)
}

func getFinalPipelineRunStatus(apiClient pps.ApiClient, pipelineRunID string) (*pps.PipelineRunStatus, error) {
	// TODO(pedge): not good
	ticker := time.NewTicker(time.Second)
	for i := 0; i < 15; i++ {
		<-ticker.C
		getPipelineRunStatusResponse, err := ppsutil.GetPipelineRunStatus(
			apiClient,
			pipelineRunID,
		)
		if err != nil {
			return nil, err
		}
		fmt.Println(getPipelineRunStatusResponse.PipelineRunStatus)
		pipelineRunStatus := getPipelineRunStatusResponse.PipelineRunStatus
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
	storeClient := store.NewInMemoryClient()
	grpctest.Run(
		t,
		testNumServers,
		func(servers map[string]*grpc.Server) {
			for _, server := range servers {
				pps.RegisterApiServer(server, newAPIServer(storeClient))
			}
		},
		func(t *testing.T, clientConns map[string]*grpc.ClientConn) {
			var clientConn *grpc.ClientConn
			for _, c := range clientConns {
				clientConn = c
				break
			}
			for _, c := range clientConns {
				if c != clientConn {
					_ = c.Close()
				}
			}
			f(
				t,
				pps.NewApiClient(
					clientConn,
				),
			)
		},
	)
}
