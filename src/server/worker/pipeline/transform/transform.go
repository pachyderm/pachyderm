package transform

import (
	"github.com/pachyderm/pachyderm/v2/src/pps"
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

	return driver.PachClient().SubscribePipelineJob(
		driver.PipelineInfo().Pipeline.Name,
		false,
		func(pipelineJobInfo *pps.PipelineJobInfo) error {
			// TODO: check that this is for the right version of the pipeline
			return reg.startPipelineJob(pipelineJobInfo)
		},
	)
}
