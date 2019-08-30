package spawner

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

type spawnerFunc func(*client.APIClient, *pps.PipelineInfo, logs.TaggedLogger, driver.Driver) error

func Run(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, logger logs.TaggedLogger, driver driver.Driver) error {
	pipelineType, runFn := func() (string, spawnerFunc) {
		switch {
		case pipelineInfo.Service != nil:
			return "service", runService
		case pipelineInfo.Spout != nil:
			return "spout", runSpout
		default:
			return "pipeline", runMap
		}
	}()

	logger.Logf("Launching %v spawner process", pipelineType)
	err := runFn(pachClient, pipelineInfo, logger, driver)
	if err != nil {
		logger.Logf("error running the %v spawner process: %v", pipelineType, err)
	}
	return err
}

// Runs the user code until cancelled by the context - used for services
// Unlike how the worker runs user code, this does not set environment variables
// or collect stats.
func runUserCode(
	ctx context.Context,
	driver driver.Driver,
	logger logs.TaggedLogger,
) error {
	return backoff.RetryUntilCancel(ctx, func() error {
		// TODO: shouldn't this set up env like the worker does?
		// TODO: what about the user error handling code?
		return driver.RunUserCode(ctx, logger, nil, &pps.ProcessStats{}, nil)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("error in runUserCode: %+v, retrying in: %+v", err, d)
		return nil
	})
}
