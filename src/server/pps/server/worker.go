package server

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
)

func (a *apiServer) worker(ctx context.Context) {
	taskSource := a.env.TaskService.NewSource(ppsTaskNamespace)
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		err := taskSource.Iterate(ctx, func(ctx context.Context, input *anypb.Any) (_ *anypb.Any, retErr error) {
			defer log.Span(ctx, "pps task", log.Proto("input", input))(log.Errorp(&retErr))
			switch {
			case datum.IsTask(input):
				pachClient := a.env.GetPachClient(ctx)
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
