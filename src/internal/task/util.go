package task

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"

	taskapi "github.com/pachyderm/pachyderm/v2/src/task"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/taskchain"
)

// DoOrdered processes tasks in parallel, but returns outputs in order via the provided callback cb.
func DoOrdered(ctx context.Context, doer Doer, inputs chan *anypb.Any, parallelism int, cb CollectFunc) error {
	taskChain := taskchain.New(ctx, semaphore.NewWeighted(int64(parallelism)))
	for {
		select {
		case input, ok := <-inputs:
			if !ok {
				return taskChain.Wait()
			}
			if err := taskChain.CreateTask(func(context.Context) (func() error, error) {
				result, err := DoOne(ctx, doer, input)
				if err != nil {
					return nil, errors.EnsureStack(err)
				}
				return func() error {
					return cb(-1, result, nil)
				}, nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
		case <-ctx.Done():
			return errors.EnsureStack(context.Cause(ctx))
		}
	}
}

// DoOne executes one task.
// NOTE: This interface is much less performant than the stream / batch interfaces for many tasks.
// Only use this interface for development / a small number of tasks.
func DoOne(ctx context.Context, doer Doer, input *anypb.Any) (*anypb.Any, error) {
	var result *anypb.Any
	if err := DoBatch(ctx, doer, []*anypb.Any{input}, func(_ int64, output *anypb.Any, err error) error {
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
func DoBatch(ctx context.Context, doer Doer, inputs []*anypb.Any, cb CollectFunc) error {
	var eg errgroup.Group
	inputChan := make(chan *anypb.Any)
	eg.Go(func() error {
		return errors.EnsureStack(doer.Do(ctx, inputChan, cb))
	})
	eg.Go(func() error {
		for _, input := range inputs {
			select {
			case inputChan <- input:
			case <-ctx.Done():
				return errors.EnsureStack(context.Cause(ctx))
			}
		}
		close(inputChan)
		return nil
	})
	return errors.EnsureStack(eg.Wait())
}

func translateTaskState(state State) taskapi.State {
	switch state {
	case State_RUNNING:
		return taskapi.State_RUNNING
	case State_SUCCESS:
		return taskapi.State_SUCCESS
	case State_FAILURE:
		return taskapi.State_FAILURE
	}
	return taskapi.State_UNKNOWN
}

// List implements the functionality for an arbitrary service's ListTask gRPC
func List(ctx context.Context, svc Service, req *taskapi.ListTaskRequest, send func(info *taskapi.TaskInfo) error) error {
	var marshaler protojson.MarshalOptions
	return errors.EnsureStack(svc.List(ctx, req.Group.Namespace, req.Group.Group, func(namespace, group string, data *Task, claimed bool) error {
		state := translateTaskState(data.State)
		if claimed {
			state = taskapi.State_CLAIMED
		}
		var inputJSON []byte
		input, err := data.Input.UnmarshalNew()
		if err != nil {
			// unmarshalling might fail due to the input type not being registered,
			// don't let this interfere with fetching or counting tasks
			log.Error(ctx, "couldn't unmarshal task input", zap.Error(err), zap.String("taskType", data.GetInput().TypeUrl), zap.String("taskID", data.GetId()))
		} else {
			inputJSON, err = marshaler.Marshal(input)
			if err != nil {
				log.Error(ctx, "couldn't marshal task input", zap.Error(err), zap.String("taskType", data.GetInput().TypeUrl), zap.String("taskID", data.GetId()))
			}
		}
		info := &taskapi.TaskInfo{
			Id: data.Id,
			Group: &taskapi.Group{
				Namespace: namespace,
				Group:     group,
			},
			State:     state,
			Reason:    data.Reason,
			InputType: data.Input.TypeUrl,
			InputData: string(inputJSON),
		}
		return errors.EnsureStack(send(info))
	}))
}

// Count returns the number of tasks and claims in the given namespace and group (if nonempty)
func Count(ctx context.Context, service Service, namespace, group string) (tasks int64, claims int64, retErr error) {
	retErr = errors.EnsureStack(service.List(ctx, namespace, group, func(_, _ string, _ *Task, claimed bool) error {
		tasks++
		if claimed {
			claims++
		}
		return nil
	}))
	if retErr != nil {
		return 0, 0, retErr
	}
	return
}
