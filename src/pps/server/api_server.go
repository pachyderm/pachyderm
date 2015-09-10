package server

import (
	"sort"

	"go.pedge.io/proto/time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/graph"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/run"
	"github.com/pachyderm/pachyderm/src/pps/source"
	"github.com/pachyderm/pachyderm/src/pps/store"
	"golang.org/x/net/context"
)

type apiServer struct {
	pfsAPIClient    pfs.ApiClient
	containerClient container.Client
	storeClient     store.Client
	timer           timing.Timer
}

func newAPIServer(pfsAPIClient pfs.ApiClient, containerClient container.Client, storeClient store.Client, timer timing.Timer) *apiServer {
	return &apiServer{pfsAPIClient, containerClient, storeClient, timer}
}

func (a *apiServer) GetPipeline(ctx context.Context, getPipelineRequest *pps.GetPipelineRequest) (*pps.GetPipelineResponse, error) {
	_, pipeline, err := source.NewSourcer().GetDirPathAndPipeline(getPipelineRequest.PipelineSource)
	if err != nil {
		return nil, err
	}
	return &pps.GetPipelineResponse{
		Pipeline: pipeline,
	}, nil
}

func (a *apiServer) StartPipelineRun(ctx context.Context, startPipelineRunRequest *pps.StartPipelineRunRequest) (*pps.StartPipelineRunResponse, error) {
	runner := run.NewRunner(
		source.NewSourcer(),
		graph.NewGrapher(),
		a.containerClient,
		a.storeClient,
		a.timer,
	)
	runID, err := runner.Start(startPipelineRunRequest.PipelineSource)
	if err != nil {
		return nil, err
	}
	return &pps.StartPipelineRunResponse{
		PipelineRunId: runID,
	}, nil
}

func (a *apiServer) GetPipelineRunStatus(ctx context.Context, getRunStatusRequest *pps.GetPipelineRunStatusRequest) (*pps.GetPipelineRunStatusResponse, error) {
	pipelineRunStatus, err := a.storeClient.GetPipelineRunStatusLatest(getRunStatusRequest.PipelineRunId)
	if err != nil {
		return nil, err
	}
	return &pps.GetPipelineRunStatusResponse{
		PipelineRunStatus: pipelineRunStatus,
	}, nil
}

func (a *apiServer) GetPipelineRunLogs(ctx context.Context, getRunLogsRequest *pps.GetPipelineRunLogsRequest) (*pps.GetPipelineRunLogsResponse, error) {
	pipelineRunLogs, err := a.storeClient.GetPipelineRunLogs(getRunLogsRequest.PipelineRunId)
	if err != nil {
		return nil, err
	}
	filteredPipelineRunLogs := pipelineRunLogs
	if getRunLogsRequest.Node != "" {
		filteredPipelineRunLogs = make([]*pps.PipelineRunLog, 0)
		for _, pipelineRunLog := range pipelineRunLogs {
			if pipelineRunLog.Node == getRunLogsRequest.Node {
				filteredPipelineRunLogs = append(filteredPipelineRunLogs, pipelineRunLog)
			}
		}
	}
	sort.Sort(sortByTimestamp(filteredPipelineRunLogs))
	return &pps.GetPipelineRunLogsResponse{
		PipelineRunLog: filteredPipelineRunLogs,
	}, nil
}

type sortByTimestamp []*pps.PipelineRunLog

func (s sortByTimestamp) Len() int          { return len(s) }
func (s sortByTimestamp) Swap(i int, j int) { s[i], s[j] = s[j], s[i] }
func (s sortByTimestamp) Less(i int, j int) bool {
	return prototime.TimestampLess(s[i].Timestamp, s[j].Timestamp)
}
