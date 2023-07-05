package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
)

// TODO: Move fan-in configuration to fileset.Storage.
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
	return c.storage.CompactLevelBased(ctx, ids, c.maxFanIn, defaultTTL, func(ctx context.Context, ids []fileset.ID, ttl time.Duration) (*fileset.ID, error) {
		return c.compact(ctx, taskDoer, ids, ttl)
	})
}

func (c *compactor) compact(ctx context.Context, taskDoer task.Doer, ids []fileset.ID, ttl time.Duration) (*fileset.ID, error) {
	var tasks []*CompactTask
	if err := log.LogStep(ctx, "shardFileSet", func(ctx context.Context) error {
		var err error
		tasks, err = c.createCompactTasks(ctx, taskDoer, ids)
		return err
	}, zap.Int("filesets", len(ids))); err != nil {
		return nil, err
	}
	var id *fileset.ID
	if err := c.storage.WithRenewer(ctx, ttl, func(ctx context.Context, renewer *fileset.Renewer) error {
		var results []fileset.ID
		if err := log.LogStep(ctx, "compactTasks", func(ctx context.Context) error {
			var err error
			results, err = c.processCompactTasks(ctx, taskDoer, renewer, tasks)
			return err
		}, zap.Int("tasks", len(tasks))); err != nil {
			return err
		}
		return log.LogStep(ctx, "concatenateFileSets", func(ctx context.Context) error {
			var err error
			id, err = c.concat(ctx, taskDoer, renewer, results)
			return err
		}, zap.Int("filesets", len(results)))
	}); err != nil {
		return nil, err
	}
	return id, nil
}

func (c *compactor) createCompactTasks(ctx context.Context, taskDoer task.Doer, ids []fileset.ID) ([]*CompactTask, error) {
	fs, err := c.storage.Open(ctx, ids)
	if err != nil {
		return nil, err
	}
	shards, err := fs.Shards(ctx, index.WithShardConfig(c.storage.ShardConfig()))
	if err != nil {
		return nil, err
	}
	for _, shard := range shards {
		log.Info(ctx, "created file set shard", zap.Stringer("range", shard))
	}
	var inputs []*anypb.Any
	for _, shard := range shards {
		input, err := serializeShardTask(&ShardTask{
			Inputs: fileset.IDsToHexStrings(ids),
			PathRange: &PathRange{
				Lower: shard.Lower,
				Upper: shard.Upper,
			},
		})
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}
	results := make([][]*CompactTask, len(inputs))
	if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
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
		return nil, err
	}
	var compactTasks []*CompactTask
	for _, result := range results {
		for _, task := range result {
			log.Info(ctx, "created compaction shard", zap.Stringer("pathRange", task.PathRange))
		}
		compactTasks = append(compactTasks, result...)
	}
	return compactTasks, nil
}

func (c *compactor) processCompactTasks(ctx context.Context, taskDoer task.Doer, renewer *fileset.Renewer, tasks []*CompactTask) ([]fileset.ID, error) {
	inputs := make([]*anypb.Any, len(tasks))
	for i, task := range tasks {
		task := proto.Clone(task).(*CompactTask)
		input, err := serializeCompactTask(task)
		if err != nil {
			return nil, err
		}
		inputs[i] = input
	}
	results := make([]fileset.ID, len(inputs))
	if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
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

func (c *compactor) concat(ctx context.Context, taskDoer task.Doer, renewer *fileset.Renewer, ids []fileset.ID) (*fileset.ID, error) {
	var serInputs []string
	for _, id := range ids {
		serInputs = append(serInputs, id.HexString())
	}
	input, err := serializeConcatTask(&ConcatTask{
		Inputs: serInputs,
	})
	if err != nil {
		return nil, err
	}
	output, err := task.DoOne(ctx, taskDoer, input)
	if err != nil {
		return nil, err
	}
	result, err := deserializeConcatTaskResult(output)
	if err != nil {
		return nil, err
	}
	id, err := fileset.ParseID(result.Id)
	if err != nil {
		return nil, err
	}
	if err := renewer.Add(ctx, *id); err != nil {
		return nil, err
	}
	return id, nil
}

func (c *compactor) Validate(ctx context.Context, taskDoer task.Doer, id fileset.ID) (string, int64, error) {
	fs, err := c.storage.Open(ctx, []fileset.ID{id})
	if err != nil {
		return "", 0, err
	}
	shards, err := fs.Shards(ctx, index.WithShardConfig(c.storage.ShardConfig()))
	if err != nil {
		return "", 0, err
	}
	var inputs []*anypb.Any
	for _, shard := range shards {
		input, err := serializeValidateTask(&ValidateTask{
			Id: id.HexString(),
			PathRange: &PathRange{
				Lower: shard.Lower,
				Upper: shard.Upper,
			},
		})
		if err != nil {
			return "", 0, err
		}
		inputs = append(inputs, input)
	}
	// TODO: Collect all of the errors or fail fast?
	results := make([]*ValidateTaskResult, len(inputs))
	if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
		if err != nil {
			return err
		}
		results[i], err = deserializeValidateTaskResult(output)
		return err
	}); err != nil {
		return "", 0, err
	}
	var errStr string
	var size int64
	for i, result := range results {
		if errStr == "" && i != 0 {
			errStr = checkIndex(results[i-1].Last, result.First)
		}
		if errStr == "" {
			errStr = result.Error
		}
		size += result.SizeBytes
	}
	return errStr, size, nil
}

func compactionWorker(ctx context.Context, taskSource task.Source, storage *fileset.Storage) error {
	log.Info(ctx, "running compaction worker")
	return backoff.RetryUntilCancel(ctx, func() error {
		err := taskSource.Iterate(ctx, func(ctx context.Context, input *anypb.Any) (*anypb.Any, error) {
			switch {
			case input.MessageIs(&ShardTask{}):
				shardTask, err := deserializeShardTask(input)
				if err != nil {
					return nil, err
				}
				return processShardTask(ctx, storage, shardTask)
			case input.MessageIs(&CompactTask{}):
				compactTask, err := deserializeCompactTask(input)
				if err != nil {
					return nil, err
				}
				return processCompactTask(ctx, storage, compactTask)
			case input.MessageIs(&ConcatTask{}):
				concatTask, err := deserializeConcatTask(input)
				if err != nil {
					return nil, err
				}
				return processConcatTask(ctx, storage, concatTask)
			case input.MessageIs(&ValidateTask{}):
				validateTask, err := deserializeValidateTask(input)
				if err != nil {
					return nil, err
				}
				return processValidateTask(ctx, storage, validateTask)
			default:
				return nil, errors.Errorf("unrecognized any type (%v) in compaction worker", input.TypeUrl)
			}
		})
		return errors.EnsureStack(err)
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Info(ctx, "error in compaction worker", zap.Error(err))
		return nil
	})
}

func processShardTask(ctx context.Context, storage *fileset.Storage, task *ShardTask) (*anypb.Any, error) {
	result := &ShardTaskResult{}
	if err := log.LogStep(ctx, "processing shard task", func(ctx context.Context) error {
		ids, err := fileset.HexStringsToIDs(task.Inputs)
		if err != nil {
			return err
		}
		pathRange := &index.PathRange{
			Lower: task.PathRange.Lower,
			Upper: task.PathRange.Upper,
		}
		shards, err := storage.Shard(ctx, ids, pathRange)
		if err != nil {
			return err
		}
		for _, shard := range shards {
			result.CompactTasks = append(result.CompactTasks, &CompactTask{
				Inputs: task.Inputs,
				PathRange: &PathRange{
					Lower: shard.Lower,
					Upper: shard.Upper,
				},
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return serializeShardTaskResult(result)
}

func processCompactTask(ctx context.Context, storage *fileset.Storage, task *CompactTask) (*anypb.Any, error) {
	result := &CompactTaskResult{}
	if err := log.LogStep(ctx, "processCompactTask", func(ctx context.Context) error {
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

func processConcatTask(ctx context.Context, storage *fileset.Storage, task *ConcatTask) (*anypb.Any, error) {
	result := &ConcatTaskResult{}
	if err := log.LogStep(ctx, "processConcatTask", func(ctx context.Context) error {
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

// TODO(2.0 optional): Improve the performance of this by doing a logarithmic lookup per new file,
// rather than a linear scan through all of the files.
func processValidateTask(ctx context.Context, storage *fileset.Storage, task *ValidateTask) (*anypb.Any, error) {
	result := &ValidateTaskResult{}
	if err := log.LogStep(ctx, "validateTask", func(ctx context.Context) error {
		id, err := fileset.ParseID(task.Id)
		if err != nil {
			return err
		}
		fs, err := storage.Open(ctx, []fileset.ID{*id})
		if err != nil {
			return err
		}
		opts := []index.Option{
			index.WithRange(&index.PathRange{
				Lower: task.PathRange.Lower,
				Upper: task.PathRange.Upper,
			}),
		}
		var prev *index.Index
		if err := fs.Iterate(ctx, func(f fileset.File) error {
			idx := f.Index()
			if result.First == nil {
				result.First = idx
			}
			result.Last = idx
			if result.Error == "" {
				result.Error = checkIndex(prev, idx)
			}
			prev = idx
			result.SizeBytes += index.SizeBytes(idx)
			return nil
		}, opts...); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return serializeValidateTaskResult(result)
}

func checkIndex(prev, curr *index.Index) string {
	if prev == nil || curr == nil {
		return ""
	}
	if curr.Path == prev.Path {
		return fmt.Sprintf("duplicate path output by different datums (%v from %v and %v from %v)", prev.Path, prev.File.Datum, curr.Path, curr.File.Datum)
	}
	if strings.HasPrefix(curr.Path, prev.Path+"/") {
		return fmt.Sprintf("file / directory path collision (%v)", curr.Path)
	}
	return ""
}

func serializeShardTask(task *ShardTask) (*anypb.Any, error) { return anypb.New(task) }

func deserializeShardTask(taskAny *anypb.Any) (*ShardTask, error) {
	task := &ShardTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeShardTaskResult(task *ShardTaskResult) (*anypb.Any, error) { return anypb.New(task) }

func deserializeShardTaskResult(taskAny *anypb.Any) (*ShardTaskResult, error) {
	task := &ShardTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeCompactTask(task *CompactTask) (*anypb.Any, error) { return anypb.New(task) }

func deserializeCompactTask(taskAny *anypb.Any) (*CompactTask, error) {
	task := &CompactTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeCompactTaskResult(res *CompactTaskResult) (*anypb.Any, error) { return anypb.New(res) }

func deserializeCompactTaskResult(any *anypb.Any) (*CompactTaskResult, error) {
	res := &CompactTaskResult{}
	if err := any.UnmarshalTo(res); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return res, nil
}

func serializeConcatTask(task *ConcatTask) (*anypb.Any, error) { return anypb.New(task) }

func deserializeConcatTask(taskAny *anypb.Any) (*ConcatTask, error) {
	task := &ConcatTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeConcatTaskResult(task *ConcatTaskResult) (*anypb.Any, error) { return anypb.New(task) }

func deserializeConcatTaskResult(taskAny *anypb.Any) (*ConcatTaskResult, error) {
	task := &ConcatTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeValidateTask(task *ValidateTask) (*anypb.Any, error) { return anypb.New(task) }

func deserializeValidateTask(taskAny *anypb.Any) (*ValidateTask, error) {
	task := &ValidateTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeValidateTaskResult(task *ValidateTaskResult) (*anypb.Any, error) {
	return anypb.New(task)
}

func deserializeValidateTaskResult(taskAny *anypb.Any) (*ValidateTaskResult, error) {
	task := &ValidateTaskResult{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}
