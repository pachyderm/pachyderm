package server

import (
	"fmt"

	"github.com/dancannon/gorethink"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/pkg/time"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
)

const (
	jobInfosTable      Table = "job_infos"
	jobStatusesTable   Table = "job_statuses"
	jobOutputsTable    Table = "job_outputs"
	jobLogsTable       Table = "job_logs"
	pipelineInfosTable Table = "pipeline_infos"

	idPrimaryKey       PrimaryKey = "id"
	jobPrimaryKey      PrimaryKey = "job.id"
	pipelinePrimaryKey PrimaryKey = "pipeline.name"

	pipelineIndex Index = "pipeline.name"
	jobIndex      Index = "job.id"
	typeIndex     Index = "type"
)

type Table string
type PrimaryKey string
type Index string

var (
	marshaller = &jsonpb.Marshaler{}

	tables = []Table{
		jobInfosTable,
		jobStatusesTable,
		jobOutputsTable,
		jobLogsTable,
		pipelineInfosTable,
	}

	tableToTableCreateOpts = map[Table][]gorethink.TableCreateOpts{
		jobInfosTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: jobPrimaryKey,
			},
		},
		jobStatusesTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: idPrimaryKey,
			},
		},
		jobOutputsTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: jobPrimaryKey,
			},
		},
		jobLogsTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: idPrimaryKey,
			},
		},
		pipelineInfosTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: pipelinePrimaryKey,
			},
		},
	}

	tableToIndexes = map[Table][]Index{
		jobInfosTable: []Index{
			pipelineIndex,
		},
		jobStatusesTable: []Index{
			jobIndex,
			typeIndex,
		},
		jobLogsTable: []Index{
			jobIndex,
		},
		pipelineInfosTable: []Index{},
	}
)

// InitDBs prepares a RethinkDB instance to be used by the rethink server.
// Rethink servers will error if they are pointed at databases that haven't had InitDBs run on them.
// InitDBs should only be run once per instance of RethinkDB, it will error if it is called a second time.
func InitDBs(address string, databaseName string) error {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: address})
	if err != nil {
		return err
	}
	if _, err := gorethink.DBCreate(databaseName).RunWrite(session); err != nil {
		return err
	}
	for _, table := range tables {
		tableCreateOpts, ok := tableToTableCreateOpts[table]
		if ok {
			if _, err := gorethink.DB(databaseName).TableCreate(table, tableCreateOpts...).RunWrite(session); err != nil {
				return err
			}
		} else {
			if _, err := gorethink.DB(databaseName).TableCreate(table).RunWrite(session); err != nil {
				return err
			}
		}
	}
	for table, indexes := range tableToIndexes {
		for _, index := range indexes {
			if _, err := gorethink.DB(databaseName).Table(table).IndexCreate(index).RunWrite(session); err != nil {
				return err
			}
		}
	}
	return nil
}

type rethinkAPIServer struct {
	session      *gorethink.Session
	databaseName string
	timer        pkgtime.Timer
}

func newRethinkAPIServer(address string, databaseName string) (*rethinkAPIServer, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: address})
	if err != nil {
		return nil, err
	}
	return &rethinkAPIServer{
		session,
		databaseName,
		pkgtime.NewSystemTimer(),
	}, nil
}

func (a *rethinkAPIServer) Close() error {
	return a.session.Close()
}

// id cannot be set
// a JobStatus of type created will also be created
func (a *rethinkAPIServer) CreateJobInfo(ctx context.Context, request *persist.JobInfo) (*persist.JobInfo, error) {
	if request.CreatedAt != nil {
		return nil, ErrTimestampSet
	}
	request.Id = newID()
	request.CreatedAt = a.now()
	if err := a.insertMessage(jobInfosTable, request); err != nil {
		return nil, err
	}
	// TODO(pedge): not transactional
	if _, err := a.CreateJobStatus(
		ctx,
		&persist.JobStatus{
			JobId: request.Id,
			Type:  pps.JobStatusType_JOB_STATUS_TYPE_CREATED,
		},
	); err != nil {
		return nil, err
	}
	return request, nil
}

func (a *rethinkAPIServer) GetJobByID(ctx context.Context, request *google_protobuf.StringValue) (*persist.Job, error) {
	job := &persist.Job{}
	if err := a.getMessageByPrimaryKey(jobInfosTable, request.Value, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (a *rethinkAPIServer) GetJobsByPipelineID(ctx context.Context, request *google_protobuf.StringValue) (*persist.Jobs, error) {
	jobObjs, err := a.getMessagesByIndex(
		jobInfosTable,
		pipelineIDIndex,
		request.Value,
		func() proto.Message { return &persist.Job{} },
		func(term gorethink.Term) gorethink.Term {
			return term.OrderBy(gorethink.Desc("created_at"))
		},
	)
	if err != nil {
		return nil, err
	}
	jobs := make([]*persist.Job, len(jobObjs))
	for i, jobObj := range jobObjs {
		jobs[i] = jobObj.(*persist.Job)
	}
	return &persist.Jobs{
		Job: jobs,
	}, nil
}

// id cannot be set
// timestamp cannot be set
func (a *rethinkAPIServer) CreateJobStatus(ctx context.Context, request *persist.JobStatus) (*persist.JobStatus, error) {
	if request.Id != "" {
		return nil, ErrIDSet
	}
	if request.Timestamp != nil {
		return nil, ErrTimestampSet
	}
	request.Id = newID()
	request.Timestamp = a.now()
	if err := a.insertMessage(jobStatusesTable, request); err != nil {
		return nil, err
	}
	return request, nil
}

// ordered by time, latest to earliest
func (a *rethinkAPIServer) GetJobStatusesByJobID(ctx context.Context, request *google_protobuf.StringValue) (*persist.JobStatuses, error) {
	jobStatusObjs, err := a.getMessagesByIndex(
		jobStatusesTable,
		jobIDIndex,
		request.Value,
		func() proto.Message { return &persist.JobStatus{} },
		func(term gorethink.Term) gorethink.Term {
			return term.OrderBy(gorethink.Desc("timestamp"))
		},
	)
	if err != nil {
		return nil, err
	}
	jobStatuses := make([]*persist.JobStatus, len(jobStatusObjs))
	for i, jobStatusObj := range jobStatusObjs {
		jobStatuses[i] = jobStatusObj.(*persist.JobStatus)
	}
	return &persist.JobStatuses{
		JobStatus: jobStatuses,
	}, nil
}

// id cannot be set
// timestamp cannot be set
func (a *rethinkAPIServer) CreateJobLog(ctx context.Context, request *persist.JobLog) (*persist.JobLog, error) {
	if request.Id != "" {
		return nil, ErrIDSet
	}
	if request.Timestamp != nil {
		return nil, ErrTimestampSet
	}
	request.Id = newID()
	request.Timestamp = a.now()
	if err := a.insertMessage(jobLogsTable, request); err != nil {
		return nil, err
	}
	return request, nil
}

// ordered by time, latest to earliest
func (a *rethinkAPIServer) GetJobLogsByJobID(ctx context.Context, request *google_protobuf.StringValue) (*persist.JobLogs, error) {
	jobLogObjs, err := a.getMessagesByIndex(
		jobLogsTable,
		jobIDIndex,
		request.Value,
		func() proto.Message { return &persist.JobLog{} },
		func(term gorethink.Term) gorethink.Term {
			return term.OrderBy(gorethink.Desc("timestamp"))
		},
	)
	if err != nil {
		return nil, err
	}
	jobLogs := make([]*persist.JobLog, len(jobLogObjs))
	for i, jobLogObj := range jobLogObjs {
		jobLogs[i] = jobLogObj.(*persist.JobLog)
	}
	return &persist.JobLogs{
		JobLog: jobLogs,
	}, nil
}

// id and previous_id cannot be set
// name must not already exist
func (a *rethinkAPIServer) CreatePipeline(ctx context.Context, request *persist.Pipeline) (*persist.Pipeline, error) {
	pipeline, err := a.getLastPipeline(ctx, request.Name)
	if err != nil {
		return nil, err
	}
	if pipeline != nil {
		return nil, fmt.Errorf("pachyderm.pps.persist.server: pipeline already exists with name %s", request.Name)
	}
	return a.createPipeline(ctx, request)
}

// id and previous_id cannot be set
// timestamp cannot be set
// update by name, name must already exist
func (a *rethinkAPIServer) UpdatePipeline(ctx context.Context, request *persist.Pipeline) (*persist.Pipeline, error) {
	pipeline, err := a.getLastPipeline(ctx, request.Name)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, fmt.Errorf("pachyderm.pps.persist.server: pipeline does not exist with name %s", request.Name)
	}
	request.PreviousId = pipeline.Id
	return a.createPipeline(ctx, request)
}

func (a *rethinkAPIServer) getLastPipeline(ctx context.Context, name string) (*persist.Pipeline, error) {
	// TODO(pedge): not transactional
	pipelines, err := a.GetPipelinesByName(ctx, &google_protobuf.StringValue{Value: name})
	if err != nil {
		return nil, err
	}
	if len(pipelines.Pipeline) == 0 {
		return nil, nil
	}
	return pipelines.Pipeline[0], nil
}

func (a *rethinkAPIServer) createPipeline(ctx context.Context, request *persist.Pipeline) (*persist.Pipeline, error) {
	if request.Id != "" {
		return nil, ErrIDSet
	}
	if request.CreatedAt != nil {
		return nil, ErrTimestampSet
	}
	request.Id = newID()
	request.CreatedAt = a.now()
	if err := a.insertMessage(pipelineInfosTable, request); err != nil {
		return nil, err
	}
	return request, nil
}

func (a *rethinkAPIServer) GetPipelineByID(ctx context.Context, request *google_protobuf.StringValue) (*persist.Pipeline, error) {
	pipeline := &persist.Pipeline{}
	if err := a.getMessageByPrimaryKey(pipelineInfosTable, request.Value, pipeline); err != nil {
		return nil, err
	}
	return pipeline, nil
}

// ordered by time, latest to earliest
func (a *rethinkAPIServer) GetPipelinesByName(ctx context.Context, request *google_protobuf.StringValue) (*persist.Pipelines, error) {
	pipelineObjs, err := a.getMessagesByIndex(
		pipelineInfosTable,
		nameIndex,
		request.Value,
		func() proto.Message { return &persist.Pipeline{} },
		func(term gorethink.Term) gorethink.Term {
			return term.OrderBy(gorethink.Desc("created_at"))
		},
	)
	if err != nil {
		return nil, err
	}
	pipelines := make([]*persist.Pipeline, len(pipelineObjs))
	for i, pipelineObj := range pipelineObjs {
		pipelines[i] = pipelineObj.(*persist.Pipeline)
	}
	return &persist.Pipelines{
		Pipeline: pipelines,
	}, nil
}

func (a *rethinkAPIServer) GetAllPipelines(ctx context.Context, request *google_protobuf.Empty) (*persist.Pipelines, error) {
	pipelineObjs, err := a.getAllMessages(
		pipelineInfosTable,
		func() proto.Message { return &persist.Pipeline{} },
		func(term gorethink.Term) gorethink.Term {
			return term.OrderBy(gorethink.Desc("created_at"))
		},
	)
	if err != nil {
		return nil, err
	}
	pipelines := make([]*persist.Pipeline, len(pipelineObjs))
	for i, pipelineObj := range pipelineObjs {
		pipelines[i] = pipelineObj.(*persist.Pipeline)
	}
	return &persist.Pipelines{
		Pipeline: pipelines,
	}, nil
}

func (a *rethinkAPIServer) insertMessage(table Table, message proto.Message) error {
	data, err := marshaller.MarshalToString(message)
	if err != nil {
		return err
	}
	_, err = a.getTerm(table).Insert(gorethink.JSON(data)).RunWrite(a.session)
	return err
}

func (a *rethinkAPIServer) getMessageByPrimaryKey(table Table, Value interface{}, message proto.Message) error {
	cursor, err := a.getTerm(table).Get(Value).ToJSON().Run(a.session)
	if err != nil {
		return err
	}
	data := ""
	if !cursor.Next(&data) {
		return cursor.Err()
	}
	if err := jsonpb.UnmarshalString(data, message); err != nil {
		return err
	}
	return nil
}

func (a *rethinkAPIServer) getMessagesByIndex(
	table Table,
	index Index,
	value interface{},
	messageConstructor func() proto.Message,
	modifiers ...func(gorethink.Term) gorethink.Term,
) ([]interface{}, error) {
	return a.getMultiple(
		a.getTerm(table).GetAllByIndex(index, value),
		messageConstructor,
		modifiers...,
	)
}

func (a *rethinkAPIServer) getAllMessages(
	table Table,
	messageConstructor func() proto.Message,
	modifiers ...func(gorethink.Term) gorethink.Term,
) ([]interface{}, error) {
	return a.getMultiple(
		a.getTerm(table),
		messageConstructor,
		modifiers...,
	)
}

func (a *rethinkAPIServer) getMultiple(
	term gorethink.Term,
	messageConstructor func() proto.Message,
	modifiers ...func(gorethink.Term) gorethink.Term,
) ([]interface{}, error) {
	for _, modifier := range modifiers {
		term = modifier(term)
	}
	term = term.Map(func(row gorethink.Term) interface{} {
		return row.ToJSON()
	})
	cursor, err := term.Run(a.session)
	if err != nil {
		return nil, err
	}
	return processMultipleCursor(cursor, messageConstructor)
}

func (a *rethinkAPIServer) getTerm(table Table) gorethink.Term {
	return gorethink.DB(a.databaseName).Table(table)
}

func (a *rethinkAPIServer) now() *google_protobuf.Timestamp {
	return prototime.TimeToTimestamp(a.timer.Now())
}

func processMultipleCursor(
	cursor *gorethink.Cursor,
	messageConstructor func() proto.Message,
) ([]interface{}, error) {
	var data []string
	if err := cursor.All(&data); err != nil {
		return nil, err
	}
	result := make([]interface{}, len(data))
	for i, datum := range data {
		message := messageConstructor()
		if err := jsonpb.UnmarshalString(datum, message); err != nil {
			return nil, err
		}
		result[i] = message
	}
	return result, nil
}

func rethinkToJSON(row gorethink.Term) interface{} {
	return row.ToJSON()
}

func newID() string {
	return uuid.NewV4().String()
}
