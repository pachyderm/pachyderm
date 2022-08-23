package server

import (
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func (a *apiServer) worker() {
	ctx := a.env.BackgroundContext
	taskSource := a.env.TaskService.NewSource(ppsTaskNamespace)
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		err := taskSource.Iterate(ctx, func(ctx context.Context, input *types.Any) (*types.Any, error) {
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
		log.Printf("error in pps worker: %v", err)
		return nil
	})
}
