package transform

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
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

type pendingJob struct {
	client        *client.APIClient
	cancel        context.CancelFunc
	logger        logs.TaggedLogger
	index         int
	orphan        bool
	ji            *pps.JobInfo
	datumsAdded   map[string]struct{}
	datumsRemoved map[string]struct{}
	datumTaskChan chan *work.Task
}

type registry struct {
	driver       driver.Driver
	logger       logs.TaggedLogger
	workMaster   *work.Master
	mutex        sync.Mutex
	concurrency  int
	numHashtrees int64
	limiter      limit.ConcurrencyLimiter
	datumsBase   map[string]struct{}
	jobs         []*pendingJob
}

// Returns the registry or lazily instantiates it
func newRegistry(
	logger logs.TaggedLogger,
	driver driver.Driver,
) (*registry, error) {
	// Determine the maximum number of concurrent tasks we will allow
	concurrency, err := driver.KubeWrapper().GetExpectedNumWorkers(driver.PipelineInfo().ParallelismSpec)
	if err != nil {
		return nil, err
	}

	numHashtrees, err := ppsutil.GetExpectedNumHashtrees(driver.PipelineInfo().HashtreeSpec)
	if err != nil {
		return nil, err
	}

	return &registry{
		driver:       driver,
		logger:       logger,
		concurrency:  concurrency,
		numHashtrees: numHashtrees,
		workMaster:   driver.NewTaskMaster(),
		limiter:      limit.New(concurrency),
	}, nil
}

// finishJob will transactionally finish the output and meta commit for the job,
// and optionally set the job's state.
func (reg *registry) finishJob(pj *pendingJob) error {
	// TODO: accept hashtrees for commits
	_, err := pj.client.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
		if pj.ji.StatsCommit != nil {
			if _, err := builder.PfsAPIClient.FinishCommit(pj.client.Ctx(), &pfs.FinishCommitRequest{
				Commit: pj.ji.StatsCommit,
				Empty:  true,
			}); err != nil {
				return err
			}
		}

		if _, err := builder.PfsAPIClient.FinishCommit(pj.client.Ctx(), &pfs.FinishCommitRequest{
			Commit: pj.ji.OutputCommit,
			Empty:  true,
		}); err != nil {
			return err
		}

		if _, err := builder.PpsAPIClient.UpdateJobState(pj.client.Ctx(), &pps.UpdateJobStateRequest{
			Job:    pj.ji.Job,
			State:  pj.ji.State,
			Reason: pj.ji.Reason,
		}); err != nil {
			return err
		}

		return nil
	})

	// If failed, forward the added/removed datums to the next job in the queue

	// If successful, assert this is the first job in the queue (or is an orphan) and merge the

	return err
}

func (reg *registry) initializeDatumsBase(commitInfo *pfs.CommitInfo) error {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

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

	return nil
}

func isDatumNew(hash string, parentJobs []*pendingJob, datumsBase map[string]struct{}) bool {
	for i := len(parentJobs) - 1; i >= 0; i-- {
		if _, ok := parentJobs[i].datumsAdded[hash]; ok {
			return false
		}
		if parentJobs[i].orphan {
			// No other jobs matter since this job doesn't depend on any previous
			// hashtree, so we can shortcut out
			return true
		}
		if _, ok := parentJobs[i].datumsRemoved[hash]; ok {
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
	parentJobs []*pendingJob,
	datumsBase map[string]struct{},
) map[string]struct{} {
	removed := make(map[string]struct{})
	skip := make(map[string]struct{})

	for i := len(parentJobs) - 1; i >= 0; i-- {
		for hash := range parentJobs[i].datumsAdded {
			if _, ok := skip[hash]; !ok {
				skip[hash] = struct{}{}
				if _, ok := allDatums[hash]; !ok {
					removed[hash] = struct{}{}
				}
			}
		}
		if parentJobs[i].orphan {
			// No other jobs matter since this job doesn't depend on any previous
			// hashtree, so we can shortcut out
			return removed
		}
		for hash := range parentJobs[i].datumsRemoved {
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
	parentJobs []*pendingJob,
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
		if isDatumNew(hash, parentJobs, datumsBase) {
			datumsAdded[hash] = struct{}{}
		}
	}

	// If datumsAdded == allDatums, this job does not depend on a parent job, unlink it
	orphan = len(datumsAdded) == len(allDatums)

	return datumsAdded, calculateRemovedDatums(allDatums, parentJobs, datumsBase), orphan
}

// Generate a datum task (and split it up into subtasks) for the added datums
// in the pending job.
func (reg *registry) makeDatumTask(
	datumsAdded map[string]struct{},
	dit datum.Iterator,
) (*work.Task, error) {
	pipelineInfo := reg.driver.PipelineInfo()

	var numTasks int
	if len(datumsAdded) < reg.concurrency {
		numTasks = len(datumsAdded)
	} else if len(datumsAdded)/maxDatumsPerTask > reg.concurrency {
		numTasks = len(datumsAdded) / maxDatumsPerTask
	} else {
		numTasks = reg.concurrency
	}
	datumsPerTask := int(math.Ceil(float64(numTasks) / float64(len(datumsAdded))))

	var putObjectWriter *client.PutObjectWriteCloserAsync
	var protoWriter pbutil.Writer

	makeWriter := func() (err error) {
		putObjectWriter, err = reg.driver.PachClient().PutObjectAsync([]*pfs.Tag{})
		if err != nil {
			return err
		}
		protoWriter = pbutil.NewWriter(putObjectWriter)
		return nil
	}

	subtasks := []*work.Task{}
	taskLen := 0

	// finishTask will finish the currently-writing object and append it to the
	// subtasks, then reset all the relevant variables
	finishTask := func() error {
		err := putObjectWriter.Close()
		if err != nil {
			return err
		}
		datums, err := putObjectWriter.Object()
		if err != nil {
			return err
		}

		taskData, err := serializeDatumData(&DatumData{Datums: datums})
		if err != nil {
			return err
		}

		subtasks = append(subtasks, &work.Task{
			Id:   uuid.NewWithoutDashes(),
			Data: taskData,
		})

		taskLen = 0
		protoWriter = nil
		putObjectWriter = nil
		return nil
	}

	// Iterate over the datum iterator and append the inputs for added datums to
	// the current object
	dit.Reset()
	for dit.Next() {
		inputs := dit.Datum()
		hash := common.HashDatum(pipelineInfo.Pipeline.Name, pipelineInfo.Salt, inputs)

		if _, ok := datumsAdded[hash]; ok {
			if protoWriter == nil {
				if err := makeWriter(); err != nil {
					return nil, err
				}
			}
			if _, err := protoWriter.Write(&DatumInputs{Inputs: inputs}); err != nil {
				putObjectWriter.Close()
				return nil, err
			}
			taskLen++
		}

		// If we hit the upper threshold for task size, finish the task
		if taskLen == datumsPerTask {
			if err := finishTask(); err != nil {
				return nil, err
			}
		}
	}

	if taskLen > 0 {
		if err := finishTask(); err != nil {
			return nil, err
		}
	}

	return &work.Task{
		Id:       uuid.NewWithoutDashes(),
		Subtasks: subtasks,
	}, nil
}

func serializeDatumData(data *DatumData) (*types.Any, error) {
	serialized, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(data),
		Value:   serialized,
	}, nil
}

func deserializeDatumData(any *types.Any) (*DatumData, error) {
	data := &DatumData{}
	if err := types.UnmarshalAny(any, data); err != nil {
		return nil, err
	}
	return data, nil
}

func serializeMergeData(data *MergeData) (*types.Any, error) {
	serialized, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(data),
		Value:   serialized,
	}, nil
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
	pachClient := reg.driver.PachClient()

	if reg.datumsBase == nil {
		if err := reg.initializeDatumsBase(commitInfo); err != nil {
			return err
		}
	}

	jobInfo, err := reg.ensureJob(commitInfo, metaCommit)
	if err != nil {
		return err
	}

	switch {
	case ppsutil.IsTerminal(jobInfo.State):
		// Make sure the output commits are closed
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

	jobCtx, cancel := context.WithCancel(pachClient.Ctx())
	jobClient := pachClient.WithCtx(jobCtx)

	// Build the pending job to send out to workers - this will block if we have
	// too many already
	pj := &pendingJob{
		client: jobClient,
		cancel: cancel,
		logger: reg.logger.WithJob(jobInfo.Job.ID),
		ji:     jobInfo,
	}

	// Watch output commit - if it gets finished early we need to abort
	go func() {
		pj.logger.Logf("master watching for output commit closure")
		backoff.RetryUntilCancel(pj.client.Ctx(), func() error {
			commitInfo, err := pj.client.PfsAPIClient.InspectCommit(pj.client.Ctx(),
				&pfs.InspectCommitRequest{
					Commit:     pj.ji.OutputCommit,
					BlockState: pfs.CommitState_FINISHED,
				})
			pj.logger.Logf("master got output commit, info: %v, err: %v", commitInfo, err)
			if err != nil {
				if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
					defer pj.cancel() // whether we return error or nil, job is done
					// Output commit was deleted. Delete job as well
					// TODO: this isn't tied to the right ctx
					if _, err := reg.driver.NewSTM(func(stm col.STM) error {
						// Delete the job if no other worker has deleted it yet
						jobPtr := &pps.EtcdJobInfo{}
						if err := reg.driver.Jobs().ReadWrite(stm).Get(pj.ji.Job.ID, jobPtr); err != nil {
							return err
						}
						return reg.driver.DeleteJob(stm, jobPtr)
					}); err != nil && !col.IsErrNotFound(err) {
						return err
					}
					return nil
				}
				return err
			}
			if commitInfo.Trees == nil {
				defer pj.cancel() // whether job state update succeeds or not, job is done
				// TODO: make sure stats commit and job are finished
			}
			return nil
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			return nil // retry again
		})
		pj.logger.Logf("master done watching for output commit closure")
	}()

	reg.limiter.Acquire()
	reg.mutex.Lock()
	pj.index = len(reg.jobs)
	reg.jobs = append(reg.jobs, pj)
	reg.mutex.Unlock()

	go func() {
		pj.logger.Logf("master running processJobs")
		defer reg.limiter.Release()
		defer pj.cancel()
		backoff.RetryUntilCancel(pj.client.Ctx(), func() error {
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
			// TODO: this uses the wrong ctx
			_, err = reg.driver.NewSTM(func(stm col.STM) error {
				jobs := reg.driver.Jobs().ReadWrite(stm)
				jobID := pj.ji.Job.ID
				jobPtr := &pps.EtcdJobInfo{}
				if err := jobs.Get(jobID, jobPtr); err != nil {
					return err
				}
				jobPtr.Restart++
				return jobs.Put(jobID, jobPtr)
			})
			if err != nil {
				pj.logger.Logf("error incrementing restart count for job (%s)", pj.ji.Job.ID)
			}
			return nil
		})
		pj.logger.Logf("master done running processJobs")
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
		return reg.processJobStarting(pj)
	case state == pps.JobState_JOB_RUNNING:
		return reg.processJobRunning(pj)
	case state == pps.JobState_JOB_MERGING:
		return reg.processJobMerging(pj)
	}
	pj.cancel()
	return fmt.Errorf("unknown job state: %v", state)
}

func updateJobState(pachClient *client.APIClient, jobInfo *pps.JobInfo) error {
	_, err := pachClient.PpsAPIClient.UpdateJobState(pachClient.Ctx(), &pps.UpdateJobStateRequest{
		Job:    jobInfo.Job,
		State:  jobInfo.State,
		Reason: jobInfo.Reason,
	})
	return err
}

func (reg *registry) processJobStarting(pj *pendingJob) error {
	// block until job inputs are ready
	failed, err := failedInputs(pj.client, pj.ji)
	if err != nil {
		return err
	}

	if len(failed) > 0 {
		pj.ji.Reason = fmt.Sprintf("inputs %s failed", strings.Join(failed, ", "))
		pj.ji.State = pps.JobState_JOB_FAILURE
	} else {
		pj.ji.State = pps.JobState_JOB_RUNNING
	}

	return updateJobState(pj.client, pj.ji)
}

func (reg *registry) processJobRunning(pj *pendingJob) error {
	// Create a datum iterator pointing at the job's inputs
	dit, err := datum.NewIterator(pj.client, pj.ji.Input)
	if err != nil {
		return err
	}

	reg.mutex.Lock()
	pj.datumTaskChan = make(chan *work.Task)
	defer func() {
		close(pj.datumTaskChan)
		pj.datumTaskChan = nil
	}()

	pj.datumsAdded, pj.datumsRemoved, pj.orphan = calculateDatumSets(
		reg.driver.PipelineInfo(),
		dit,
		reg.jobs[0:pj.index],
		reg.datumsBase,
	)

	datumTask, err := reg.makeDatumTask(pj.datumsAdded, dit)
	if err != nil {
		reg.mutex.Unlock()
		return err
	}
	pj.datumTaskChan <- datumTask
	if pj.orphan || pj.index == 0 {
		close(pj.datumTaskChan)
	}

	reg.mutex.Unlock()

	processStats := &pps.ProcessStats{}

	eg := errgroup.Group{}

	// Run tasks in the datumTaskChan until we are done
	for task := range pj.datumTaskChan {
		eg.Go(func() error {
			return reg.workMaster.Run(
				reg.driver.PachClient().Ctx(),
				task,
				func(ctx context.Context, subtask *work.Task) error {
					pj.logger.Logf("datum task complete: %v\n", subtask)
					data, err := deserializeDatumData(subtask.Data)
					if err != nil {
						return err
					}
					mergeStats(processStats, data.ProcessStats)
					return nil
				},
			)
		})
	}

	// Wait for datums to complete
	if err := eg.Wait(); err != nil {
		// TODO: some sort of error state?
		return fmt.Errorf("process datum error: %v", err)
	}

	pj.ji.State = pps.JobState_JOB_MERGING
	return updateJobState(pj.client, pj.ji)
}

func (reg *registry) processJobMerging(pj *pendingJob) error {
	mergeSubtasks := []*work.Task{}
	for i := int64(0); i < reg.numHashtrees; i++ {
		data, err := serializeMergeData(&MergeData{Shard: i})
		if err != nil {
			return fmt.Errorf("failed to serialize merge task: %v", err)
		}

		mergeSubtasks = append(mergeSubtasks, &work.Task{
			Id:   uuid.NewWithoutDashes(),
			Data: data,
		})
	}

	// Generate merge task
	mergeTask := &work.Task{
		Id:       uuid.NewWithoutDashes(),
		Subtasks: mergeSubtasks,
	}

	// Wait for merges to complete
	err := reg.workMaster.Run(
		reg.driver.PachClient().Ctx(),
		mergeTask,
		func(ctx context.Context, subtask *work.Task) error {
			fmt.Printf("merge task complete: %v\n", subtask)
			// TODO: something
			return nil
		},
	)
	if err != nil {
		// TODO: persist error to job?
		return fmt.Errorf("merge error: %v", err)
	}

	if err := reg.egress(pj); err != nil {
		return fmt.Errorf("egress error: %v", err)
	}

	pj.ji.State = pps.JobState_JOB_SUCCESS
	if err := updateJobState(pj.client, pj.ji); err != nil {
		return fmt.Errorf("failed to update job state to success: %v", err)
	}

	reg.mutex.Lock()
	// TODO: propagate failed datums to downstream job

	// TODO: close downstream job's datumTask channel

	// TODO: remove job
	defer reg.mutex.Unlock()
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
	pachClient := pj.client.WithCtx(pj.client.Ctx())
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
