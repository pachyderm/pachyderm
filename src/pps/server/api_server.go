package server

import (
	"os"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/graph"
	"github.com/pachyderm/pachyderm/src/pps/run"
	"github.com/pachyderm/pachyderm/src/pps/source"
	"github.com/pachyderm/pachyderm/src/pps/store"
	"github.com/peter-edge/go-google-protobuf"
	"golang.org/x/net/context"
)

var (
	emptyInstance              = &google_protobuf.Empty{}
	getVersionResponseInstance = &pps.GetVersionResponse{
		Version: &pps.Version{
			Major:      common.MajorVersion,
			Minor:      common.MinorVersion,
			Micro:      common.MicroVersion,
			Additional: common.AdditionalVersion,
		},
	}
)

type apiServer struct {
	storeClient store.Client
}

func newAPIServer(storeClient store.Client) *apiServer {
	return &apiServer{storeClient}
}

func (a *apiServer) GetVersion(ctx context.Context, empty *google_protobuf.Empty) (*pps.GetVersionResponse, error) {
	return getVersionResponseInstance, nil
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
