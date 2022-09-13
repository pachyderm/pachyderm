// Package ppsdb contains the database schema that PPS uses.
package ppsdb

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
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
		return VersionKey(info.Pipeline.Name, info.Version)
	},
}

func VersionKey(pipeline string, version uint64) string {
	// zero pad in case we want to sort
	return fmt.Sprintf("%s@%08d", pipeline, version)
}

// PipelinesNameIndex records the name of pipelines
var PipelinesNameIndex = &col.Index{
	Name: "name",
	Extract: func(val proto.Message) string {
		info := val.(*pps.PipelineInfo)
		return info.Pipeline.Name
	},
}

var pipelinesIndexes = []*col.Index{
	PipelinesVersionIndex,
	PipelinesNameIndex,
}

func ParsePipelineKey(key string) (string, string, error) {
	parts := strings.Split(key, "@")
	if len(parts) != 2 || !uuid.IsUUIDWithoutDashes(parts[1]) {
		return "", "", errors.Errorf("key %s is not of form <pipeline>@<id>", key)
	}
	return parts[0], parts[1], nil
}

// Pipelines returns a PostgresCollection of pipelines
func Pipelines(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
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
		col.WithKeyCheck(func(key string) error {
			_, _, err := ParsePipelineKey(key)
			return err
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
		return JobTerminalKey(jobInfo.Job.Pipeline, pps.IsTerminal(jobInfo.State))
	},
}

var JobsJobSetIndex = &col.Index{
	Name: "jobset",
	Extract: func(val proto.Message) string {
		return val.(*pps.JobInfo).Job.ID
	},
}

var jobsIndexes = []*col.Index{JobsPipelineIndex, JobsTerminalIndex, JobsJobSetIndex}

// JobKey is the string representation of a Job suitable for use as an indexing key
func JobKey(job *pps.Job) string {
	return job.String()
}

// Jobs returns a PostgresCollection of Jobs
func Jobs(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		jobsCollectionName,
		db,
		listener,
		&pps.JobInfo{},
		jobsIndexes,
	)
}

// CollectionsV0 returns a list of all the PPS API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CollectionsV0() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(pipelinesCollectionName, nil, nil, nil, pipelinesIndexes),
		col.NewPostgresCollection(jobsCollectionName, nil, nil, nil, jobsIndexes),
	}
}
