package ppsdb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// Pipeline wraps a PipelineInfo protobuf message.
type Pipeline struct {
	*pps.PipelineInfo
}

// GetPipeline reads a pipeline from the database.
func GetPipeline(ctx context.Context, tx *pachsql.Tx, projectName, pipelineName string) (Pipeline, error) {
	var commit = (&pfs.Repo{
		Project: &pfs.Project{Name: projectName},
		Name:    pipelineName,
		Type:    pfs.SpecRepoType,
	}).NewCommit("master", "")
	var err error
	branch, err := pfsdb.GetBranch(ctx, tx, commit.Branch)
	if err != nil {
		return Pipeline{}, errors.Wrap(err, "getting branch")
	}
	commit.Branch = branch.Branch
	commit.Id = branch.Head.Id
	c, err := pfsdb.GetCommitByKey(ctx, tx, commit)
	if err != nil {
		return Pipeline{}, errors.Wrapf(err, "pipeline was not inspected: couldn't find up to date spec for pipeline %q", pipelineName)
	}
	var pi = new(pps.PipelineInfo)
	if err := Pipelines(tx, nil).ReadOnly().Get(ctx, c.Commit, pi); err != nil {
		return Pipeline{}, errors.Wrapf(err, "could not read pipeline %s %v", pipelineName)
	}
	return Pipeline{pi}, nil
}

// UpsertPipeline attempts to update (or create) a pipeline.  It does not return an
// ID because ppsdb has not yet been converted to be fully relational yet and
// pipeline IDs do not yet exist.
func UpsertPipeline(ctx context.Context, tx *pachsql.Tx, pi *pps.PipelineInfo) error {
	var old = new(pps.PipelineInfo)
	return errors.Wrap(Pipelines(tx, nil).ReadWrite(tx).Upsert(ctx, pi.SpecCommit, old, func() error {
		var (
			o   = old.ProtoReflect()
			p   = pi.ProtoReflect()
			fds = p.Descriptor().Fields()
		)
		for i := 0; i < fds.Len(); i++ {
			var fd = fds.Get(i)
			if p.Has(fd) {
				o.Set(fd, p.Get(fd))
			}
		}
		return nil
	}), "upserting pipeline")
}

// PickPipeline picks a pipeline from the database.
func PickPipeline(ctx context.Context, pp *pps.PipelinePicker, tx *pachsql.Tx) (Pipeline, error) {
	if pp == nil || pp.Picker == nil {
		return Pipeline{}, errors.New("pipeline picker cannot be nil")
	}
	switch pp := pp.Picker.(type) {
	case *pps.PipelinePicker_Name:
		proj, err := pfsdb.PickProject(ctx, pp.Name.Project, tx)
		if err != nil {
			return Pipeline{}, errors.Wrap(err, "picking project")
		}
		repo, err := GetPipeline(ctx, tx, proj.Project.Name, pp.Name.Name)
		if err != nil {
			return Pipeline{}, errors.Wrap(err, "getting pipeline")
		}
		return repo, nil
	default:
		return Pipeline{}, errors.Errorf("pipeline picker is of an unknown type: %T", pp)
	}
}
