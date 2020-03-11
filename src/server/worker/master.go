package worker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
)

const (
	masterLockPath = "_master_worker_lock"
)

func (a *APIServer) getMasterLogger() *taggedLogger {
	result := &taggedLogger{
		template:  a.logMsgTemplate, // Copy struct
		stderrLog: log.Logger{},
		marshaler: &jsonpb.Marshaler{},
	}
	result.stderrLog.SetOutput(os.Stderr)
	result.stderrLog.SetFlags(log.LstdFlags | log.Llongfile) // Log file/line
	result.template.Master = true
	return result
}

func (a *APIServer) getWorkerLogger() *taggedLogger {
	result := &taggedLogger{
		template:  a.logMsgTemplate, // Copy struct
		stderrLog: log.Logger{},
		marshaler: &jsonpb.Marshaler{},
	}
	result.stderrLog.SetOutput(os.Stderr)
	result.stderrLog.SetFlags(log.LstdFlags | log.Llongfile) // Log file/line
	return result
}

func (a *APIServer) master(masterType string, spawner func(*client.APIClient) error) {
	masterLock := dlock.NewDLock(a.etcdClient, path.Join(a.etcdPrefix, masterLockPath, a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt))
	logger := a.getMasterLogger()
	b := backoff.NewInfiniteBackOff()
	// Setting a high backoff so that when this master fails, the other
	// workers are more likely to become the master.
	// Also, we've observed race conditions where StopPipeline would cause
	// a master to restart before it's deleted.  PPS would then get confused
	// by the restart and create the workers again, because the restart would
	// bring the pipeline state from PAUSED to RUNNING.  By setting a high
	// retry interval, the master would be deleted before it gets a chance
	// to restart.
	b.InitialInterval = 10 * time.Second
	backoff.RetryNotify(func() error {
		// We use pachClient.Ctx here because it contains auth information.
		ctx, cancel := context.WithCancel(a.pachClient.Ctx())
		defer cancel() // make sure that everything this loop might spawn gets cleaned up
		ctx, err := masterLock.Lock(ctx)
		pachClient := a.pachClient.WithCtx(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)
		logger.Logf("Launching %v master process", masterType)
		return spawner(pachClient)
	}, b, func(err error, d time.Duration) error {
		if auth.IsErrNotAuthorized(err) {
			logger.Logf("failing %q due to auth rejection", a.pipelineInfo.Pipeline.Name)
			return ppsutil.FailPipeline(a.pachClient.Ctx(), a.etcdClient, a.pipelines,
				a.pipelineInfo.Pipeline.Name, "worker master could not access output "+
					"repo to watch for new commits")
		}
		logger.Logf("master: error running the %v master process: %v; retrying in %v", masterType, err, d)
		return nil
	})
}

func (a *APIServer) jobSpawner(pachClient *client.APIClient) error {
	logger := a.getMasterLogger()
	// Listen for new commits, and create jobs when they arrive
	commitIter, err := pachClient.SubscribeCommit(a.pipelineInfo.Pipeline.Name, "",
		client.NewCommitProvenance(ppsconsts.SpecRepo, a.pipelineInfo.Pipeline.Name, a.pipelineInfo.SpecCommit.ID),
		"", pfs.CommitState_READY)
	if err != nil {
		return err
	}
	defer commitIter.Close()
	for {
		commitInfo, err := commitIter.Next()
		if err != nil {
			return err
		}
		if commitInfo.Finished != nil {
			continue
		}
		// Inspect the commit and check again if it has been finished (it may have
		// been closed since it was queued, e.g. by StopPipeline or StopJob)
		commitInfo, err = pachClient.InspectCommit(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID)
		if err != nil {
			return err
		}
		// Determine stats commit.
		var statsCommit *pfs.Commit
		if a.pipelineInfo.EnableStats {
			for _, commitRange := range commitInfo.Subvenance {
				if commitRange.Lower.Repo.Name == a.pipelineInfo.Pipeline.Name && commitRange.Upper.Repo.Name == a.pipelineInfo.Pipeline.Name {
					statsCommit = commitRange.Lower
				}
			}
		}
		if commitInfo.Finished != nil {
			// Finish stats commit if it has not been finished.
			if statsCommit != nil {
				if _, err := pachClient.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
					Commit: statsCommit,
					Empty:  true,
				}); err != nil && !pfsserver.IsCommitFinishedErr(err) {
					return err
				}
			}
			// Make sure that the job has been correctly finished as the commit has.
			ji, err := pachClient.InspectJobOutputCommit(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, false)
			if err != nil {
				// If no job was created for the commit, then we are done.
				if strings.Contains(err.Error(), fmt.Sprintf("job with output commit %s not found", commitInfo.Commit.ID)) {
					continue
				}
				return err
			}
			if !ppsutil.IsTerminal(ji.State) {
				if len(commitInfo.Trees) == 0 {
					if err := a.updateJobState(pachClient.Ctx(), ji,
						pps.JobState_JOB_KILLED, "output commit is finished without data, but job state has not been updated"); err != nil {
						return fmt.Errorf("could not kill job with finished output commit: %v", err)
					}
				} else {
					if err := a.updateJobState(pachClient.Ctx(), ji, pps.JobState_JOB_SUCCESS, ""); err != nil {
						return fmt.Errorf("could not mark job with finished output commit as successful: %v", err)
					}
				}
			}
			continue // commit finished after queueing
		}

		// Check if a job was previously created for this commit. If not, make one
		var jobInfo *pps.JobInfo // job for commitInfo (new or old)
		jobInfos, err := pachClient.ListJob("", nil, commitInfo.Commit, -1, true)
		if err != nil {
			return err
		}
		if len(jobInfos) > 1 {
			return fmt.Errorf("multiple jobs found for commit: %s/%s", commitInfo.Commit.Repo.Name, commitInfo.Commit.ID)
		} else if len(jobInfos) < 1 {
			job, err := pachClient.CreateJob(a.pipelineInfo.Pipeline.Name, commitInfo.Commit, statsCommit)
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
			continue
		case jobInfo.PipelineVersion < a.pipelineInfo.Version:
			// kill unfinished jobs from old pipelines (should generally be cleaned
			// up by PPS master, but the PPS master can fail, and if these jobs
			// aren't killed, future jobs will hang indefinitely waiting for their
			// parents to finish)
			if err := a.updateJobState(pachClient.Ctx(), jobInfo,
				pps.JobState_JOB_KILLED, "pipeline has been updated"); err != nil {
				return fmt.Errorf("could not kill stale job: %v", err)
			}
			continue
		case jobInfo.PipelineVersion > a.pipelineInfo.Version:
			return fmt.Errorf("job %s's version (%d) greater than pipeline's "+
				"version (%d), this should automatically resolve when the worker "+
				"is updated", jobInfo.Job.ID, jobInfo.PipelineVersion, a.pipelineInfo.Version)
		}

		// Now that the jobInfo is persisted, wait until all input commits are
		// ready, split the input datums into chunks and merge the results of
		// chunks as they're processed
		if err := a.waitJob(pachClient, jobInfo, logger); err != nil {
			return err
		}
	}
}

func (a *APIServer) spoutSpawner(pachClient *client.APIClient) error {
	ctx := pachClient.Ctx()

	var dir string

	logger, err := a.getTaggedLogger(pachClient, "spout", nil, false)
	if err != nil {
		return fmt.Errorf("getTaggedLogger: %v", err)
	}
	puller := filesync.NewPuller()

	if err := a.unlinkData(nil); err != nil {
		return fmt.Errorf("unlinkData: %v", err)
	}
	// If this is our second time through the loop cleanup the old data.
	if dir != "" {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("os.RemoveAll: %v", err)
		}
	}
	dir, err = a.downloadData(pachClient, logger, nil, puller, &pps.ProcessStats{}, nil)
	if err != nil {
		return err
	}
	if err := a.linkData(nil, dir); err != nil {
		return fmt.Errorf("linkData: %v", err)
	}

	err = a.runService(ctx, logger)
	if err != nil {
		logger.Logf("error from runService: %+v", err)
	}
	return nil
}

func (a *APIServer) serviceSpawner(pachClient *client.APIClient) error {
	ctx := pachClient.Ctx()
	commitIter, err := pachClient.SubscribeCommit(a.pipelineInfo.Pipeline.Name, "",
		client.NewCommitProvenance(ppsconsts.SpecRepo, a.pipelineInfo.Pipeline.Name, a.pipelineInfo.SpecCommit.ID),
		"", pfs.CommitState_READY)
	if err != nil {
		return err
	}
	defer commitIter.Close()
	var serviceCtx context.Context
	var serviceCancel func()
	var dir string
	for {
		commitInfo, err := commitIter.Next()
		if err != nil {
			return err
		}
		if commitInfo.Finished != nil {
			continue
		}

		// Create a job document matching the service's output commit
		jobInput := ppsutil.JobInput(a.pipelineInfo, commitInfo)
		job, err := pachClient.CreateJob(a.pipelineInfo.Pipeline.Name, commitInfo.Commit, nil)
		if err != nil {
			return err
		}
		df, err := NewDatumIterator(pachClient, jobInput)
		if err != nil {
			return err
		}
		if df.Len() != 1 {
			return fmt.Errorf("services must have a single datum")
		}
		data := df.DatumN(0)
		logger, err := a.getTaggedLogger(pachClient, job.ID, data, false)
		if err != nil {
			return fmt.Errorf("getTaggedLogger: %v", err)
		}
		puller := filesync.NewPuller()
		// If this is our second time through the loop cleanup the old data.
		if dir != "" {
			if err := a.unlinkData(data); err != nil {
				return fmt.Errorf("unlinkData: %v", err)
			}
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("os.RemoveAll: %v", err)
			}
		}
		dir, err = a.downloadData(pachClient, logger, data, puller, &pps.ProcessStats{}, nil)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(client.PPSInputPrefix, 0666); err != nil {
			return err
		}
		if err := a.linkData(data, dir); err != nil {
			return fmt.Errorf("linkData: %v", err)
		}
		if serviceCancel != nil {
			serviceCancel()
		}
		serviceCtx, serviceCancel = context.WithCancel(ctx)
		defer serviceCancel() // make go vet happy: infinite loop obviates 'defer'
		go func() {
			serviceCtx := serviceCtx
			if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
				jobs := a.jobs.ReadWrite(stm)
				jobPtr := &pps.EtcdJobInfo{}
				if err := jobs.Get(job.ID, jobPtr); err != nil {
					return err
				}
				return ppsutil.UpdateJobState(a.pipelines.ReadWrite(stm), a.jobs.ReadWrite(stm), jobPtr, pps.JobState_JOB_RUNNING, "")
			}); err != nil {
				logger.Logf("error updating job state: %+v", err)
			}
			err := a.runService(serviceCtx, logger)
			if err != nil {
				logger.Logf("error from runService: %+v", err)
			}
			select {
			case <-serviceCtx.Done():
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					jobs := a.jobs.ReadWrite(stm)
					jobPtr := &pps.EtcdJobInfo{}
					if err := jobs.Get(job.ID, jobPtr); err != nil {
						return err
					}
					return ppsutil.UpdateJobState(a.pipelines.ReadWrite(stm), a.jobs.ReadWrite(stm), jobPtr, pps.JobState_JOB_SUCCESS, "")
				}); err != nil {
					logger.Logf("error updating job progress: %+v", err)
				}
				if err := pachClient.FinishCommit(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID); err != nil {
					logger.Logf("could not finish output commit: %v", err)
				}
			default:
			}
		}()
	}
}

func plusDuration(x *types.Duration, y *types.Duration) (*types.Duration, error) {
	var xd time.Duration
	var yd time.Duration
	var err error
	if x != nil {
		xd, err = types.DurationFromProto(x)
		if err != nil {
			return nil, err
		}
	}
	if y != nil {
		yd, err = types.DurationFromProto(y)
		if err != nil {
			return nil, err
		}
	}
	return types.DurationProto(xd + yd), nil
}

func (a *APIServer) chunks(jobID string) col.Collection {
	return col.NewCollection(a.etcdClient, path.Join(a.etcdPrefix, chunkPrefix, jobID), nil, &ChunkState{}, nil, nil)
}

func (a *APIServer) merges(jobID string) col.Collection {
	return col.NewCollection(a.etcdClient, path.Join(a.etcdPrefix, mergePrefix, jobID), nil, &MergeState{}, nil, nil)
}

func newPlan(df DatumIterator, spec *pps.ChunkSpec, parallelism int, numHashtrees int64) *Plan {
	if spec == nil {
		spec = &pps.ChunkSpec{}
	}
	if spec.Number == 0 && spec.SizeBytes == 0 {
		spec.Number = int64(df.Len() / (parallelism * 10))
		if spec.Number == 0 {
			spec.Number = 1
		}
	}
	plan := &Plan{}
	if spec.Number != 0 {
		for i := spec.Number; i < int64(df.Len()); i += spec.Number {
			plan.Chunks = append(plan.Chunks, int64(i))
		}
	} else {
		size := int64(0)
		for i := 0; i < df.Len(); i++ {
			for _, input := range df.DatumN(i) {
				size += int64(input.FileInfo.SizeBytes)
			}
			if size > spec.SizeBytes {
				plan.Chunks = append(plan.Chunks, int64(i))
			}
		}
	}
	plan.Chunks = append(plan.Chunks, int64(df.Len()))
	plan.Merges = numHashtrees
	return plan
}

func (a *APIServer) failedInputs(ctx context.Context, jobInfo *pps.JobInfo) ([]string, error) {
	var failedInputs []string
	var vistErr error
	blockCommit := func(name string, commit *pfs.Commit) {
		ci, err := a.pachClient.PfsAPIClient.InspectCommit(ctx,
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
			failedInputs = append(failedInputs, name)
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
	return failedInputs, vistErr
}

// waitJob waits for the job in 'jobInfo' to finish, and then it collects the
// output from the job's workers and merges it into a commit (and may merge
// stats into a commit in the stats branch as well)
func (a *APIServer) waitJob(pachClient *client.APIClient, jobInfo *pps.JobInfo, logger *taggedLogger) (retErr error) {
	logger.Logf("waiting on job %q (pipeline version: %d, state: %s)", jobInfo.Job.ID, jobInfo.PipelineVersion, jobInfo.State)
	ctx, cancel := context.WithCancel(pachClient.Ctx())
	pachClient = pachClient.WithCtx(ctx)

	// Watch the output commit to see if it's terminated (KILLED, FAILED, or
	// SUCCESS) and if so, cancel the current context
	go func() {
		backoff.RetryNotify(func() error {
			commitInfo, err := pachClient.PfsAPIClient.InspectCommit(ctx,
				&pfs.InspectCommitRequest{
					Commit:     jobInfo.OutputCommit,
					BlockState: pfs.CommitState_FINISHED,
				})
			if err != nil {
				if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
					defer cancel() // whether we return error or nil, job is done
					// Output commit was deleted. Delete job as well
					if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
						// Delete the job if no other worker has deleted it yet
						jobPtr := &pps.EtcdJobInfo{}
						if err := a.jobs.ReadWrite(stm).Get(jobInfo.Job.ID, jobPtr); err != nil {
							return err
						}
						return a.deleteJob(stm, jobPtr)
					}); err != nil && !col.IsErrNotFound(err) {
						return err
					}
					return nil
				}
				return err
			}
			if commitInfo.Trees == nil {
				defer cancel() // whether job state update succeeds or not, job is done
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					// Read an up to date version of the jobInfo so that we
					// don't overwrite changes that have happened since this
					// function started.
					jobPtr := &pps.EtcdJobInfo{}
					if err := a.jobs.ReadWrite(stm).Get(jobInfo.Job.ID, jobPtr); err != nil {
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
					if !ppsutil.IsTerminal(jobPtr.State) {
						return ppsutil.UpdateJobState(a.pipelines.ReadWrite(stm), a.jobs.ReadWrite(stm), jobPtr, pps.JobState_JOB_KILLED, "")
					}
					return nil
				}); err != nil {
					return err
				}
			}
			return nil
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			if isDone(ctx) {
				return err // exit retry loop
			}
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
		failedInputs, err := a.failedInputs(ctx, jobInfo)
		if err != nil {
			return err
		}
		if len(failedInputs) > 0 {
			reason := fmt.Sprintf("inputs %s failed", strings.Join(failedInputs, ", "))
			if jobInfo.EnableStats {
				if _, err = pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
					Commit: jobInfo.StatsCommit,
					Empty:  true,
				}); err != nil && !pfsserver.IsCommitFinishedErr(err) {
					return err
				}
			}
			if err := a.updateJobState(ctx, jobInfo, pps.JobState_JOB_FAILURE, reason); err != nil {
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
		df, err := NewDatumIterator(pachClient, jobInfo.Input)
		if err != nil {
			return err
		}
		pipelinePtr := &pps.EtcdPipelineInfo{}
		if err := a.pipelines.ReadOnly(ctx).Get(a.pipelineInfo.Pipeline.Name, pipelinePtr); err != nil {
			return err
		}
		parallelism := int(pipelinePtr.Parallelism)
		numHashtrees, err := ppsutil.GetExpectedNumHashtrees(a.pipelineInfo.HashtreeSpec)
		if err != nil {
			return fmt.Errorf("error from GetExpectedNumHashtrees: %v", err)
		}
		plan := &Plan{}
		// Read the job document, and either resume (if we're recovering from a
		// crash) or mark it running. Also write the input chunks calculated above
		// into plansCol
		jobID := jobInfo.Job.ID
		if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobPtr := &pps.EtcdJobInfo{}
			if err := jobs.Get(jobID, jobPtr); err != nil {
				return err
			}
			if jobPtr.State == pps.JobState_JOB_KILLED {
				return nil
			}
			jobPtr.DataTotal = int64(df.Len())
			if err := ppsutil.UpdateJobState(a.pipelines.ReadWrite(stm), a.jobs.ReadWrite(stm), jobPtr, pps.JobState_JOB_RUNNING, ""); err != nil {
				return err
			}
			plansCol := a.plans.ReadWrite(stm)
			if err := plansCol.Get(jobID, plan); err == nil {
				return nil
			}
			plan = newPlan(df, jobInfo.ChunkSpec, parallelism, numHashtrees)
			return plansCol.Put(jobID, plan)
		}); err != nil {
			return err
		}
		defer func() {
			if retErr == nil {
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					chunksCol := a.chunks(jobID).ReadWrite(stm)
					chunksCol.DeleteAll()
					plansCol := a.plans.ReadWrite(stm)
					return plansCol.Delete(jobID)
				}); err != nil {
					retErr = err
				}
			}
		}()
		// Watch the chunks in order
		chunks := a.chunks(jobInfo.Job.ID).ReadOnly(ctx)
		var failedDatumID string
		recoveredDatums := make(map[string]uint64)
		for _, high := range plan.Chunks {
			chunkState := &ChunkState{}
			if err := chunks.WatchOneF(fmt.Sprint(high), func(e *watch.Event) error {
				var key string
				if err := e.Unmarshal(&key, chunkState); err != nil {
					return err
				}
				if key != fmt.Sprint(high) {
					return nil
				}
				if chunkState.State != State_RUNNING {
					if chunkState.State == State_FAILED {
						failedDatumID = chunkState.DatumID
					} else if chunkState.State == State_COMPLETE {
						// if the chunk has been completed, grab the recovered datums from the chunk
						chunkRecoveredDatums, err := a.getDatumMap(ctx, pachClient, chunkState.RecoveredDatums)
						if err != nil {
							return err
						}
						for k, count := range chunkRecoveredDatums {
							recoveredDatums[k] += count
						}
					}
					return errutil.ErrBreak
				}
				return nil
			}); err != nil {
				return err
			}
		}
		if err := a.updateJobState(ctx, jobInfo, pps.JobState_JOB_MERGING, ""); err != nil {
			return err
		}
		var trees []*pfs.Object
		var size uint64
		var statsTrees []*pfs.Object
		var statsSize uint64
		if failedDatumID == "" || jobInfo.EnableStats {
			// Wait for all merges to happen.
			merges := a.merges(jobInfo.Job.ID).ReadOnly(ctx)
			for merge := int64(0); merge < plan.Merges; merge++ {
				mergeState := &MergeState{}
				if err := merges.WatchOneF(fmt.Sprint(merge), func(e *watch.Event) error {
					var key string
					if err := e.Unmarshal(&key, mergeState); err != nil {
						return err
					}
					if key != fmt.Sprint(merge) {
						return nil
					}
					if mergeState.State != State_RUNNING {
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
			if err := a.updateJobState(ctx, jobInfo, pps.JobState_JOB_FAILURE, reason); err != nil {
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
		for i := 0; i < df.Len(); i++ {
			files := df.DatumN(i)
			datumHash := HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, files)
			// recovered datums were not processed, and thus should not be skipped
			if count := recoveredDatums[a.DatumID(files)]; count > 0 {
				recoveredDatums[a.DatumID(files)]--
				continue // so we won't write them to the processed datums object
			}
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
		if err := a.egress(pachClient, logger, jobInfo); err != nil {
			reason := fmt.Sprintf("egress error: %v", err)
			return a.updateJobState(ctx, jobInfo, pps.JobState_JOB_FAILURE, reason)
		}
		return a.updateJobState(ctx, jobInfo, pps.JobState_JOB_SUCCESS, "")
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("error in waitJob %v, retrying in %v", err, d)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		// Increment the job's restart count
		_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobID := jobInfo.Job.ID
			jobPtr := &pps.EtcdJobInfo{}
			if err := jobs.Get(jobID, jobPtr); err != nil {
				return err
			}
			jobPtr.Restart++
			return jobs.Put(jobID, jobPtr)
		})
		if err != nil {
			logger.Logf("error incrementing job %s's restart count", jobInfo.Job.ID)
		}
		return nil
	})
	return nil
}

func (a *APIServer) updateJobState(ctx context.Context, info *pps.JobInfo, state pps.JobState, reason string) error {
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		jobs := a.jobs.ReadWrite(stm)
		jobID := info.Job.ID
		jobPtr := &pps.EtcdJobInfo{}
		if err := jobs.Get(jobID, jobPtr); err != nil {
			return err
		}
		return ppsutil.UpdateJobState(a.pipelines.ReadWrite(stm), a.jobs.ReadWrite(stm), jobPtr, state, reason)
	})
	return err
}

// deleteJob is identical to updateJobState, except that jobPtr points to a job
// that should be deleted rather than marked failed. Jobs may be deleted if
// their output commit is deleted.
func (a *APIServer) deleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {
	pipelinePtr := &pps.EtcdPipelineInfo{}
	if err := a.pipelines.ReadWrite(stm).Update(jobPtr.Pipeline.Name, pipelinePtr, func() error {
		if pipelinePtr.JobCounts == nil {
			pipelinePtr.JobCounts = make(map[int32]int32)
		}
		if pipelinePtr.JobCounts[int32(jobPtr.State)] != 0 {
			pipelinePtr.JobCounts[int32(jobPtr.State)]--
		}
		return nil
	}); err != nil {
		return err
	}
	return a.jobs.ReadWrite(stm).Delete(jobPtr.Job.ID)
}

func (a *APIServer) egress(pachClient *client.APIClient, logger *taggedLogger, jobInfo *pps.JobInfo) error {
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
			if err := filesync.PushObj(pachClient, jobInfo.OutputCommit, objClient, url.Object); err != nil {
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

func (a *APIServer) receiveSpout(ctx context.Context, logger *taggedLogger) error {
	return backoff.RetryNotify(func() error {
		repo := a.pipelineInfo.Pipeline.Name
		for {
			// this extra closure is so that we can scope the defer
			if err := func() (retErr error) {
				// open a read connection to the /pfs/out named pipe
				out, err := os.Open("/pfs/out")
				if err != nil {
					return err
				}
				// and close it at the end of each loop
				defer func() {
					if err := out.Close(); err != nil && retErr == nil {
						// this lets us pass the error through if Close fails
						retErr = err
					}
				}()
				outTar := tar.NewReader(out)

				// start commit
				commit, err := a.pachClient.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
					Parent:     client.NewCommit(repo, ""),
					Branch:     a.pipelineInfo.OutputBranch,
					Provenance: []*pfs.CommitProvenance{client.NewCommitProvenance(ppsconsts.SpecRepo, repo, a.pipelineInfo.SpecCommit.ID)},
				})
				if err != nil {
					return err
				}

				// finish the commit even if there was an issue
				defer func() {
					if err := a.pachClient.FinishCommit(repo, commit.ID); err != nil && retErr == nil {
						// this lets us pass the error through if FinishCommit fails
						retErr = err
					}
				}()
				// this loops through all the files in the tar that we've read from /pfs/out
				for {
					fileHeader, err := outTar.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						return err
					}
					// put files into pachyderm
					if a.pipelineInfo.Spout.Marker != "" && strings.HasPrefix(path.Clean(fileHeader.Name), a.pipelineInfo.Spout.Marker) {
						// we'll check that this is the latest version of the spout, and then commit to it
						// we need to do this atomically because we otherwise might hit a subtle race condition

						// check to see if this spout is the latest version of this spout by seeing if its spec commit has any children
						spec, err := a.pachClient.InspectCommit(ppsconsts.SpecRepo, a.pipelineInfo.SpecCommit.ID)
						if err != nil && !errutil.IsNotFoundError(err) {
							return err
						}
						if spec != nil && len(spec.ChildCommits) != 0 {
							return fmt.Errorf("outdated spout, now shutting down")
						}
						_, err = a.pachClient.PutFileOverwrite(repo, ppsconsts.SpoutMarkerBranch, fileHeader.Name, outTar, 0)
						if err != nil {
							return err
						}

					} else if a.pipelineInfo.Spout.Overwrite {
						_, err = a.pachClient.PutFileOverwrite(repo, commit.ID, fileHeader.Name, outTar, 0)
						if err != nil {
							return err
						}
					} else {
						_, err = a.pachClient.PutFile(repo, commit.ID, fileHeader.Name, outTar)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			return err
		default:
			logger.Logf("error running spout: %+v, retrying in: %+v", err, d)
			return nil
		}
	})
}

func (a *APIServer) runService(ctx context.Context, logger *taggedLogger) error {
	return backoff.RetryNotify(func() error {
		// if we have a spout, then asynchronously receive spout data
		if a.pipelineInfo.Spout != nil {
			go a.receiveSpout(ctx, logger)
		}
		return a.runUserCode(ctx, logger, nil, &pps.ProcessStats{}, nil)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			return err
		default:
			logger.Logf("error running user code: %+v, retrying in: %+v", err, d)
			return nil
		}
	})
}
