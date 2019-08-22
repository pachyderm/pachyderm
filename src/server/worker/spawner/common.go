package spawner

type Spawner interface {
	RunService(Context.context) error
	RunSpout(Context.context) error
	RunMap(Context.context) error
}

type spawner struct {
}

func NewSpawner() Spawner {
}

func (s *spawner) runServiceUserCode(ctx context.Context, logger logs.TaggedLogger) error {
	return backoff.RetryNotify(func() error {
		// if we have a spout, then asynchronously receive spout data
		if a.pipelineInfo.Spout != nil {
			go a.receiveSpout(ctx, logger)
		}
		return a.runUserCode(ctx, logger, nil, &pps.ProcessStats{}, nil)
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
