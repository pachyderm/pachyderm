package fileset

import (
	"context"
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

// IsCompacted returns true if the filesets are already in compacted form.
func (s *Storage) IsCompacted(ctx context.Context, ids []ID) (bool, error) {
	prims, err := s.flattenPrimitives(ctx, ids)
	if err != nil {
		return false, err
	}
	return isCompacted(s.compactionConfig, prims), nil
}

func isCompacted(config *CompactionConfig, prims []*Primitive) bool {
	return config.FixedDelay > int64(len(prims)) || indexOfCompacted(config.LevelFactor, prims) == len(prims)
}

// indexOfCompacted returns the last value of i for which the "compacted relationship"
// is maintained for all layers[:i+1]
// the "compacted relationship" is defined as leftSize >= (rightSize * factor)
// If there is an element at i+1. It will be the first element which does not satisfy
// the compacted relationship with i.
func indexOfCompacted(factor int64, inputs []*Primitive) int {
	l := len(inputs)
	for i := 0; i < l-1; i++ {
		leftSize := inputs[i].SizeBytes
		rightSize := inputs[i+1].SizeBytes
		if leftSize < rightSize*factor {
			return i
		}
	}
	return l
}

// Compact compacts the contents of ids into a new fileset with the specified ttl and returns the ID.
// Compact always returns the ID of a primitive fileset.
// Compact does not renew ids.
// It is the responsibility of the caller to renew ids.  In some cases they may be permanent and not require renewal.
func (s *Storage) Compact(ctx context.Context, ids []ID, ttl time.Duration, opts ...index.Option) (*ID, error) {
	var size int64
	w := s.newWriter(ctx, WithTTL(ttl), WithIndexCallback(func(idx *index.Index) error {
		size += index.SizeBytes(idx)
		return nil
	}))
	fs, err := s.Open(ctx, ids, opts...)
	if err != nil {
		return nil, err
	}
	if err := CopyFiles(ctx, w, fs, true); err != nil {
		return nil, err
	}
	return w.Close()
}

// CompactionTask contains everything needed to perform the smallest unit of compaction
type CompactionTask struct {
	Inputs    []ID
	PathRange *index.PathRange
}

// Compactor performs compaction
type Compactor interface {
	Compact(ctx context.Context, ids []ID, ttl time.Duration) (*ID, error)
}

// CompactionWorker can perform CompactionTasks
type CompactionWorker func(ctx context.Context, spec CompactionTask) (*ID, error)

// CompactionBatchWorker can perform batches of CompactionTasks
type CompactionBatchWorker func(ctx context.Context, renewer *Renewer, spec []CompactionTask) ([]ID, error)

// DistributedCompactor performs compaction by fanning out tasks to workers.
type DistributedCompactor struct {
	s          *Storage
	maxFanIn   int
	workerFunc CompactionBatchWorker
}

// NewDistributedCompactor returns a DistributedCompactor which will compact by fanning out
// work to workerFunc, while respecting maxFanIn
// TODO: change this to CompactionWorker after work package changes.
func NewDistributedCompactor(s *Storage, maxFanIn int, workerFunc CompactionBatchWorker) *DistributedCompactor {
	return &DistributedCompactor{
		s:          s,
		maxFanIn:   maxFanIn,
		workerFunc: workerFunc,
	}
}

// Compact compacts the contents of ids into a new fileset with the specified ttl and returns the ID.
// Compact always returns the ID of a primitive fileset.
func (c *DistributedCompactor) Compact(ctx context.Context, ids []ID, ttl time.Duration) (_ *ID, retErr error) {
	var err error
	ids, err = c.s.Flatten(ctx, ids)
	if err != nil {
		return nil, err
	}
	var tasks []CompactionTask
	var taskLens []int
	for start := 0; start < len(ids); start += c.maxFanIn {
		end := start + c.maxFanIn
		if end > len(ids) {
			end = len(ids)
		}
		ids := ids[start:end]
		taskLen := len(tasks)
		if err := miscutil.LogStep(fmt.Sprintf("sharding %v file sets", len(ids)), func() error {
			return c.s.Shard(ctx, ids, func(pathRange *index.PathRange) error {
				tasks = append(tasks, CompactionTask{
					Inputs:    ids,
					PathRange: pathRange,
				})
				return nil
			})
		}); err != nil {
			return nil, err
		}
		taskLens = append(taskLens, len(tasks)-taskLen)
	}
	var id *ID
	if err := c.s.WithRenewer(ctx, ttl, func(ctx context.Context, renewer *Renewer) error {
		var resultIds []ID
		if err := miscutil.LogStep(fmt.Sprintf("compacting %v tasks", len(tasks)), func() error {
			results, err := c.workerFunc(ctx, renewer, tasks)
			if err != nil {
				return err
			}
			if len(results) != len(tasks) {
				return errors.Errorf("results are a different length than tasks")
			}
			for _, taskLen := range taskLens {
				id, err := c.s.Concat(ctx, results[:taskLen], ttl)
				if err != nil {
					return err
				}
				if err := renewer.Add(ctx, *id); err != nil {
					return err
				}
				resultIds = append(resultIds, *id)
				results = results[taskLen:]
			}
			return nil
		}); err != nil {
			return err
		}
		if len(resultIds) == 1 {
			id = &resultIds[0]
			return nil
		}
		id, err = c.Compact(ctx, resultIds, ttl)
		return err
	}); err != nil {
		return nil, err
	}
	return id, nil
}

// CompactCallback is the standard callback signature for a compaction operation.
type CompactCallback func(context.Context, []ID, time.Duration) (*ID, error)

// CompactLevelBased performs a level-based compaction on the passed in filesets.
func (s *Storage) CompactLevelBased(ctx context.Context, ids []ID, ttl time.Duration, compact CompactCallback) (*ID, error) {
	ids, err := s.Flatten(ctx, ids)
	if err != nil {
		return nil, err
	}
	prims, err := s.getPrimitives(ctx, ids)
	if err != nil {
		return nil, err
	}
	if isCompacted(s.compactionConfig, prims) {
		return s.Compose(ctx, ids, ttl)
	}
	i := indexOfCompacted(s.compactionConfig.LevelFactor, prims)
	var id *ID
	if err := miscutil.LogStep(fmt.Sprintf("compacting %v levels out of %v", len(ids)-i, len(ids)), func() error {
		id, err = compact(ctx, ids[i:], ttl)
		return err
	}); err != nil {
		return nil, err
	}
	return s.CompactLevelBased(ctx, append(ids[:i], *id), ttl, compact)
}
