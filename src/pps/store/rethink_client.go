package store

import (
	"github.com/dancannon/gorethink"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/pps"
)

const (
	runTable       = "pipeline_runs"
	statusTable    = "pipeline_run_statuses"
	containerTable = "pipeline_containers"
)

var (
	marshaller = &jsonpb.Marshaller{}
)

// InitDBs prepares a RethinkDB instance to be used by rethinkClient.
// rethinkClients will error if they are pointed at databases that haven't had
// InitDBs run on them
// InitDBs should only be run once per instance of RethinkDB, it will error if
// it's called a second time.
func InitDBs(address string, databaseName string) error {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: address})
	if err != nil {
		return err
	}
	if _, err := gorethink.DBCreate(databaseName).RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).TableCreate(runTable).RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).TableCreate(statusTable).RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).TableCreate(
		containerTable,
		gorethink.TableCreateOpts{
			PrimaryKey: "container_id",
		},
	).RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).Table(statusTable).
		IndexCreate("pipeline_run_id").RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).Table(containerTable).
		IndexCreate("pipeline_run_id").RunWrite(session); err != nil {
		return err
	}
	return nil
}

type rethinkClient struct {
	session      *gorethink.Session
	databaseName string
	timer        timer
	runs         gorethink.Term
	statuses     gorethink.Term
	containers   gorethink.Term
}

func newRethinkClient(address string, databaseName string) (*rethinkClient, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: address})
	if err != nil {
		return nil, err
	}
	return &rethinkClient{
		session,
		databaseName,
		defaultTimer,
		gorethink.DB(databaseName).Table(runTable),
		gorethink.DB(databaseName).Table(statusTable),
		gorethink.DB(databaseName).Table(containerTable),
	}, nil
}

func (c *rethinkClient) Close() error {
	return c.session.Close()
}

func (c *rethinkClient) AddPipelineRun(pipelineRun *pps.PipelineRun) error {
	data, err := marshaller.MarshalToString(pipelineRun)
	if err != nil {
		return err
	}
	_, err = c.runs.Insert(gorethink.JSON(data)).RunWrite(c.session)
	return err
}

func (c *rethinkClient) GetPipelineRun(id string) (*pps.PipelineRun, error) {
	cursor, err := c.runs.Get(id).ToJSON().Run(c.session)
	if err != nil {
		return nil, err
	}
	data := ""
	if !cursor.Next(&data) {
		return nil, cursor.Err()
	}
	var pipelineRun pps.PipelineRun
	if err := jsonpb.UnmarshalString(data, &pipelineRun); err != nil {
		return nil, err
	}
	return &pipelineRun, nil
}

func (c *rethinkClient) AddPipelineRunStatus(id string, statusType pps.PipelineRunStatusType) error {
	runStatus := &pps.PipelineRunStatus{
		PipelineRunId:         id,
		PipelineRunStatusType: statusType,
		Timestamp:             timeToTimestamp(c.timer.Now()),
	}
	data, err := marshaller.MarshalToString(runStatus)
	if err != nil {
		return err
	}
	_, err = c.statuses.Insert(gorethink.JSON(data)).RunWrite(c.session)
	return err
}

func (c *rethinkClient) GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error) {
	cursor, err := c.statuses.
		GetAllByIndex("pipeline_run_id", id).
		OrderBy(gorethink.Desc("timestamp")).
		Nth(0).
		ToJSON().
		Run(c.session)
	if err != nil {
		return nil, err
	}
	data := ""
	if !cursor.Next(&data) {
		return nil, cursor.Err()
	}
	var pipelineRunStatus pps.PipelineRunStatus
	if err := jsonpb.UnmarshalString(data, &pipelineRunStatus); err != nil {
		return nil, err
	}
	return &pipelineRunStatus, nil
}

func (c *rethinkClient) AddPipelineRunContainerIDs(id string, containerIDs ...string) error {
	var pipelineContainers []gorethink.Term
	for _, containerID := range containerIDs {
		pipelineContainer := PipelineContainer{id, containerID}
		data, err := marshaller.MarshalToString(&pipelineContainer)
		if err != nil {
			return err
		}
		pipelineContainers = append(pipelineContainers, gorethink.JSON(data))
	}
	if _, err := c.containers.
		Insert(pipelineContainers).
		RunWrite(c.session); err != nil {
		return err
	}
	return nil
}

func (c *rethinkClient) GetPipelineRunContainers(id string) ([]*PipelineContainer, error) {
	cursor, err := c.containers.
		GetAllByIndex("pipeline_run_id", id).
		Map(func(row gorethink.Term) interface{} {
		return row.ToJSON()
	}).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	var pipelineContainers []string
	if err := cursor.All(&pipelineContainers); err != nil {
		return nil, err
	}
	var result []*PipelineContainer
	for _, data := range pipelineContainers {
		var pipelineContainer PipelineContainer
		if err := jsonpb.UnmarshalString(data, &pipelineContainer); err != nil {
			return nil, err
		}
		result = append(result, &pipelineContainer)
	}
	return result, nil
}
