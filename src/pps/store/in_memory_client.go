package store

import (
	"fmt"
	"sync"

	"go.pedge.io/proto/time"

	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
)

type inMemoryClient struct {
	idToRun                          map[string]*pps.PipelineRun
	idToRunStatuses                  map[string][]*pps.PipelineRunStatus
	idToContainers                   map[string][]*pps.PipelineRunContainer
	idToLogs                         map[string][]*pps.PipelineRunLog
	inputCommitIDToPfsCommitMappings map[string][]*pps.PfsCommitMapping

	timer timing.Timer

	runLock               *sync.RWMutex
	runStatusesLock       *sync.RWMutex
	containersLock        *sync.RWMutex
	logsLock              *sync.RWMutex
	pfsCommitMappingsLock *sync.RWMutex
}

func newInMemoryClient() *inMemoryClient {
	return &inMemoryClient{
		make(map[string]*pps.PipelineRun),
		make(map[string][]*pps.PipelineRunStatus),
		make(map[string][]*pps.PipelineRunContainer),
		make(map[string][]*pps.PipelineRunLog),
		make(map[string][]*pps.PfsCommitMapping),
		timing.NewSystemTimer(),
		&sync.RWMutex{},
		&sync.RWMutex{},
		&sync.RWMutex{},
		&sync.RWMutex{},
		&sync.RWMutex{},
	}
}

func (c *inMemoryClient) Close() error {
	return nil
}

func (c *inMemoryClient) AddPipelineRun(pipelineRun *pps.PipelineRun) error {
	c.runLock.Lock()
	defer c.runLock.Unlock()
	c.runStatusesLock.Lock()
	defer c.runStatusesLock.Unlock()
	c.containersLock.Lock()
	defer c.containersLock.Unlock()
	c.logsLock.Lock()
	defer c.logsLock.Unlock()

	if _, ok := c.idToRun[pipelineRun.Id]; ok {
		return fmt.Errorf("run with id %s already added", pipelineRun.Id)
	}
	c.idToRun[pipelineRun.Id] = pipelineRun
	c.idToRunStatuses[pipelineRun.Id] = make([]*pps.PipelineRunStatus, 1)
	c.idToRunStatuses[pipelineRun.Id][0] = &pps.PipelineRunStatus{
		PipelineRunStatusType: pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_ADDED,
		Timestamp:             prototime.TimeToTimestamp(c.timer.Now()),
	}
	c.idToContainers[pipelineRun.Id] = make([]*pps.PipelineRunContainer, 0)
	c.idToLogs[pipelineRun.Id] = make([]*pps.PipelineRunLog, 0)
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

func (c *inMemoryClient) AddPipelineRunStatus(id string, statusType pps.PipelineRunStatusType) error {
	runStatus := &pps.PipelineRunStatus{
		PipelineRunId:         id,
		PipelineRunStatusType: statusType,
		Timestamp:             prototime.TimeToTimestamp(c.timer.Now()),
	}
	c.runStatusesLock.Lock()
	defer c.runStatusesLock.Unlock()

	_, ok := c.idToRunStatuses[id]
	if !ok {
		return fmt.Errorf("no run for id %s", runStatus.PipelineRunId)
	}
	c.idToRunStatuses[id] =
		append(c.idToRunStatuses[id], runStatus)
	return nil
}

func (c *inMemoryClient) GetPipelineRunContainers(id string) ([]*pps.PipelineRunContainer, error) {
	c.containersLock.RLock()
	defer c.containersLock.RUnlock()

	containers, ok := c.idToContainers[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return containers, nil
}

func (c *inMemoryClient) AddPipelineRunContainers(pipelineContainers ...*pps.PipelineRunContainer) error {
	c.containersLock.Lock()
	defer c.containersLock.Unlock()

	for _, container := range pipelineContainers {
		_, ok := c.idToContainers[container.PipelineRunId]
		if !ok {
			return fmt.Errorf("no run for id %s", container.PipelineRunId)
		}
		c.idToContainers[container.PipelineRunId] = append(c.idToContainers[container.PipelineRunId], container)
	}
	return nil
}

func (c *inMemoryClient) GetPipelineRunLogs(id string) ([]*pps.PipelineRunLog, error) {
	c.logsLock.RLock()
	defer c.logsLock.RUnlock()

	logs, ok := c.idToLogs[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return logs, nil
}

func (c *inMemoryClient) AddPipelineRunLogs(pipelineLogs ...*pps.PipelineRunLog) error {
	c.logsLock.Lock()
	defer c.logsLock.Unlock()

	for _, log := range pipelineLogs {
		if log.Timestamp == nil {
			return fmt.Errorf("timestamp not set for %v", log)
		}
		_, ok := c.idToLogs[log.PipelineRunId]
		if !ok {
			return fmt.Errorf("no run for id %s", log.PipelineRunId)
		}
		c.idToLogs[log.PipelineRunId] = append(c.idToLogs[log.PipelineRunId], log)
	}
	return nil
}

func (c *inMemoryClient) AddPfsCommitMapping(pfsCommitMapping *pps.PfsCommitMapping) error {
	if pfsCommitMapping.Timestamp == nil {
		return fmt.Errorf("timestamp not set for %v", pfsCommitMapping)
	}
	c.pfsCommitMappingsLock.Lock()
	defer c.pfsCommitMappingsLock.Unlock()
	if _, ok := c.inputCommitIDToPfsCommitMappings[pfsCommitMapping.InputCommitId]; !ok {
		c.inputCommitIDToPfsCommitMappings[pfsCommitMapping.InputCommitId] = make([]*pps.PfsCommitMapping, 0)
	}
	c.inputCommitIDToPfsCommitMappings[pfsCommitMapping.InputCommitId] = append(
		c.inputCommitIDToPfsCommitMappings[pfsCommitMapping.InputCommitId],
		pfsCommitMapping,
	)
	return nil
}

func (c *inMemoryClient) GetPfsCommitMappingLatest(inputRepositoryName string, inputCommitID string) (*pps.PfsCommitMapping, error) {
	c.pfsCommitMappingsLock.RLock()
	defer c.pfsCommitMappingsLock.RUnlock()
	pfsCommitMappings, ok := c.inputCommitIDToPfsCommitMappings[inputCommitID]
	if !ok {
		return nil, nil
	}
	return getPfsCommitMappingLatestInMemory(pfsCommitMappings, inputRepositoryName)
}
