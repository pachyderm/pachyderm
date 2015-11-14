package pps

import (
	"google.golang.org/grpc"
)

type apiClient struct {
	JobAPIClient
	InternalJobAPIClient
	PipelineAPIClient
}

func newAPIClient(clientConn *grpc.ClientConn) apiClient {
	return apiClient{
		NewJobAPIClient(clientConn),
		NewInternalJobAPIClient(clientConn),
		NewPipelineAPIClient(clientConn),
	}
}
