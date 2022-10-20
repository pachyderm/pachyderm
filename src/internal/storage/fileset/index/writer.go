package index

import (
	"bytes"
	"context"
	"sync"

	"github.com/docker/go-units"
	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
)

const (
	DefaultBatchThreshold = units.MB
)

type levelWriter struct {
	buf               *bytes.Buffer
	batcher           *chunk.Batcher
	firstIdx, lastIdx *Index
}

// Writer is used for creating a multilevel index into a serialized file set.
// Each index level is a stream of byte length encoded index entries that are
// stored in chunk storage. Both file and range type indexes can be written to
// a writer. New levels above the written indexes will be created when the
// serialized indexes reach the batching threshold.
type Writer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	chunks   *chunk.Storage
	tmpID    string
	levelsMu sync.RWMutex
	levels   []*levelWriter
}

// NewWriter create a new Writer.
func NewWriter(ctx context.Context, chunks *chunk.Storage, tmpID string) *Writer {
	ctx, cancel := context.WithCancel(ctx)
	return &Writer{
		ctx:    ctx,
		cancel: cancel,
		chunks: chunks,
		tmpID:  tmpID,
	}
}

// WriteIndex writes an index entry.
func (w *Writer) WriteIndex(idx *Index) error {
	idx = proto.Clone(idx).(*Index)
	return w.writeIndex(idx, 0)
}

func (w *Writer) writeIndex(idx *Index, level int) error {
	w.setupLevel(idx, level)
	l := w.getLevel(level)
	l.lastIdx = idx
	// pull the data refs from the index
	var pointsTo []*chunk.DataRef
	if idx.Range != nil && idx.File != nil {
		return errors.New("either Index.Range or Index.File must be set, but not both.")
	} else if idx.Range != nil {
		pointsTo = []*chunk.DataRef{idx.Range.ChunkRef}
	} else if idx.File != nil {
		pointsTo = append(pointsTo, idx.File.DataRefs...)
	} else {
		return errors.New("either Index.Range or Index.File must be set, but not both.")
	}
	l.buf.Reset()
	pbw := pbutil.NewWriter(l.buf)
	_, err := pbw.Write(idx)
	if err != nil {
		return errors.EnsureStack(err)
	}
	return l.batcher.Add(idx, l.buf.Bytes(), pointsTo)
}

func (w *Writer) setupLevel(idx *Index, level int) {
	if level == w.numLevels() {
		batcher := w.chunks.NewBatcher(w.ctx, w.tmpID, DefaultBatchThreshold, chunk.WithChunkCallback(w.callback(level)))
		w.createLevel(&levelWriter{
			buf:      &bytes.Buffer{},
			batcher:  batcher,
			firstIdx: idx,
		})
	}
}

func (w *Writer) callback(level int) chunk.ChunkFunc {
	return func(metas []interface{}, chunkRef *chunk.DataRef) error {
		idx := proto.Clone(metas[0].(*Index)).(*Index)
		lastIdx := metas[len(metas)-1].(*Index)
		lastPath := lastIdx.Path
		if lastIdx.Range != nil {
			lastPath = lastIdx.Range.LastPath
		}
		idx.Range = &Range{
			LastPath: lastPath,
			ChunkRef: chunkRef,
		}
		idx.File = nil
		idx.NumFiles = 0
		for _, meta := range metas {
			idx.NumFiles += meta.(*Index).NumFiles
		}
		idx.SizeBytes = 0
		for _, meta := range metas {
			idx.SizeBytes += meta.(*Index).SizeBytes
		}
		// Write index entry in next index level.
		return w.writeIndex(idx, level+1)
	}
}

// Close finishes the index, and returns the serialized top index level.
func (w *Writer) Close() (*Index, error) {
	// Close levels until a level with only one index entry (first index == last index)
	// exists and return the index.
	// Note: new levels can be created while closing, so the number of iterations
	// necessary can increase as the levels are being closed. Cancelling the context
	// will stop any open chunk writers, ending the callback recursion.
	// Any additional chunks that are created at or above the level with the one index
	// entry (chunk split point in the middle) will be unreferenced and therefore cleaned
	// up by garbage collection.
	defer w.cancel()
	for i := 0; i < w.numLevels(); i++ {
		l := w.getLevel(i)
		if proto.Equal(l.firstIdx, l.lastIdx) {
			return l.firstIdx, nil
		}
		if err := l.batcher.Close(); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (w *Writer) createLevel(l *levelWriter) {
	w.levelsMu.Lock()
	defer w.levelsMu.Unlock()
	w.levels = append(w.levels, l)
}

func (w *Writer) getLevel(level int) *levelWriter {
	w.levelsMu.RLock()
	defer w.levelsMu.RUnlock()
	return w.levels[level]
}

func (w *Writer) numLevels() int {
	w.levelsMu.RLock()
	defer w.levelsMu.RUnlock()
	return len(w.levels)
}
