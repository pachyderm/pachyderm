package serviceenv

import (
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// ChunkStorageOptions returns the chunk storage options for the service environment.
func (env *ServiceEnv) ChunkStorageOptions() ([]chunk.StorageOption, error) {
	var opts []chunk.StorageOption
	if env.StorageUploadConcurrencyLimit > 0 {
		opts = append(opts, chunk.WithMaxConcurrentObjects(0, env.StorageUploadConcurrencyLimit))
	}
	if env.StorageDiskCacheSize > 0 {
		diskCache, err := obj.NewLocalClient(filepath.Join(os.TempDir(), "pfs-cache", uuid.NewWithoutDashes()))
		if err != nil {
			return nil, err
		}
		opts = append(opts, chunk.WithObjectCache(diskCache, env.StorageDiskCacheSize))
	}
	return opts, nil
}

// FileSetStorageOptions returns the fileset storage options for the service environment.
func (env *ServiceEnv) FileSetStorageOptions() []fileset.StorageOption {
	var opts []fileset.StorageOption
	if env.StorageMemoryThreshold > 0 {
		opts = append(opts, fileset.WithMemoryThreshold(env.StorageMemoryThreshold))
	}
	if env.StorageShardThreshold > 0 {
		opts = append(opts, fileset.WithShardThreshold(env.StorageShardThreshold))
	}
	if env.StorageLevelZeroSize > 0 {
		opts = append(opts, fileset.WithLevelZeroSize(env.StorageLevelZeroSize))
	}
	if env.StorageLevelSizeBase > 0 {
		opts = append(opts, fileset.WithLevelSizeBase(env.StorageLevelSizeBase))
	}
	return opts
}
