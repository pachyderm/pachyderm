package store

import (
	"time"

	"github.com/dancannon/gorethink"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/pps"
)

type rethinkClient struct {
	session    *gorethink.Session
	marshaller *jsonpb.Marshaller
}

func newRethinkClient(address string) (*rethinkClient, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: address})
	if err != nil {
		return nil, err
	}
	if _, err := gorethink.DBCreate("pachyderm").RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB("pachyderm").TableCreate("pipeline_runs").RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB("pachyderm").TableCreate("pipeline_status").RunWrite(session); err != nil {
		return nil, err
	}
	return &rethinkClient{session, &jsonpb.Marshaller{EnumsAsString: true}}, nil
}

func (c *rethinkClient) AddPipelineRun(pipelineRun *pps.PipelineRun) error {
	data, err := c.marshaller.MarshalToString(pipelineRun)
	if err != nil {
		return err
	}
	_, err = gorethink.DB("pachyderm").Table("pipeline_runs").Insert(gorethink.JSON(data)).RunWrite(c.session)
	return err
}

func (c *rethinkClient) GetPipelineRun(id string) (*pps.PipelineRun, error) {
	data := ""
	cursor, err := gorethink.DB("pachyderm").Table("pipeline_runs").Get(id).ToJSON().Run(c.session)
	if err != nil {
		return nil, err
	}
	if !cursor.Next(&data) {
		return nil, cursor.Err()
	}
	var result pps.PipelineRun
	jsonpb.UnmarshalString(data, &result)
	return &result, nil
}

type pipelineRunStatus struct {
	pipelineRun           string                    `gorethink:"pipeline_run,omitempty"`
	pipelineRunStatusType pps.PipelineRunStatusType `gorethink:"pipeline_run_status_type,omitempty"`
	timestamp             gorethink.Term            `gorethink:"timestamp,omitempty"`
}

func (c *rethinkClient) AddPipelineRunStatus(id string, runStatusType pps.PipelineRunStatusType) error {
	doc := pipelineRunStatus{
		id,
		runStatusType,
		gorethink.Time(time.Now().Unix()),
	}
	_, err := gorethink.DB("pachyderm").Table("pipeline_status").Insert(doc).RunWrite(c.session)
	return err
}

type pipelineContainer struct {
	id          string `gorethink:"id,omitempty"`
	pipelineRun string `gorethink:"pipeline_run,omitempty"`
}

func (c *rethinkClient) GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error) {
	var result pps.PipelineRunStatus
	cursor, err := gorethink.DB("pachyderm").Table("pipeline_status").
		GetAllByIndex("pipeline_run", id).
		OrderBy(gorethink.Desc("timestamp")).
		Nth(0).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	if !cursor.Next(&result) {
		return nil, cursor.Err()
	}
	return &result, nil
}

func (c *rethinkClient) AddPipelineRunContainerIDs(id string, containerIDs ...string) error {
	var toInsert []pipelineContainer
	for _, containerID := range containerIDs {
		toInsert = append(toInsert, pipelineContainer{containerID, id})
	}
	if _, err := gorethink.DB("pachyderm").Table("pipeline_container").
		Insert(toInsert).
		RunWrite(c.session); err != nil {
		return err
	}
	return nil
}

func (c *rethinkClient) GetPipelineRunContainerIDs(id string) ([]string, error) {
	cursor, err := gorethink.DB("pachyderm").Table("pipeline_status").
		GetAllByIndex("pipeline_run", id).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	var containers []pipelineContainer
	if err := cursor.All(&containers); err != nil {
		return nil, err
	}
	var result []string
	for _, container := range containers {
		result = append(result, container.id)
	}
	return result, nil
}
