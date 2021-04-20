package server

import "github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"

// SourceOption configures a source.
type SourceOption func(*sourceConfig)

type sourceConfig struct {
	full   bool
	filter func(fileset.FileSet) fileset.FileSet
}

// WithFull sets the source to generate the full metadata for the returned files.
// This means resolving the indexes for computing the true hashes of the files,
// and computing the directory hashes / sizes by scanning ahead in a parallel iteration.
func WithFull() SourceOption {
	return func(sc *sourceConfig) {
		sc.full = true
	}
}

// WithFilter applies a filter to the fileset after it has been set up by the source.
func WithFilter(filter func(fileset.FileSet) fileset.FileSet) SourceOption {
	return func(sc *sourceConfig) {
		sc.filter = filter
	}
}
