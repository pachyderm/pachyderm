package store

import (
	"fmt"
	"sync"

	"github.com/pachyderm/pachyderm/src/pps"
)

type inMemoryClient struct {
	idToRun         map[string]*pps.PipelineRun
	idToRunStatuses map[string][]*pps.PipelineRunStatus
	idToContainers  map[string][]*PipelineContainer

	timer timer

	runLock          *sync.RWMutex
	runStatusesLock  *sync.RWMutex
	containerIDsLock *sync.RWMutex
}

func newInMemoryClient() *inMemoryClient {
	return &inMemoryClient{
		make(map[string]*pps.PipelineRun),
		make(map[string][]*pps.PipelineRunStatus),
		make(map[string][]*PipelineContainer),
		defaultTimer,
		&sync.RWMutex{},
		&sync.RWMutex{},
		&sync.RWMutex{},
	}
}

func (c *inMemoryClient) AddPipelineRun(pipelineRun *pps.PipelineRun) error {
	c.runLock.Lock()
	defer c.runLock.Unlock()
	c.runStatusesLock.Lock()
	defer c.runStatusesLock.Unlock()

	if _, ok := c.idToRun[pipelineRun.Id]; ok {
		return fmt.Errorf("run with id %s already added", pipelineRun.Id)
	}
	c.idToRun[pipelineRun.Id] = pipelineRun
	c.idToRunStatuses[pipelineRun.Id] = make([]*pps.PipelineRunStatus, 1)
	c.idToRunStatuses[pipelineRun.Id][0] = &pps.PipelineRunStatus{
		PipelineRunStatusType: pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_ADDED,
		Timestamp:             timeToTimestamp(c.timer.Now()),
	}
	c.idToContainers[pipelineRun.Id] = make([]*PipelineContainer, 0)
	return nil
}

func (c *inMemoryClient) GetPipelineRun(id string) (*pps.PipelineRun, error) {
	c.runLock.RLock()
	defer c.runLock.RUnlock()

	pipelineRun, ok := c.idToRun[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return pipelineRun, nil
}

func (c *inMemoryClient) GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error) {
	c.runStatusesLock.RLock()
	defer c.runStatusesLock.RUnlock()

	runStatuses, ok := c.idToRunStatuses[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return runStatuses[len(runStatuses)-1], nil
}

func (c *inMemoryClient) AddPipelineRunStatus(runStatus *pps.PipelineRunStatus) error {
	c.runStatusesLock.Lock()
	defer c.runStatusesLock.Unlock()

	_, ok := c.idToRunStatuses[runStatus.PipelineRunId]
	if !ok {
		return fmt.Errorf("no run for id %s", runStatus.PipelineRunId)
	}
	c.idToRunStatuses[runStatus.PipelineRunId] =
		append(c.idToRunStatuses[runStatus.PipelineRunId], runStatus)
	return nil
}

func (c *inMemoryClient) GetPipelineRunContainers(id string) ([]*PipelineContainer, error) {
	c.containerIDsLock.RLock()
	defer c.containerIDsLock.RUnlock()

	containers, ok := c.idToContainers[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return containers, nil
}

func (c *inMemoryClient) AddPipelineRunContainerIDs(id string, containerIDs ...string) error {
	c.containerIDsLock.Lock()
	defer c.containerIDsLock.Unlock()

	_, ok := c.idToContainers[id]
	if !ok {
		return fmt.Errorf("no run for id %s", id)
	}
	for _, containerID := range containerIDs {
		c.idToContainers[id] = append(c.idToContainers[id], &PipelineContainer{PipelineRunId: id, ContainerId: containerID})
	}
	return nil
}
