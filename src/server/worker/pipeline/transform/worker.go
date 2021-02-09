package transform

import (
	"context"
	"path/filepath"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
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
// s3 input / gateway stuff (need more information here).
// spouts.
// joins.
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
			return handleDatumSet(driver, logger, datumSet, status)
		}); err != nil {
			return err
		}
		subtask.Data, err = serializeDatumSet(datumSet)
		return err
	})
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
				di := datum.NewFileSetIterator(pachClient, client.TmpRepoName, datumSet.FileSet)
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
