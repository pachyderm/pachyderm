package server

import "github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"

// SourceOption configures a source.
type SourceOption func(*sourceConfig)

type sourceConfig struct {
	filter func(fileset.FileSet) fileset.FileSet
}

// WithFilter applies a filter to the fileset after it has been set up by the source.
func WithFilter(filter func(fileset.FileSet) fileset.FileSet) SourceOption {
	return func(sc *sourceConfig) {
		sc.filter = filter
	}
}
