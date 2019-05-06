package chunk

import (
	"context"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// Storage is the abstraction that manages chunk storage.
type Storage struct {
	objC   obj.Client
	prefix string
}

// NewStorage creates a new Storage.
func NewStorage(objC obj.Client, prefix string) *Storage {
	return &Storage{
		objC:   objC,
		prefix: prefix,
	}
}

// NewReader creates an io.ReadCloser for a chunk.
// (bryce) The whole chunk is in-memory right now. Could be a problem with
// concurrency, particularly the merge process.
// May want to handle concurrency here (pass in multiple data refs)
func (s *Storage) NewReader(ctx context.Context, dataRefs ...*DataRef) *Reader {
	return newReader(ctx, s.objC, s.prefix, dataRefs...)
}

// NewWriter creates an io.WriteCloser for a stream of bytes to be chunked.
// Chunks are created based on the content, then hashed and deduplicated/uploaded to
// object storage.
// The callback arguments are the chunk hash and content.
func (s *Storage) NewWriter(ctx context.Context) *Writer {
	return newWriter(ctx, s.objC, s.prefix)
}

// DeleteAll deletes all of the chunks in object storage.
func (s *Storage) DeleteAll(ctx context.Context) error {
	return s.objC.Walk(ctx, s.prefix, func(hash string) error {
		return s.objC.Delete(ctx, hash)
	})
}
