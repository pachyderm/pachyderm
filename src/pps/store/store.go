package store

import "github.com/pachyderm/pachyderm/src/pps"

type Client interface {
	Init() error
	AddPipelineRun(pipelineRun *pps.PipelineRun) error
	GetPipelineRun(id string) (*pps.PipelineRun, error)
	AddPipelineRunStatus(runStatus *pps.PipelineRunStatus) error
	GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error)
	AddPipelineRunContainerIDs(id string, containerIDs ...string) error
	GetPipelineRunContainers(id string) ([]*PipelineContainer, error)
}

func NewInMemoryClient() Client {
	return newInMemoryClient()
}

func NewRethinkClient(address string, databaseName string) (Client, error) {
	return newRethinkClient(address, databaseName)
}
