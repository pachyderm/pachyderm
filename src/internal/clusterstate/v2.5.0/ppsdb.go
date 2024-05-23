package v2_5_0

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// pipelinesVersionIndex records the version numbers of pipelines
var pipelinesVersionIndex = &index{
	Name: "version",
	Extract: func(val proto.Message) string {
		info := val.(*pps.PipelineInfo)
		return versionKey(info.Pipeline, info.Version)
	},
}

// versionKey return a unique key for the given project, pipeline & version.  If
// the project is the empty string it will return an old-style key without a
// project; otherwise the key will include the project.  The version is
// zero-padded in order to facilitate sorting.
func versionKey(p *pps.Pipeline, version uint64) string {
	// zero pad in case we want to sort
	return fmt.Sprintf("%s@%08d", p, version)
}

// pipelinesNameKey returns the key used by PipelinesNameIndex to index a
// PipelineInfo.
func pipelinesNameKey(p *pps.Pipeline) string {
	return p.String()
}

// pipelinesNameIndex records the name of pipelines
var pipelinesNameIndex = &index{
	Name: "name",
	Extract: func(val proto.Message) string {
		return pipelinesNameKey(val.(*pps.PipelineInfo).Pipeline)
	},
}

var pipelinesIndexes = []*index{
	pipelinesVersionIndex,
	pipelinesNameIndex,
}

// parsePipelineKey expects keys to either be of the form <pipeline>@<id> or
// <project>/<pipeline>@<id>.
func parsePipelineKey(key string) (projectName, pipelineName, id string, err error) {
	parts := strings.Split(key, "@")
	if len(parts) != 2 || !uuid.IsUUIDWithoutDashes(parts[1]) {
		return "", "", "", errors.Errorf("key %s is not of form [<project>/]<pipeline>@<id>", key)
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
	if projectName := commit.Branch.Repo.Project.GetName(); projectName != "" {
		return fmt.Sprintf("%s/%s@%s", projectName, commit.Branch.Repo.Name, commit.Id), nil
	}
	return fmt.Sprintf("%s@%s", commit.Branch.Repo.Name, commit.Id), nil
}

func jobsPipelineKey(p *pps.Pipeline) string {
	return p.String()
}

// jobsPipelineIndex maps pipeline to Jobs started by the pipeline
var jobsPipelineIndex = &index{
	Name: "pipeline",
	Extract: func(val proto.Message) string {
		return jobsPipelineKey(val.(*pps.JobInfo).Job.Pipeline)
	},
}

func jobsTerminalKey(pipeline *pps.Pipeline, isTerminal bool) string {
	return fmt.Sprintf("%s_%v", pipeline, isTerminal)
}

var jobsTerminalIndex = &index{
	Name: "job_state",
	Extract: func(val proto.Message) string {
		jobInfo := val.(*pps.JobInfo)
		return jobsTerminalKey(jobInfo.Job.Pipeline, pps.IsTerminal(jobInfo.State))
	},
}

var jobsJobSetIndex = &index{
	Name: "jobset",
	Extract: func(val proto.Message) string {
		return val.(*pps.JobInfo).Job.Id
	},
}

var jobsIndexes = []*index{jobsPipelineIndex, jobsTerminalIndex, jobsJobSetIndex}

// jobKey is a string representation of a Job suitable for use as an indexing
// key.  It will include the project if the project name is not the empty
// string.
func jobKey(j *pps.Job) string {
	return fmt.Sprintf("%s@%s", j.Pipeline, j.Id)
}

func migrateJobInfo(j *pps.JobInfo) (*pps.JobInfo, error) {
	j.Job.Pipeline.Project = migrateProject(j.Job.Pipeline.Project)
	j.OutputCommit = migrateCommit(j.OutputCommit)
	if j.Details == nil {
		return j, nil
	}
	if err := pps.VisitInput(j.Details.Input, func(i *pps.Input) error {
		if i.Pfs != nil && i.Pfs.Project == "" {
			i.Pfs.Project = j.Job.Pipeline.Project.Name
		}
		if i.Cron != nil && i.Cron.Project == "" {
			i.Cron.Project = j.Job.Pipeline.Project.Name
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return j, nil
}

func migratePipelineInfo(p *pps.PipelineInfo) (*pps.PipelineInfo, error) {
	p.Pipeline.Project = migrateProject(p.Pipeline.Project)
	p.SpecCommit = migrateCommit(p.SpecCommit)
	if err := pps.VisitInput(p.Details.Input, func(i *pps.Input) error {
		if i.Pfs != nil && i.Pfs.Project == "" {
			i.Pfs.Project = "default"
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return p, nil
}

func migratePPSDB(ctx context.Context, tx *pachsql.Tx) error {
	var oldJob = new(pps.JobInfo)
	if err := migratePostgreSQLCollection(ctx, tx, "jobs", jobsIndexes, oldJob, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		if oldJob, err = migrateJobInfo(oldJob); err != nil {
			return "", nil, err
		}
		return jobKey(oldJob.Job), oldJob, nil

	}); err != nil {
		return errors.Wrap(err, "could not migrate jobs")
	}
	var oldPipeline = new(pps.PipelineInfo)
	if err := migratePostgreSQLCollection(ctx, tx, "pipelines", pipelinesIndexes, oldPipeline, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		if oldPipeline, err = migratePipelineInfo(oldPipeline); err != nil {
			return "", nil, err
		}
		if newKey, err = pipelineCommitKey(oldPipeline.SpecCommit); err != nil {
			return
		}
		return newKey, oldPipeline, nil

	},
		withKeyGen(func(key interface{}) (string, error) {
			if commit, ok := key.(*pfs.Commit); ok {
				return pipelineCommitKey(commit)
			}
			return "", errors.New("must provide a spec commit")
		}),
		withKeyCheck(func(key string) error {
			_, _, _, err := parsePipelineKey(key)
			return err
		}),
	); err != nil {
		return errors.Wrap(err, "could not migrate jobs")
	}
	return nil
}
