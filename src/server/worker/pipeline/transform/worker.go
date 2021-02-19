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
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/work"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Worker handles a transform pipeline work subtask, then returns.
// TODO:
// datum queuing (probably should be handled by datum package).
// spouts.
// capture datum logs.
// git inputs.
// handle custom user set for execution.
func Worker(driver driver.Driver, logger logs.TaggedLogger, subtask *work.Task, status *Status) (retErr error) {
	datumSet, err := deserializeDatumSet(subtask.Data)
	if err != nil {
		return err
	}
	return status.withJob(datumSet.JobID, func() error {
		logger = logger.WithJob(datumSet.JobID)
		if err := logger.LogStep("datum task", func() error {
			if ppsutil.ContainsS3Inputs(driver.PipelineInfo().Input) || driver.PipelineInfo().S3Out {
				if err := checkS3Gateway(driver, logger); err != nil {
					return err
				}
			}
			return handleDatumSet(driver, logger, datumSet, status)
		}); err != nil {
			return err
		}
		subtask.Data, err = serializeDatumSet(datumSet)
		return err
	})
}

func checkS3Gateway(driver driver.Driver, logger logs.TaggedLogger) error {
	return backoff.RetryNotify(func() error {
		endpoint := fmt.Sprintf("http://%s:%s/", ppsutil.SidecarS3GatewayService(logger.JobID()), os.Getenv("S3GATEWAY_PORT"))
		_, err := (&http.Client{Timeout: 5 * time.Second}).Get(endpoint)
		logger.Logf("checking s3 gateway service for job %q: %v", logger.JobID(), err)
		return err
	}, backoff.New60sBackOff(), func(err error, d time.Duration) error {
		logger.Logf("worker could not connect to s3 gateway for %q: %v", logger.JobID(), err)
		return nil
	})
	// TODO: `master` implementation fails the job here, we may need to do the same
	// We would need to load the jobInfo first for this:
	// }); err != nil {
	//   reason := fmt.Sprintf("could not connect to s3 gateway for %q: %v", logger.JobID(), err)
	//   logger.Logf("failing job with reason: %s", reason)
	//   // NOTE: this is the only place a worker will reach over and change the job state, this should not generally be done.
	//   return finishJob(driver.PipelineInfo(), driver.PachClient(), jobInfo, pps.JobState_JOB_FAILURE, reason, nil, nil, 0, nil, 0)
	// }
	// return nil
}

// TODO: It would probably be better to write the output to temporary file sets and expose an operation through pfs for adding a temporary fileset to a commit.
func handleDatumSet(driver driver.Driver, logger logs.TaggedLogger, datumSet *DatumSet, status *Status) error {
	pachClient := driver.PachClient()
	storageRoot := filepath.Join(driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
	datumSet.Stats = &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	// Setup file operation client for output meta commit.
	metaCommit := datumSet.MetaCommit
	return pachClient.WithModifyFileClient(metaCommit.Repo.Name, metaCommit.ID, func(mfcMeta *client.ModifyFileClient) error {
		// Setup file operation client for output PFS commit.
		outputCommit := datumSet.OutputCommit
		return pachClient.WithModifyFileClient(outputCommit.Repo.Name, outputCommit.ID, func(mfcPFS *client.ModifyFileClient) (retErr error) {
			opts := []datum.SetOption{
				datum.WithMetaOutput(newDatumClient(mfcMeta, pachClient, metaCommit)),
				datum.WithPFSOutput(newDatumClient(mfcPFS, pachClient, outputCommit)),
				datum.WithStats(datumSet.Stats),
			}
			// Setup datum set for processing.
			return datum.WithSet(pachClient, storageRoot, func(s *datum.Set) error {
				di := datum.NewFileSetIterator(pachClient, datumSet.FileSet)
				// Process each datum in the assigned datum set.
				return di.Iterate(func(meta *datum.Meta) error {
					ctx := pachClient.Ctx()
					inputs := meta.Inputs
					logger = logger.WithData(inputs)
					env := driver.UserCodeEnv(logger.JobID(), outputCommit, inputs)
					var opts []datum.Option
					if driver.PipelineInfo().DatumTimeout != nil {
						timeout, err := types.DurationFromProto(driver.PipelineInfo().DatumTimeout)
						if err != nil {
							return err
						}
						opts = append(opts, datum.WithTimeout(timeout))
					}
					if driver.PipelineInfo().Transform.ErrCmd != nil {
						opts = append(opts, datum.WithRecoveryCallback(func(runCtx context.Context) error {
							return driver.RunUserErrorHandlingCode(runCtx, logger, env)
						}))
					}
					return s.WithDatum(ctx, meta, func(d *datum.Datum) error {
						cancelCtx, cancel := context.WithCancel(ctx)
						defer cancel()
						return status.withDatum(inputs, cancel, func() error {
							return driver.WithActiveData(inputs, d.PFSStorageRoot(), func() error {
								return d.Run(cancelCtx, func(runCtx context.Context) error {
									return driver.RunUserCode(runCtx, logger, env)
								})
							})
						})
					}, opts...)

				})
			}, opts...)
		})
	})
}

// TODO: This should be removed when CopyFile is a part of ModifyFile.
type datumClient struct {
	*client.ModifyFileClient
	pachClient *client.APIClient
	commit     *pfs.Commit
}

func newDatumClient(mfc *client.ModifyFileClient, pachClient *client.APIClient, commit *pfs.Commit) datum.Client {
	return &datumClient{
		ModifyFileClient: mfc,
		pachClient:       pachClient,
		commit:           commit,
	}
}

func (dc *datumClient) CopyFile(dst string, srcFile *pfs.File, tag string) error {
	return dc.pachClient.CopyFile(srcFile.Commit.Repo.Name, srcFile.Commit.ID, srcFile.Path, dc.commit.Repo.Name, dc.commit.ID, dst, false, tag)
}

type datumClientFileset struct {
	*client.CreateFilesetClient
}

func newDatumClientFileset(cfc *client.CreateFilesetClient) datum.Client {
	return &datumClientFileset{
		CreateFilesetClient: cfc,
	}
}

func (dcf *datumClientFileset) CopyFile(_ string, _ *pfs.File, _ string) error {
	panic("attempted copy file in fileset datum client")
}
