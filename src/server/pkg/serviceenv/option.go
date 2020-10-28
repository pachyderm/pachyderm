package serviceenv

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
)

var localDiskCachePath = filepath.Join(os.TempDir(), "pfs-cache")

// ChunkStorageOptions returns the chunk storage options for the service environment.
func (env *ServiceEnv) ChunkStorageOptions() ([]chunk.StorageOption, error) {
	var opts []chunk.StorageOption
	if env.StorageUploadConcurrencyLimit > 0 {
		opts = append(opts, chunk.WithMaxConcurrentObjects(0, env.StorageUploadConcurrencyLimit))
	}
	if env.StorageGCTimeout != "" {
		timeout, err := time.ParseDuration(env.StorageGCTimeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, chunk.WithGCTimeout(timeout))
	}
	if env.StorageDiskCacheSize > 0 {
		diskCache, err := obj.NewLocalClient(localDiskCachePath)
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
