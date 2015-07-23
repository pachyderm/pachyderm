package store

import "github.com/pachyderm/pachyderm/src/pps"

type Client interface {
	AddPipelineRun(id string, pipelineSource *pps.PipelineSource, pipeline *pps.Pipeline) error
	GetPipelineRunPipelineSource(id string) (*pps.PipelineSource, error)
	GetPipelineRunPipeline(id string) (*pps.Pipeline, error)
	GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error)
	AddPipelineRunStatus(id string, runStatusType pps.PipelineRunStatusType) error
}

func NewInMemoryClient() Client {
	return newInMemoryClient()
}
