package transform

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	pfssync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/src/server/worker/pipeline/transform/chain"
)

func jobTagPrefix(jobID string) string {
	return fmt.Sprintf("job-%s", jobID)
}

func jobHashtreesTag(jobID string) string {
	return fmt.Sprintf("%s-hashtrees", jobTagPrefix(jobID))
}

func jobRecoveredDatumTagsTag(jobID string) string {
	return fmt.Sprintf("%s-recovered", jobTagPrefix(jobID))
}

type pendingJob struct {
	driver     driver.Driver
	commitInfo *pfs.CommitInfo
	cancel     context.CancelFunc
	logger     logs.TaggedLogger
	ji         *pps.JobInfo
	jobData    *types.Any
	jdit       chain.JobDatumIterator

	// These are filled in when the RUNNING phase completes, but may be re-fetched
	// from object storage.
	hashtrees []*HashtreeInfo
}

type registry struct {
	driver      driver.Driver
	logger      logs.TaggedLogger
	workMaster  *work.Master
	mutex       sync.Mutex
	concurrency uint64
	limiter     limit.ConcurrencyLimiter
	jobChain    chain.JobChain
}

type hasher struct {
	name string
	salt string
}

func (h *hasher) Hash(inputs []*common.Input) string {
	return common.HashDatum(h.name, h.salt, inputs)
}

// Returns the registry or lazily instantiates it
func newRegistry(
	logger logs.TaggedLogger,
	driver driver.Driver,
) (*registry, error) {
	// Determine the maximum number of concurrent tasks we will allow
	concurrency, err := driver.GetExpectedNumWorkers()
	if err != nil {
		return nil, err
	}

	return &registry{
		driver:      driver,
		logger:      logger,
		concurrency: concurrency,
		workMaster:  driver.NewTaskMaster(),
		limiter:     limit.New(int(concurrency)),
		jobChain:    nil,
	}, nil
}

// Helper function for succeedJob and failJob, do not use directly
func finishJob(
	pachClient *client.APIClient,
	jobInfo *pps.JobInfo,
	state pps.JobState,
	reason string,
	datums *pfs.Object,
	trees []*pfs.Object,
	size uint64,
	statsTrees []*pfs.Object,
	statsSize uint64,
) error {
	// Optimistically update the local state and reason - if any errors occur the
	// local state will be reloaded way up the stack
	jobInfo.State = state
	jobInfo.Reason = reason

	if _, err := pachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
		if jobInfo.StatsCommit != nil {
			if _, err := builder.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
				Commit:    jobInfo.StatsCommit,
				Empty:     statsTrees == nil,
				Trees:     statsTrees,
				SizeBytes: statsSize,
			}); err != nil {
				return err
			}
		}

		if _, err := builder.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
			Commit:    jobInfo.OutputCommit,
			Empty:     trees == nil,
			Datums:    datums,
			Trees:     trees,
			SizeBytes: size,
		}); err != nil {
			return err
		}

		return writeJobInfo(&builder.APIClient, jobInfo)
	}); err != nil {
		if pfsserver.IsCommitFinishedErr(err) || pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
			// For certain types of errors, we want to reattempt these operations
			// outside of a transaction (in case the job or commits were affected by
			// some non-transactional code elsewhere, we can attempt to recover)
			return recoverFinishedJob(pachClient, jobInfo, state, reason, datums, trees, size, statsTrees, statsSize)
		}
		// For other types of errors, we want to fail the job supervision and let it
		// reattempt later
		return err
	}
	return nil
}

// recoverFinishedJob performs job and output commit updates outside of a
// transaction in an attempt to get everything in a consistent state if they
// were modified non-transactionally elsewhere.
func recoverFinishedJob(
	pachClient *client.APIClient,
	jobInfo *pps.JobInfo,
	state pps.JobState,
	reason string,
	datums *pfs.Object,
	trees []*pfs.Object,
	size uint64,
	statsTrees []*pfs.Object,
	statsSize uint64,
) error {
	if jobInfo.StatsCommit != nil {
		if _, err := pachClient.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
			Commit:    jobInfo.StatsCommit,
			Empty:     statsTrees == nil,
			Trees:     statsTrees,
			SizeBytes: statsSize,
		}); err != nil {
			if !pfsserver.IsCommitFinishedErr(err) && !pfsserver.IsCommitNotFoundErr(err) && !pfsserver.IsCommitDeletedErr(err) {
				return err
			}
		}
	}

	if _, err := pachClient.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
		Commit:    jobInfo.OutputCommit,
		Empty:     trees == nil,
		Datums:    datums,
		Trees:     trees,
		SizeBytes: size,
	}); err != nil {
		if !pfsserver.IsCommitFinishedErr(err) && !pfsserver.IsCommitNotFoundErr(err) && !pfsserver.IsCommitDeletedErr(err) {
			return err
		}
	}

	return writeJobInfo(pachClient, jobInfo)
}

// succeedJob will move a job to the successful state and propagate to any
// dependent jobs in the jobChain.
func (reg *registry) succeedJob(
	pj *pendingJob,
	trees []*pfs.Object,
	size uint64,
	statsTrees []*pfs.Object,
	statsSize uint64,
) error {
	datums, err := reg.storeJobDatums(pj)
	if err != nil {
		return err
	}

	var newState pps.JobState
	if pj.ji.Egress == nil {
		pj.logger.Logf("job successful, closing commits")
		newState = pps.JobState_JOB_SUCCESS
	} else {
		pj.logger.Logf("job successful, advancing to egress")
		newState = pps.JobState_JOB_EGRESS
	}

	if err := finishJob(pj.driver.PachClient(), pj.ji, newState, "", datums, trees, size, statsTrees, statsSize); err != nil {
		return err
	}

	return reg.jobChain.Succeed(pj)
}

func (reg *registry) failJob(
	pj *pendingJob,
	reason string,
) error {
	pj.logger.Logf("failing job with reason: %s", reason)

	if err := finishJob(pj.driver.PachClient(), pj.ji, pps.JobState_JOB_FAILURE, reason, nil, nil, 0, nil, 0); err != nil {
		return err
	}

	return reg.jobChain.Fail(pj)
}

func (reg *registry) killJob(
	pj *pendingJob,
	reason string,
) error {
	pj.logger.Logf("killing job with reason: %s", reason)

	if err := finishJob(pj.driver.PachClient(), pj.ji, pps.JobState_JOB_KILLED, reason, nil, nil, 0, nil, 0); err != nil {
		return err
	}

	return reg.jobChain.Fail(pj)
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

func (pj *pendingJob) writeJobInfo() error {
	pj.logger.Logf("updating job info, state: %s", pj.ji.State)
	return writeJobInfo(pj.driver.PachClient(), pj.ji)
}

func (reg *registry) initializeJobChain(commitInfo *pfs.CommitInfo) error {
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

		reg.jobChain = chain.NewJobChain(
			&hasher{
				name: reg.driver.PipelineInfo().Pipeline.Name,
				salt: reg.driver.PipelineInfo().Salt,
			},
			baseDatums,
		)
	}

	return nil
}

// Generate a datum task (and split it up into subtasks) for the added datums
// in the pending job.
func (reg *registry) makeDatumTask(ctx context.Context, pj *pendingJob, numDatums uint64) (*work.Task, error) {
	chunkSpec := pj.ji.ChunkSpec
	if chunkSpec == nil {
		chunkSpec = &pps.ChunkSpec{}
	}

	maxDatumsPerTask := uint64(chunkSpec.Number)
	maxBytesPerTask := uint64(chunkSpec.SizeBytes)
	driver := pj.driver.WithCtx(ctx)
	var numTasks uint64
	if numDatums < reg.concurrency {
		numTasks = numDatums
	} else if maxDatumsPerTask > 0 && numDatums/maxDatumsPerTask > reg.concurrency {
		numTasks = numDatums / maxDatumsPerTask
	} else {
		numTasks = reg.concurrency
	}
	datumsPerTask := uint64(math.Ceil(float64(numDatums) / float64(numTasks)))

	datumsSize := uint64(0)
	datums := []*DatumInputs{}
	subtasks := []*work.Task{}

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

		taskData, err := serializeDatumData(&DatumData{Datums: datumsObject, OutputCommit: pj.ji.OutputCommit})
		if err != nil {
			return err
		}

		subtasks = append(subtasks, &work.Task{
			ID:   uuid.NewWithoutDashes(),
			Data: taskData,
		})

		datumsSize = 0
		datums = []*DatumInputs{}
		return nil
	}

	// Build up chunks to be put into work tasks from the datum iterator
	for i := uint64(0); i < numDatums; i++ {
		inputs := pj.jdit.NextDatum()
		if inputs == nil {
			return nil, fmt.Errorf("job datum iterator returned nil inputs")
		}

		datums = append(datums, &DatumInputs{Inputs: inputs})

		// If we have enough input bytes, finish the task
		if maxBytesPerTask != 0 {
			for _, input := range inputs {
				datumsSize += input.FileInfo.SizeBytes
			}
			if datumsSize >= maxBytesPerTask {
				if err := finishTask(); err != nil {
					return nil, err
				}
			}
		}

		// If we hit the upper threshold for task size, finish the task
		if uint64(len(datums)) >= datumsPerTask {
			if err := finishTask(); err != nil {
				return nil, err
			}
		}
	}

	if len(datums) > 0 {
		if err := finishTask(); err != nil {
			return nil, err
		}
	}

	return &work.Task{
		ID:       uuid.NewWithoutDashes(),
		Data:     pj.jobData,
		Subtasks: subtasks,
	}, nil
}

func serializeJobData(data *JobData) (*types.Any, error) {
	serialized, err := types.MarshalAny(data)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

func deserializeJobData(any *types.Any) (*JobData, error) {
	data := &JobData{}
	if err := types.UnmarshalAny(any, data); err != nil {
		return nil, err
	}
	return data, nil
}

func serializeDatumData(data *DatumData) (*types.Any, error) {
	serialized, err := types.MarshalAny(data)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

func deserializeDatumData(any *types.Any) (*DatumData, error) {
	data := &DatumData{}
	if err := types.UnmarshalAny(any, data); err != nil {
		return nil, err
	}
	return data, nil
}

func serializeMergeData(data *MergeData) (*types.Any, error) {
	serialized, err := types.MarshalAny(data)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

func deserializeMergeData(any *types.Any) (*MergeData, error) {
	data := &MergeData{}
	if err := types.UnmarshalAny(any, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (reg *registry) getDatumSet(datumsObj *pfs.Object) (_ chain.DatumSet, retErr error) {
	pachClient := reg.driver.PachClient()
	if datumsObj == nil {
		return nil, nil
	}
	r, err := pachClient.GetObjectReader(datumsObj.Hash)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := r.Close(); err != nil && retErr != nil {
			retErr = err
		}
	}()
	pbr := pbutil.NewReader(r)
	datums := make(chain.DatumSet)
	for {
		k, err := pbr.ReadBytes()
		if err != nil {
			if err == io.EOF {
				return datums, retErr
			}
			return nil, err
		}
		datums[string(k)]++
	}
}

// Walk from the given commit back to a successfully completed commit so we can
// get the initial state of datumsBase in the registry.
func (reg *registry) getParentCommitInfo(commitInfo *pfs.CommitInfo) (*pfs.CommitInfo, error) {
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
func (reg *registry) ensureJob(
	commitInfo *pfs.CommitInfo,
	metaCommit *pfs.Commit,
) (*pps.JobInfo, error) {
	pachClient := reg.driver.PachClient()

	// Check if a job was previously created for this commit. If not, make one
	jobInfos, err := pachClient.ListJob("", nil, commitInfo.Commit, -1, true)
	if err != nil {
		return nil, err
	}
	if len(jobInfos) > 1 {
		return nil, fmt.Errorf("multiple jobs found for commit: %s/%s", commitInfo.Commit.Repo.Name, commitInfo.Commit.ID)
	} else if len(jobInfos) < 1 {
		job, err := pachClient.CreateJob(reg.driver.PipelineInfo().Pipeline.Name, commitInfo.Commit, metaCommit)
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

	// get latest job state
	reg.logger.Logf("found existing job %q for output commit %q", jobInfos[0].Job.ID, commitInfo.Commit.ID)
	return pachClient.InspectJob(jobInfos[0].Job.ID, false)
}

func (reg *registry) startJob(commitInfo *pfs.CommitInfo, metaCommit *pfs.Commit) error {
	if err := reg.initializeJobChain(commitInfo); err != nil {
		return err
	}

	jobInfo, err := reg.ensureJob(commitInfo, metaCommit)
	if err != nil {
		return err
	}

	jobCtx, cancel := context.WithCancel(reg.driver.PachClient().Ctx())
	driver := reg.driver.WithCtx(jobCtx)

	// Build the pending job to send out to workers - this will block if we have
	// too many already
	pj := &pendingJob{
		driver:     driver,
		commitInfo: commitInfo,
		cancel:     cancel,
		logger:     reg.logger.WithJob(jobInfo.Job.ID),
		ji:         jobInfo,
	}

	switch {
	case ppsutil.IsTerminal(jobInfo.State):
		// Make sure the output commits are closed
		if err := recoverFinishedJob(pj.driver.PachClient(), pj.ji, jobInfo.State, "", nil, nil, 0, nil, 0); err != nil {
			return err
		}
		// ignore finished jobs (e.g. old pipeline & already killed)
		return nil
	case jobInfo.PipelineVersion < reg.driver.PipelineInfo().Version:
		// kill unfinished jobs from old pipelines (should generally be cleaned
		// up by PPS master, but the PPS master can fail, and if these jobs
		// aren't killed, future jobs will hang indefinitely waiting for their
		// parents to finish)
		pj.ji.State = pps.JobState_JOB_KILLED
		pj.ji.Reason = "pipeline has been updated"
		if err := pj.writeJobInfo(); err != nil {
			return fmt.Errorf("failed to kill stale job: %v", err)
		}
		return nil
	case jobInfo.PipelineVersion > reg.driver.PipelineInfo().Version:
		return fmt.Errorf("job %s's version (%d) greater than pipeline's "+
			"version (%d), this should automatically resolve when the worker "+
			"is updated", jobInfo.Job.ID, jobInfo.PipelineVersion, reg.driver.PipelineInfo().Version)
	}

	pj.jobData, err = serializeJobData(&JobData{JobID: pj.ji.Job.ID})
	if err != nil {
		return fmt.Errorf("failed to serialize job data: %v", err)
	}

	// Inputs must be ready before we can construct a datum iterator, so do this
	// synchronously to ensure correct order in the jobChain.
	if err := pj.logger.LogStep("waiting for job inputs", func() error {
		return reg.processJobStarting(pj)
	}); err != nil {
		return err
	}

	// TODO: cancel job if output commit is finished early

	reg.limiter.Acquire()
	// TODO: we should NOT start the job this way if it is in EGRESS
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

	go func() {
		defer pj.cancel()
		if pj.ji.JobTimeout != nil {
			pj.logger.Logf("cancelling job at: %+v", afterTime)
			timer := time.AfterFunc(afterTime, func() {
				reg.killJob(pj, "job timed out")
			})
			defer timer.Stop()
		}

		backoff.RetryUntilCancel(pj.driver.PachClient().Ctx(), func() error {
			return reg.superviseJob(pj)
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			pj.logger.Logf("error in superviseJob: %v, retrying in %+v", err, d)
			return nil
		})
	}()

	go func() {
		pj.logger.Logf("master running processJobs")
		defer reg.limiter.Release()
		defer pj.cancel()
		backoff.RetryUntilCancel(pj.driver.PachClient().Ctx(), func() error {
			var err error
			for err == nil {
				err = reg.processJob(pj)
			}
			return err
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			pj.logger.Logf("processJob failed: %v; retrying in %v", err, d)

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

			return nil
		})
		pj.logger.Logf("master done running processJobs")
		// TODO: remove job, make sure that all paths close the commit correctly
	}()

	return nil
}

// superviseJob watches for the output commit closing and cancels the job, or
// deletes it if the output commit is removed.
func (reg *registry) superviseJob(pj *pendingJob) error {
	commitInfo, err := pj.driver.PachClient().PfsAPIClient.InspectCommit(pj.driver.PachClient().Ctx(),
		&pfs.InspectCommitRequest{
			Commit:     pj.ji.OutputCommit,
			BlockState: pfs.CommitState_FINISHED,
		})
	if err != nil {
		if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
			defer pj.cancel() // whether we return error or nil, job is done
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
	if commitInfo.Trees == nil {
		defer pj.cancel() // whether job state update succeeds or not, job is done
		return reg.killJob(pj, "output commit closed")
	}
	return nil
}

func (reg *registry) processJob(pj *pendingJob) error {
	state := pj.ji.State
	switch {
	case ppsutil.IsTerminal(state):
		pj.cancel()
		return errutil.ErrBreak
	case state == pps.JobState_JOB_STARTING:
		return fmt.Errorf("job should have been moved out of the STARTING state before processJob")
	case state == pps.JobState_JOB_RUNNING:
		return pj.logger.LogStep("processing job datums", func() error {
			return reg.processJobRunning(pj)
		})
	case state == pps.JobState_JOB_MERGING:
		return pj.logger.LogStep("merging job hashtrees", func() error {
			return reg.processJobMerging(pj)
		})
	case state == pps.JobState_JOB_EGRESS:
		return pj.logger.LogStep("egressing job data", func() error {
			return reg.processJobEgress(pj)
		})
	}
	pj.cancel()
	return fmt.Errorf("unknown job state: %v", state)
}

func (reg *registry) processJobStarting(pj *pendingJob) error {
	// block until job inputs are ready
	failed, err := failedInputs(pj.driver.PachClient(), pj.ji)
	if err != nil {
		return err
	}

	if len(failed) > 0 {
		reason := fmt.Sprintf("inputs failed: %s", strings.Join(failed, ", "))
		return reg.failJob(pj, reason)
	}

	pj.ji.State = pps.JobState_JOB_RUNNING
	return pj.writeJobInfo()
}

// Iterator fulfills the chain.JobData interface for pendingJob
func (pj *pendingJob) Iterator() (datum.Iterator, error) {
	var dit datum.Iterator
	err := pj.logger.LogStep("constructing datum iterator", func() (err error) {
		dit, err = datum.NewIterator(pj.driver.PachClient(), pj.ji.Input)
		return
	})
	return dit, err
}

func (reg *registry) processJobRunning(pj *pendingJob) error {
	pj.logger.Logf("processJobRunning creating task channel")
	datumTaskChan := make(chan *work.Task, 10)

	ctx, cancel := context.WithCancel(reg.driver.PachClient().Ctx())
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	// Spawn a goroutine to emit tasks on the datum task channel
	eg.Go(func() error {
		defer close(datumTaskChan)
		for {
			pj.logger.Logf("processJobRunning making datum task")

			numDatums, err := pj.jdit.NextBatch(ctx)
			if err != nil {
				return err
			}
			if numDatums == 0 {
				return nil
			}

			datumTask, err := reg.makeDatumTask(ctx, pj, numDatums)
			if err != nil {
				return err
			}
			datumTaskChan <- datumTask
		}
	})

	mutex := &sync.Mutex{}
	stats := &DatumStats{ProcessStats: &pps.ProcessStats{}}
	hashtrees := []*HashtreeInfo{}
	recoveredTags := []string{}

	// Run tasks in the datumTaskChan until we are done
	pj.logger.Logf("processJobRunning running datum tasks")
	for task := range datumTaskChan {
		eg.Go(func() error {
			pj.logger.Logf("processJobRunning async task running")
			err := reg.workMaster.Run(
				ctx,
				task,
				func(ctx context.Context, subtask *work.Task) error {
					pj.logger.Logf("datum task complete: %v\n", subtask)
					data, err := deserializeDatumData(subtask.Data)
					if err != nil {
						return err
					}

					mutex.Lock()
					defer mutex.Unlock()

					mergeStats(stats, data.Stats)
					if stats.DatumsFailed > 0 {
						cancel() // Fail fast if any datum failed
						return fmt.Errorf("datum processing failed on datum (%s)", stats.FailedDatumID)
					}

					if data.Hashtree != nil {
						hashtrees = append(hashtrees, data.Hashtree)
					}
					if data.RecoveredDatumsTag != "" {
						recoveredTags = append(recoveredTags, data.RecoveredDatumsTag)
					}
					return nil
				},
			)
			pj.logger.Logf("processJobRunning async task complete")
			return err
		})
	}

	pj.logger.Logf("processJobRunning waiting for task complete")
	// Wait for datums to complete
	err := eg.Wait()
	pj.saveJobStats(stats)
	if err != nil {
		// If we have failed datums in the stats, fail the job, otherwise we can reattempt later
		if stats.FailedDatumID != "" {
			return reg.failJob(pj, "datum failed")
		}
		return fmt.Errorf("process datum error: %v", err)
	}

	// Write the hashtrees list and recovered datums list to object storage
	if err := pj.storeHashtreeInfos(hashtrees); err != nil {
		return err
	}
	if err := pj.storeRecoveredTags(recoveredTags); err != nil {
		return err
	}

	pj.logger.Logf("processJobRunning updating task to merging, total stats: %v", stats)
	pj.ji.State = pps.JobState_JOB_MERGING
	return pj.writeJobInfo()
}

func (pj *pendingJob) saveJobStats(stats *DatumStats) {
	// Any unaccounted-for datums were skipped in the job datum iterator
	pj.ji.DataSkipped = int64(pj.jdit.MaxLen()) - stats.DatumsProcessed - stats.DatumsFailed - stats.DatumsRecovered
	pj.ji.DataProcessed = stats.DatumsProcessed
	pj.ji.DataFailed = stats.DatumsFailed
	pj.ji.DataRecovered = stats.DatumsRecovered
	pj.ji.DataTotal = int64(pj.jdit.MaxLen())
	pj.ji.Stats = stats.ProcessStats
}

func (pj *pendingJob) storeHashtreeInfos(hashtrees []*HashtreeInfo) error {
	pj.hashtrees = hashtrees

	tags := &HashtreeTags{Tags: []string{}}
	for _, info := range hashtrees {
		tags.Tags = append(tags.Tags, info.Tag)
	}

	buf := &bytes.Buffer{}
	pbw := pbutil.NewWriter(buf)
	if _, err := pbw.Write(tags); err != nil {
		return nil
	}

	_, _, err := pj.driver.PachClient().PutObject(buf, jobHashtreesTag(pj.ji.Job.ID))
	return err
}

func (pj *pendingJob) storeRecoveredTags(recoveredTags []string) error {
	if len(recoveredTags) == 0 {
		return nil
	}

	tags := &RecoveredDatumTags{Tags: recoveredTags}

	buf := &bytes.Buffer{}
	pbw := pbutil.NewWriter(buf)
	if _, err := pbw.Write(tags); err != nil {
		return nil
	}

	_, _, err := pj.driver.PachClient().PutObject(buf, jobRecoveredDatumTagsTag(pj.ji.Job.ID))
	return err
}

func (pj *pendingJob) initializeHashtrees() error {
	if pj.hashtrees == nil {
		// We are picking up an old job and don't have the hashtrees generated by
		// the 'running' state, load them from object storage
		buf := &bytes.Buffer{}
		if err := pj.driver.PachClient().GetTag(jobHashtreesTag(pj.ji.Job.ID), buf); err != nil {
			return err
		}
		protoReader := pbutil.NewReader(buf)

		hashtreeTags := &HashtreeTags{}
		if err := protoReader.Read(hashtreeTags); err != nil {
			return err
		}

		hashtreeInfos := []*HashtreeInfo{}
		for _, tag := range hashtreeTags.Tags {
			hashtreeInfos = append(hashtreeInfos, &HashtreeInfo{Tag: tag})
		}

		pj.hashtrees = hashtreeInfos
	}
	return nil
}

func (pj *pendingJob) loadRecoveredDatums() (chain.DatumSet, error) {
	datumSet := make(chain.DatumSet)

	// We are picking up an old job and don't have the recovered datums generated by
	// the 'running' state, load them from object storage
	buf := &bytes.Buffer{}
	if err := pj.driver.PachClient().GetTag(jobRecoveredDatumTagsTag(pj.ji.Job.ID), buf); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return datumSet, nil
		}
		return nil, err
	}
	protoReader := pbutil.NewReader(buf)

	recoveredTags := &RecoveredDatumTags{}
	if err := protoReader.Read(recoveredTags); err != nil {
		return nil, err
	}

	recoveredDatums := &RecoveredDatums{}

	for _, tag := range recoveredTags.Tags {
		buf := &bytes.Buffer{}
		if err := pj.driver.PachClient().GetTag(tag, buf); err != nil {
			return nil, err
		}

		protoReader := pbutil.NewReader(buf)
		if err := protoReader.Read(recoveredDatums); err != nil {
			return nil, err
		}

		for _, hash := range recoveredDatums.Hashes {
			datumSet[hash]++
		}
	}

	return datumSet, nil
}

func (reg *registry) processJobMerging(pj *pendingJob) error {
	if err := pj.initializeHashtrees(); err != nil {
		return err
	}

	// For jobs that can base their hashtree off of the parent hashtree, fetch the
	// object information for the parent hashtrees
	// TODO: this is risky - there's no check that the parent the jobChain is
	// thinking of is the same one we find.  Add some extra guarantees here if we can.
	var parentHashtrees []*pfs.Object
	if pj.jdit.AdditiveOnly() {
		parentCommitInfo, err := reg.getParentCommitInfo(pj.commitInfo)
		if err != nil {
			return err
		}

		if parentCommitInfo != nil {
			parentHashtrees = parentCommitInfo.Trees
			if len(parentHashtrees) != int(reg.driver.NumShards()) {
				return fmt.Errorf("unexpected number of hashtrees between the parent commit (%d) and the pipeline spec (%d)", len(parentHashtrees), reg.driver.NumShards())
			}
		}
	}

	if parentHashtrees != nil {
		pj.logger.Logf("merging %d hashtrees with parent hashtree across %d shards", len(pj.hashtrees), reg.driver.NumShards())
	} else {
		pj.logger.Logf("merging %d hashtrees across %d shards", len(pj.hashtrees), reg.driver.NumShards())
	}

	mergeSubtasks := []*work.Task{}
	for i := int64(0); i < reg.driver.NumShards(); i++ {
		mergeData := &MergeData{Hashtrees: pj.hashtrees, Shard: i}

		if parentHashtrees != nil {
			mergeData.Parent = parentHashtrees[i]
		}

		data, err := serializeMergeData(mergeData)
		if err != nil {
			return fmt.Errorf("failed to serialize merge data: %v", err)
		}

		mergeSubtasks = append(mergeSubtasks, &work.Task{
			ID:   uuid.NewWithoutDashes(),
			Data: data,
		})
	}

	trees := make([]*pfs.Object, reg.driver.NumShards())
	size := uint64(0)

	pj.logger.Logf("sending out %d merge tasks", len(trees))

	// Generate merge task and wait for merges to complete
	if err := reg.workMaster.Run(
		reg.driver.PachClient().Ctx(),
		&work.Task{
			ID:       uuid.NewWithoutDashes(),
			Data:     pj.jobData,
			Subtasks: mergeSubtasks,
		},
		func(ctx context.Context, subtask *work.Task) error {
			data, err := deserializeMergeData(subtask.Data)
			if err != nil {
				return err
			}
			if data.Tree == nil {
				return fmt.Errorf("merge task for shard %d failed, no tree returned", data.Shard)
			}
			trees[data.Shard] = data.Tree
			size += data.TreeSize
			return nil
		},
	); err != nil {
		// TODO: persist error to job?
		return fmt.Errorf("merge error: %v", err)
	}

	if err := reg.succeedJob(pj, trees, size, []*pfs.Object{}, 0); err != nil {
		return err
	}

	reg.mutex.Lock()
	// TODO: propagate failed datums to downstream job

	// TODO: close downstream job's datumTask channel
	defer reg.mutex.Unlock()
	return nil
}

func (reg *registry) storeJobDatums(pj *pendingJob) (*pfs.Object, error) {
	// Update recovered datums through the job chain, which will update its internal DatumSet
	recoveredDatums, err := pj.loadRecoveredDatums()
	if err != nil {
		return nil, err
	}

	if err := reg.jobChain.RecoveredDatums(pj, recoveredDatums); err != nil {
		return nil, err
	}

	// Write out the datums processed/skipped and merged for this job
	buf := &bytes.Buffer{}
	pbw := pbutil.NewWriter(buf)
	for hash, count := range pj.jdit.DatumSet() {
		for i := uint64(0); i < count; i++ {
			if _, err := pbw.WriteBytes([]byte(hash)); err != nil {
				return nil, err
			}
		}
	}
	datums, _, err := pj.driver.PachClient().PutObject(buf)
	return datums, err
}

func (reg *registry) processJobEgress(pj *pendingJob) error {
	if err := reg.egress(pj); err != nil {
		return reg.failJob(pj, fmt.Sprintf("egress error: %v", err))
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
				vistErr = fmt.Errorf("error blocking on commit %s/%s: %v",
					commit.Repo.Name, commit.ID, err)
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

func (reg *registry) egress(pj *pendingJob) error {
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
