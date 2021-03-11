package spout

import (
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Run will run a spout pipeline until the driver is canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	logger = logger.WithJob("spout")
	return driver.RunUserCode(driver.PachClient().Ctx(), logger, nil)
}
