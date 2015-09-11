package store

import "github.com/pachyderm/pachyderm/src/pps"

type Client interface {
	Close() error
	CreatePipelineRun(pipelineRun *pps.PipelineRun) error
	GetPipelineRun(id string) (*pps.PipelineRun, error)
	GetAllPipelineRuns(pipelineID string) ([]*pps.PipelineRun, error)
	CreatePipeline(pipeline *pps.Pipeline) error
	GetPipeline(id string) (*pps.Pipeline, error)
	GetAllPipelines(pipelineSourceID string) ([]*pps.Pipeline, error)
	CreatePipelineRunStatus(id string, statusType pps.PipelineRunStatusType) error
	GetAllPipelineRunStatuses(pipelineRunID string) ([]*pps.PipelineRunStatus, error)
	CreatePipelineRunContainers(pipelineContainers ...*pps.PipelineRunContainer) error
	GetPipelineRunContainers(pipelineRunID string) ([]*pps.PipelineRunContainer, error)
	CreatePipelineRunLogs(pipelineRunLogs ...*pps.PipelineRunLog) error
	GetPipelineRunLogs(pipelineRunID string) ([]*pps.PipelineRunLog, error)
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
