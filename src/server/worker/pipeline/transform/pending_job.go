package transform

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
	"google.golang.org/protobuf/types/known/anypb"
)

type pendingJob struct {
	driver                     driver.Driver
	logger                     logs.TaggedLogger
	cancel                     context.CancelFunc
	ji                         *pps.JobInfo
	commitInfo, metaCommitInfo *pfs.CommitInfo
	baseMetaCommit             *pfs.Commit
	noSkip                     bool
	cache                      *cache
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
	pj.ji.DataTotal += stats.Total
	pj.ji.DataFailed += stats.Failed
	pj.ji.DataRecovered += stats.Recovered
}

func (pj *pendingJob) load() error {
	pachClient := pj.driver.PachClient()
	var err error
	// Load and clear the output commit.
	pj.commitInfo, err = pachClient.PfsAPIClient.InspectCommit(
		pachClient.Ctx(),
		&pfs.InspectCommitRequest{
			Commit: pj.ji.OutputCommit,
			Wait:   pfs.CommitState_STARTED,
		})
	if err != nil {
		return errors.EnsureStack(err)
	}
	if _, err := pachClient.PfsAPIClient.ClearCommit(
		pachClient.Ctx(),
		&pfs.ClearCommitRequest{
			Commit: pj.ji.OutputCommit,
		}); err != nil {
		return errors.EnsureStack(err)
	}
	// Load and clear the meta commit.
	pj.metaCommitInfo, err = pachClient.PfsAPIClient.InspectCommit(
		pachClient.Ctx(),
		&pfs.InspectCommitRequest{
			Commit: ppsutil.MetaCommit(pj.ji.OutputCommit),
			Wait:   pfs.CommitState_STARTED,
		})
	if err != nil {
		return errors.EnsureStack(err)
	}
	if _, err := pachClient.PfsAPIClient.ClearCommit(
		pachClient.Ctx(),
		&pfs.ClearCommitRequest{
			Commit: ppsutil.MetaCommit(pj.ji.OutputCommit),
		}); err != nil {
		return errors.EnsureStack(err)
	}
	// Find the most recent successful ancestor commit to use as the
	// base for this job.
	// TODO: This should be an operation supported and exposed by PFS.
	pj.baseMetaCommit = pj.metaCommitInfo.ParentCommit
	for pj.baseMetaCommit != nil {
		metaCI, err := pachClient.PfsAPIClient.InspectCommit(
			pachClient.Ctx(),
			&pfs.InspectCommitRequest{
				Commit: pj.baseMetaCommit,
				Wait:   pfs.CommitState_STARTED,
			})
		if err != nil {
			return errors.EnsureStack(err)
		}
		if metaCI.Origin.Kind == pfs.OriginKind_AUTO {
			outputCI, err := pachClient.InspectCommit(pj.baseMetaCommit.Repo.Project.GetName(), pj.baseMetaCommit.Repo.Name, "", pj.baseMetaCommit.Id)
			if err != nil {
				return errors.EnsureStack(err)
			}
			// both commits must have succeeded - a validation error will only show up in the output
			if metaCI.Error == "" && outputCI.Error == "" {
				// Load the job info.
				if _, err = pachClient.InspectJob(pj.ji.Job.Pipeline.Project.GetName(), pj.ji.Job.Pipeline.Name, metaCI.Commit.Id, false); err != nil {
					if !errutil.IsNotFoundError(err) {
						return err
					}
					// Jobs are always expected to exist. In case a pachyderm instance contains a meta commit without a job,
					// we gracefully handle it by skipping it for base meta commit selection.
					pj.logger.Logf("base meta commit %q could not be selected for job %q because it was missing a job was missing an associated job",
						metaCI.Commit.String(), pj.ji.Job.String())
				} else {
					break
				}
			}
		}
		pj.baseMetaCommit = metaCI.ParentCommit
	}
	if pj.baseMetaCommit != nil {
		pj.logger.Logf("base meta commit for job %q is %q", pj.ji.Job.String(), pj.baseMetaCommit.Id)
	} else {
		pj.logger.Logf("base meta commit for job %q not selected", pj.ji.Job.String())
	}
	pj.ji, err = pachClient.InspectJob(pj.ji.Job.Pipeline.Project.GetName(), pj.ji.Job.Pipeline.Name, pj.ji.Job.Id, true)
	if err != nil {
		return err
	}
	pj.clearJobStats()
	return nil
}

func (pj *pendingJob) clearJobStats() {
	pj.ji.Stats = &pps.ProcessStats{}
	pj.ji.DataProcessed = 0
	pj.ji.DataSkipped = 0
	pj.ji.DataFailed = 0
	pj.ji.DataRecovered = 0
	pj.ji.DataTotal = 0
}

// TODO: Remove when job state transition operations are handled by a background process.
func (pj *pendingJob) clearCache() {
	if err := pj.cache.clear(pj.driver.PachClient().Ctx()); err != nil {
		pj.logger.Logf("errored clearing job cache: %s", err)
	}
}

// The datums that can be processed in parallel are the datums that exist in the current job and do not exist in the base job.
func (pj *pendingJob) createParallelDatums(ctx context.Context, taskDoer task.Doer) (string, error) {
	pachClient := pj.driver.PachClient().WithCtx(ctx)
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		// Upload the datums from the current job into the datum file set format.
		var fileSetID string
		var err error
		if err := pj.logger.LogStep("creating full job datum file set", func() error {
			fileSetID, err = createDatums(pachClient, taskDoer, pj.ji.Job)
			if err != nil {
				return err
			}
			return renewer.Add(ctx, fileSetID)
		}); err != nil {
			return errors.EnsureStack(err)
		}
		var baseFileSetID string
		if pj.baseMetaCommit != nil {
			// Upload the datums from the base job into the datum file set format.
			if err := pj.logger.LogStep("creating full base job datum file set", func() error {
				baseFileSetID, err = createDatums(pachClient, taskDoer, client.NewJob(pj.ji.Job.Pipeline.Project.GetName(), pj.ji.Job.Pipeline.Name, pj.baseMetaCommit.Id))
				if err != nil {
					return err
				}
				return renewer.Add(ctx, baseFileSetID)
			}); err != nil {
				return errors.EnsureStack(err)
			}
		}
		// Create the output datum file set for the new datums (datums that do not exist in the base job).
		outputFileSetID, err = pj.createJobDatumFileSetParallel(ctx, taskDoer, renewer, fileSetID, baseFileSetID)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createDatums(pachClient *client.APIClient, taskDoer task.Doer, job *pps.Job) (string, error) {
	jobInfo, err := pachClient.InspectJob(job.Pipeline.Project.GetName(), job.Pipeline.Name, job.Id, true)
	if err != nil {
		return "", err
	}
	metaCommitInfo, err := pachClient.PfsAPIClient.InspectCommit(
		pachClient.Ctx(),
		&pfs.InspectCommitRequest{
			Commit: ppsutil.MetaCommit(jobInfo.OutputCommit),
		})
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	if metaCommitInfo.Finished != nil {
		resp, err := pachClient.PfsAPIClient.GetFileSet(
			pachClient.Ctx(),
			&pfs.GetFileSetRequest{
				Commit: metaCommitInfo.Commit,
			},
		)
		if err != nil {
			return "", errors.EnsureStack(err)
		}
		return resp.FileSetId, nil
	}
	return datum.Create(pachClient.Ctx(), pachClient.PfsAPIClient, taskDoer, jobInfo.Details.Input)
}

func (pj *pendingJob) createJobDatumFileSetParallel(ctx context.Context, taskDoer task.Doer, renewer *renew.StringSet, fileSetID, baseFileSetID string) (string, error) {
	var outputFileSetID string
	if err := pj.logger.LogStep("creating job datum file set (parallel jobs)", func() error {
		pachClient := pj.driver.PachClient().WithCtx(ctx)
		return pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
			pachClient := pachClient.WithCtx(ctx)
			// TODO: We may want to create additional shards if the new job has much less datums.
			shards, err := pachClient.ShardFileSet(fileSetID)
			if err != nil {
				return err
			}
			var inputs []*anypb.Any
			for _, shard := range shards {
				input, err := serializeCreateParallelDatumsTask(&CreateParallelDatumsTask{
					Job:           pj.ji.Job,
					Salt:          pj.ji.Details.Salt,
					FileSetId:     fileSetID,
					BaseFileSetId: baseFileSetID,
					PathRange:     shard,
					AuthToken:     pachClient.AuthToken(),
				})
				if err != nil {
					return err
				}
				inputs = append(inputs, input)
			}
			resultFileSetIDs := make([]string, len(inputs))
			if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
				if err != nil {
					return err
				}
				result, err := deserializeCreateParallelDatumsTaskResult(output)
				if err != nil {
					return err
				}
				if err := renewer.Add(ctx, result.FileSetId); err != nil {
					return err
				}
				resultFileSetIDs[i] = result.FileSetId
				pj.saveJobStats(result.Stats)
				return pj.writeJobInfo()
			}); err != nil {
				return err
			}
			outputFileSetID, err = datum.ComposeFileSets(pachClient.Ctx(), pachClient.PfsAPIClient, taskDoer, resultFileSetIDs)
			return err
		})
	}); err != nil {
		return "", errors.EnsureStack(err)
	}
	return outputFileSetID, nil
}

// The datums that must be processed serially (with respect to the base job) are the datums that exist in both the current and base job.
// A datum is skipped if it exists in both jobs with the same hash and was successfully processed by the base.
// Deletion operations are created for the datums that need to be removed from the base job output commits.
func (pj *pendingJob) createSerialDatums(ctx context.Context, taskDoer task.Doer) (string, error) {
	pachClient := pj.driver.PachClient().WithCtx(ctx)
	// There are no serial datums if no base exists.
	if pj.baseMetaCommit == nil {
		return datum.CreateEmptyFileSet(pachClient.Ctx(), pachClient.PfsAPIClient)
	}
	var ci *pfs.CommitInfo
	var err error
	// Wait for the base job to finish.
	if err := backoff.RetryUntilCancel(ctx, func() error {
		ci, err = pachClient.WaitCommit(pj.baseMetaCommit.Repo.Project.GetName(), pj.baseMetaCommit.Repo.Name, "", pj.baseMetaCommit.Id)
		if errutil.IsDatabaseDisconnect(err) {
			return backoff.ErrContinue
		}
		if err != nil {
			return err
		}
		return nil
	}, backoff.RetryEvery(time.Second), nil); err != nil {
		return "", errors.Wrapf(err, "wait for meta commit %q", pj.baseMetaCommit.String())
	}
	if ci.Error != "" {
		return "", pfsserver.ErrCommitError{Commit: ci.Commit}
	}
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		// Upload the datums from the current job into the datum file set format.
		var fileSetID string
		if err := pj.logger.LogStep("creating full job datum file set", func() error {
			fileSetID, err = createDatums(pachClient, taskDoer, pj.ji.Job)
			if err != nil {
				return err
			}
			return renewer.Add(ctx, fileSetID)
		}); err != nil {
			return errors.EnsureStack(err)
		}
		// Create the output datum file set for the datums that were not processed by the base (failed, recovered, etc.).
		outputFileSetID, err = pj.createJobDatumFileSetSerial(ctx, taskDoer, renewer, fileSetID, pj.baseMetaCommit)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func (pj *pendingJob) createJobDatumFileSetSerial(ctx context.Context, taskDoer task.Doer, renewer *renew.StringSet, fileSetID string, baseMetaCommit *pfs.Commit) (string, error) {
	var outputFileSetID string
	if err := pj.logger.LogStep("creating job datum file set (serial jobs)", func() error {
		pachClient := pj.driver.PachClient().WithCtx(ctx)
		return pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
			pachClient := pachClient.WithCtx(ctx)
			// TODO: We may want to create additional shards if the new job has much less datums.
			shards, err := pachClient.ShardFileSet(fileSetID)
			if err != nil {
				return err
			}
			var inputs []*anypb.Any
			for _, shard := range shards {
				input, err := serializeCreateSerialDatumsTask(&CreateSerialDatumsTask{
					Job:            pj.ji.Job,
					Salt:           pj.ji.Details.Salt,
					FileSetId:      fileSetID,
					BaseMetaCommit: baseMetaCommit,
					NoSkip:         pj.noSkip,
					PathRange:      shard,
					AuthToken:      pachClient.AuthToken(),
				})
				if err != nil {
					return err
				}
				inputs = append(inputs, input)
			}
			resultFileSetIDs := make([]string, len(inputs))
			if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
				if err != nil {
					return err
				}
				result, err := deserializeCreateSerialDatumsTaskResult(output)
				if err != nil {
					return err
				}
				if err := renewer.Add(ctx, result.FileSetId); err != nil {
					return err
				}
				resultFileSetIDs[i] = result.FileSetId
				if _, err := pachClient.PfsAPIClient.AddFileSet(
					ctx,
					&pfs.AddFileSetRequest{
						Commit:    pj.commitInfo.Commit,
						FileSetId: result.OutputDeleteFileSetId,
					},
				); err != nil {
					return errors.EnsureStack(err)
				}
				if _, err := pachClient.PfsAPIClient.AddFileSet(
					ctx,
					&pfs.AddFileSetRequest{
						Commit:    pj.metaCommitInfo.Commit,
						FileSetId: result.MetaDeleteFileSetId,
					},
				); err != nil {
					return errors.EnsureStack(err)
				}
				pj.saveJobStats(result.Stats)
				return pj.writeJobInfo()
			}); err != nil {
				return err
			}
			outputFileSetID, err = datum.ComposeFileSets(pachClient.Ctx(), pachClient.PfsAPIClient, taskDoer, resultFileSetIDs)
			return err
		})
	}); err != nil {
		return "", errors.EnsureStack(err)
	}
	return outputFileSetID, nil
}

// TODO: We might be better off removing the serialize boilerplate and switching to types.MarshalAny.
func serializeCreateParallelDatumsTask(task *CreateParallelDatumsTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeCreateParallelDatumsTaskResult(taskAny *anypb.Any) (*CreateParallelDatumsTaskResult, error) {
	task := &CreateParallelDatumsTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeCreateSerialDatumsTask(task *CreateSerialDatumsTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeCreateSerialDatumsTaskResult(taskAny *anypb.Any) (*CreateSerialDatumsTaskResult, error) {
	task := &CreateSerialDatumsTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}
