package task

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/sync/errgroup"
)

// DoOne executes one task.
// NOTE: This interface is much less performant than the stream / batch interfaces for many tasks.
// Only use this interface for development / a small number of tasks.
func DoOne(ctx context.Context, doer Doer, input *types.Any) (*types.Any, error) {
	var result *types.Any
	if err := DoBatch(ctx, doer, []*types.Any{input}, func(_ int64, output *types.Any, err error) error {
		if err != nil {
			return err
		}
		result = output
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// DoBatch executes a batch of tasks.
func DoBatch(ctx context.Context, doer Doer, inputs []*types.Any, cb CollectFunc) error {
	var eg errgroup.Group
	inputChan := make(chan *types.Any)
	eg.Go(func() error {
		return errors.EnsureStack(doer.Do(ctx, inputChan, cb))
	})
	eg.Go(func() error {
		for _, input := range inputs {
			select {
			case inputChan <- input:
			case <-ctx.Done():
				return errors.EnsureStack(ctx.Err())
			}
		}
		close(inputChan)
		return nil
	})
	return errors.EnsureStack(eg.Wait())
}
