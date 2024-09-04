package server

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"github.com/pachyderm/pachyderm/v2/src/server/pps/internal/worker"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/pipeline/transform"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
)

type WorkerEnv struct {
	PFS         pfs.APIClient
	TaskService task.Service
	PJS         pjs.APIClient
}

type Worker struct {
	env WorkerEnv
}

func NewWorker(env WorkerEnv) *Worker {
	return &Worker{env: env}
}

func (w *Worker) Run(ctx context.Context) (retErr error) {
	// The PPS worker listens on both the task service and the job service for work.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		errors.JoinInto(&retErr, backoff.RetryUntilCancel(ctx, func() error {
			return transform.PreprocessingWorker(ctx, w.env.PFS, w.env.TaskService, nil)
		}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
			log.Debug(ctx, "error in pps worker", zap.Error(err))
			return nil
		}))
	}()
	go func() {
		defer wg.Done()
		if w.env.PJS == nil {
			return
		}
		errors.JoinInto(&retErr, backoff.RetryUntilCancel(ctx, func() error {
			q, err := w.env.PJS.ProcessQueue(ctx)
			if err != nil {
				return errors.Wrap(err, "PJS.ProcessQueue")
			}
			if err := q.Send(&pjs.ProcessQueueRequest{Queue: &pjs.Queue{Id: worker.ProgramHash()}}); err != nil {
				return errors.Wrap(err, "send initial request")
			}
			for r, err := q.Recv(); err == nil; r, err = q.Recv() {
				log.Info(ctx, "response received", zap.Strings("inputs", r.Input), zap.String("context", r.Context))
				if err := q.Send(&pjs.ProcessQueueRequest{Result: &pjs.ProcessQueueRequest_Failed{Failed: true}}); err != nil {
					return errors.Wrap(err, "sending request")
				}
			}
			return errors.Wrap(err, "receive subsequent responses")
		}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
			log.Debug(ctx, "error in pps worker", zap.Error(err))
			return nil
		}))
	}()
	wg.Wait()
	return
}
