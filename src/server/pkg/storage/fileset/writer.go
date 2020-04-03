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

type data struct {
	idx *index.Index
}

// Writer writes the serialized format of a fileset.
// The serialized format of a fileset consists of indexes and content.
type Writer struct {
	ctx       context.Context
	tw        *tar.Writer
	cw        *chunk.Writer
	iw        *index.Writer
	indexFunc func(*index.Index) error
	lastIdx   *index.Index
	first     bool
}

func newWriter(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string, indexFunc func(*index.Index) error) *Writer {
	w := &Writer{
		ctx:       ctx,
		first:     true,
		indexFunc: indexFunc,
	}
	if w.indexFunc == nil {
		w.iw = index.NewWriter(ctx, objC, chunks, path)
	}
	cw := chunks.NewWriter(ctx, averageBits, math.MaxInt64, indexFunc != nil, w.callback())
	w.cw = cw
	w.tw = tar.NewWriter(cw)
	return w
}

// WriteHeader writes a tar header and prepares to accept the file's contents.
func (w *Writer) WriteHeader(hdr *tar.Header) error {
	// Finish prior file.
	if err := w.finishFile(); err != nil {
		return err
	}
	// Setup annotation in chunk writer.
	w.setupAnnotation(hdr.Name)
	// Setup header tag for the file.
	w.cw.Tag(headerTag)
	// Write file header.
	return w.tw.WriteHeader(hdr)
}

func (w *Writer) finishFile() error {
	if w.first {
		w.first = false
		return nil
	}
	w.cw.Tag(paddingTag)
	// Flush the last file's content.
	return w.tw.Flush()
}

func (w *Writer) setupAnnotation(path string) {
	w.cw.Annotate(&chunk.Annotation{
		NextDataRef: &chunk.DataRef{},
		Data: &data{
			idx: &index.Index{
				Path:   path,
				DataOp: &index.DataOp{},
			},
		},
	})
}

func (w *Writer) callback() chunk.WriterFunc {
	return func(annotations []*chunk.Annotation) error {
		if len(annotations) == 0 {
			return nil
		}
		var idxs []*index.Index
		// Edge case where the last file from the prior chunk ended at the chunk split point.
		firstIdx := annotations[0].Data.(*data).idx
		if w.lastIdx != nil && firstIdx.Path != w.lastIdx.Path {
			idxs = append(idxs, w.lastIdx)
		}
		w.lastIdx = annotations[len(annotations)-1].Data.(*data).idx
		// Update the file indexes.
		for i := 0; i < len(annotations); i++ {
			idx := annotations[i].Data.(*data).idx
			idx.DataOp.DataRefs = append(idx.DataOp.DataRefs, annotations[i].NextDataRef)
			for _, tag := range annotations[i].NextDataRef.Tags {
				if tag.Id != headerTag && tag.Id != paddingTag {
					idx.SizeBytes += int64(tag.SizeBytes)
				}
			}
			idxs = append(idxs, idx)
		}
		// Don't write out the last file index (it may have more content in the next chunk).
		idxs = idxs[:len(idxs)-1]
		if w.indexFunc != nil {
			for _, idx := range idxs {
				if err := w.indexFunc(idx); err != nil {
					return err
				}
			}
			return nil
		}
		return w.iw.WriteIndexes(idxs)
	}
}

// Tag starts a tag for the next set of bytes (used for the reverse index, mapping file output to datums).
func (w *Writer) Tag(id string) {
	w.cw.Tag(id)
}

// Write writes to the current file in the tar stream.
func (w *Writer) Write(data []byte) (int, error) {
	return w.tw.Write(data)
}

// CopyFile copies a file (header and tags included).
func (w *Writer) CopyFile(fr *FileReader) error {
	// Finish prior file.
	if err := w.finishFile(); err != nil {
		return err
	}
	w.setupAnnotation(fr.Index().Path)
	return fr.Iterate(func(dr *chunk.DataReader) error {
		return w.cw.Copy(dr)
	})
}

// CopyTags copies the tagged data from the passed in data reader.
func (w *Writer) CopyTags(dr *chunk.DataReader) error {
	if err := w.tw.Skip(dr.Len()); err != nil {
		return err
	}
	return w.cw.Copy(dr)
}

// Close closes the writer.
func (w *Writer) Close() error {
	// Finish prior file.
	if err := w.finishFile(); err != nil {
		return err
	}
	// Close the chunk writer.
	if err := w.cw.Close(); err != nil {
		return err
	}
	// Write out the last index.
	if w.lastIdx != nil {
		if w.indexFunc != nil {
			if err := w.indexFunc(w.lastIdx); err != nil {
				return err
			}
		} else if err := w.iw.WriteIndexes([]*index.Index{w.lastIdx}); err != nil {
			return err
		}
	}
	if w.indexFunc != nil {
		return nil
	}
	// Close the index writer.
	return w.iw.Close()
}
