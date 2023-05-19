package pfsload

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"golang.org/x/sync/errgroup"
)

func Modification(env *Env, commit *pfs.Commit, spec *ModificationSpec) error {
	taskDoer := env.TaskDoer()
	client := env.Client()
	eg, ctx := errgroup.WithContext(client.Ctx())
	inputChan := make(chan *types.Any)
	eg.Go(func() error {
		defer close(inputChan)
		for i := 0; i < int(spec.Count); i++ {
			input, err := serializePutFileTask(&PutFileTask{
				Count:      spec.PutFile.Count,
				FileSource: env.FileSource(spec.PutFile.Source),
				Seed:       env.Seed(),
				AuthToken:  env.AuthToken(),
			})
			if err != nil {
				return err
			}
			select {
			case inputChan <- input:
			case <-ctx.Done():
				return errors.EnsureStack(context.Cause(ctx))
			}
		}
		return nil
	})
	eg.Go(func() error {
		return errors.EnsureStack(taskDoer.Do(
			ctx,
			inputChan,
			func(_ int64, output *types.Any, err error) error {
				if err != nil {
					return err
				}
				data, err := deserializePutFileTaskResult(output)
				if err != nil {
					return err
				}
				if err := client.AddFileSet(ctx, commit, data.FileSetId); err != nil {
					return errors.EnsureStack(err)
				}
				if data.Hash != nil {
					env.Validator().AddHash(data.Hash)
				}
				return nil
			},
		))
	})
	return errors.EnsureStack(eg.Wait())
}

func serializePutFileTask(task *PutFileTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializePutFileTaskResult(taskAny *types.Any) (*PutFileTaskResult, error) {
	task := &PutFileTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func PutFile(ctx context.Context, c Client, fileSource FileSource, count int) (string, error) {
	resp, err := c.WithCreateFileSetClient(ctx, func(mf client.ModifyFile) error {
		files, err := Files(fileSource, count)
		if err != nil {
			return err
		}
		for _, file := range files {
			if err := mf.PutFile(file.Path(), file); err != nil {
				return errors.EnsureStack(err)
			}
		}
		return nil
	})
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	return resp.FileSetId, nil
}
