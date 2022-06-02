package fileset

import (
	"bytes"
	"context"
	"io"
	"sync/atomic"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

// TODO: Size zero files need to be addressed now that we are moving away from storing tar headers.
// We can run into the same issue as deletions where a lot of size zero files can cause us to get backed up
// since no chunks will get created. The solution we have in mind is to write a small number of bytes
// for a size zero file, then either not store references to them or ignore them at read time.

// TODO: Might need to think a bit more about fileset sizes and whether deletes should be represented.

// Writer provides functionality for writing a file set.
type Writer struct {
	ctx                                 context.Context
	storage                             *Storage
	additive, additiveBatched, deletive *index.Writer
	uploader                            *chunk.Uploader
	batchThreshold                      int
	batcher                             *chunk.Batcher
	idx                                 *index.Index
	deleteIdx                           *index.Index
	ttl                                 time.Duration
	sizeBytes                           int64
}

func newWriter(ctx context.Context, storage *Storage, opts ...WriterOption) *Writer {
	w := &Writer{
		ctx:            ctx,
		storage:        storage,
		batchThreshold: DefaultBatchThreshold,
	}
	for _, opt := range opts {
		opt(w)
	}
	w.additive = index.NewWriter(ctx, storage.ChunkStorage(), "additive-index-writer")
	w.uploader = storage.ChunkStorage().NewUploader(ctx, "chunk-uploader", false, func(meta interface{}, dataRefs []*chunk.DataRef) error {
		idx := meta.(*index.Index)
		idx.File.DataRefs = dataRefs
		atomic.AddInt64(&w.sizeBytes, index.SizeBytes(idx))
		return w.additive.WriteIndex(idx)
	})
	w.additiveBatched = index.NewWriter(ctx, storage.ChunkStorage(), "additive-batched-index-writer")
	w.batcher = storage.ChunkStorage().NewBatcher(ctx, "chunk-batcher", w.batchThreshold, chunk.WithEntryCallback(func(meta interface{}, dataRef *chunk.DataRef) error {
		idx := meta.(*index.Index)
		if dataRef != nil {
			idx.File.DataRefs = []*chunk.DataRef{dataRef}
		}
		atomic.AddInt64(&w.sizeBytes, index.SizeBytes(idx))
		return w.additiveBatched.WriteIndex(idx)
	}))
	w.deletive = index.NewWriter(ctx, storage.ChunkStorage(), "deletive-index-writer")
	return w
}

func (w *Writer) Add(path, datum string, r io.Reader) error {
	idx := &index.Index{
		Path: path,
		File: &index.File{
			Datum: datum,
		},
	}
	if err := w.checkIndex(w.idx, idx); err != nil {
		return err
	}
	w.idx = idx
	// Handle files less than the batch threshold.
	buf := &bytes.Buffer{}
	_, err := io.CopyN(buf, r, int64(w.batchThreshold))
	if err != nil {
		if errors.Is(err, io.EOF) {
			return w.batcher.Add(idx, buf.Bytes(), nil)
		}
		return errors.EnsureStack(err)
	}
	// Handle files greater than or equal to the batch threshold.
	r = io.MultiReader(buf, r)
	return w.uploader.Upload(idx, r)
}

func (w *Writer) checkIndex(prevIdx, idx *index.Index) error {
	if prevIdx == nil {
		return nil
	}
	if prevIdx.Path == idx.Path && prevIdx.File.Datum == idx.File.Datum {
		return errors.Errorf("cannot write same path (%s) and datum (%s) twice", idx.Path, idx.File.Datum)
	}
	if prevIdx.Path == idx.Path && prevIdx.File.Datum > idx.File.Datum {
		return errors.Errorf("cannot write same path (%s) for datum (%s) after datum (%s)", idx.Path, idx.File.Datum, prevIdx.File.Datum)
	}
	if prevIdx.Path > idx.Path {
		return errors.Errorf("cannot write path (%s) after (%s)", idx.Path, prevIdx.Path)
	}
	return nil
}

// Delete creates a delete operation for a file.
func (w *Writer) Delete(path, datum string) error {
	idx := &index.Index{
		Path: path,
		File: &index.File{
			Datum: datum,
		},
	}
	if err := w.checkIndex(w.deleteIdx, idx); err != nil {
		return err
	}
	w.deleteIdx = idx
	return w.deletive.WriteIndex(idx)
}

// Copy copies a file to the file set writer.
func (w *Writer) Copy(file File, datum string) error {
	idx := file.Index()
	size := index.SizeBytes(idx)
	if size >= int64(w.batchThreshold) && chunk.StableDataRefs(idx.File.DataRefs) {
		if err := w.checkIndex(w.idx, idx); err != nil {
			return err
		}
		w.idx = idx
		copyIdx := &index.Index{
			Path: idx.Path,
			File: &index.File{
				Datum: datum,
			},
		}
		return w.uploader.Copy(copyIdx, idx.File.DataRefs)
	}
	// TODO: Optimize to handle large files that can mostly be copy by reference?
	return miscutil.WithPipe(func(w2 io.Writer) error {
		r := w.storage.ChunkStorage().NewReader(w.ctx, idx.File.DataRefs)
		return r.Get(w2)
	}, func(r io.Reader) error {
		return w.Add(idx.Path, datum, r)
	})
}

// Close closes the writer.
func (w *Writer) Close() (*ID, error) {
	// The uploader writes to the additive index, so close it first.
	if err := w.uploader.Close(); err != nil {
		return nil, err
	}
	additiveIdx, err := w.additive.Close()
	if err != nil {
		return nil, err
	}
	// The batcher writes to the additive batched index, so close it first.
	if err := w.batcher.Close(); err != nil {
		return nil, err
	}
	additiveBatchedIdx, err := w.additiveBatched.Close()
	if err != nil {
		return nil, err
	}
	// TODO: Rethink / rework merging the two additive indexes?
	additiveMerge := index.NewWriter(w.ctx, w.storage.ChunkStorage(), "additive-merge-index-writer")
	if err := index.Merge(w.ctx, w.storage.ChunkStorage(), []*index.Index{additiveIdx, additiveBatchedIdx}, func(idx *index.Index) error {
		return additiveMerge.WriteIndex(idx)
	}); err != nil {
		return nil, err
	}
	additiveMergeIdx, err := additiveMerge.Close()
	if err != nil {
		return nil, err
	}
	deletiveIdx, err := w.deletive.Close()
	if err != nil {
		return nil, err
	}
	return w.storage.newPrimitive(w.ctx, &Primitive{
		Additive:  additiveMergeIdx,
		Deletive:  deletiveIdx,
		SizeBytes: w.sizeBytes,
	}, w.ttl)
}
