package chunk

import (
	"bytes"
	"context"
	"path"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/gc"
)

const (
	prefix = "chunks"
)

// Annotation is used to associate information with data
// written into the chunk storage layer.
type Annotation struct {
	RefDataRefs []*DataRef
	NextDataRef *DataRef
	Data        interface{}
	buf         *bytes.Buffer
	tags        []*Tag
	drs         []*DataReader
}

// Storage is the abstraction that manages chunk storage.
type Storage struct {
	objClient obj.Client
	gcClient  gc.Client
}

// NewStorage creates a new Storage.
func NewStorage(objClient obj.Client, opts ...StorageOption) *Storage {
	s := &Storage{
		objClient: objClient,
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
// The callback arguments are the chunk hash and annotations.
func (s *Storage) NewWriter(ctx context.Context, averageBits int, seed int64, noUpload bool, tmpID string, f WriterFunc) *Writer {
	return newWriter(ctx, s.objClient, s.gcClient, averageBits, f, seed, noUpload, tmpID)
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

func (s *Storage) AddSemanticReference(ctx context.Context, name string, chunk *Chunk, tmpID string) error {
	if err := s.gcClient.AddReference(ctx, semanticReference(name, chunk.Hash)); err != nil {
		return err
	}
	// (bryce) removing the temporary reference will eventually be handled by gc.
	return s.gcClient.RemoveReference(ctx, &gc.Reference{
		Sourcetype: "temporary",
		Source:     tmpID,
	})
}

func (s *Storage) RemoveSemanticReference(ctx context.Context, name string) error {
	return s.gcClient.RemoveReference(ctx, semanticReference(name, ""))
}

func semanticReference(name, chunk string) *gc.Reference {
	return &gc.Reference{
		Sourcetype: "semantic",
		Source:     name,
		Chunk:      path.Join(prefix, chunk),
	}
}
