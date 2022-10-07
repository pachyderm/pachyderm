package ppsdb

import (
	"context"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

func MigrateV2_4_0(ctx context.Context, tx *pachsql.Tx) error {
	var oldJob = new(pps.JobInfo)
	if err := col.MigratePostgreSQLCollection(ctx, tx, "jobs", jobsIndexes, oldJob, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		if oldJob.Job.Pipeline.Project.GetName() == "" {
			oldJob.Job.Pipeline.Project = &pfs.Project{Name: "default"}
		}
		return JobKey(oldJob.Job), oldJob, nil

	}); err != nil {
		return errors.Wrap(err, "could not migrate jobs")
	}
	var oldPipeline = new(pps.PipelineInfo)
	if err := col.MigratePostgreSQLCollection(ctx, tx, "pipelines", pipelinesIndexes, oldPipeline, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		if oldPipeline.Pipeline.Project.GetName() == "" {
			oldPipeline.Pipeline.Project = &pfs.Project{Name: "default"}
		}
		if newKey, err = pipelineCommitKey(oldPipeline.SpecCommit); err != nil {
			return
		}
		return newKey, oldPipeline, nil

	},
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
	); err != nil {
		return errors.Wrap(err, "could not migrate jobs")
	}
	return nil
}
