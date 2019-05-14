package fileset

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

const (
	prefix = "pfs"
)

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

// NewWriter creates a new Writer.
func (s *Storage) NewWriter(ctx context.Context, fileSet string) *Writer {
	fileSet = path.Join(prefix, fileSet)
	return newWriter(ctx, s.objC, s.chunks, fileSet)
}

// NewReader creates a new Reader.
func (s *Storage) NewReader(ctx context.Context, fileSet, idxPrefix string) *Reader {
	fileSet = path.Join(prefix, fileSet)
	return newReader(ctx, s.objC, s.chunks, fileSet, idxPrefix)
}

// NewIndexWriter creates a new index.Writer.
func (s *Storage) NewIndexWriter(ctx context.Context, fileSet string) *index.Writer {
	fileSet = path.Join(prefix, fileSet)
	return index.NewWriter(ctx, s.objC, s.chunks, fileSet)
}

// NewIndexReader creates a new index.Reader.
func (s *Storage) NewIndexReader(ctx context.Context, fileSet, idxPrefix string) *index.Reader {
	fileSet = path.Join(prefix, fileSet)
	return index.NewReader(ctx, s.objC, s.chunks, fileSet, idxPrefix)
}
