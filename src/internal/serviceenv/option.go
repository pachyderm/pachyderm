package serviceenv

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
)

// ChunkMemoryCache returns the in memory cache for chunks, pre-configured to the desired size
func (conf *StorageConfiguration) ChunkMemoryCache() kv.GetPut {
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
		GlobalConfiguration:             &GlobalConfiguration{},
		PachdSpecificConfiguration:      &PachdSpecificConfiguration{},
		WorkerSpecificConfiguration:     &WorkerSpecificConfiguration{},
		EnterpriseSpecificConfiguration: &EnterpriseSpecificConfiguration{},
	}
	ApplyOptions(result, opts...)
	return result
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

func WithOidcPort(port uint16) ConfigOption {
	return func(config *Configuration) {
		config.OidcPort = port
	}
}
