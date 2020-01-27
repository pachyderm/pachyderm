package transform

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

// synchronously start job for commit - this will do the initial task creation
// waits until the number of running jobs
// asynchronously wait for datums and then merge
// semaphore/pool to control how many

func getStatsCommit(pipelineInfo *pps.PipelineInfo, commitInfo *pfs.CommitInfo) *pfs.Commit {
	for _, commitRange := range commitInfo.Subvenance {
		if commitRange.Lower.Repo.Name == pipelineInfo.Pipeline.Name && commitRange.Upper.Repo.Name == pipelineInfo.Pipeline.Name {
			return commitRange.Lower
		}
	}
	return nil
}

func findAncestorCommit(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo) (*pfs.Commit, error) {
	return &pfs.Commit{}, nil
}

// forEachCommit listens for each READY output commit in the pipeline, and calls
// the given callback once for each such commit, synchronously.
func forEachCommit(
	driver driver.Driver,
	cb func(*pfs.CommitInfo, *pfs.Commit) error,
) error {
	pachClient := driver.PachClient()
	pi := driver.PipelineInfo()

	return pachClient.SubscribeCommitF(
		pi.Pipeline.Name,
		"",
		client.NewCommitProvenance(ppsconsts.SpecRepo, pi.Pipeline.Name, pi.SpecCommit.ID),
		"",
		pfs.CommitState_READY,
		func(ci *pfs.CommitInfo) error {
			statsCommit := getStatsCommit(driver.PipelineInfo(), ci)
			// TODO: ensure ci and statsCommit are in a consistent state
			if ci.Finished == nil {
				// Inspect the commit and check again if it has been finished (it may have
				// been closed since it was queued, e.g. by StopPipeline or StopJob)
				if ci, err := pachClient.InspectCommit(ci.Commit.Repo.Name, ci.Commit.ID); err != nil {
					return err
				} else if ci.Finished == nil {
					return cb(ci, statsCommit)
				} else {
					// Make sure that the job has been correctly finished as the commit has.
					ji, err := pachClient.InspectJobOutputCommit(ci.Commit.Repo.Name, ci.Commit.ID, true)
					if err != nil {
						return err
					}
					if !ppsutil.IsTerminal(ji.State) {
						if len(ci.Trees) == 0 {
							ji.State = pps.JobState_JOB_KILLED
							ji.Reason = "output commit is finished without data, but job state has not been updated"
							if err := updateJobState(pachClient, ji); err != nil {
								return fmt.Errorf("could not kill job with finished output commit: %v", err)
							}
						} else {
							ji.State = pps.JobState_JOB_SUCCESS
							if err := updateJobState(pachClient, ji); err != nil {
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

// dumb function to get a job and associated commits into a consistent state in case something broke it
func recoverFinishJob(jobInfo *pps.JobInfo, jobState *pps.JobState) error {
	return nil
}

// Run will run a transform pipeline until the driver is canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	reg, err := newRegistry(logger, driver)
	if err != nil {
		return err
	}

	logger.Logf("transform spawner started")

	// TODO: goroutine linearly waiting on jobs in the registry and cleaning up
	// after them, bubbling up errors, canceling

	return forEachCommit(driver, func(commitInfo *pfs.CommitInfo, statsCommit *pfs.Commit) error {
		return reg.startJob(commitInfo, statsCommit)
	})
}

// waitJob waits for the job in 'jobInfo' to finish, and then it collects the
// output from the job's workers and merges it into a commit (and may merge
// stats into a commit in the stats branch as well)
/*
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
		afterTime := time.Until(startTime.Add(timeout))
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
		// Create a datum iterator pointing at the job's inputs and split up the
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
		recoveredDatums := make(map[string]bool)
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
					} else if chunkState.State == State_COMPLETE {
						// if the chunk has been completed, grab the recovered datums from the chunk
						chunkRecoveredDatums, err := driver.GetDatumMap(ctx, chunkState.RecoveredDatums)
						if err != nil {
							return err
						}
						for k := range chunkRecoveredDatums {
							recoveredDatums[k] = true
						}
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
*/

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
