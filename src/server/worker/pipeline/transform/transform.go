package transform

import (
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Run will run a transform pipeline until the driver is canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	reg, err := newRegistry(driver, logger)
	if err != nil {
		return err
	}
	logger.Logf("transform spawner started")
	return forEachCommit(driver, reg.startPipelineJob)
}

func forEachCommit(driver driver.Driver, cb func(*pfs.CommitInfo) error) error {
	pachClient := driver.PachClient()
	pipelineInfo := driver.PipelineInfo()
	return pachClient.SubscribeCommit(
		client.NewRepo(pipelineInfo.Pipeline.Name),
		"",
		pipelineInfo.SpecCommit.NewProvenance(),
		"",
		pfs.CommitState_READY,
		func(commitInfo *pfs.CommitInfo) error {
			err := cb(commitInfo)
			// TODO: Figure out how to clean up jobs after deleted commit. Just cleaning up here is not a good solution because
			// we are not guaranteed to hit this code path after a deletion.
			if pfsserver.IsCommitFinishedErr(err) || pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) || ppsserver.IsPipelineJobFinishedErr(err) {
				return nil
			}
			return err
		},
	)
}
