package store

import (
	"github.com/pachyderm/pachyderm/src/pps"
)

type rethinkClient struct {
}

func newRethinkClient() *rethinkClient {
	return &rethinkClient{}
}

func (c *rethinkClient) AddPipelineRun(pipelineRun *pps.PipelineRun) error {
	// will look like this:
	// data, err := jsonpb.Marshal(pipelineRun) // marshals directly
	// gorethink.Database(dbName).Table("pipeline_runs").Insert(data).RunWrite(session)
	return nil
}

func (c *rethinkClient) GetPipelineRun(id string) (*pps.PipelineRun, error) {
	return nil, nil
}

func (c *rethinkClient) GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error) {
	return nil, nil
}

func (c *rethinkClient) AddPipelineRunStatus(id string, runStatusType pps.PipelineRunStatusType) error {
	return nil
}

func (c *rethinkClient) GetPipelineRunContainerIDs(id string) ([]string, error) {
	return nil, nil
}

func (c *rethinkClient) AddPipelineRunContainerIDs(id string, containerIDs ...string) error {
	return nil
}
