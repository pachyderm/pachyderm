package server

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/pipeline/transform"
	"go.uber.org/zap"
)

type WorkerEnv struct {
	PachClient  *client.APIClient
	TaskService task.Service
}

type Worker struct {
	env WorkerEnv
}

func NewWorker(env WorkerEnv) *Worker {
	return &Worker{env: env}
}

func (w *Worker) Run(ctx context.Context) error {
	pachClient := w.env.PachClient.WithCtx(ctx)
	return backoff.RetryUntilCancel(ctx, func() error {
		return transform.PreprocessingWorker(pachClient, w.env.TaskService, nil)
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Debug(ctx, "error in pps worker", zap.Error(err))
		return nil
	})
}
