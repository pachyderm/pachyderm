package storage

import (
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// maxKeySize is the maximum size of keys in key-value stores.
const (
	maxKeySize = 512
)

// MakeChunkOptions returns the chunk storage options for the config.
func makeChunkOptions(conf *pachconfig.StorageConfiguration) (opts []chunk.StorageOption) {
	if conf.StorageMemoryCacheSize > 0 {
		opts = append(opts, chunk.WithMemoryCacheSize(conf.StorageMemoryCacheSize))
	}
	return opts
}

func makeFilesetOptions(conf *pachconfig.StorageConfiguration) (opts []fileset.StorageOption) {
	if conf.StorageMemoryThreshold > 0 {
		opts = append(opts, fileset.WithMemoryThreshold(conf.StorageMemoryThreshold))
	}
	if conf.StorageCompactionShardSizeThreshold > 0 {
		opts = append(opts, fileset.WithShardSizeThreshold(conf.StorageCompactionShardSizeThreshold))
	}
	if conf.StorageCompactionShardCountThreshold > 0 {
		opts = append(opts, fileset.WithShardCountThreshold(conf.StorageCompactionShardCountThreshold))
	}
	if conf.StorageLevelFactor > 0 {
		opts = append(opts, fileset.WithLevelFactor(conf.StorageLevelFactor))
	}
	return opts
}

// wrapStore adds layers around the store based on the storage configuration.
// This includes:
// - DiskCache
// - Upload/Download concurrency limits
// this is done below the chunk layer.
func wrapStore(conf *pachconfig.StorageConfiguration, store kv.Store) kv.Store {
	if conf.StorageUploadConcurrencyLimit > 0 || conf.StorageDownloadConcurrencyLimit > 0 {
		store = kv.NewSemaphored(store, conf.StorageDownloadConcurrencyLimit, conf.StorageUploadConcurrencyLimit)
	}
	if conf.StorageDiskCacheSize > 0 {
		p := filepath.Join(os.TempDir(), "pss-cache", uuid.NewWithoutDashes())
		var diskCache kv.Store = kv.NewFSStore(p, maxKeySize, chunk.DefaultMaxChunkSize)
		diskCache = kv.NewMetered(diskCache, "disk")
		return kv.NewLRUCache(store, diskCache, conf.StorageDiskCacheSize)
	}
	return store
}
