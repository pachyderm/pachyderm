package transform

import (
	"context"
	"path/filepath"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

// TODO:
// datum queuing (probably should be handled by datum package).
// s3 input / gateway stuff (need more information here).
// spouts.
// joins.
// capture datum logs.
// file download features (empty / lazy files). Need to check over the pipe logic.
// git inputs.
// datum specific stats (need to refactor datum stats into datum package and figure out prometheus stuff).
// handle custom user set for execution.
// Taking advantage of symlinks during upload?
func WorkerV2(driver driver.Driver, logger logs.TaggedLogger, subtask *work.Task, status *Status) (retErr error) {
	datumSet, err := deserializeDatumSetV2(subtask.Data)
	if err != nil {
		return err
	}
	return status.withJob(datumSet.JobID, func() error {
		logger = logger.WithJob(datumSet.JobID)
		if err := logger.LogStep("datum task", func() error {
			return handleDatumSetV2(driver, logger, datumSet, status)
		}); err != nil {
			return err
		}
		subtask.Data, err = serializeDatumSetV2(datumSet)
		return err
	})
}

// TODO: It would probably be better to write the output to temporary file sets and expose an operation through pfs for adding a temporary fileset to a commit.
func handleDatumSetV2(driver driver.Driver, logger logs.TaggedLogger, datumSet *DatumSetV2, status *Status) error {
	pachClient := driver.PachClient()
	storageRoot := filepath.Join(driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
	datumSet.Stats = &datum.Stats{ProcessStats: &pps.ProcessStats{}}
	// Setup file operation client for output meta commit.
	metaCommit := datumSet.MetaCommit
	return pachClient.WithFileOperationClientV2(metaCommit.Repo.Name, metaCommit.ID, func(focMeta *client.FileOperationClient) error {
		// Setup file operation client for output PFS commit.
		outputCommit := datumSet.OutputCommit
		return pachClient.WithFileOperationClientV2(outputCommit.Repo.Name, outputCommit.ID, func(focPFS *client.FileOperationClient) error {
			// Setup datum set for processing.
			return datum.WithSet(pachClient, storageRoot, func(s *datum.Set) error {
				di := datum.NewFileSetIterator(pachClient, tmpRepo, datumSet.FileSet)
				// Process each datum in the assigned datum set.
				return di.Iterate(func(meta *datum.Meta) error {
					ctx, cancel := context.WithCancel(pachClient.Ctx())
					defer cancel()
					driver := driver.WithContext(ctx)
					inputs := meta.Inputs
					env := driver.UserCodeEnv(logger.JobID(), outputCommit, inputs)
					var opts []datum.DatumOption
					if driver.PipelineInfo().DatumTimeout != nil {
						timeout, err := types.DurationFromProto(driver.PipelineInfo().DatumTimeout)
						if err != nil {
							return err
						}
						opts = append(opts, datum.WithTimeout(timeout))
					}
					if driver.PipelineInfo().Transform.ErrCmd != nil {
						opts = append(opts, datum.WithRecoveryCallback(func(runCtx context.Context) error {
							return driver.RunUserErrorHandlingCodeV2(runCtx, logger, env)
						}))
					}
					return s.WithDatum(ctx, meta, func(d *datum.Datum) error {
						return status.withDatum(inputs, cancel, func() error {
							return driver.WithActiveData(inputs, d.PFSStorageRoot(), func() error {
								return d.Run(ctx, func(runCtx context.Context) error {
									return driver.RunUserCodeV2(runCtx, logger, env)
								})
							})
						})
					}, opts...)

				})
			}, datum.WithMetaOutput(focMeta), datum.WithPFSOutput(focPFS), datum.WithStats(datumSet.Stats))
		})
	})
}
