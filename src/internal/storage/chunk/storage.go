package chunk

import (
	"bytes"
	"context"
	"encoding/hex"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

const (
	// TrackerPrefix is the prefix used when creating tracker objects for chunks
	TrackerPrefix        = "chunk/"
	defaultChunkTTL      = 30 * time.Minute
	DefaultPrefetchLimit = 10
)

// Storage is an abstraction for interfacing with chunk storage.
// A storage instance:
// - Provides methods for uploading data to chunks and downloading data from chunks.
// - Manages tracker state to keep chunks alive while uploading.
// - Manages an internal chunk cache and work deduplicator (parallel downloads of the same chunk will be deduplicated).
type Storage struct {
	db            *pachsql.DB
	tracker       track.Tracker
	store         kv.Store
	memCache      *memoryCache
	deduper       *miscutil.WorkDeduper[pachhash.Output]
	pool          *kv.Pool
	prefetchLimit int

	createOpts CreateOptions
}

// NewStorage creates a new Storage.
func NewStorage(store kv.Store, db *pachsql.DB, tracker track.Tracker, opts ...StorageOption) *Storage {
	s := &Storage{
		db:            db,
		store:         store,
		tracker:       tracker,
		memCache:      newMemoryCache(50),
		deduper:       &miscutil.WorkDeduper[pachhash.Output]{},
		pool:          kv.NewPool(DefaultMaxChunkSize),
		prefetchLimit: DefaultPrefetchLimit,
		createOpts: CreateOptions{
			Compression: CompressionAlgo_GZIP_BEST_SPEED,
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// NewReader creates a new Reader.
func (s *Storage) NewReader(ctx context.Context, dataRefs []*DataRef, opts ...ReaderOption) *Reader {
	client := NewClient(s.store, s.db, s.tracker, nil, s.pool)
	defaultOpts := []ReaderOption{WithPrefetchLimit(s.prefetchLimit)}
	return newReader(ctx, s, client, dataRefs, append(defaultOpts, opts...)...)
}

func (s *Storage) NewDataReader(ctx context.Context, dataRef *DataRef) *DataReader {
	client := NewClient(s.store, s.db, s.tracker, nil, s.pool)
	return newDataReader(ctx, s, client, dataRef, 0)
}

func (s *Storage) PrefetchData(ctx context.Context, dataRef *DataRef) error {
	return s.NewDataReader(ctx, dataRef).fetchData()
}

// ListStore lists all of the chunks in object storage.
// This is not the same as listing the chunk entries in the database.
func (s *Storage) ListStore(ctx context.Context, cb func(id ID, gen uint64) error) error {
	it := s.store.NewKeyIterator(kv.Span{})
	return stream.ForEach(ctx, it, func(key []byte) error {
		chunkID, gen, err := parseKey(key)
		if err != nil {
			return err
		}
		return cb(chunkID, gen)
	})
}

// NewDeleter creates a deleter for use with a tracker.GC
func (s *Storage) NewDeleter() track.Deleter {
	return &deleter{}
}

// Check runs an integrity check on the objects in object storage.
// It will check objects for chunks with IDs in the range [first, last)
// As a special case: if len(end) == 0 then it is ignored.
func (s *Storage) Check(ctx context.Context, begin, end []byte, readChunks bool) (int, error) {
	c := NewClient(s.store, s.db, s.tracker, nil, s.pool).(*trackedClient)
	first := append([]byte{}, begin...)
	var count int
	for {
		n, last, err := c.CheckEntries(ctx, first, 100, readChunks)
		count += n
		if err != nil {
			return count, err
		}
		if last == nil {
			break
		}
		if len(end) > 0 && bytes.Compare(last, end) > 0 {
			break
		}
		first = kv.KeyAfter(last)
	}
	return count, nil
}

type memoryCache = lru.Cache[pachhash.Output, []byte]

func newMemoryCache(size int) *memoryCache {
	c, err := lru.New[pachhash.Output, []byte](size)
	if err != nil {
		panic(err)
	}
	return c
}

func getFromCache(cache *memoryCache, ref *Ref) ([]byte, error) {
	key := ref.Key()
	chunkData, ok := cache.Get(key)
	if !ok {
		return nil, pacherr.NewNotExist("chunk-memory", hex.EncodeToString(key[:]))
	}
	return chunkData, nil
}

// putInCache takes data and inserts it into the cache.
// putInCache consumes data, and data must not be modified after passing it to putInCache.
func putInCache(cache *memoryCache, ref *Ref, data []byte) {
	key := ref.Key()
	cache.Add(key, data)
}
