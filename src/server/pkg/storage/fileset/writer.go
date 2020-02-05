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
	ctx                   context.Context
	chunks                *chunk.Storage
	tw                    *tar.Writer
	cw                    *chunk.Writer
	iw                    *index.Writer
	hdr, copyHdr, lastHdr *index.Header
	closed                bool
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
	// Setup index header for next file.
	if hdr.Idx == nil {
		hdr.Idx = &index.Index{}
	}
	w.hdr = &index.Header{
		Hdr: hdr.Hdr,
		Idx: proto.Clone(hdr.Idx).(*index.Index),
	}
	w.cw.Annotate(&chunk.Annotation{
		NextDataRef: &chunk.DataRef{},
		Meta: &meta{
			hdr: w.hdr,
		},
	})
	// The DataOp should only be non-nil when copying.
	if w.hdr.Idx.DataOp == nil {
		w.hdr.Idx.DataOp = &index.DataOp{}
	} else {
		w.hdr.Idx.DataOp.DataRefs = nil
		w.copyHdr = hdr
	}
	if err := w.tw.WriteHeader(w.hdr.Hdr); err != nil {
		return err
	}
	// Setup first tag for header.
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
func (w *Writer) CopyFiles(r *Reader, pathBound ...string) error {
	f := r.readCopyFiles(pathBound...)
	for {
		c, err := f()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		// Copy the content level.
		for _, file := range c.files {
			if err := w.WriteHeader(file.hdr); err != nil {
				return err
			}
			if err := w.writeCopyTags(file); err != nil {
				return err
			}
		}
		// Copy the index level(s).
		if c.indexCopyF != nil {
			if err := w.finishFile(); err != nil {
				return err
			}
			if err := w.iw.WriteCopyFunc(c.indexCopyF); err != nil {
				return err
			}
		}
	}
}

func (w *Writer) finishFile() error {
	// Pull the last data reference and the corresponding last path in the last chunk from the copy header.
	if err := w.cw.Flush(); err != nil {
		return err
	}
	// We need to delete the last chunk after the flush because it will be contained within the copied chunks.
	if err := w.chunks.Delete(w.ctx, w.lastHdr.Idx.DataOp.DataRefs[len(w.lastHdr.Idx.DataOp.DataRefs)-1].Chunk.Hash); err != nil {
		return err
	}
	w.lastHdr.Idx.DataOp.DataRefs = w.lastHdr.Idx.DataOp.DataRefs[:len(w.lastHdr.Idx.DataOp.DataRefs)-1]
	dataRefs := w.lastHdr.Idx.DataOp.DataRefs
	copyDataRefs := w.copyHdr.Idx.DataOp.DataRefs
	if len(dataRefs) == 0 || dataRefs[len(dataRefs)-1].Chunk.Hash != copyDataRefs[len(copyDataRefs)-1].Chunk.Hash {
		w.lastHdr.Idx.DataOp.DataRefs = append(w.lastHdr.Idx.DataOp.DataRefs, copyDataRefs[len(copyDataRefs)-1])
	}
	w.lastHdr.Idx.LastPathChunk = w.copyHdr.Idx.LastPathChunk
	// Write the last hdr and clear the content level writers.
	w.iw.WriteHeaders([]*index.Header{w.lastHdr})
	w.lastHdr = nil
	w.tw = tar.NewWriter(w.cw)
	w.cw.Reset()
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
