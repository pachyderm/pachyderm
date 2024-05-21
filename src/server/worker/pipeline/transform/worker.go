package transform

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfssync"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

type hasher struct {
	salt string
}

func (h *hasher) Hash(inputs []*common.Input) string {
	return common.HashDatum(h.salt, inputs)
}

func PreprocessingWorker(pachClient *client.APIClient, taskService task.Service, pipelineInfo *pps.PipelineInfo) error {
	taskSource := taskService.NewSource(driver.PreprocessingTaskNamespace(pipelineInfo))
	return errors.EnsureStack(taskSource.Iterate(
		pachClient.Ctx(),
		func(ctx context.Context, input *anypb.Any) (*anypb.Any, error) {
			pachClient := pachClient.WithCtx(ctx)
			switch {
			case datum.IsTask(input):
				return datum.ProcessTask(ctx, pachClient.PfsAPIClient, input)
			case input.MessageIs(&CreateParallelDatumsTask{}):
				createParallelDatumsTask, err := deserializeCreateParallelDatumsTask(input)
				if err != nil {
					return nil, err
				}
				pachClient.SetAuthToken(createParallelDatumsTask.AuthToken)
				return processCreateParallelDatumsTask(pachClient, createParallelDatumsTask)
			case input.MessageIs(&CreateSerialDatumsTask{}):
				createSerialDatumsTask, err := deserializeCreateSerialDatumsTask(input)
				if err != nil {
					return nil, err
				}
				pachClient.SetAuthToken(createSerialDatumsTask.AuthToken)
				return processCreateSerialDatumsTask(pachClient, createSerialDatumsTask)
			case input.MessageIs(&CreateDatumSetsTask{}):
				createDatumSetsTask, err := deserializeCreateDatumSetsTask(input)
				if err != nil {
					return nil, err
				}
				pachClient.SetAuthToken(createDatumSetsTask.AuthToken)
				return processCreateDatumSetsTask(pachClient, createDatumSetsTask)
			default:
				return nil, errors.Errorf("unrecognized any type (%v) in preprocessing worker", input.TypeUrl)
			}
		},
	))
}

func ProcessingWorker(ctx context.Context, driver driver.Driver, logger logs.TaggedLogger, status *Status) error {
	return errors.EnsureStack(driver.NewTaskSource().Iterate(
		ctx,
		func(ctx context.Context, input *anypb.Any) (*anypb.Any, error) {
			switch {
			case input.MessageIs(&DatumSetTask{}):
				datumSetTask, err := deserializeDatumSetTask(input)
				if err != nil {
					return nil, err
				}
				driver := driver.WithContext(ctx)
				return processDatumSetTask(driver, logger, datumSetTask, status)
			default:
				return nil, errors.Errorf("unrecognized any type (%v) in processing worker", input.TypeUrl)
			}
		},
	))
}

func processCreateParallelDatumsTask(pachClient *client.APIClient, task *CreateParallelDatumsTask) (*anypb.Any, error) {
	var dits []datum.Iterator
	if task.BaseFileSetId != "" {
		dits = append(dits, datum.NewFileSetIterator(pachClient.Ctx(), pachClient.PfsAPIClient, task.BaseFileSetId, task.PathRange))
	}
	dit := datum.NewFileSetIterator(pachClient.Ctx(), pachClient.PfsAPIClient, task.FileSetId, task.PathRange)
	dit = datum.NewJobIterator(dit, task.Job, &hasher{salt: task.Salt})
	dits = append(dits, dit)
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	outputFileSetID, err := datum.WithCreateFileSet(pachClient.Ctx(), pachClient.PfsAPIClient, "pachyderm-create-parallel-datums", func(outputSet *datum.Set) error {
		return datum.Merge(dits, func(metas []*datum.Meta) error {
			// Datum exists in both jobs.
			if len(metas) > 1 {
				stats.Total++
				return nil
			}
			// Datum only exists in the parent job.
			if !proto.Equal(metas[0].Job, task.Job) {
				return nil
			}
			if err := outputSet.UploadMeta(metas[0], datum.WithPrefixIndex()); err != nil {
				return err
			}
			stats.Total++
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return serializeCreateParallelDatumsTaskResult(&CreateParallelDatumsTaskResult{
		FileSetId: outputFileSetID,
		Stats:     stats,
	})
}

func processCreateSerialDatumsTask(pachClient *client.APIClient, task *CreateSerialDatumsTask) (*anypb.Any, error) {
	dit := datum.NewFileSetIterator(pachClient.Ctx(), pachClient.PfsAPIClient, task.FileSetId, task.PathRange)
	dit = datum.NewJobIterator(dit, task.Job, &hasher{salt: task.Salt})
	dits := []datum.Iterator{
		datum.NewCommitIterator(pachClient.Ctx(), pachClient.PfsAPIClient, task.BaseMetaCommit, task.PathRange),
		dit,
	}
	var metaDeleteFileSetID, outputDeleteFileSetID string
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	outputFileSetID, err := datum.WithCreateFileSet(pachClient.Ctx(), pachClient.PfsAPIClient, "pachyderm-create-serial-datums", func(outputSet *datum.Set) error {
		var err error
		outputDeleteFileSetID, metaDeleteFileSetID, err = withDeleter(pachClient, task.BaseMetaCommit, func(deleter datum.Deleter) error {
			return datum.Merge(dits, func(metas []*datum.Meta) error {
				if len(metas) == 1 {
					// Datum was processed in the parallel step.
					if proto.Equal(metas[0].Job, task.Job) {
						return nil
					}
					// Datum only exists in the parent job.
					return deleter(metas[0])
				}
				// Check if a skippable datum was successfully processed by the parent.
				if !task.NoSkip && skippableDatum(metas[1], metas[0]) {
					stats.Skipped++
					return nil
				}
				if err := deleter(metas[0]); err != nil {
					return err
				}
				return outputSet.UploadMeta(metas[1], datum.WithPrefixIndex())
			})
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return serializeCreateSerialDatumsTaskResult(&CreateSerialDatumsTaskResult{
		FileSetId:             outputFileSetID,
		OutputDeleteFileSetId: outputDeleteFileSetID,
		MetaDeleteFileSetId:   metaDeleteFileSetID,
		Stats:                 stats,
	})
}

func skippableDatum(meta1, meta2 *datum.Meta) bool {
	// If the hashes are equal and the second datum was processed, then skip it.
	return meta1.Hash == meta2.Hash && meta2.State == datum.State_PROCESSED
}

func withDeleter(pachClient *client.APIClient, baseMetaCommit *pfs.Commit, cb func(datum.Deleter) error) (string, string, error) {
	var outputFileSetID, metaFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		resp, err := pachClient.WithCreateFileSetClient(func(mfMeta client.ModifyFile) error {
			resp, err := pachClient.WithCreateFileSetClient(func(mfPFS client.ModifyFile) error {
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
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, resp.FileSetId); err != nil {
				return err
			}
			outputFileSetID = resp.FileSetId
			return nil
		})
		if err != nil {
			return err
		}
		metaFileSetID = resp.FileSetId
		return nil
	}); err != nil {
		return "", "", err
	}
	return outputFileSetID, metaFileSetID, nil
}

func processCreateDatumSetsTask(pachClient *client.APIClient, task *CreateDatumSetsTask) (*anypb.Any, error) {
	datumSets, err := datum.CreateSets(pachClient, task.SetSpec, task.FileSetId, task.PathRange)
	if err != nil {
		return nil, err
	}
	return serializeCreateDatumSetsTaskResult(&CreateDatumSetsTaskResult{
		DatumSets: datumSets,
	})
}

// Worker handles a transform pipeline work subtask, then returns.
// TODO:
// datum queuing (probably should be handled by datum package).
// capture datum logs.
// git inputs.
func processDatumSetTask(driver driver.Driver, logger logs.TaggedLogger, task *DatumSetTask, status *Status) (*anypb.Any, error) {
	var output *anypb.Any
	if err := status.withJob(task.Job.Id, func() error {
		logger = logger.WithJob(task.Job.Id)
		return errors.EnsureStack(logger.LogStep("process datum set task", func() error {
			if ppsutil.ContainsS3Inputs(driver.PipelineInfo().Details.Input) || driver.PipelineInfo().Details.S3Out {
				if err := checkS3Gateway(driver, logger); err != nil {
					return err
				}
			}
			var err error
			output, err = handleDatumSet(driver, logger, task, status)
			return err
		}))
	}); err != nil {
		return nil, err
	}
	return output, nil
}

func checkS3Gateway(driver driver.Driver, logger logs.TaggedLogger) error {
	return backoff.RetryNotify(func() error {
		jobDomain := ppsutil.SidecarS3GatewayService(driver.PipelineInfo().Pipeline, logger.JobID())
		endpoint := fmt.Sprintf("http://%s:%s/", jobDomain, os.Getenv("S3GATEWAY_PORT"))
		_, err := (&http.Client{Timeout: 5 * time.Second}).Get(endpoint)
		logger.Logf("checking s3 gateway service for job %q: %v", logger.JobID(), err)
		return errors.EnsureStack(err)
	}, backoff.New60sBackOff(), func(err error, d time.Duration) error {
		logger.Logf("worker could not connect to s3 gateway for %q: %v", logger.JobID(), err)
		return nil
	})
	// TODO: `master` implementation fails the job here, we may need to do the same
	// We would need to load the JobInfo first for this:
	// }); err != nil {
	//   reason := fmt.Sprintf("could not connect to s3 gateway for %q: %v", logger.JobID(), err)
	//   logger.Logf("failing job with reason: %s", reason)
	//   // NOTE: this is the only place a worker will reach over and change the job state, this should not generally be done.
	//   return finishJob(driver.PipelineInfo(), driver.PachClient(), jobInfo, pps.JobState_JOB_FAILURE, reason, nil, nil, 0, nil, 0)
	// }
	// return nil
}

func handleDatumSet(driver driver.Driver, logger logs.TaggedLogger, task *DatumSetTask, status *Status) (*anypb.Any, error) {
	pachClient := driver.PachClient()
	var outputFileSetID, metaFileSetID string
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		cacheClient := pfssync.NewCacheClient(pachClient, renewer)
		// Setup file operation client for output meta commit.
		resp, err := cacheClient.WithCreateFileSetClient(func(mfMeta client.ModifyFile) error {
			// Setup file operation client for output PFS commit.
			resp, err := cacheClient.WithCreateFileSetClient(func(mfPFS client.ModifyFile) error {
				di := datum.NewFileSetIterator(pachClient.Ctx(), pachClient.PfsAPIClient, task.FileSetId, task.PathRange)
				opts := []datum.SetOption{
					datum.WithMetaOutput(mfMeta),
					datum.WithPFSOutput(mfPFS),
					datum.WithStats(stats),
				}
				if driver.PipelineInfo().Details.Transform.DatumBatching {
					return handleDatumSetBatching(ctx, driver, logger, task, status, cacheClient, di, opts)
				}
				return forEachDatum(ctx, driver, logger, task, status, cacheClient, di, opts, func(ctx context.Context, logger logs.TaggedLogger, env []string) error {
					return errors.EnsureStack(driver.RunUserCode(ctx, logger, env))
				})
			})
			if err != nil {
				return err
			}
			outputFileSetID = resp.FileSetId
			return renewer.Add(ctx, outputFileSetID)
		})
		if err != nil {
			return err
		}
		metaFileSetID = resp.FileSetId
		return renewer.Add(ctx, metaFileSetID)
	}); err != nil {
		return nil, err
	}
	return serializeDatumSetTaskResult(&DatumSetTaskResult{
		OutputFileSetId: outputFileSetID,
		MetaFileSetId:   metaFileSetID,
		Stats:           stats,
	})
}

func handleDatumSetBatching(ctx context.Context, driver driver.Driver, logger logs.TaggedLogger, task *DatumSetTask, status *Status, cacheClient *pfssync.CacheClient, di datum.Iterator, setOpts []datum.SetOption) error {
	return status.withDatumBatch(func(nextChan <-chan error, setupChan chan<- []string) error {
		// Set up the restart mechanism for the user code.
		var cancel context.CancelFunc
		var errChan chan error
		stop := func() {
			if cancel != nil {
				cancel()
				<-errChan
			}
		}
		defer stop()
		start := func() error {
			var cancelCtx context.Context
			cancelCtx, cancel = pctx.WithCancel(ctx)
			errChan = make(chan error, 1)
			go func() {
				err := driver.RunUserCode(cancelCtx, logger, nil)
				if err == nil {
					err = errors.New("user code exited prematurely")
				}
				errChan <- err
				close(errChan)
			}()
			select {
			case <-nextChan:
				return nil
			case err := <-errChan:
				return errors.Wrap(err, "error running user code")
			case <-ctx.Done():
				return errors.EnsureStack(context.Cause(ctx))
			}
		}
		// Start the user code, then iterate through the datums.
		if err := start(); err != nil {
			return errors.Wrap(err, "error starting user code")
		}
		return forEachDatum(ctx, driver, logger, task, status, cacheClient, di, setOpts, func(ctx context.Context, logger logs.TaggedLogger, env []string) (retErr error) {
			defer func() {
				// Restart the user code if an error occurred.
				if retErr != nil {
					stop()
					errors.Invoke(&retErr, start, "error restarting user code")
				}
			}()
			select {
			case status.setupChan <- env:
			case err := <-errChan:
				return errors.Wrap(err, "error running user code")
			case <-ctx.Done():
				return errors.EnsureStack(context.Cause(ctx))
			}
			select {
			case err := <-nextChan:
				return err
			case err := <-errChan:
				return errors.Wrap(err, "error running user code")
			case <-ctx.Done():
				return errors.EnsureStack(context.Cause(ctx))
			}
		})
	})
}

type datumCallback = func(ctx context.Context, logger logs.TaggedLogger, env []string) error

// TODO: There are way too many parameters here. Consider grouping them into a reasonable struct or storing some of these in the driver.
func forEachDatum(ctx context.Context, driver driver.Driver, baseLogger logs.TaggedLogger, task *DatumSetTask, status *Status, cacheClient *pfssync.CacheClient, di datum.Iterator, setOpts []datum.SetOption, cb datumCallback) error {
	jobInfo, err := driver.GetJobInfo(task.Job)
	if err != nil {
		return errors.Wrapf(err, "load datum set's job info for job %q", task.Job.String())
	}
	// TODO: Can this just be refactored into the datum package such that we don't need to specify a storage root for the sets?
	// The sets would just create a temporary directory under /tmp.
	storageRoot := filepath.Join(driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
	userImageID, err := driver.GetContainerImageID(ctx, "user")
	if err != nil {
		return errors.Wrap(err, "could not get user image ID")
	}
	// Setup datum set for processing.
	return datum.WithSet(cacheClient, storageRoot, func(s *datum.Set) error {
		// Process each datum in the datum set.
		err := di.Iterate(func(meta *datum.Meta) error {
			meta = proto.Clone(meta).(*datum.Meta)
			meta.ImageId = userImageID
			inputs := meta.Inputs
			logger := baseLogger.WithData(inputs)

			env := driver.UserCodeEnv(logger.JobID(), task.OutputCommit, inputs, jobInfo.GetAuthToken())
			opts := []datum.Option{datum.WithEnv(env)}
			if driver.PipelineInfo().Details.DatumTimeout != nil {
				timeout := driver.PipelineInfo().Details.DatumTimeout.AsDuration()
				opts = append(opts, datum.WithTimeout(timeout))
			}
			if driver.PipelineInfo().Details.DatumTries > 0 {
				opts = append(opts, datum.WithRetry(int(driver.PipelineInfo().Details.DatumTries)-1))
			}
			if driver.PipelineInfo().Details.Transform.ErrCmd != nil {
				opts = append(opts, datum.WithRecoveryCallback(func(runCtx context.Context) error {
					return errors.EnsureStack(driver.RunUserErrorHandlingCode(runCtx, logger, env))
				}))
			}
			return s.WithDatum(meta, func(d *datum.Datum) error {
				cancelCtx, cancel := pctx.WithCancel(ctx)
				defer cancel()
				err := status.withDatum(inputs, cancel, func() error {
					err := driver.WithActiveData(inputs, d.PFSStorageRoot(), func() error {
						err := d.Run(cancelCtx, func(runCtx context.Context) error {
							return cb(runCtx, logger, env)
						})
						return errors.EnsureStack(err)
					})
					return errors.EnsureStack(err)
				})
				return errors.EnsureStack(err)
			}, opts...)
		})
		return errors.EnsureStack(err)
	}, setOpts...)
}

func deserializeCreateParallelDatumsTask(taskAny *anypb.Any) (*CreateParallelDatumsTask, error) {
	task := &CreateParallelDatumsTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeCreateParallelDatumsTaskResult(task *CreateParallelDatumsTaskResult) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeCreateSerialDatumsTask(taskAny *anypb.Any) (*CreateSerialDatumsTask, error) {
	task := &CreateSerialDatumsTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeCreateSerialDatumsTaskResult(task *CreateSerialDatumsTaskResult) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeCreateDatumSetsTask(taskAny *anypb.Any) (*CreateDatumSetsTask, error) {
	task := &CreateDatumSetsTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeCreateDatumSetsTaskResult(task *CreateDatumSetsTaskResult) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeDatumSetTask(taskAny *anypb.Any) (*DatumSetTask, error) {
	task := &DatumSetTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeDatumSetTaskResult(task *DatumSetTaskResult) (*anypb.Any, error) {
	return anypb.New(task)
}
