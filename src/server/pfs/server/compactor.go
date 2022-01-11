package server

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type compactor struct {
	storage  *fileset.Storage
	maxFanIn int
}

func newCompactor(storage *fileset.Storage, maxFanIn int) *compactor {
	return &compactor{
		storage:  storage,
		maxFanIn: maxFanIn,
	}
}

func (c *compactor) Compact(ctx context.Context, taskDoer task.Doer, ids []fileset.ID, ttl time.Duration) (*fileset.ID, error) {
	return c.storage.CompactLevelBased(ctx, ids, defaultTTL, func(ctx context.Context, ids []fileset.ID, ttl time.Duration) (*fileset.ID, error) {
		return c.compact(ctx, taskDoer, ids, ttl)
	})
}

func (c *compactor) compact(ctx context.Context, taskDoer task.Doer, ids []fileset.ID, ttl time.Duration) (*fileset.ID, error) {
	ids, err := c.storage.Flatten(ctx, ids)
	if err != nil {
		return nil, err
	}
	var tasks []*CompactTask
	var taskLens []int
	if err := miscutil.LogStep(fmt.Sprintf("sharding %v file sets", len(ids)), func() error {
		var err error
		tasks, taskLens, err = c.createCompactTasks(ctx, taskDoer, ids)
		return err
	}); err != nil {
		return nil, err
	}
	var id *fileset.ID
	if err := c.storage.WithRenewer(ctx, ttl, func(ctx context.Context, renewer *fileset.Renewer) error {
		var compactResults []fileset.ID
		if err := miscutil.LogStep(fmt.Sprintf("compacting %v tasks", len(tasks)), func() error {
			var err error
			compactResults, err = c.processCompactTasks(ctx, taskDoer, renewer, tasks)
			return err
		}); err != nil {
			return err
		}
		var concatResults []fileset.ID
		if err := miscutil.LogStep(fmt.Sprintf("concatenating %v file sets", len(compactResults)), func() error {
			var err error
			concatResults, err = c.concat(ctx, taskDoer, renewer, compactResults, taskLens)
			return err
		}); err != nil {
			return err
		}
		if len(concatResults) == 1 {
			id = &concatResults[0]
			return nil
		}
		id, err = c.compact(ctx, taskDoer, concatResults, ttl)
		return err
	}); err != nil {
		return nil, err
	}
	return id, nil
}

func (c *compactor) createCompactTasks(ctx context.Context, taskDoer task.Doer, ids []fileset.ID) ([]*CompactTask, []int, error) {
	var inputs []*types.Any
	for start := 0; start < len(ids); start += int(c.maxFanIn) {
		end := start + int(c.maxFanIn)
		if end > len(ids) {
			end = len(ids)
		}
		ids := ids[start:end]
		input, err := serializeShardTask(&ShardTask{
			Inputs: fileset.IDsToHexStrings(ids),
		})
		if err != nil {
			return nil, nil, err
		}
		inputs = append(inputs, input)
	}
	results := make([][]*CompactTask, len(inputs))
	if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *types.Any, err error) error {
		if err != nil {
			return err
		}
		result, err := deserializeShardTaskResult(output)
		if err != nil {
			return err
		}
		results[i] = result.CompactTasks
		return nil
	}); err != nil {
		return nil, nil, err
	}
	var tasks []*CompactTask
	var taskLens []int
	for _, res := range results {
		taskLen := len(tasks)
		tasks = append(tasks, res...)
		taskLens = append(taskLens, len(tasks)-taskLen)
	}
	return tasks, taskLens, nil
}

func (c *compactor) processCompactTasks(ctx context.Context, taskDoer task.Doer, renewer *fileset.Renewer, tasks []*CompactTask) ([]fileset.ID, error) {
	inputs := make([]*types.Any, len(tasks))
	for i, task := range tasks {
		task := proto.Clone(task).(*CompactTask)
		input, err := serializeCompactTask(task)
		if err != nil {
			return nil, err
		}
		inputs[i] = input
	}
	results := make([]fileset.ID, len(inputs))
	if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *types.Any, err error) error {
		if err != nil {
			return err
		}
		result, err := deserializeCompactTaskResult(output)
		if err != nil {
			return err
		}
		id, err := fileset.ParseID(result.Id)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, *id); err != nil {
			return err
		}
		results[i] = *id
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *compactor) concat(ctx context.Context, taskDoer task.Doer, renewer *fileset.Renewer, ids []fileset.ID, taskLens []int) ([]fileset.ID, error) {
	var inputs []*types.Any
	for _, taskLen := range taskLens {
		var serInputs []string
		for _, id := range ids[:taskLen] {
			serInputs = append(serInputs, id.HexString())
		}
		ids = ids[taskLen:]
		input, err := serializeConcatTask(&ConcatTask{
			Inputs: serInputs,
		})
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}
	results := make([]fileset.ID, len(inputs))
	if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *types.Any, err error) error {
		if err != nil {
			return err
		}
		result, err := deserializeConcatTaskResult(output)
		if err != nil {
			return err
		}
		id, err := fileset.ParseID(result.Id)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, *id); err != nil {
			return err
		}
		results[i] = *id
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func compactionWorker(ctx context.Context, taskSource task.Source, storage *fileset.Storage) error {
	return backoff.RetryUntilCancel(ctx, func() error {
		err := taskSource.Iterate(ctx, func(ctx context.Context, input *types.Any) (*types.Any, error) {
			switch {
			case types.Is(input, &ShardTask{}):
				shardTask, err := deserializeShardTask(input)
				if err != nil {
					return nil, err
				}
				return processShardTask(ctx, storage, shardTask)
			case types.Is(input, &CompactTask{}):
				compactTask, err := deserializeCompactTask(input)
				if err != nil {
					return nil, err
				}
				return processCompactTask(ctx, storage, compactTask)
			case types.Is(input, &ConcatTask{}):
				concatTask, err := deserializeConcatTask(input)
				if err != nil {
					return nil, err
				}
				return processConcatTask(ctx, storage, concatTask)
			default:
				return nil, errors.Errorf("unrecognized any type (%v) in compaction worker", input.TypeUrl)
			}
		})
		return errors.EnsureStack(err)
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Printf("error in compaction worker: %v", err)
		return nil
	})
}

func processShardTask(ctx context.Context, storage *fileset.Storage, task *ShardTask) (*types.Any, error) {
	result := &ShardTaskResult{}
	if err := miscutil.LogStep("processing shard task", func() error {
		ids, err := fileset.HexStringsToIDs(task.Inputs)
		if err != nil {
			return err
		}
		return storage.Shard(ctx, ids, func(pathRange *index.PathRange) error {
			result.CompactTasks = append(result.CompactTasks, &CompactTask{
				Inputs: task.Inputs,
				PathRange: &PathRange{
					Lower: pathRange.Lower,
					Upper: pathRange.Upper,
				},
			})
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return serializeShardTaskResult(result)
}

func processCompactTask(ctx context.Context, storage *fileset.Storage, task *CompactTask) (*types.Any, error) {
	result := &CompactTaskResult{}
	if err := miscutil.LogStep("processing compact task", func() error {
		ids, err := fileset.HexStringsToIDs(task.Inputs)
		if err != nil {
			return err
		}
		pathRange := &index.PathRange{
			Lower: task.PathRange.Lower,
			Upper: task.PathRange.Upper,
		}
		id, err := storage.Compact(ctx, ids, defaultTTL, index.WithRange(pathRange))
		if err != nil {
			return err
		}
		result.Id = id.HexString()
		return err
	}); err != nil {
		return nil, err
	}
	return serializeCompactTaskResult(result)
}

func processConcatTask(ctx context.Context, storage *fileset.Storage, task *ConcatTask) (*types.Any, error) {
	result := &ConcatTaskResult{}
	if err := miscutil.LogStep("processing concat task", func() error {
		ids, err := fileset.HexStringsToIDs(task.Inputs)
		if err != nil {
			return err
		}
		id, err := storage.Concat(ctx, ids, defaultTTL)
		if err != nil {
			return err
		}
		result.Id = id.HexString()
		return err
	}); err != nil {
		return nil, err
	}
	return serializeConcatTaskResult(result)
}

func serializeShardTask(task *ShardTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeShardTask(taskAny *types.Any) (*ShardTask, error) {
	task := &ShardTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeShardTaskResult(task *ShardTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeShardTaskResult(taskAny *types.Any) (*ShardTaskResult, error) {
	task := &ShardTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeCompactTask(task *CompactTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeCompactTask(taskAny *types.Any) (*CompactTask, error) {
	task := &CompactTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeCompactTaskResult(res *CompactTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(res)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(res),
		Value:   data,
	}, nil
}

func deserializeCompactTaskResult(any *types.Any) (*CompactTaskResult, error) {
	res := &CompactTaskResult{}
	if err := types.UnmarshalAny(any, res); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return res, nil
}

func serializeConcatTask(task *ConcatTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeConcatTask(taskAny *types.Any) (*ConcatTask, error) {
	task := &ConcatTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeConcatTaskResult(task *ConcatTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeConcatTaskResult(taskAny *types.Any) (*ConcatTaskResult, error) {
	task := &ConcatTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}
