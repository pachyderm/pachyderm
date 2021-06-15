package transform

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/client/limit"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/work"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/pipeline/transform/chain"
)

// TODO: Job failures are propagated through commits with pfs.EmptyStr in the description, would be better to have general purpose metadata associated with a commit.

const (
	defaultDatumSetsPerWorker int64 = 4
)

type hasher struct {
	name string
	salt string
}

func (h *hasher) Hash(inputs []*common.Input) string {
	return common.HashDatum(h.name, h.salt, inputs)
}

type pendingJob struct {
	driver                     driver.Driver
	logger                     logs.TaggedLogger
	ji                         *pps.JobInfo
	commitInfo, metaCommitInfo *pfs.CommitInfo
	jdit                       *chain.JobDatumIterator
	taskMaster                 *work.Master
	cancel                     context.CancelFunc
}

func (pj *pendingJob) writeJobInfo() error {
	pj.logger.Logf("updating job info, state: %s", pj.ji.State)
	return ppsutil.WriteJobInfo(pj.driver.PachClient(), pj.ji)
}

// TODO: The job info should eventually just have a field with type *datum.Stats
func (pj *pendingJob) saveJobStats(stats *datum.Stats) {
	// TODO: Need to clean up the setup of process stats.
	if pj.ji.Stats == nil {
		pj.ji.Stats = &pps.ProcessStats{}
	}
	datum.MergeProcessStats(pj.ji.Stats, stats.ProcessStats)
	pj.ji.DataProcessed += stats.Processed
	pj.ji.DataSkipped += stats.Skipped
	pj.ji.DataFailed += stats.Failed
	pj.ji.DataRecovered += stats.Recovered
	pj.ji.DataTotal += stats.Processed + stats.Skipped + stats.Failed + stats.Recovered
}

func (pj *pendingJob) clearJobStats() {
	pj.ji.Stats = &pps.ProcessStats{}
	pj.ji.DataProcessed = 0
	pj.ji.DataSkipped = 0
	pj.ji.DataFailed = 0
	pj.ji.DataRecovered = 0
	pj.ji.DataTotal = 0
}

func (pj *pendingJob) withDeleter(pachClient *client.APIClient, cb func() error) error {
	defer pj.jdit.SetDeleter(nil)
	// Setup file operation client for output Meta commit.
	metaCommit := pj.metaCommitInfo.Commit
	return pachClient.WithModifyFileClient(metaCommit, func(mfMeta client.ModifyFile) error {
		// Setup file operation client for output PFS commit.
		outputCommit := pj.commitInfo.Commit
		return pachClient.WithModifyFileClient(outputCommit, func(mfPFS client.ModifyFile) error {
			parentMetaCommit := pj.metaCommitInfo.ParentCommit
			metaFileWalker := func(path string) ([]string, error) {
				var files []string
				if err := pachClient.WalkFile(parentMetaCommit, path, func(fi *pfs.FileInfo) error {
					if fi.FileType == pfs.FileType_FILE {
						files = append(files, fi.File.Path)
					}
					return nil
				}); err != nil {
					return nil, err
				}
				return files, nil
			}
			pj.jdit.SetDeleter(datum.NewDeleter(metaFileWalker, mfMeta, mfPFS))
			return cb()
		})
	})
}

type registry struct {
	driver      driver.Driver
	logger      logs.TaggedLogger
	taskQueue   *work.TaskQueue
	concurrency int64
	limiter     limit.ConcurrencyLimiter
	jobChain    *chain.JobChain
}

// TODO:
// Prometheus stats? (previously in the driver, which included testing we should reuse if possible)
// capture logs (reuse driver tests and reintroduce tagged logger).
func newRegistry(driver driver.Driver, logger logs.TaggedLogger) (*registry, error) {
	// Determine the maximum number of concurrent tasks we will allow
	concurrency, err := driver.ExpectedNumWorkers()
	if err != nil {
		return nil, err
	}
	taskQueue, err := driver.NewTaskQueue()
	if err != nil {
		return nil, err
	}
	return &registry{
		driver:      driver,
		logger:      logger,
		taskQueue:   taskQueue,
		concurrency: concurrency,
		limiter:     limit.New(int(concurrency)),
	}, nil
}

func (reg *registry) succeedJob(pj *pendingJob) error {
	pj.logger.Logf("job successful, closing commits")
	defer pj.jdit.Finish()
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	return ppsutil.FinishJob(reg.driver.PachClient(), pj.ji, pps.JobState_JOB_SUCCESS, "")
}

func (reg *registry) failJob(pj *pendingJob, reason string) error {
	pj.logger.Logf("failing job with reason: %s", reason)
	defer pj.jdit.Finish()
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	return ppsutil.FinishJob(reg.driver.PachClient(), pj.ji, pps.JobState_JOB_FAILURE, reason)
}

func (reg *registry) killJob(pj *pendingJob, reason string) error {
	pj.logger.Logf("killing job with reason: %s", reason)
	defer pj.jdit.Finish()
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	_, err := reg.driver.PachClient().PpsAPIClient.StopJob(
		reg.driver.PachClient().Ctx(),
		&pps.StopJobRequest{
			Job:    pj.ji.Job,
			Reason: reason,
		},
	)
	return err
}

func (reg *registry) initializeJobChain(metaCommitInfo *pfs.CommitInfo) error {
	if reg.jobChain == nil {
		pi := reg.driver.PipelineInfo()
		hasher := &hasher{
			name: pi.Pipeline.Name,
			salt: pi.Salt,
		}
		var opts []chain.JobChainOption
		if pi.ReprocessSpec == client.ReprocessSpecEveryJob || pi.S3Out {
			opts = append(opts, chain.WithNoSkip())
		}
		pachClient := reg.driver.PachClient()
		if metaCommitInfo.ParentCommit == nil {
			reg.jobChain = chain.NewJobChain(pachClient, hasher, opts...)
			return nil
		}
		parentMetaCommitInfo, err := pachClient.PfsAPIClient.InspectCommit(pachClient.Ctx(),
			&pfs.InspectCommitRequest{
				Commit: metaCommitInfo.ParentCommit,
				Wait:   pfs.CommitState_FINISHED,
			})
		if err != nil {
			return err
		}
		commit := parentMetaCommitInfo.Commit
		reg.jobChain = chain.NewJobChain(
			pachClient,
			hasher,
			append(opts, chain.WithBase(datum.NewCommitIterator(pachClient, commit)))...,
		)
	}
	return nil
}

func (reg *registry) startJob(jobInfo *pps.JobInfo) error {
	var asyncEg *errgroup.Group
	reg.limiter.Acquire()
	defer func() {
		if asyncEg == nil {
			// The async errgroup never got started, so give up the limiter lock
			reg.limiter.Release()
		}
	}()
	commitInfo, err := reg.driver.PachClient().PfsAPIClient.InspectCommit(
		reg.driver.PachClient().Ctx(),
		&pfs.InspectCommitRequest{
			Commit: jobInfo.OutputCommit,
			Wait:   pfs.CommitState_STARTED,
		})
	if err != nil {
		return err
	}
	metaCommitInfo, err := reg.driver.PachClient().PfsAPIClient.InspectCommit(
		reg.driver.PachClient().Ctx(),
		&pfs.InspectCommitRequest{
			Commit: ppsutil.MetaCommit(jobInfo.OutputCommit),
			Wait:   pfs.CommitState_STARTED,
		})
	if err != nil {
		return err
	}
	if err := reg.initializeJobChain(metaCommitInfo); err != nil {
		return err
	}
	asyncStarted := false
	jobCtx, cancel := context.WithCancel(reg.driver.PachClient().Ctx())
	// Don't leak the cancellation if we error before starting the async code
	defer func() {
		if !asyncStarted {
			cancel()
		}
	}()
	driver := reg.driver.WithContext(jobCtx)
	// Build the pending job to send out to workers - this will block if
	// we have too many already
	pj := &pendingJob{
		driver:         driver,
		ji:             jobInfo,
		logger:         reg.logger.WithJob(jobInfo.Job.ID),
		commitInfo:     commitInfo,
		metaCommitInfo: metaCommitInfo,
		cancel:         cancel,
	}
	if jobInfo.State == pps.JobState_JOB_CREATED {
		jobInfo.State = pps.JobState_JOB_STARTING
		if err := pj.writeJobInfo(); err != nil {
			return err
		}
	}
	// Inputs must be ready before we can construct a datum iterator, so do this
	// synchronously to ensure correct order in the jobChain.
	if err := pj.logger.LogStep("waiting for job inputs", func() error {
		return reg.processJobStarting(pj)
	}); err != nil {
		return err
	}
	// TODO: This could probably be scoped to a callback, and we could move job specific features
	// in the chain package (timeouts for example).
	// TODO: I use the registry pachclient for the iterators, so I can reuse across jobs for skipping.
	pachClient := reg.driver.PachClient()
	dit, err := datum.NewIterator(pachClient, pj.ji.Input)
	if err != nil {
		return err
	}
	outputDit := datum.NewCommitIterator(pachClient, pj.metaCommitInfo.Commit)
	pj.jdit = reg.jobChain.CreateJob(pj.driver.PachClient().Ctx(), pj.ji.Job, dit, outputDit)
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
	asyncStarted = true

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
				if errors.Is(err, errutil.ErrBreak) {
					return nil
				}
				return err
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				pj.logger.Logf("processJob error: %v, retrying in %v", err, d)
				for err != nil {
					if st, ok := err.(errors.StackTracer); ok {
						pj.logger.Logf("error stack: %+v", st.StackTrace())
					}
					err = errors.Unwrap(err)
				}
				// Get job state, increment restarts, write job state
				pj.ji, err = pj.driver.PachClient().InspectJob(pj.ji.Job.Pipeline.Name, pj.ji.Job.ID, false)
				if err != nil {
					return err
				}
				pj.clearJobStats()
				pj.ji.Restart++
				if err := pj.writeJobInfo(); err != nil {
					pj.logger.Logf("error incrementing restart count for job (%s): %v", pj.ji.Job.ID, err)
				}
				// Reload the job's commitInfo(s) as they may have changed and clear the state of the commit(s).
				pachClient := reg.driver.PachClient()
				pj.commitInfo, err = pachClient.PfsAPIClient.InspectCommit(
					pachClient.Ctx(),
					&pfs.InspectCommitRequest{
						Commit: pj.commitInfo.Commit,
						Wait:   pfs.CommitState_STARTED,
					})
				if err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				if _, err := pachClient.PfsAPIClient.ClearCommit(
					pachClient.Ctx(),
					&pfs.ClearCommitRequest{
						Commit: pj.commitInfo.Commit,
					}); err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				pj.metaCommitInfo, err = pachClient.PfsAPIClient.InspectCommit(
					pachClient.Ctx(),
					&pfs.InspectCommitRequest{
						Commit: pj.metaCommitInfo.Commit,
						Wait:   pfs.CommitState_STARTED,
					})
				if err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				_, err = pachClient.PfsAPIClient.ClearCommit(
					pachClient.Ctx(),
					&pfs.ClearCommitRequest{
						Commit: pj.metaCommitInfo.Commit,
					})
				return grpcutil.ScrubGRPC(err)
			})
			pj.logger.Logf("master done running processJobs")
		}); err != nil {
			return err
		}
		// This should block until the callback has completed
		mutex.Lock()
		return nil
	})

	go func() {
		defer reg.limiter.Release()
		// Make sure the job has been removed from the job chain.
		defer pj.jdit.Finish()
		if err := asyncEg.Wait(); err != nil {
			pj.logger.Logf("fatal job error: %v", err)
		}
	}()
	return nil
}

// superviseJob watches for the output commit closing and cancels the job, or
// deletes it if the output commit is removed.
func (reg *registry) superviseJob(pj *pendingJob) error {
	defer pj.cancel()
	ci, err := pj.driver.PachClient().PfsAPIClient.InspectCommit(pj.driver.PachClient().Ctx(),
		&pfs.InspectCommitRequest{
			Commit: pj.ji.OutputCommit,
			Wait:   pfs.CommitState_FINISHED,
		})
	if err != nil {
		if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
			// Stop the job and clean up any job state in the registry
			if err := reg.killJob(pj, "output commit missing"); err != nil {
				return err
			}
			// Output commit was deleted. Delete job as well
			if err := pj.driver.NewSQLTx(func(sqlTx *sqlx.Tx) error {
				// Delete the job if no other worker has deleted it yet
				jobPtr := &pps.StoredJobInfo{}
				if err := pj.driver.Jobs().ReadWrite(sqlTx).Get(ppsdb.JobKey(pj.ji.Job), jobPtr); err != nil {
					return err
				}
				return pj.driver.DeleteJob(sqlTx, jobPtr)
			}); err != nil && !col.IsErrNotFound(err) {
				return err
			}
			return nil
		}
		return err
	}
	if strings.Contains(ci.Description, pfs.EmptyStr) {
		return reg.killJob(pj, "output commit closed")
	}
	return nil

}

func (reg *registry) processJob(pj *pendingJob) error {
	state := pj.ji.State
	if ppsutil.IsTerminal(state) {
		return errutil.ErrBreak
	}
	switch state {
	case pps.JobState_JOB_STARTING:
		return errors.New("job should have been moved out of the STARTING state before processJob")
	case pps.JobState_JOB_RUNNING:
		return pj.logger.LogStep("processing job running", func() error {
			return reg.processJobRunning(pj)
		})
	case pps.JobState_JOB_EGRESSING:
		return pj.logger.LogStep("processing job egressing", func() error {
			return reg.processJobEgressing(pj)
		})
	}
	panic(fmt.Sprintf("unrecognized job state: %s", state))
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

// TODO:
// Need to put some more thought into the context use.
func (reg *registry) processJobRunning(pj *pendingJob) error {
	pachClient := pj.driver.PachClient()
	eg, ctx := errgroup.WithContext(pachClient.Ctx())
	pachClient = pachClient.WithCtx(ctx)
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	// TODO: We need to delete the output for S3Out since we don't have a clear way to track the output in the stats commit (which means datums cannot be skipped with S3Out).
	// If we had a way to map the output added through the S3 gateway back to the datums, and stored this in the appropriate place in the stats commit, then we would be able
	// handle datums the same way we handle normal pipelines.
	if pj.driver.PipelineInfo().S3Out {
		if err := pachClient.DeleteFile(pj.commitInfo.Commit, "/"); err != nil {
			return err
		}
	}
	// Generate the deletion operations and count the number of datums for the job.
	var numDatums int64
	if err := pj.logger.LogStep("computing skipped and deleted datums", func() error {
		return pj.withDeleter(pachClient, func() error {
			return pj.jdit.Iterate(func(_ *datum.Meta) error {
				numDatums++
				return nil
			})
		})
	}); err != nil {
		return err
	}
	pj.saveJobStats(pj.jdit.Stats())
	if err := pj.writeJobInfo(); err != nil {
		return err
	}
	// Set up the datum set spec for the job.
	// When the datum set spec is not set, evenly distribute the datums.
	var setSpec *datum.SetSpec
	datumSetsPerWorker := defaultDatumSetsPerWorker
	if pj.driver.PipelineInfo().DatumSetSpec != nil {
		setSpec = &datum.SetSpec{
			Number:    pj.driver.PipelineInfo().DatumSetSpec.Number,
			SizeBytes: pj.driver.PipelineInfo().DatumSetSpec.SizeBytes,
		}
		datumSetsPerWorker = pj.driver.PipelineInfo().DatumSetSpec.PerWorker
	}
	if setSpec == nil || (setSpec.Number == 0 && setSpec.SizeBytes == 0) {
		setSpec = &datum.SetSpec{Number: numDatums / (int64(reg.concurrency) * datumSetsPerWorker)}
		if setSpec.Number == 0 {
			setSpec.Number = 1
		}
	}
	// Setup datum set subtask channel.
	subtasks := make(chan *work.Task)
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		// Setup goroutine for creating datum set subtasks.
		eg.Go(func() error {
			defer close(subtasks)
			storageRoot := filepath.Join(pj.driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
			return datum.CreateSets(pj.jdit, storageRoot, setSpec, func(upload func(client.ModifyFile) error) error {
				subtask, err := createDatumSetSubtask(pachClient, pj, upload, renewer)
				if err != nil {
					return err
				}
				select {
				case subtasks <- subtask:
				case <-ctx.Done():
					return ctx.Err()
				}
				return nil
			})
		})
		// Setup goroutine for running and collecting datum set subtasks.
		eg.Go(func() error {
			return pj.logger.LogStep("running and collecting datum set subtasks", func() error {
				return pj.taskMaster.RunSubtasksChan(
					subtasks,
					func(ctx context.Context, taskInfo *work.TaskInfo) error {
						if taskInfo.State == work.State_FAILURE {
							return errors.Errorf("datum set subtask failed: %s", taskInfo.Reason)
						}
						data, err := deserializeDatumSet(taskInfo.Task.Data)
						if err != nil {
							return err
						}
						renewer.Remove(data.FilesetId)
						if _, err := pachClient.PfsAPIClient.AddFileset(
							pachClient.Ctx(),
							&pfs.AddFilesetRequest{
								Commit:    pj.commitInfo.Commit,
								FilesetId: data.OutputFilesetId,
							},
						); err != nil {
							return grpcutil.ScrubGRPC(err)
						}
						if _, err := pachClient.PfsAPIClient.AddFileset(
							pachClient.Ctx(),
							&pfs.AddFilesetRequest{
								Commit:    pj.metaCommitInfo.Commit,
								FilesetId: data.MetaFilesetId,
							},
						); err != nil {
							return grpcutil.ScrubGRPC(err)
						}
						if err := datum.MergeStats(stats, data.Stats); err != nil {
							return err
						}
						pj.saveJobStats(data.Stats)
						return pj.writeJobInfo()
					},
				)
			})
		})
		return eg.Wait()
	}); err != nil {
		return err
	}
	// TODO: This shouldn't be necessary.
	select {
	case <-pj.driver.PachClient().Ctx().Done():
		return pj.driver.PachClient().Ctx().Err()
	default:
	}
	if stats.FailedID != "" {
		return reg.failJob(pj, fmt.Sprintf("datum %v failed", stats.FailedID))
	}
	if pj.ji.Egress != nil {
		pj.ji.State = pps.JobState_JOB_EGRESSING
		return pj.writeJobInfo()
	}
	return reg.succeedJob(pj)
}

func createDatumSetSubtask(pachClient *client.APIClient, pj *pendingJob, upload func(client.ModifyFile) error, renewer *renew.StringSet) (*work.Task, error) {
	resp, err := pachClient.WithCreateFilesetClient(func(mf client.ModifyFile) error {
		return upload(mf)
	})
	if err != nil {
		return nil, err
	}
	renewer.Add(resp.FilesetId)
	data, err := serializeDatumSet(&DatumSet{
		JobID:        pj.ji.Job.ID,
		OutputCommit: pj.commitInfo.Commit,
		// TODO: It might make sense for this to be a hash of the constituent datums?
		// That could make it possible to recover from a master restart.
		FilesetId: resp.FilesetId,
	})
	if err != nil {
		return nil, err
	}
	return &work.Task{
		// TODO: Should this just be a uuid?
		ID:   uuid.NewWithoutDashes(),
		Data: data,
	}, nil
}

func serializeDatumSet(data *DatumSet) (*types.Any, error) {
	serialized, err := types.MarshalAny(data)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

func deserializeDatumSet(any *types.Any) (*DatumSet, error) {
	data := &DatumSet{}
	if err := types.UnmarshalAny(any, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (reg *registry) processJobEgressing(pj *pendingJob) error {
	url := pj.ji.Egress.URL
	if err := pj.driver.PachClient().GetFileURL(pj.commitInfo.Commit, "/", url); err != nil {
		return err
	}
	return reg.succeedJob(pj)
}

func failedInputs(pachClient *client.APIClient, jobInfo *pps.JobInfo) ([]string, error) {
	var failed []string
	waitCommit := func(name string, commit *pfs.Commit) error {
		ci, err := pachClient.WaitCommit(commit.Branch.Repo.Name, commit.Branch.Name, commit.ID)
		if err != nil {
			return errors.Wrapf(err, "error blocking on commit %s", pfsdb.CommitKey(commit))
		}
		if strings.Contains(ci.Description, pfs.EmptyStr) {
			failed = append(failed, name)
		}
		return nil
	}
	visitErr := pps.VisitInput(jobInfo.Input, func(input *pps.Input) error {
		if input.Pfs != nil && input.Pfs.Commit != "" {
			if err := waitCommit(input.Pfs.Name, client.NewCommit(input.Pfs.Repo, input.Pfs.Branch, input.Pfs.Commit)); err != nil {
				return err
			}
		}
		return nil
	})
	if visitErr != nil {
		return nil, visitErr
	}
	return failed, nil
}
