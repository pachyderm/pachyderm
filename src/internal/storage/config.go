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

// MakeChunkOptions returns the chunk storage options for the config.
func makeChunkOptions(conf *pachconfig.StorageConfiguration) (opts []chunk.StorageOption) {
	if conf.StorageUploadConcurrencyLimit > 0 {
		opts = append(opts, chunk.WithMaxConcurrentObjects(0, conf.StorageUploadConcurrencyLimit))
	}
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

// addDiskCache adds a disk cache based on the storage configuration.
// this is done below the chunk layer.
func addDiskCache(conf *pachconfig.StorageConfiguration, store kv.Store) kv.Store {
	if conf.StorageDiskCacheSize > 0 {
		p := filepath.Join(os.TempDir(), "pss-cache", uuid.NewWithoutDashes())
		diskCache := kv.NewFSStore(p)
		return kv.NewLRUCache(store, diskCache, conf.StorageDiskCacheSize)
	} else {
		return store
	}
}
