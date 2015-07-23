package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/peter-edge/go-google-protobuf"
)

var (
	defaultTimer = &systemTimer{}
)

// TODO(pedge): storing pointers not great, need a copy constructor
type runInfo struct {
	pipelineSource *pps.PipelineSource
	pipeline       *pps.Pipeline
}

type inMemoryClient struct {
	idToRunInfo      map[string]*runInfo
	idToRunStatuses  map[string][]*pps.PipelineRunStatus
	idToContainerIDs map[string]map[string]bool

	timer timer

	runInfoLock      *sync.RWMutex
	runStatusesLock  *sync.RWMutex
	containerIDsLock *sync.RWMutex
}

func newInMemoryClient() *inMemoryClient {
	return &inMemoryClient{
		make(map[string]*runInfo),
		make(map[string][]*pps.PipelineRunStatus),
		make(map[string]map[string]bool),
		defaultTimer,
		&sync.RWMutex{},
		&sync.RWMutex{},
		&sync.RWMutex{},
	}
}

func (c *inMemoryClient) AddPipelineRun(id string, pipelineSource *pps.PipelineSource, pipeline *pps.Pipeline) error {
	c.runInfoLock.Lock()
	defer c.runInfoLock.Unlock()
	c.runStatusesLock.Lock()
	defer c.runStatusesLock.Unlock()

	if _, ok := c.idToRunInfo[id]; ok {
		return fmt.Errorf("run with id %s already added", id)
	}
	c.idToRunInfo[id] = &runInfo{
		pipelineSource,
		pipeline,
	}
	c.idToRunStatuses[id] = make([]*pps.PipelineRunStatus, 1)
	c.idToRunStatuses[id][0] = &pps.PipelineRunStatus{
		PipelineRunStatusType: pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_ADDED,
		Timestamp:             timeToTimestamp(c.timer.Now()),
	}
	c.idToContainerIDs[id] = make(map[string]bool)
	return nil
}

func (c *inMemoryClient) GetPipelineRunPipelineSource(id string) (*pps.PipelineSource, error) {
	c.runInfoLock.RLock()
	defer c.runInfoLock.RUnlock()

	runInfo, ok := c.idToRunInfo[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return runInfo.pipelineSource, nil
}

func (c *inMemoryClient) GetPipelineRunPipeline(id string) (*pps.Pipeline, error) {
	c.runInfoLock.RLock()
	defer c.runInfoLock.RUnlock()

	runInfo, ok := c.idToRunInfo[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return runInfo.pipeline, nil
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

func (c *inMemoryClient) AddPipelineRunStatus(id string, pipelineRunStatusType pps.PipelineRunStatusType) error {
	c.runStatusesLock.Lock()
	defer c.runStatusesLock.Unlock()

	_, ok := c.idToRunStatuses[id]
	if !ok {
		return fmt.Errorf("no run for id %s", id)
	}
	c.idToRunStatuses[id] = append(c.idToRunStatuses[id], &pps.PipelineRunStatus{PipelineRunStatusType: pipelineRunStatusType, Timestamp: timeToTimestamp(c.timer.Now())})
	return nil
}

func (c *inMemoryClient) GetPipelineRunContainerIDs(id string) ([]string, error) {
	c.containerIDsLock.RLock()
	defer c.containerIDsLock.RUnlock()

	containerIDsMap, ok := c.idToContainerIDs[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	containerIDsSlice := make([]string, len(containerIDsMap))
	i := 0
	for containerID := range containerIDsMap {
		containerIDsSlice[i] = containerID
		i++
	}
	return containerIDsSlice, nil
}

func (c *inMemoryClient) AddPipelineRunContainerIDs(id string, containerIDs ...string) error {
	c.containerIDsLock.Lock()
	defer c.containerIDsLock.Unlock()

	containerIDsMap, ok := c.idToContainerIDs[id]
	if !ok {
		return fmt.Errorf("no run for id %s", id)
	}
	for _, containerID := range containerIDs {
		containerIDsMap[containerID] = true
	}
	return nil
}

func timeToTimestamp(t time.Time) *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{
		Seconds: t.UnixNano() / int64(time.Second),
		Nanos:   int32(t.UnixNano() % int64(time.Second)),
	}
}

type timer interface {
	Now() time.Time
}

type systemTimer struct{}

func (t *systemTimer) Now() time.Time {
	return time.Now().UTC()
}
