package chunk

import (
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
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

// StorageOptions returns the chunk storage options for the config.
func StorageOptions(conf *pachconfig.StorageConfiguration) ([]StorageOption, error) {
	var opts []StorageOption
	if conf.StorageUploadConcurrencyLimit > 0 {
		opts = append(opts, WithMaxConcurrentObjects(0, conf.StorageUploadConcurrencyLimit))
	}
	if conf.StorageDiskCacheSize > 0 {
		diskCache, err := obj.NewLocalClient(filepath.Join(os.TempDir(), "pfs-cache", uuid.NewWithoutDashes()))
		if err != nil {
			return nil, err
		}
		diskCache = obj.TracingObjClient("DiskCache", diskCache)
		opts = append(opts, WithObjectCache(diskCache, conf.StorageDiskCacheSize))
	}
	return opts, nil
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
