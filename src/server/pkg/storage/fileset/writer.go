package fileset

import (
	"archive/tar"
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

// Writer writes the serialized format of a fileset.
// The serialized format of a fileset consists of indexes and content which are both realized as compressed tar stream chunks.
type Writer struct {
	tw    *tar.Writer
	cw    *chunk.Writer
	iw    *index.Writer
	first bool
}

// NewWriter creates a new Writer.
func NewWriter(ctx context.Context, chunks *chunk.Storage) *Writer {
	cw := chunks.NewWriter(ctx)
	return &Writer{
		tw:    tar.NewWriter(cw),
		cw:    cw,
		iw:    index.NewWriter(ctx, chunks),
		first: true,
	}
}

// WriteHeader writes a tar header and prepares to accept the file's contents.
func (w *Writer) WriteHeader(hdr *index.Header) error {
	// Flush out preceding file's content before starting new range.
	if !w.first {
		if err := w.tw.Flush(); err != nil {
			return err
		}
	}
	w.first = false
	w.cw.StartRange(w.callback(hdr))
	return w.tw.WriteHeader(hdr.Hdr)
}

func (w *Writer) callback(hdr *index.Header) func([]*chunk.DataRef) error {
	return func(dataRefs []*chunk.DataRef) error {
		if hdr.Idx == nil {
			hdr.Idx = &index.Index{}
		}
		hdr.Idx.DataOp = &index.DataOp{DataRefs: dataRefs}
		return w.iw.WriteHeader(hdr)
	}
}

// Write writes to the current file in the tar stream.
func (w *Writer) Write(data []byte) (int, error) {
	return w.tw.Write(data)
}

// Close closes the writer.
// (bryce) is not closing the tar writer the right choice here?
// We cannot include it in the last file's range because it will
// effect the hash.
// Our indexing is what will exit the reading of the tar stream,
// not the end of tar entry (two empty 512 bytes).
func (w *Writer) Close() (io.Reader, error) {
	// Flush the last file's content.
	if err := w.tw.Flush(); err != nil {
		return nil, err
	}
	// Close chunk and index writer.
	if err := w.cw.Close(); err != nil {
		return nil, err
	}
	return w.iw.Close()
}
