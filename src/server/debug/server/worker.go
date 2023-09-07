package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
)

type Worker struct {
	env Env
}

func NewWorker(env Env) *Worker {
	return &Worker{
		env: env,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	return pfsload.Worker(w.env.GetPachClient(pctx.Child(ctx, "pfsload")), w.env.TaskService) //nolint:errcheck
}
