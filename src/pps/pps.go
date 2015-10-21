package pps // import "go.pachyderm.com/pachyderm/src/pps"

func NewLocalJobAPIClient(jobAPIServer JobAPIServer) JobAPIClient {
	return newLocalJobAPIClient(jobAPIServer)
}

func NewLocalPipelineAPIClient(pipelineAPIServer PipelineAPIServer) PipelineAPIClient {
	return newLocalPipelineAPIClient(pipelineAPIServer)
}
