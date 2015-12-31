package server

import (
	"sort"
	"time"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/pkg/time"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"

	"github.com/dancannon/gorethink"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
)

const (
	jobInfosTable      Table = "job_infos"
	pipelineInfosTable Table = "pipeline_infos"

	pipelineNameIndex         Index = "pipeline_name"
	pipelineNameAndInputIndex Index = "pipeline_name_and_input"
	inputIndex                Index = "input"
)

type Table string
type PrimaryKey string
type Index string

var (
	marshaller = &jsonpb.Marshaler{}

	tables = []Table{
		jobInfosTable,
		pipelineInfosTable,
	}

	tableToTableCreateOpts = map[Table][]gorethink.TableCreateOpts{
		jobInfosTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: "job_id",
			},
		},
		pipelineInfosTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: "pipeline_name",
			},
		},
	}
)

// InitDBs prepares a RethinkDB instance to be used by the rethink server.
// Rethink servers will error if they are pointed at databases that haven't had InitDBs run on them.
// InitDBs is idempotent (unless rethink dies in the middle of the function)
func InitDBs(address string, databaseName string) error {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: address})
	if err != nil {
		return err
	}
	if _, err := gorethink.DBCreate(databaseName).RunWrite(session); err != nil {
		if _, ok := err.(gorethink.RQLRuntimeError); ok {
			return nil
		}
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

	// Create some indexes for the jobInfosTable
	if _, err := gorethink.DB(databaseName).Table(jobInfosTable).IndexCreate(pipelineNameIndex).RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).Table(jobInfosTable).IndexCreateFunc(
		pipelineNameAndInputIndex,
		func(row gorethink.Term) interface{} {
			return []interface{}{
				row.Field("pipeline_name"),
				row.Field("input_index"),
			}
		}).RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).Table(jobInfosTable).IndexCreateFunc(
		inputIndex,
		func(row gorethink.Term) interface{} {
			return row.Field("input_index")
		}).RunWrite(session); err != nil {
		return err
	}
	return nil
}

type rethinkAPIServer struct {
	protorpclog.Logger
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
		protorpclog.NewLogger("pachyderm.pps.persist.API"),
		session,
		databaseName,
		pkgtime.NewSystemTimer(),
	}, nil
}

func (a *rethinkAPIServer) Close() error {
	return a.session.Close()
}

// job_id cannot be set
// timestamp cannot be set
func (a *rethinkAPIServer) CreateJobInfo(ctx context.Context, request *persist.JobInfo) (response *persist.JobInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	request.CommitIndex = commitIndex(request.InputCommit)
	if err := a.insertMessage(jobInfosTable, request); err != nil {
		return nil, err
	}
	return request, nil
}

func (a *rethinkAPIServer) InspectJob(ctx context.Context, request *pps.InspectJobRequest) (response *persist.JobInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	jobInfo := &persist.JobInfo{}
	if err := a.getMessageByPrimaryKey(jobInfosTable, request.Job.Id, jobInfo); err != nil {
		return nil, err
	}
	return jobInfo, nil
}

func (a *rethinkAPIServer) GetJobInfosByPipeline(ctx context.Context, request *pps.Pipeline) (response *persist.JobInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	jobInfoObjs, err := a.getMessagesByIndex(
		jobInfosTable,
		pipelineNameIndex,
		request.Name,
		func() proto.Message { return &persist.JobInfo{} },
	)
	if err != nil {
		return nil, err
	}
	var jobInfos []*persist.JobInfo
	for _, jobInfoObj := range jobInfoObjs {
		jobInfos = append(jobInfos, jobInfoObj.(*persist.JobInfo))
	}
	sortJobInfosByTimestampDesc(jobInfos)
	return &persist.JobInfos{
		JobInfo: jobInfos,
	}, nil
}

func (a *rethinkAPIServer) ListJobInfos(ctx context.Context, request *pps.ListJobRequest) (response *persist.JobInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	query := a.getTerm(jobInfosTable)
	if request.Pipeline != nil && len(request.InputCommit) > 0 {
		query = query.GetAllByIndex(
			pipelineNameAndInputIndex,
			gorethink.Expr([]interface{}{request.Pipeline.Name, commitIndex(request.InputCommit)}),
		)
	} else if request.Pipeline != nil {
		query = query.GetAllByIndex(
			pipelineNameIndex,
			request.Pipeline.Name,
		)
	} else if len(request.InputCommit) > 0 {
		query = query.GetAllByIndex(
			inputIndex,
			gorethink.Expr(commitIndex(request.InputCommit)),
		)
	}
	jobInfoObjs, err := a.getAllMessages(
		jobInfosTable,
		func() proto.Message { return &persist.JobInfo{} },
	)
	if err != nil {
		return nil, err
	}
	jobInfos := make([]*persist.JobInfo, len(jobInfoObjs))
	for i, jobInfoObj := range jobInfoObjs {
		jobInfos[i] = jobInfoObj.(*persist.JobInfo)
	}
	sortJobInfosByTimestampDesc(jobInfos)
	return &persist.JobInfos{
		JobInfo: jobInfos,
	}, nil
}

func (a *rethinkAPIServer) DeleteJobInfo(ctx context.Context, request *pps.Job) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	if err := a.deleteMessageByPrimaryKey(jobInfosTable, request.Id); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *rethinkAPIServer) CreateJobOutput(ctx context.Context, request *persist.JobOutput) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	if err := a.updateMessage(jobInfosTable, request); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *rethinkAPIServer) CreateJobState(ctx context.Context, request *persist.JobState) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	if err := a.updateMessage(jobInfosTable, request); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

// timestamp cannot be set
func (a *rethinkAPIServer) CreatePipelineInfo(ctx context.Context, request *persist.PipelineInfo) (response *persist.PipelineInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	if request.CreatedAt != nil {
		return nil, ErrTimestampSet
	}
	request.CreatedAt = a.now()
	if err := a.insertMessage(pipelineInfosTable, request); err != nil {
		return nil, err
	}
	return request, nil
}

func (a *rethinkAPIServer) GetPipelineInfo(ctx context.Context, request *pps.Pipeline) (response *persist.PipelineInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	pipelineInfo := &persist.PipelineInfo{}
	if err := a.getMessageByPrimaryKey(pipelineInfosTable, request.Name, pipelineInfo); err != nil {
		return nil, err
	}
	return pipelineInfo, nil
}

func (a *rethinkAPIServer) ListPipelineInfos(ctx context.Context, request *google_protobuf.Empty) (response *persist.PipelineInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	pipelineInfoObjs, err := a.getAllMessages(
		pipelineInfosTable,
		func() proto.Message { return &persist.PipelineInfo{} },
	)
	if err != nil {
		return nil, err
	}
	pipelineInfos := make([]*persist.PipelineInfo, len(pipelineInfoObjs))
	for i, pipelineInfoObj := range pipelineInfoObjs {
		pipelineInfos[i] = pipelineInfoObj.(*persist.PipelineInfo)
	}
	sortPipelineInfosByTimestampDesc(pipelineInfos)
	return &persist.PipelineInfos{
		PipelineInfo: pipelineInfos,
	}, nil
}

func (a *rethinkAPIServer) DeletePipelineInfo(ctx context.Context, request *pps.Pipeline) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	if err := a.deleteMessageByPrimaryKey(pipelineInfosTable, request.Name); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *rethinkAPIServer) insertMessage(table Table, message proto.Message) error {
	data, err := marshaller.MarshalToString(message)
	if err != nil {
		return err
	}
	_, err = a.getTerm(table).Insert(gorethink.JSON(data)).RunWrite(a.session)
	return err
}

func (a *rethinkAPIServer) updateMessage(table Table, message proto.Message) error {
	data, err := marshaller.MarshalToString(message)
	if err != nil {
		return err
	}
	_, err = a.getTerm(table).Update(gorethink.JSON(data)).RunWrite(a.session)
	return err
}

func (a *rethinkAPIServer) getMessageByPrimaryKey(table Table, key interface{}, message proto.Message) error {
	cursor, err := a.getTerm(table).Get(key).Default(gorethink.Error("value not found")).ToJSON().Run(a.session)
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

func (a *rethinkAPIServer) deleteMessageByPrimaryKey(table Table, value interface{}) error {
	_, err := a.getTerm(table).Get(value).Delete().RunWrite(a.session)
	return err
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

func commitIndex(commits []*pfs.Commit) string {
	var commitIDs []string
	for _, commit := range commits {
		commitIDs = append(commitIDs, commit.Id[0:10])
	}
	sort.Strings(commitIDs)
	var result []byte
	for _, commitID := range commitIDs {
		result = append(result, commitID...)
	}
	return string(result)
}
