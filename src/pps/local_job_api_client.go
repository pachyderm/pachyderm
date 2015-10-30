package pps

import (
	"google.golang.org/grpc"

	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
)

type localJobAPIClient struct {
	jobAPIServer JobAPIServer
}

func newLocalJobAPIClient(jobAPIServer JobAPIServer) *localJobAPIClient {
	return &localJobAPIClient{jobAPIServer}
}

func (a *localJobAPIClient) CreateJob(ctx context.Context, request *CreateJobRequest, _ ...grpc.CallOption) (response *Job, err error) {
	return a.jobAPIServer.CreateJob(ctx, request)
}

func (a *localJobAPIClient) InspectJob(ctx context.Context, request *InspectJobRequest, _ ...grpc.CallOption) (response *JobInfo, err error) {
	return a.jobAPIServer.InspectJob(ctx, request)
}

func (a *localJobAPIClient) ListJob(ctx context.Context, request *ListJobRequest, _ ...grpc.CallOption) (response *JobInfos, err error) {
	return a.jobAPIServer.ListJob(ctx, request)
}

func (a *localJobAPIClient) GetJobLogs(ctx context.Context, request *GetJobLogsRequest, _ ...grpc.CallOption) (client JobAPI_GetJobLogsClient, err error) {
	steamingBytesRelayer := protostream.NewStreamingBytesRelayer(ctx)
	if err := a.jobAPIServer.GetJobLogs(request, steamingBytesRelayer); err != nil {
		return nil, err
	}
	return steamingBytesRelayer, nil
}
