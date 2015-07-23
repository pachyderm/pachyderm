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
	idToRunInfo     map[string]*runInfo
	idToRunStatuses map[string][]*pps.RunStatus
	timer           timer

	runInfoLock     *sync.RWMutex
	runStatusesLock *sync.RWMutex
}

func newInMemoryClient() *inMemoryClient {
	return &inMemoryClient{
		make(map[string]*runInfo),
		make(map[string][]*pps.RunStatus),
		defaultTimer,
		&sync.RWMutex{},
		&sync.RWMutex{},
	}
}

func (c *inMemoryClient) AddRun(id string, pipelineSource *pps.PipelineSource, pipeline *pps.Pipeline) error {
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
	c.idToRunStatuses[id] = make([]*pps.RunStatus, 1)
	c.idToRunStatuses[id][0] = &pps.RunStatus{
		RunStatusType: pps.RunStatusType_RUN_STATUS_TYPE_ADDED,
		Timestamp:     timeToTimestamp(c.timer.Now()),
	}
	return nil
}

func (c *inMemoryClient) GetRunPipelineSource(id string) (*pps.PipelineSource, error) {
	c.runInfoLock.RLock()
	defer c.runInfoLock.RUnlock()

	runInfo, ok := c.idToRunInfo[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return runInfo.pipelineSource, nil
}

func (c *inMemoryClient) GetRunPipeline(id string) (*pps.Pipeline, error) {
	c.runInfoLock.RLock()
	defer c.runInfoLock.RUnlock()

	runInfo, ok := c.idToRunInfo[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return runInfo.pipeline, nil
}

func (c *inMemoryClient) GetRunStatusLatest(id string) (*pps.RunStatus, error) {
	c.runStatusesLock.RLock()
	defer c.runStatusesLock.RUnlock()

	runStatuses, ok := c.idToRunStatuses[id]
	if !ok {
		return nil, fmt.Errorf("no run for id %s", id)
	}
	return runStatuses[len(runStatuses)-1], nil
}

func (c *inMemoryClient) AddRunStatus(id string, runStatusType pps.RunStatusType) error {
	c.runStatusesLock.Lock()
	defer c.runStatusesLock.Unlock()

	_, ok := c.idToRunStatuses[id]
	if !ok {
		return fmt.Errorf("no run for id %s", id)
	}
	c.idToRunStatuses[id] = append(c.idToRunStatuses[id], &pps.RunStatus{RunStatusType: runStatusType, Timestamp: timeToTimestamp(c.timer.Now())})
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
