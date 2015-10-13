package persist

import (
	"github.com/dancannon/gorethink"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/pkg/time"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
)

const (
	jobsTable        Table = "jobs"
	jobStatusesTable Table = "job_statuses"
	jobLogsTable     Table = "job_logs"
	pipelinesTable   Table = "pipelines"

	idPrimaryKey PrimaryKey = "id"

	pipelineIDIndex Index = "pipeline_id"
	jobIDIndex      Index = "job_id"
	typeIndex       Index = "type"
	nameIndex       Index = "name"
)

type Table string
type PrimaryKey string
type Index string

var (
	marshaller = &jsonpb.Marshaler{}

	tables = []Table{
		jobsTable,
		jobStatusesTable,
		jobLogsTable,
		pipelinesTable,
	}

	tableToTableCreateOpts = map[Table][]gorethink.TableCreateOpts{
		jobsTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: idPrimaryKey,
			},
		},
		jobStatusesTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: idPrimaryKey,
			},
		},
		jobLogsTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: idPrimaryKey,
			},
		},
		pipelinesTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: idPrimaryKey,
			},
		},
	}

	tableToIndexes = map[Table][]Index{
		jobsTable: []Index{
			pipelineIDIndex,
		},
		jobStatusesTable: []Index{
			jobIDIndex,
			typeIndex,
		},
		jobLogsTable: []Index{
			jobIDIndex,
		},
		pipelinesTable: []Index{
			nameIndex,
		},
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
func (a *rethinkAPIServer) CreateJob(ctx context.Context, request *Job) (*Job, error) {
	if request.Id != "" {
		return nil, ErrIDSet
	}
	request.Id = newID()
	if err := a.insertMessage(jobsTable, request); err != nil {
		return nil, err
	}
	// TODO(pedge): not transactional
	if _, err := a.CreateJobStatus(
		ctx,
		&JobStatus{
			JobId: request.Id,
			Type:  pps.JobStatusType_JOB_STATUS_TYPE_CREATED,
		},
	); err != nil {
		return nil, err
	}
	return request, nil
}

func (a *rethinkAPIServer) GetJobByID(ctx context.Context, request *google_protobuf.StringValue) (*Job, error) {
	job := &Job{}
	if err := a.getMessageByPrimaryKey(jobsTable, request.Value, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (a *rethinkAPIServer) GetJobsByPipelineID(ctx context.Context, request *google_protobuf.StringValue) (*Jobs, error) {
	jobObjs, err := a.getMessageByIndex(
		jobsTable,
		pipelineIDIndex,
		request.Value,
		func() proto.Message { return &Job{} },
	)
	if err != nil {
		return nil, err
	}
	jobs := make([]*Job, len(jobObjs))
	for i, jobObj := range jobObjs {
		jobs[i] = jobObj.(*Job)
	}
	return &Jobs{
		Job: jobs,
	}, nil
}

// id cannot be set
// timestamp cannot be set
func (a *rethinkAPIServer) CreateJobStatus(ctx context.Context, request *JobStatus) (*JobStatus, error) {
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
func (a *rethinkAPIServer) GetJobStatusesByJobID(ctx context.Context, request *google_protobuf.StringValue) (*JobStatuses, error) {
	return nil, nil
}

// id cannot be set
func (a *rethinkAPIServer) CreateJobLog(ctx context.Context, request *JobLog) (*JobLog, error) {
	return nil, nil
}

// ordered by time, latest to earliest
func (a *rethinkAPIServer) GetJobLogsByJobID(ctx context.Context, request *google_protobuf.StringValue) (*JobLogs, error) {
	return nil, nil
}

// id and previous_id cannot be set
// name must not already exist
func (a *rethinkAPIServer) CreatePipeline(ctx context.Context, request *Pipeline) (*Pipeline, error) {
	return nil, nil
}

// id and previous_id cannot be set
// timestamp cannot be set
// update by name, name must already exist
func (a *rethinkAPIServer) UpdatePipeline(ctx context.Context, request *Pipeline) (*Pipeline, error) {
	return nil, nil
}

func (a *rethinkAPIServer) GetPipelineByID(ctx context.Context, request *google_protobuf.StringValue) (*Pipeline, error) {
	return nil, nil
}

// ordered by time, latest to earliest
func (a *rethinkAPIServer) GetPipelinesByName(ctx context.Context, request *google_protobuf.StringValue) (*Pipeline, error) {
	return nil, nil
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

func (a *rethinkAPIServer) getMessageByIndex(
	table Table,
	index Index,
	value interface{},
	messageConstructor func() proto.Message,
	modifiers ...func(gorethink.Term) gorethink.Term,
) ([]interface{}, error) {
	term := a.getTerm(table).GetAllByIndex(index, value).Map(func(row gorethink.Term) interface{} {
		return row.ToJSON()
	})
	for _, modifier := range modifiers {
		term = modifier(term)
	}
	cursor, err := term.Run(a.session)
	if err != nil {
		return nil, err
	}
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

func (a *rethinkAPIServer) getTerm(table Table) gorethink.Term {
	return gorethink.DB(a.databaseName).Table(table)
}

func (a *rethinkAPIServer) now() *google_protobuf.Timestamp {
	return prototime.TimeToTimestamp(a.timer.Now())
}

func rethinkToJSON(row gorethink.Term) interface{} {
	return row.ToJSON()
}

func newID() string {
	return uuid.NewV4().String()
}
