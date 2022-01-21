package transform

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfssync"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Worker handles a transform pipeline work subtask, then returns.
// TODO:
// datum queuing (probably should be handled by datum package).
// capture datum logs.
// git inputs.
func Worker(driver driver.Driver, logger logs.TaggedLogger, input *types.Any, status *Status) (*types.Any, error) {
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
