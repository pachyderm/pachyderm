package gc

import (
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
)

// Option configures the garbage collector.
type Option func(gc *garbageCollector)

// WithPolling sets the polling duration.
func WithPolling(polling time.Duration) Option {
	return func(gc *garbageCollector) {
		gc.polling = polling
	}
}

// ServiceEnvToOptions converts a service environment configuration (specifically
// the garbage collection configuration) to a set of options.
func ServiceEnvToOptions(env *serviceenv.ServiceEnv) ([]Option, error) {
	var opts []Option
	if env.StorageGCPolling != "" {
		polling, err := time.ParseDuration(env.StorageGCPolling)
		if err != nil {
			return nil, err
		}
		opts = append(opts, WithPolling(polling))
	}
	return opts, nil
}
