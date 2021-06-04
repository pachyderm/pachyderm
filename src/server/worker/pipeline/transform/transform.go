package transform

import (
	"github.com/gogo/protobuf/proto"

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

	return driver.PachClient().SubscribeJob(
		driver.PipelineInfo().Pipeline.Name,
		true,
		func(jobInfo *pps.JobInfo) error {
			if jobInfo.PipelineVersion != driver.PipelineInfo().Version {
				// Skip this job - we should be shut down soon, but don't error out in the meantime
				return nil
			}
			return reg.startJob(proto.Clone(jobInfo).(*pps.JobInfo))
		},
	)
}
