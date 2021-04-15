package server

import "github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"

type SourceOption func(*sourceConfig)

type sourceConfig struct {
	full   bool
	filter func(fileset.FileSet) fileset.FileSet
}

func WithFull() SourceOption {
	return func(sc *sourceConfig) {
		sc.full = true
	}
}

func WithFilter(filter func(fileset.FileSet) fileset.FileSet) SourceOption {
	return func(sc *sourceConfig) {
		sc.filter = filter
	}
}
