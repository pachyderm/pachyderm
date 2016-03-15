package pps

import (
	"go.pedge.io/pb/go/google/protobuf"
	"google.golang.org/grpc"

	"golang.org/x/net/context"
)

type localPipelineAPIClient struct {
	pipelineAPIServer PipelineAPIServer
}

func newLocalPipelineAPIClient(pipelineAPIServer PipelineAPIServer) *localPipelineAPIClient {
	return &localPipelineAPIClient{pipelineAPIServer}
}

func (a *localPipelineAPIClient) CreatePipeline(ctx context.Context, request *CreatePipelineRequest, _ ...grpc.CallOption) (response *google_protobuf.Empty, err error) {
	return a.pipelineAPIServer.CreatePipeline(ctx, request)
}

func (a *localPipelineAPIClient) InspectPipeline(ctx context.Context, request *InspectPipelineRequest, _ ...grpc.CallOption) (response *PipelineInfo, err error) {
	return a.pipelineAPIServer.InspectPipeline(ctx, request)
}

func (a *localPipelineAPIClient) ListPipeline(ctx context.Context, request *ListPipelineRequest, _ ...grpc.CallOption) (response *PipelineInfos, err error) {
	return a.pipelineAPIServer.ListPipeline(ctx, request)
}

func (a *localPipelineAPIClient) DeletePipeline(ctx context.Context, request *DeletePipelineRequest, _ ...grpc.CallOption) (response *google_protobuf.Empty, err error) {
	return a.pipelineAPIServer.DeletePipeline(ctx, request)
}
