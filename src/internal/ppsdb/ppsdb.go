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
		return VersionKey(info.Pipeline.Project.GetName(), info.Pipeline.Name, info.Version)
	},
}

// VersionKey return a unique key for the given project, pipeline & version.  If
// the project is the empty string it will return an old-style key without a
// project; otherwise the key will include the project.  The version is
// zero-padded in order to facilitate sorting.
func VersionKey(projectName, pipelineName string, version uint64) string {
	// zero pad in case we want to sort
	if projectName == "" {
		return fmt.Sprintf("%s@%08d", pipelineName, version) // pre-projects style
	}
	return fmt.Sprintf("%s/%s@%08d", projectName, pipelineName, version) // projects style
}

// PipelinesNameKey returns the key used by PipelinesNameIndex to index a
// PipelineInfo.
func PipelinesNameKey(pi *pps.PipelineInfo) string {
	if projectName := pi.Pipeline.Project.GetName(); projectName != "" {
		return fmt.Sprintf("%s/%s", projectName, pi.Pipeline.Name)
	}
	return pi.Pipeline.Name
}

// PipelinesNameIndex records the name of pipelines
var PipelinesNameIndex = &col.Index{
	Name: "name",
	Extract: func(val proto.Message) string {
		return PipelinesNameKey(val.(*pps.PipelineInfo))
	},
}

var pipelinesIndexes = []*col.Index{
	PipelinesVersionIndex,
	PipelinesNameIndex,
}

// ParsePipelineKey expects keys to either be of the form <pipeline>@<id> or
// <project>/<pipeline>@<id>.
func ParsePipelineKey(key string) (projectName, pipelineName, id string, err error) {
	parts := strings.Split(key, "@")
	if len(parts) != 2 || !uuid.IsUUIDWithoutDashes(parts[1]) {
		return "", "", "", errors.Errorf("key %s is not of form [<project>/]<pipeline>@<id>")
	}
	id = parts[1]
	parts = strings.Split(parts[0], "/")
	if len(parts) == 0 {
		return "", "", "", errors.Errorf("key %s is not of form [<project>/]<pipeline>@<id>")
	}
	pipelineName = parts[len(parts)-1]
	if len(parts) == 1 {
		return
	}
	projectName = strings.Join(parts[0:len(parts)-1], "/")
	return
}

func pipelineCommitKey(commit *pfs.Commit) (string, error) {
	if commit.Branch.Repo.Type != pfs.SpecRepoType {
		return "", errors.Errorf("commit %s is not from a spec repo", commit)
	}
	// FIXME: include project
	if projectName := commit.Branch.Repo.Project.GetName(); projectName != "" {
		return fmt.Sprintf("%s/%s@%s", projectName, commit.Branch.Repo.Name, commit.ID), nil
	}
	return fmt.Sprintf("%s@%s", commit.Branch.Repo.Name, commit.ID), nil
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
				return pipelineCommitKey(commit)
			}
			return "", errors.New("must provide a spec commit")
		}),
		col.WithKeyCheck(func(key string) error {
			_, _, _, err := ParsePipelineKey(key)
			return err
		}),
	)
}

func jobsPipelineKey(j *pps.Job) string {
	if projectName := j.Pipeline.Project.GetName(); projectName != "" {
		return fmt.Sprintf("%s/%s", projectName, j.Pipeline.Name)
	}
	return j.Pipeline.Name
}

// JobsPipelineIndex maps pipeline to Jobs started by the pipeline
var JobsPipelineIndex = &col.Index{
	Name: "pipeline",
	Extract: func(val proto.Message) string {
		return jobsPipelineKey(val.(*pps.JobInfo).Job)
	},
}

func JobTerminalKey(pipeline *pps.Pipeline, isTerminal bool) string {
	if projectName := pipeline.Project.GetName(); projectName != "" {
		return fmt.Sprintf("%s/%s_%v", projectName, pipeline.Name, isTerminal)
	}
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

// JobKey is a string representation of a Job suitable for use as an indexing
// key.  It will include the project if the project name is not the empty
// string.
func JobKey(j *pps.Job) string {
	if projectName := j.Pipeline.Project.GetName(); projectName != "" {
		return fmt.Sprintf("%s/%s@%s", projectName, j.Pipeline.Name, j.ID)
	}
	return fmt.Sprintf("%s@%s", j.Pipeline.Name, j.ID)
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
