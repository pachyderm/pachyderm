package task

import (
	"context"

	taskapi "github.com/pachyderm/pachyderm/v2/src/task"

	"github.com/gogo/protobuf/jsonpb"
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

// HandleList implements the functionality for an arbitrary service's ListTask gRPC
func HandleList(ctx context.Context, svc Service, req *taskapi.ListTaskRequest, send func(info *taskapi.TaskInfo) error) error {
	var marshaler jsonpb.Marshaler
	return errors.EnsureStack(svc.List(ctx, req.Group.Namespace, req.Group.Group, func(key *TaskKey, data *Task, claimed bool) error {
		state := translateTaskState(data.State)
		if claimed {
			state = taskapi.State_CLAIMED
		}
		var input types.DynamicAny
		if err := types.UnmarshalAny(data.Input, &input); err != nil {
			return errors.EnsureStack(err)
		}
		inputJSON, err := marshaler.MarshalToString(input.Message)
		if err != nil {
			return errors.EnsureStack(err)
		}
		info := &taskapi.TaskInfo{
			ID: data.ID,
			Group: &taskapi.Group{
				Group:     key.Group,
				Namespace: key.Namespace,
			},
			Key:       key.Key,
			State:     state,
			Reason:    data.Reason,
			InputType: data.Input.TypeUrl,
			InputData: inputJSON,
		}
		return errors.EnsureStack(send(info))
	}))
}

// Count returns the number of tasks and claims in the given namespace and group (if nonempty)
func Count(ctx context.Context, namespace, group string, service Service) (tasks int64, claims int64, retErr error) {
	retErr = errors.EnsureStack(service.List(ctx, namespace, group, func(_ *TaskKey, _ *Task, claimed bool) error {
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
