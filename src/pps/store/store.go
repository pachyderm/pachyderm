package store

import "github.com/pachyderm/pachyderm/src/pps"

type Client interface {
	Close() error
	CreatePipelineRun(pipelineRun *pps.PipelineRun) error
	GetPipelineRun(id string) (*pps.PipelineRun, error)
	CreatePipelineRunStatus(id string, statusType pps.PipelineRunStatusType) error
	GetAllPipelineRunStatuses(id string) ([]*pps.PipelineRunStatus, error)
	CreatePipelineRunContainers(pipelineContainers ...*pps.PipelineRunContainer) error
	GetPipelineRunContainers(id string) ([]*pps.PipelineRunContainer, error)
	CreatePipelineRunLogs(pipelineRunLogs ...*pps.PipelineRunLog) error
	GetPipelineRunLogs(id string) ([]*pps.PipelineRunLog, error)
	CreatePfsCommitMapping(pfsCommitMapping *pps.PfsCommitMapping) error
	GetPfsCommitMappingLatest(inputRepository string, inputCommitID string) (*pps.PfsCommitMapping, error)
	CreatePipelineSource(pipelineSource *pps.PipelineSource) error
	GetPipelineSource(id string) (*pps.PipelineSource, error)
	UpdatePipelineSource(pipelineSource *pps.PipelineSource) error
	ArchivePipelineSource(id string) error
	GetAllPipelineSources() ([]*pps.PipelineSource, error)
}

func NewRethinkClient(address string, databaseName string) (Client, error) {
	return newRethinkClient(address, databaseName)
}
