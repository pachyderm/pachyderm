package server

import (
	"os"
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
	t.Skip()
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

	var pipelineRunStatus pps.PipelineRunStatus
	// TODO(pedge): not good
	ticker := time.NewTicker(time.Second)
	for i := 0; i < 15; i++ {
		<-ticker.C
		getPipelineRunStatusResponse, err := ppsutil.GetPipelineRunStatus(
			apiClient,
			pipelineRunID,
		)
		require.NoError(t, err)
		pipelineRunStatus := getPipelineRunStatusResponse.PipelineRunStatus
		switch pipelineRunStatus.PipelineRunStatusType {
		case pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_ERROR:
			fallthrough
		case pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS:
			break
		default:
			pipelineRunStatus = nil
		}
	}
	require.NotNil(t, pipelineRunStatus)
	require.Equal(t, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS, pipelineRunStatus.PipelineRunStatusType)
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
