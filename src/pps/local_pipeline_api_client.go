package pps

import (
	"google.golang.org/grpc"

	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

type localPipelineAPIClient struct {
	pipelineAPIServer PipelineAPIServer
}

func newLocalPipelineAPIClient(pipelineAPIServer PipelineAPIServer) *localPipelineAPIClient {
	return &localPipelineAPIClient{pipelineAPIServer}
}

func (a *localPipelineAPIClient) CreatePipeline(ctx context.Context, request *CreatePipelineRequest, _ ...grpc.CallOption) (response *Pipeline, err error) {
	return a.pipelineAPIServer.CreatePipeline(ctx, request)
}

func (a *localPipelineAPIClient) GetPipeline(ctx context.Context, request *GetPipelineRequest, _ ...grpc.CallOption) (response *Pipeline, err error) {
	return a.pipelineAPIServer.GetPipeline(ctx, request)
}

func (a *localPipelineAPIClient) GetAllPipelines(ctx context.Context, request *google_protobuf.Empty, _ ...grpc.CallOption) (response *Pipelines, err error) {
	return a.pipelineAPIServer.GetAllPipelines(ctx, request)
}
