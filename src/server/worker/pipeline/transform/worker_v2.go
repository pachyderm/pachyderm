package transform

import (
	"context"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
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

func handleDatumSetV2(driver driver.Driver, logger logs.TaggedLogger, datumSet *DatumSetV2, status *Status) error {
	pachClient := driver.PachClient()
	storageRoot := filepath.Join(driver.InputDir(), client.PPSScratchSpace, uuid.NewWithoutDashes())
	// TODO: Unify these stats with datum.Stats
	processStats := &pps.ProcessStats{}
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
					inputs := inputV2ToV1(meta.Inputs)
					env := driver.UserCodeEnv(logger.JobID(), outputCommit, inputs)
					opts := []datum.DatumOption{}
					if driver.PipelineInfo().Transform.ErrCmd != nil {
						opts = append(opts, datum.WithRecoveryCallback(func() error {
							// TODO: Datum timeout should be an functional option on a datum.
							return driver.RunUserErrorHandlingCode(logger, env, processStats, driver.PipelineInfo().DatumTimeout)
						}))
					}
					return s.WithDatum(ctx, meta, func(d *datum.Datum) error {
						return driver.WithActiveData(inputs, d.PFSStorageRoot(), func() error {
							return status.withDatum(inputs, cancel, func() error {
								return driver.RunUserCode(logger, env, processStats, driver.PipelineInfo().DatumTimeout)
							})
						})
					}, opts...)
				})
			}, datum.WithMetaOutput(focMeta), datum.WithPFSOutput(focPFS))
		})
	})
}

// TODO: temporary hack to workaround refactoring WithActiveData
func inputV2ToV1(inputs []*common.InputV2) []*common.Input {
	var result []*common.Input
	for _, input := range inputs {
		result = append(result, &common.Input{
			ParentCommit: input.ParentCommit,
			Name:         input.Name,
			JoinOn:       input.JoinOn,
			Lazy:         input.Lazy,
			Branch:       input.Branch,
			GitURL:       input.GitURL,
			EmptyFiles:   input.EmptyFiles,
			S3:           input.S3,
		})
	}
	return result
}
