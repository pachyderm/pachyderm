package store

import "github.com/pachyderm/pachyderm/src/pps"

type Client interface {
	AddPipelineRun(pipelineRun *pps.PipelineRun) error
	GetPipelineRun(id string) (*pps.PipelineRun, error)
	GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error)
	AddPipelineRunStatus(id string, runStatusType pps.PipelineRunStatusType) error
	GetPipelineRunContainerIDs(id string) ([]string, error)
	AddPipelineRunContainerIDs(id string, containerIDs ...string) error
}

func NewInMemoryClient() Client {
	return newInMemoryClient()
}
