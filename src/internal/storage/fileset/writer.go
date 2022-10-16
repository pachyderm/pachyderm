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

// TODO: Might need to think a bit more about file set sizes and whether deletes should be represented.
// Also, we should consider removing the file set size field and leaning on the new index size field.

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
		idx.NumFiles = 1
		size := index.SizeBytes(idx)
		idx.SizeBytes = size
		atomic.AddInt64(&w.sizeBytes, size)
		return w.additive.WriteIndex(idx)
	})
	w.additiveBatched = index.NewWriter(ctx, storage.ChunkStorage(), "additive-batched-index-writer")
	w.batcher = storage.ChunkStorage().NewBatcher(ctx, "chunk-batcher", w.batchThreshold, chunk.WithEntryCallback(func(meta interface{}, dataRef *chunk.DataRef) error {
		idx := meta.(*index.Index)
		if dataRef != nil {
			idx.File.DataRefs = []*chunk.DataRef{dataRef}
		}
		idx.NumFiles = 1
		size := index.SizeBytes(idx)
		idx.SizeBytes = size
		atomic.AddInt64(&w.sizeBytes, size)
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
	// TODO: Readd if we should block this outright.
	// We support this in the storage layer to allow the reading of invalid commits, but we may want to
	// block it outright in the future. This check can no longer be used because compaction will write file sets
	// like this in certain cases since validation happens after compaction.
	//if prevIdx.Path == idx.Path && prevIdx.File.Datum == idx.File.Datum {
	//	return errors.Errorf("cannot write same path (%s) and datum (%s) twice", idx.Path, idx.File.Datum)
	//}
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
		NumFiles: 1,
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
	if len(idx.File.DataRefs) == 0 {
		return w.Add(idx.Path, datum, &bytes.Buffer{})
	}
	if len(idx.File.DataRefs) == 1 {
		r := w.storage.ChunkStorage().NewDataReader(w.ctx, idx.File.DataRefs[0])
		return w.Add(idx.Path, datum, r)
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
	var additiveMergeIdx *index.Index
	switch {
	case additiveIdx == nil:
		additiveMergeIdx = additiveBatchedIdx
	case additiveBatchedIdx == nil:
		additiveMergeIdx = additiveIdx
	default:
		// TODO: Rethink / rework merging the two additive indexes?
		additiveMerge := index.NewWriter(w.ctx, w.storage.ChunkStorage(), "additive-merge-index-writer")
		if err := index.Merge(w.ctx, w.storage.ChunkStorage(), []*index.Index{additiveIdx, additiveBatchedIdx}, func(idx *index.Index) error {
			return additiveMerge.WriteIndex(idx)
		}); err != nil {
			return nil, err
		}
		additiveMergeIdx, err = additiveMerge.Close()
		if err != nil {
			return nil, err
		}
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
