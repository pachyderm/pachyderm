package server

import (
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

type apiServer struct{}

func newAPIServer() *apiServer {
	return &apiServer{}
}

func (a *apiServer) CreateJob(ctx context.Context, request *pps.CreateJobRequest) (response *pps.Job, err error) {
	return nil, nil
}

func (a *apiServer) GetJob(ctx context.Context, request *pps.GetJobRequest) (response *pps.Job, err error) {
	return nil, nil
}

func (a *apiServer) GetJobsByPipelineName(ctx context.Context, request *pps.GetJobsByPipelineNameRequest) (response *pps.Jobs, err error) {
	return nil, nil
}

func (a *apiServer) StartJob(ctx context.Context, request *pps.StartJobRequest) (response *google_protobuf.Empty, err error) {
	return nil, nil
}

func (a *apiServer) GetJobStatus(ctx context.Context, request *pps.GetJobStatusRequest) (response *pps.JobStatus, err error) {
	return nil, nil
}

func (a *apiServer) GetJobLogs(request *pps.GetJobLogsRequest, responseServer pps.API_GetJobLogsServer) (err error) {
	return nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *pps.Pipeline, err error) {
	return nil, nil
}

func (a *apiServer) GetPipeline(ctx context.Context, request *pps.GetPipelineRequest) (response *pps.Pipeline, err error) {
	return nil, nil
}

func (a *apiServer) GetAllPipelines(ctx context.Context, request *google_protobuf.Empty) (response *pps.Pipelines, err error) {
	return nil, nil
}
