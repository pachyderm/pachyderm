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

// TODO: PipelineJob failures are propagated through commits with pfs.EmptyStr in the description, would be better to have general purpose metadata associated with a commit.

type hasher struct {
	name string
	salt string
}

func (h *hasher) Hash(inputs []*common.Input) string {
	return common.HashDatum(h.name, h.salt, inputs)
}

type pendingPipelineJob struct {
	driver                     driver.Driver
	logger                     logs.TaggedLogger
	pji                        *pps.PipelineJobInfo
	commitInfo, metaCommitInfo *pfs.CommitInfo
	jdit                       *chain.JobDatumIterator
	taskMaster                 *work.Master
	cancel                     context.CancelFunc
}

func (ppj *pendingPipelineJob) writeJobInfo() error {
	ppj.logger.Logf("updating job info, state: %s", ppj.pji.State)
	return ppsutil.WriteJobInfo(ppj.driver.PachClient(), ppj.pji)
}

// TODO: The job info should eventually just have a field with type *datum.Stats
func (ppj *pendingPipelineJob) saveJobStats(stats *datum.Stats) {
	// TODO: Need to clean up the setup of process stats.
	if ppj.pji.Stats == nil {
		ppj.pji.Stats = &pps.ProcessStats{}
	}
	datum.MergeProcessStats(ppj.pji.Stats, stats.ProcessStats)
	ppj.pji.DataProcessed += stats.Processed
	ppj.pji.DataSkipped += stats.Skipped
	ppj.pji.DataFailed += stats.Failed
	ppj.pji.DataRecovered += stats.Recovered
	ppj.pji.DataTotal += stats.Processed + stats.Skipped + stats.Failed + stats.Recovered
}

func (ppj *pendingPipelineJob) withDeleter(pachClient *client.APIClient, cb func() error) error {
	defer ppj.jdit.SetDeleter(nil)
	// Setup file operation client for output Meta commit.
	metaCommit := ppj.metaCommitInfo.Commit
	return pachClient.WithModifyFileClient(metaCommit, func(mfMeta client.ModifyFile) error {
		// Setup file operation client for output PFS commit.
		outputCommit := ppj.commitInfo.Commit
		return pachClient.WithModifyFileClient(outputCommit, func(mfPFS client.ModifyFile) error {
			parentMetaCommit := ppj.metaCommitInfo.ParentCommit
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
			ppj.jdit.SetDeleter(datum.NewDeleter(metaFileWalker, mfMeta, mfPFS))
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

func (reg *registry) succeedPipelineJob(ppj *pendingPipelineJob) error {
	ppj.logger.Logf("pipeline job successful, closing commits")
	defer ppj.jdit.Finish()
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	return ppsutil.FinishPipelineJob(reg.driver.PachClient(), ppj.pji, pps.PipelineJobState_JOB_SUCCESS, "")
}

func (reg *registry) failPipelineJob(ppj *pendingPipelineJob, reason string) error {
	ppj.logger.Logf("failing pipeline job with reason: %s", reason)
	defer ppj.jdit.Finish()
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	return ppsutil.FinishPipelineJob(reg.driver.PachClient(), ppj.pji, pps.PipelineJobState_JOB_FAILURE, reason)
}

func (reg *registry) killPipelineJob(ppj *pendingPipelineJob, reason string) error {
	ppj.logger.Logf("killing pipeline job with reason: %s", reason)
	defer ppj.jdit.Finish()
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	_, err := reg.driver.PachClient().PpsAPIClient.StopPipelineJob(
		reg.driver.PachClient().Ctx(),
		&pps.StopPipelineJobRequest{
			PipelineJob: ppj.pji.PipelineJob,
			Reason:      reason,
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
				Commit:     metaCommitInfo.ParentCommit,
				BlockState: pfs.CommitState_FINISHED,
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

func (reg *registry) ensurePipelineJob(commitInfo *pfs.CommitInfo) (*pps.PipelineJobInfo, error) {
	pachClient := reg.driver.PachClient()
	pipelineInfo := reg.driver.PipelineInfo()
	pipelineJobInfo, err := pachClient.InspectPipelineJobOutputCommit(pipelineInfo.Pipeline.Name, commitInfo.Commit.Branch.Name, commitInfo.Commit.ID, false)
	if err != nil {
		// TODO: It would be better for this to be a structured error.
		if strings.Contains(err.Error(), "not found") {
			pipelineJob, err := pachClient.CreatePipelineJob(pipelineInfo.Pipeline.Name, commitInfo.Commit, ppsutil.GetStatsCommit(commitInfo))
			if err != nil {
				return nil, err
			}
			pipelineJobInfo, err = pachClient.InspectPipelineJob(pipelineJob.ID, false)
			if err != nil {
				return nil, err
			}
			reg.logger.Logf("created new pipeline job %q for output commit %q", pipelineJobInfo.PipelineJob.ID, pipelineJobInfo.OutputCommit.ID)
			return pipelineJobInfo, nil
		}
		return nil, err
	}
	reg.logger.Logf("found existing pipeline job %q for output commit %q", pipelineJobInfo.PipelineJob.ID, commitInfo.Commit.ID)
	return pipelineJobInfo, nil
}

func (reg *registry) startPipelineJob(commitInfo *pfs.CommitInfo) error {
	var asyncEg *errgroup.Group
	reg.limiter.Acquire()
	defer func() {
		if asyncEg == nil {
			// The async errgroup never got started, so give up the limiter lock
			reg.limiter.Release()
		}
	}()
	pipelineJobInfo, err := reg.ensurePipelineJob(commitInfo)
	if err != nil {
		return err
	}
	metaCommitInfo, err := reg.driver.PachClient().PfsAPIClient.InspectCommit(
		reg.driver.PachClient().Ctx(),
		&pfs.InspectCommitRequest{
			Commit:     pipelineJobInfo.StatsCommit,
			BlockState: pfs.CommitState_STARTED,
		})
	if err != nil {
		return err
	}
	if err := reg.initializeJobChain(metaCommitInfo); err != nil {
		return err
	}
	jobCtx, cancel := context.WithCancel(reg.driver.PachClient().Ctx())
	driver := reg.driver.WithContext(jobCtx)
	// Build the pending pipeline job to send out to workers - this will block if
	// we have too many already
	ppj := &pendingPipelineJob{
		driver:         driver,
		pji:            pipelineJobInfo,
		logger:         reg.logger.WithPipelineJob(pipelineJobInfo.PipelineJob.ID),
		commitInfo:     commitInfo,
		metaCommitInfo: metaCommitInfo,
		cancel:         cancel,
	}
	// Inputs must be ready before we can construct a datum iterator, so do this
	// synchronously to ensure correct order in the jobChain.
	if err := ppj.logger.LogStep("waiting for pipeline job inputs", func() error {
		return reg.processJobStarting(ppj)
	}); err != nil {
		return err
	}
	// TODO: This could probably be scoped to a callback, and we could move job specific features
	// in the chain package (timeouts for example).
	// TODO: I use the registry pachclient for the iterators, so I can reuse across jobs for skipping.
	pachClient := reg.driver.PachClient()
	dit, err := datum.NewIterator(pachClient, ppj.pji.Input)
	if err != nil {
		return err
	}
	outputDit := datum.NewCommitIterator(pachClient, ppj.metaCommitInfo.Commit)
	ppj.jdit = reg.jobChain.CreateJob(ppj.driver.PachClient().Ctx(), ppj.pji.PipelineJob.ID, dit, outputDit)
	var afterTime time.Duration
	if ppj.pji.JobTimeout != nil {
		startTime, err := types.TimestampFromProto(ppj.pji.Started)
		if err != nil {
			return err
		}
		timeout, err := types.DurationFromProto(ppj.pji.JobTimeout)
		if err != nil {
			return err
		}
		afterTime = time.Until(startTime.Add(timeout))
	}
	asyncEg, jobCtx = errgroup.WithContext(ppj.driver.PachClient().Ctx())
	ppj.driver = reg.driver.WithContext(jobCtx)
	asyncEg.Go(func() error {
		defer ppj.cancel()
		if ppj.pji.JobTimeout != nil {
			ppj.logger.Logf("cancelling job at: %+v", afterTime)
			timer := time.AfterFunc(afterTime, func() {
				reg.killPipelineJob(ppj, "job timed out")
			})
			defer timer.Stop()
		}
		return backoff.RetryUntilCancel(ppj.driver.PachClient().Ctx(), func() error {
			return reg.superviseJob(ppj)
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			ppj.logger.Logf("error in superviseJob: %v, retrying in %+v", err, d)
			return nil
		})
	})
	asyncEg.Go(func() error {
		defer ppj.cancel()
		mutex := &sync.Mutex{}
		mutex.Lock()
		defer mutex.Unlock()
		// This runs the callback asynchronously, but we want to block the errgroup until it completes
		if err := reg.taskQueue.RunTask(ppj.driver.PachClient().Ctx(), func(master *work.Master) {
			defer mutex.Unlock()
			ppj.taskMaster = master
			backoff.RetryUntilCancel(ppj.driver.PachClient().Ctx(), func() error {
				var err error
				for err == nil {
					err = reg.processJob(ppj)
				}
				if errors.Is(err, errutil.ErrBreak) {
					return nil
				}
				return err
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				ppj.logger.Logf("processJob error: %v, retrying in %v", err, d)
				for err != nil {
					if st, ok := err.(errors.StackTracer); ok {
						ppj.logger.Logf("error stack: %+v", st.StackTrace())
					}
					err = errors.Unwrap(err)
				}
				// Get job state, increment restarts, write job state
				ppj.pji, err = ppj.driver.PachClient().InspectPipelineJob(ppj.pji.PipelineJob.ID, false)
				if err != nil {
					return err
				}
				ppj.pji.Restart++
				if err := ppj.writeJobInfo(); err != nil {
					ppj.logger.Logf("error incrementing restart count for job (%s): %v", ppj.pji.PipelineJob.ID, err)
				}
				// Reload the job's commitInfo(s) as they may have changed and clear the state of the commit(s).
				pachClient := reg.driver.PachClient()
				ppj.commitInfo, err = pachClient.PfsAPIClient.InspectCommit(
					pachClient.Ctx(),
					&pfs.InspectCommitRequest{
						Commit:     ppj.commitInfo.Commit,
						BlockState: pfs.CommitState_STARTED,
					})
				if err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				if _, err := pachClient.PfsAPIClient.ClearCommit(
					pachClient.Ctx(),
					&pfs.ClearCommitRequest{
						Commit: ppj.commitInfo.Commit,
					}); err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				ppj.metaCommitInfo, err = pachClient.PfsAPIClient.InspectCommit(
					pachClient.Ctx(),
					&pfs.InspectCommitRequest{
						Commit:     ppj.metaCommitInfo.Commit,
						BlockState: pfs.CommitState_STARTED,
					})
				if err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				_, err = pachClient.PfsAPIClient.ClearCommit(
					pachClient.Ctx(),
					&pfs.ClearCommitRequest{
						Commit: ppj.metaCommitInfo.Commit,
					})
				return grpcutil.ScrubGRPC(err)
			})
			ppj.logger.Logf("master done running processJobs")
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
		defer ppj.jdit.Finish()
		if err := asyncEg.Wait(); err != nil {
			ppj.logger.Logf("fatal job error: %v", err)
		}
	}()
	return nil
}

// superviseJob watches for the output commit closing and cancels the job, or
// deletes it if the output commit is removed.
func (reg *registry) superviseJob(ppj *pendingPipelineJob) error {
	defer ppj.cancel()
	ci, err := ppj.driver.PachClient().PfsAPIClient.InspectCommit(ppj.driver.PachClient().Ctx(),
		&pfs.InspectCommitRequest{
			Commit:     ppj.pji.OutputCommit,
			BlockState: pfs.CommitState_FINISHED,
		})
	if err != nil {
		if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
			// Stop the job and clean up any job state in the registry
			if err := reg.killPipelineJob(ppj, "output commit missing"); err != nil {
				return err
			}
			// Output commit was deleted. Delete job as well
			if err := ppj.driver.NewSQLTx(func(sqlTx *sqlx.Tx) error {
				// Delete the job if no other worker has deleted it yet
				jobPtr := &pps.StoredPipelineJobInfo{}
				if err := ppj.driver.PipelineJobs().ReadWrite(sqlTx).Get(ppj.pji.PipelineJob.ID, jobPtr); err != nil {
					return err
				}
				return ppj.driver.DeletePipelineJob(sqlTx, jobPtr)
			}); err != nil && !col.IsErrNotFound(err) {
				return err
			}
			return nil
		}
		return err
	}
	if strings.Contains(ci.Description, pfs.EmptyStr) {
		return reg.killPipelineJob(ppj, "output commit closed")
	}
	return nil

}

func (reg *registry) processJob(ppj *pendingPipelineJob) error {
	state := ppj.pji.State
	if ppsutil.IsTerminal(state) {
		return errutil.ErrBreak
	}
	switch state {
	case pps.PipelineJobState_JOB_STARTING:
		return errors.New("job should have been moved out of the STARTING state before processJob")
	case pps.PipelineJobState_JOB_RUNNING:
		return ppj.logger.LogStep("processing job running", func() error {
			return reg.processJobRunning(ppj)
		})
	case pps.PipelineJobState_JOB_EGRESSING:
		return ppj.logger.LogStep("processing job egressing", func() error {
			return reg.processJobEgressing(ppj)
		})
	}
	panic(fmt.Sprintf("unrecognized job state: %s", state))
}

func (reg *registry) processJobStarting(ppj *pendingPipelineJob) error {
	// block until job inputs are ready
	failed, err := failedInputs(ppj.driver.PachClient(), ppj.pji)
	if err != nil {
		return err
	}
	if len(failed) > 0 {
		reason := fmt.Sprintf("inputs failed: %s", strings.Join(failed, ", "))
		return reg.failPipelineJob(ppj, reason)
	}
	ppj.pji.State = pps.PipelineJobState_JOB_RUNNING
	return ppj.writeJobInfo()
}

// TODO:
// Need to put some more thought into the context use.
func (reg *registry) processJobRunning(ppj *pendingPipelineJob) error {
	pachClient := ppj.driver.PachClient()
	eg, ctx := errgroup.WithContext(pachClient.Ctx())
	pachClient = pachClient.WithCtx(ctx)
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	// TODO: We need to delete the output for S3Out since we don't have a clear way to track the output in the stats commit (which means datums cannot be skipped with S3Out).
	// If we had a way to map the output added through the S3 gateway back to the datums, and stored this in the appropriate place in the stats commit, then we would be able
	// handle datums the same way we handle normal pipelines.
	if ppj.driver.PipelineInfo().S3Out {
		if err := pachClient.DeleteFile(ppj.commitInfo.Commit, "/"); err != nil {
			return err
		}
	}
	// Generate the deletion operations and count the number of datums for the job.
	var numDatums int64
	if err := ppj.withDeleter(pachClient, func() error {
		return ppj.jdit.Iterate(func(_ *datum.Meta) error {
			numDatums++
			return nil
		})
	}); err != nil {
		return err
	}
	// Set up the datum set spec for the job.
	// When the datum set spec is not set, evenly distribute the datums.
	var setSpec *datum.SetSpec
	if ppj.driver.PipelineInfo().ChunkSpec != nil {
		setSpec = &datum.SetSpec{
			Number:    ppj.driver.PipelineInfo().ChunkSpec.Number,
			SizeBytes: ppj.driver.PipelineInfo().ChunkSpec.SizeBytes,
		}
	}
	if setSpec == nil || (setSpec.Number == 0 && setSpec.SizeBytes == 0) {
		setSpec = &datum.SetSpec{Number: numDatums / int64(reg.concurrency)}
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
			storageRoot := filepath.Join(ppj.driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
			return datum.CreateSets(ppj.jdit, storageRoot, setSpec, func(upload func(client.ModifyFile) error) error {
				subtask, err := createDatumSetSubtask(pachClient, ppj, upload, renewer)
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
			return ppj.logger.LogStep("running and collecting datum set subtasks", func() error {
				return ppj.taskMaster.RunSubtasksChan(
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
								Commit:    ppj.commitInfo.Commit,
								FilesetId: data.OutputFilesetId,
							},
						); err != nil {
							return grpcutil.ScrubGRPC(err)
						}
						if _, err := pachClient.PfsAPIClient.AddFileset(
							pachClient.Ctx(),
							&pfs.AddFilesetRequest{
								Commit:    ppj.metaCommitInfo.Commit,
								FilesetId: data.MetaFilesetId,
							},
						); err != nil {
							return grpcutil.ScrubGRPC(err)
						}
						return datum.MergeStats(stats, data.Stats)
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
	case <-ppj.driver.PachClient().Ctx().Done():
		return ppj.driver.PachClient().Ctx().Err()
	default:
	}
	ppj.saveJobStats(ppj.jdit.Stats())
	ppj.saveJobStats(stats)
	if stats.FailedID != "" {
		return reg.failPipelineJob(ppj, fmt.Sprintf("datum %v failed", stats.FailedID))
	}
	if ppj.pji.Egress != nil {
		ppj.pji.State = pps.PipelineJobState_JOB_EGRESSING
		return ppj.writeJobInfo()
	}
	return reg.succeedPipelineJob(ppj)
}

func createDatumSetSubtask(pachClient *client.APIClient, ppj *pendingPipelineJob, upload func(client.ModifyFile) error, renewer *renew.StringSet) (*work.Task, error) {
	resp, err := pachClient.WithCreateFilesetClient(func(mf client.ModifyFile) error {
		return upload(mf)
	})
	if err != nil {
		return nil, err
	}
	renewer.Add(resp.FilesetId)
	data, err := serializeDatumSet(&DatumSet{
		PipelineJobID: ppj.pji.PipelineJob.ID,
		OutputCommit:  ppj.commitInfo.Commit,
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

func (reg *registry) processJobEgressing(ppj *pendingPipelineJob) error {
	url := ppj.pji.Egress.URL
	if err := ppj.driver.PachClient().GetFileURL(ppj.commitInfo.Commit, "/", url); err != nil {
		return err
	}
	return reg.succeedPipelineJob(ppj)
}

func failedInputs(pachClient *client.APIClient, pipelineJobInfo *pps.PipelineJobInfo) ([]string, error) {
	var failed []string
	blockCommit := func(name string, commit *pfs.Commit) error {
		ci, err := pachClient.PfsAPIClient.InspectCommit(pachClient.Ctx(),
			&pfs.InspectCommitRequest{
				Commit:     commit,
				BlockState: pfs.CommitState_FINISHED,
			})
		if err != nil {
			return errors.Wrapf(err, "error blocking on commit %s@%s",
				commit.Branch.Repo.Name, commit.Branch.Name, commit.ID)
		}
		if strings.Contains(ci.Description, pfs.EmptyStr) {
			failed = append(failed, name)
		}
		return nil
	}
	visitErr := pps.VisitInput(pipelineJobInfo.Input, func(input *pps.Input) error {
		if input.Pfs != nil && input.Pfs.Commit != "" {
			if err := blockCommit(input.Pfs.Name, client.NewCommit(input.Pfs.Repo, input.Pfs.Branch, input.Pfs.Commit)); err != nil {
				return err
			}
		}
		if input.Git != nil && input.Git.Commit != "" {
			if err := blockCommit(input.Git.Name, client.NewCommit(input.Git.Name, input.Git.Branch, input.Git.Commit)); err != nil {
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
