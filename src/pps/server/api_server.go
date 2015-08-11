package server

import (
	"os"

	"github.com/pachyderm/pachyderm/src/pkg/graph"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/run"
	"github.com/pachyderm/pachyderm/src/pps/source"
	"github.com/pachyderm/pachyderm/src/pps/store"
	"golang.org/x/net/context"
)

type apiServer struct {
	storeClient store.Client
}

func newAPIServer(storeClient store.Client) *apiServer {
	return &apiServer{storeClient}
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
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		dockerHost = "unix:///var/run/docker.sock"
	}
	containerClient, err := container.NewDockerClient(
		container.DockerClientOptions{
			Host: dockerHost,
		},
	)
	if err != nil {
		return nil, err
	}
	runner := run.NewRunner(
		source.NewSourcer(),
		graph.NewGrapher(),
		containerClient,
		a.storeClient,
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
	return &pps.GetPipelineRunLogsResponse{
		PipelineRunLog: filteredPipelineRunLogs,
	}, nil
}
