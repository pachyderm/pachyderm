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
	idx *index.Index
}

// Writer writes the serialized format of a fileset.
// The serialized format of a fileset consists of indexes and content.
type Writer struct {
	ctx          context.Context
	tw           *tar.Writer
	cw           *chunk.Writer
	iw           *index.Writer
	idx, lastIdx *index.Index
}

func newWriter(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string) *Writer {
	w := &Writer{
		ctx: ctx,
		iw:  index.NewWriter(ctx, objC, chunks, path),
	}
	cw := chunks.NewWriter(ctx, averageBits, w.callback(), math.MaxInt64)
	w.cw = cw
	w.tw = tar.NewWriter(cw)
	return w
}

// WriteHeader writes a tar header and prepares to accept the file's contents.
func (w *Writer) WriteHeader(hdr *tar.Header) error {
	// Flush out preceding file's content.
	if err := w.tw.Flush(); err != nil {
		return err
	}
	w.idx = &index.Index{
		Path:   hdr.Name,
		DataOp: &index.DataOp{},
	}
	// Setup annotation in chunk writer.
	w.cw.Annotate(&chunk.Annotation{
		NextDataRef: &chunk.DataRef{},
		Meta:        &meta{idx: w.idx},
	})
	// Write file header.
	if err := w.tw.WriteHeader(hdr); err != nil {
		return err
	}
	// Setup header tag for the file.
	w.idx.DataOp.Tags = []*index.Tag{&index.Tag{Id: headerTag, SizeBytes: w.cw.AnnotatedBytesSize()}}
	return nil
}

func (w *Writer) callback() chunk.WriterFunc {
	return func(_ *chunk.DataRef, annotations []*chunk.Annotation) error {
		if len(annotations) == 0 {
			return nil
		}
		var idxs []*index.Index
		// Edge case where the last file from the prior chunk ended at the chunk split point.
		firstIdx := annotations[0].Meta.(*meta).idx
		if w.lastIdx != nil && firstIdx.Path != w.lastIdx.Path {
			idxs = append(idxs, w.lastIdx)
		}
		w.lastIdx = annotations[len(annotations)-1].Meta.(*meta).idx
		// Update the file indexes.
		for i := 0; i < len(annotations); i++ {
			idx := annotations[i].Meta.(*meta).idx
			idx.DataOp.DataRefs = append(idx.DataOp.DataRefs, annotations[i].NextDataRef)
			idx.LastPathChunk = w.lastIdx.Path
			idxs = append(idxs, idx)
		}
		// Don't write out the last file index (it may have more content in the next chunk).
		idxs = idxs[:len(idxs)-1]
		return w.iw.WriteIndexes(idxs)
	}
}

// StartTag starts a tag for the next set of bytes (used for the reverse index, mapping file output to datums).
func (w *Writer) StartTag(id string) {
	w.idx.DataOp.Tags = append(w.idx.DataOp.Tags, &index.Tag{Id: id})
}

// Write writes to the current file in the tar stream.
func (w *Writer) Write(data []byte) (int, error) {
	n, err := w.tw.Write(data)
	w.idx.SizeBytes += int64(n)
	w.idx.DataOp.Tags[len(w.idx.DataOp.Tags)-1].SizeBytes += int64(n)
	return n, err
}

// Close closes the writer.
func (w *Writer) Close() error {
	// Flush the last file's content.
	if err := w.tw.Flush(); err != nil {
		return err
	}
	// Close the chunk writer.
	if err := w.cw.Close(); err != nil {
		return err
	}
	// Write out the last index.
	if w.lastIdx != nil {
		if err := w.iw.WriteIndexes([]*index.Index{w.lastIdx}); err != nil {
			return err
		}
	}
	// Close the index writer.
	return w.iw.Close()
}

// (bryce) commented out for now to checkpoint changes to reader/writer
//// CopyTags does a cheap copy of tagged file data from a reader to a writer.
//func (w *Writer) CopyTags(r *Reader, tagBound ...string) error {
//	c, err := r.readCopyTags(tagBound...)
//	if err != nil {
//		return err
//	}
//	return w.writeCopyTags(c)
//}
//
//func (w *Writer) writeCopyTags(c *copyTags) error {
//	beforeSize := w.cw.AnnotatedBytesSize()
//	if err := w.cw.WriteCopy(c.content); err != nil {
//		return err
//	}
//	w.hdr.Idx.DataOp.Tags = append(w.hdr.Idx.DataOp.Tags, c.tags...)
//	return w.tw.Skip(w.cw.AnnotatedBytesSize() - beforeSize)
//}
//
//// CopyFiles does a cheap copy of files from a reader to a writer.
//// (bryce) need to handle delete operations.
//func (w *Writer) CopyFiles(r *Reader, pathBound ...string) error {
//	f := r.readCopyFiles(pathBound...)
//	for {
//		c, err := f()
//		if err != nil {
//			if err == io.EOF {
//				return nil
//			}
//			return err
//		}
//		// Copy the content level.
//		for _, file := range c.files {
//			if err := w.WriteHeader(file.hdr); err != nil {
//				return err
//			}
//			if err := w.writeCopyTags(file); err != nil {
//				return err
//			}
//		}
//		// Copy the index level(s).
//		if c.indexCopyF != nil {
//			if err := w.finishFile(); err != nil {
//				return err
//			}
//			if err := w.iw.WriteCopyFunc(c.indexCopyF); err != nil {
//				return err
//			}
//		}
//	}
//}
//
//func (w *Writer) finishFile() error {
//	// Pull the last data reference and the corresponding last path in the last chunk from the copy header.
//	if err := w.cw.Flush(); err != nil {
//		return err
//	}
//	// We need to delete the last chunk after the flush because it will be contained within the copied chunks.
//	if err := w.chunks.Delete(w.ctx, w.lastHdr.Idx.DataOp.DataRefs[len(w.lastHdr.Idx.DataOp.DataRefs)-1].Chunk.Hash); err != nil {
//		return err
//	}
//	w.lastHdr.Idx.DataOp.DataRefs = w.lastHdr.Idx.DataOp.DataRefs[:len(w.lastHdr.Idx.DataOp.DataRefs)-1]
//	dataRefs := w.lastHdr.Idx.DataOp.DataRefs
//	copyDataRefs := w.copyHdr.Idx.DataOp.DataRefs
//	if len(dataRefs) == 0 || dataRefs[len(dataRefs)-1].Chunk.Hash != copyDataRefs[len(copyDataRefs)-1].Chunk.Hash {
//		w.lastHdr.Idx.DataOp.DataRefs = append(w.lastHdr.Idx.DataOp.DataRefs, copyDataRefs[len(copyDataRefs)-1])
//	}
//	w.lastHdr.Idx.LastPathChunk = w.copyHdr.Idx.LastPathChunk
//	// Write the last hdr and clear the content level writers.
//	w.iw.WriteHeaders([]*index.Header{w.lastHdr})
//	w.lastHdr = nil
//	w.tw = tar.NewWriter(w.cw)
//	w.cw.Reset()
//	return nil
//}
