package server

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/graph"
	"github.com/pachyderm/pachyderm/src/pkg/protoutil"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/run"
	"github.com/pachyderm/pachyderm/src/pps/source"
	"github.com/pachyderm/pachyderm/src/pps/store"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

var (
	emptyInstance := &google_protobuf.Empty{}
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

func (a *apiServer) CreatePipelineSource(ctx context.Context, request *pps.CreatePipelineSourceRequest) (*pps.CreatePipelineSourceResponse, error) {
	pipelineSource := request.PipelineSource
	if pipelineSource.Id != "" {
		return nil, fmt.Errorf("cannot set id when creating a pipeline source: %+v", pipelineSource)
	}
	pipelineSource.Id = strings.Replace(uuid.NewV4().String, "-", "", -1)
	if err := a.storeClient.CreatePipelineSource(pipelineSource); err != nil {
		return nil, err
	}
	return &pps.CreatePipelineSourceResponse{
		PipelineSource: pipelineSource,
	}, nil
}

func (a *apiServer) GetPipelineSource(ctx context.Context, request *pps.GetPipelineSourceRequest) (*pps.GetPipelineSourceResponse, error) {
	pipelineSource, err := a.storeClient.GetPipelineSource(request.PipelineSourceId)
	if err != nil {
		return nil, err
	}
	return &pps.GetPipelineSourceResponse{
		PipelineSource: pipelineSource,
	}, nil
}

func (a *apiServer) UpdatePipelineSource(ctx context.Context, request *pps.UpdatePipelineSourceRequest) (*pps.UpdatePipelineSourceResponse, error) {
	return nil, errors.New("not implemented")
}

func (a *apiServer) DeletePipelineSource(ctx context.Context, request *pps.DeletePipelineSourceRequest) (*google_protobuf.Empty, error) {
	if err := a.storeClient.DeletePipelineSource(request.PipelineSourceId); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *apiServer) ListPipelineSources(ctx context.Context, request *pps.ListPipelineSourcesRequest) (*pps.ListPipelineSourcesResponse, error) {
	pipelineSources, err := a.storeClient.GetAllPipelineSources()
	if err != nil {
		return nil, err
	}
	if request.Tags == nil || len(request.Tags) == 0 {
		return &pps.ListPipelineSourcesResponse{
			PipelineSource: pipelineSources,
		}, nil
	}
	var filteredPipelineSources []*pps.PipelineSource
	for _, pipelineSource := range pipelineSources {
		if tagsMatch(request.Tags, pipelineSource.Tags) {
			filteredPipelineSources = append(filteredPipelineSources, pipelineSource)
		}
	}
	return &pps.ListPipelineSourcesResponse{
		PipelineSource: pipelineSources,
	}, nil
}

func tagsMatch(expected map[string]string, tags map[string]string) bool {
	for key, value := range expected {
		if tags[key] != value {
			return false
		}
	}
	return true
}

func (a *apiServer) GetPipeline(ctx context.Context, request *pps.GetPipelineRequest) (*pps.GetPipelineResponse, error) {
	pipelineSource, err := a.storeClient.GetPipelineSource(request.PipelineSourceId)
	if err != nil {
		return nil, err
	}
	_, pipeline, err := source.NewSourcer().GetDirPathAndPipeline(pipelineSource)
	if err != nil {
		return nil, err
	}
	return &pps.GetPipelineResponse{
		Pipeline: pipeline,
	}, nil
}

func (a *apiServer) CreatePipelineRun(ctx context.Context, request *pps.CreatePipelineRunRequest) (*pps.CreatePipelineRunResponse, error) {
	pipelineSource, err := a.storeClient.GetPipelineSource(request.PipelineSourceId)
	if err != nil {
		return nil, err
	}
	dirPath, pipeline, err := source.NewSourcer().GetDirPathAndPipeline(pipelineSource)
	if err != nil {
		return "", err
	}
	pipelineRunID := strings.Replace(uuid.NewV4().String(), "-", "", -1)
	pipelineRun := &pps.PipelineRun{
		Id:             pipelineRunID,
		Pipeline:       pipeline,
		PipelineSource: pipelineSource,
	}
	if err := r.storeClient.AddPipelineRun(pipelineRun); err != nil {
		return "", err
	}
	protolog.Info(
		&AddedPipelineRun{
			PipelineRun: pipelineRun,
		},
	)
	return &pps.CreatePipelineRunResponse{
		PipelineRun: pipelineRun,
	}
}

func (a *apiServer) StartPipelineRun(ctx context.Context, request *pps.StartPipelineRunRequest) (*google_protobuf.Empty, error) {
	pipelineRun, err := a.storeClient.GetPipelineRun(request.PipelineRunId)
	if err != nil {
		return nil, err
	}
	runner := run.NewRunner(
		graph.NewGrapher(),
		a.containerClient,
		a.storeClient,
		a.timer,
	)
	if err := runner.Start(startPipelineRunRequest.PipelineSource); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *apiServer) ListPipelineRunsRequest(ctx context.Context, request *pps.ListPipelineRunsRequest) (*pps.ListPipelineRunsResponse, error) {
	return nil, errors.New("not implemented")
}

func (a *apiServer) GetPipelineRunStatus(ctx context.Context, request *pps.GetPipelineRunStatusRequest) (*pps.GetPipelineRunStatusResponse, error) {
	pipelineRunStatuses, err := a.storeClient.GetAllPipelineRunStatuses(request.PipelineRunId)
	if err != nil {
		return nil, err
	}
	if !request.All {
		pipelineRunStatuses = []*pps.PipelineRunStatus{pipelineRunStatus[0]}
	}
	return &pps.GetPipelineRunStatusResponse{
		PipelineRunStatus: pipelineRunStatuses,
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
	return protoutil.TimestampLess(s[i].Timestamp, s[j].Timestamp)
}
