package chunk

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
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
	store     kv.Store
	memCache  kv.GetPut
	tracker   track.Tracker
	db        *sqlx.DB

	createOpts CreateOptions
}

// NewStorage creates a new Storage.
func NewStorage(objC obj.Client, memCache kv.GetPut, db *sqlx.DB, tracker track.Tracker, opts ...StorageOption) *Storage {
	s := &Storage{
		objClient: objC,
		memCache:  memCache,
		db:        db,
		tracker:   tracker,
		createOpts: CreateOptions{
			Compression: CompressionAlgo_GZIP_BEST_SPEED,
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	s.store = kv.NewFromObjectClient(s.objClient)
	s.objClient = nil
	return s
}

// NewReader creates a new Reader.
func (s *Storage) NewReader(ctx context.Context, dataRefs []*DataRef, opts ...ReaderOption) *Reader {
	// using the empty string for the tmp id to disable the renewer
	client := NewClient(s.store, s.db, s.tracker, "")
	return newReader(ctx, client, s.memCache, dataRefs, opts...)
}

// NewWriter creates a new Writer for a stream of bytes to be chunked.
// Chunks are created based on the content, then hashed and deduplicated/uploaded to
// object storage.
func (s *Storage) NewWriter(ctx context.Context, name string, cb WriterCallback, opts ...WriterOption) *Writer {
	if name == "" {
		panic("name must not be empty")
	}
	client := NewClient(s.store, s.db, s.tracker, name)
	return newWriter(ctx, client, s.memCache, s.createOpts, cb, opts...)
}

// List lists all of the chunks in object storage.
func (s *Storage) List(ctx context.Context, cb func(id ID) error) error {
	return s.store.Walk(ctx, nil, func(key []byte) error {
		return cb(ID(key))
	})
}

// NewDeleter creates a deleter for use with a tracker.GC
func (s *Storage) NewDeleter() track.Deleter {
	return &deleter{}
}
