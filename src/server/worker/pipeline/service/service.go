package service

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/src/server/worker/pipeline"
)

// Repeatedly runs the given callback with the latest commit for the pipeline.
// The given context will be canceled if a newer commit is ready, then this will
// wait for the previous callback to return before calling the callback again
// with the latest commit.
func forLatestCommit(
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger,
	cb func(context.Context, *pfs.CommitInfo) error,
) error {
	// These are used to cancel the existing service and wait for it to finish
	var cancel func()
	var eg *errgroup.Group

	return pachClient.SubscribeCommitF(
		pipelineInfo.Pipeline.Name,
		"",
		client.NewCommitProvenance(ppsconsts.SpecRepo, pipelineInfo.Pipeline.Name, pipelineInfo.SpecCommit.ID),
		"",
		pfs.CommitState_READY,
		func(ci *pfs.CommitInfo) error {
			if cancel != nil {
				logger.Logf("canceling previous service, new commit ready")
				cancel()
				if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
					return err
				} else if common.IsDone(pachClient.Ctx()) {
					return pachClient.Ctx().Err()
				}
			}

			logger.Logf("starting new service, commit: %s", ci.Commit.ID)

			var ctx context.Context
			ctx, cancel = context.WithCancel(pachClient.Ctx())
			eg, ctx = errgroup.WithContext(ctx)
			eg.Go(func() error { return cb(ctx, ci) })

			return nil
		},
	)
}

// Run will run a service pipeline until the driver is canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	pachClient := driver.PachClient()
	pipelineInfo := driver.PipelineInfo()

	// The serviceCtx is only used for canceling user code (due to a new output
	// commit being ready)
	return forLatestCommit(pachClient, driver.PipelineInfo(), logger, func(serviceCtx context.Context, commitInfo *pfs.CommitInfo) error {
		// Create a job document matching the service's output commit
		jobInput := ppsutil.JobInput(pipelineInfo, commitInfo)
		job, err := pachClient.CreateJob(pipelineInfo.Pipeline.Name, commitInfo.Commit, nil)
		if err != nil {
			return err
		}
		logger := logger.WithJob(job.ID)

		dit, err := datum.NewIterator(pachClient, jobInput)
		if err != nil {
			return err
		}
		if dit.Len() != 1 {
			return errors.New("services must have a single datum")
		}
		inputs := dit.DatumN(0)
		logger = logger.WithData(inputs)

		// TODO: do something with stats? - this isn't an output repo so there's nowhere to put them
		_, err = driver.WithData(inputs, nil, logger, func(dir string, stats *pps.ProcessStats) error {
			if err := driver.UpdateJobState(job.ID, pps.JobState_JOB_RUNNING, ""); err != nil {
				logger.Logf("error updating job state: %+v", err)
			}

			eg, serviceCtx := errgroup.WithContext(serviceCtx)
			eg.Go(func() error {
				return driver.WithActiveData(inputs, dir, func() error {
					return pipeline.RunUserCode(driver.WithContext(serviceCtx), logger, nil, inputs)
				})
			})
			if pipelineInfo.Spout != nil {
				eg.Go(func() error { return pipeline.ReceiveSpout(serviceCtx, pachClient, pipelineInfo, logger) })
			}

			if err := eg.Wait(); err != nil {
				logger.Logf("error running user code: %+v", err)
			}

			// Only want to update this stuff if we were canceled due to a new commit
			if common.IsDone(serviceCtx) {
				// TODO: do this in a transaction
				if err := driver.UpdateJobState(job.ID, pps.JobState_JOB_SUCCESS, ""); err != nil {
					logger.Logf("error updating job progress: %+v", err)
				}
				if err := pachClient.FinishCommit(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID); err != nil {
					logger.Logf("could not finish output commit: %v", err)
				}
			}
			return nil
		})
		return err
	})
}
