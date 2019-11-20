package chunk

import (
	"bytes"
	"context"
	"path"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

const (
	prefix = "chunks"
)

// Annotation is used to associate information with data
// written into the chunk storage layer.
type Annotation struct {
	RefDataRefs []*DataRef
	NextDataRef *DataRef
	Meta        interface{}
	buf         *bytes.Buffer
	tags        []*Tag
	drs         []*DataReader
}

// Storage is the abstraction that manages chunk storage.
type Storage struct {
	objC obj.Client
}

// NewStorage creates a new Storage.
func NewStorage(objC obj.Client, opts ...StorageOption) *Storage {
	s := &Storage{
		objC: objC,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// NewReader creates a new Reader.
func (s *Storage) NewReader(ctx context.Context, dataRefs ...*DataRef) *Reader {
	return newReader(ctx, s.objC, dataRefs...)
}

// NewWriter creates a new Writer for a stream of bytes to be chunked.
// Chunks are created based on the content, then hashed and deduplicated/uploaded to
// object storage.
// The callback arguments are the chunk hash and annotations.
func (s *Storage) NewWriter(ctx context.Context, averageBits int, f WriterFunc, seed int64) *Writer {
	return newWriter(ctx, s.objC, averageBits, f, seed)
}

// List lists all of the chunks in object storage.
func (s *Storage) List(ctx context.Context, f func(string) error) error {
	return s.objC.Walk(ctx, prefix, f)
}

// DeleteAll deletes all of the chunks in object storage.
func (s *Storage) DeleteAll(ctx context.Context) error {
	return s.objC.Walk(ctx, prefix, func(hash string) error {
		return s.objC.Delete(ctx, hash)
	})
}

// Delete deletes a chunk in object storage.
func (s *Storage) Delete(ctx context.Context, hash string) error {
	return s.objC.Delete(ctx, path.Join(prefix, hash))
}
