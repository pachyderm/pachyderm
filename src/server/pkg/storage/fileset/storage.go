package fileset

import (
	"context"
	"fmt"
	"io"
	"math"
	"path"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

const (
	// (bryce) Not sure if these are the tags we should use, but the header and padding tag should show up before and after respectively in the
	// lexicographical ordering of tags.
	headerTag  = ""
	paddingTag = "~"
	prefix     = "pfs"
	// DefaultMemoryThreshold is the default for the memory threshold that must
	// be met before a file set part is serialized (excluding close).
	DefaultMemoryThreshold = 1024 * chunk.MB
	// DefaultShardThreshold is the default for the size threshold that must
	// be met before a shard is created by the shard function.
	DefaultShardThreshold = 1024 * chunk.MB
	// DefaultLevelZeroSize is the default size for level zero in the compacted
	// representation of a file set.
	DefaultLevelZeroSize = 1 * chunk.MB
	// DefaultLevelSizeBase is the default base of the exponential growth function
	// for level sizes in the compacted representation of a file set.
	DefaultLevelSizeBase = 10
	// Diff is the suffix of a path that points to the diff of the prefix.
	Diff = "diff"
	// Compacted is the suffix of a path that points to the compaction of the prefix.
	Compacted = "compacted"
)

// Storage is the abstraction that manages fileset storage.
type Storage struct {
	objC                         obj.Client
	chunks                       *chunk.Storage
	memThreshold, shardThreshold int64
	levelZeroSize                int64
	levelSizeBase                int
}

// NewStorage creates a new Storage.
func NewStorage(objC obj.Client, chunks *chunk.Storage, opts ...StorageOption) *Storage {
	s := &Storage{
		objC:           objC,
		chunks:         chunks,
		memThreshold:   DefaultMemoryThreshold,
		shardThreshold: DefaultShardThreshold,
		levelZeroSize:  DefaultLevelZeroSize,
		levelSizeBase:  DefaultLevelSizeBase,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ChunkStorage returns the underlying chunk storage instance for this storage instance.
func (s *Storage) ChunkStorage() *chunk.Storage {
	return s.chunks
}

// New creates a new in-memory fileset.
func (s *Storage) New(ctx context.Context, fileSet, tag string, opts ...Option) *FileSet {
	fileSet = applyPrefix(fileSet)
	return newFileSet(ctx, s, fileSet, s.memThreshold, tag, opts...)
}

func (s *Storage) newWriter(ctx context.Context, fileSet string, indexFunc func(*index.Index) error) *Writer {
	fileSet = applyPrefix(fileSet)
	return newWriter(ctx, s.objC, s.chunks, fileSet, indexFunc)
}

// (bryce) expose some notion of read ahead (read a certain number of chunks in parallel).
// this will be necessary to speed up reading large files.
func (s *Storage) newReader(ctx context.Context, fileSet string, opts ...index.Option) *Reader {
	fileSet = applyPrefix(fileSet)
	return newReader(ctx, s.objC, s.chunks, fileSet, opts...)
}

// NewMergeReader returns a merge reader for a set for filesets.
func (s *Storage) NewMergeReader(ctx context.Context, fileSets []string, opts ...index.Option) (*MergeReader, error) {
	fileSets = applyPrefixes(fileSets)
	var rs []*Reader
	for _, fileSet := range fileSets {
		if err := s.objC.Walk(ctx, fileSet, func(name string) error {
			rs = append(rs, s.newReader(ctx, name, opts...))
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return newMergeReader(rs), nil
}

// ResolveIndexes resolves index entries that are spread across multiple filesets.
func (s *Storage) ResolveIndexes(ctx context.Context, fileSets []string, f func(*index.Index) error, opts ...index.Option) error {
	mr, err := s.NewMergeReader(ctx, fileSets, opts...)
	if err != nil {
		return err
	}
	w := s.newWriter(ctx, "", f)
	if err := mr.WriteTo(w); err != nil {
		return err
	}
	return w.Close()
}

// Shard shards the merge of the file sets with the passed in prefix into file ranges.
// (bryce) this should be extended to be more configurable (different criteria
// for creating shards).
func (s *Storage) Shard(ctx context.Context, fileSets []string, shardFunc ShardFunc) error {
	fileSets = applyPrefixes(fileSets)
	mr, err := s.NewMergeReader(ctx, fileSets)
	if err != nil {
		return err
	}
	return shard(mr, s.shardThreshold, shardFunc)
}

// Compact compacts a set of filesets into an output fileset.
func (s *Storage) Compact(ctx context.Context, outputFileSet string, inputFileSets []string, opts ...index.Option) error {
	outputFileSet = applyPrefix(outputFileSet)
	inputFileSets = applyPrefixes(inputFileSets)
	w := s.newWriter(ctx, outputFileSet, nil)
	mr, err := s.NewMergeReader(ctx, inputFileSets, opts...)
	if err != nil {
		return err
	}
	if err := mr.WriteTo(w); err != nil {
		return err
	}
	return w.Close()
}

// CompactSpec specifies the input and output for a compaction operation.
type CompactSpec struct {
	Output string
	Input  []string
}

// CompactSpec returns a compaction specification that determines the input filesets (the diff file set and potentially
// compacted filesets) and output fileset.
func (s *Storage) CompactSpec(ctx context.Context, fileSet string, compactedFileSet ...string) (*CompactSpec, error) {
	fileSet = applyPrefix(fileSet)
	idx, err := index.GetTopLevelIndex(ctx, s.objC, path.Join(fileSet, Diff))
	if err != nil {
		return nil, err
	}
	size := idx.SizeBytes
	spec := &CompactSpec{
		Input: []string{path.Join(fileSet, Diff)},
	}
	var level int
	// Handle first commit being compacted.
	if len(compactedFileSet) == 0 {
		for size > s.levelZeroSize*int64(math.Pow(float64(s.levelSizeBase), float64(level))) {
			level++
		}
		spec.Output = path.Join(fileSet, Compacted, strconv.Itoa(level))
		return spec, nil
	}
	// Handle commits with a parent commit.
	compactedFileSet[0] = applyPrefix(compactedFileSet[0])
	if err := s.objC.Walk(ctx, path.Join(compactedFileSet[0], Compacted), func(name string) error {
		nextLevel, err := strconv.Atoi(path.Base(name))
		if err != nil {
			return err
		}
		// Handle levels that are non-empty.
		if nextLevel == level {
			idx, err := index.GetTopLevelIndex(ctx, s.objC, name)
			if err != nil {
				return err
			}
			size += idx.SizeBytes
			// If the output level has not been determined yet, then the current level will be an input
			// to the compaction.
			// If the output level has been determined, then the current level will be copied.
			// The copied levels are above the output level.
			if spec.Output == "" {
				spec.Input = append(spec.Input, name)
			} else {
				w, err := s.objC.Writer(ctx, path.Join(fileSet, Compacted, strconv.Itoa(level)))
				if err != nil {
					return err
				}
				r, err := s.objC.Reader(ctx, name, 0, 0)
				if err != nil {
					return err
				}
				if _, err := io.Copy(w, r); err != nil {
					return err
				}
			}
		}
		// If the output level has not been determined yet and the compaction size is less than the threshold for
		// the current level, then the current level becomes the output level.
		if spec.Output == "" && size <= s.levelZeroSize*int64(math.Pow(float64(s.levelSizeBase), float64(level))) {
			spec.Output = path.Join(fileSet, Compacted, strconv.Itoa(level))
		}
		level++
		return nil
	}); err != nil {
		return nil, err
	}
	return spec, nil
}

// Delete deletes a fileset.
func (s *Storage) Delete(ctx context.Context, fileSet string) error {
	fileSet = applyPrefix(fileSet)
	return s.objC.Walk(ctx, fileSet, func(name string) error {
		idx, err := index.GetTopLevelIndex(ctx, s.objC, name)
		if err != nil {
			return err
		}
		chunk := idx.DataOp.DataRefs[0].ChunkInfo.Chunk
		if err := s.chunks.DeleteSemanticReference(ctx, name, chunk, ""); err != nil {
			return err
		}
		return s.objC.Delete(ctx, name)
	})
}

func applyPrefix(fileSet string) string {
	if strings.HasPrefix(fileSet, prefix) {
		return fileSet
	}
	return path.Join(prefix, fileSet)
}

func applyPrefixes(fileSets []string) []string {
	var prefixedFileSets []string
	for _, fileSet := range fileSets {
		prefixedFileSets = append(prefixedFileSets, applyPrefix(fileSet))
	}
	return prefixedFileSets
}

// SubFileSetStr returns the string representation of a subfileset.
func SubFileSetStr(subFileSet int64) string {
	return fmt.Sprintf("%020d", subFileSet)
}
