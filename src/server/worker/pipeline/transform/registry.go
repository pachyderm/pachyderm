package transform

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/limit"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
)

const (
	defaultDatumSetsPerWorker int64 = 4
)

type registry struct {
	driver  driver.Driver
	logger  logs.TaggedLogger
	limiter limit.ConcurrencyLimiter
}

// NewRegistry is not properly documented.
//
// TODO: document.
//
// TODO:
// Prometheus stats? (previously in the driver, which included testing we should reuse if possible)
// capture logs (reuse driver tests and reintroduce tagged logger).
func NewRegistry(driver driver.Driver, logger logs.TaggedLogger) (*registry, error) {
	// Determine the maximum number of concurrent tasks we will allow
	concurrency, err := driver.ExpectedNumWorkers()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &registry{
		driver:  driver,
		logger:  logger,
		limiter: limit.New(int(concurrency)),
	}, nil
}

func (reg *registry) succeedJob(pj *pendingJob) error {
	pj.logger.Logf("job successful, closing commits")
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	if err := ppsutil.FinishJob(reg.driver.PachClient(), pj.ji, pps.JobState_JOB_FINISHING, ""); err != nil {
		return err
	}
	pj.clearCache()
	return nil
}

func (reg *registry) failJob(pj *pendingJob, reason string) error {
	pj.logger.Logf("failing job with reason: %s", reason)
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	if err := ppsutil.FinishJob(reg.driver.PachClient(), pj.ji, pps.JobState_JOB_FAILURE, reason); err != nil {
		return err
	}
	pj.clearCache()
	return nil
}

func (reg *registry) markJobUnrunnable(pj *pendingJob, reason string) error {
	pj.logger.Logf("marking job unrunnable with reason: %s", reason)
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	if err := ppsutil.FinishJob(reg.driver.PachClient(), pj.ji, pps.JobState_JOB_UNRUNNABLE, reason); err != nil {
		return err
	}
	pj.clearCache()
	return nil
}

func (reg *registry) killJob(pj *pendingJob, reason string) error {
	pj.logger.Logf("killing job with reason: %s", reason)
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	_, err := reg.driver.PachClient().PpsAPIClient.StopJob(
		reg.driver.PachClient().Ctx(),
		&pps.StopJobRequest{
			Job:    pj.ji.Job,
			Reason: reason,
		},
	)
	return errors.EnsureStack(err)
}

func (reg *registry) StartJob(jobInfo *pps.JobInfo) (retErr error) {
	reg.limiter.Acquire()
	defer func() {
		// TODO(2.0 optional): The error handling during job setup needs more work.
		// For a commit that is squashed, we would want to exit.
		// For transient errors, we would want to retry, not just give up on the job.
		if retErr != nil {
			reg.limiter.Release()
		}
	}()
	pi := reg.driver.PipelineInfo()
	pj := &pendingJob{
		driver: reg.driver,
		logger: reg.logger.WithJob(jobInfo.Job.Id),
		ji:     jobInfo,
		noSkip: pi.Details.ReprocessSpec == client.ReprocessSpecEveryJob || pi.Details.S3Out,
		cache:  newCache(reg.driver.PachClient(), ppsdb.JobKey(jobInfo.Job)),
	}
	if pj.ji.State == pps.JobState_JOB_CREATED {
		pj.ji.State = pps.JobState_JOB_STARTING
		if err := pj.writeJobInfo(); err != nil {
			return err
		}
	}
	// Inputs must be ready before we can construct a datum iterator.
	if err := pj.logger.LogStep("waiting for job inputs", func() error {
		return reg.processJobStarting(pj)
	}); err != nil {
		return errors.EnsureStack(err)
	}
	if err := pj.load(); err != nil {
		return err
	}
	// TODO: This could probably be scoped to a callback.
	var afterTime time.Duration
	if pj.ji.Details.JobTimeout != nil {
		startTime := pj.ji.Started.AsTime()
		timeout := pj.ji.Details.JobTimeout.AsDuration()
		afterTime = time.Until(startTime.Add(timeout))
	}
	go func() {
		defer reg.limiter.Release()
		if pj.ji.Details.JobTimeout != nil {
			pj.logger.Logf("cancelling job at: %+v", afterTime)
			timer := time.AfterFunc(afterTime, func() {
				if err := reg.killJob(pj, " jobtimed out"); err != nil {
					pj.logger.Logf("error killing job: %v", err)
				}
			})
			defer timer.Stop()
		}
		if err := backoff.RetryUntilCancel(reg.driver.PachClient().Ctx(), func() error {
			ctx, cancel := pctx.WithCancel(reg.driver.PachClient().Ctx())
			defer cancel()
			eg, jobCtx := errgroup.WithContext(ctx)
			pj.driver = reg.driver.WithContext(jobCtx)
			pj.cancel = cancel
			eg.Go(func() error {
				return reg.superviseJob(pj)
			})
			eg.Go(func() error {
				var err error
				for err == nil {
					err = reg.processJob(pj)
				}
				if errors.Is(err, errutil.ErrBreak) || errors.Is(context.Cause(ctx), context.Canceled) {
					return nil
				}
				return err
			})
			return errors.EnsureStack(eg.Wait())
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			pj.logger.Logf("error processing job: %v, retrying in %v", err, d)
			for err != nil {
				if st, ok := err.(errors.StackTracer); ok {
					pj.logger.Logf("error stack: %+v", st.StackTrace())
				}
				err = errors.Unwrap(err)
			}
			pj.driver = reg.driver
			return backoff.RetryUntilCancel(reg.driver.PachClient().Ctx(), func() error {
				// Reload the job's commits and info as they may have changed.
				if err := pj.load(); err != nil {
					return err
				}
				pj.ji.Restart++
				return pj.writeJobInfo()
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
					return err
				}
				pj.logger.Logf("error restarting job: %v, retrying in %v", err, d)
				return nil
			})
		}); err != nil {
			pj.logger.Logf("fatal job error: %v", err)
		}
	}()
	return nil
}

func (reg *registry) superviseJob(pj *pendingJob) error {
	defer pj.cancel()
	// wrap InspectCommit() in a retry to mitigate database connection flakiness.
	if err := backoff.RetryUntilCancel(pj.driver.PachClient().Ctx(), func() error {
		_, err := pj.driver.PachClient().PfsAPIClient.InspectCommit(pj.driver.PachClient().Ctx(),
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
				// TODO: This should be handled through a transaction defer when the commit is squashed.
				if err := pj.driver.NewSQLTx(func(ctx context.Context, sqlTx *pachsql.Tx) error {
					// Delete the job if no other worker has deleted it yet
					jobInfo := &pps.JobInfo{}
					if err := pj.driver.Jobs().ReadWrite(sqlTx).Get(ctx, ppsdb.JobKey(pj.ji.Job), jobInfo); err != nil {
						return errors.EnsureStack(err)
					}
					return errors.EnsureStack(pj.driver.DeleteJob(ctx, sqlTx, jobInfo))
				}); err != nil && !col.IsErrNotFound(err) {
					return errors.EnsureStack(err)
				}
				return nil
			}
			if errutil.IsDatabaseDisconnect(err) {
				pj.logger.Logf("retry InspectCommit() in registry.superviseJob(); err %v", err)
				return backoff.ErrContinue
			}
			return errors.EnsureStack(err)
		}
		return nil
	}, backoff.RetryEvery(time.Second), nil); err != nil {
		return err
	}
	return nil

}

func (reg *registry) processJob(pj *pendingJob) error {
	state := pj.ji.State
	if pps.IsTerminal(state) || state == pps.JobState_JOB_FINISHING {
		return errutil.ErrBreak
	}
	switch state {
	case pps.JobState_JOB_STARTING:
		return errors.New("job should have been moved out of the STARTING state before processJob")
	case pps.JobState_JOB_RUNNING:
		return errors.EnsureStack(pj.logger.LogStep("processing job running", func() error {
			return reg.processJobRunning(pj)
		}))
	case pps.JobState_JOB_EGRESSING:
		return errors.EnsureStack(pj.logger.LogStep("processing job egressing", func() error {
			return reg.processJobEgressing(pj)
		}))
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
		reason := fmt.Sprintf("unrunnable because the following upstream pipelines failed: %s", strings.Join(failed, ", "))
		return reg.markJobUnrunnable(pj, reason)
	}
	pj.ji.State = pps.JobState_JOB_RUNNING
	return pj.writeJobInfo()
}

func (reg *registry) processJobRunning(pj *pendingJob) error {
	pachClient := pj.driver.PachClient()
	// TODO: We need to delete the output for S3Out since we don't have a clear way to track the output in the stats commit (which means datums cannot be skipped with S3Out).
	// If we had a way to map the output added through the S3 gateway back to the datums, and stored this in the appropriate place in the stats commit, then we would be able
	// handle datums the same way we handle normal pipelines.
	if pj.driver.PipelineInfo().Details.S3Out {
		if err := pachClient.DeleteFile(pj.commitInfo.Commit, "/"); err != nil {
			return err
		}
	}
	preprocessingTaskDoer := reg.driver.NewPreprocessingTaskDoer(pj.ji.Job.Id, pj.cache)
	processingTaskDoer := reg.driver.NewProcessingTaskDoer(pj.ji.Job.Id, pj.cache)
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		fileSetID, err := pj.createParallelDatums(ctx, preprocessingTaskDoer)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, fileSetID); err != nil {
			return err
		}
		pachClient := pachClient.WithCtx(ctx)
		if err := reg.processDatums(pachClient, pj, preprocessingTaskDoer, processingTaskDoer, fileSetID); err != nil {
			return err
		}
		fileSetID, err = pj.createSerialDatums(ctx, preprocessingTaskDoer)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, fileSetID); err != nil {
			return err
		}
		return reg.processDatums(pachClient, pj, preprocessingTaskDoer, processingTaskDoer, fileSetID)
	}); err != nil {
		if errors.Is(err, errutil.ErrBreak) {
			return nil
		}
		return err
	}
	if pj.ji.Details.Egress != nil {
		pj.ji.State = pps.JobState_JOB_EGRESSING
		return pj.writeJobInfo()
	}
	return reg.succeedJob(pj)
}

func (reg *registry) processDatums(pachClient *client.APIClient, pj *pendingJob, preprocessingTaskDoer, processingTaskDoer task.Doer, fileSetID string) error {
	datumSets, err := createDatumSets(pachClient, pj, preprocessingTaskDoer, fileSetID)
	if err != nil {
		return err
	}
	var inputs []*anypb.Any
	for _, datumSet := range datumSets {
		input, err := serializeDatumSetTask(&DatumSetTask{
			Job:          pj.ji.Job,
			FileSetId:    fileSetID,
			PathRange:    datumSet,
			OutputCommit: pj.commitInfo.Commit,
		})
		if err != nil {
			return err
		}
		inputs = append(inputs, input)
	}
	ctx := pachClient.Ctx()
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	if err := task.DoBatch(ctx, processingTaskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
		if err != nil {
			return err
		}
		result, err := deserializeDatumSetTaskResult(output)
		if err != nil {
			return err
		}
		if _, err := pachClient.PfsAPIClient.AddFileSet(
			ctx,
			&pfs.AddFileSetRequest{
				Commit:    pj.commitInfo.Commit,
				FileSetId: result.OutputFileSetId,
			},
		); err != nil {
			return errors.EnsureStack(err)
		}
		if _, err := pachClient.PfsAPIClient.AddFileSet(
			ctx,
			&pfs.AddFileSetRequest{
				Commit:    pj.metaCommitInfo.Commit,
				FileSetId: result.MetaFileSetId,
			},
		); err != nil {
			return errors.EnsureStack(err)
		}
		datum.MergeStats(stats, result.Stats)
		pj.saveJobStats(result.Stats)
		return pj.writeJobInfo()
	}); err != nil {
		return err
	}
	if stats.FailedId != "" {
		if err := reg.failJob(pj, fmt.Sprintf("datum %v failed", stats.FailedId)); err != nil {
			return err
		}
		return errutil.ErrBreak
	}
	return nil
}

func createDatumSets(pachClient *client.APIClient, pj *pendingJob, taskDoer task.Doer, fileSetID string) ([]*pfs.PathRange, error) {
	setSpec, err := createSetSpec(pj)
	if err != nil {
		return nil, err
	}
	var datumSets []*pfs.PathRange
	if err := pj.logger.LogStep("creating datum sets", func() error {
		shards, err := pachClient.ShardFileSet(fileSetID)
		if err != nil {
			return err
		}
		var inputs []*anypb.Any
		for _, shard := range shards {
			input, err := serializeCreateDatumSetsTask(&CreateDatumSetsTask{
				FileSetId: fileSetID,
				PathRange: shard,
				SetSpec:   setSpec,
				AuthToken: pachClient.AuthToken(),
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		return task.DoBatch(pachClient.Ctx(), taskDoer, inputs, func(_ int64, output *anypb.Any, err error) error {
			if err != nil {
				return err
			}
			result, err := deserializeCreateDatumSetsTaskResult(output)
			if err != nil {
				return err
			}
			datumSets = append(datumSets, result.DatumSets...)
			return nil
		})
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return datumSets, nil
}

func createSetSpec(pj *pendingJob) (*datum.SetSpec, error) {
	var setSpec *datum.SetSpec
	datumSetsPerWorker := defaultDatumSetsPerWorker
	if pj.driver.PipelineInfo().Details.DatumSetSpec != nil {
		setSpec = &datum.SetSpec{
			Number:    pj.driver.PipelineInfo().Details.DatumSetSpec.Number,
			SizeBytes: pj.driver.PipelineInfo().Details.DatumSetSpec.SizeBytes,
		}
		datumSetsPerWorker = pj.driver.PipelineInfo().Details.DatumSetSpec.PerWorker
	}
	// When the datum set spec is not set, evenly distribute the datums.
	if setSpec == nil || (setSpec.Number == 0 && setSpec.SizeBytes == 0) {
		concurrency, err := pj.driver.ExpectedNumWorkers()
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		numDatums := pj.ji.DataTotal - pj.ji.DataSkipped
		setSpec = &datum.SetSpec{Number: numDatums / (int64(concurrency) * datumSetsPerWorker)}
		if setSpec.Number == 0 {
			setSpec.Number = 1
		}
	}
	return setSpec, nil
}

func serializeCreateDatumSetsTask(task *CreateDatumSetsTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeCreateDatumSetsTaskResult(taskAny *anypb.Any) (*CreateDatumSetsTaskResult, error) {
	task := &CreateDatumSetsTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeDatumSetTask(task *DatumSetTask) (*anypb.Any, error) { return anypb.New(task) }

func deserializeDatumSetTaskResult(taskAny *anypb.Any) (*DatumSetTaskResult, error) {
	task := &DatumSetTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func (reg *registry) processJobEgressing(pj *pendingJob) error {
	client := pj.driver.PachClient()
	egress := pj.ji.Details.Egress
	var request pfs.EgressRequest
	request.Commit = pj.commitInfo.Commit
	// For backwards compatibility, egress still works with just a URL to object storage.
	if egress.URL != "" {
		request.Target = &pfs.EgressRequest_ObjectStorage{ObjectStorage: &pfs.ObjectStorageEgress{Url: egress.URL}}
	} else {
		switch egress.Target.(type) {
		case *pps.Egress_ObjectStorage:
			request.Target = &pfs.EgressRequest_ObjectStorage{ObjectStorage: egress.GetObjectStorage()}
		case *pps.Egress_SqlDatabase:
			request.Target = &pfs.EgressRequest_SqlDatabase{SqlDatabase: egress.GetSqlDatabase()}
		}
	}
	if err := backoff.RetryUntilCancel(client.Ctx(), func() error {
		// file not found means the commit is empty, nothing to egress
		// TODO explicitly handle/swallow more unrecoverable errors from SqlDatabase, to prevent job being stuck in egressing.
		if _, err := client.Egress(client.Ctx(), &request); err != nil && !pfsserver.IsFileNotFoundErr(err) {
			return err
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		pj.logger.Logf("error processing egress: %v, retrying in %v", err, d)
		return nil
	}); err != nil {
		return err
	}
	return reg.succeedJob(pj)
}

func failedInputs(pachClient *client.APIClient, jobInfo *pps.JobInfo) ([]string, error) {
	var failed []string
	waitCommit := func(name string, commit *pfs.Commit) error {
		ci, err := pachClient.WaitCommit(commit.Repo.Project.GetName(), commit.Repo.Name, commit.Branch.Name, commit.Id)
		if err != nil {
			return errors.Wrapf(err, "error blocking on commit %s", commit)
		}
		if ci.Error != "" {
			failed = append(failed, name)
		}
		return nil
	}
	visitErr := pps.VisitInput(jobInfo.Details.Input, func(input *pps.Input) error {
		if input.Pfs != nil && input.Pfs.Commit != "" {
			if err := waitCommit(input.Pfs.Name, client.NewCommit(input.Pfs.Project, input.Pfs.Repo, input.Pfs.Branch, input.Pfs.Commit)); err != nil {
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
