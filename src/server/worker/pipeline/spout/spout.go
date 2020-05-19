package spout

import (
	"golang.org/x/sync/errgroup"

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

	// TODO: do something with stats?
	_, err := driver.WithData(nil, nil, logger, func(dir string, stats *pps.ProcessStats) error {
		return driver.WithActiveData([]*common.Input{}, dir, func() error {
			eg, serviceCtx := errgroup.WithContext(pachClient.Ctx())
			eg.Go(func() error { return pipeline.RunUserCode(driver.WithContext(serviceCtx), logger) })
			eg.Go(func() error { return pipeline.ReceiveSpout(serviceCtx, pachClient, pipelineInfo, logger) })
			return eg.Wait()
		})
	})
	return err
}
