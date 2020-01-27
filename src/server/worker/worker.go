package worker

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/src/server/worker/pipeline/transform"
)

func (a *APIServer) worker() {
	ctx := a.driver.PachClient().Ctx()
	logger := logs.NewStatlessLogger(a.driver.PipelineInfo())

	backoff.RetryUntilCancel(ctx, func() error {
		return a.driver.NewTaskWorker().Run(
			a.driver.PachClient().Ctx(),
			func(ctx context.Context, task *work.Task, subtask *work.Task) error {
				return transform.Worker(a.driver, logger, task)
			},
		)
	}, backoff.NewConstantBackOff(200*time.Millisecond), func(err error, d time.Duration) error {
		logger.Logf("worker failed: %v; retrying in %v", err, d)
		return nil
	})
}
