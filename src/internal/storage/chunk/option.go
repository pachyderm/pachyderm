package chunk

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
)

// StorageOption configures a storage.
type StorageOption func(s *Storage)

// WithMaxConcurrentObjects sets the maximum number of object writers (upload)
// and readers (download) that can be open at a time.
func WithMaxConcurrentObjects(maxDownload, maxUpload int) StorageOption {
	return func(s *Storage) {
		s.objClient = obj.NewLimitedClient(s.objClient, maxDownload, maxUpload)
	}
}

// WithObjectCache adds a cache around the currently configured object client
func WithObjectCache(fastLayer obj.Client, size int) StorageOption {
	return func(s *Storage) {
		s.objClient = obj.NewCacheClient(s.objClient, fastLayer, size)
	}
}

// WithMemoryCacheSize sets the number of decrypted, uncompressed chunks that will be stored in memory.
func WithMemoryCacheSize(size int) StorageOption {
	return func(s *Storage) {
		mc, err := lru.New[pachhash.Output, []byte](size)
		if err != nil {
			panic(err)
		}
		s.memCache = mc
	}
}

// WithSecret sets the secret used to generate chunk encryption keys
func WithSecret(secret []byte) StorageOption {
	return func(s *Storage) {
		s.createOpts.Secret = append([]byte{}, secret...)
	}
}

// WithCompression sets the compression algorithm used to compress chunks
func WithCompression(algo CompressionAlgo) StorageOption {
	return func(s *Storage) {
		s.createOpts.Compression = algo
	}
}

type BatcherOption func(b *Batcher)

func WithChunkCallback(cb ChunkFunc) BatcherOption {
	return func(b *Batcher) {
		b.chunkFunc = cb
		b.entryFunc = nil
	}
}

func WithEntryCallback(cb EntryFunc) BatcherOption {
	return func(b *Batcher) {
		b.chunkFunc = nil
		b.entryFunc = cb
	}
}

type ReaderOption func(*Reader)

func WithOffsetBytes(offsetBytes int64) ReaderOption {
	return func(r *Reader) {
		r.offsetBytes = offsetBytes
	}
}

func WithPrefetchLimit(limit int) ReaderOption {
	return func(r *Reader) {
		r.prefetchLimit = limit
	}
}
