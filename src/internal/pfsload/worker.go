package pfsload

import (
	"context"
	"math/rand"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
)

const namespace = "pfsload"

func Worker(ctx context.Context, c pfs.APIClient, taskService task.Service) error {
	taskSource := taskService.NewSource(namespace)
	return backoff.RetryUntilCancel(ctx, func() error {
		err := taskSource.Iterate(ctx, func(ctx context.Context, input *anypb.Any) (*anypb.Any, error) {
			switch {
			case input.MessageIs(&PutFileTask{}):
				putFileTask, err := deserializePutFileTask(input)
				if err != nil {
					return nil, err
				}
				return processPutFileTask(ctx, c, putFileTask)
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

func processPutFileTask(ctx context.Context, c pfs.APIClient, task *PutFileTask) (*anypb.Any, error) {
	result := &PutFileTaskResult{}
	if err := log.LogStep(ctx, "putFileTask", func(ctx context.Context) error {
		ctx = client.SetAuthToken(ctx, task.AuthToken)
		client := NewValidatorClient(NewPachClient(c))
		fileSource := NewFileSource(task.FileSource, rand.New(rand.NewSource(task.Seed)))
		fileSetId, err := PutFile(ctx, client, fileSource, int(task.Count))
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

func deserializePutFileTask(taskAny *anypb.Any) (*PutFileTask, error) {
	task := &PutFileTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializePutFileTaskResult(task *PutFileTaskResult) (*anypb.Any, error) {
	return anypb.New(task)
}
