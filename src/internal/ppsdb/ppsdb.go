// Package ppsdb contains the database schema that PPS uses.
package ppsdb

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// Pipelines returns an EtcdCollection of pipelines
func Pipelines(ctx context.Context, db *sqlx.DB, listener *col.PostgresListener) (col.PostgresCollection, error) {
	return col.NewPostgresCollection(
		ctx,
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
func Jobs(ctx context.Context, db *sqlx.DB, listener *col.PostgresListener) (col.PostgresCollection, error) {
	return col.NewPostgresCollection(
		ctx,
		db,
		listener,
		&pps.EtcdJobInfo{},
		[]*col.Index{JobsPipelineIndex, JobsOutputIndex},
		nil,
	)
}
