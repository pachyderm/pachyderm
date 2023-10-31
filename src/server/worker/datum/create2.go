package datum

import (
	"context"
	"math"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
)

type taskResult struct {
	fsid string
	err  error
}

func Create2(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, input *pps.Input, taskCh chan taskResult) {
	defer close(taskCh)
	switch {
	case input.Pfs != nil:
		createPFS2(ctx, c, taskDoer, input.Pfs, taskCh)
	// case input.Union != nil:
	// 	return createUnion(ctx, c, taskDoer, input.Union, fsidCh)
	// case input.Cross != nil:
	// 	return createCross(ctx, c, taskDoer, input.Cross, fsidCh)
	// case input.Join != nil:
	// 	return createJoin(ctx, c, taskDoer, input.Join, fsidCh)
	// case input.Group != nil:
	// 	return createGroup(ctx, c, taskDoer, input.Group, fsidCh)
	// case input.Cron != nil:
	// 	return createCron(ctx, c, taskDoer, input.Cron, fsidCh)
	default:
		taskCh <- taskResult{err: errors.Errorf("unrecognized input type: %v", input)}
	}
}

func createPFS2(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, input *pps.PFSInput, parentTaskCh chan taskResult) {
	authToken := getAuthToken(ctx)
	// var outputFileSetID string
	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
		fileSetID, err := client.GetFileSet(ctx, c, input.Project, input.Repo, input.Branch, input.Commit)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, fileSetID); err != nil {
			return err
		}
		// TODO: Stream back first shard?
		log.Info(ctx, "Sharding files")
		shards, err := client.ShardFileSet(ctx, c, fileSetID)
		if err != nil {
			return err
		}
		var inputs []*anypb.Any
		for i, shard := range shards {
			input, err := serializePFSTask(&PFSTask{
				Input:     input,
				PathRange: shard,
				BaseIndex: createBaseIndex(int64(i)),
				AuthToken: authToken,
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		log.Info(ctx, "Shards created", zap.Int("numShards", len(inputs)))
		// resultFileSetIDs := make([]string, len(inputs))
		err = task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
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
			// log.Info(ctx, "Sending back task result", zap.Int("index", int(i)))
			parentTaskCh <- taskResult{fsid: result.FileSetId}
			// log.Info(ctx, "Sent result", zap.Int("index", int(i)))
			// resultFileSetIDs[i] = result.FileSetId
			return nil
		})
		log.Info(ctx, "All tasks done")
		// outputFileSetID, err = ComposeFileSets(ctx, c, taskDoer, resultFileSetIDs)
		return err
	}); err != nil {
		parentTaskCh <- taskResult{err: err}
	}
}

func createBaseIndex2(index int64) int64 {
	return index * int64(math.Pow(float64(10), float64(16)))
}

func ComposeFileSets2(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, fileSetIDs []string) (string, error) {
	input, err := serializeComposeTask(&ComposeTask{
		FileSetIds: fileSetIDs,
		AuthToken:  getAuthToken(ctx),
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

// func createUnion2(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, inputs []*pps.Input, parentTaskCh chan taskResult) {
// 	// var outputFileSetID string
// 	defer close(parentTaskCh)
// 	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
// 		taskCh := make(chan taskResult) // TODO: Buffered??
// 		go createInputs2(ctx, c, taskDoer, renewer, inputs, taskCh)
// 		// outputFileSetID, err = ComposeFileSets(ctx, c, taskDoer, fileSetIDs)
// 		for task := range taskCh {
// 			parentTaskCh <- task
// 		}
// 	}); err != nil {
// 		return err
// 	}
// }

// func createInputs2(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, renewer *renew.StringSet, inputs []*pps.Input, parentTaskCh chan taskResult) error {
// 	eg, ctx := errgroup.WithContext(ctx)
// 	outputFileSetIDs := make([]string, len(inputs))
// 	for i, input := range inputs {
// 		i := i
// 		input := input
// 		eg.Go(func() error {
// 			fsidCh := make(chan string)
// 			err := Create2(ctx, c, taskDoer, input, fsidCh)
// 			if err != nil {
// 				return err
// 			}
// 			if err := renewer.Add(ctx, outputFileSetID); err != nil {
// 				return err
// 			}
// 			outputFileSetIDs[i] = outputFileSetID
// 			return nil
// 		})
// 	}
// 	if err := eg.Wait(); err != nil {
// 		return nil, errors.EnsureStack(err)
// 	}
// 	return outputFileSetIDs, nil
// }

// func createCross(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, inputs []*pps.Input, fsidCh chan string) error {
// 	var outputFileSetID string
// 	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
// 		fileSetIDs, err := createInputs(ctx, c, taskDoer, renewer, inputs)
// 		if err != nil {
// 			return err
// 		}
// 		var baseFileSetIndex int
// 		var baseFileSetShards []*pfs.PathRange
// 		for i, fileSetID := range fileSetIDs {
// 			shards, err := client.ShardFileSet(ctx, c, fileSetID)
// 			if err != nil {
// 				return err
// 			}
// 			if len(shards) > len(baseFileSetShards) {
// 				baseFileSetIndex = i
// 				baseFileSetShards = shards
// 			}
// 		}
// 		var inputs []*anypb.Any
// 		for i, shard := range baseFileSetShards {
// 			input, err := serializeCrossTask(&CrossTask{
// 				FileSetIds:           fileSetIDs,
// 				BaseFileSetIndex:     int64(baseFileSetIndex),
// 				BaseFileSetPathRange: shard,
// 				BaseIndex:            createBaseIndex(int64(i)),
// 				AuthToken:            getAuthToken(ctx),
// 			})
// 			if err != nil {
// 				return err
// 			}
// 			inputs = append(inputs, input)
// 		}
// 		resultFileSetIDs := make([]string, len(inputs))
// 		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			result, err := deserializeCrossTaskResult(output)
// 			if err != nil {
// 				return err
// 			}
// 			if err := renewer.Add(ctx, result.FileSetId); err != nil {
// 				return err
// 			}
// 			resultFileSetIDs[i] = result.FileSetId
// 			return nil
// 		}); err != nil {
// 			return err
// 		}
// 		outputFileSetID, err = ComposeFileSets(ctx, c, taskDoer, resultFileSetIDs)
// 		return err
// 	}); err != nil {
// 		return "", err
// 	}
// 	return outputFileSetID, nil
// }

// func createJoin(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, inputs []*pps.Input, fsidCh chan string) error {
// 	var outputFileSetID string
// 	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
// 		fileSetIDs, err := createInputs(ctx, c, taskDoer, renewer, inputs)
// 		if err != nil {
// 			return err
// 		}
// 		keyFileSetIDs, err := createKeyFileSets(ctx, c, taskDoer, renewer, fileSetIDs, KeyTask_JOIN)
// 		if err != nil {
// 			return err
// 		}
// 		outputFileSetID, err = mergeKeyFileSets(ctx, c, taskDoer, keyFileSetIDs, MergeTask_JOIN)
// 		return err
// 	}); err != nil {
// 		return "", err
// 	}
// 	return outputFileSetID, nil
// }

// func createKeyFileSets(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, renewer *renew.StringSet, fileSetIDs []string, keyType KeyTask_Type) ([]string, error) {
// 	eg, ctx := errgroup.WithContext(ctx)
// 	outputFileSetIDs := make([]string, len(fileSetIDs))
// 	for i, fileSetID := range fileSetIDs {
// 		i := i
// 		fileSetID := fileSetID
// 		eg.Go(func() error {
// 			outputFileSetID, err := createKeyFileSet(ctx, c, taskDoer, fileSetID, keyType)
// 			if err != nil {
// 				return err
// 			}
// 			if err := renewer.Add(ctx, outputFileSetID); err != nil {
// 				return err
// 			}
// 			outputFileSetIDs[i] = outputFileSetID
// 			return nil
// 		})
// 	}
// 	if err := eg.Wait(); err != nil {
// 		return nil, errors.EnsureStack(err)
// 	}
// 	return outputFileSetIDs, nil
// }

// func createKeyFileSet(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, fileSetID string, keyType KeyTask_Type) (string, error) {
// 	var outputFileSetID string
// 	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
// 		shards, err := client.ShardFileSet(ctx, c, fileSetID)
// 		if err != nil {
// 			return err
// 		}
// 		var inputs []*anypb.Any
// 		for _, shard := range shards {
// 			input, err := serializeKeyTask(&KeyTask{
// 				FileSetId: fileSetID,
// 				PathRange: shard,
// 				Type:      keyType,
// 				AuthToken: getAuthToken(ctx),
// 			})
// 			if err != nil {
// 				return err
// 			}
// 			inputs = append(inputs, input)
// 		}
// 		resultFileSetIDs := make([]string, len(inputs))
// 		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			result, err := deserializeKeyTaskResult(output)
// 			if err != nil {
// 				return err
// 			}
// 			if err := renewer.Add(ctx, result.FileSetId); err != nil {
// 				return err
// 			}
// 			resultFileSetIDs[i] = result.FileSetId
// 			return nil
// 		}); err != nil {
// 			return err
// 		}
// 		outputFileSetID, err = ComposeFileSets(ctx, c, taskDoer, resultFileSetIDs)
// 		return err
// 	}); err != nil {
// 		return "", err
// 	}
// 	return outputFileSetID, nil
// }

// func mergeKeyFileSets(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, fileSetIDs []string, mergeType MergeTask_Type) (string, error) {
// 	var outputFileSetID string
// 	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
// 		shards, err := common.Shard(ctx, c, fileSetIDs)
// 		if err != nil {
// 			return err
// 		}
// 		var inputs []*anypb.Any
// 		for _, shard := range shards {
// 			input, err := serializeMergeTask(&MergeTask{
// 				FileSetIds: fileSetIDs,
// 				PathRange:  shard,
// 				Type:       mergeType,
// 				AuthToken:  getAuthToken(ctx),
// 			})
// 			if err != nil {
// 				return err
// 			}
// 			inputs = append(inputs, input)
// 		}
// 		resultFileSetIDs := make([]string, len(inputs))
// 		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			result, err := deserializeMergeTaskResult(output)
// 			if err != nil {
// 				return err
// 			}
// 			if err := renewer.Add(ctx, result.FileSetId); err != nil {
// 				return err
// 			}
// 			resultFileSetIDs[i] = result.FileSetId
// 			return nil
// 		}); err != nil {
// 			return err
// 		}
// 		outputFileSetID, err = ComposeFileSets(ctx, c, taskDoer, resultFileSetIDs)
// 		return err
// 	}); err != nil {
// 		return "", err
// 	}
// 	return outputFileSetID, nil
// }

// func createGroup(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, inputs []*pps.Input, fsidCh chan string) error {
// 	var outputFileSetID string
// 	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
// 		fileSetIDs, err := createInputs(ctx, c, taskDoer, renewer, inputs)
// 		if err != nil {
// 			return err
// 		}
// 		keyFileSetIDs, err := createKeyFileSets(ctx, c, taskDoer, renewer, fileSetIDs, KeyTask_GROUP)
// 		if err != nil {
// 			return err
// 		}
// 		outputFileSetID, err = mergeKeyFileSets(ctx, c, taskDoer, keyFileSetIDs, MergeTask_GROUP)
// 		return err
// 	}); err != nil {
// 		return "", err
// 	}
// 	return outputFileSetID, nil
// }

// func createCron(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, input *pps.CronInput, fsidCh chan string) error {
// 	return createPFS(ctx, c, taskDoer, &pps.PFSInput{
// 		Name:    input.Name,
// 		Project: input.Project,
// 		Repo:    input.Repo,
// 		Branch:  "master",
// 		Commit:  input.Commit,
// 		Glob:    "/*",
// 	}, fsidCh)
// }

func serializePFSTask2(task *PFSTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializePFSTaskResult2(taskAny *anypb.Any) (*PFSTaskResult, error) {
	task := &PFSTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

// func serializeCrossTask(task *CrossTask) (*anypb.Any, error) {
// 	return anypb.New(task)
// }

// func deserializeCrossTaskResult(taskAny *anypb.Any) (*CrossTaskResult, error) {
// 	task := &CrossTaskResult{}
// 	if err := taskAny.UnmarshalTo(task); err != nil {
// 		return nil, errors.EnsureStack(err)
// 	}
// 	return task, nil
// }

// func serializeKeyTask(task *KeyTask) (*anypb.Any, error) {
// 	return anypb.New(task)
// }

// func deserializeKeyTaskResult(taskAny *anypb.Any) (*KeyTaskResult, error) {
// 	task := &KeyTaskResult{}
// 	if err := taskAny.UnmarshalTo(task); err != nil {
// 		return nil, errors.EnsureStack(err)
// 	}
// 	return task, nil
// }

// func serializeMergeTask(task *MergeTask) (*anypb.Any, error) {
// 	return anypb.New(task)
// }

// func deserializeMergeTaskResult(taskAny *anypb.Any) (*MergeTaskResult, error) {
// 	task := &MergeTaskResult{}
// 	if err := taskAny.UnmarshalTo(task); err != nil {
// 		return nil, errors.EnsureStack(err)
// 	}
// 	return task, nil
// }

func serializeComposeTask2(task *ComposeTask) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeComposeTaskResult2(taskAny *anypb.Any) (*ComposeTaskResult, error) {
	task := &ComposeTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func getAuthToken2(ctx context.Context) string {
	authToken, err := auth.GetAuthToken(ctx)
	if err != nil {
		log.Error(ctx, "no auth token", zap.Error(err))
	}
	return authToken
}
