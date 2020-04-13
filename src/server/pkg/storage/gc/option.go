package gc

import (
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
)

type Option func(gc *garbageCollector)

func WithPolling(polling time.Duration) Option {
	return func(gc *garbageCollector) {
		gc.polling = polling
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(gc *garbageCollector) {
		gc.timeout = timeout
	}
}

func ServiceEnvToOptions(env *serviceenv.ServiceEnv) ([]Option, error) {
	var opts []Option
	if env.StorageGCPolling != "" {
		polling, err := time.ParseDuration(env.StorageGCPolling)
		if err != nil {
			return nil, err
		}
		opts = append(opts, WithPolling(polling))
	}
	if env.StorageGCTimeout != "" {
		timeout, err := time.ParseDuration(env.StorageGCTimeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, WithTimeout(timeout))
	}
	return opts, nil
}
