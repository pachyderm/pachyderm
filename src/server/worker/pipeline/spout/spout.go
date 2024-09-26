// Package spout needs to be documented.
//
// TODO: document
package spout

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Run will run a spout pipeline until the driver is canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	logger = logger.WithJob("spout")
	return errors.EnsureStack(driver.RunUserCode(driver.PachClient().Ctx(), logger, nil))
}
