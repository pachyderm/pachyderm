package persist

import (
	"google.golang.org/grpc"

	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

type localAPIClient struct {
	apiServer APIServer
}

func newLocalAPIClient(apiServer APIServer) *localAPIClient {
	return &localAPIClient{apiServer}
}

func (a *localAPIClient) CreateJob(ctx context.Context, request *Job, _ ...grpc.CallOption) (response *Job, err error) {
	return a.apiServer.CreateJob(ctx, request)
}

func (a *localAPIClient) GetJobByID(ctx context.Context, request *google_protobuf.StringValue, _ ...grpc.CallOption) (response *Job, err error) {
	return a.apiServer.GetJobByID(ctx, request)
}

func (a *localAPIClient) GetJobsByPipelineID(ctx context.Context, request *google_protobuf.StringValue, _ ...grpc.CallOption) (response *Jobs, err error) {
	return a.apiServer.GetJobsByPipelineID(ctx, request)
}

func (a *localAPIClient) CreateJobStatus(ctx context.Context, request *JobStatus, _ ...grpc.CallOption) (response *JobStatus, err error) {
	return a.apiServer.CreateJobStatus(ctx, request)
}

func (a *localAPIClient) GetJobStatusesByJobID(ctx context.Context, request *google_protobuf.StringValue, _ ...grpc.CallOption) (response *JobStatuses, err error) {
	return a.apiServer.GetJobStatusesByJobID(ctx, request)
}

func (a *localAPIClient) CreateJobLog(ctx context.Context, request *JobLog, _ ...grpc.CallOption) (response *JobLog, err error) {
	return a.apiServer.CreateJobLog(ctx, request)
}

func (a *localAPIClient) GetJobLogsByJobID(ctx context.Context, request *google_protobuf.StringValue, _ ...grpc.CallOption) (response *JobLogs, err error) {
	return a.apiServer.GetJobLogsByJobID(ctx, request)
}

func (a *localAPIClient) CreatePipeline(ctx context.Context, request *Pipeline, _ ...grpc.CallOption) (response *Pipeline, err error) {
	return a.apiServer.CreatePipeline(ctx, request)
}

func (a *localAPIClient) UpdatePipeline(ctx context.Context, request *Pipeline, _ ...grpc.CallOption) (response *Pipeline, err error) {
	return a.apiServer.UpdatePipeline(ctx, request)
}

func (a *localAPIClient) GetPipelineByID(ctx context.Context, request *google_protobuf.StringValue, _ ...grpc.CallOption) (response *Pipeline, err error) {
	return a.apiServer.GetPipelineByID(ctx, request)
}

func (a *localAPIClient) GetPipelinesByName(ctx context.Context, request *google_protobuf.StringValue, _ ...grpc.CallOption) (response *Pipelines, err error) {
	return a.apiServer.GetPipelinesByName(ctx, request)
}
