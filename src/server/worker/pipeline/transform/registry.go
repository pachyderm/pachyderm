package transform

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
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
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

const (
	defaultDatumSetsPerWorker int64 = 4
)

type registry struct {
	driver  driver.Driver
	logger  logs.TaggedLogger
	limiter limit.ConcurrencyLimiter
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
				if err := reg.killJob(pj, " jobtimed out"); err != nil {
					pj.logger.Logf("error killing job: %v", err)
				}
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
	ctx := pachClient.Ctx()
	taskDoer := reg.driver.NewTaskDoer(pj.ji.Job.ID, pj.cache)
	if err := func() error {
		if err := pj.writeDatumCount(ctx, taskDoer); err != nil {
			return err
		}
		if err := pj.withParallelDatums(ctx, taskDoer, func(ctx context.Context, fileSetID string) error {
			return reg.processDatums(ctx, pj, taskDoer, fileSetID)
		}); err != nil {
			return err
		}
		return pj.withSerialDatums(ctx, taskDoer, func(ctx context.Context, fileSetID string) error {
			return reg.processDatums(ctx, pj, taskDoer, fileSetID)
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

func (reg *registry) processDatums(ctx context.Context, pj *pendingJob, taskDoer task.Doer, fileSetID string) error {
	pachClient := pj.driver.PachClient().WithCtx(ctx)
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		datumSetsFileSetID, err := createDatumSets(ctx, pj, taskDoer, renewer, fileSetID)
		if err != nil {
			return err
		}
		return processDatumSets(pachClient.WithCtx(ctx), pj, taskDoer, datumSetsFileSetID, stats)
	}); err != nil {
		return errors.EnsureStack(err)
	}
	if stats.FailedID != "" {
		if err := reg.failJob(pj, fmt.Sprintf("datum %v failed", stats.FailedID)); err != nil {
			return err
		}
		return errutil.ErrBreak
	}
	return nil
}

func createDatumSets(ctx context.Context, pj *pendingJob, taskDoer task.Doer, renewer *renew.StringSet, fileSetID string) (string, error) {
	var datumSetsFileSetID string
	if err := pj.logger.LogStep("creating datum sets", func() error {
		input, err := serializeCreateDatumSetsTask(&CreateDatumSetsTask{
			Job:          pj.ji.Job,
			OutputCommit: pj.ji.OutputCommit,
			FileSetId:    fileSetID,
		})
		if err != nil {
			return err
		}
		output, err := task.DoOne(ctx, taskDoer, input)
		if err != nil {
			return err
		}
		result, err := deserializeCreateDatumSetsTaskResult(output)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, result.FileSetId); err != nil {
			return err
		}
		if err := renewer.Add(ctx, result.InputFileSetsId); err != nil {
			return err
		}
		datumSetsFileSetID = result.FileSetId
		return nil
	}); err != nil {
		return "", errors.EnsureStack(err)
	}
	return datumSetsFileSetID, nil
}

func processDatumSets(pachClient *client.APIClient, pj *pendingJob, taskDoer task.Doer, fileSetID string, stats *datum.Stats) error {
	return errors.EnsureStack(pj.logger.LogStep("processing datum sets", func() error {
		eg, ctx := errgroup.WithContext(pachClient.Ctx())
		pachClient := pachClient.WithCtx(ctx)
		inputChan := make(chan *types.Any)
		eg.Go(func() error {
			defer close(inputChan)
			commit := client.NewRepo(client.FileSetsRepoName).NewCommit("", fileSetID)
			r, err := pachClient.GetFileTAR(commit, "/*")
			if err != nil {
				return err
			}
			if err := tarutil.Iterate(r, func(f tarutil.File) error {
				buf := &bytes.Buffer{}
				if err := f.Content(buf); err != nil {
					return errors.EnsureStack(err)
				}
				input := &types.Any{}
				if err := proto.Unmarshal(buf.Bytes(), input); err != nil {
					return errors.EnsureStack(err)
				}
				select {
				case inputChan <- input:
				case <-ctx.Done():
					return errors.EnsureStack(ctx.Err())
				}
				return nil
			}, true); err != nil && !pfsserver.IsFileNotFoundErr(err) {
				return err
			}
			return nil
		})
		eg.Go(func() error {
			return errors.EnsureStack(taskDoer.Do(
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
			))
		})
		return errors.EnsureStack(eg.Wait())
	}))
}

func serializeCreateDatumSetsTask(task *CreateDatumSetsTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeCreateDatumSetsTaskResult(taskAny *types.Any) (*CreateDatumSetsTaskResult, error) {
	task := &CreateDatumSetsTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
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
