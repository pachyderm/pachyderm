package pfssync

import "archive/tar"

// ImportOption configures an import call.
type ImportOption func(i *importConfig)

// WithLazy configures the import call to lazily import files.
func WithLazy() ImportOption {
	return func(i *importConfig) {
		i.lazy = true
	}
}

// WithEmpty configures the import call to just import the file info.
func WithEmpty() ImportOption {
	return func(i *importConfig) {
		i.empty = true
	}
}

// WithHeaderCallback configures the import call to execute the callback for each tar file imported.
func WithHeaderCallback(cb func(*tar.Header) error) ImportOption {
	return func(i *importConfig) {
		i.headerCallback = cb
	}
}
