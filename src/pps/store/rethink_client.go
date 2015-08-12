package store

import (
	"github.com/dancannon/gorethink"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/pkg/protoutil"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
)

const (
	runTable       = "pipeline_runs"
	statusTable    = "pipeline_run_statuses"
	containerTable = "pipeline_containers"
	logTable       = "pipeline_logs"
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
	if _, err := gorethink.DB(databaseName).TableCreate(logTable).RunWrite(session); err != nil {
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
	if _, err := gorethink.DB(databaseName).Table(logTable).
		IndexCreate("pipeline_run_id").RunWrite(session); err != nil {
		return err
	}
	return nil
}

type rethinkClient struct {
	session      *gorethink.Session
	databaseName string
	timer        timing.Timer
	runs         gorethink.Term
	statuses     gorethink.Term
	containers   gorethink.Term
	logs         gorethink.Term
}

func newRethinkClient(address string, databaseName string) (*rethinkClient, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: address})
	if err != nil {
		return nil, err
	}
	return &rethinkClient{
		session,
		databaseName,
		timing.NewSystemTimer(),
		gorethink.DB(databaseName).Table(runTable),
		gorethink.DB(databaseName).Table(statusTable),
		gorethink.DB(databaseName).Table(containerTable),
		gorethink.DB(databaseName).Table(logTable),
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
		Timestamp:             protoutil.TimeToTimestamp(c.timer.Now()),
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

func (c *rethinkClient) AddPipelineRunContainers(containers ...*pps.PipelineRunContainer) error {
	var pipelineContainers []gorethink.Term
	for _, pipelineContainer := range containers {
		data, err := marshaller.MarshalToString(pipelineContainer)
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

func (c *rethinkClient) GetPipelineRunContainers(id string) ([]*pps.PipelineRunContainer, error) {
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
	var result []*pps.PipelineRunContainer
	for _, data := range pipelineContainers {
		var pipelineContainer pps.PipelineRunContainer
		if err := jsonpb.UnmarshalString(data, &pipelineContainer); err != nil {
			return nil, err
		}
		result = append(result, &pipelineContainer)
	}
	return result, nil
}

func (c *rethinkClient) AddPipelineRunLogs(logs ...*pps.PipelineRunLog) error {
	var pipelineLogs []gorethink.Term
	for _, pipelineLog := range logs {
		data, err := marshaller.MarshalToString(pipelineLog)
		if err != nil {
			return err
		}
		pipelineLogs = append(pipelineLogs, gorethink.JSON(data))
	}
	if _, err := c.logs.
		Insert(pipelineLogs).
		RunWrite(c.session); err != nil {
		return err
	}
	return nil
}

func (c *rethinkClient) GetPipelineRunLogs(id string) ([]*pps.PipelineRunLog, error) {
	cursor, err := c.logs.
		GetAllByIndex("pipeline_run_id", id).
		Map(func(row gorethink.Term) interface{} {
		return row.ToJSON()
	}).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	var pipelineLogs []string
	if err := cursor.All(&pipelineLogs); err != nil {
		return nil, err
	}
	var result []*pps.PipelineRunLog
	for _, data := range pipelineLogs {
		var pipelineLog pps.PipelineRunLog
		if err := jsonpb.UnmarshalString(data, &pipelineLog); err != nil {
			return nil, err
		}
		result = append(result, &pipelineLog)
	}
	return result, nil
}

func (c *rethinkClient) AddPfsCommitMapping(pfsCommitMapping *pps.PfsCommitMapping) error {
	return nil
}

func (c *rethinkClient) GetPfsCommitMappingLatest(inputRepositoryName string, inputCommitID string) (*pps.PfsCommitMapping, error) {
	return nil, nil
}
