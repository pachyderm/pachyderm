package server

import (
	"context"
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

func (c *compactor) Compact(ctx context.Context, taskDoer task.Doer, handles []*fileset.Handle, ttl time.Duration) (*fileset.Handle, error) {
	return c.storage.CompactLevelBased(ctx, handles, c.maxFanIn, defaultTTL, func(ctx context.Context, handles []*fileset.Handle, ttl time.Duration) (*fileset.Handle, error) {
		return c.compact(ctx, taskDoer, handles, ttl)
	})
}

func (c *compactor) compact(ctx context.Context, taskDoer task.Doer, handles []*fileset.Handle, ttl time.Duration) (*fileset.Handle, error) {
	var tasks []*CompactTask
	if err := log.LogStep(ctx, "shardFileSet", func(ctx context.Context) error {
		var err error
		tasks, err = c.createCompactTasks(ctx, taskDoer, handles)
		return err
	}, zap.Int("filesets", len(handles))); err != nil {
		return nil, err
	}
	var handle *fileset.Handle
	if err := c.storage.WithRenewer(ctx, ttl, func(ctx context.Context, renewer *fileset.Renewer) error {
		var results []*fileset.Handle
		if err := log.LogStep(ctx, "compactTasks", func(ctx context.Context) error {
			var err error
			results, err = c.processCompactTasks(ctx, taskDoer, renewer, tasks)
			return err
		}, zap.Int("tasks", len(tasks))); err != nil {
			return err
		}
		return log.LogStep(ctx, "concatenateFileSets", func(ctx context.Context) error {
			var err error
			handle, err = c.concat(ctx, taskDoer, renewer, results)
			return err
		}, zap.Int("filesets", len(results)))
	}); err != nil {
		return nil, err
	}
	return handle, nil
}

func (c *compactor) createCompactTasks(ctx context.Context, taskDoer task.Doer, handles []*fileset.Handle) ([]*CompactTask, error) {
	fs, err := c.storage.Open(ctx, handles)
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
			Inputs: fileset.HandlesToHexStrings(handles),
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

func (c *compactor) processCompactTasks(ctx context.Context, taskDoer task.Doer, renewer *fileset.Renewer, tasks []*CompactTask) ([]*fileset.Handle, error) {
	inputs := make([]*anypb.Any, len(tasks))
	for i, task := range tasks {
		task := proto.Clone(task).(*CompactTask)
		input, err := serializeCompactTask(task)
		if err != nil {
			return nil, err
		}
		inputs[i] = input
	}
	results := make([]*fileset.Handle, len(inputs))
	if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *anypb.Any, err error) error {
		if err != nil {
			return err
		}
		result, err := deserializeCompactTaskResult(output)
		if err != nil {
			return err
		}
		handle, err := fileset.ParseHandle(result.Handle)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, handle); err != nil {
			return err
		}
		results[i] = handle
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *compactor) concat(ctx context.Context, taskDoer task.Doer, renewer *fileset.Renewer, handles []*fileset.Handle) (*fileset.Handle, error) {
	var serInputs []string
	for _, handle := range handles {
		serInputs = append(serInputs, handle.HexString())
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
	handle, err := fileset.ParseHandle(result.Handle)
	if err != nil {
		return nil, err
	}
	if err := renewer.Add(ctx, handle); err != nil {
		return nil, err
	}
	return handle, nil
}

func (c *compactor) Validate(ctx context.Context, taskDoer task.Doer, handle *fileset.Handle) (string, int64, error) {
	fs, err := c.storage.Open(ctx, []*fileset.Handle{handle})
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
			Handle: handle.HexString(),
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
			if err := fileset.CheckIndex(results[i-1].Last, result.First); err != nil {
				errStr = err.Error()
			}
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
		handles, err := fileset.HexStringsToHandles(task.Inputs)
		if err != nil {
			return err
		}
		pathRange := &index.PathRange{
			Lower: task.PathRange.Lower,
			Upper: task.PathRange.Upper,
		}
		shards, err := storage.Shard(ctx, handles, pathRange)
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
		handles, err := fileset.HexStringsToHandles(task.Inputs)
		if err != nil {
			return err
		}
		pathRange := &index.PathRange{
			Lower: task.PathRange.Lower,
			Upper: task.PathRange.Upper,
		}
		handle, err := storage.Compact(ctx, handles, defaultTTL, index.WithRange(pathRange))
		if err != nil {
			return err
		}
		result.Handle = handle.HexString()
		return err
	}); err != nil {
		return nil, err
	}
	return serializeCompactTaskResult(result)
}

func processConcatTask(ctx context.Context, storage *fileset.Storage, task *ConcatTask) (*anypb.Any, error) {
	result := &ConcatTaskResult{}
	if err := log.LogStep(ctx, "processConcatTask", func(ctx context.Context) error {
		handles, err := fileset.HexStringsToHandles(task.Inputs)
		if err != nil {
			return err
		}
		handle, err := storage.Concat(ctx, handles, defaultTTL)
		if err != nil {
			return err
		}
		result.Handle = handle.HexString()
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
		handle, err := fileset.ParseHandle(task.Handle)
		if err != nil {
			return err
		}
		fs, err := storage.Open(ctx, []*fileset.Handle{handle})
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
				if err := fileset.CheckIndex(prev, idx); err != nil {
					result.Error = err.Error()
				}
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
