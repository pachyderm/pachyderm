package spawner

func Run(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, logger logs.TaggedLogger, utils common.Utils) error {
	pipelineType, runFn := func() (string, SpawnerFunc) {
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

func runUserCode(
	ctx context.Context,
	logger logs.TaggedLogger,
) error {
	return runUntil(ctx, "runUserCode", func() error {
		return s.utils.runUserCode(ctx, logger, nil, &pps.ProcessStats{}, nil)
	})
}

func runUntil(ctx context.Context, name string, cb func() error) error {
	return backoff.RetryNotify(cb, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if common.IsDone(ctx) {
			return err
		}
		logger.Logf("error in %s: %+v, retrying in: %+v", name, err, d)
		return nil
	})
}
