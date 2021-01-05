package fileset

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

// IsCompacted returns true if the fileset is already in compacted form.
func (s *Storage) IsCompacted(ctx context.Context, id ID, ttl time.Duration) (bool, error) {
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
		layers, err := s.collectLayers(ctx, ids)
		if err != nil {
			return false, err
		}
		return isCompacted(s.levelFactor, layers), nil
	default:
		return true, nil
	}
}

func (s *Storage) collectLayers(ctx context.Context, ids []ID) ([]*Primitive, error) {
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

// CompactStats contains information about what was compacted.
type CompactStats struct {
	OutputSize int64
}

// Compact compacts a set of filesets into an output fileset.
func (s *Storage) Compact(ctx context.Context, ids []string, ttl time.Duration, opts ...index.Option) (*ID, *CompactStats, error) {
	var size int64
	w := s.newWriter(ctx, WithTTL(ttl), WithIndexCallback(func(idx *index.Index) error {
		size += index.SizeBytes(idx)
		return nil
	}))
	fs, err := s.Open(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	if err := CopyFiles(ctx, w, fs, true); err != nil {
		return nil, nil, err
	}
	id, err := w.Close()
	if err != nil {
		return nil, nil, err
	}
	return id, &CompactStats{OutputSize: size}, nil
}

// // CompactSpec specifies the input and output for a compaction operation.
// type CompactSpec struct {
// 	Output string
// 	Input  []string
// }

// // CompactSpec returns a compaction specification that determines the input filesets (the diff file set and potentially
// // compacted filesets) and output fileset.
// func (s *Storage) CompactSpec(ctx context.Context, fileSet string, compactedFileSet ...string) (*CompactSpec, error) {
// 	if len(compactedFileSet) > 1 {
// 		return nil, errors.Errorf("multiple compacted FileSets")
// 	}
// 	spec, err := s.compactSpec(ctx, fileSet, compactedFileSet...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return spec, nil
// }

// func (s *Storage) compactSpec(ctx context.Context, fileSet string, compactedFileSet ...string) (ret *CompactSpec, retErr error) {
// 	md, err := s.store.Get(ctx, path.Join(fileSet, Diff))
// 	if err != nil {
// 		return nil, err
// 	}
// 	size := md.SizeBytes
// 	spec := &CompactSpec{
// 		Input: []string{path.Join(fileSet, Diff)},
// 	}
// 	var level int
// 	// Handle first commit being compacted.
// 	if len(compactedFileSet) == 0 {
// 		for size > s.levelSize(level) {
// 			level++
// 		}
// 		spec.Output = path.Join(fileSet, Compacted, levelName(level))
// 		return spec, nil
// 	}
// 	// While we can't fit it all in the current level
// 	for {
// 		levelPath := path.Join(compactedFileSet[0], Compacted, levelName(level))
// 		md, err := s.store.Get(ctx, levelPath)
// 		if err != nil {
// 			if err != ErrPathNotExists {
// 				return nil, err
// 			}
// 		} else {
// 			spec.Input = append(spec.Input, levelPath)
// 			size += md.SizeBytes
// 		}
// 		if size <= s.levelSize(level) {
// 			break
// 		}
// 		level++
// 	}
// 	// Now we know the output level
// 	spec.Output = path.Join(fileSet, Compacted, levelName(level))
// 	// Copy the other levels that may exist
// 	if err := s.store.Walk(ctx, path.Join(compactedFileSet[0], Compacted), func(src string) error {
// 		lName := path.Base(src)
// 		l, err := parseLevel(lName)
// 		if err != nil {
// 			return err
// 		}
// 		if l > level {
// 			dst := path.Join(fileSet, Compacted, levelName(l))
// 			if err := copyPath(ctx, s.store, s.store, src, dst, s.tracker, 0); err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	}); err != nil {
// 		return nil, err
// 	}
// 	// Inputs should be ordered with priority from least to greatest.
// 	for i := 0; i < len(spec.Input)/2; i++ {
// 		spec.Input[i], spec.Input[len(spec.Input)-1-i] = spec.Input[len(spec.Input)-1-i], spec.Input[i]
// 	}
// 	return spec, nil
// }

// func (s *Storage) levelSize(i int) int64 {
// 	return s.levelZeroSize * int64(math.Pow(float64(s.levelSizeBase), float64(i)))
// }

// const subFileSetFmt = "%020d"
// const levelFmt = "level_" + subFileSetFmt

// func levelName(i int) string {
// 	return fmt.Sprintf(levelFmt, i)
// }

// func parseLevel(x string) (int, error) {
// 	var y int
// 	_, err := fmt.Sscanf(x, levelFmt, &y)
// 	return y, err
// }
