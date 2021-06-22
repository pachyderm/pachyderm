package server

import "github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"

// SourceOption configures a source.
type SourceOption func(*sourceConfig)

type sourceConfig struct {
	details bool
	filter  func(fileset.FileSet) fileset.FileSet
}

// WithDetails sets the source to populate the 'Details' field of the returned
// files.  This means resolving the indexes for computing the true hashes of the
// files, and computing the directory hashes / sizes by scanning ahead in a
// parallel iteration.
func WithDetails() SourceOption {
	return func(sc *sourceConfig) {
		sc.details = true
	}
}

// WithFilter applies a filter to the fileset after it has been set up by the source.
func WithFilter(filter func(fileset.FileSet) fileset.FileSet) SourceOption {
	return func(sc *sourceConfig) {
		sc.filter = filter
	}
}
