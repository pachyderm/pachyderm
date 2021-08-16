// Package ppsdb contains the database schema that PPS uses.
package ppsdb

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	pipelinesCollectionName = "pipelines"
	jobsCollectionName      = "jobs"
)

// PipelinesVersionIndex records the version numbers of pipelines
var PipelinesVersionIndex = &col.Index{
	Name: "version",
	Extract: func(val proto.Message) string {
		info := val.(*pps.PipelineInfo)
		return fmt.Sprintf("%s@%d", info.Pipeline.Name, info.Version)
	},
}

var pipelinesIndexes = []*col.Index{
	PipelinesVersionIndex,
}

// Pipelines returns a PostgresCollection of pipelines
func Pipelines(db *sqlx.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		pipelinesCollectionName,
		db,
		listener,
		&pps.PipelineInfo{},
		pipelinesIndexes,
		col.WithKeyGen(func(key interface{}) (string, error) {
			if commit, ok := key.(*pfs.Commit); ok {
				if commit.Branch.Repo.Type != pfs.SpecRepoType {
					return "", errors.Errorf("commit %s is not from a spec repo", commit)
				}
				return fmt.Sprintf("%s@%s", commit.Branch.Repo.Name, commit.ID), nil
			}
			return "", errors.New("must provide a spec commit")
		}),
	)
}

// JobsPipelineIndex maps pipeline to Jobs started by the pipeline
var JobsPipelineIndex = &col.Index{
	Name: "pipeline",
	Extract: func(val proto.Message) string {
		return val.(*pps.JobInfo).Job.Pipeline.Name
	},
}

func JobTerminalKey(pipeline *pps.Pipeline, isTerminal bool) string {
	return fmt.Sprintf("%s_%v", pipeline.Name, isTerminal)
}

var JobsTerminalIndex = &col.Index{
	Name: "job_state",
	Extract: func(val proto.Message) string {
		jobInfo := val.(*pps.JobInfo)
		return JobTerminalKey(jobInfo.Job.Pipeline, ppsutil.IsTerminal(jobInfo.State))
	},
}

var JobsJobSetIndex = &col.Index{
	Name: "jobset",
	Extract: func(val proto.Message) string {
		return val.(*pps.JobInfo).Job.ID
	},
}

var jobsIndexes = []*col.Index{JobsPipelineIndex, JobsTerminalIndex, JobsJobSetIndex}

var JobKey = ppsutil.JobKey

// Jobs returns a PostgresCollection of Jobs
func Jobs(db *sqlx.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		jobsCollectionName,
		db,
		listener,
		&pps.JobInfo{},
		jobsIndexes,
	)
}

// AllCollections returns a list of all the PPS API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(pipelinesCollectionName, nil, nil, nil, pipelinesIndexes),
		col.NewPostgresCollection(jobsCollectionName, nil, nil, nil, jobsIndexes),
	}
}
