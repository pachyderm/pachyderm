package tarutil

import "archive/tar"

// ExportOption configures an export call.
type ExportOption func(*exportConfig)

// WithHeaderCallback configures the export call to execute the callback for each tar file exported.
func WithHeaderCallback(cb func(*tar.Header) error) ExportOption {
	return func(ec *exportConfig) {
		ec.headerCallback = cb
	}
}
