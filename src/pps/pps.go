package pps

import ()

func NewLocalJobAPIClient(jobAPIServer JobAPIServer) JobAPIClient {
	return newLocalJobAPIClient(jobAPIServer)
}

func NewLocalPipelineAPIClient(pipelineAPIServer PipelineAPIServer) PipelineAPIClient {
	return newLocalPipelineAPIClient(pipelineAPIServer)
}
