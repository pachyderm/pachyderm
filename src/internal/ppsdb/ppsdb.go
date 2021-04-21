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

// AllCollections returns a list of all the PPS API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(pipelinesCollectionName, nil, nil, nil, nil, nil),
		col.NewPostgresCollection(jobsCollectionName, nil, nil, nil, []*col.Index{JobsPipelineIndex, JobsOutputIndex}, nil),
	}
}

// Pipelines returns an EtcdCollection of pipelines
func Pipelines(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		pipelinesCollectionName,
		db,
		listener,
		&pps.EtcdPipelineInfo{},
		nil,
		nil,
	)
}

var (
	// JobsPipelineIndex maps pipeline to jobs started by the pipeline
	JobsPipelineIndex = &col.Index{
		Name: "Pipeline",
		Extract: func(val proto.Message) string {
			return val.(*pps.EtcdJobInfo).Pipeline.Name
		},
	}

	// JobsOutputIndex maps job outputs to the job that create them.
	JobsOutputIndex = &col.Index{
		Name: "OutputCommit",
		Extract: func(val proto.Message) string {
			return pfsdb.CommitKey(val.(*pps.EtcdJobInfo).OutputCommit)
		},
	}
)

// Jobs returns an EtcdCollection of jobs
func Jobs(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		jobsCollectionName,
		db,
		listener,
		&pps.EtcdJobInfo{},
		[]*col.Index{JobsPipelineIndex, JobsOutputIndex},
		nil,
	)
}
