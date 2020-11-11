package fileset

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/tracker"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

type FileWriter struct {
	idx  *index.Index
	cw   *chunk.Writer
	part *index.Part
}

func (fw *FileWriter) Append(tag string) {
	fw.idx.File.Parts = append(fw.idx.File.Parts, &index.Part{Tag: tag})
}

func (fw *FileWriter) Write(data []byte) (int, error) {
	if fw.part.Tag != headerTag {
		fw.idx.SizeBytes += int64(len(data))
	}
	fw.part.SizeBytes += int64(len(data))
	return fw.cw.Write(data)
}

// Writer provides functionality for writing a file set.
type Writer struct {
	ctx                context.Context
	tracker            tracker.Tracker
	store              Store
	path               string
	additive, deletive *index.Writer
	cw                 *chunk.Writer
	idx                *index.Index
	lastIdx            *index.Index
	noUpload           bool
	indexFunc          func(*index.Index) error
	ttl                time.Duration
}

func newWriter(ctx context.Context, store Store, tracker tracker.Tracker, chunks *chunk.Storage, path string, opts ...WriterOption) *Writer {
	uuidStr := uuid.NewWithoutDashes()
	w := &Writer{
		ctx:     ctx,
		store:   store,
		tracker: tracker,
		path:    path,
	}
	for _, opt := range opts {
		opt(w)
	}
	var chunkWriterOpts []chunk.WriterOption
	if w.noUpload {
		chunkWriterOpts = append(chunkWriterOpts, chunk.WithNoUpload())
	}
	w.additive = index.NewWriter(ctx, chunks, "additive-index-writer-"+uuidStr)
	w.deletive = index.NewWriter(ctx, chunks, "deletive-index-writer-"+uuidStr)
	w.cw = chunks.NewWriter(ctx, "chunk-writer-"+uuidStr, w.callback, chunkWriterOpts...)
	return w
}

// Append creates an append operation for a file and provides a scoped file writer.
func (w *Writer) Append(p string, cb func(*FileWriter) error) error {
	fw, err := w.newFileWriter(p, w.cw)
	if err != nil {
		return err
	}
	return cb(fw)
}

func (w *Writer) newFileWriter(p string, cw *chunk.Writer) (*FileWriter, error) {
	idx := &index.Index{
		Path: p,
		File: &index.File{},
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
// TODO: Check path order.
func (w *Writer) Delete(p string, tags ...string) error {
	idx := &index.Index{
		Path: p,
		File: &index.File{},
	}
	for _, tag := range tags {
		idx.File.Parts = append(idx.File.Parts, &index.Part{Tag: tag})
	}
	return w.deletive.WriteIndex(idx)
}

// Copy copies a file to the file set writer.
func (w *Writer) Copy(file File) error {
	idx := file.Index()
	copyIdx := &index.Index{
		Path: idx.Path,
		File: &index.File{
			Parts: idx.File.Parts,
		},
		SizeBytes: idx.SizeBytes,
	}
	if err := w.nextIdx(copyIdx); err != nil {
		return err
	}
	// Copy the file data refs if they are resolved.
	if idx.File.DataRefs != nil {
		for _, dataRef := range idx.File.DataRefs {
			if err := w.cw.Copy(dataRef); err != nil {
				return err
			}
		}
		return nil
	}
	// Copy the file part data refs otherwise.
	for _, part := range idx.File.Parts {
		for _, dataRef := range part.DataRefs {
			if err := w.cw.Copy(dataRef); err != nil {
				return err
			}
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
				if err := w.additive.WriteIndex(w.lastIdx); err != nil {
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
			w.lastIdx.File.DataRefs = append(w.lastIdx.File.DataRefs, annotation.NextDataRef)
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
			if err := w.additive.WriteIndex(idx); err != nil {
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
	// Close the index writers.
	additiveIdx, err := w.additive.Close()
	if err != nil {
		return err
	}
	deletiveIdx, err := w.deletive.Close()
	if err != nil {
		return err
	}
	var pointsTo []string
	for _, cid := range index.PointsTo(additiveIdx) {
		pointsTo = append(pointsTo, chunk.ChunkObjectID(cid))
	}
	for _, cid := range index.PointsTo(deletiveIdx) {
		pointsTo = append(pointsTo, chunk.ChunkObjectID(cid))
	}
	if err := w.tracker.CreateObject(w.ctx, filesetObjectID(w.path), pointsTo, w.ttl); err != nil {
		return err
	}
	return w.store.Set(w.ctx, w.path, &Metadata{
		Path:     w.path,
		Additive: additiveIdx,
		Deletive: deletiveIdx,
	})
}
