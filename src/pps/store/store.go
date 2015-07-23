package store

import "github.com/pachyderm/pachyderm/src/pps"

type Client interface {
	AddRun(id string, pipelineSource *pps.PipelineSource, pipeline *pps.Pipeline) error
	GetRunPipelineSource(id string) (*pps.Pipeline, error)
	GetRunPipeline(id string) (*pps.Pipeline, error)
	GetRunStatusLatest(id string) (*pps.RunStatus, error)
	AddRunStatus(id string, runStatusType pps.RunStatusType) error
}

func NewInMemoryClient() Client {
	return newInMemoryClient()
}
