package chunk

import (
	"context"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/gc"
)

const (
	prefix          = "chunks"
	defaultChunkTTL = 30 * time.Minute
)

// Storage is the abstraction that manages chunk storage.
type Storage struct {
	objClient obj.Client
	gcClient  gc.Client

	defaultChunkTTL time.Duration
}

// NewStorage creates a new Storage.
func NewStorage(objClient obj.Client, opts ...StorageOption) *Storage {
	s := &Storage{
		objClient:       objClient,
		defaultChunkTTL: defaultChunkTTL,
	}
	s.gcClient = gc.NewMockClient()
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// NewReader creates a new Reader.
func (s *Storage) NewReader(ctx context.Context, dataRefs ...*DataRef) *Reader {
	return newReader(ctx, s.objClient, dataRefs...)
}

// NewWriter creates a new Writer for a stream of bytes to be chunked.
// Chunks are created based on the content, then hashed and deduplicated/uploaded to
// object storage.
func (s *Storage) NewWriter(ctx context.Context, tmpID string, f WriterFunc, opts ...WriterOption) *Writer {
	return newWriter(ctx, s.objClient, s.gcClient, tmpID, f, append(opts, WithChunkTTL(defaultChunkTTL))...)
}

// List lists all of the chunks in object storage.
func (s *Storage) List(ctx context.Context, f func(string) error) error {
	return s.objClient.Walk(ctx, prefix, f)
}

// DeleteAll deletes all of the chunks in object storage.
func (s *Storage) DeleteAll(ctx context.Context) error {
	return s.objClient.Walk(ctx, prefix, func(hash string) error {
		return s.objClient.Delete(ctx, hash)
	})
}

// Delete deletes a chunk in object storage.
func (s *Storage) Delete(ctx context.Context, hash string) error {
	return s.objClient.Delete(ctx, path.Join(prefix, hash))
}

// CreateSemanticReference creates a semantic reference to a chunk.
func (s *Storage) CreateSemanticReference(ctx context.Context, name string, chunk *Chunk) error {
	return s.gcClient.CreateReference(ctx, semanticReference(name, chunk.Hash))
}

// CreateSemanticReference creates a semantic reference to a chunk.
func (s *Storage) CreateTemporaryReference(ctx context.Context, name string, chunk *Chunk, ttl time.Duration) (*time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	if err := s.gcClient.CreateReference(ctx, temporaryReference(name, chunk.Hash, expiresAt)); err != nil {
		return nil, err
	}
	return &expiresAt, nil
}

// DeleteSemanticReference deletes a semantic reference.
func (s *Storage) DeleteSemanticReference(ctx context.Context, name string) error {
	return s.gcClient.DeleteReference(ctx, semanticReference(name, ""))
}

// RewnewReference sets the time to live for a reference to now + ttl, and returns the new expire time.
func (s *Storage) RenewReference(ctx context.Context, name string, ttl time.Duration) (*time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	if err := s.gcClient.RenewReference(ctx, name, expiresAt); err != nil {
		return nil, err
	}
	return &expiresAt, nil
}

func semanticReference(name, chunk string) *gc.Reference {
	return &gc.Reference{
		Sourcetype: gc.STSemantic,
		Source:     name,
		Chunk:      path.Join(prefix, chunk),
	}
}

func temporaryReference(name, chunk string, expiresAt time.Time) *gc.Reference {
	return &gc.Reference{
		Sourcetype: gc.STTemporary,
		Source:     name,
		Chunk:      path.Join(prefix, chunk),
		ExpiresAt:  &expiresAt,
	}
}
