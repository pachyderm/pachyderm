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
)

type Table string
type Index string

var (
	marshaller = &jsonpb.Marshaler{}

	tables = []Table{
		jobsTable,
		jobStatusesTable,
		jobLogsTable,
		pipelinesTable,
	}

	tableToTableCreateOpts = map[Table][]gorethink.TableCreateOpts{}

	tableToIndexes = map[Table][]Index{}
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
	if err := a.insertMessage(
		jobStatusesTable,
		&JobStatus{
			Id:        newID(),
			JobId:     request.Id,
			Type:      pps.JobStatusType_JOB_STATUS_TYPE_CREATED,
			Timestamp: a.now(),
		},
	); err != nil {
		return nil, err
	}
	return request, nil
}

func (a *rethinkAPIServer) GetJobByID(ctx context.Context, request *google_protobuf.StringValue) (*Job, error) {
	return nil, nil
}

func (a *rethinkAPIServer) GetJobsByPipelineID(ctx context.Context, request *google_protobuf.StringValue) (*Jobs, error) {
	return nil, nil
}

// id cannot be set
func (a *rethinkAPIServer) CreateJobStatus(ctx context.Context, request *JobStatus) (*JobStatus, error) {
	return nil, nil
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

func (a *rethinkAPIServer) getMessageByID(table Table, id string, message proto.Message) error {
	cursor, err := a.getTerm(table).Get(id).ToJSON().Run(a.session)
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
