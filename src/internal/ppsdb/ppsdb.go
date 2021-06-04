// Package ppsdb contains the database schema that PPS uses.
package ppsdb

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	pipelinesCollectionName    = "pipelines"
	pipelineJobsCollectionName = "pipeline_jobs"
)

var pipelinesIndexes = []*col.Index{}

// Pipelines returns a PostgresCollection of pipelines
func Pipelines(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		pipelinesCollectionName,
		db,
		listener,
		&pps.StoredPipelineInfo{},
		nil,
		nil,
	)
}

// PipelineJobsPipelineIndex maps pipeline to PipelineJobs started by the pipeline
var PipelineJobsPipelineIndex = &col.Index{
	Name: "Pipeline",
	Extract: func(val proto.Message) string {
		return val.(*pps.StoredPipelineJobInfo).Pipeline.Name
	},
}

// PipelineJobsOutputIndex maps job outputs to the PipelineJob that create them.
var PipelineJobsOutputIndex = &col.Index{
	Name: "OutputCommit",
	Extract: func(val proto.Message) string {
		return pfsdb.CommitKey(val.(*pps.StoredPipelineJobInfo).OutputCommit)
	},
}

func PipelineJobTerminalKey(pipeline *pps.Pipeline, isTerminal bool) string {
	return fmt.Sprintf("%s_%v", pipeline.Name, isTerminal)
}

// PipelineJobsOutputIndex maps job outputs to the PipelineJob that create them.
var PipelineJobsTerminalIndex = &col.Index{
	Name: "PipelineJobState",
	Extract: func(val proto.Message) string {
		jobInfo := val.(*pps.StoredPipelineJobInfo)
		return PipelineJobTerminalKey(jobInfo.Pipeline, ppsutil.IsTerminal(jobInfo.State))
	},
}

var pipelineJobsIndexes = []*col.Index{PipelineJobsPipelineIndex, PipelineJobsOutputIndex, PipelineJobsTerminalIndex}

// PipelineJobs returns a PostgresCollection of PipelineJobs
func PipelineJobs(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		pipelineJobsCollectionName,
		db,
		listener,
		&pps.StoredPipelineJobInfo{},
		pipelineJobsIndexes,
		nil,
	)
}

// AllCollections returns a list of all the PPS API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(pipelinesCollectionName, nil, nil, nil, pipelinesIndexes, nil),
		col.NewPostgresCollection(pipelineJobsCollectionName, nil, nil, nil, pipelineJobsIndexes, nil),
	}
}
