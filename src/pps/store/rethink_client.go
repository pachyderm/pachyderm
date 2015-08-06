package store

import (
	"time"

	"github.com/dancannon/gorethink"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/peter-edge/go-google-protobuf"
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
	if _, err := gorethink.DB("pachyderm").TableCreate("pipeline_run").RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB("pachyderm").TableCreate("pipeline_status").RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB("pachyderm").
		TableCreate(
		"pipeline_container",
		gorethink.TableCreateOpts{
			PrimaryKey: "container_id"},
	).RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB("pachyderm").Table("pipeline_status").
		IndexCreate("run_id").RunWrite(session); err != nil {
		return nil, err
	}
	if _, err := gorethink.DB("pachyderm").Table("pipeline_container").
		IndexCreate("run_id").RunWrite(session); err != nil {
		return nil, err
	}
	return &rethinkClient{session, &jsonpb.Marshaller{EnumsAsString: true}}, nil
}

func (c *rethinkClient) AddPipelineRun(pipelineRun *pps.PipelineRun) error {
	data, err := c.marshaller.MarshalToString(pipelineRun)
	if err != nil {
		return err
	}
	_, err = gorethink.DB("pachyderm").Table("pipeline_run").Insert(gorethink.JSON(data)).RunWrite(c.session)
	return err
}

func (c *rethinkClient) GetPipelineRun(id string) (*pps.PipelineRun, error) {
	data := ""
	cursor, err := gorethink.DB("pachyderm").Table("pipeline_run").Get(id).ToJSON().Run(c.session)
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
	RunID                 string                    `gorethink:"run_id"`
	PipelineRunStatusType pps.PipelineRunStatusType `gorethink:"pipeline_run_status_type"`
	Seconds               int64                     `gorethink:"seconds"`
	Nanos                 int64                     `gorethink:"nanos"`
}

func (c *rethinkClient) AddPipelineRunStatus(id string, runStatusType pps.PipelineRunStatusType) error {
	now := time.Now()
	doc := pipelineRunStatus{
		id,
		runStatusType,
		now.Unix(),
		now.UnixNano(),
	}
	_, err := gorethink.DB("pachyderm").Table("pipeline_status").Insert(doc).RunWrite(c.session)
	return err
}

func (c *rethinkClient) GetPipelineRunStatusLatest(id string) (*pps.PipelineRunStatus, error) {
	var doc pipelineRunStatus
	cursor, err := gorethink.DB("pachyderm").Table("pipeline_status").
		GetAllByIndex("run_id", id).
		OrderBy(gorethink.Desc("seconds"), gorethink.Desc("nanos")).
		Nth(0).
		Run(c.session)
	if err != nil {
		return nil, err
	}
	if !cursor.Next(&doc) {
		return nil, cursor.Err()
	}
	var result pps.PipelineRunStatus
	result.PipelineRunStatusType = doc.PipelineRunStatusType
	result.Timestamp = &google_protobuf.Timestamp{Seconds: doc.Seconds, Nanos: int32(doc.Nanos)}
	return &result, nil
}

type pipelineContainer struct {
	ContainerID string `gorethink:"container_id,omitempty"`
	RunID       string `gorethink:"run_id,omitempty"`
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
	cursor, err := gorethink.DB("pachyderm").Table("pipeline_container").
		GetAllByIndex("run_id", id).
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
		result = append(result, container.ContainerID)
	}
	return result, nil
}
