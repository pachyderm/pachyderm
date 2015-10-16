package watch

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

func (a *localAPIClient) Start(ctx context.Context, request *google_protobuf.Empty, _ ...grpc.CallOption) (response *google_protobuf.Empty, err error) {
	return a.apiServer.Start(ctx, request)
}

func (a *localAPIClient) RegisterChangeEvent(ctx context.Context, request *ChangeEvent, _ ...grpc.CallOption) (response *google_protobuf.Empty, err error) {
	return a.apiServer.RegisterChangeEvent(ctx, request)
}
