package serviceenv

import (
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
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
	if env.StorageLevelFactor > 0 {
		opts = append(opts, fileset.WithLevelFactor(env.StorageLevelFactor))
	}
	return opts
}

// ChunkMemoryCache returns the in memory cache for chunks, pre-configured to the desired size
func (env *ServiceEnv) ChunkMemoryCache() kv.GetPut {
	size := env.StorageMemoryCacheSize
	if size < 1 {
		size = 1
	}
	return kv.NewMemCache(size)
}

type ConfigOption = func(*Configuration)

func ApplyOptions(config *Configuration, opts ...ConfigOption) {
	for _, opt := range opts {
		opt(config)
	}
}

func NewConfig(opts ...ConfigOption) *Configuration {
	result := &Configuration{}
	ApplyOptions(result, opts...)
	return result
}

func WithPostgresHostPort(host string, port string) ConfigOption {
	return func(config *Configuration) {
		config.PostgresServiceHost = host
		config.PostgresServicePort = port
	}
}

func WithEtcdHostPort(host string, port string) ConfigOption {
	return func(config *Configuration) {
		config.EtcdHost = host
		config.EtcdPort = port
	}
}

func WithPachdPeerPort(port uint16) ConfigOption {
	return func(config *Configuration) {
		config.PeerPort = port
	}
}
