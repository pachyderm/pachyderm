package service

import (
	"context"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfssync"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Run will run a service pipeline until the driver is canceled.
// TODO: The context handling is wonky here, the pachClient context is above the service context in the hierarchy.
// This is necessary to ensure we can finish the job when the service gets canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	pachClient := driver.PachClient()
	pipelineInfo := driver.PipelineInfo()
	return forEachJob(pachClient, pipelineInfo, logger, func(ctx context.Context, jobInfo *pps.JobInfo) (retErr error) {
		driver := driver.WithContext(ctx)
		if err := driver.UpdateJobState(jobInfo.Job, pps.JobState_JOB_RUNNING, ""); err != nil {
			return errors.EnsureStack(err)
		}
		jobInput := ppsutil.JobInput(pipelineInfo, jobInfo.OutputCommit)
		di, err := datum.NewIterator(pachClient, jobInput)
		if err != nil {
			return err
		}
		var meta *datum.Meta
		if err := di.Iterate(func(m *datum.Meta) error {
			if meta != nil {
				return errors.New("services must have a single datum")
			}
			meta = m
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
		if meta == nil {
			return errors.New("services must have a single datum")
		}
		meta.Job = jobInfo.Job
		defer func() {
			select {
			case <-ctx.Done():
				retErr = ppsutil.FinishJob(pachClient, jobInfo, pps.JobState_JOB_FINISHING, "")
			default:
			}
		}()
		// now that we're actually running the datum, use a pachClient which is bound to the job-scoped context
		pachClient := pachClient.WithCtx(ctx)
		storageRoot := filepath.Join(driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
		return pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
			pachClient := pachClient.WithCtx(ctx)
			cacheClient := pfssync.NewCacheClient(pachClient, renewer)
			return datum.WithSet(cacheClient, storageRoot, func(s *datum.Set) error {
				inputs := meta.Inputs
				logger = logger.WithData(inputs)
				env := driver.UserCodeEnv(logger.JobID(), jobInfo.OutputCommit, inputs)
				return s.WithDatum(meta, func(d *datum.Datum) error {
					err := driver.WithActiveData(inputs, d.PFSStorageRoot(), func() error {
						return d.Run(ctx, func(runCtx context.Context) error {
							return errors.EnsureStack(driver.RunUserCode(runCtx, logger, env))
						})
					})
					return errors.EnsureStack(err)
				})

			})
		})
	})
}

// Repeatedly runs the given callback with the latest job for the pipeline.
// The given context will be canceled if a newer job is ready, then this will
// wait for the previous callback to return before calling the callback again
// with the latest job.
func forEachJob(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, logger logs.TaggedLogger, cb func(context.Context, *pps.JobInfo) error) error {
	// These are used to cancel the existing service and wait for it to finish
	var cancel func()
	var eg *errgroup.Group
	return pachClient.SubscribeJob(pipelineInfo.Pipeline.Name, true, func(ji *pps.JobInfo) error {
		if ji.State == pps.JobState_JOB_FINISHING {
			return nil // don't pick up a "finishing" job
		}
		if cancel != nil {
			logger.Logf("canceling previous service, new job ready")
			cancel()
			if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
				return errors.EnsureStack(err)
			}
		}
		logger.Logf("starting new service, job: %s", ji.Job.ID)
		var ctx context.Context
		ctx, cancel = context.WithCancel(pachClient.Ctx())
		eg, ctx = errgroup.WithContext(ctx)
		eg.Go(func() error { return cb(ctx, ji) })
		return nil
	},
	)
}
