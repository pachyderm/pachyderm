package pps

import (
	"google.golang.org/grpc"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
)

type localJobAPIClient struct {
	jobAOIServer JobAPIServer
}

func newLocalJobAPIClient(jobAOIServer JobAPIServer) *localJobAPIClient {
	return &localJobAPIClient{jobAOIServer}
}

func (a *localJobAPIClient) CreateJob(ctx context.Context, request *CreateJobRequest, _ ...grpc.CallOption) (response *Job, err error) {
	return a.jobAOIServer.CreateJob(ctx, request)
}

func (a *localJobAPIClient) GetJob(ctx context.Context, request *GetJobRequest, _ ...grpc.CallOption) (response *Job, err error) {
	return a.jobAOIServer.GetJob(ctx, request)
}

func (a *localJobAPIClient) GetJobsByPipelineName(ctx context.Context, request *GetJobsByPipelineNameRequest, _ ...grpc.CallOption) (response *Jobs, err error) {
	return a.jobAOIServer.GetJobsByPipelineName(ctx, request)
}

func (a *localJobAPIClient) StartJob(ctx context.Context, request *StartJobRequest, _ ...grpc.CallOption) (response *google_protobuf.Empty, err error) {
	return a.jobAOIServer.StartJob(ctx, request)
}

func (a *localJobAPIClient) GetJobStatus(ctx context.Context, request *GetJobStatusRequest, _ ...grpc.CallOption) (response *JobStatus, err error) {
	return a.jobAOIServer.GetJobStatus(ctx, request)
}

func (a *localJobAPIClient) GetJobLogs(ctx context.Context, request *GetJobLogsRequest, _ ...grpc.CallOption) (client JobAPI_GetJobLogsClient, err error) {
	steamingBytesRelayer := protostream.NewStreamingBytesRelayer(ctx)
	if err := a.jobAOIServer.GetJobLogs(request, steamingBytesRelayer); err != nil {
		return nil, err
	}
	return steamingBytesRelayer, nil
}
