// Package ppsdb contains the database schema that PPS uses.
package ppsdb

import (
	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	pipelinesCollectionName = "pipelines"
	jobsCollectionName      = "jobs"
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

// JobsPipelineIndex maps pipeline to Jobs started by the pipeline
var JobsPipelineIndex = &col.Index{
	Name: "pipeline",
	Extract: func(val proto.Message) string {
		return val.(*pps.StoredJobInfo).Pipeline.Name
	},
}

// JobsOutputIndex maps job outputs to the Job that create them.
var JobsOutputIndex = &col.Index{
	Name: "output_commit",
	Extract: func(val proto.Message) string {
		return pfsdb.CommitKey(val.(*pps.StoredJobInfo).OutputCommit)
	},
}

var jobsIndexes = []*col.Index{JobsPipelineIndex, JobsOutputIndex}

// Jobs returns a PostgresCollection of Jobs
func Jobs(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		jobsCollectionName,
		db,
		listener,
		&pps.StoredJobInfo{},
		jobsIndexes,
		nil,
	)
}

// AllCollections returns a list of all the PPS API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(pipelinesCollectionName, nil, nil, nil, pipelinesIndexes, nil),
		col.NewPostgresCollection(jobsCollectionName, nil, nil, nil, jobsIndexes, nil),
	}
}
