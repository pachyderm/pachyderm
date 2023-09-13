package server

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
)

type WorkerEnv struct {
	TaskService   task.Service
	GetPachClient func(context.Context) *client.APIClient
}

type Worker struct {
	env WorkerEnv
}

func NewWorker(env WorkerEnv) *Worker {
	return &Worker{env: env}
}

func (w *Worker) Run(ctx context.Context) error {
	taskSource := w.env.TaskService.NewSource(ppsTaskNamespace)
	return backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		err := taskSource.Iterate(ctx, func(ctx context.Context, input *anypb.Any) (_ *anypb.Any, retErr error) {
			defer log.Span(ctx, "pps task", log.Proto("input", input))(log.Errorp(&retErr))
			switch {
			case datum.IsTask(input):
				pachClient := w.env.GetPachClient(ctx)
				return datum.ProcessTask(pachClient, input)
			default:
				return nil, errors.Errorf("unrecognized any type (%v) in pps worker", input.TypeUrl)
			}
		})
		return errors.EnsureStack(err)
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Debug(ctx, "error in pps worker", zap.Error(err))
		return nil
	})
}
