package store

import "github.com/pachyderm/pachyderm/src/pps"

type Client interface {
	Close() error
	AddPipelineRun(pipelineRun *pps.PipelineRun) error
	GetPipelineRun(id string) (*pps.PipelineRun, error)
	AddPipelineRunStatus(id string, statusType pps.PipelineRunStatusType) error
	GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error)
	AddPipelineRunContainers(pipelineContainers ...*pps.PipelineRunContainer) error
	GetPipelineRunContainers(id string) ([]*pps.PipelineRunContainer, error)
}

func NewInMemoryClient() Client {
	return newInMemoryClient()
}

func NewRethinkClient(address string, databaseName string) (Client, error) {
	return newRethinkClient(address, databaseName)
}
