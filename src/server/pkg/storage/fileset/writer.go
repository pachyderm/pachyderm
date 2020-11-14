package fileset

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

// FileWriter provides functionality for writing a file.
type FileWriter struct {
	idx    *index.Index
	cw     *chunk.Writer
	dataOp *index.DataOp
}

// Append sets an append tag for the next set of bytes.
func (fw *FileWriter) Append(tag string) {
	fw.newDataOp(index.Op_APPEND, tag)
}

// Overwrite sets an overwrite tag for the next set of bytes.
func (fw *FileWriter) Overwrite(tag string) {
	fw.newDataOp(index.Op_OVERWRITE, tag)
}

// Delete sets a delete tag.
func (fw *FileWriter) Delete(tag string) {
	fw.newDataOp(index.Op_DELETE, tag)
}

func (fw *FileWriter) newDataOp(op index.Op, tag string) {
	fw.dataOp = &index.DataOp{
		Op:  op,
		Tag: tag,
	}
	fw.idx.FileOp.DataOps = append(fw.idx.FileOp.DataOps, fw.dataOp)
}

// Copy copies a set of data ops to the file writer.
func (fw *FileWriter) Copy(dataOps []*index.DataOp) error {
	fw.idx.FileOp.DataOps = append(fw.idx.FileOp.DataOps, dataOps...)
	for _, dataOp := range dataOps {
		for _, dataRef := range dataOp.DataRefs {
			fw.idx.SizeBytes += int64(dataRef.SizeBytes)
			if err := fw.cw.Copy(dataRef); err != nil {
				return err
			}
		}
	}
	return nil
}

func (fw *FileWriter) Write(data []byte) (int, error) {
	if fw.dataOp.Tag != headerTag && fw.dataOp.Tag != paddingTag {
		fw.idx.SizeBytes += int64(len(data))
	}
	fw.dataOp.SizeBytes += int64(len(data))
	return fw.cw.Write(data)
}

// Writer provides functionality for writing a file set.
type Writer struct {
	ctx       context.Context
	iw        *index.Writer
	cw        *chunk.Writer
	idx       *index.Index
	lastIdx   *index.Index
	noUpload  bool
	indexFunc func(*index.Index) error
	ttl       time.Duration
}

func newWriter(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string, opts ...WriterOption) *Writer {
	tmpID := path + uuid.NewWithoutDashes()
	w := &Writer{ctx: ctx}
	for _, opt := range opts {
		opt(w)
	}
	var chunkWriterOpts []chunk.WriterOption
	if w.noUpload {
		chunkWriterOpts = append(chunkWriterOpts, chunk.WithNoUpload())
	}
	var indexWriterOpts []index.WriterOption
	if w.ttl > 0 {
		indexWriterOpts = append(indexWriterOpts, index.WithRootTTL(w.ttl))
	}
	w.iw = index.NewWriter(ctx, objC, chunks, path, tmpID, indexWriterOpts...)
	w.cw = chunks.NewWriter(ctx, tmpID, w.callback, chunkWriterOpts...)
	return w
}

// Append creates an append operation for a file and provides a scoped file writer.
func (w *Writer) Append(p string, cb func(*FileWriter) error) error {
	fw, err := w.newFileWriter(p, index.Op_APPEND, w.cw)
	if err != nil {
		return err
	}
	return cb(fw)
}

// Overwrite creates an overwrite operation for a file and provides a scoped file writer.
func (w *Writer) Overwrite(p string, cb func(*FileWriter) error) error {
	fw, err := w.newFileWriter(p, index.Op_OVERWRITE, w.cw)
	if err != nil {
		return err
	}
	return cb(fw)
}

func (w *Writer) newFileWriter(p string, op index.Op, cw *chunk.Writer) (*FileWriter, error) {
	idx := &index.Index{
		Path: p,
		FileOp: &index.FileOp{
			Op: op,
		},
	}
	if err := w.nextIdx(idx); err != nil {
		return nil, err
	}
	return &FileWriter{
		idx: idx,
		cw:  cw,
	}, nil
}

func (w *Writer) nextIdx(idx *index.Index) error {
	if err := w.checkPath(idx.Path); err != nil {
		return err
	}
	return w.cw.Annotate(&chunk.Annotation{
		Data: idx,
	})
}

func (w *Writer) checkPath(p string) error {
	if w.idx != nil {
		if w.idx.Path > p {
			return errors.Errorf("can't write path (%s) after (%s)", p, w.idx.Path)
		}
		parent := parentOf(p)
		if w.idx.Path < parent {
			return errors.Errorf("cannot write path (%s) without first writing parent (%s)", p, parent)
		}
	}
	return nil
}

// Delete creates a delete operation for a file.
func (w *Writer) Delete(p string) error {
	idx := &index.Index{
		Path: p,
		FileOp: &index.FileOp{
			Op: index.Op_DELETE,
		},
	}
	return w.nextIdx(idx)
}

// Copy copies a file to the file set writer.
func (w *Writer) Copy(file File) error {
	idx := file.Index()
	if idx.FileOp.DataRefs != nil {
		return w.copyIndex(idx)
	}
	switch idx.FileOp.Op {
	case index.Op_APPEND:
		return w.Append(idx.Path, func(fw *FileWriter) error {
			hdr, err := file.Header()
			if err != nil {
				return err
			}
			return WithTarFileWriter(fw, hdr, func(tfw *TarFileWriter) error {
				return tfw.Copy(getContentDataOps(idx.FileOp.DataOps))
			})
		})
	case index.Op_OVERWRITE:
		return w.Overwrite(idx.Path, func(fw *FileWriter) error {
			hdr, err := file.Header()
			if err != nil {
				return err
			}
			return WithTarFileWriter(fw, hdr, func(tfw *TarFileWriter) error {
				return tfw.Copy(getContentDataOps(idx.FileOp.DataOps))
			})
		})
	case index.Op_DELETE:
		return w.Delete(idx.Path)
	}
	return nil
}

func (w *Writer) copyIndex(idx *index.Index) error {
	copyIdx := &index.Index{
		Path: idx.Path,
		FileOp: &index.FileOp{
			Op:      idx.FileOp.Op,
			DataOps: idx.FileOp.DataOps,
		},
		SizeBytes: idx.SizeBytes,
	}
	if err := w.nextIdx(copyIdx); err != nil {
		return err
	}
	for _, dataRef := range idx.FileOp.DataRefs {
		if err := w.cw.Copy(dataRef); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) callback(annotations []*chunk.Annotation) error {
	for _, annotation := range annotations {
		idx := annotation.Data.(*index.Index)
		if w.lastIdx == nil {
			w.lastIdx = idx
		}
		if idx.Path != w.lastIdx.Path {
			if !w.noUpload {
				if err := w.iw.WriteIndex(w.lastIdx); err != nil {
					return err
				}
			}
			if w.indexFunc != nil {
				if err := w.indexFunc(w.lastIdx); err != nil {
					return err
				}
			}
			w.lastIdx = idx
		}
		if annotation.NextDataRef != nil {
			w.lastIdx.FileOp.DataRefs = append(w.lastIdx.FileOp.DataRefs, annotation.NextDataRef)
		}
	}
	return nil
}

// Close closes the writer.
func (w *Writer) Close() error {
	if err := w.cw.Close(); err != nil {
		return err
	}
	// Write out the last index.
	if w.lastIdx != nil {
		idx := w.lastIdx
		if !w.noUpload {
			if err := w.iw.WriteIndex(idx); err != nil {
				return err
			}
		}
		if w.indexFunc != nil {
			if err := w.indexFunc(idx); err != nil {
				return err
			}
		}
	}
	if w.noUpload {
		return nil
	}
	return w.iw.Close()
}
