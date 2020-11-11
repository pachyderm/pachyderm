package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
)

// File represents a file.
type File interface {
	// Index returns the index for the file.
	Index() *index.Index
	// Header returns the tar header for the file.
	Header() (*tar.Header, error)
	// Content writes the content of the file.
	Content(w io.Writer) error
}

var _ File = &MergeFileReader{}
var _ File = &FileReader{}

// FileSet represents a set of files.
type FileSet interface {
	// Iterate iterates over the files in the file set.
	Iterate(ctx context.Context, cb func(File) error, deletive ...bool) error
}

var _ FileSet = &MergeReader{}
var _ FileSet = &Reader{}
