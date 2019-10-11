package fileset

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

const (
	headerTag = ""
	prefix    = "pfs"
	// FullMergeSuffix is appended to a prefix to create a path to be used for referencing
	// the full merge of that prefix. The full merge of a prefix means that every
	// path with the prefix has been merged into the referenced object.
	FullMergeSuffix = "~"
)

// ShardFunc is a callback that returns a PathRange for each shard.
type ShardFunc func(*index.PathRange) error

// Storage is the abstraction that manages fileset storage.
type Storage struct {
	objC   obj.Client
	chunks *chunk.Storage
	opts   []Option
}

// NewStorage creates a new Storage.
func NewStorage(objC obj.Client, chunks *chunk.Storage, opts ...Option) *Storage {
	return &Storage{
		objC:   objC,
		chunks: chunks,
		opts:   opts,
	}
}

// New creates a new in-memory fileset.
func (s *Storage) New(ctx context.Context, fileSet string, opts ...Option) *FileSet {
	fileSet = applyPrefix(fileSet)
	return newFileSet(ctx, s, fileSet, append(s.opts, opts...)...)
}

// NewWriter creates a new Writer.
func (s *Storage) NewWriter(ctx context.Context, fileSet string) *Writer {
	fileSet = applyPrefix(fileSet)
	return newWriter(ctx, s.objC, s.chunks, fileSet)
}

// NewReader creates a new Reader for a full file set.
func (s *Storage) NewReader(ctx context.Context, fileSet string, opts ...index.Option) *Reader {
	fileSet = applyPrefix(fileSet)
	return newReader(ctx, s.objC, s.chunks, fileSet, opts...)
}

// Shard shards the merge of the file sets with the passed in prefix into file ranges.
// (bryce) this should be extended to be more configurable (different criteria
// for creating shards).
func (s *Storage) Shard(ctx context.Context, fileSets []string, shardFunc ShardFunc) error {
	fileSets = applyPrefixes(fileSets)
	return s.merge(ctx, fileSets, shardMergeFunc(shardFunc))
}

// Merge merges the file sets with the passed in prefix.
func (s *Storage) Merge(ctx context.Context, outputFileSet string, inputFileSets []string, opts ...index.Option) error {
	outputFileSet = applyPrefix(outputFileSet)
	inputFileSets = applyPrefixes(inputFileSets)
	w := s.NewWriter(ctx, outputFileSet)
	if err := s.merge(ctx, inputFileSets, contentMergeFunc(w), opts...); err != nil {
		return err
	}
	return w.Close()
}

func (s *Storage) merge(ctx context.Context, fileSets []string, f mergeFunc, opts ...index.Option) error {
	var rs []*Reader
	for _, fileSet := range fileSets {
		if err := s.objC.Walk(ctx, fileSet, func(name string) error {
			rs = append(rs, s.NewReader(ctx, name, opts...))
			return nil
		}); err != nil {
			return err
		}
	}
	var fileStreams []stream
	for _, r := range rs {
		fileStreams = append(fileStreams, &fileStream{r: r})
	}
	return merge(fileStreams, f)
}

func applyPrefix(fileSet string) string {
	return path.Join(prefix, fileSet)
}

func applyPrefixes(fileSets []string) []string {
	for i := 0; i < len(fileSets); i++ {
		fileSets[i] = path.Join(prefix, fileSets[i])
	}
	return fileSets
}
