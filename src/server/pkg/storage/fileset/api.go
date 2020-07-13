package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

// File represents a file in a fileset
type File interface {
	// Index returns the index for the file
	Index() *index.Index
	// Header returns a *tar.Header with metadata about the file, or an error
	Header() (*tar.Header, error)
	// Content writes the content of the file to w
	Content(w io.Writer) error
}

var _ File = &FileMergeReader{}
var _ File = &FileReader{}

type FileSource interface {
	// Iterate calls cb for each File in the FileSource in lexigraphical order.
	Iterate(ctx context.Context, cb func(File) error, stopBefore ...string) error
}

var _ FileSource = &mergeSource{}
