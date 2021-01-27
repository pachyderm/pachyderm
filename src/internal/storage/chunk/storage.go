package chunk

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/internal/obj"
	"github.com/pachyderm/pachyderm/src/internal/storage/track"
)

const (
	// TrackerPrefix is the prefix used when creating tracker objects for chunks
	TrackerPrefix   = "chunk/"
	prefix          = "chunk"
	defaultChunkTTL = 30 * time.Minute
)

// Storage is the abstraction that manages chunk storage.
type Storage struct {
	objClient obj.Client
	tracker   track.Tracker
	mdstore   MetadataStore

	defaultChunkTTL time.Duration
}

// NewStorage creates a new Storage.
func NewStorage(objClient obj.Client, mdstore MetadataStore, tracker track.Tracker, opts ...StorageOption) *Storage {
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
func (s *Storage) NewReader(ctx context.Context, dataRefs []*DataRef) *Reader {
	// using the empty string for the tmp id to disable the renewer
	client := NewClient(s.objClient, s.mdstore, s.tracker, "")
	return newReader(ctx, client, dataRefs)
}

// NewWriter creates a new Writer for a stream of bytes to be chunked.
// Chunks are created based on the content, then hashed and deduplicated/uploaded to
// object storage.
func (s *Storage) NewWriter(ctx context.Context, tmpID string, cb WriterCallback, opts ...WriterOption) *Writer {
	client := NewClient(s.objClient, s.mdstore, s.tracker, tmpID)
	return newWriter(ctx, client, cb, opts...)
}

// List lists all of the chunks in object storage.
func (s *Storage) List(ctx context.Context, cb func(string) error) error {
	return s.objClient.Walk(ctx, prefix, cb)
}

// NewDeleter creates a deleter for use with a tracker.GC
func (s *Storage) NewDeleter() track.Deleter {
	return &deleter{
		mdstore: s.mdstore,
		objc:    s.objClient,
	}
}
