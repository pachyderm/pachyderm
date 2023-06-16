package storage

import (
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// MakeChunkOptions returns the chunk storage options for the config.
func makeChunkOptions(conf *pachconfig.StorageConfiguration) ([]chunk.StorageOption, error) {
	var opts []chunk.StorageOption
	if conf.StorageUploadConcurrencyLimit > 0 {
		opts = append(opts, chunk.WithMaxConcurrentObjects(0, conf.StorageUploadConcurrencyLimit))
	}
	if conf.StorageDiskCacheSize > 0 {
		p := filepath.Join(os.TempDir(), "pfs-cache", uuid.NewWithoutDashes())
		diskCache, err := obj.NewLocalClient(p)
		if err != nil {
			return nil, err
		}
		opts = append(opts, chunk.WithObjectCache(diskCache, conf.StorageDiskCacheSize))
	}
	if conf.StorageMemoryCacheSize > 0 {
		opts = append(opts, chunk.WithMemoryCacheSize(conf.StorageMemoryCacheSize))
	}
	return opts, nil
}

func makeFilesetOptions(conf *pachconfig.StorageConfiguration) []fileset.StorageOption {
	var opts []fileset.StorageOption
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
