package spout

import (
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/src/server/worker/pipeline"
)

// Run will run a spout pipeline until the driver is canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	pachClient := driver.PachClient()
	pipelineInfo := driver.PipelineInfo()
	logger = logger.WithJob("spout")

	// Spouts typically have an open commit waiting for new data. So if the spout needs to be updated, and
	// thus spoutSpawner is called, it might hang if the commit never gets closed. So to avoid this, we
	// delete open commits that we see here.
	// We probably only need to check the first commit, but doing 10 to be safe
	pachClient.ListCommitF(pipelineInfo.Pipeline.Name, "", "", 10, false, func(c *pfs.CommitInfo) error {
		if c.Finished != nil {
			return nil
		}
		return pachClient.DeleteCommit(pipelineInfo.Pipeline.Name, c.Commit.ID)
	})

	// TODO: do something with stats?
	_, err := driver.WithData(nil, nil, logger, func(dir string, stats *pps.ProcessStats) error {
		inputs := []*common.Input{} // Spouts take no inputs
		return driver.WithActiveData(inputs, dir, func() error {
			eg, serviceCtx := errgroup.WithContext(pachClient.Ctx())

			// While spouts do write to output commits, the output commit changes
			// frequently and we do not restart the user code for each one. Therefore,
			// we leave the output commit out of the user code env.
			eg.Go(func() error { return pipeline.RunUserCode(driver.WithContext(serviceCtx), logger, nil, inputs) })
			eg.Go(func() error { return pipeline.ReceiveSpout(serviceCtx, pachClient, pipelineInfo, logger) })
			return eg.Wait()
		})
	})
	return err
}
