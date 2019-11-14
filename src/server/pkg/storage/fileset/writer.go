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
	ctx     context.Context
	tw      *tar.Writer
	cw      *chunk.Writer
	iw      *index.Writer
	lastIdx *index.Index
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
	// Finish prior file.
	if err := finishFile(); err != nil {
		return err
	}
	// Setup annotation in chunk writer.
	w.setupAnnotation(hdr.Name)
	// Setup header tag for the file.
	w.cw.Tag(headerTag)
	// Write file header.
	return w.tw.WriteHeader(hdr)
}

func (w *Writer) setupAnnotation(path string) {
	w.cw.Annotate(&chunk.Annotation{
		NextDataRef: &chunk.DataRef{},
		Meta: &meta{
			idx: &index.Index{
				Path:   path,
				DataOp: &index.DataOp{},
			},
		},
	})
}

func (w *Writer) finishFile() {
	w.cw.FinishTag()
	// Flush the last file's content.
	if err := w.tw.Flush(); err != nil {
		return err
	}
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
			for _, tag := range annotations[i].NextDataRef.Tags {
				idx.SizeBytes += int64(tag.SizeBytes)
			}
			idxs = append(idxs, idx)
		}
		// Don't write out the last file index (it may have more content in the next chunk).
		idxs = idxs[:len(idxs)-1]
		return w.iw.WriteIndexes(idxs)
	}
}

// StartTag starts a tag for the next set of bytes (used for the reverse index, mapping file output to datums).
func (w *Writer) StartTag(id string) {
	w.cw.StartTag(id)
}

// Write writes to the current file in the tar stream.
func (w *Writer) Write(data []byte) (int, error) {
	return w.tw.Write(data)
}

func (w *Writer) CopyFile(fr *FileReader) error {
	w.setupAnnotation(fr.Index().Path)
	return fr.cr.Iterate(func(dr *DataReader) error {
		return w.cw.Copy(dr)
	})
}

func (w *Writer) CopyTags(dr *DataReader, tagBound ...string) error {
	return w.cw.Copy(dr, tagBound...)
}

// Close closes the writer.
func (w *Writer) Close() error {
	// Finish prior file.
	if err := finishFile(); err != nil {
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
