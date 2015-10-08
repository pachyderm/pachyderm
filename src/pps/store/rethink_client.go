package store

import (
	"errors"

	"github.com/dancannon/gorethink"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pedge.io/pkg/time"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
)

const (
	pfsCommitMappingsTable   = "pfs_commit_mappings"
	pipelinesTable           = "pipelines"
	pipelineContainersTable  = "pipeline_containers"
	pipelineLogsTable        = "pipeline_logs"
	pipelineRunsTable        = "pipeline_runs"
	pipelineRunStatusesTable = "pipeline_run_statuses"
	pipelineSourcesTable     = "pipeline_sources"
)

var (
	marshaller = &jsonpb.Marshaler{}

	tables = []string{
		pfsCommitMappingsTable,
		pipelinesTable,
		pipelineContainersTable,
		pipelineLogsTable,
		pipelineRunsTable,
		pipelineRunStatusesTable,
		pipelineSourcesTable,
	}

	tableToTableCreateOpts = map[string][]gorethink.TableCreateOpts{
		pipelineContainersTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: "container_id",
			},
		},
	}

	tableToIndexes = map[string][]string{
		pfsCommitMappingsTable: []string{
			"input_commit_id",
		},
		pipelinesTable: []string{
			"pipeline_source_id",
		},
		pipelineContainersTable: []string{
			"pipeline_run_id",
		},
		pipelineLogsTable: []string{
			"pipeline_run_id",
		},
		pipelineRunsTable: []string{
			"pipeline_id",
		},
		pipelineRunStatusesTable: []string{
			"pipeline_run_id",
		},
	}
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

type rethinkClient struct {
	session             *gorethink.Session
	databaseName        string
	timer               pkgtime.Timer
	pfsCommitMappings   gorethink.Term
	pipelines           gorethink.Term
	pipelineContainers  gorethink.Term
	pipelineLogs        gorethink.Term
	pipelineRuns        gorethink.Term
	pipelineRunStatuses gorethink.Term
	pipelineSources     gorethink.Term
}

func newRethinkClient(address string, databaseName string) (*rethinkClient, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: address})
	if err != nil {
		return nil, err
	}
	return &rethinkClient{
		session,
		databaseName,
		pkgtime.NewSystemTimer(),
		gorethink.DB(databaseName).Table(pfsCommitMappingsTable),
		gorethink.DB(databaseName).Table(pipelinesTable),
		gorethink.DB(databaseName).Table(pipelineContainersTable),
		gorethink.DB(databaseName).Table(pipelineLogsTable),
		gorethink.DB(databaseName).Table(pipelineRunsTable),
		gorethink.DB(databaseName).Table(pipelineRunStatusesTable),
		gorethink.DB(databaseName).Table(pipelineSourcesTable),
	}, nil
}

func (c *rethinkClient) Close() error {
	return c.session.Close()
}

func (c *rethinkClient) CreatePipelineRun(pipelineRun *pps.PipelineRun) error {
	data, err := marshaller.MarshalToString(pipelineRun)
	if err != nil {
		return err
	}
	_, err = c.pipelineRuns.Insert(gorethink.JSON(data)).RunWrite(c.session)
	return err
}

func (c *rethinkClient) GetPipelineRun(id string) (*pps.PipelineRun, error) {
	cursor, err := c.pipelineRuns.Get(id).ToJSON().Run(c.session)
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

func (c *rethinkClient) GetAllPipelineRuns(pipelineID string) ([]*pps.PipelineRun, error) {
	cursor, err := c.pipelineRuns.
		GetAllByIndex("pipeline_id", pipelineID).
		Map(func(row gorethink.Term) interface{} {
		return row.ToJSON()
	}).Run(c.session)
	if err != nil {
		return nil, err
	}
	var pipelineRuns []string
	if err := cursor.All(&pipelineRuns); err != nil {
		return nil, err
	}
	var result []*pps.PipelineRun
	for _, data := range pipelineRuns {
		var pipelineRun pps.PipelineRun
		if err := jsonpb.UnmarshalString(data, &pipelineRun); err != nil {
			return nil, err
		}
		result = append(result, &pipelineRun)
	}
	return result, nil
}

func (c *rethinkClient) CreatePipeline(pipeline *pps.Pipeline) error {
	return c.addMessage(c.pipelines, pipeline)
}

func (c *rethinkClient) GetPipeline(id string) (*pps.Pipeline, error) {
	var pipeline pps.Pipeline
	if err := c.getMessageByID(c.pipelines, id, &pipeline); err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func (c *rethinkClient) GetAllPipelines(pipelineSourceID string) ([]*pps.Pipeline, error) {
	cursor, err := c.pipelines.
		GetAllByIndex("pipeline_source_id", pipelineSourceID).
		Map(func(row gorethink.Term) interface{} {
		return row.ToJSON()
	}).Run(c.session)
	if err != nil {
		return nil, err
	}
	var pipelines []string
	if err := cursor.All(&pipelines); err != nil {
		return nil, err
	}
	var result []*pps.Pipeline
	for _, data := range pipelines {
		var pipeline pps.Pipeline
		if err := jsonpb.UnmarshalString(data, &pipeline); err != nil {
			return nil, err
		}
		result = append(result, &pipeline)
	}
	return result, nil
}

func (c *rethinkClient) CreatePipelineRunStatus(id string, statusType pps.PipelineRunStatusType) error {
	runStatus := &pps.PipelineRunStatus{
		PipelineRunId:         id,
		PipelineRunStatusType: statusType,
		Timestamp:             prototime.TimeToTimestamp(c.timer.Now()),
	}
	data, err := marshaller.MarshalToString(runStatus)
	if err != nil {
		return err
	}
	_, err = c.pipelineRunStatuses.Insert(gorethink.JSON(data)).RunWrite(c.session)
	protolog.Debugf("created pipeline run status: %v\n", runStatus)
	return err
}

func (c *rethinkClient) GetAllPipelineRunStatuses(pipelineRunID string) ([]*pps.PipelineRunStatus, error) {
	cursor, err := c.pipelineRunStatuses.
		GetAllByIndex("pipeline_run_id", pipelineRunID).
		OrderBy(gorethink.Desc("timestamp")).
		Without("id").
		Map(rethinkToJSON).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	var pipelineRunStatuses []*pps.PipelineRunStatus
	data := ""
	for cursor.Next(&data) {
		var pipelineRunStatus pps.PipelineRunStatus
		if err := jsonpb.UnmarshalString(data, &pipelineRunStatus); err != nil {
			return nil, err
		}
		pipelineRunStatuses = append(pipelineRunStatuses, &pipelineRunStatus)
	}
	return pipelineRunStatuses, cursor.Err()
}

func (c *rethinkClient) CreatePipelineRunContainers(pipelineContainers ...*pps.PipelineRunContainer) error {
	var pipelineContainerTerms []gorethink.Term
	for _, pipelineContainer := range pipelineContainers {
		data, err := marshaller.MarshalToString(pipelineContainer)
		if err != nil {
			return err
		}
		pipelineContainerTerms = append(pipelineContainerTerms, gorethink.JSON(data))
	}
	if _, err := c.pipelineContainers.Insert(pipelineContainerTerms).RunWrite(c.session); err != nil {
		return err
	}
	return nil
}

func (c *rethinkClient) GetPipelineRunContainers(pipelineRunID string) ([]*pps.PipelineRunContainer, error) {
	cursor, err := c.pipelineContainers.
		GetAllByIndex("pipeline_run_id", pipelineRunID).
		Without("id").
		Map(rethinkToJSON).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	var pipelineRunContainers []string
	if err := cursor.All(&pipelineRunContainers); err != nil {
		return nil, err
	}
	var result []*pps.PipelineRunContainer
	for _, data := range pipelineRunContainers {
		var pipelineRunContainer pps.PipelineRunContainer
		if err := jsonpb.UnmarshalString(data, &pipelineRunContainer); err != nil {
			return nil, err
		}
		result = append(result, &pipelineRunContainer)
	}
	return result, nil
}

func (c *rethinkClient) CreatePipelineRunLogs(pipelineLogs ...*pps.PipelineRunLog) error {
	var pipelineLogTerms []gorethink.Term
	for _, pipelineLog := range pipelineLogs {
		data, err := marshaller.MarshalToString(pipelineLog)
		if err != nil {
			return err
		}
		pipelineLogTerms = append(pipelineLogTerms, gorethink.JSON(data))
	}
	if _, err := c.pipelineLogs.Insert(pipelineLogTerms).RunWrite(c.session); err != nil {
		return err
	}
	return nil
}

func (c *rethinkClient) GetPipelineRunLogs(pipelineRunID string) ([]*pps.PipelineRunLog, error) {
	cursor, err := c.pipelineLogs.
		GetAllByIndex("pipeline_run_id", pipelineRunID).
		Map(func(row gorethink.Term) interface{} {
		return row.ToJSON()
	}).Run(c.session)
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

func (c *rethinkClient) CreatePfsCommitMapping(pfsCommitMapping *pps.PfsCommitMapping) error {
	data, err := marshaller.MarshalToString(pfsCommitMapping)
	if err != nil {
		return err
	}
	if _, err := c.pfsCommitMappings.Insert(gorethink.JSON(data)).RunWrite(c.session); err != nil {
		return err
	}
	return nil
}

func (c *rethinkClient) GetPfsCommitMappingLatest(inputRepositoryName string, inputCommitID string) (*pps.PfsCommitMapping, error) {
	cursor, err := c.pfsCommitMappings.
		GetAllByIndex("input_commit_id", inputCommitID).
		Map(func(row gorethink.Term) interface{} {
		return row.ToJSON()
	}).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	var pipelinePfsCommitMappings []string
	if err := cursor.All(&pipelinePfsCommitMappings); err != nil {
		return nil, err
	}
	var result []*pps.PfsCommitMapping
	for _, data := range pipelinePfsCommitMappings {
		var pfsCommitMapping pps.PfsCommitMapping
		if err := jsonpb.UnmarshalString(data, &pfsCommitMapping); err != nil {
			return nil, err
		}
		result = append(result, &pfsCommitMapping)
	}
	return getPfsCommitMappingLatestInMemory(result, inputRepositoryName)
}

func (c *rethinkClient) CreatePipelineSource(pipelineSource *pps.PipelineSource) error {
	return c.addMessage(c.pipelineSources, pipelineSource)
}

func (c *rethinkClient) GetPipelineSource(id string) (*pps.PipelineSource, error) {
	var pipelineSource pps.PipelineSource
	if err := c.getMessageByID(c.pipelineSources, id, &pipelineSource); err != nil {
		return nil, err
	}
	return &pipelineSource, nil
}

func (c *rethinkClient) UpdatePipelineSource(pipelineSource *pps.PipelineSource) error {
	return errors.New("not implemented")
}

func (c *rethinkClient) ArchivePipelineSource(id string) error {
	// TODO(pedge): do not delete
	_, err := c.pipelineSources.Get(id).Delete().RunWrite(c.session)
	return err
}

func (c *rethinkClient) GetAllPipelineSources() ([]*pps.PipelineSource, error) {
	cursor, err := c.pipelineSources.ToJSON().Run(c.session)
	if err != nil {
		return nil, err
	}
	var pipelineSources []*pps.PipelineSource
	data := ""
	for cursor.Next(&data) {
		var pipelineSource pps.PipelineSource
		if err := jsonpb.UnmarshalString(data, &pipelineSource); err != nil {
			return nil, err
		}
		pipelineSources = append(pipelineSources, &pipelineSource)
		data = ""
	}
	return pipelineSources, cursor.Err()
}

func (c *rethinkClient) addMessage(term gorethink.Term, message proto.Message) error {
	data, err := marshaller.MarshalToString(message)
	if err != nil {
		return err
	}
	_, err = term.Insert(gorethink.JSON(data)).RunWrite(c.session)
	return err
}

func (c *rethinkClient) getMessageByID(term gorethink.Term, id string, message proto.Message) error {
	cursor, err := term.Get(id).ToJSON().Run(c.session)
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

func rethinkToJSON(row gorethink.Term) interface{} {
	return row.ToJSON()
}
