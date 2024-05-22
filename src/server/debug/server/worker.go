package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type WorkerEnv struct {
	PFS         pfs.APIClient
	TaskService task.Service
}

type Worker struct {
	env WorkerEnv
}

func NewWorker(env WorkerEnv) *Worker {
	return &Worker{
		env: env,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	return pfsload.Worker(pctx.Child(ctx, "pfsload"), w.env.PFS, w.env.TaskService)
}
