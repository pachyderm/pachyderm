package pfsload

import (
	"context"
	"math/rand"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	log "github.com/sirupsen/logrus"
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
		log.Printf("error in pfsload worker: %v", err)
		return nil
	})
}

func processPutFileTask(pachClient *client.APIClient, task *PutFileTask) (*types.Any, error) {
	result := &PutFileTaskResult{}
	if err := miscutil.LogStep(pachClient.Ctx(), log.NewEntry(log.StandardLogger()), "processing put file task", func() error {
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
