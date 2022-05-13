package transform

import (
	"bytes"
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
			case types.Is(input, &UploadDatumsTask{}):
				uploadDatumsTask, err := deserializeUploadDatumsTask(input)
				if err != nil {
					return nil, err
				}
				pachClient := driver.PachClient().WithCtx(ctx)
				return processUploadDatumsTask(pachClient, uploadDatumsTask)
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

func processUploadDatumsTask(pachClient *client.APIClient, task *UploadDatumsTask) (*types.Any, error) {
	jobInfo, err := pachClient.InspectJob(task.Job.Pipeline.Name, task.Job.ID, true)
	if err != nil {
		return nil, err
	}
	metaCommitInfo, err := pachClient.PfsAPIClient.InspectCommit(
		pachClient.Ctx(),
		&pfs.InspectCommitRequest{
			Commit: ppsutil.MetaCommit(jobInfo.OutputCommit),
		})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	var dit datum.Iterator
	if metaCommitInfo.Finishing != nil {
		dit = datum.NewCommitIterator(pachClient, metaCommitInfo.Commit)
	} else {
		dit, err = datum.NewIterator(pachClient, jobInfo.Details.Input)
		if err != nil {
			return nil, err
		}
		dit = datum.NewJobIterator(dit, jobInfo.Job, &hasher{salt: jobInfo.Details.Salt})
	}
	fileSetID, count, err := uploadDatumFileSet(pachClient, dit)
	if err != nil {
		return nil, err
	}
	return serializeUploadDatumsTaskResult(&UploadDatumsTaskResult{FileSetId: fileSetID, Count: int64(count)})
}

func uploadDatumFileSet(pachClient *client.APIClient, dit datum.Iterator) (string, int, error) {
	var count int
	s, err := withDatumFileSet(pachClient, func(s *datum.Set) error {
		return errors.EnsureStack(dit.Iterate(func(meta *datum.Meta) error {
			count++
			return s.UploadMeta(meta)
		}))
	})
	if err != nil {
		return "", 0, err
	}
	return s, count, nil
}

func withDatumFileSet(pachClient *client.APIClient, cb func(*datum.Set) error) (string, error) {
	resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
		storageRoot := filepath.Join(os.TempDir(), "pachyderm-skipped-tmp", uuid.NewWithoutDashes())
		return datum.WithSet(nil, storageRoot, cb, datum.WithMetaOutput(mf))
	})
	if err != nil {
		return "", err
	}
	return resp.FileSetId, nil
}

func processComputeParallelDatumsTask(pachClient *client.APIClient, task *ComputeParallelDatumsTask) (*types.Any, error) {
	var dits []datum.Iterator
	if task.BaseFileSetId != "" {
		dits = append(dits, datum.NewFileSetIterator(pachClient, task.BaseFileSetId))
	}
	dits = append(dits, datum.NewFileSetIterator(pachClient, task.FileSetId))
	outputFileSetID, err := withDatumFileSet(pachClient, func(outputSet *datum.Set) error {
		return datum.Merge(dits, func(metas []*datum.Meta) error {
			if len(metas) > 1 || !proto.Equal(metas[0].Job, task.Job) {
				return nil
			}
			return outputSet.UploadMeta(metas[0], datum.WithPrefixIndex())
		})
	})
	if err != nil {
		return nil, err
	}
	return serializeComputeParallelDatumsTaskResult(&ComputeParallelDatumsTaskResult{FileSetId: outputFileSetID})
}

func processComputeSerialDatumsTask(pachClient *client.APIClient, task *ComputeSerialDatumsTask) (*types.Any, error) {
	dits := []datum.Iterator{
		datum.NewCommitIterator(pachClient, task.BaseMetaCommit),
		datum.NewFileSetIterator(pachClient, task.FileSetId),
	}
	var deleteFileSetID string
	var skipped int64
	outputFileSetID, err := withDatumFileSet(pachClient, func(outputSet *datum.Set) error {
		var err error
		deleteFileSetID, err = withDatumFileSet(pachClient, func(deleteSet *datum.Set) error {
			return datum.Merge(dits, func(metas []*datum.Meta) error {
				if len(metas) == 1 {
					// Datum was processed in the parallel step.
					if proto.Equal(metas[0].Job, task.Job) {
						return nil
					}
					// Datum only exists in the parent job.
					return deleteSet.UploadMeta(metas[0])
				}
				// Check if a skippable datum was successfully processed by the parent.
				if !task.NoSkip && skippableDatum(metas[1], metas[0]) {
					skipped++
					return nil
				}
				if err := deleteSet.UploadMeta(metas[0]); err != nil {
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
		FileSetId:       outputFileSetID,
		DeleteFileSetId: deleteFileSetID,
		Skipped:         skipped,
	})
}

func skippableDatum(meta1, meta2 *datum.Meta) bool {
	// If the hashes are equal and the second datum was processed, then skip it.
	return meta1.Hash == meta2.Hash && meta2.State == datum.State_PROCESSED
}

func processCreateDatumSetsTask(driver driver.Driver, task *CreateDatumSetsTask) (*types.Any, error) {
	setSpec, err := createSetSpec(driver, task.FileSetId)
	if err != nil {
		return nil, err
	}
	pachClient := driver.PachClient()
	var inputFileSetsId string
	resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
		return pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
			pachClient := pachClient.WithCtx(ctx)
			dit := datum.NewFileSetIterator(pachClient, task.FileSetId)
			storageRoot := filepath.Join(driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
			var count int64
			if err := datum.CreateSets(dit, storageRoot, setSpec, func(upload func(client.ModifyFile) error) error {
				resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
					return upload(mf)
				})
				if err != nil {
					return err
				}
				if err := renewer.Add(pachClient.Ctx(), resp.FileSetId); err != nil {
					return err
				}
				name := fmt.Sprintf("%016d", count)
				count++
				input, err := serializeDatumSet(&DatumSet{
					JobID:        task.Job.ID,
					OutputCommit: task.OutputCommit,
					FileSetId:    resp.FileSetId,
				})
				if err != nil {
					return err
				}
				data, err := proto.Marshal(input)
				if err != nil {
					return errors.EnsureStack(err)
				}
				return errors.EnsureStack(mf.PutFile(name, bytes.NewReader(data)))
			}); err != nil {
				return err
			}
			var err error
			inputFileSetsId, err = renewer.Compose(ctx)
			return err
		})
	})
	if err != nil {
		return nil, err
	}
	return serializeCreateDatumSetsTaskResult(&CreateDatumSetsTaskResult{
		FileSetId:       resp.FileSetId,
		InputFileSetsId: inputFileSetsId,
	})
}

func createSetSpec(driver driver.Driver, fileSetID string) (*datum.SetSpec, error) {
	pachClient := driver.PachClient()
	dit := datum.NewFileSetIterator(pachClient, fileSetID)
	var numDatums int64
	if err := dit.Iterate(func(_ *datum.Meta) error {
		numDatums++
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	// When the datum set spec is not set, evenly distribute the datums.
	var setSpec *datum.SetSpec
	datumSetsPerWorker := defaultDatumSetsPerWorker
	if driver.PipelineInfo().Details.DatumSetSpec != nil {
		setSpec = &datum.SetSpec{
			Number:    driver.PipelineInfo().Details.DatumSetSpec.Number,
			SizeBytes: driver.PipelineInfo().Details.DatumSetSpec.SizeBytes,
		}
		datumSetsPerWorker = driver.PipelineInfo().Details.DatumSetSpec.PerWorker
	}
	if setSpec == nil || (setSpec.Number == 0 && setSpec.SizeBytes == 0) {
		concurrency, err := driver.ExpectedNumWorkers()
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		setSpec = &datum.SetSpec{Number: numDatums / (int64(concurrency) * datumSetsPerWorker)}
		if setSpec.Number == 0 {
			setSpec.Number = 1
		}
	}
	return setSpec, nil
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
	// Setup file operation client for output meta commit.
	resp, err := pachClient.WithCreateFileSetClient(func(mfMeta client.ModifyFile) error {
		// Setup file operation client for output PFS commit.
		resp, err := pachClient.WithCreateFileSetClient(func(mfPFS client.ModifyFile) (retErr error) {
			opts := []datum.SetOption{
				datum.WithMetaOutput(mfMeta),
				datum.WithPFSOutput(mfPFS),
				datum.WithStats(datumSet.Stats),
			}
			return pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
				pachClient := pachClient.WithCtx(ctx)
				cacheClient := pfssync.NewCacheClient(pachClient, renewer)
				// Setup datum set for processing.
				return datum.WithSet(cacheClient, storageRoot, func(s *datum.Set) error {
					di := datum.NewFileSetIterator(pachClient, datumSet.FileSetId)
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
		})
		if err != nil {
			return err
		}
		datumSet.OutputFileSetId = resp.FileSetId
		return nil
	})
	if err != nil {
		return err
	}
	datumSet.MetaFileSetId = resp.FileSetId
	return nil
}

func deserializeUploadDatumsTask(taskAny *types.Any) (*UploadDatumsTask, error) {
	task := &UploadDatumsTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeUploadDatumsTaskResult(task *UploadDatumsTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
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
