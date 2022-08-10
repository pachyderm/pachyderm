package transform

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
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
	if err := datum.MergeProcessStats(pj.ji.Stats, stats.ProcessStats); err != nil {
		pj.logger.Logf("errored merging process stats: %v", err)
	}
	pj.ji.DataProcessed += stats.Processed
	pj.ji.DataSkipped += stats.Skipped
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
		outputCI, err := pachClient.InspectCommit(pj.baseMetaCommit.Branch.Repo.Name,
			pj.baseMetaCommit.Branch.Name, pj.baseMetaCommit.ID)
		if err != nil {
			return errors.EnsureStack(err)
		}
		// both commits must have succeeded - a validation error will only show up in the output
		if metaCI.Error == "" && outputCI.Error == "" {
			break
		}
		pj.baseMetaCommit = metaCI.ParentCommit
	}
	// Load the job info.
	pj.ji, err = pachClient.InspectJob(pj.ji.Job.Pipeline.Name, pj.ji.Job.ID, true)
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

func (pj *pendingJob) withDeleter(pachClient *client.APIClient, cb func(datum.Deleter) error) error {
	// Setup modify file client for meta commit.
	metaCommit := pj.metaCommitInfo.Commit
	return pachClient.WithModifyFileClient(metaCommit, func(mfMeta client.ModifyFile) error {
		// Setup modify file client for output commit.
		outputCommit := pj.commitInfo.Commit
		return pachClient.WithModifyFileClient(outputCommit, func(mfPFS client.ModifyFile) error {
			baseMetaCommit := pj.baseMetaCommit
			metaFileWalker := func(path string) ([]string, error) {
				var files []string
				if err := pachClient.WalkFile(baseMetaCommit, path, func(fi *pfs.FileInfo) error {
					if fi.FileType == pfs.FileType_FILE {
						files = append(files, fi.File.Path)
					}
					return nil
				}); err != nil {
					return nil, err
				}
				return files, nil
			}
			return cb(datum.NewDeleter(metaFileWalker, mfMeta, mfPFS))
		})
	})
}

// writeDatumCount gets the total count of input datums for the job and updates
// its DataTotal field.
func (pj *pendingJob) writeDatumCount(ctx context.Context, taskDoer task.Doer) error {
	pachClient := pj.driver.PachClient().WithCtx(ctx)
	var count int
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		// Upload the datums from the current job into the datum file set format.
		var err error
		_, count, err = pj.createFullJobDatumFileSet(ctx, taskDoer, renewer)
		return err
	}); err != nil {
		return err
	}
	pj.ji.DataTotal = int64(count)
	return pj.writeJobInfo()
}

// The datums that can be processed in parallel are the datums that exist in the current job and do not exist in the base job.
func (pj *pendingJob) withParallelDatums(ctx context.Context, taskDoer task.Doer, cb func(context.Context, string) error) error {
	pachClient := pj.driver.PachClient().WithCtx(ctx)
	return pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		// Upload the datums from the current job into the datum file set format.
		fileSetID, _, err := pj.createFullJobDatumFileSet(ctx, taskDoer, renewer)
		if err != nil {
			return err
		}
		var baseFileSetID string
		if pj.baseMetaCommit != nil {
			// Upload the datums from the base job into the datum file set format.
			var err error
			baseFileSetID, err = pj.createFullBaseJobDatumFileSet(ctx, taskDoer, renewer)
			if err != nil {
				return err
			}
		}
		// Create the output datum file set for the new datums (datums that do not exist in the base job).
		outputFileSetID, err := pj.createJobDatumFileSetParallel(ctx, taskDoer, renewer, fileSetID, baseFileSetID)
		if err != nil {
			return err
		}
		return cb(ctx, outputFileSetID)
	})
}

// TODO: There is probably a way to reduce the boilerplate for running all of these preprocessing tasks.
func (pj *pendingJob) createFullJobDatumFileSet(ctx context.Context, taskDoer task.Doer, renewer *renew.StringSet) (string, int, error) {
	var (
		fileSetID string
		count     int
	)
	if err := pj.logger.LogStep("creating full job datum file set", func() error {
		input, err := serializeUploadDatumsTask(&UploadDatumsTask{Job: pj.ji.Job})
		if err != nil {
			return err
		}
		output, err := task.DoOne(ctx, taskDoer, input)
		if err != nil {
			return err
		}
		result, err := deserializeUploadDatumsTaskResult(output)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, result.FileSetId); err != nil {
			return err
		}
		fileSetID = result.FileSetId
		count = int(result.Count)
		return nil
	}); err != nil {
		return "", 0, errors.EnsureStack(err)
	}
	return fileSetID, count, nil
}

func (pj *pendingJob) createFullBaseJobDatumFileSet(ctx context.Context, taskDoer task.Doer, renewer *renew.StringSet) (string, error) {
	var baseFileSetID string
	if err := pj.logger.LogStep("creating full base job datum file set", func() error {
		input, err := serializeUploadDatumsTask(&UploadDatumsTask{Job: client.NewJob(pj.ji.Job.Pipeline.Name, pj.baseMetaCommit.ID)})
		if err != nil {
			return err
		}
		output, err := task.DoOne(ctx, taskDoer, input)
		if err != nil {
			return err
		}
		result, err := deserializeUploadDatumsTaskResult(output)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, result.FileSetId); err != nil {
			return err
		}
		baseFileSetID = result.FileSetId
		return nil
	}); err != nil {
		return "", errors.EnsureStack(err)
	}
	return baseFileSetID, nil
}

func (pj *pendingJob) createJobDatumFileSetParallel(ctx context.Context, taskDoer task.Doer, renewer *renew.StringSet, fileSetID, baseFileSetID string) (string, error) {
	var outputFileSetID string
	if err := pj.logger.LogStep("creating job datum file set (parallel jobs)", func() error {
		input, err := serializeComputeParallelDatumsTask(&ComputeParallelDatumsTask{
			Job:           pj.ji.Job,
			FileSetId:     fileSetID,
			BaseFileSetId: baseFileSetID,
		})
		if err != nil {
			return err
		}
		output, err := task.DoOne(ctx, taskDoer, input)
		if err != nil {
			return err
		}
		result, err := deserializeComputeParallelDatumsTaskResult(output)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, result.FileSetId); err != nil {
			return err
		}
		outputFileSetID = result.FileSetId
		return nil
	}); err != nil {
		return "", errors.EnsureStack(err)
	}
	return outputFileSetID, nil
}

// The datums that must be processed serially (with respect to the base job) are the datums that exist in both the current and base job.
// A datum is skipped if it exists in both jobs with the same hash and was successfully processed by the base.
// Deletion operations are created for the datums that need to be removed from the base job output commits.
func (pj *pendingJob) withSerialDatums(ctx context.Context, taskDoer task.Doer, cb func(context.Context, string) error) error {
	// There are no serial datums if no base exists.
	if pj.baseMetaCommit == nil {
		return nil
	}
	pachClient := pj.driver.PachClient().WithCtx(ctx)
	// Wait for the base job to finish.
	ci, err := pachClient.WaitCommit(pj.baseMetaCommit.Branch.Repo.Name, pj.baseMetaCommit.Branch.Name, pj.baseMetaCommit.ID)
	if err != nil {
		return err
	}
	if ci.Error != "" {
		return pfsserver.ErrCommitError{Commit: ci.Commit}
	}
	return pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		// Upload the datums from the current job into the datum file set format.
		fileSetID, _, err := pj.createFullJobDatumFileSet(ctx, taskDoer, renewer)
		if err != nil {
			return err
		}
		// Create the output datum file set for the datums that were not processed by the base (failed, recovered, etc.).
		outputFileSetID, deleteFileSetID, skipped, err := pj.createJobDatumFileSetSerial(ctx, taskDoer, renewer, fileSetID, pj.baseMetaCommit)
		if err != nil {
			return err
		}
		// Delete the appropriate files.
		if err := pj.logger.LogStep("deleting old datum outputs", func() error {
			pachClient := pachClient.WithCtx(ctx)
			return pj.withDeleter(pachClient, func(deleter datum.Deleter) error {
				dit := datum.NewFileSetIterator(pachClient, deleteFileSetID)
				return errors.EnsureStack(dit.Iterate(func(meta *datum.Meta) error {
					return deleter(meta)
				}))
			})
		}); err != nil {
			return errors.EnsureStack(err)
		}
		// Record the skipped datums.
		stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
		stats.Skipped = skipped
		pj.saveJobStats(stats)
		if err := pj.writeJobInfo(); err != nil {
			return err
		}
		return cb(ctx, outputFileSetID)
	})
}

func (pj *pendingJob) createJobDatumFileSetSerial(ctx context.Context, taskDoer task.Doer, renewer *renew.StringSet, fileSetID string, baseMetaCommit *pfs.Commit) (string, string, int64, error) {
	var outputFileSetID, deleteFileSetID string
	var skipped int64
	if err := pj.logger.LogStep("creating job datum file set (serial jobs)", func() error {
		input, err := serializeComputeSerialDatumsTask(&ComputeSerialDatumsTask{
			Job:            pj.ji.Job,
			FileSetId:      fileSetID,
			BaseMetaCommit: baseMetaCommit,
			NoSkip:         pj.noSkip,
		})
		if err != nil {
			return err
		}
		output, err := task.DoOne(ctx, taskDoer, input)
		if err != nil {
			return err
		}
		result, err := deserializeComputeSerialDatumsTaskResult(output)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, result.FileSetId); err != nil {
			return err
		}
		if err := renewer.Add(ctx, result.DeleteFileSetId); err != nil {
			return err
		}
		outputFileSetID = result.FileSetId
		deleteFileSetID = result.DeleteFileSetId
		skipped = result.Skipped
		return nil
	}); err != nil {
		return "", "", 0, errors.EnsureStack(err)
	}
	return outputFileSetID, deleteFileSetID, skipped, nil
}

// TODO: We might be better off removing the serialize boilerplate and switching to types.MarshalAny.
func serializeUploadDatumsTask(task *UploadDatumsTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeUploadDatumsTaskResult(taskAny *types.Any) (*UploadDatumsTaskResult, error) {
	task := &UploadDatumsTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeComputeParallelDatumsTask(task *ComputeParallelDatumsTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeComputeParallelDatumsTaskResult(taskAny *types.Any) (*ComputeParallelDatumsTaskResult, error) {
	task := &ComputeParallelDatumsTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeComputeSerialDatumsTask(task *ComputeSerialDatumsTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeComputeSerialDatumsTaskResult(taskAny *types.Any) (*ComputeSerialDatumsTaskResult, error) {
	task := &ComputeSerialDatumsTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}
