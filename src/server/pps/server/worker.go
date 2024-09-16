package server

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/pipeline/transform"
	"github.com/pachyderm/pachyderm/v2/src/storage"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
)

type WorkerEnv struct {
	PFS         pfs.APIClient
	TaskService task.Service
	PJS         pjs.APIClient
	Fileset     storage.FilesetClient
}

type Worker struct {
	env WorkerEnv
}

func NewWorker(env WorkerEnv) *Worker {
	return &Worker{env: env}
}

func (w *Worker) Run(ctx context.Context) error {
	return backoff.RetryUntilCancel(ctx, func() error {
		return transform.PreprocessingWorker(ctx, w.env.PFS, w.env.TaskService, nil, w.env.PJS, w.env.Fileset)
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Debug(ctx, "error in pps worker", zap.Error(err))
		return nil
	})
}
