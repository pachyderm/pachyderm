package v2_6_0

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

func branchlessCommitsPPS(ctx context.Context, tx *pachsql.Tx) error {
	jis, err := listCollectionProtos(ctx, tx, "jobs", &pps.JobInfo{})
	if err != nil {
		return errors.Wrap(err, "collecting jobs")
	}
	// Update the output commit for jobs.
	if err := func() (retErr error) {
		ctx, end := log.SpanContext(ctx, "updateJobs")
		defer end(log.Errorp(&retErr))
		batcher := newPostgresBatcher(ctx, tx)
		for _, ji := range jis {
			ji.OutputCommit.Repo = ji.OutputCommit.Branch.Repo
			data, err := proto.Marshal(ji)
			if err != nil {
				return errors.EnsureStack(err)
			}
			stmt := fmt.Sprintf("UPDATE collections.jobs SET proto=decode('%v', 'hex') WHERE key='%v'", hex.EncodeToString(data), jobKey(ji.Job))
			if err := batcher.Add(stmt); err != nil {
				return err
			}
		}
		return batcher.Close()
	}(); err != nil {
		return err
	}
	pis, err := listCollectionProtos(ctx, tx, "pipelines", &pps.PipelineInfo{})
	if err != nil {
		return errors.Wrap(err, "collecting pipelines")
	}
	for _, pi := range pis {
		pi.SpecCommit.Repo = pi.SpecCommit.Branch.Repo
		if err := updateCollectionProto(ctx, tx, "pipelines", pipelineCommitKey(pi.SpecCommit), pipelineCommitKey(pi.SpecCommit), pi); err != nil {
			return errors.Wrapf(err, "update collections.pipelines with key %q", pipelineCommitKey(pi.SpecCommit))
		}
	}
	return nil
}
