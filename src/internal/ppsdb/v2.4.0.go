package ppsdb

import (
	"context"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
)

func migrateJobInfoV2_4_0(j *pps.JobInfo) *pps.JobInfo {
	j.Job.Pipeline.Project = pfsdb.MigrateProjectV2_4_0(j.Job.Pipeline.Project)
	j.OutputCommit = pfsdb.MigrateCommitV2_4_0(j.OutputCommit)
	pps.VisitInput(j.Details.Input, func(i *pps.Input) error {
		if i.Pfs != nil && i.Pfs.Project == "" {
			i.Pfs.Project = "default"
		}
		return nil
	})
	return j
}

func migratePipelineInfoV2_4_0(p *pps.PipelineInfo) *pps.PipelineInfo {
	p.Pipeline.Project = pfsdb.MigrateProjectV2_4_0(p.Pipeline.Project)
	p.SpecCommit = pfsdb.MigrateCommitV2_4_0(p.SpecCommit)
	pps.VisitInput(p.Details.Input, func(i *pps.Input) error {
		if i.Pfs != nil && i.Pfs.Project == "" {
			i.Pfs.Project = "default"
		}
		return nil
	})
	return p
}

func MigrateV2_4_0(ctx context.Context, tx *pachsql.Tx) error {
	var oldJob = new(pps.JobInfo)
	if err := col.MigratePostgreSQLCollection(ctx, tx, "jobs", jobsIndexes, oldJob, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		oldJob = migrateJobInfoV2_4_0(oldJob)
		return JobKey(oldJob.Job), oldJob, nil

	}); err != nil {
		return errors.Wrap(err, "could not migrate jobs")
	}
	var oldPipeline = new(pps.PipelineInfo)
	if err := col.MigratePostgreSQLCollection(ctx, tx, "pipelines", pipelinesIndexes, oldPipeline, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		oldPipeline = migratePipelineInfoV2_4_0(oldPipeline)
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
