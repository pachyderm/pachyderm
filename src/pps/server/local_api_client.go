package server

import (
	"google.golang.org/grpc"

	"go.pachyderm.com/pachyderm/src/pps"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

type localAPIClient struct {
	apiServer pps.APIServer
}

func newLocalAPIClient(apiServer pps.APIServer) *localAPIClient {
	return &localAPIClient{apiServer}
}

func (a *localAPIClient) CreateJob(ctx context.Context, request *pps.CreateJobRequest, _ ...grpc.CallOption) (response *pps.Job, err error) {
	return a.apiServer.CreateJob(ctx, request)
}

func (a *localAPIClient) GetJob(ctx context.Context, request *pps.GetJobRequest, _ ...grpc.CallOption) (response *pps.Job, err error) {
	return a.apiServer.GetJob(ctx, request)
}

func (a *localAPIClient) GetJobsByPipelineName(ctx context.Context, request *pps.GetJobsByPipelineNameRequest, _ ...grpc.CallOption) (response *pps.Jobs, err error) {
	return a.apiServer.GetJobsByPipelineName(ctx, request)
}

func (a *localAPIClient) StartJob(ctx context.Context, request *pps.StartJobRequest, _ ...grpc.CallOption) (response *google_protobuf.Empty, err error) {
	return a.apiServer.StartJob(ctx, request)
}

func (a *localAPIClient) GetJobStatus(ctx context.Context, request *pps.GetJobStatusRequest, _ ...grpc.CallOption) (response *pps.JobStatus, err error) {
	return a.apiServer.GetJobStatus(ctx, request)
}

func (a *localAPIClient) GetJobLogs(request *pps.GetJobLogsRequest, responseServer pps.API_GetJobLogsServer) (err error) {
}

func (a *localAPIClient) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest, _ ...grpc.CallOption) (response *pps.Pipeline, err error) {
	return a.apiServer.CreatePipeline(ctx, request)
}

func (a *localAPIClient) GetPipeline(ctx context.Context, request *pps.GetPipelineRequest, _ ...grpc.CallOption) (response *pps.Pipeline, err error) {
	return a.apiServer.GetPipeline(ctx, request)
}

func (a *localAPIClient) GetAllPipelines(ctx context.Context, request *google_protobuf.Empty, _ ...grpc.CallOption) (response *pps.Pipelines, err error) {
	return a.apiServer.GetAllPipelines(ctx, request)
}
