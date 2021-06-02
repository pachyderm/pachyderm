package service

import (
	"context"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Run will run a service pipeline until the driver is canceled.
// TODO: The context handling is wonky here, the pachClient context is above the service context in the hierarchy.
// This is necessary to ensure we can finish the job when the service gets canceled. Services will probably be reworked to not
// be triggered by output commits, so this is probably fine for now.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	pachClient := driver.PachClient()
	pipelineInfo := driver.PipelineInfo()
	return forEachCommit(pachClient, pipelineInfo, logger, func(ctx context.Context, commitInfo *pfs.CommitInfo) (retErr error) {
		driver := driver.WithContext(ctx)
		jobInfo, err := ensureJob(pachClient, pipelineInfo.Pipeline.Name, commitInfo.Commit, logger)
		if err != nil {
			return err
		}
		if err := driver.UpdateJobState(jobInfo.Job.ID, pps.JobState_JOB_RUNNING, ""); err != nil {
			return err
		}
		jobInput := ppsutil.JobInput(pipelineInfo, commitInfo)
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
			return err
		}
		if meta == nil {
			return errors.New("services must have a single datum")
		}
		defer func() {
			if common.IsDone(ctx) {
				retErr = finishJob(pachClient, jobInfo)
			}
		}()
		storageRoot := filepath.Join(driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
		return datum.WithSet(pachClient, storageRoot, func(s *datum.Set) error {
			inputs := meta.Inputs
			logger = logger.WithData(inputs)
			env := driver.UserCodeEnv(logger.JobID(), commitInfo.Commit, inputs)
			return s.WithDatum(ctx, meta, func(d *datum.Datum) error {
				return driver.WithActiveData(inputs, d.PFSStorageRoot(), func() error {
					return d.Run(ctx, func(runCtx context.Context) error {
						return driver.RunUserCode(runCtx, logger, env)
					})
				})
			})

		})
	})
}

// Repeatedly runs the given callback with the latest commit for the pipeline.
// The given context will be canceled if a newer commit is ready, then this will
// wait for the previous callback to return before calling the callback again
// with the latest commit.
func forEachCommit(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, logger logs.TaggedLogger, cb func(context.Context, *pfs.CommitInfo) error) error {
	// These are used to cancel the existing service and wait for it to finish
	var cancel func()
	var eg *errgroup.Group
	// TODO: Readd subscribe on spec commit provenance. Current code simplifies correctness in terms
	// of commits being closed / jobs being finished.
	return pachClient.SubscribeCommit(
		client.NewRepo(pipelineInfo.Pipeline.Name),
		"",
		nil,
		"",
		pfs.CommitState_READY,
		func(ci *pfs.CommitInfo) error {
			if cancel != nil {
				logger.Logf("canceling previous service, new commit ready")
				cancel()
				if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
					return err
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

func ensureJob(pachClient *client.APIClient, pipeline string, commit *pfs.Commit, logger logs.TaggedLogger) (*pps.JobInfo, error) {
	// Check if a job was previously created for this commit. If not, make one
	jobInfos, err := pachClient.ListJob("", nil, commit, -1, true)
	if err != nil {
		return nil, err
	}
	if len(jobInfos) > 1 {
		return nil, errors.Errorf("multiple jobs found for commit: %s@%s", commit.Branch.Repo.Name, commit.ID)
	} else if len(jobInfos) < 1 {
		job, err := pachClient.CreateJob(pipeline, commit, nil)
		if err != nil {
			return nil, err
		}
		logger.Logf("created new job %q for output commit %q", job.ID, commit.ID)
		// get JobInfo to look up spec commit, pipeline version, etc (if this
		// worker is stale and about to be killed, the new job may have a newer
		// pipeline version than the master. Or if the commit is stale, it may
		// have an older pipeline version than the master)
		return pachClient.InspectJob(job.ID, false)
	}
	// Get latest job state.
	logger.Logf("found existing job %q for output commit %q", jobInfos[0].Job.ID, commit.ID)
	return pachClient.InspectJob(jobInfos[0].Job.ID, false)
}

func finishJob(pachClient *client.APIClient, jobInfo *pps.JobInfo) error {
	_, err := pachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
		if _, err := builder.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
			Commit: jobInfo.OutputCommit,
		}); err != nil {
			return err
		}
		_, err := builder.PpsAPIClient.UpdateJobState(pachClient.Ctx(), &pps.UpdateJobStateRequest{
			Job:   jobInfo.Job,
			State: pps.JobState_JOB_SUCCESS,
		})
		return err
	})
	return err
}
