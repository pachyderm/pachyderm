package chunk

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/tracker"
)

const (
	prefix          = "chunks"
	defaultChunkTTL = 30 * time.Minute
)

// Storage is the abstraction that manages chunk storage.
type Storage struct {
	objClient obj.Client
	tracker   tracker.Tracker
	mdstore   MetadataStore

	defaultChunkTTL time.Duration
}

// NewStorage creates a new Storage.
func NewStorage(objClient obj.Client, mdstore MetadataStore, tracker tracker.Tracker, opts ...StorageOption) *Storage {
	s := &Storage{
		objClient:       objClient,
		mdstore:         mdstore,
		defaultChunkTTL: defaultChunkTTL,
		tracker:         tracker,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// NewReader creates a new Reader.
func (s *Storage) NewReader(ctx context.Context, dataRefs ...*DataRef) *Reader {
	// using the empty chunkset for the reader
	client := NewClient(s.objClient, s.mdstore, s.tracker, "")
	return newReader(ctx, client, dataRefs...)
}

// NewWriter creates a new Writer for a stream of bytes to be chunked.
// Chunks are created based on the content, then hashed and deduplicated/uploaded to
// object storage.
func (s *Storage) NewWriter(ctx context.Context, tmpID string, f WriterFunc, opts ...WriterOption) *Writer {
	opts = append([]WriterOption{WithChunkTTL(defaultChunkTTL)}, opts...)
	client := NewClient(s.objClient, s.mdstore, s.tracker, tmpID)
	return newWriter(ctx, client, f, opts...)
}

// List lists all of the chunks in object storage.
func (s *Storage) List(ctx context.Context, f func(string) error) error {
	return s.objClient.Walk(ctx, prefix, f)
}

// // DeleteAll deletes all of the chunks in object storage.
// func (s *Storage) DeleteAll(ctx context.Context) error {
// 	panic("don't call this")
// 	return s.objClient.Walk(ctx, prefix, func(hash string) error {
// 		return s.objClient.Delete(ctx, hash)
// 	})
// }

// // Delete deletes a chunk in object storage.
// func (s *Storage) Delete(ctx context.Context, hash string) error {
// 	panic("don't call this")
// 	return s.objClient.Delete(ctx, path.Join(prefix, hash))
// }

// NewDeleter creates a deleter for use with a tracker.GC
func (s *Storage) NewDeleter() tracker.Deleter {
	return &deleter{
		mdstore: s.mdstore,
		objc:    s.objClient,
	}
}
