package fileset

import (
	"context"
	"fmt"
	"time"

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
			var size int64
			for j := i; j < l; j++ {
				size += inputs[j].SizeBytes
			}
			for ; i > 0; i-- {
				if inputs[i-1].SizeBytes >= size*factor {
					return i
				}
				size += inputs[i-1].SizeBytes
			}
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
