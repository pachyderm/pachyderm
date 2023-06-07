package datum

import (
	"context"
	"math"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/anypb"
)

func Create(pachClient *client.APIClient, taskDoer task.Doer, input *pps.Input) (string, error) {
	pachClient = pachClient.WithCtx(pachClient.Ctx())
	authToken, _ := auth.GetAuthToken(pachClient.Ctx())
	pachClient.SetAuthToken(authToken)
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
		fileSetID, err := pachClient.GetFileSet(input.Project, input.Repo, input.Branch, input.Commit)
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
		var inputs []*anypb.Any
		for i, shard := range shards {
			input, err := serializePFSTask(&PFSTask{
				Input:     input,
				PathRange: shard,
				BaseIndex: createBaseIndex(int64(i)),
				AuthToken: pachClient.AuthToken(),
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
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
		outputFileSetID, err = ComposeFileSets(pachClient, taskDoer, resultFileSetIDs)
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createBaseIndex(index int64) int64 {
	return index * int64(math.Pow(float64(10), float64(16)))
}

func ComposeFileSets(pachClient *client.APIClient, taskDoer task.Doer, fileSetIDs []string) (string, error) {
	input, err := serializeComposeTask(&ComposeTask{
		FileSetIds: fileSetIDs,
		AuthToken:  pachClient.AuthToken(),
	})
	if err != nil {
		return "", err
	}
	output, err := task.DoOne(pachClient.Ctx(), taskDoer, input)
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
		outputFileSetID, err = ComposeFileSets(pachClient, taskDoer, fileSetIDs)
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
		var baseFileSetIndex int
		var baseFileSetShards []*pfs.PathRange
		for i, fileSetID := range fileSetIDs {
			shards, err := pachClient.ShardFileSet(fileSetID)
			if err != nil {
				return err
			}
			if len(shards) > len(baseFileSetShards) {
				baseFileSetIndex = i
				baseFileSetShards = shards
			}
		}
		var inputs []*anypb.Any
		for i, shard := range baseFileSetShards {
			input, err := serializeCrossTask(&CrossTask{
				FileSetIds:           fileSetIDs,
				BaseFileSetIndex:     int64(baseFileSetIndex),
				BaseFileSetPathRange: shard,
				BaseIndex:            createBaseIndex(int64(i)),
				AuthToken:            pachClient.AuthToken(),
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
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
		outputFileSetID, err = ComposeFileSets(pachClient, taskDoer, resultFileSetIDs)
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
		var inputs []*anypb.Any
		for _, shard := range shards {
			input, err := serializeKeyTask(&KeyTask{
				FileSetId: fileSetID,
				PathRange: shard,
				Type:      keyType,
				AuthToken: pachClient.AuthToken(),
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
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
		outputFileSetID, err = ComposeFileSets(pachClient, taskDoer, resultFileSetIDs)
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
		var inputs []*anypb.Any
		for _, shard := range shards {
			input, err := serializeMergeTask(&MergeTask{
				FileSetIds: fileSetIDs,
				PathRange:  shard,
				Type:       mergeType,
				AuthToken:  pachClient.AuthToken(),
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
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
		outputFileSetID, err = ComposeFileSets(pachClient, taskDoer, resultFileSetIDs)
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
		Name:    input.Name,
		Project: input.Project,
		Repo:    input.Repo,
		Branch:  "master",
		Commit:  input.Commit,
		Glob:    "/*",
	})
}

func serializePFSTask(task *PFSTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializePFSTaskResult(taskAny *anypb.Any) (*PFSTaskResult, error) {
	task := &PFSTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeCrossTask(task *CrossTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeCrossTaskResult(taskAny *anypb.Any) (*CrossTaskResult, error) {
	task := &CrossTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeKeyTask(task *KeyTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeKeyTaskResult(taskAny *anypb.Any) (*KeyTaskResult, error) {
	task := &KeyTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeMergeTask(task *MergeTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeMergeTaskResult(taskAny *anypb.Any) (*MergeTaskResult, error) {
	task := &MergeTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeComposeTask(task *ComposeTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeComposeTaskResult(taskAny *anypb.Any) (*ComposeTaskResult, error) {
	task := &ComposeTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}
