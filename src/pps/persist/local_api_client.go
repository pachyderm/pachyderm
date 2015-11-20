package persist

import (
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/pps"
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

func (a *localAPIClient) ListJobInfos(ctx context.Context, request *pps.ListJobRequest, _ ...grpc.CallOption) (response *JobInfos, err error) {
	return a.apiServer.ListJobInfos(ctx, request)
}

func (a *localAPIClient) DeleteJobInfo(ctx context.Context, request *pps.Job, _ ...grpc.CallOption) (response *google_protobuf.Empty, err error) {
	return a.apiServer.DeleteJobInfo(ctx, request)
}

func (a *localAPIClient) CreateJobOutput(ctx context.Context, request *JobOutput, _ ...grpc.CallOption) (response *JobOutput, err error) {
	return a.apiServer.CreateJobOutput(ctx, request)
}

func (a *localAPIClient) GetJobOutput(ctx context.Context, request *pps.Job, _ ...grpc.CallOption) (response *JobOutput, err error) {
	return a.apiServer.GetJobOutput(ctx, request)
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
