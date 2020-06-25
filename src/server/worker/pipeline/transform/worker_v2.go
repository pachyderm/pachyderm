package transform

import (
	"context"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

func WorkerV2(driver driver.Driver, logger logs.TaggedLogger, subtask *work.Task, status *Status) (retErr error) {
	datumData, err := deserializeDatumDataV2(subtask.Data)
	if err != nil {
		return err
	}
	return status.withJob(datumData.JobID, func() error {
		logger = logger.WithJob(datumData.JobID)
		if err := logger.LogStep("datum task", func() error {
			return handleDatumTaskV2(driver, logger, datumData, subtask.ID, status)
		}); err != nil {
			return err
		}
		subtask.Data, err = serializeDatumData(datumData)
		return err
	})
}

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
func handleDatumTaskV2(driver driver.Driver, logger logs.TaggedLogger, datumData *DatumDataV2, subtaskID string, status *Status) error {
	storageRoot := filepath.Join(d.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
	h := &hasher{
		name: driver.PipelineInfo().Pipeline.Name,
		salt: driver.PipelineInfo().Salt,
	}
	pachClient := driver.PachClient()
	outputCommit := datumData.OutputCommit
	// Setup file operation client for output commit.
	return pachClient.WithFileOperationClientV2(outputCommit.Repo.Name, outputCommit.ID, func(foc *client.FileOperationClient) error {
		statsCommit := datumData.StatsCommit
		// Setup file operation client for stats commit.
		return pachClient.WithFileOperationClientV2(statsCommit.Repo.Name, statsCommit.ID, func(focStats *client.FileOperationClient) error {
			// Setup datum set for processing.
			return datum.WithSet(pachClient, storageRoot, h, func(s *Set) error {
				// TODO: Figure out temporary filesets in pfs.
				di := datum.NewFileSetIterator(pachClient, "/tmp", data.DatumFileSet)
				// Process each datum in the assigned datum set.
				return di.Iterate(func(inputs []*common.Input) error {
					ctx, cancel := context.WithCancel(pachClient.Ctx())
					defer cancel()
					driver := driver.WithContext(ctx)
					env := userCodeEnv(driver, logger.JobID(), outputCommit, inputs)
					opts := []datum.DatumOption{}
					if driver.PipelineInfo().Transform.ErrCmd != nil {
						opts = append(opts, datum.WithRecoveryCallback(func() error {
							return driver.RunUserErrorHandlingCode(logger, env, processStats, driver.PipelineInfo().DatumTimeout)
						}))
					}
					return s.WithDatum(inputs, func(datum *Datum) error {
						return driver.WithActiveData(inputs, d.StorageRoot(), func() error {
							return status.withDatum(inputs, cancel, func() error {
								return driver.RunUserCode(logger, env, processStats, driver.PipelineInfo().DatumTimeout)
							})
						})
					}, opts...)
				})
			}, datum.WithPFSOutput(foc), datum.WithMetaOutput(focStats))
		})
	})
}
