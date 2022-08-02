package transform

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfssync"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

type hasher struct {
	salt string
}

func (h *hasher) Hash(inputs []*common.Input) string {
	return common.HashDatum(h.salt, inputs)
}

func Worker(ctx context.Context, driver driver.Driver, logger logs.TaggedLogger, status *Status) error {
	return errors.EnsureStack(driver.NewTaskSource().Iterate(
		ctx,
		func(ctx context.Context, input *types.Any) (*types.Any, error) {
			switch {
			case datum.IsTask(input):
				pachClient := driver.PachClient().WithCtx(ctx)
				return datum.ProcessTask(pachClient, input)
			case types.Is(input, &ComputeParallelDatumsTask{}):
				computeParallelDatumsTask, err := deserializeComputeParallelDatumsTask(input)
				if err != nil {
					return nil, err
				}
				pachClient := driver.PachClient().WithCtx(ctx)
				return processComputeParallelDatumsTask(pachClient, computeParallelDatumsTask)
			case types.Is(input, &ComputeSerialDatumsTask{}):
				computeSerialDatumsTask, err := deserializeComputeSerialDatumsTask(input)
				if err != nil {
					return nil, err
				}
				pachClient := driver.PachClient().WithCtx(ctx)
				return processComputeSerialDatumsTask(pachClient, computeSerialDatumsTask)
			case types.Is(input, &CreateDatumSetsTask{}):
				createDatumSetsTask, err := deserializeCreateDatumSetsTask(input)
				if err != nil {
					return nil, err
				}
				driver := driver.WithContext(ctx)
				return processCreateDatumSetsTask(driver, createDatumSetsTask)
			case types.Is(input, &DatumSet{}):
				driver := driver.WithContext(ctx)
				return processDatumSet(driver, logger, input, status)
			default:
				return nil, errors.Errorf("unrecognized any type (%v) in transform worker", input.TypeUrl)
			}
		},
	))
}

func processComputeParallelDatumsTask(pachClient *client.APIClient, task *ComputeParallelDatumsTask) (*types.Any, error) {
	var dits []datum.Iterator
	if task.BaseFileSetId != "" {
		dits = append(dits, datum.NewFileSetIterator(pachClient, task.BaseFileSetId, task.PathRange))
	}
	dit := datum.NewFileSetIterator(pachClient, task.FileSetId, task.PathRange)
	dit = datum.NewJobIterator(dit, task.Job, &hasher{salt: task.Salt})
	dits = append(dits, dit)
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	outputFileSetID, err := datum.WithCreateFileSet(pachClient, "pachyderm-datums-compute-parallel", func(outputSet *datum.Set) error {
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
	return serializeComputeParallelDatumsTaskResult(&ComputeParallelDatumsTaskResult{
		FileSetId: outputFileSetID,
		Stats:     stats,
	})
}

func processComputeSerialDatumsTask(pachClient *client.APIClient, task *ComputeSerialDatumsTask) (*types.Any, error) {
	dit := datum.NewFileSetIterator(pachClient, task.FileSetId, task.PathRange)
	dit = datum.NewJobIterator(dit, task.Job, &hasher{salt: task.Salt})
	dits := []datum.Iterator{
		datum.NewCommitIterator(pachClient, task.BaseMetaCommit, task.PathRange),
		dit,
	}
	var metaDeleteFileSetID, outputDeleteFileSetID string
	stats := &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	outputFileSetID, err := datum.WithCreateFileSet(pachClient, "pachyderm-datums-compute-serial", func(outputSet *datum.Set) error {
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
	return serializeComputeSerialDatumsTaskResult(&ComputeSerialDatumsTaskResult{
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

func processCreateDatumSetsTask(driver driver.Driver, task *CreateDatumSetsTask) (*types.Any, error) {
	datumSets, err := datum.CreateSets(driver.PachClient(), task.SetSpec, task.FileSetId, task.PathRange)
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
func processDatumSet(driver driver.Driver, logger logs.TaggedLogger, input *types.Any, status *Status) (*types.Any, error) {
	datumSet, err := deserializeDatumSet(input)
	if err != nil {
		return nil, err
	}
	var output *types.Any
	if err := status.withJob(datumSet.JobID, func() error {
		logger = logger.WithJob(datumSet.JobID)
		if err := logger.LogStep("datum task", func() error {
			if ppsutil.ContainsS3Inputs(driver.PipelineInfo().Details.Input) || driver.PipelineInfo().Details.S3Out {
				if err := checkS3Gateway(driver, logger); err != nil {
					return err
				}
			}
			return handleDatumSet(driver, logger, datumSet, status)
		}); err != nil {
			return errors.EnsureStack(err)
		}
		output, err = serializeDatumSet(datumSet)
		return err
	}); err != nil {
		return nil, err
	}
	return output, nil
}

func checkS3Gateway(driver driver.Driver, logger logs.TaggedLogger) error {
	return backoff.RetryNotify(func() error {
		jobDomain := ppsutil.SidecarS3GatewayService(driver.PipelineInfo().Pipeline.Name, logger.JobID())
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

func handleDatumSet(driver driver.Driver, logger logs.TaggedLogger, datumSet *DatumSet, status *Status) error {
	pachClient := driver.PachClient()
	// TODO: Can this just be refactored into the datum package such that we don't need to specify a storage root for the sets?
	// The sets would just create a temporary directory under /tmp.
	storageRoot := filepath.Join(driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
	datumSet.Stats = &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	userImageID, err := driver.GetContainerImageID(pachClient.Ctx(), "user")
	if err != nil {
		return errors.Wrap(err, "could not get user image ID")
	}
	return pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		// Setup file operation client for output meta commit.
		resp, err := pachClient.WithCreateFileSetClient(func(mfMeta client.ModifyFile) error {
			// Setup file operation client for output PFS commit.
			resp, err := pachClient.WithCreateFileSetClient(func(mfPFS client.ModifyFile) (retErr error) {
				opts := []datum.SetOption{
					datum.WithMetaOutput(mfMeta),
					datum.WithPFSOutput(mfPFS),
					datum.WithStats(datumSet.Stats),
				}
				pachClient := pachClient.WithCtx(ctx)
				cacheClient := pfssync.NewCacheClient(pachClient, renewer)
				// Setup datum set for processing.
				return datum.WithSet(cacheClient, storageRoot, func(s *datum.Set) error {
					di := datum.NewFileSetIterator(pachClient, datumSet.FileSetId, datumSet.PathRange)
					// Process each datum in the assigned datum set.
					err := di.Iterate(func(meta *datum.Meta) error {
						ctx := pachClient.Ctx()
						meta = proto.Clone(meta).(*datum.Meta)
						meta.ImageId = userImageID
						inputs := meta.Inputs
						logger = logger.WithData(inputs)
						env := driver.UserCodeEnv(logger.JobID(), datumSet.OutputCommit, inputs)
						var opts []datum.Option
						if driver.PipelineInfo().Details.DatumTimeout != nil {
							timeout, err := types.DurationFromProto(driver.PipelineInfo().Details.DatumTimeout)
							if err != nil {
								return errors.EnsureStack(err)
							}
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
							cancelCtx, cancel := context.WithCancel(ctx)
							defer cancel()
							err := status.withDatum(inputs, cancel, func() error {
								err := driver.WithActiveData(inputs, d.PFSStorageRoot(), func() error {
									err := d.Run(cancelCtx, func(runCtx context.Context) error {
										return errors.EnsureStack(driver.RunUserCode(runCtx, logger, env))
									})
									return errors.EnsureStack(err)
								})
								return errors.EnsureStack(err)
							})
							return errors.EnsureStack(err)
						}, opts...)
					})
					return errors.EnsureStack(err)
				}, opts...)
			})
			if err != nil {
				return err
			}
			datumSet.OutputFileSetId = resp.FileSetId
			return renewer.Add(ctx, datumSet.OutputFileSetId)
		})
		if err != nil {
			return err
		}
		datumSet.MetaFileSetId = resp.FileSetId
		return renewer.Add(ctx, datumSet.MetaFileSetId)
	})
}

func deserializeComputeParallelDatumsTask(taskAny *types.Any) (*ComputeParallelDatumsTask, error) {
	task := &ComputeParallelDatumsTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeComputeParallelDatumsTaskResult(task *ComputeParallelDatumsTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeComputeSerialDatumsTask(taskAny *types.Any) (*ComputeSerialDatumsTask, error) {
	task := &ComputeSerialDatumsTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeComputeSerialDatumsTaskResult(task *ComputeSerialDatumsTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeCreateDatumSetsTask(taskAny *types.Any) (*CreateDatumSetsTask, error) {
	task := &CreateDatumSetsTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeCreateDatumSetsTaskResult(task *CreateDatumSetsTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}
