package fileset

import (
	"context"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

// TODO: Size zero files need to be addressed now that we are moving away from storing tar headers.
// We can run into the same issue as deletions where a lot of size zero files can cause us to get backed up
// since no chunks will get created. The solution we have in mind is to write a small number of bytes
// for a size zero file, then either not store references to them or ignore them at read time.

// TODO: Might need to think a bit more about fileset sizes and whether deletes should be represented.

// Writer provides functionality for writing a file set.
type Writer struct {
	ctx                context.Context
	tracker            track.Tracker
	storage            *Storage
	additive, deletive *index.Writer
	sizeBytes          int64
	cw                 *chunk.Writer
	idx                *index.Index
	deletePath         string
	lastIdx            *index.Index
	noUpload           bool
	indexFunc          func(*index.Index) error
	ttl                time.Duration
}

func newWriter(ctx context.Context, storage *Storage, tracker track.Tracker, chunks *chunk.Storage, opts ...WriterOption) *Writer {
	w := &Writer{
		ctx:     ctx,
		storage: storage,
		tracker: tracker,
	}
	for _, opt := range opts {
		opt(w)
	}
	var chunkWriterOpts []chunk.WriterOption
	if w.noUpload {
		chunkWriterOpts = append(chunkWriterOpts, chunk.WithNoUpload())
	}
	w.additive = index.NewWriter(ctx, chunks, "additive-index-writer")
	w.deletive = index.NewWriter(ctx, chunks, "deletive-index-writer")
	w.cw = chunks.NewWriter(ctx, "chunk-writer", w.callback, chunkWriterOpts...)
	return w
}

func (w *Writer) Add(path string, r io.Reader, tag ...string) error {
	idx := &index.Index{
		Path: path,
		File: &index.File{},
	}
	if len(tag) > 0 {
		idx.File.Tag = tag[0]
	}
	if err := w.nextIdx(idx); err != nil {
		return err
	}
	n, err := io.Copy(w.cw, r)
	w.sizeBytes += n
	return err
}

func (w *Writer) nextIdx(idx *index.Index) error {
	if w.idx != nil {
		if err := w.checkPath(w.idx.Path, idx.Path); err != nil {
			return err
		}
	}
	w.idx = idx
	return w.cw.Annotate(&chunk.Annotation{
		Data: idx,
	})
}

// Delete creates a delete operation for a file.
func (w *Writer) Delete(path string) error {
	if w.deletePath != "" {
		if err := w.checkPath(w.deletePath, path); err != nil {
			return err
		}
	}
	w.deletePath = path
	idx := &index.Index{
		Path: path,
		File: &index.File{},
	}
	return w.deletive.WriteIndex(idx)
}

func (w *Writer) checkPath(prev, p string) error {
	if prev == p {
		return errors.Errorf("cannot write same path (%s) twice", p)
	}
	if prev > p {
		return errors.Errorf("cannot write path (%s) after (%s)", p, prev)
	}
	return nil
}

// Copy copies a file to the file set writer.
func (w *Writer) Copy(file File, tag ...string) error {
	idx := file.Index()
	copyIdx := &index.Index{
		Path: idx.Path,
		File: &index.File{},
	}
	if len(tag) > 0 {
		copyIdx.File.Tag = tag[0]
	}
	if err := w.nextIdx(copyIdx); err != nil {
		return err
	}
	// Copy the file data refs if they are resolved.
	for _, dataRef := range idx.File.DataRefs {
		w.sizeBytes += dataRef.SizeBytes
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
func (w *Writer) Close() (*ID, error) {
	if err := w.cw.Close(); err != nil {
		return nil, err
	}
	// Write out the last index.
	if w.lastIdx != nil {
		idx := w.lastIdx
		if !w.noUpload {
			if err := w.additive.WriteIndex(idx); err != nil {
				return nil, err
			}
		}
		if w.indexFunc != nil {
			if err := w.indexFunc(idx); err != nil {
				return nil, err
			}
		}
	}
	if w.noUpload {
		return nil, nil
	}
	// Close the index writers.
	additiveIdx, err := w.additive.Close()
	if err != nil {
		return nil, err
	}
	deletiveIdx, err := w.deletive.Close()
	if err != nil {
		return nil, err
	}
	return w.storage.newPrimitive(w.ctx, &Primitive{
		Additive:  additiveIdx,
		Deletive:  deletiveIdx,
		SizeBytes: w.sizeBytes,
	}, w.ttl)
}
