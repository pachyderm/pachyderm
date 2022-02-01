package transform

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/client/limit"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

const (
	defaultDatumSetsPerWorker int64 = 4
)

type hasher struct {
	salt string
}

func (h *hasher) Hash(inputs []*common.Input) string {
	return common.HashDatum(h.salt, inputs)
}

type registry struct {
	driver      driver.Driver
	logger      logs.TaggedLogger
	concurrency int64
	limiter     limit.ConcurrencyLimiter
}

// TODO:
// Prometheus stats? (previously in the driver, which included testing we should reuse if possible)
// capture logs (reuse driver tests and reintroduce tagged logger).
func newRegistry(driver driver.Driver, logger logs.TaggedLogger) (*registry, error) {
	// Determine the maximum number of concurrent tasks we will allow
	concurrency, err := driver.ExpectedNumWorkers()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &registry{
		driver:      driver,
		logger:      logger,
		concurrency: concurrency,
		limiter:     limit.New(int(concurrency)),
	}, nil
}

func (reg *registry) succeedJob(pj *pendingJob) error {
	pj.logger.Logf("job successful, closing commits")
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	return ppsutil.FinishJob(reg.driver.PachClient(), pj.ji, pps.JobState_JOB_FINISHING, "")
}

func (reg *registry) failJob(pj *pendingJob, reason string) error {
	pj.logger.Logf("failing job with reason: %s", reason)
	// Use the registry's driver so that the job's supervision goroutine cannot cancel us
	return ppsutil.FinishJob(reg.driver.PachClient(), pj.ji, pps.JobState_JOB_FAILURE, reason)
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

func (reg *registry) startJob(jobInfo *pps.JobInfo) (retErr error) {
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
		logger: reg.logger.WithJob(jobInfo.Job.ID),
		ji:     jobInfo,
		hasher: &hasher{
			salt: pi.Details.Salt,
		},
		noSkip: pi.Details.ReprocessSpec == client.ReprocessSpecEveryJob || pi.Details.S3Out,
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
		startTime, err := types.TimestampFromProto(pj.ji.Started)
		if err != nil {
			return errors.EnsureStack(err)
		}
		timeout, err := types.DurationFromProto(pj.ji.Details.JobTimeout)
		if err != nil {
			return errors.EnsureStack(err)
		}
		afterTime = time.Until(startTime.Add(timeout))
	}
	go func() {
		defer reg.limiter.Release()
		if pj.ji.Details.JobTimeout != nil {
			pj.logger.Logf("cancelling job at: %+v", afterTime)
			timer := time.AfterFunc(afterTime, func() {
				reg.killJob(pj, "job timed out")
			})
			defer timer.Stop()
		}
		if err := backoff.RetryUntilCancel(reg.driver.PachClient().Ctx(), func() error {
			ctx, cancel := context.WithCancel(reg.driver.PachClient().Ctx())
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
				if errors.Is(err, errutil.ErrBreak) || errors.Is(ctx.Err(), context.Canceled) {
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
			pj.ji.Restart++
			return backoff.RetryUntilCancel(reg.driver.PachClient().Ctx(), func() error {
				// Reload the job's commits and info as they may have changed.
				if err := pj.load(); err != nil {
					return err
				}
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
			if err := pj.driver.NewSQLTx(func(sqlTx *pachsql.Tx) error {
				// Delete the job if no other worker has deleted it yet
				jobInfo := &pps.JobInfo{}
				if err := pj.driver.Jobs().ReadWrite(sqlTx).Get(ppsdb.JobKey(pj.ji.Job), jobInfo); err != nil {
					return errors.EnsureStack(err)
				}
				return errors.EnsureStack(pj.driver.DeleteJob(sqlTx, jobInfo))
			}); err != nil && !col.IsErrNotFound(err) {
				return errors.EnsureStack(err)
			}
			return nil
		}
		return errors.EnsureStack(err)
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
		reason := fmt.Sprintf("inputs failed: %s", strings.Join(failed, ", "))
		return reg.failJob(pj, reason)
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
	ctx := pachClient.Ctx()
	taskDoer := reg.driver.NewTaskDoer(pj.ji.Job.ID)
	if err := func() error {
		if err := pj.withParallelDatums(ctx, func(ctx context.Context, dit datum.Iterator) error {
			return reg.processDatums(ctx, pj, taskDoer, dit)
		}); err != nil {
			return err
		}
		return pj.withSerialDatums(ctx, func(ctx context.Context, dit datum.Iterator) error {
			return reg.processDatums(ctx, pj, taskDoer, dit)
		})
	}(); err != nil {
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

func (reg *registry) processDatums(ctx context.Context, pj *pendingJob, taskDoer task.Doer, dit datum.Iterator) error {
	var numDatums int64
	if err := dit.Iterate(func(_ *datum.Meta) error {
		numDatums++
		return nil
	}); err != nil {
		return errors.EnsureStack(err)
	}
	// Set up the datum set spec for the job.
	// When the datum set spec is not set, evenly distribute the datums.
	var setSpec *datum.SetSpec
	datumSetsPerWorker := defaultDatumSetsPerWorker
	if pj.driver.PipelineInfo().Details.DatumSetSpec != nil {
		setSpec = &datum.SetSpec{
			Number:    pj.driver.PipelineInfo().Details.DatumSetSpec.Number,
			SizeBytes: pj.driver.PipelineInfo().Details.DatumSetSpec.SizeBytes,
		}
		datumSetsPerWorker = pj.driver.PipelineInfo().Details.DatumSetSpec.PerWorker
	}
	if setSpec == nil || (setSpec.Number == 0 && setSpec.SizeBytes == 0) {
		setSpec = &datum.SetSpec{Number: numDatums / (int64(reg.concurrency) * datumSetsPerWorker)}
		if setSpec.Number == 0 {
			setSpec.Number = 1
		}
	}
	pachClient := pj.driver.PachClient().WithCtx(ctx)
	inputChan := make(chan *types.Any)
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		eg, ctx := errgroup.WithContext(ctx)
		pachClient = pachClient.WithCtx(ctx)
		// Setup goroutine for creating datum set subtasks.
		eg.Go(func() error {
			defer close(inputChan)
			storageRoot := filepath.Join(pj.driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
			// TODO: The dit needs to iterate with the inner context.
			return datum.CreateSets(dit, storageRoot, setSpec, func(upload func(client.ModifyFile) error) error {
				input, err := createDatumSetTask(pachClient, pj, upload, renewer)
				if err != nil {
					return err
				}
				select {
				case inputChan <- input:
				case <-ctx.Done():
					return errors.EnsureStack(ctx.Err())
				}
				return nil
			})
		})
		// Setup goroutine for running and collecting datum set subtasks.
		eg.Go(func() error {
			err := pj.logger.LogStep("running and collecting datum set subtasks", func() error {
				err := taskDoer.Do(
					ctx,
					inputChan,
					func(_ int64, output *types.Any, err error) error {
						if err != nil {
							return err
						}
						data, err := deserializeDatumSet(output)
						if err != nil {
							return err
						}
						if _, err := pachClient.PfsAPIClient.AddFileSet(
							pachClient.Ctx(),
							&pfs.AddFileSetRequest{
								Commit:    pj.commitInfo.Commit,
								FileSetId: data.OutputFileSetId,
							},
						); err != nil {
							return grpcutil.ScrubGRPC(err)
						}
						if _, err := pachClient.PfsAPIClient.AddFileSet(
							pachClient.Ctx(),
							&pfs.AddFileSetRequest{
								Commit:    pj.metaCommitInfo.Commit,
								FileSetId: data.MetaFileSetId,
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
				return errors.EnsureStack(err)
			})
			return errors.EnsureStack(err)
		})
		return errors.EnsureStack(eg.Wait())
	}); err != nil {
		return err
	}
	if stats.FailedID != "" {
		if err := reg.failJob(pj, fmt.Sprintf("datum %v failed", stats.FailedID)); err != nil {
			return err
		}
		return errutil.ErrBreak
	}
	return nil
}

func createDatumSetTask(pachClient *client.APIClient, pj *pendingJob, upload func(client.ModifyFile) error, renewer *renew.StringSet) (*types.Any, error) {
	resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
		return upload(mf)
	})
	if err != nil {
		return nil, err
	}
	if err := renewer.Add(pachClient.Ctx(), resp.FileSetId); err != nil {
		return nil, err
	}
	return serializeDatumSet(&DatumSet{
		JobID:        pj.ji.Job.ID,
		OutputCommit: pj.commitInfo.Commit,
		// TODO: It might make sense for this to be a hash of the constituent datums?
		// That could make it possible to recover from a master restart.
		FileSetId: resp.FileSetId,
	})
}

func serializeDatumSet(data *DatumSet) (*types.Any, error) {
	serialized, err := types.MarshalAny(data)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return serialized, nil
}

func deserializeDatumSet(any *types.Any) (*DatumSet, error) {
	data := &DatumSet{}
	if err := types.UnmarshalAny(any, data); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return data, nil
}

func (reg *registry) processJobEgressing(pj *pendingJob) error {
	url := pj.ji.Details.Egress.URL
	err := pj.driver.PachClient().GetFileURL(pj.commitInfo.Commit, "/", url)
	// file not found means the commit is empty, nothing to egress
	if err != nil && !pfsserver.IsFileNotFoundErr(err) {
		return err
	}
	return reg.succeedJob(pj)
}

func failedInputs(pachClient *client.APIClient, jobInfo *pps.JobInfo) ([]string, error) {
	var failed []string
	waitCommit := func(name string, commit *pfs.Commit) error {
		ci, err := pachClient.WaitCommit(commit.Branch.Repo.Name, commit.Branch.Name, commit.ID)
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
