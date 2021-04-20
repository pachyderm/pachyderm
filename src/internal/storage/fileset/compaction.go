package fileset

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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

// CompactionWorker can perform CompactionTasks
type CompactionWorker func(ctx context.Context, spec CompactionTask) (*ID, error)

// CompactionBatchWorker can perform batches of CompactionTasks
type CompactionBatchWorker func(ctx context.Context, spec []CompactionTask) ([]ID, error)

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
func (c *DistributedCompactor) Compact(ctx context.Context, ids []ID, ttl time.Duration) (*ID, error) {
	var err error
	ids, err = c.s.Flatten(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(ids) <= c.maxFanIn {
		return c.shardedCompact(ctx, ids, ttl)
	}
	childSize := c.maxFanIn
	for len(ids)/childSize > c.maxFanIn {
		childSize *= c.maxFanIn
	}
	results := []ID{}
	for start := 0; start < len(ids); start += childSize {
		end := start + childSize
		if end > len(ids) {
			end = len(ids)
		}
		id, err := c.Compact(ctx, ids[start:end], ttl)
		if err != nil {
			return nil, err
		}
		results = append(results, *id)
	}
	return c.Compact(ctx, results, ttl)
}

func (c *DistributedCompactor) shardedCompact(ctx context.Context, ids []ID, ttl time.Duration) (*ID, error) {
	var tasks []CompactionTask
	if err := c.s.Shard(ctx, ids, func(pathRange *index.PathRange) error {
		tasks = append(tasks, CompactionTask{
			Inputs:    ids,
			PathRange: pathRange,
		})
		return nil
	}); err != nil {
		return nil, err
	}
	results, err := c.workerFunc(ctx, tasks)
	if err != nil {
		return nil, err
	}
	if len(results) != len(tasks) {
		return nil, errors.Errorf("results are a different length than tasks")
	}
	return c.s.Concat(ctx, results, ttl)
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
	id, err := compact(ctx, ids[i:], ttl)
	if err != nil {
		return nil, err
	}
	return s.CompactLevelBased(ctx, append(ids[:i], *id), ttl, compact)
}
