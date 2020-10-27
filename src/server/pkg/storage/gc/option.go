package gc

import (
	"time"
)

// Option configures the garbage collector.
type Option func(gc *garbageCollector)

// WithPolling sets the polling duration.
func WithPolling(polling time.Duration) Option {
	return func(gc *garbageCollector) {
		gc.polling = polling
	}
}
