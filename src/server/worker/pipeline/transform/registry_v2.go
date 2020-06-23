package transform

import (
	"context"
	"math"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
)

type pendingJobV2 struct {
	driver          driver.Driver
	commitInfo      *pfs.CommitInfo
	statsCommitInfo *pfs.CommitInfo
	cancel          context.CancelFunc
	logger          logs.TaggedLogger
	ji              *pps.JobInfo
	jdit            chain.JobDatumIterator
	taskMaster      *work.Master
}

func (pj *pendingJobV2) writeJobInfo() error {
	pj.logger.Logf("updating job info, state: %s", pj.ji.State)
	return writeJobInfo(pj.driver.PachClient(), pj.ji)
}

func writeJobInfo(pachClient *client.APIClient, jobInfo *pps.JobInfo) error {
	_, err := pachClient.PpsAPIClient.UpdateJobState(pachClient.Ctx(), &pps.UpdateJobStateRequest{
		Job:           jobInfo.Job,
		State:         jobInfo.State,
		Reason:        jobInfo.Reason,
		Restart:       jobInfo.Restart,
		DataProcessed: jobInfo.DataProcessed,
		DataSkipped:   jobInfo.DataSkipped,
		DataTotal:     jobInfo.DataTotal,
		DataFailed:    jobInfo.DataFailed,
		DataRecovered: jobInfo.DataRecovered,
		Stats:         jobInfo.Stats,
	})
	return err
}

func (pj *pendingJobV2) saveJobStats(stats *DatumStats) {
	// Any unaccounted-for datums were skipped in the job datum iterator
	pj.ji.DataSkipped = int64(pj.jdit.MaxLen()) - stats.DatumsProcessed - stats.DatumsFailed - stats.DatumsRecovered
	pj.ji.DataProcessed = stats.DatumsProcessed
	pj.ji.DataFailed = stats.DatumsFailed
	pj.ji.DataRecovered = stats.DatumsRecovered
	pj.ji.DataTotal = int64(pj.jdit.MaxLen())
	pj.ji.Stats = stats.ProcessStats
}

type registryV2 struct {
	driver      driver.Driver
	logger      logs.TaggedLogger
	taskQueue   *work.TaskQueue
	concurrency int64
	limiter     limit.ConcurrencyLimiter
	jobChain    chain.JobChainV2
}

func newRegistryV2(logger logs.TaggedLogger, driver driver.Driver) (*registryV2, error) {
	// Determine the maximum number of concurrent tasks we will allow
	concurrency, err := driver.ExpectedNumWorkers()
	if err != nil {
		return nil, err
	}
	taskQueue, err := driver.NewTaskQueue()
	if err != nil {
		return nil, err
	}
	return &registryV2{
		driver:      driver,
		logger:      logger,
		concurrency: concurrency,
		taskQueue:   taskQueue,
		limiter:     limit.New(int(concurrency)),
	}, nil
}

func (reg *registryV2) succeedJob(pj *pendingJobV2) error {
	var newState pps.JobState
	if pj.ji.Egress == nil {
		pj.logger.Logf("job successful, closing commits")
		newState = pps.JobState_JOB_SUCCESS
	} else {
		pj.logger.Logf("job successful, advancing to egress")
		newState = pps.JobState_JOB_EGRESSING
	}
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	if err := finishJobV2(reg.driver.PipelineInfo(), reg.driver.PachClient(), pj.ji, newState, ""); err != nil {
		return err
	}
	return reg.jobChain.Succeed(pj)
}

func (reg *registryV2) failJob(pj *pendingJobV2, reason string) error {
	pj.logger.Logf("failing job with reason: %s", reason)
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	if err := finishJobV2(reg.driver.PipelineInfo(), reg.driver.PachClient(), pj.ji, pps.JobState_JOB_FAILURE, reason); err != nil {
		return err
	}
	return reg.jobChain.Fail(pj)
}

func (reg *registry) killJob(pj *pendingJobV2, reason string) error {
	pj.logger.Logf("killing job with reason: %s", reason)
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	if err := finishJob(reg.driver.PipelineInfo(), reg.driver.PachClient(), pj.ji, pps.JobState_JOB_KILLED, reason); err != nil {
		return err
	}
	return reg.jobChain.Fail(pj)
}

func (reg *registryV2) initializeJobChain(commitInfo *pfs.CommitInfo) error {
	if reg.jobChain == nil {
		// Get the most recent successful commit starting from the given commit
		parentCommitInfo, err := reg.getParentCommitInfo(commitInfo)
		if err != nil {
			return err
		}

		var baseDatums chain.DatumSet
		if parentCommitInfo != nil {
			baseDatums, err = reg.getDatumSet(parentCommitInfo.Datums)
			if err != nil {
				return err
			}
		} else {
			baseDatums = make(chain.DatumSet)
		}

		if reg.driver.PipelineInfo().S3Out {
			// When running a pipeline with S3Out, we need to yield every datum for
			// every job, use a no-skip job chain for this.
			reg.jobChain = chain.NewNoSkipJobChain(
				&hasher{
					name: reg.driver.PipelineInfo().Pipeline.Name,
					salt: reg.driver.PipelineInfo().Salt,
				},
			)
		} else {
			reg.jobChain = chain.NewJobChain(
				&hasher{
					name: reg.driver.PipelineInfo().Pipeline.Name,
					salt: reg.driver.PipelineInfo().Salt,
				},
				baseDatums,
			)
		}
	}

	return nil
}

func serializeDatumDataV2(data *DatumDataV2) (*types.Any, error) {
	serialized, err := types.MarshalAny(data)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

func deserializeDatumDataV2(any *types.Any) (*DatumDataV2, error) {
	data := &DatumDataV2{}
	if err := types.UnmarshalAny(any, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (reg *registryV2) getParentCommitInfo(commitInfo *pfs.CommitInfo) (*pfs.CommitInfo, error) {
	pachClient := reg.driver.PachClient()
	outputCommitID := commitInfo.Commit.ID
	// Walk up the commit chain to find a successfully finished commit
	for commitInfo.ParentCommit != nil {
		reg.logger.Logf(
			"blocking on parent commit %q before writing to output commit %q",
			commitInfo.ParentCommit.ID, outputCommitID,
		)
		parentCommitInfo, err := pachClient.PfsAPIClient.InspectCommit(pachClient.Ctx(),
			&pfs.InspectCommitRequest{
				Commit:     commitInfo.ParentCommit,
				BlockState: pfs.CommitState_FINISHED,
			})
		if err != nil {
			return nil, err
		}
		// TODO: should we be checking `.Tree` here as well?
		if parentCommitInfo.Trees != nil {
			return parentCommitInfo, nil
		}
		commitInfo = parentCommitInfo
	}
	return nil, nil
}

// ensureJob loads an existing job for the given commit in the pipeline, or
// creates it if there is none. If more than one such job exists, an error will
// be generated.
func (reg *registryV2) ensureJob(commitInfo *pfs.CommitInfo, statsCommit *pfs.Commit) (*pps.JobInfo, error) {
	pachClient := reg.driver.PachClient()
	// Check if a job was previously created for this commit. If not, make one
	jobInfos, err := pachClient.ListJob("", nil, commitInfo.Commit, -1, true)
	if err != nil {
		return nil, err
	}
	if len(jobInfos) > 1 {
		return nil, errors.Errorf("multiple jobs found for commit: %s/%s", commitInfo.Commit.Repo.Name, commitInfo.Commit.ID)
	} else if len(jobInfos) < 1 {
		job, err := pachClient.CreateJob(reg.driver.PipelineInfo().Pipeline.Name, commitInfo.Commit, statsCommit)
		if err != nil {
			return nil, err
		}
		reg.logger.Logf("created new job %q for output commit %q", job.ID, commitInfo.Commit.ID)
		// get jobInfo to look up spec commit, pipeline version, etc (if this
		// worker is stale and about to be killed, the new job may have a newer
		// pipeline version than the master. Or if the commit is stale, it may
		// have an older pipeline version than the master)
		return pachClient.InspectJob(job.ID, false)
	}
	// Get latest job state.
	reg.logger.Logf("found existing job %q for output commit %q", jobInfos[0].Job.ID, commitInfo.Commit.ID)
	return pachClient.InspectJob(jobInfos[0].Job.ID, false)
}

func (reg *registryV2) startJob(commitInfo *pfs.CommitInfo, statsCommit *pfs.Commit) error {
	if err := reg.initializeJobChain(commitInfo); err != nil {
		return err
	}
	var asyncEg *errgroup.Group
	reg.limiter.Acquire()
	defer func() {
		if asyncEg == nil {
			// The async errgroup never got started, so give up the limiter lock
			reg.limiter.Release()
		}
	}()
	jobInfo, err := reg.ensureJob(commitInfo, statsCommit)
	if err != nil {
		return err
	}
	var statsCommitInfo *pfs.CommitInfo
	if statsCommit != nil {
		statsCommitInfo, err = reg.driver.PachClient().InspectCommit(statsCommit.Repo.Name, statsCommit.ID)
		if err != nil {
			return err
		}
	}
	jobCtx, cancel := context.WithCancel(reg.driver.PachClient().Ctx())
	driver := reg.driver.WithContext(jobCtx)
	// Build the pending job to send out to workers - this will block if we have
	// too many already
	pj := &pendingJobV2{
		driver:          driver,
		commitInfo:      commitInfo,
		statsCommitInfo: statsCommitInfo,
		logger:          reg.logger.WithJob(jobInfo.Job.ID),
		ji:              jobInfo,
		cancel:          cancel,
	}
	switch {
	case ppsutil.IsTerminal(jobInfo.State):
		// TODO: V2 recovery for inconsistent output commit / job state?
		return nil
	case jobInfo.PipelineVersion < reg.driver.PipelineInfo().Version:
		// kill unfinished jobs from old pipelines (should generally be cleaned
		// up by PPS master, but the PPS master can fail, and if these jobs
		// aren't killed, future jobs will hang indefinitely waiting for their
		// parents to finish)
		pj.ji.State = pps.JobState_JOB_KILLED
		pj.ji.Reason = "pipeline has been updated"
		if err := pj.writeJobInfo(); err != nil {
			return errors.Wrap(err, "failed to kill stale job")
		}
		return nil
	case jobInfo.PipelineVersion > reg.driver.PipelineInfo().Version:
		return errors.Errorf("job %s's version (%d) greater than pipeline's "+
			"version (%d), this should automatically resolve when the worker "+
			"is updated", jobInfo.Job.ID, jobInfo.PipelineVersion, reg.driver.PipelineInfo().Version)
	}
	// Inputs must be ready before we can construct a datum iterator, so do this
	// synchronously to ensure correct order in the jobChain.
	if err := pj.logger.LogStep("waiting for job inputs", func() error {
		return reg.processJobStarting(pj)
	}); err != nil {
		return err
	}
	// TODO: we should NOT start the job this way if it is in EGRESSING
	pj.jdit, err = reg.jobChain.Start(pj)
	if err != nil {
		return err
	}
	var afterTime time.Duration
	if pj.ji.JobTimeout != nil {
		startTime, err := types.TimestampFromProto(pj.ji.Started)
		if err != nil {
			return err
		}
		timeout, err := types.DurationFromProto(pj.ji.JobTimeout)
		if err != nil {
			return err
		}
		afterTime = time.Until(startTime.Add(timeout))
	}
	asyncEg, jobCtx = errgroup.WithContext(pj.driver.PachClient().Ctx())
	pj.driver = reg.driver.WithContext(jobCtx)
	asyncEg.Go(func() error {
		defer pj.cancel()
		if pj.ji.JobTimeout != nil {
			pj.logger.Logf("cancelling job at: %+v", afterTime)
			timer := time.AfterFunc(afterTime, func() {
				reg.killJob(pj, "job timed out")
			})
			defer timer.Stop()
		}
		return backoff.RetryUntilCancel(pj.driver.PachClient().Ctx(), func() error {
			return reg.superviseJob(pj)
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			pj.logger.Logf("error in superviseJob: %v, retrying in %+v", err, d)
			return nil
		})
	})
	asyncEg.Go(func() error {
		defer pj.cancel()
		mutex := &sync.Mutex{}
		mutex.Lock()
		defer mutex.Unlock()
		// This runs the callback asynchronously, but we want to block the errgroup until it completes
		if err := reg.taskQueue.RunTask(pj.driver.PachClient().Ctx(), func(master *work.Master) {
			defer mutex.Unlock()
			pj.taskMaster = master
			backoff.RetryUntilCancel(pj.driver.PachClient().Ctx(), func() error {
				var err error
				for err == nil {
					err = reg.processJob(pj)
				}
				return err
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				pj.logger.Logf("processJob error: %v, retring in %v", err, d)
				for err != nil {
					if st, ok := err.(errors.StackTracer); ok {
						pj.logger.Logf("error stack: %+v", st.StackTrace())
					}
					err = errors.Unwrap(err)
				}
				pj.jdit.Reset()
				// Get job state, increment restarts, write job state
				pj.ji, err = pj.driver.PachClient().InspectJob(pj.ji.Job.ID, false)
				if err != nil {
					return err
				}
				pj.ji.Restart++
				if err := pj.writeJobInfo(); err != nil {
					pj.logger.Logf("error incrementing restart count for job (%s): %v", pj.ji.Job.ID, err)
				}
				// Reload the job's commitInfo as it may have changed
				pj.commitInfo, err = reg.driver.PachClient().InspectCommit(pj.commitInfo.Commit.Repo.Name, pj.commitInfo.Commit.ID)
				if err != nil {
					return err
				}
				if statsCommit != nil {
					pj.statsCommitInfo, err = reg.driver.PachClient().InspectCommit(statsCommit.Repo.Name, statsCommit.ID)
					if err != nil {
						return err
					}
				}
				return nil
			})
			pj.logger.Logf("master done running processJobs")
			// TODO: make sure that all paths close the commit correctly
		}); err != nil {
			return err
		}
		// This should block until the callback has completed
		mutex.Lock()
		return nil
	})
	go func() {
		defer reg.limiter.Release()
		// Make sure the job has been removed from the job chain, ignore any errors
		defer reg.jobChain.Fail(pj)
		if err := asyncEg.Wait(); err != nil {
			pj.logger.Logf("fatal job error: %v", err)
		}
	}()
	return nil
}

// superviseJob watches for the output commit closing and cancels the job, or
// deletes it if the output commit is removed.
func (reg *registryV2) superviseJob(pj *pendingJobV2) error {
	commitInfo, err := pj.driver.PachClient().PfsAPIClient.InspectCommit(pj.driver.PachClient().Ctx(),
		&pfs.InspectCommitRequest{
			Commit:     pj.ji.OutputCommit,
			BlockState: pfs.CommitState_FINISHED,
		})
	if err != nil {
		if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
			defer pj.cancel() // whether we return error or nil, job is done
			// Stop the job and clean up any job state in the registry
			if err := reg.killJob(pj, "output commit missing"); err != nil {
				return err
			}
			// Output commit was deleted. Delete job as well
			if _, err := pj.driver.NewSTM(func(stm col.STM) error {
				// Delete the job if no other worker has deleted it yet
				jobPtr := &pps.EtcdJobInfo{}
				if err := pj.driver.Jobs().ReadWrite(stm).Get(pj.ji.Job.ID, jobPtr); err != nil {
					return err
				}
				return pj.driver.DeleteJob(stm, jobPtr)
			}); err != nil && !col.IsErrNotFound(err) {
				return err
			}
			return nil
		}
		return err
	}
	// commitInfo.Trees is set by non-S3-output jobs, while commitInfo.Tree is
	// set by S3-output jobs
	// TODO: why don't we cancel in all cases?
	if commitInfo.Trees == nil && commitInfo.Tree == nil {
		defer pj.cancel() // whether job state update succeeds or not, job is done
		return reg.killJob(pj, "output commit closed")
	}
	return nil
}

func (reg *registryV2) processJob(pj *pendingJobV2) error {
	state := pj.ji.State
	switch {
	case ppsutil.IsTerminal(state):
		pj.cancel()
		return errutil.ErrBreak
	case state == pps.JobState_JOB_STARTING:
		return errors.New("job should have been moved out of the STARTING state before processJob")
	case state == pps.JobState_JOB_RUNNING:
		return pj.logger.LogStep("processing job datums", func() error {
			return reg.processJobRunningV2(pj)
		})
	case state == pps.JobState_JOB_EGRESSING:
		return pj.logger.LogStep("egressing job data", func() error {
			return reg.processJobEgressV2(pj)
		})
	}
	pj.cancel()
	return errors.Errorf("unknown job state: %v", state)
}

func (reg *registryV2) processJobStarting(pj *pendingJobV2) error {
	// block until job inputs are ready
	failed, err := failedInputs(pj.driver.PachClient(), pj.ji)
	if err != nil {
		return err
	}

	if len(failed) > 0 {
		reason := fmt.Sprintf("inputs failed: %s", strings.Join(failed, ", "))
		return reg.failJob(pj, reason, nil, 0)
	}

	if pj.driver.PipelineInfo().S3Out && pj.commitInfo.ParentCommit != nil {
		// We don't want S3-out pipelines to merge datum output with the parent
		// commit, so we create a PutFile record to delete "/". Doing it before
		// we move the job to the RUNNING state ensures that:
		// 1) workers can't process datums unless DeleteFile("/") has run
		// 2) DeleteFile("/") won't run after work has started
		if err := pj.driver.PachClient().DeleteFile(
			pj.commitInfo.Commit.Repo.Name,
			pj.commitInfo.Commit.ID,
			"/",
		); err != nil {
			return errors.Wrap(err, "couldn't prepare output commit for S3-out job")
		}
	}

	pj.ji.State = pps.JobState_JOB_RUNNING
	return pj.writeJobInfo()
}

// Iterator fulfills the chain.JobData interface for pendingJob
func (pj *pendingJobV2) Iterator() (datum.Iterator, error) {
	var dit datum.Iterator
	err := pj.logger.LogStep("constructing datum iterator", func() (err error) {
		dit, err = datum.NewIterator(pj.driver.PachClient(), pj.ji.Input)
		return
	})
	return dit, err
}

func (reg *registryV2) processJobRunning(pj *pendingJobV2) error {
	pj.logger.Logf("processJobRunning creating task channel")
	subtasks := make(chan *work.Task, 10)
	eg, ctx := errgroup.WithContext(reg.driver.PachClient().Ctx())
	// Spawn a goroutine to emit tasks on the datum task channel
	eg.Go(func() error {
		defer close(subtasks)
		return pj.logger.LogStep("collecting datums for tasks", func() error {
				// TODO: logic for datum set size.
				if err := pj.jdit.Iterate(ctx, func(iter *datum.Iterator) {
					if err := reg.sendDatumTasksV2(ctx, pj, iter, subtasks); err != nil {
						return err
					}
				}); err != nil {
					return err
				}
			}
		})
	})
	mutex := &sync.Mutex{}
	stats := &DatumStats{ProcessStats: &pps.ProcessStats{}}
	recoveredTags := []string{}
	// Run subtasks until we are done
	eg.Go(func() error {
		return pj.logger.LogStep("running datum tasks", func() error {
			return pj.taskMaster.RunSubtasksChan(
				subtasks,
				func(ctx context.Context, taskInfo *work.TaskInfo) error {
					if taskInfo.State == work.State_FAILURE {
						return errors.Errorf("datum task failed: %s", taskInfo.Reason)
					}
					data, err := deserializeDatumData(taskInfo.Task.Data)
					if err != nil {
						return err
					}
					mutex.Lock()
					defer mutex.Unlock()
					mergeStats(stats, data.Stats)
					if data.RecoveredDatumsTag != "" {
						recoveredTags = append(recoveredTags, data.RecoveredDatumsTag)
					}
					return nil
				},
			)
		})
	})
	err := eg.Wait()
	pj.saveJobStats(stats)
	if err != nil {
		// If these was no failed datum, we can reattempt later
		return errors.Wrap(err, "process datum error")
	}
	// S3Out pipelines don't use hashtrees, so skip over the MERGING state - this
	// will go to EGRESSING, if applicable.
	if pj.driver.PipelineInfo().S3Out {
		if stats.FailedDatumID != "" {
			return reg.failJob(pj, "datum failed", nil, 0)
		}
		pj.logger.Logf("processJobRunning succeeding s3out job, total stats: %v", stats)
		return reg.succeedJob(pj, nil, 0, nil, 0)
	}
	// Write the hashtrees list and recovered datums list to object storage
	if err := pj.storeHashtreeInfos(chunkHashtrees, statsHashtrees); err != nil {
		return err
	}
	if err := pj.storeRecoveredTags(recoveredTags); err != nil {
		return err
	}
	pj.logger.Logf("processJobRunning updating task to merging, total stats: %v", stats)
	pj.ji.State = pps.JobState_JOB_MERGING
	return pj.writeJobInfo()
}

func (reg *registryV2) sendDatumTasks(ctx context.Context, pj *pendingJobV2, iter datum.Iterator, subtasks chan<- *work.Task) error {
	chunkSpec := pj.ji.ChunkSpec
	if chunkSpec == nil {
		chunkSpec = &pps.ChunkSpec{}
	}
	maxDatumsPerTask := int64(chunkSpec.Number)
	maxBytesPerTask := int64(chunkSpec.SizeBytes)
	driver := pj.driver.WithContext(ctx)
	numDatums := iter.Len()
	var numTasks int64
	if numDatums < reg.concurrency {
		numTasks = numDatums
	} else if maxDatumsPerTask > 0 && numDatums/maxDatumsPerTask > reg.concurrency {
		numTasks = numDatums / maxDatumsPerTask
	} else {
		numTasks = reg.concurrency
	}
	datumsPerTask := int64(math.Ceil(float64(numDatums) / float64(numTasks)))
	datumsSize := int64(0)
	datums := []*DatumInputs{}
	// finishTask will finish the currently-writing object and append it to the
	// subtasks, then reset all the relevant variables
	finishTask := func() error {
		putObjectWriter, err := driver.PachClient().PutObjectAsync([]*pfs.Tag{})
		if err != nil {
			return err
		}
		protoWriter := pbutil.NewWriter(putObjectWriter)
		if _, err := protoWriter.Write(&DatumInputsList{Datums: datums}); err != nil {
			putObjectWriter.Close()
			return err
		}
		if err := putObjectWriter.Close(); err != nil {
			return err
		}
		datumsObject, err := putObjectWriter.Object()
		if err != nil {
			return err
		}
		taskData, err := serializeDatumData(&DatumData{Datums: datumsObject, OutputCommit: pj.ji.OutputCommit, JobID: pj.ji.Job.ID})
		if err != nil {
			return err
		}
		select {
		case subtasks <- &work.Task{ID: uuid.NewWithoutDashes(), Data: taskData}:
		case <-ctx.Done():
			return ctx.Err()
		}
		datumsSize = 0
		datums = []*DatumInputs{}
		return nil
	}
	// Build up chunks to be put into work tasks from the datum iterator
	for i := int64(0); i < numDatums; i++ {
		inputs, index := pj.jdit.NextDatum()
		if inputs == nil {
			return errors.New("job datum iterator returned nil inputs")
		}
		datums = append(datums, &DatumInputs{Inputs: inputs, Index: index})
		// If we have enough input bytes, finish the task
		if maxBytesPerTask != 0 {
			for _, input := range inputs {
				datumsSize += int64(input.FileInfo.SizeBytes)
			}
			if datumsSize >= maxBytesPerTask {
				if err := finishTask(); err != nil {
					return err
				}
			}
		}
		// If we hit the upper threshold for task size, finish the task
		if int64(len(datums)) >= datumsPerTask {
			if err := finishTask(); err != nil {
				return err
			}
		}
	}
	if len(datums) > 0 {
		if err := finishTask(); err != nil {
			return err
		}
	}
	return nil
}

func (reg *registryV2) processJobEgress(pj *pendingJobV2) error {
	if err := reg.egress(pj); err != nil {
		return reg.failJob(pj, fmt.Sprintf("egress error: %v", err), nil, 0)
	}

	pj.ji.State = pps.JobState_JOB_SUCCESS
	return pj.writeJobInfo()
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
		if err != nil {
			if vistErr == nil {
				vistErr = errors.Wrapf(err, "error blocking on commit %s/%s",
					commit.Repo.Name, commit.ID)
			}
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

func (reg *registryV2) egress(pj *pendingJobV2) error {
	// copy the pach client (preserving auth info) so we can set a different
	// number of concurrent streams
	pachClient := pj.driver.PachClient().WithCtx(pj.driver.PachClient().Ctx())
	pachClient.SetMaxConcurrentStreams(100)

	var egressFailureCount int
	return backoff.RetryNotify(func() (retErr error) {
		if pj.ji.Egress != nil {
			pj.logger.Logf("Starting egress upload")
			start := time.Now()
			url, err := obj.ParseURL(pj.ji.Egress.URL)
			if err != nil {
				return err
			}
			objClient, err := obj.NewClientFromURLAndSecret(url, false)
			if err != nil {
				return err
			}
			if err := pfssync.PushObj(pachClient, pj.ji.OutputCommit, objClient, url.Object); err != nil {
				return err
			}
			pj.logger.Logf("Completed egress upload, duration (%v)", time.Since(start))
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		egressFailureCount++
		if egressFailureCount > 3 {
			return err
		}
		pj.logger.Logf("egress failed: %v; retrying in %v", err, d)
		return nil
	})
}

func finishJobV2(pipelineInfo *pps.PipelineInfo, pachClient *client.APIClient, jobInfo *pps.JobInfo, state pps.JobState, reason string) {
	// Optimistically update the local state and reason - if any errors occur the
	// local state will be reloaded way up the stack
	jobInfo.State = state
	jobInfo.Reason = reason

	if _, err := pachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
		if pipelineInfo.S3Out {
			if err := builder.FinishCommit(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID); err != nil {
				return err
			}
		} else {
			if jobInfo.StatsCommit != nil {
				if _, err := builder.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
					Commit: jobInfo.StatsCommit,
					Empty:  statsTrees == nil,
				}); err != nil {
					return err
				}
			}

			if _, err := builder.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
				Commit: jobInfo.OutputCommit,
				Empty:  trees == nil,
			}); err != nil {
				return err
			}
		}

		return writeJobInfo(&builder.APIClient, jobInfo)
	}); err != nil {
		if pfsserver.IsCommitFinishedErr(err) || pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
			// For certain types of errors, we want to reattempt these operations
			// outside of a transaction (in case the job or commits were affected by
			// some non-transactional code elsewhere, we can attempt to recover)
			return recoverFinishedJob(pipelineInfo, pachClient, jobInfo, state, reason, datums, trees, size, statsTrees, statsSize)
		}
		// For other types of errors, we want to fail the job supervision and let it
		// reattempt later
		return err
	}
	return nil
}




