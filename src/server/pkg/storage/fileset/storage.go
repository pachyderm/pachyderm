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
}

// NewStorage creates a new Storage.
func NewStorage(objC obj.Client, chunks *chunk.Storage) *Storage {
	return &Storage{
		objC:   objC,
		chunks: chunks,
	}
}

// New creates a new in-memory fileset.
func (s *Storage) New(ctx context.Context, name string, opts ...Option) *FileSet {
	return newFileSet(ctx, s, name, opts...)
}

// NewWriter creates a new Writer.
func (s *Storage) NewWriter(ctx context.Context, fileSet string) *Writer {
	return newWriter(ctx, s.objC, s.chunks, fileSet)
}

// NewReader creates a new Reader for a full file set.
func (s *Storage) NewReader(ctx context.Context, fileSet string, opts ...index.Option) *Reader {
	// (bryce) lazy merge logic here?
	fileSet = path.Join(fileSet, FullMergeSuffix)
	return s.newReader(ctx, fileSet, opts...)
}

func (s *Storage) newReader(ctx context.Context, fileSet string, opts ...index.Option) *Reader {
	return newReader(ctx, s.objC, s.chunks, fileSet, opts...)
}

// NewIndexWriter creates a new index.Writer.
func (s *Storage) NewIndexWriter(ctx context.Context, fileSet string) *index.Writer {
	return index.NewWriter(ctx, s.objC, s.chunks, fileSet)
}

// NewIndexReader creates a new index.Reader.
func (s *Storage) NewIndexReader(ctx context.Context, fileSet string, opts ...index.Option) *index.Reader {
	return index.NewReader(ctx, s.objC, s.chunks, fileSet, opts...)
}

// Shard shards the merge of the file sets with the passed in prefix into file ranges.
// (bryce) this should be extended to be more configurable (different criteria
// for creating shards).
func (s *Storage) Shard(ctx context.Context, fileSet string, shardFunc ShardFunc) error {
	return s.merge(ctx, fileSet, shardMergeFunc(shardFunc))
}

// Merge merges the file sets with the passed in prefix.
func (s *Storage) Merge(ctx context.Context, outputFileSet, inputFileSet string, opts ...index.Option) error {
	w := s.NewWriter(ctx, outputFileSet)
	if err := s.merge(ctx, inputFileSet, contentMergeFunc(w), opts...); err != nil {
		return err
	}
	return w.Close()
}

func (s *Storage) merge(ctx context.Context, fileSet string, f mergeFunc, opts ...index.Option) error {
	var rs []*Reader
	if err := s.objC.Walk(ctx, fileSet, func(name string) error {
		rs = append(rs, s.newReader(ctx, name, opts...))
		return nil
	}); err != nil {
		return err
	}
	var fileStreams []stream
	for _, r := range rs {
		fileStreams = append(fileStreams, &fileStream{r: r})
	}
	return merge(fileStreams, f)
}
