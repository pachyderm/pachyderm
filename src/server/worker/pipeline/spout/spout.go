package spout

import (
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Run will run a spout pipeline until the driver is canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	pachClient := driver.PachClient()
	pipelineInfo := driver.PipelineInfo()

	// Spouts typically have an open commit waiting for new data. So if the spout needs to be updated, and
	// thus spoutSpawner is called, it might hang if the commit never gets closed. So to avoid this, we
	// delete open commits that we see here.
	// We probably only need to check the first commit, but doing 10 to be safe
	// TODO: This seems like something the user should handle, rather than us. I don't think we want to impose
	// semantics on the output repo across runs of the spout code.
	pachClient.ListCommitF(pipelineInfo.Pipeline.Name, "", "", 10, false, func(ci *pfs.CommitInfo) error {
		if ci.Finished != nil {
			return nil
		}
		return pachClient.SquashCommit(pipelineInfo.Pipeline.Name, ci.Commit.ID)
	})

	logger = logger.WithJob("spout")
	return driver.RunUserCode(pachClient.Ctx(), logger, nil)
}
