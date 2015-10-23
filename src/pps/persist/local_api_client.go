package persist

import (
	"google.golang.org/grpc"

	"go.pachyderm.com/pachyderm/src/pps"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

type localAPIClient struct {
	apiServer APIServer
}

func newLocalAPIClient(apiServer APIServer) *localAPIClient {
	return &localAPIClient{apiServer}
}

func (a *localAPIClient) CreateJobInfo(ctx context.Context, request *JobInfo, _ ...grpc.CallOption) (response *JobInfo, err error) {
	return a.apiServer.CreateJobInfo(ctx, request)
}

func (a *localAPIClient) GetJobInfo(ctx context.Context, request *pps.Job, _ ...grpc.CallOption) (response *JobInfo, err error) {
	return a.apiServer.GetJobInfo(ctx, request)
}

func (a *localAPIClient) GetJobInfosByPipeline(ctx context.Context, request *pps.Pipeline, _ ...grpc.CallOption) (response *JobInfos, err error) {
	return a.apiServer.GetJobInfosByPipeline(ctx, request)
}

func (a *localAPIClient) CreateJobStatus(ctx context.Context, request *JobStatus, _ ...grpc.CallOption) (response *JobStatus, err error) {
	return a.apiServer.CreateJobStatus(ctx, request)
}

func (a *localAPIClient) GetJobStatuses(ctx context.Context, request *pps.Job, _ ...grpc.CallOption) (response *JobStatuses, err error) {
	return a.apiServer.GetJobStatuses(ctx, request)
}

func (a *localAPIClient) CreateJobLog(ctx context.Context, request *JobLog, _ ...grpc.CallOption) (response *JobLog, err error) {
	return a.apiServer.CreateJobLog(ctx, request)
}

func (a *localAPIClient) GetJobLogs(ctx context.Context, request *pps.Job, _ ...grpc.CallOption) (response *JobLogs, err error) {
	return a.apiServer.GetJobLogs(ctx, request)
}

func (a *localAPIClient) CreatePipelineInfo(ctx context.Context, request *PipelineInfo, _ ...grpc.CallOption) (response *PipelineInfo, err error) {
	return a.apiServer.CreatePipelineInfo(ctx, request)
}

func (a *localAPIClient) GetPipelineInfo(ctx context.Context, request *pps.Pipeline, _ ...grpc.CallOption) (response *PipelineInfo, err error) {
	return a.apiServer.GetPipelineInfo(ctx, request)
}

func (a *localAPIClient) ListPipelineInfos(ctx context.Context, request *google_protobuf.Empty, _ ...grpc.CallOption) (response *PipelineInfos, err error) {
	return a.apiServer.ListPipelineInfos(ctx, request)
}

func (a *localAPIClient) DeletePipelineInfo(ctx context.Context, request *pps.Pipeline, _ ...grpc.CallOption) (response *google_protobuf.Empty, err error) {
	return a.apiServer.DeletePipelineInfo(ctx, request)
}
