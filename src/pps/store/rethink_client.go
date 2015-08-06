package store

import (
	"github.com/dancannon/gorethink"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/peter-edge/go-google-protobuf"
)

var (
	marshaller = &jsonpb.Marshaller{EnumsAsString: true}
)

type pipelineRunStatus struct {
	PipelineRunID         string                    `gorethink:"pipeline_run_id,omitempty"`
	PipelineRunStatusType pps.PipelineRunStatusType `gorethink:"pipeline_run_status_type,omitempty"`
	Seconds               int64                     `gorethink:"seconds,omitempty"`
	Nanos                 int32                     `gorethink:"nanos,omitempty"`
}

type pipelineContainer struct {
	PipelineRunID string `gorethink:"pipeline_run_id,omitempty"`
	ContainerID   string `gorethink:"container_id,omitempty"`
}

type rethinkClient struct {
	session      *gorethink.Session
	databaseName string
	timer        timer
}

func newRethinkClient(address string, databaseName string) (*rethinkClient, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: address})
	if err != nil {
		return nil, err
	}
	if _, err := gorethink.DBCreate(databaseName).RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB(databaseName).TableCreate("pipeline_runs").RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB(databaseName).TableCreate("pipeline_run_statuses").RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB(databaseName).TableCreate(
		"pipeline_containers",
		gorethink.TableCreateOpts{
			PrimaryKey: "container_id",
		},
	).RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB(databaseName).Table("pipeline_run_statuses").
		IndexCreate("pipeline_run_id").RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB(databaseName).Table("pipeline_containers").
		IndexCreate("pipeline_run_id").RunWrite(session); err != nil {
		return nil, err
	}
	return &rethinkClient{
		session,
		databaseName,
		defaultTimer,
	}, nil
}

func (c *rethinkClient) AddPipelineRun(pipelineRun *pps.PipelineRun) error {
	data, err := marshaller.MarshalToString(pipelineRun)
	if err != nil {
		return err
	}
	_, err = gorethink.DB(c.databaseName).Table("pipeline_runs").Insert(gorethink.JSON(data)).RunWrite(c.session)
	return err
}

func (c *rethinkClient) GetPipelineRun(id string) (*pps.PipelineRun, error) {
	data := ""
	cursor, err := gorethink.DB(c.databaseName).Table("pipeline_runs").Get(id).ToJSON().Run(c.session)
	if err != nil {
		return nil, err
	}
	if !cursor.Next(&data) {
		return nil, cursor.Err()
	}
	var pipelineRun pps.PipelineRun
	if err := jsonpb.UnmarshalString(data, &pipelineRun); err != nil {
		return nil, err
	}
	return &pipelineRun, nil
}

func (c *rethinkClient) AddPipelineRunStatus(id string, runStatusType pps.PipelineRunStatusType) error {
	now := timeToTimestamp(c.timer.Now())
	doc := pipelineRunStatus{
		id,
		runStatusType,
		now.Seconds,
		now.Nanos,
	}
	_, err := gorethink.DB(c.databaseName).Table("pipeline_run_statuses").Insert(doc).RunWrite(c.session)
	return err
}

func (c *rethinkClient) GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error) {
	var doc pipelineRunStatus
	cursor, err := gorethink.DB(c.databaseName).Table("pipeline_run_statuses").
		GetAllByIndex("pipeline_run_id", id).
		OrderBy(gorethink.Desc("seconds"), gorethink.Desc("nanos")).
		Nth(0).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	if !cursor.Next(&doc) {
		return nil, cursor.Err()
	}
	var pipelineRunStatus pps.PipelineRunStatus
	pipelineRunStatus.PipelineRunStatusType = doc.PipelineRunStatusType
	pipelineRunStatus.Timestamp = &google_protobuf.Timestamp{Seconds: doc.Seconds, Nanos: doc.Nanos}
	return &pipelineRunStatus, nil
}

func (c *rethinkClient) AddPipelineRunContainerIDs(id string, containerIDs ...string) error {
	var pipelineContainers []pipelineContainer
	for _, containerID := range containerIDs {
		pipelineContainers = append(pipelineContainers, pipelineContainer{id, containerID})
	}
	if _, err := gorethink.DB(c.databaseName).Table("pipeline_containers").
		Insert(pipelineContainers).
		RunWrite(c.session); err != nil {
		return err
	}
	return nil
}

func (c *rethinkClient) GetPipelineRunContainerIDs(id string) ([]string, error) {
	cursor, err := gorethink.DB(c.databaseName).Table("pipeline_containers").
		GetAllByIndex("pipeline_run_id", id).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	var pipelineContainers []pipelineContainer
	if err := cursor.All(&pipelineContainers); err != nil {
		return nil, err
	}
	var containerIDs []string
	for _, pipelineContainer := range pipelineContainers {
		containerIDs = append(containerIDs, pipelineContainer.ContainerID)
	}
	return containerIDs, nil
}
