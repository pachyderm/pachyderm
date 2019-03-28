package files

import (
	"archive/tar"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// Writer is a wrapper for a tar writer that writes
// Pachyderm specific flags.
type Writer struct {
	r tar.Writer
	// Generate index and chunks?
	// Would need direct access to object storage and be able to page indexes
	// out to object storage and understand the rolling hash.
}

func NewWriter(client *obj.Client, prefix string) *Writer {
	if len(rs) == 1 {
		// Just a reader
	} else {
		// Merge into reader
		// This would essentially function as the merge reader.
	}
}

func (w *Writer) WriteHeader() error {

}

func (w *Writer) Write(b []byte) (int, error) {

}
