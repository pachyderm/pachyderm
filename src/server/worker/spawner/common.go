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

func (s *spawner) runServiceUserCode(ctx context.Context, logger logs.TaggedLogger) error {
	return backoff.RetryNotify(func() error {
		// if we have a spout, then asynchronously receive spout data
		if s.pipelineInfo.Spout != nil {
			go s.receiveSpout(ctx, logger)
		}
		return s.utils.runUserCode(ctx, logger, nil, &pps.ProcessStats{}, nil)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			return err
		default:
			logger.Logf("error running user code: %+v, retrying in: %+v", err, d)
			return nil
		}
	})
}
