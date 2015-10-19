package pps

import (
	"google.golang.org/grpc"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
)

type localAPIClient struct {
	apiServer APIServer
}

func newLocalAPIClient(apiServer APIServer) *localAPIClient {
	return &localAPIClient{apiServer}
}

func (a *localAPIClient) CreateJob(ctx context.Context, request *CreateJobRequest, _ ...grpc.CallOption) (response *Job, err error) {
	return a.apiServer.CreateJob(ctx, request)
}

func (a *localAPIClient) GetJob(ctx context.Context, request *GetJobRequest, _ ...grpc.CallOption) (response *Job, err error) {
	return a.apiServer.GetJob(ctx, request)
}

func (a *localAPIClient) GetJobsByPipelineName(ctx context.Context, request *GetJobsByPipelineNameRequest, _ ...grpc.CallOption) (response *Jobs, err error) {
	return a.apiServer.GetJobsByPipelineName(ctx, request)
}

func (a *localAPIClient) StartJob(ctx context.Context, request *StartJobRequest, _ ...grpc.CallOption) (response *google_protobuf.Empty, err error) {
	return a.apiServer.StartJob(ctx, request)
}

func (a *localAPIClient) GetJobStatus(ctx context.Context, request *GetJobStatusRequest, _ ...grpc.CallOption) (response *JobStatus, err error) {
	return a.apiServer.GetJobStatus(ctx, request)
}

func (a *localAPIClient) GetJobLogs(ctx context.Context, request *GetJobLogsRequest, _ ...grpc.CallOption) (client API_GetJobLogsClient, err error) {
	steamingBytesRelayer := protostream.NewStreamingBytesRelayer(ctx)
	if err := a.apiServer.GetJobLogs(request, steamingBytesRelayer); err != nil {
		return nil, err
	}
	return steamingBytesRelayer, nil
}

func (a *localAPIClient) CreatePipeline(ctx context.Context, request *CreatePipelineRequest, _ ...grpc.CallOption) (response *Pipeline, err error) {
	return a.apiServer.CreatePipeline(ctx, request)
}

func (a *localAPIClient) GetPipeline(ctx context.Context, request *GetPipelineRequest, _ ...grpc.CallOption) (response *Pipeline, err error) {
	return a.apiServer.GetPipeline(ctx, request)
}

func (a *localAPIClient) GetAllPipelines(ctx context.Context, request *google_protobuf.Empty, _ ...grpc.CallOption) (response *Pipelines, err error) {
	return a.apiServer.GetAllPipelines(ctx, request)
}
