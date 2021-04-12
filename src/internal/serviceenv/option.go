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
func (conf *Configuration) ChunkStorageOptions() ([]chunk.StorageOption, error) {
	var opts []chunk.StorageOption
	if conf.StorageUploadConcurrencyLimit > 0 {
		opts = append(opts, chunk.WithMaxConcurrentObjects(0, conf.StorageUploadConcurrencyLimit))
	}
	if conf.StorageDiskCacheSize > 0 {
		diskCache, err := obj.NewLocalClient(filepath.Join(os.TempDir(), "pfs-cache", uuid.NewWithoutDashes()))
		if err != nil {
			return nil, err
		}
		opts = append(opts, chunk.WithObjectCache(diskCache, conf.StorageDiskCacheSize))
	}
	return opts, nil
}

// FileSetStorageOptions returns the fileset storage options for the service environment.
func (conf *Configuration) FileSetStorageOptions() []fileset.StorageOption {
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

// ChunkMemoryCache returns the in memory cache for chunks, pre-configured to the desired size
func (conf *Configuration) ChunkMemoryCache() kv.GetPut {
	size := conf.StorageMemoryCacheSize
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

// ConfigFromOptions is for use in tests where some tests may want to generate a
// config with certain default values overridden via options.
func ConfigFromOptions(opts ...ConfigOption) *Configuration {
	result := &Configuration{
		GlobalConfiguration:         &GlobalConfiguration{},
		PachdSpecificConfiguration:  &PachdSpecificConfiguration{},
		WorkerSpecificConfiguration: &WorkerSpecificConfiguration{},
	}
	ApplyOptions(result, opts...)
	return result
}

func WithPostgresHostPort(host string, port int) ConfigOption {
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
