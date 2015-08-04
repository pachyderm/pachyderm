package store

import "github.com/pachyderm/pachyderm/src/pps"

type rethinkClient struct {
}

func newRethinkClient() *rethinkClient {
	return &rethinkClient{}
}

func (c *rethinkClient) AddPipelineRun(pipelineRun *pps.PipelineRun) error {
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
