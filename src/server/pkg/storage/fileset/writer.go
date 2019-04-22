package fileset

import (
	"archive/tar"
	"context"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

var (
	// pad is used to pad chunk.WindowSize bytes between the prior file content and the next.
	// The reason this is necessary is because we do not want the chunker to produce new
	// chunks in the content due to changes in the metadata (this can happen when the window
	// is sliding across the metadata/data boundary).
	pad = make([]byte, chunk.WindowSize)
)

// Writer writes the serialized format of a fileset.
// The serialized format of a fileset consists of indexes and content which are both realized as compressed tar stream chunks.
// The headers of the files in the indexes contain additional metadata about the indexed files (offset, hash, etc.).
type Writer struct {
	cw  chunk.Writer
	tw  tar.Writer
	idx *IndexWriter
}

// NewWriter creates a new Writer.
func NewWriter(ctx context.Context, chunks *chunk.Storage) *Writer {
	cw := chunks.NewWriter(ctx)
	return &Writer{
		cw:  cw,
		tw:  tar.NewWriter(cw),
		idx: index.NewWriter(ctx, chunks),
	}
}

// WriteHeader writes a tar header and prepares to accept the file's contents.
func (w *Writer) WriteHeader(hdr *Header) error {
	w.cw.RangeStart(callback(hdr))
	if err := w.tw.Write(pad); err != nil {
		return err
	}
	if err := w.tw.WriteHeader(hdr.hdr); err != nil {
		return err
	}
}

func (w *Writer) callback(hdr *Header) func([]*chunk.DataRef) error {
	return func(dataRefs []*chunk.DataRef) error {
		if hdr.info == nil {
			hdr.info = &FileInfo{}
		}
		for _, dataRef := range dataRefs {
			hdr.info.DataOps = append(hdr.info.DataOps, &DataOps{DataRef: dataRef})
		}
		return w.idx.WriteHeader(hdr)
	}
}

// Write writes to the current file in the tar archive.
func (w *Writer) Write(data []byte) (int, error) {
	return w.tw.Write(data)
}

func (w *Writer) Close() (*Header, error) {
	if err := w.cw.Close(); err != nil {
		return nil, err
	}
	if err := w.tw.Close(); err != nil {
		return nil, err
	}
	return w.idx.Close()
}
