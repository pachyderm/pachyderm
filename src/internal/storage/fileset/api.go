package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/internal/storage/fileset/index"
)

// File represents a file.
type File interface {
	// Index returns the index for the file.
	Index() *index.Index
	// Content writes the content of the file.
	Content(w io.Writer) error
}

var _ File = &MergeFileReader{}
var _ File = &FileReader{}

// FileSet represents a set of files.
type FileSet interface {
	// Iterate iterates over the files in the file set.
	Iterate(ctx context.Context, cb func(File) error, deletive ...bool) error
	// TODO: Implement IterateDeletes or pull deletion information out of the fileset API.
}

var _ FileSet = &MergeReader{}
var _ FileSet = &Reader{}
