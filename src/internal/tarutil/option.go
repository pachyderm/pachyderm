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

// WithSymlink configures the export call to execute the callback for each symlink encountered.
// The callback that is passed into the callback should be executed if the symlinked file should be written to the tar stream.
func WithSymlinkCallback(cb func(string, string, func() error) error) ExportOption {
	return func(ec *exportConfig) {
		ec.symlinkCallback = cb
	}
}
