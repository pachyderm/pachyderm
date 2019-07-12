package fileset

import (
	"context"
	"math"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

var (
	averageBits = 23
)

type meta struct {
	hdr *index.Header
}

// Writer writes the serialized format of a fileset.
// The serialized format of a fileset consists of indexes and content which are both realized as compressed tar stream chunks.
type Writer struct {
	tw      *tar.Writer
	cw      *chunk.Writer
	iw      *index.Writer
	first   bool
	hdr     *index.Header
	lastHdr *index.Header
	closed  bool
}

func newWriter(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string) *Writer {
	w := &Writer{
		iw:    index.NewWriter(ctx, objC, chunks, path),
		first: true,
	}
	cw := chunks.NewWriter(ctx, averageBits, w.callback(), math.MaxInt64)
	w.cw = cw
	w.tw = tar.NewWriter(cw)
	return w
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
	w.hdr = &index.Header{Hdr: hdr.Hdr}
	if w.hdr.Idx == nil {
		w.hdr.Idx = &index.Index{}
	}
	w.hdr.Idx.DataOp = &index.DataOp{}
	w.cw.Annotate(&chunk.Annotation{
		NextDataRef: &chunk.DataRef{},
		Meta: &meta{
			hdr: w.hdr,
		},
	})
	if err := w.tw.WriteHeader(w.hdr.Hdr); err != nil {
		return err
	}
	// Setup first tag for header.
	w.hdr.Idx.DataOp.Tags = []*index.Tag{&index.Tag{Id: headerTag, SizeBytes: w.cw.AnnotationSize()}}
	return nil
}

func (w *Writer) callback() chunk.WriterFunc {
	return func(_ *chunk.DataRef, annotations []*chunk.Annotation) error {
		if len(annotations) == 0 {
			return nil
		}
		var hdrs []*index.Header
		// Edge case where the last file from the prior chunk ended at the chunk split point.
		firstHdr := annotations[0].Meta.(*meta).hdr
		if w.lastHdr != nil && firstHdr.Hdr.Name != w.lastHdr.Hdr.Name {
			hdrs = append(hdrs, w.lastHdr)
		}
		w.lastHdr = annotations[len(annotations)-1].Meta.(*meta).hdr
		// Update the file headers.
		for i := 0; i < len(annotations); i++ {
			hdr := annotations[i].Meta.(*meta).hdr
			hdr.Idx.DataOp.DataRefs = append(hdr.Idx.DataOp.DataRefs, annotations[i].NextDataRef)
			hdr.Idx.LastPathChunk = w.lastHdr.Hdr.Name
			hdrs = append(hdrs, hdr)
		}
		// Don't write out the last file header (it may have more content in the next chunk).
		hdrs = hdrs[:len(hdrs)-1]
		return w.iw.WriteHeaders(hdrs)
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
	if len(w.hdr.Idx.DataOp.Tags) > 1 {
		w.hdr.Idx.DataOp.Tags[len(w.hdr.Idx.DataOp.Tags)-1].SizeBytes += int64(n)
	}
	return n, err
}

func (w *Writer) finishFile(hdr *index.Header) error {
	// (bryce) this needs to be changed when the chunk storage layer is parallelized.
	w.lastHdr = nil
	w.hdr.Idx.DataOp.DataRefs = append(w.hdr.Idx.DataOp.DataRefs, hdr.Idx.DataOp.DataRefs[1:]...)
	w.hdr.Idx.DataOp.Tags = append(w.hdr.Idx.DataOp.Tags, hdr.Idx.DataOp.Tags[1:]...)
	w.hdr.Idx.SizeBytes = hdr.Idx.SizeBytes
	w.hdr.Idx.LastPathChunk = hdr.Idx.LastPathChunk
	return w.iw.WriteHeaders([]*index.Header{w.hdr})
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
	w.closed = true
	// Flush the last file's content.
	if err := w.tw.Flush(); err != nil {
		return err
	}
	// Close the chunk writer.
	if err := w.cw.Close(); err != nil {
		return err
	}
	// Write out the last header.
	if w.lastHdr != nil {
		if err := w.iw.WriteHeaders([]*index.Header{w.lastHdr}); err != nil {
			return err
		}
	}
	// Close the index writer.
	return w.iw.Close()
}
