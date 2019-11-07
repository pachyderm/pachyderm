package spawner

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	pfssync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

// forEachCommit listens for each READY output commit in the pipeline, and calls
// the given callback once for each such commit, synchronously.
func forEachCommit(
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	driver driver.Driver,
	cb func(*pfs.CommitInfo) error,
) error {
	return pachClient.SubscribeCommitF(
		pipelineInfo.Pipeline.Name,
		"",
		client.NewCommitProvenance(ppsconsts.SpecRepo, pipelineInfo.Pipeline.Name, pipelineInfo.SpecCommit.ID),
		"",
		pfs.CommitState_READY,
		func(ci *pfs.CommitInfo) error {
			if ci.Finished == nil {
				// Inspect the commit and check again if it has been finished (it may have
				// been closed since it was queued, e.g. by StopPipeline or StopJob)
				if ci, err := pachClient.InspectCommit(ci.Commit.Repo.Name, ci.Commit.ID); err != nil {
					return err
				} else if ci.Finished == nil {
					return cb(ci)
				} else {
					// Make sure that the job has been correctly finished as the commit has.
					ji, err := pachClient.InspectJobOutputCommit(ci.Commit.Repo.Name, ci.Commit.ID, true)
					if err != nil {
						return err
					}
					if !ppsutil.IsTerminal(ji.State) {
						if len(ci.Trees) == 0 {
							if err := driver.UpdateJobState(pachClient.Ctx(), ji.Job.ID,
								pps.JobState_JOB_KILLED, "output commit is finished without data, but job state has not been updated"); err != nil {
								return fmt.Errorf("could not kill job with finished output commit: %v", err)
							}
						} else {
							if err := driver.UpdateJobState(pachClient.Ctx(), ji.Job.ID, pps.JobState_JOB_SUCCESS, ""); err != nil {
								return fmt.Errorf("could not mark job with finished output commit as successful: %v", err)
							}
						}
					}
				}
			}
			return nil
		},
	)
}

func runTransform(
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger,
	driver driver.Driver,
) error {
	return forEachCommit(pachClient, pipelineInfo, driver, func(commitInfo *pfs.CommitInfo) error {
		// Check if a job was previously created for this commit. If not, make one
		var jobInfo *pps.JobInfo // job for commitInfo (new or old)
		jobInfos, err := pachClient.ListJob("", nil, commitInfo.Commit, -1, true)
		if err != nil {
			return err
		}
		if len(jobInfos) > 1 {
			return fmt.Errorf("multiple jobs found for commit: %s/%s", commitInfo.Commit.Repo.Name, commitInfo.Commit.ID)
		} else if len(jobInfos) < 1 {
			var statsCommit *pfs.Commit
			if pipelineInfo.EnableStats {
				for _, commitRange := range commitInfo.Subvenance {
					if commitRange.Lower.Repo.Name == pipelineInfo.Pipeline.Name && commitRange.Upper.Repo.Name == pipelineInfo.Pipeline.Name {
						statsCommit = commitRange.Lower
					}
				}
			}
			job, err := pachClient.CreateJob(pipelineInfo.Pipeline.Name, commitInfo.Commit, statsCommit)
			if err != nil {
				return err
			}
			logger.Logf("creating new job %q for output commit %q", job.ID, commitInfo.Commit.ID)
			// get jobInfo to look up spec commit, pipeline version, etc (if this
			// worker is stale and about to be killed, the new job may have a newer
			// pipeline version than the master. Or if the commit is stale, it may
			// have an older pipeline version than the master)
			jobInfo, err = pachClient.InspectJob(job.ID, false)
			if err != nil {
				return err
			}
		} else {
			// get latest job state
			jobInfo, err = pachClient.InspectJob(jobInfos[0].Job.ID, false)
			logger.Logf("found existing job %q for output commit %q", jobInfo.Job.ID, commitInfo.Commit.ID)
			if err != nil {
				return err
			}
		}

		switch {
		case ppsutil.IsTerminal(jobInfo.State):
			if jobInfo.StatsCommit != nil {
				if _, err := pachClient.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
					Commit: jobInfo.StatsCommit,
					Empty:  true,
				}); err != nil && !pfsserver.IsCommitFinishedErr(err) {
					return err
				}
			}
			if _, err := pachClient.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
				Commit: jobInfo.OutputCommit,
				Empty:  true,
			}); err != nil && !pfsserver.IsCommitFinishedErr(err) {
				return err
			}
			// ignore finished jobs (e.g. old pipeline & already killed)
			return nil
		case jobInfo.PipelineVersion < pipelineInfo.Version:
			// kill unfinished jobs from old pipelines (should generally be cleaned
			// up by PPS master, but the PPS master can fail, and if these jobs
			// aren't killed, future jobs will hang indefinitely waiting for their
			// parents to finish)
			if err := driver.UpdateJobState(pachClient.Ctx(), jobInfo.Job.ID,
				pps.JobState_JOB_KILLED, "pipeline has been updated"); err != nil {
				return fmt.Errorf("could not kill stale job: %v", err)
			}
			return nil
		case jobInfo.PipelineVersion > pipelineInfo.Version:
			return fmt.Errorf("job %s's version (%d) greater than pipeline's "+
				"version (%d), this should automatically resolve when the worker "+
				"is updated", jobInfo.Job.ID, jobInfo.PipelineVersion, pipelineInfo.Version)
		}

		// Now that the jobInfo is persisted, wait until all input commits are
		// ready, split the input datums into chunks and merge the results of
		// chunks as they're processed
		return waitJob(pachClient, pipelineInfo, logger, driver, jobInfo)
	})
}

// waitJob waits for the job in 'jobInfo' to finish, and then it collects the
// output from the job's workers and merges it into a commit (and may merge
// stats into a commit in the stats branch as well)
func waitJob(
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger,
	driver driver.Driver,
	jobInfo *pps.JobInfo,
) (retErr error) {
	logger.Logf("waiting on job %q (pipeline version: %d, state: %s)", jobInfo.Job.ID, jobInfo.PipelineVersion, jobInfo.State)
	ctx, cancel := context.WithCancel(pachClient.Ctx())
	pachClient = pachClient.WithCtx(ctx)

	// Watch the output commit to see if it's terminated (KILLED, FAILURE, or
	// SUCCESS) and if so, cancel the current context
	go func() {
		backoff.RetryUntilCancel(ctx, func() error {
			commitInfo, err := pachClient.PfsAPIClient.InspectCommit(ctx,
				&pfs.InspectCommitRequest{
					Commit:     jobInfo.OutputCommit,
					BlockState: pfs.CommitState_FINISHED,
				})
			if err != nil {
				if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
					defer cancel() // whether we return error or nil, job is done
					// Output commit was deleted. Delete job as well
					if _, err := driver.NewSTM(ctx, func(stm col.STM) error {
						// Delete the job if no other worker has deleted it yet
						jobPtr := &pps.EtcdJobInfo{}
						if err := driver.Jobs().ReadWrite(stm).Get(jobInfo.Job.ID, jobPtr); err != nil {
							return err
						}
						return driver.DeleteJob(stm, jobPtr)
					}); err != nil && !col.IsErrNotFound(err) {
						return err
					}
					return nil
				}
				return err
			}
			if commitInfo.Trees == nil {
				defer cancel() // whether job state update succeeds or not, job is done
				if _, err := driver.NewSTM(ctx, func(stm col.STM) error {
					// Read an up to date version of the jobInfo so that we
					// don't overwrite changes that have happened since this
					// function started.
					jobPtr := &pps.EtcdJobInfo{}
					if err := driver.Jobs().ReadWrite(stm).Get(jobInfo.Job.ID, jobPtr); err != nil {
						return err
					}
					if jobInfo.EnableStats {
						if _, err = pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
							Commit: jobInfo.StatsCommit,
							Empty:  true,
						}); err != nil && !pfsserver.IsCommitFinishedErr(err) {
							logger.Logf("error from FinishCommit for stats while cleaning up job: %+v", err)
						}
					}
					// TODO: this should just be part of UpdateJobState
					if !ppsutil.IsTerminal(jobPtr.State) {
						return ppsutil.UpdateJobState(driver.Pipelines().ReadWrite(stm), driver.Jobs().ReadWrite(stm), jobPtr, pps.JobState_JOB_KILLED, "")
					}
					return nil
				}); err != nil {
					return err
				}
			}
			return nil
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			return nil // retry again
		})
	}()
	if jobInfo.JobTimeout != nil {
		startTime, err := types.TimestampFromProto(jobInfo.Started)
		if err != nil {
			return err
		}
		timeout, err := types.DurationFromProto(jobInfo.JobTimeout)
		if err != nil {
			return err
		}
		afterTime := startTime.Add(timeout).Sub(time.Now())
		logger.Logf("cancelling job at: %+v", afterTime)
		timer := time.AfterFunc(afterTime, func() {
			if jobInfo.EnableStats {
				if _, err = pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
					Commit: jobInfo.StatsCommit,
					Empty:  true,
				}); err != nil {
					logger.Logf("error from FinishCommit for stats while timing out job: %+v", err)
				}
			}
			if _, err := pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
				Commit: jobInfo.OutputCommit,
				Empty:  true,
			}); err != nil {
				logger.Logf("error from FinishCommit while timing out job: %+v", err)
			}
		})
		defer timer.Stop()
	}
	backoff.RetryNotify(func() (retErr error) {
		// block until job inputs are ready
		failed, err := failedInputs(pachClient, jobInfo)
		if err != nil {
			return err
		}
		if len(failed) > 0 {
			reason := fmt.Sprintf("inputs %s failed", strings.Join(failed, ", "))
			if jobInfo.EnableStats {
				if _, err = pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
					Commit: jobInfo.StatsCommit,
					Empty:  true,
				}); err != nil && !pfsserver.IsCommitFinishedErr(err) {
					return err
				}
			}
			if err := driver.UpdateJobState(ctx, jobInfo.Job.ID, pps.JobState_JOB_FAILURE, reason); err != nil {
				return err
			}
			if _, err := pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
				Commit: jobInfo.OutputCommit,
				Empty:  true,
			}); err != nil && !pfsserver.IsCommitFinishedErr(err) {
				return err
			}
			return nil
		}
		// Create a datum factory pointing at the job's inputs and split up the
		// input data into chunks
		dit, err := datum.NewIterator(pachClient, jobInfo.Input)
		if err != nil {
			return err
		}
		parallelism, err := driver.GetExpectedNumWorkers()
		if err != nil {
			return fmt.Errorf("error from GetExpectedNumWorkers: %v", err)
		}
		numHashtrees, err := ppsutil.GetExpectedNumHashtrees(pipelineInfo.HashtreeSpec)
		if err != nil {
			return fmt.Errorf("error from GetExpectedNumHashtrees: %v", err)
		}
		plan := &common.Plan{}
		// Read the job document, and either resume (if we're recovering from a
		// crash) or mark it running. Also write the input chunks calculated above
		// into plansCol
		jobID := jobInfo.Job.ID
		if _, err := driver.NewSTM(ctx, func(stm col.STM) error {
			jobs := driver.Jobs().ReadWrite(stm)
			jobPtr := &pps.EtcdJobInfo{}
			if err := jobs.Get(jobID, jobPtr); err != nil {
				return err
			}
			if jobPtr.State == pps.JobState_JOB_KILLED {
				return nil
			}
			jobPtr.DataTotal = int64(dit.Len())
			if err := ppsutil.UpdateJobState(driver.Pipelines().ReadWrite(stm), driver.Jobs().ReadWrite(stm), jobPtr, pps.JobState_JOB_RUNNING, ""); err != nil {
				return err
			}
			plansCol := driver.Plans().ReadWrite(stm)
			if err := plansCol.Get(jobID, plan); err == nil {
				return nil
			}
			plan = newPlan(dit, jobInfo.ChunkSpec, parallelism, numHashtrees)
			return plansCol.Put(jobID, plan)
		}); err != nil {
			return err
		}
		defer func() {
			if retErr == nil {
				if _, err := driver.NewSTM(ctx, func(stm col.STM) error {
					chunksCol := driver.Chunks(jobID).ReadWrite(stm)
					chunksCol.DeleteAll()
					plansCol := driver.Plans().ReadWrite(stm)
					return plansCol.Delete(jobID)
				}); err != nil {
					retErr = err
				}
			}
		}()
		// Watch the chunks in order
		chunks := driver.Chunks(jobInfo.Job.ID).ReadOnly(ctx)
		var failedDatumID string
		for _, high := range plan.Chunks {
			chunkState := &common.ChunkState{}
			if err := chunks.WatchOneF(fmt.Sprint(high), func(e *watch.Event) error {
				var key string
				if err := e.Unmarshal(&key, chunkState); err != nil {
					return err
				}
				if key != fmt.Sprint(high) {
					return nil
				}
				if chunkState.State != common.State_RUNNING {
					if chunkState.State == common.State_FAILURE {
						failedDatumID = chunkState.DatumID
					}
					return errutil.ErrBreak
				}
				return nil
			}); err != nil {
				return err
			}
		}
		if err := driver.UpdateJobState(ctx, jobInfo.Job.ID, pps.JobState_JOB_MERGING, ""); err != nil {
			return err
		}
		var trees []*pfs.Object
		var size uint64
		var statsTrees []*pfs.Object
		var statsSize uint64
		if failedDatumID == "" || jobInfo.EnableStats {
			// Wait for all merges to happen.
			merges := driver.Merges(jobInfo.Job.ID).ReadOnly(ctx)
			for merge := int64(0); merge < plan.Merges; merge++ {
				mergeState := &common.MergeState{}
				if err := merges.WatchOneF(fmt.Sprint(merge), func(e *watch.Event) error {
					var key string
					if err := e.Unmarshal(&key, mergeState); err != nil {
						return err
					}
					if key != fmt.Sprint(merge) {
						return nil
					}
					if mergeState.State != common.State_RUNNING {
						trees = append(trees, mergeState.Tree)
						size += mergeState.SizeBytes
						statsTrees = append(statsTrees, mergeState.StatsTree)
						statsSize += mergeState.StatsSizeBytes
						return errutil.ErrBreak
					}
					return nil
				}); err != nil {
					return err
				}
			}
		}
		if jobInfo.EnableStats {
			if _, err = pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
				Commit:    jobInfo.StatsCommit,
				Trees:     statsTrees,
				SizeBytes: statsSize,
			}); err != nil && !pfsserver.IsCommitFinishedErr(err) {
				return err
			}
		}
		// If the job failed we finish the commit with an empty tree but only
		// after we've set the state, otherwise the job will be considered
		// killed.
		if failedDatumID != "" {
			reason := fmt.Sprintf("failed to process datum: %v", failedDatumID)
			if err := driver.UpdateJobState(ctx, jobInfo.Job.ID, pps.JobState_JOB_FAILURE, reason); err != nil {
				return err
			}
			if _, err = pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
				Commit: jobInfo.OutputCommit,
				Empty:  true,
			}); err != nil && !pfsserver.IsCommitFinishedErr(err) {
				return err
			}
			return nil
		}
		// Write out the datums processed/skipped and merged for this job
		buf := &bytes.Buffer{}
		pbw := pbutil.NewWriter(buf)
		for i := 0; i < dit.Len(); i++ {
			files := dit.DatumN(i)
			datumHash := common.HashDatum(pipelineInfo.Pipeline.Name, pipelineInfo.Salt, files)
			if _, err := pbw.WriteBytes([]byte(datumHash)); err != nil {
				return err
			}
		}
		datums, _, err := pachClient.PutObject(buf)
		if err != nil {
			return err
		}
		// Finish the job's output commit
		_, err = pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
			Commit:    jobInfo.OutputCommit,
			Trees:     trees,
			SizeBytes: size,
			Datums:    datums,
		})
		if err != nil && !pfsserver.IsCommitFinishedErr(err) {
			if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
				// output commit was deleted during e.g. FinishCommit, which means this job
				// should be deleted. Goro from top of waitJob() will observe the deletion,
				// delete the jobPtr and call cancel()--wait for that.
				<-ctx.Done() // wait for cancel()
				return nil
			}
			return err
		}
		// Handle egress
		if err := egress(pachClient, logger, jobInfo); err != nil {
			reason := fmt.Sprintf("egress error: %v", err)
			return driver.UpdateJobState(ctx, jobInfo.Job.ID, pps.JobState_JOB_FAILURE, reason)
		}
		return driver.UpdateJobState(ctx, jobInfo.Job.ID, pps.JobState_JOB_SUCCESS, "")
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("error in waitJob %v, retrying in %v", err, d)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		// Increment the job's restart count
		_, err = driver.NewSTM(ctx, func(stm col.STM) error {
			jobs := driver.Jobs().ReadWrite(stm)
			jobID := jobInfo.Job.ID
			jobPtr := &pps.EtcdJobInfo{}
			if err := jobs.Get(jobID, jobPtr); err != nil {
				return err
			}
			jobPtr.Restart++
			return jobs.Put(jobID, jobPtr)
		})
		if err != nil {
			logger.Logf("error incrementing restart count for job (%s)", jobInfo.Job.ID)
		}
		return nil
	})
	return nil
}

func egress(pachClient *client.APIClient, logger logs.TaggedLogger, jobInfo *pps.JobInfo) error {
	// copy the pach client (preserving auth info) so we can set a different
	// number of concurrent streams
	pachClient = pachClient.WithCtx(pachClient.Ctx())
	pachClient.SetMaxConcurrentStreams(100)
	var egressFailureCount int
	return backoff.RetryNotify(func() (retErr error) {
		if jobInfo.Egress != nil {
			logger.Logf("Starting egress upload for job (%v)", jobInfo)
			start := time.Now()
			url, err := obj.ParseURL(jobInfo.Egress.URL)
			if err != nil {
				return err
			}
			objClient, err := obj.NewClientFromURLAndSecret(url, false)
			if err != nil {
				return err
			}
			if err := pfssync.PushObj(pachClient, jobInfo.OutputCommit, objClient, url.Object); err != nil {
				return err
			}
			logger.Logf("Completed egress upload for job (%v), duration (%v)", jobInfo, time.Since(start))
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		egressFailureCount++
		if egressFailureCount > 3 {
			return err
		}
		logger.Logf("egress failed: %v; retrying in %v", err, d)
		return nil
	})
}

func failedInputs(pachClient *client.APIClient, jobInfo *pps.JobInfo) ([]string, error) {
	var failed []string
	var vistErr error
	blockCommit := func(name string, commit *pfs.Commit) {
		ci, err := pachClient.PfsAPIClient.InspectCommit(pachClient.Ctx(),
			&pfs.InspectCommitRequest{
				Commit:     commit,
				BlockState: pfs.CommitState_FINISHED,
			})
		if err != nil && vistErr == nil {
			vistErr = fmt.Errorf("error blocking on commit %s/%s: %v",
				commit.Repo.Name, commit.ID, err)
			return
		}
		if ci.Tree == nil && ci.Trees == nil {
			failed = append(failed, name)
		}
	}
	pps.VisitInput(jobInfo.Input, func(input *pps.Input) {
		if input.Pfs != nil && input.Pfs.Commit != "" {
			blockCommit(input.Pfs.Name, client.NewCommit(input.Pfs.Repo, input.Pfs.Commit))
		}
		if input.Cron != nil && input.Cron.Commit != "" {
			blockCommit(input.Cron.Name, client.NewCommit(input.Cron.Repo, input.Cron.Commit))
		}
		if input.Git != nil && input.Git.Commit != "" {
			blockCommit(input.Git.Name, client.NewCommit(input.Git.Name, input.Git.Commit))
		}
	})
	return failed, vistErr
}

func newPlan(dit datum.Iterator, spec *pps.ChunkSpec, parallelism int, numHashtrees int64) *common.Plan {
	if spec == nil {
		spec = &pps.ChunkSpec{}
	}
	if spec.Number == 0 && spec.SizeBytes == 0 {
		spec.Number = int64(dit.Len() / (parallelism * 10))
		if spec.Number == 0 {
			spec.Number = 1
		}
	}
	plan := &common.Plan{}
	if spec.Number != 0 {
		for i := spec.Number; i < int64(dit.Len()); i += spec.Number {
			plan.Chunks = append(plan.Chunks, int64(i))
		}
	} else {
		size := int64(0)
		for i := 0; i < dit.Len(); i++ {
			for _, input := range dit.DatumN(i) {
				size += int64(input.FileInfo.SizeBytes)
			}
			if size > spec.SizeBytes {
				plan.Chunks = append(plan.Chunks, int64(i))
			}
		}
	}
	plan.Chunks = append(plan.Chunks, int64(dit.Len()))
	plan.Merges = numHashtrees
	return plan
}
