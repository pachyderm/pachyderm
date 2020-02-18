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
)

const maxDatumsPerTask = 10000

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
	driver        driver.Driver
	commitInfo    *pfs.CommitInfo
	cancel        context.CancelFunc
	logger        logs.TaggedLogger
	index         int
	orphan        bool
	ji            *pps.JobInfo
	jobData       *types.Any
	datumsAdded   map[string]struct{}
	datumsRemoved map[string]struct{}
	datumTaskChan chan *work.Task
	dit           datum.Iterator

	// These are filled in when the RUNNING phase completes, but may be re-fetched
	// from object storage.
	hashtrees       []*HashtreeInfo
	datumsRecovered map[string]struct{}
}

type registry struct {
	driver      driver.Driver
	logger      logs.TaggedLogger
	workMaster  *work.Master
	mutex       sync.Mutex
	concurrency int
	limiter     limit.ConcurrencyLimiter

	datumsBase map[string]struct{}
	jobs       []*pendingJob
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
		limiter:     limit.New(concurrency),
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

		if _, err := builder.PpsAPIClient.UpdateJobState(pachClient.Ctx(), &pps.UpdateJobStateRequest{
			Job:    jobInfo.Job,
			State:  state,
			Reason: reason,
		}); err != nil {
			return err
		}

		return nil
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

	_, err := pachClient.PpsAPIClient.UpdateJobState(pachClient.Ctx(), &pps.UpdateJobStateRequest{
		Job:    jobInfo.Job,
		State:  state,
		Reason: reason,
	})
	return err
}

// finishJob will transactionally finish the output and meta commit for the job,
// and optionally set the job's state.
func (reg *registry) succeedJob(
	pj *pendingJob,
	datums *pfs.Object,
	trees []*pfs.Object,
	size uint64,
	statsTrees []*pfs.Object,
	statsSize uint64,
) error {
	newState := pps.JobState_JOB_SUCCESS
	if pj.ji.Egress != nil {
		newState = pps.JobState_JOB_EGRESS
	}
	if err := finishJob(pj.driver.PachClient(), pj.ji, newState, "", datums, trees, size, statsTrees, statsSize); err != nil {
		return err
	}

	recoveredDatums, err := pj.loadRecoveredDatums()
	if err != nil {
		return err
	}

	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	// For a successful job, we must merge added and removed datums into the registry's base datum set
	if pj.orphan {
		reg.datumsBase = pj.datumsAdded
	} else {
		// TODO: it is possible for the child of an orphan job to finish before the first job in the queue
		if pj.index != 0 || reg.jobs[0] != pj {
			return fmt.Errorf("unexpected job ordering, aborting")
		}

		for hash, _ := range pj.datumsAdded {
			reg.datumsBase[hash] = struct{}{}
		}

		for hash, _ := range pj.datumsRemoved {
			delete(reg.datumsBase[hash])
		}
	}

	for hash, _ := range recoveredDatums {
		delete(reg.datumsBase[hash])
	}

	// Recovered datums must be passed on to the next job
	if len(reg.jobs) > 0 && len(recoveredDatums) > 0 {
		err := pj.initializeIterator()
		if err != nil {
			return err
		}

		// TODO: construct tasks for downstream jobs
		for hash, _ := range recoveredDatums {

		}
	}

	return nil
}

func (reg *registry) failJob(
	pj *pendingJob,
	reason string,
) error {
	if err := finishJob(pj.driver.PachClient(), pj.ji, pps.JobState_JOB_FAILURE, reason, nil, nil, 0, nil, 0); err != nil {
		return err
	}

	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	// TODO: forward the added/removed datums to the next job in the queue, send a new datum task

	return nil
}

func (pj *pendingJob) writeJobInfo() error {
	_, err := pj.driver.NewSTM(func(stm col.STM) error {
		pipelines := pj.driver.Pipelines().ReadWrite(stm)
		jobs := pj.driver.Jobs().ReadWrite(stm)

		etcdJobInfo := &pps.EtcdJobInfo{}
		if err := jobs.Get(pj.ji.Job.ID, etcdJobInfo); err != nil {
			return err
		}

		etcdJobInfo.Restart = pj.ji.Restart
		etcdJobInfo.DataProcessed = pj.ji.DataProcessed
		etcdJobInfo.DataSkipped = pj.ji.DataSkipped
		etcdJobInfo.DataTotal = pj.ji.DataTotal
		etcdJobInfo.DataFailed = pj.ji.DataFailed
		etcdJobInfo.DataRecovered = pj.ji.DataRecovered
		etcdJobInfo.Stats = pj.ji.Stats

		return ppsutil.UpdateJobState(pipelines, jobs, etcdJobInfo, pj.ji.State, pj.ji.Reason)
	})
	return err
}

func (reg *registry) initializeDatumsBase(commitInfo *pfs.CommitInfo) error {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	if reg.datumsBase == nil {
		// Get the most recent successful commit starting from the given commit
		parentCommitInfo, err := reg.getParentCommitInfo(commitInfo)
		if err != nil {
			return err
		}

		// Initialize the datumsBase value in the registry
		if parentCommitInfo != nil {
			var err error
			reg.datumsBase, err = reg.getDatumSet(parentCommitInfo.Datums)
			if err != nil {
				return err
			}
		} else {
			reg.datumsBase = make(map[string]struct{})
		}
	}

	return nil
}

func isDatumNew(hash string, ancestorJobs []*pendingJob, datumsBase map[string]struct{}) bool {
	for i := len(ancestorJobs) - 1; i >= 0; i-- {
		if _, ok := ancestorJobs[i].datumsAdded[hash]; ok {
			return false
		}
		if ancestorJobs[i].orphan {
			// No other jobs matter since this job doesn't depend on any previous
			// hashtree, so we can shortcut out
			return true
		}
		if _, ok := ancestorJobs[i].datumsRemoved[hash]; ok {
			return true
		}
	}

	if _, ok := datumsBase[hash]; ok {
		return false
	}
	return true
}

// calculateRemovedDatums walks back through the sets of added and removed
// datums in upstream pending jobs and determine which datums were removed for
// the latest pending job. This is done so we can properly reprocess a datum if
// it is removed then readded.
func calculateRemovedDatums(
	allDatums map[string]struct{},
	ancestorJobs []*pendingJob,
	datumsBase map[string]struct{},
) map[string]struct{} {
	removed := make(map[string]struct{})
	skip := make(map[string]struct{})

	for i := len(ancestorJobs) - 1; i >= 0; i-- {
		for hash := range ancestorJobs[i].datumsAdded {
			if _, ok := skip[hash]; !ok {
				skip[hash] = struct{}{}
				if _, ok := allDatums[hash]; !ok {
					removed[hash] = struct{}{}
				}
			}
		}
		if ancestorJobs[i].orphan {
			// No other jobs matter since this job doesn't depend on any previous
			// hashtree, so we can shortcut out
			return removed
		}
		for hash := range ancestorJobs[i].datumsRemoved {
			skip[hash] = struct{}{}
		}
	}

	for hash := range datumsBase {
		if _, ok := skip[hash]; !ok {
			skip[hash] = struct{}{}
			if _, ok := allDatums[hash]; !ok {
				removed[hash] = struct{}{}
			}
		}
	}

	return removed
}

// Based on the current state of the registry and pending jobs, calculate the
// set of new and removed datums in the datum iterator. Also add the input
// metadata for new datums to the metadata commit.
func calculateDatumSets(
	pipelineInfo *pps.PipelineInfo,
	dit datum.Iterator,
	ancestorJobs []*pendingJob,
	datumsBase map[string]struct{},
) (datumsAdded map[string]struct{}, datumsRemoved map[string]struct{}, orphan bool) {
	allDatums := make(map[string]struct{})
	datumsAdded = make(map[string]struct{})

	// Iterate over datum iterator, comparing to upstream jobs datum sets to find
	// new datums
	dit.Reset()
	for dit.Next() {
		hash := common.HashDatum(pipelineInfo.Pipeline.Name, pipelineInfo.Salt, dit.Datum())

		allDatums[hash] = struct{}{}

		// Check if the hash was added or removed in any upstream jobs
		if isDatumNew(hash, ancestorJobs, datumsBase) {
			datumsAdded[hash] = struct{}{}
		}
	}

	// If datumsAdded == allDatums, this job does not depend on a parent job, unlink it
	orphan = len(datumsAdded) == len(allDatums)

	return datumsAdded, calculateRemovedDatums(allDatums, ancestorJobs, datumsBase), orphan
}

// Generate a datum task (and split it up into subtasks) for the added datums
// in the pending job.
func (reg *registry) makeDatumTask(
	pj *pendingJob,
) (*work.Task, error) {
	pipelineInfo := reg.driver.PipelineInfo()

	var numTasks int
	if len(pj.datumsAdded) < reg.concurrency {
		numTasks = len(pj.datumsAdded)
	} else if len(pj.datumsAdded)/maxDatumsPerTask > reg.concurrency {
		numTasks = len(pj.datumsAdded) / maxDatumsPerTask
	} else {
		numTasks = reg.concurrency
	}
	datumsPerTask := int(math.Ceil(float64(len(pj.datumsAdded)) / float64(numTasks)))

	datums := []*DatumInputs{}
	subtasks := []*work.Task{}

	// finishTask will finish the currently-writing object and append it to the
	// subtasks, then reset all the relevant variables
	finishTask := func() error {
		putObjectWriter, err := reg.driver.PachClient().PutObjectAsync([]*pfs.Tag{})
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

		datums = []*DatumInputs{}
		return nil
	}

	fmt.Printf("iterating over dit with length %d\n", pj.dit.Len())
	// Iterate over the datum iterator and append the inputs for added datums to
	// the current object
	pj.dit.Reset()
	for pj.dit.Next() {
		fmt.Printf("dit datum: %v\n", pj.dit.Datum())
		inputs := pj.dit.Datum()
		hash := common.HashDatum(pipelineInfo.Pipeline.Name, pipelineInfo.Salt, inputs)

		if _, ok := pj.datumsAdded[hash]; ok {
			datums = append(datums, &DatumInputs{Inputs: inputs})

			// If we hit the upper threshold for task size, finish the task
			if len(datums) == datumsPerTask {
				if err := finishTask(); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(datums) > 0 {
		if err := finishTask(); err != nil {
			return nil, err
		}
	}

	fmt.Printf("returning work task, subtasks: %d\n", len(subtasks))
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

func (reg *registry) getDatumSet(datumsObj *pfs.Object) (_ map[string]struct{}, retErr error) {
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
	datums := make(map[string]struct{})
	for {
		k, err := pbr.ReadBytes()
		if err != nil {
			if err == io.EOF {
				return datums, retErr
			}
			return nil, err
		}
		datums[string(k)] = struct{}{}
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
	if err := reg.initializeDatumsBase(commitInfo); err != nil {
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
		if err := reg.driver.UpdateJobState(jobInfo.Job.ID,
			pps.JobState_JOB_KILLED, "pipeline has been updated"); err != nil {
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

	if err := pj.logger.LogStep("waiting for job inputs (JOB_STARTING)", func() error {
		return reg.processJobStarting(pj)
	}); err != nil {
		return err
	}

	// TODO: cancel job if output commit is finished early

	reg.limiter.Acquire()
	reg.mutex.Lock()
	pj.index = len(reg.jobs)
	reg.jobs = append(reg.jobs, pj)
	reg.mutex.Unlock()

	go func() {
		pj.logger.Logf("master running processJobs")
		defer reg.limiter.Release()
		defer pj.cancel()
		backoff.RetryUntilCancel(pj.driver.PachClient().Ctx(), func() error {
			for {
				err := reg.processJob(pj)
				pj.logger.Logf("processJob result: %v, state: %v", err, pj.ji.State)
				if err != nil {
					return err
				}
			}
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			pj.logger.Logf("processJob failed: %v; retrying in %v", err, d)

			// Increment the job's restart count
			_, err = pj.driver.NewSTM(func(stm col.STM) error {
				jobs := pj.driver.Jobs().ReadWrite(stm)
				jobID := pj.ji.Job.ID
				jobPtr := &pps.EtcdJobInfo{}
				if err := jobs.Get(jobID, jobPtr); err != nil {
					return err
				}
				jobPtr.Restart++
				// TODO: copy the up-to-date etcd job info into the pj.ji struct, we may
				// be recovering from a failed update
				return jobs.Put(jobID, jobPtr)
			})
			if err != nil {
				pj.logger.Logf("error incrementing restart count for job (%s)", pj.ji.Job.ID)
			}
			return nil
		})
		pj.logger.Logf("master done running processJobs")
		// TODO: remove job, make sure that all paths close the commit correctly
	}()

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
		return pj.logger.LogStep("processing job datums (JOB_RUNNING)", func() error {
			return reg.processJobRunning(pj)
		})
	case state == pps.JobState_JOB_MERGING:
		return pj.logger.LogStep("merging job hashtrees (JOB_MERGING)", func() error {
			return reg.processJobMerging(pj)
		})
	case state == pps.JobState_JOB_EGRESS:
		return pj.logger.LogStep("egressing job data (JOB_EGRESS)", func() error {
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

func (pj *pendingJob) initializeIterator() error {
	if pj.dit == nil {
		// Create a datum iterator pointing at the job's inputs
		return pj.logger.LogStep("constructing datum iterator", func() error {
			var err error
			pj.dit, err = datum.NewIterator(pj.driver.PachClient(), pj.ji.Input)
			return err
		})
	}
	return nil
}

func (reg *registry) processJobRunning(pj *pendingJob) error {
	if err := pj.initializeIterator(); err != nil {
		return err
	}

	pj.logger.Logf("processJobRunning creating task channel")
	reg.mutex.Lock()
	datumTaskChan := make(chan *work.Task, 10)

	pj.logger.Logf("processJobRunning calculating datum sets")
	pj.datumsAdded, pj.datumsRemoved, pj.orphan = calculateDatumSets(
		reg.driver.PipelineInfo(),
		pj.dit,
		reg.jobs[0:pj.index],
		reg.datumsBase,
	)

	pj.logger.Logf("processJobRunning making datum task")
	datumTask, err := reg.makeDatumTask(pj)
	if err != nil {
		reg.mutex.Unlock()
		close(datumTaskChan)
		return err
	}
	pj.logger.Logf("sending datum task")
	datumTaskChan <- datumTask
	pj.logger.Logf("done sending datum task")
	if pj.orphan || pj.index == 0 {
		pj.logger.Logf("processJobRunning closing orphan channel")
		close(datumTaskChan)
	} else {
		pj.datumTaskChan = datumTaskChan

		defer func() {
			reg.mutex.Lock()
			defer reg.mutex.Unlock()

			if pj.datumTaskChan != nil {
				close(pj.datumTaskChan)
				pj.datumTaskChan = nil
			}
		}()
	}

	reg.mutex.Unlock()

	stats := &DatumStats{ProcessStats: &pps.ProcessStats{}}
	hashtrees := []*HashtreeInfo{}
	recoveredTags := []string{}

	ctx, cancel := context.WithCancel(reg.driver.PachClient().Ctx())
	eg, ctx := errgroup.WithContext(ctx)

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
					mergeStats(stats, data.Stats)
					if stats.DatumsFailed > 0 {
						cancel() // Fail fast if any datum failed
						return fmt.Errorf("datum processing failed on datum (%s)", stats.FailedDatumID)
					} else {
						if data.Hashtree != nil {
							hashtrees = append(hashtrees, data.Hashtree)
						}
						if data.RecoveredDatumsTag != "" {
							recoveredTags = append(recoveredTags, data.RecoveredDatumsTag)
						}
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
	if err := eg.Wait(); err != nil {
		// If we have failed datums in the stats, fail the job, otherwise we can reattempt later
		if stats.FailedDatumID != "" {
			fmt.Printf("failing job\n")
			return reg.failJob(pj, "datum failed")
		}
		return fmt.Errorf("process datum error: %v", err)
	}

	if len(pj.datumsRemoved) > 0 {
		// In order to merge, we'll need to generate hashtrees for the unprocessed datums as well
		return fmt.Errorf("removing datums not yet implemented")
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

func (pj *pendingJob) loadRecoveredDatums() (map[string]struct{}, error) {
	if pj.datumsRecovered == nil {
		// We are picking up an old job and don't have the recovered datums generated by
		// the 'running' state, load them from object storage
		buf := &bytes.Buffer{}
		if err := pj.driver.PachClient().GetTag(jobRecoveredDatumTagsTag(pj.ji.Job.ID), buf); err != nil {
			if strings.Contains(err.Error(), "not found") {
				pj.datumsRecovered = make(map[string]struct{})
				return pj.datumsRecovered, nil
			}
			return nil, err
		}
		protoReader := pbutil.NewReader(buf)

		recoveredTags := &RecoveredDatumTags{}
		if err := protoReader.Read(recoveredTags); err != nil {
			return nil, err
		}

		pj.datumsRecovered = make(map[string]struct{})
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
				pj.datumsRecovered[hash] = struct{}{}
			}
		}
	}

	return pj.datumsRecovered, nil
}

func (reg *registry) processJobMerging(pj *pendingJob) error {
	if err := pj.initializeHashtrees(); err != nil {
		return err
	}

	// For jobs that can base their hashtree off of the parent hashtree, fetch the
	// object information for the parent hashtrees
	var parentHashtrees []*pfs.Object
	if !pj.orphan && len(pj.datumsRemoved) == 0 {
		parentCommitInfo, err := reg.getParentCommitInfo(pj.commitInfo)
		if err != nil {
			return err
		}

		parentHashtrees = parentCommitInfo.Trees
		if len(parentHashtrees) != int(reg.driver.NumShards()) {
			return fmt.Errorf("unexpected number of hashtrees between the parent commit (%d) and the pipeline spec (%d)", len(parentHashtrees), reg.driver.NumShards())
		}
		pj.logger.Logf("merging %d hashtrees with parent hashtree across %d shards", len(pj.hashtrees), reg.driver.NumShards())
	} else {
		pj.logger.Logf("merging %d hashtrees across %d shards", len(pj.hashtrees), reg.driver.NumShards())
	}

	mergeSubtasks := []*work.Task{}
	for i := int64(0); i < reg.driver.NumShards(); i++ {
		mergeData := &MergeData{Hashtrees: pj.hashtrees, Shard: i}

		if parentHashtrees != nil {
			mergeData.Parent = &HashtreeInfo{Object: parentHashtrees[i]}
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

	datums, err := reg.storeJobDatums(pj)
	if err != nil {
		return err
	}

	if err := reg.succeedJob(pj, datums, trees, size, []*pfs.Object{}, 0); err != nil {
		return err
	}

	reg.mutex.Lock()
	// TODO: propagate failed datums to downstream job

	// TODO: close downstream job's datumTask channel
	defer reg.mutex.Unlock()
	return nil
}

func (reg *registry) storeJobDatums(pj *pendingJob) (*pfs.Object, error) {
	if err := pj.initializeIterator(); err != nil {
		return nil, err
	}

	recoveredDatums, err := pj.loadRecoveredDatums()
	if err != nil {
		return nil, err
	}

	// Write out the datums processed/skipped and merged for this job
	buf := &bytes.Buffer{}
	pbw := pbutil.NewWriter(buf)
	for i := 0; i < pj.dit.Len(); i++ {
		files := pj.dit.DatumN(i)
		datumHash := common.HashDatum(reg.driver.PipelineInfo().Pipeline.Name, reg.driver.PipelineInfo().Salt, files)
		// recovered datums were not processed, and so we won't write them to the processed datums object
		if _, ok := recoveredDatums[datumHash]; !ok {
			if _, err := pbw.WriteBytes([]byte(datumHash)); err != nil {
				return nil, err
			}
		}
	}
	datums, _, err := pj.driver.PachClient().PutObject(buf)
	return datums, err
}

func (reg *registry) processJobEgress(pj *pendingJob) error {
	if err := reg.egress(pj); err != nil {
		return fmt.Errorf("egress error: %v", err)
	}

	pj.ji.State = pps.JobState_JOB_SUCCESS
	if err := pj.writeJobInfo(); err != nil {
		return fmt.Errorf("failed to update job state to success: %v", err)
	}

	return nil
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
