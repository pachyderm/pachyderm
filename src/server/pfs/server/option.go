package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// SourceOption configures a source.
type SourceOption func(*sourceConfig)

type sourceConfig struct {
	prefix    string
	pathRange *pfs.PathRange
	datum     string
	filter    func(fileset.FileSet) fileset.FileSet
}

func WithPrefix(prefix string) SourceOption {
	return func(sc *sourceConfig) {
		sc.prefix = prefix
		sc.pathRange = nil
	}
}

func WithPathRange(pathRange *pfs.PathRange) SourceOption {
	return func(sc *sourceConfig) {
		sc.pathRange = pathRange
		sc.prefix = ""
	}
}

func WithDatum(datum string) SourceOption {
	return func(sc *sourceConfig) {
		sc.datum = datum
	}
}

// WithFilter applies a filter to the file set after it has been set up by the source.
func WithFilter(filter func(fileset.FileSet) fileset.FileSet) SourceOption {
	return func(sc *sourceConfig) {
		sc.filter = filter
	}
}
