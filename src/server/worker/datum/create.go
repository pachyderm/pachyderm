package datum

import (
	"context"
	"math"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"golang.org/x/sync/errgroup"
)

func Create(pachClient *client.APIClient, taskDoer task.Doer, input *pps.Input) (string, error) {
	switch {
	case input.Pfs != nil:
		return createPFS(pachClient, taskDoer, input.Pfs)
	case input.Union != nil:
		return createUnion(pachClient, taskDoer, input.Union)
	case input.Cross != nil:
		return createCross(pachClient, taskDoer, input.Cross)
	case input.Join != nil:
		return createJoin(pachClient, taskDoer, input.Join)
	case input.Group != nil:
		return createGroup(pachClient, taskDoer, input.Group)
	case input.Cron != nil:
		return createCron(pachClient, taskDoer, input.Cron)
	default:
		return "", errors.Errorf("unrecognized input type: %v", input)
	}
}

func createPFS(pachClient *client.APIClient, taskDoer task.Doer, input *pps.PFSInput) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		fileSetID, err := pachClient.GetFileSet(input.Repo, input.Branch, input.Commit)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, fileSetID); err != nil {
			return err
		}
		shards, err := pachClient.ShardFileSet(fileSetID)
		if err != nil {
			return err
		}
		var inputs []*types.Any
		for i, shard := range shards {
			input, err := serializePFSTask(&PFSTask{
				Input:     input,
				PathRange: shard,
				BaseIndex: createBaseIndex(int64(i)),
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *types.Any, err error) error {
			if err != nil {
				return err
			}
			result, err := deserializePFSTaskResult(output)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, result.FileSetId); err != nil {
				return err
			}
			resultFileSetIDs[i] = result.FileSetId
			return nil
		}); err != nil {
			return err
		}
		outputFileSetID, err = ComposeFileSets(ctx, taskDoer, resultFileSetIDs)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createBaseIndex(index int64) int64 {
	return index * int64(math.Pow(float64(10), float64(16)))
}

func ComposeFileSets(ctx context.Context, taskDoer task.Doer, fileSetIDs []string) (string, error) {
	input, err := serializeComposeTask(&ComposeTask{
		FileSetIds: fileSetIDs,
	})
	if err != nil {
		return "", err
	}
	output, err := task.DoOne(ctx, taskDoer, input)
	if err != nil {
		return "", err
	}
	result, err := deserializeComposeTaskResult(output)
	if err != nil {
		return "", err
	}
	return result.FileSetId, nil
}

func createUnion(pachClient *client.APIClient, taskDoer task.Doer, inputs []*pps.Input) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		fileSetIDs, err := createInputs(pachClient, taskDoer, renewer, inputs)
		if err != nil {
			return err
		}
		outputFileSetID, err = ComposeFileSets(ctx, taskDoer, fileSetIDs)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createInputs(pachClient *client.APIClient, taskDoer task.Doer, renewer *renew.StringSet, inputs []*pps.Input) ([]string, error) {
	eg, ctx := errgroup.WithContext(pachClient.Ctx())
	pachClient = pachClient.WithCtx(ctx)
	outputFileSetIDs := make([]string, len(inputs))
	for i, input := range inputs {
		i := i
		input := input
		eg.Go(func() error {
			outputFileSetID, err := Create(pachClient, taskDoer, input)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, outputFileSetID); err != nil {
				return err
			}
			outputFileSetIDs[i] = outputFileSetID
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return outputFileSetIDs, nil
}

func createCross(pachClient *client.APIClient, taskDoer task.Doer, inputs []*pps.Input) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		fileSetIDs, err := createInputs(pachClient, taskDoer, renewer, inputs)
		if err != nil {
			return err
		}
		var maxIdx int
		var baseFileSetID string
		var baseFileSetShards []*pfs.PathRange
		for i, fileSetID := range fileSetIDs {
			shards, err := pachClient.ShardFileSet(fileSetID)
			if err != nil {
				return err
			}
			if len(shards) > len(baseFileSetShards) {
				maxIdx = i
				baseFileSetID = fileSetID
				baseFileSetShards = shards
			}
		}
		fileSetIDs = append(fileSetIDs[:maxIdx], fileSetIDs[maxIdx+1:]...)
		var inputs []*types.Any
		for i, shard := range baseFileSetShards {
			input, err := serializeCrossTask(&CrossTask{
				BaseFileSetId:        baseFileSetID,
				BaseFileSetPathRange: shard,
				FileSetIds:           fileSetIDs,
				BaseIndex:            createBaseIndex(int64(i)),
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *types.Any, err error) error {
			if err != nil {
				return err
			}
			result, err := deserializeCrossTaskResult(output)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, result.FileSetId); err != nil {
				return err
			}
			resultFileSetIDs[i] = result.FileSetId
			return nil
		}); err != nil {
			return err
		}
		outputFileSetID, err = ComposeFileSets(ctx, taskDoer, resultFileSetIDs)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createJoin(pachClient *client.APIClient, taskDoer task.Doer, inputs []*pps.Input) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		fileSetIDs, err := createInputs(pachClient, taskDoer, renewer, inputs)
		if err != nil {
			return err
		}
		keyFileSetIDs, err := createKeyFileSets(pachClient, taskDoer, renewer, fileSetIDs, KeyTask_JOIN)
		if err != nil {
			return err
		}
		outputFileSetID, err = mergeKeyFileSets(pachClient, taskDoer, keyFileSetIDs, MergeTask_JOIN)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createKeyFileSets(pachClient *client.APIClient, taskDoer task.Doer, renewer *renew.StringSet, fileSetIDs []string, keyType KeyTask_Type) ([]string, error) {
	eg, ctx := errgroup.WithContext(pachClient.Ctx())
	pachClient = pachClient.WithCtx(ctx)
	outputFileSetIDs := make([]string, len(fileSetIDs))
	for i, fileSetID := range fileSetIDs {
		i := i
		fileSetID := fileSetID
		eg.Go(func() error {
			outputFileSetID, err := createKeyFileSet(pachClient, taskDoer, fileSetID, keyType)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, outputFileSetID); err != nil {
				return err
			}
			outputFileSetIDs[i] = outputFileSetID
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return outputFileSetIDs, nil
}

func createKeyFileSet(pachClient *client.APIClient, taskDoer task.Doer, fileSetID string, keyType KeyTask_Type) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		shards, err := pachClient.ShardFileSet(fileSetID)
		if err != nil {
			return err
		}
		var inputs []*types.Any
		for _, shard := range shards {
			input, err := serializeKeyTask(&KeyTask{
				FileSetId: fileSetID,
				PathRange: shard,
				Type:      keyType,
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *types.Any, err error) error {
			if err != nil {
				return err
			}
			result, err := deserializeKeyTaskResult(output)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, result.FileSetId); err != nil {
				return err
			}
			resultFileSetIDs[i] = result.FileSetId
			return nil
		}); err != nil {
			return err
		}
		outputFileSetID, err = ComposeFileSets(ctx, taskDoer, resultFileSetIDs)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func mergeKeyFileSets(pachClient *client.APIClient, taskDoer task.Doer, fileSetIDs []string, mergeType MergeTask_Type) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		shards, err := common.Shard(pachClient, fileSetIDs)
		if err != nil {
			return err
		}
		var inputs []*types.Any
		for _, shard := range shards {
			input, err := serializeMergeTask(&MergeTask{
				FileSetIds: fileSetIDs,
				PathRange:  shard,
				Type:       mergeType,
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *types.Any, err error) error {
			if err != nil {
				return err
			}
			result, err := deserializeMergeTaskResult(output)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, result.FileSetId); err != nil {
				return err
			}
			resultFileSetIDs[i] = result.FileSetId
			return nil
		}); err != nil {
			return err
		}
		outputFileSetID, err = ComposeFileSets(ctx, taskDoer, resultFileSetIDs)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createGroup(pachClient *client.APIClient, taskDoer task.Doer, inputs []*pps.Input) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		fileSetIDs, err := createInputs(pachClient, taskDoer, renewer, inputs)
		if err != nil {
			return err
		}
		keyFileSetIDs, err := createKeyFileSets(pachClient, taskDoer, renewer, fileSetIDs, KeyTask_GROUP)
		if err != nil {
			return err
		}
		outputFileSetID, err = mergeKeyFileSets(pachClient, taskDoer, keyFileSetIDs, MergeTask_GROUP)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createCron(pachClient *client.APIClient, taskDoer task.Doer, input *pps.CronInput) (string, error) {
	return createPFS(pachClient, taskDoer, &pps.PFSInput{
		Name:   input.Name,
		Repo:   input.Repo,
		Branch: "master",
		Commit: input.Commit,
		Glob:   "/*",
	})
}

func serializePFSTask(task *PFSTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializePFSTaskResult(taskAny *types.Any) (*PFSTaskResult, error) {
	task := &PFSTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeCrossTask(task *CrossTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeCrossTaskResult(taskAny *types.Any) (*CrossTaskResult, error) {
	task := &CrossTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeKeyTask(task *KeyTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeKeyTaskResult(taskAny *types.Any) (*KeyTaskResult, error) {
	task := &KeyTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeMergeTask(task *MergeTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeMergeTaskResult(taskAny *types.Any) (*MergeTaskResult, error) {
	task := &MergeTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeComposeTask(task *ComposeTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeComposeTaskResult(taskAny *types.Any) (*ComposeTaskResult, error) {
	task := &ComposeTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}
