package store

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/pps"
)

type runInfo struct {
	pipelineSource *pps.PipelineSource
	pipeline       *pps.Pipeline
}

type inMemoryClient struct {
	idToRunInfo     map[string]*runInfo
	idToRunStatuses map[string][]*pps.RunStatus

	runInfoLock      *sync.RWMutex
	runStatuesesLock *sync.RWMutex
}

func newInMemoryClient() *inMemoryClient {
	return &inMemoryClient{
		make(map[string]*runInfo),
		make(map[string][]*pps.RunStatus),
		&sync.RWMutex{},
		&sync.RWMutex{},
	}
}

func (c *inMemoryClient) AddRun(id string, pipelineSource *pps.PipelineSource, pipeline *pps.Pipeline) error {
	return nil
}

func (c *inMemoryClient) GetRunPipelineSource(id string) (*pps.Pipeline, error) {
	return nil, nil
}

func (c *inMemoryClient) GetRunPipeline(id string) (*pps.Pipeline, error) {
	return nil, nil
}

func (c *inMemoryClient) GetRunStatusLatest(id string) (*pps.RunStatus, error) {
	return nil, nil
}

func (c *inMemoryClient) AddRunStatus(id string, runStatusType pps.RunStatusType) error {
	return nil
}
