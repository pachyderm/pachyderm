package pfsload

import (
	"context"
	"math/rand"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"go.uber.org/zap"
)

const namespace = "pfsload"

func Worker(pachClient *client.APIClient, taskService task.Service) error {
	ctx := pachClient.Ctx()
	taskSource := taskService.NewSource(namespace)
	return backoff.RetryUntilCancel(ctx, func() error {
		err := taskSource.Iterate(ctx, func(ctx context.Context, input *types.Any) (*types.Any, error) {
			switch {
			case types.Is(input, &PutFileTask{}):
				putFileTask, err := deserializePutFileTask(input)
				if err != nil {
					return nil, err
				}
				return processPutFileTask(pachClient.WithCtx(ctx), putFileTask)
			default:
				return nil, errors.Errorf("unrecognized any type (%v) in pfsload worker", input.TypeUrl)
			}
		})
		return errors.EnsureStack(err)
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Info(ctx, "error in pfsload worker", zap.Error(err))
		return nil
	})
}

func processPutFileTask(pachClient *client.APIClient, task *PutFileTask) (*types.Any, error) {
	result := &PutFileTaskResult{}
	if err := log.LogStep(pachClient.Ctx(), "putFileTask", func(ctx context.Context) error {
		pachClient = pachClient.WithCtx(ctx)
		pachClient.SetAuthToken(task.AuthToken)
		client := NewValidatorClient(NewPachClient(pachClient))
		fileSource := NewFileSource(task.FileSource, rand.New(rand.NewSource(task.Seed)))
		fileSetId, err := PutFile(pachClient.Ctx(), client, fileSource, int(task.Count))
		if err != nil {
			return err
		}
		result.FileSetId = fileSetId
		result.Hash = client.hash
		return nil
	}); err != nil {
		return nil, err
	}
	return serializePutFileTaskResult(result)
}

func deserializePutFileTask(taskAny *types.Any) (*PutFileTask, error) {
	task := &PutFileTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializePutFileTaskResult(task *PutFileTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}
