package pps

import (
	"google.golang.org/grpc"
)

func NewLocalJobAPIClient(jobAPIServer JobAPIServer) JobAPIClient {
	return newLocalJobAPIClient(jobAPIServer)
}

func NewLocalPipelineAPIClient(pipelineAPIServer PipelineAPIServer) PipelineAPIClient {
	return newLocalPipelineAPIClient(pipelineAPIServer)
}

type APIClient interface {
	JobAPIClient
	InternalJobAPIClient
	PipelineAPIClient
}

func NewAPIClient(clientConn *grpc.ClientConn) APIClient {
	return newAPIClient(clientConn)
}
