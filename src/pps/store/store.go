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
	AddPipelineRunLogs(pipelineRunLogs ...*pps.PipelineRunLog) error
	GetPipelineRunLogs(id string) ([]*pps.PipelineRunLog, error)
	AddPfsCommitMapping(pfsCommitMapping *pps.PfsCommitMapping) error
	GetPfsCommitMappingLatest(inputRepository string, inputCommitID string) (*pps.PfsCommitMapping, error)
	AddPipelineSource(pipelineSource *pps.PipelineSource) error
	GetPipelineSource(id string) (*pps.PipelineSource, error)
	UpdatePipelineSource(pipelineSource *pps.PipelineSource) error
	DeletePipelineSource(pipelineSource *pps.PipelineSource) error
	GetAllPipelineSources() ([]*pps.PipelineSource, error)
}

func NewRethinkClient(address string, databaseName string) (Client, error) {
	return newRethinkClient(address, databaseName)
}
