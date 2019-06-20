package fileset

import (
	"context"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

// Writer writes the serialized format of a fileset.
// The serialized format of a fileset consists of indexes and content which are both realized as compressed tar stream chunks.
type Writer struct {
	tw    *tar.Writer
	cw    *chunk.Writer
	iw    *index.Writer
	first bool
	hdr   *index.Header
}

func newWriter(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string) *Writer {
	cw := chunks.NewWriter(ctx)
	return &Writer{
		tw:    tar.NewWriter(cw),
		cw:    cw,
		iw:    index.NewWriter(ctx, objC, chunks, path),
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
	// Setup index header for next file.
	// (bryce) might want to deep copy the passed in header.
	w.hdr = &index.Header{
		Hdr: hdr.Hdr,
		Idx: &index.Index{DataOp: &index.DataOp{}},
	}
	w.cw.StartRange(w.callback(w.hdr))
	if err := w.tw.WriteHeader(w.hdr.Hdr); err != nil {
		return err
	}
	// Setup first tag for header.
	w.hdr.Idx.DataOp.Tags = []*index.Tag{&index.Tag{Id: headerTag, SizeBytes: w.cw.RangeSize()}}
	return nil
}

func (w *Writer) callback(hdr *index.Header) func([]*chunk.DataRef) error {
	return func(dataRefs []*chunk.DataRef) error {
		hdr.Idx.DataOp.DataRefs = dataRefs
		return w.iw.WriteHeader(hdr)
	}
}

// StartTag starts a tag for the next set of bytes (used for the reverse index, mapping file output to datums).
func (w *Writer) StartTag(id string) {
	w.hdr.Idx.DataOp.Tags = append(w.hdr.Idx.DataOp.Tags, &index.Tag{Id: id})
}

// Write writes to the current file in the tar stream.
func (w *Writer) Write(data []byte) (int, error) {
	n, err := w.tw.Write(data)
	w.hdr.Idx.SizeBytes += int64(n)
	w.hdr.Idx.DataOp.Tags[len(w.hdr.Idx.DataOp.Tags)-1].SizeBytes += int64(n)
	return n, err
}

func (w *Writer) writeTags(tags []*index.Tag) error {
	w.hdr.Idx.DataOp.Tags = append(w.hdr.Idx.DataOp.Tags, tags...)
	var numBytes int64
	for _, tag := range tags {
		numBytes += tag.SizeBytes
	}
	w.hdr.Idx.SizeBytes += numBytes
	return w.tw.Skip(numBytes)
}

// Close closes the writer.
// (bryce) is not closing the tar writer the right choice here?
// We cannot include it in the last file's range because it will
// effect the hash.
// Our indexing is what will exit the reading of the tar stream,
// not the end of tar entry (two empty 512 bytes).
func (w *Writer) Close() error {
	// Flush the last file's content.
	if err := w.tw.Flush(); err != nil {
		return err
	}
	// Close chunk and index writer.
	if err := w.cw.Close(); err != nil {
		return err
	}
	return w.iw.Close()
}
