package testing

import (
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
)

// TODO Will use later, commenting to make linter happy.
//func WithPachdConfig(opts ...PachdConfigOption) loadConfigOption {
//	return func(config *loadConfig) {
//		config.pachdConfig = newPachdConfig(opts...)
//	}
//}

// PachdConfigOption configures Pachd.
// TODO This should probably be moved to the corresponding packages with configuration available
type PachdConfigOption func(*serviceenv.PachdFullConfiguration)

// NewPachdConfig creates a new default Pachd configuration for V2.
func NewPachdConfig(opts ...PachdConfigOption) *serviceenv.PachdFullConfiguration {
	config := &serviceenv.PachdFullConfiguration{}
	config.StorageV2 = true
	config.StorageMemoryThreshold = units.GB
	config.StorageShardThreshold = units.GB
	config.StorageLevelZeroSize = units.MB
	config.StorageGCPolling = "30s"
	config.StorageCompactionMaxFanIn = 10
	for _, opt := range opts {
		opt(config)
	}
	return config
}
