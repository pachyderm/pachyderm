package storage

import (
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// FilesetOptions returns the fileset storage options for the config.
func FilesetOptions(conf *serviceenv.Configuration) []fileset.StorageOption {
	var opts []fileset.StorageOption
	if conf.StorageMemoryThreshold > 0 {
		opts = append(opts, fileset.WithMemoryThreshold(conf.StorageMemoryThreshold))
	}
	if conf.StorageShardThreshold > 0 {
		opts = append(opts, fileset.WithShardThreshold(conf.StorageShardThreshold))
	}
	if conf.StorageLevelFactor > 0 {
		opts = append(opts, fileset.WithLevelFactor(conf.StorageLevelFactor))
	}
	return opts
}

// ChunkOptions returns the chunk storage options for the config.
func ChunkOptions(conf *serviceenv.Configuration) ([]chunk.StorageOption, error) {
	var opts []chunk.StorageOption
	if conf.StorageUploadConcurrencyLimit > 0 {
		opts = append(opts, chunk.WithMaxConcurrentObjects(0, conf.StorageUploadConcurrencyLimit))
	}
	if conf.StorageDiskCacheSize > 0 {
		diskCache, err := obj.NewLocalClient(filepath.Join(os.TempDir(), "pfs-cache", uuid.NewWithoutDashes()))
		if err != nil {
			return nil, err
		}
		diskCache = obj.TracingObjClient("DiskCache", diskCache)
		opts = append(opts, chunk.WithObjectCache(diskCache, conf.StorageDiskCacheSize))
	}
	return opts, nil
}
