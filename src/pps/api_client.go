package pps

import (
	"google.golang.org/grpc"
)

type apiClient struct {
	JobAPIClient
	PipelineAPIClient
}

func newAPIClient(clientConn *grpc.ClientConn) apiClient {
	return apiClient{
		NewJobAPIClient(clientConn),
		NewPipelineAPIClient(clientConn),
	}
}
