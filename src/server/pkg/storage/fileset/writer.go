package fileset

import (
	"context"
	"io"
	"math"

	"github.com/gogo/protobuf/proto"
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
	ctx          context.Context
	chunks       *chunk.Storage
	tw           *tar.Writer
	cw           *chunk.Writer
	iw           *index.Writer
	hdr, lastHdr *index.Header
	closed       bool
}

func newWriter(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string) *Writer {
	w := &Writer{
		ctx:    ctx,
		chunks: chunks,
		iw:     index.NewWriter(ctx, objC, chunks, path),
	}
	cw := chunks.NewWriter(ctx, averageBits, w.callback(), math.MaxInt64)
	w.cw = cw
	w.tw = tar.NewWriter(cw)
	return w
}

// WriteHeader writes a tar header and prepares to accept the file's contents.
func (w *Writer) WriteHeader(hdr *index.Header) error {
	// Flush out preceding file's content.
	if err := w.tw.Flush(); err != nil {
		return err
	}
	// (bryce) need a deep copy library for this.
	w.hdr = &index.Header{Hdr: hdr.Hdr}
	// Setup annotation in chunk writer.
	w.cw.Annotate(&chunk.Annotation{
		NextDataRef: &chunk.DataRef{},
		Meta: &meta{
			hdr: w.hdr,
		},
	})
	// The index should only be non-nil when cheap copying.
	if hdr.Idx != nil {
		w.hdr.Idx = proto.Clone(hdr.Idx).(*index.Index)
		// Data references will be set by the callback.
		w.hdr.Idx.DataOp.DataRefs = nil
		return nil
	}
	// Write content header.
	if err := w.tw.WriteHeader(w.hdr.Hdr); err != nil {
		return err
	}
	// Setup index header for the file.
	w.hdr.Idx = &index.Index{DataOp: &index.DataOp{}}
	// Setup header tag for the file.
	w.hdr.Idx.DataOp.Tags = []*index.Tag{&index.Tag{Id: headerTag, SizeBytes: w.cw.AnnotatedBytesSize()}}
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
	w.hdr.Idx.DataOp.Tags[len(w.hdr.Idx.DataOp.Tags)-1].SizeBytes += int64(n)
	return n, err
}

// CopyTags does a cheap copy of tagged file data from a reader to a writer.
func (w *Writer) CopyTags(r *Reader, tagBound ...string) error {
	c, err := r.readCopyTags(tagBound...)
	if err != nil {
		return err
	}
	return w.writeCopyTags(c)
}

func (w *Writer) writeCopyTags(c *copyTags) error {
	beforeSize := w.cw.AnnotatedBytesSize()
	if err := w.cw.WriteCopy(c.content); err != nil {
		return err
	}
	w.hdr.Idx.DataOp.Tags = append(w.hdr.Idx.DataOp.Tags, c.tags...)
	return w.tw.Skip(w.cw.AnnotatedBytesSize() - beforeSize)
}

// CopyFiles does a cheap copy of files from a reader to a writer.
// (bryce) need to handle delete operations.
func (w *Writer) CopyFiles(r *Reader, pathBound ...string) (retErr error) {
	var cheapCopy bool
	slr := r.newSplitLimitReader()
	for {
		hdr, err := r.Peek()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		// If the header is past the path bound, then we are done.
		if !index.BeforeBound(hdr.Hdr.Name, pathBound...) {
			return nil
		}
		if _, err = r.Next(); err != nil {
			return err
		}
		// If the last path in a referenced chunk is past the path bound,
		// then we are done cheap copying index entries.
		if !index.BeforeBound(hdr.Idx.LastPathChunk, pathBound...) {
			cheapCopy = false
		}
		if cheapCopy {
			if err := w.iw.WriteHeaders([]*index.Header{hdr}); err != nil {
				return err
			}
			continue
		}
		if err := w.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := io.Copy(w.cw, slr); err != nil && err != io.EOF {
			return err
		}
		if slr.AtSplit() {
			if err := w.finishFile(hdr); err != nil {
				return err
			}
			slr.Reset()
			cheapCopy = true
		}
	}
}

func (w *Writer) finishFile(hdr *index.Header) error {
	// Flush the current file's content.
	if err := w.cw.Flush(); err != nil {
		return err
	}
	// Update the current file's index entry with the pre-copy data references.
	hdr.Idx.DataOp.DataRefs = updateDataRefs(hdr.Idx.DataOp.DataRefs, w.lastHdr.Idx.DataOp.DataRefs)
	if err := w.iw.WriteHeaders([]*index.Header{hdr}); err != nil {
		return err
	}
	w.tw = tar.NewWriter(w.cw)
	w.lastHdr = nil
	return nil
}

func updateDataRefs(dataRefs, updateDataRefs []*chunk.DataRef) []*chunk.DataRef {
	var updateSize int64
	for _, dataRef := range updateDataRefs {
		updateSize += dataRef.SizeBytes
	}
	var size int64
	for i, dataRef := range dataRefs {
		size += dataRef.SizeBytes
		if size == updateSize {
			return append(updateDataRefs, dataRefs[i+1:]...)
		}
	}
	return nil
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
