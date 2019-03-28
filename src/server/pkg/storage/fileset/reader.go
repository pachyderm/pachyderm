package fileset

import (
	"archive/tar"
	"io"
)

// Reader is a wrapper for a tar reader that interprets
// Pachyderm specific flags (particularly MergeType).
type Reader struct {
	r tar.Reader
}

func NewReader(rs []io.Reader) *Reader {
	if len(rs) == 1 {
		// Just a reader
	} else {
		// Merge into reader
	}
}

func (r *Reader) Next() (*Header, error) {
	// Next will interpret and apply Pachyderm
	// specific flags, then return a header that
	// should be useable by external programs.
}

func (r *Reader) Read(b []byte) (int, error) {
	// Reads some content of a file.
}
