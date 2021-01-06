package fileset

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

// IsCompacted returns true if the fileset is already in compacted form.
func (s *Storage) IsCompacted(ctx context.Context, id ID) (bool, error) {
	md, err := s.store.Get(ctx, id)
	if err != nil {
		return false, err
	}
	switch x := md.Value.(type) {
	case *Metadata_Composite:
		ids, err := s.Flatten(ctx, x.Composite.Layers)
		if err != nil {
			return false, err
		}
		layers, err := s.getPrimitiveBatch(ctx, ids)
		if err != nil {
			return false, err
		}
		return isCompacted(s.levelFactor, layers), nil
	case *Metadata_Primitive:
		return true, nil
	default:
		// TODO: should it be?
		return false, errors.Errorf("IsCompacted is not defined for empty filesets")
	}
}

func (s *Storage) getPrimitiveBatch(ctx context.Context, ids []ID) ([]*Primitive, error) {
	var layers []*Primitive
	for _, id := range ids {
		prim, err := s.getPrimitive(ctx, id)
		if err != nil {
			return nil, err
		}
		layers = append(layers, prim)
	}
	return layers, nil
}

func isCompacted(factor int64, layers []*Primitive) bool {
	return indexOfCompacted(factor, layers) == len(layers)
}

// indexOfCompacted returns the last value of i for which the "compacted relationship"
// is maintained for all layers[:i]
// the "compacted relationship" is defined as leftSize >= (rightSize * factor)
// If there is an element at i. It will be the first element which does not satisfy
// the compacted relationship with i-1.
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

// Compact compacts a set of filesets into an output fileset.
func (s *Storage) Compact(ctx context.Context, ids []ID, ttl time.Duration, opts ...index.Option) (*ID, error) {
	ids, err := s.Flatten(ctx, ids)
	if err != nil {
		return nil, err
	}
	inputs, err := s.getPrimitiveBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	i := indexOfCompacted(s.levelFactor, inputs)
	if i == len(inputs) {
		return s.Compose(ctx, ids, ttl)
	}
	// merge everything from i-1 onward into a single layer
	merged, err := s.Merge(ctx, ids[i-1:], ttl)
	if err != nil {
		return nil, err
	}
	// replace everything from i-1 onward with the merged version
	ids2 := append(ids[:i-1], *merged)
	return s.Compact(ctx, ids2, ttl)
}

type CompactionWorker func(ctx context.Context, ids []ID) ([]ID, error)

func CompactDistributed(ctx context.Context, s *Storage, workerFunc CompactionWorker, maxFanIn int, ids []ID, ttl time.Duration) (ID, error) {
	panic("not implemented")
}
