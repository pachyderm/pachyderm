package spawner

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/src/server/worker/utils"
)

type spawnerFunc func(*client.APIClient, *pps.PipelineInfo, logs.TaggedLogger, utils.Utils) error

func Run(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, logger logs.TaggedLogger, utils utils.Utils) error {
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
	err := runFn(pachClient, pipelineInfo, logger, utils)
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
	utils utils.Utils,
	logger logs.TaggedLogger,
) error {
	return runUntilCancel(ctx, logger, "runUserCode", func() error {
		// TODO: shouldn't this set up env like the worker does?
		// TODO: what about the user error handling code?
		return utils.RunUserCode(ctx, logger, nil, &pps.ProcessStats{}, nil)
	})
}

func runUntilCancel(ctx context.Context, logger logs.TaggedLogger, name string, cb func() error) error {
	return backoff.RetryNotify(cb, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if common.IsDone(ctx) {
			return err
		}
		logger.Logf("error in %s: %+v, retrying in: %+v", name, err, d)
		return nil
	})
}
