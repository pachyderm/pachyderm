package index

import (
	"context"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
)

var (
	averageBits = 20
)

type levelWriter struct {
	muIn              sync.Mutex
	cw                *chunk.Writer
	pbw               pbutil.Writer
	firstIdx, lastIdx *Index

	muOut     sync.Mutex
	lastCbIdx *Index
}

// Writer is used for creating a multilevel index into a serialized file set.
// Each index level is a stream of byte length encoded index entries that are stored in chunk storage.
type Writer struct {
	ctx    context.Context
	cancel context.CancelFunc
	chunks *chunk.Storage
	tmpID  string

	// levelsMu guards levels.
	// All reads from levels should have the lock in at least read mode
	// All writes to levels (slice expansion) need levelsMu in write mode
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
	return w.withLevelIn(level, func(l *levelWriter) error {
		l.lastIdx = idx
		// pull the data refs from the index
		var refDataRefs []*chunk.DataRef
		if idx.Range != nil && idx.File != nil {
			return errors.New("either Index.Range or Index.File must be set, but not both.")
		} else if idx.Range != nil {
			refDataRefs = []*chunk.DataRef{idx.Range.ChunkRef}
		} else if idx.File != nil {
			refDataRefs = append(refDataRefs, idx.File.DataRefs...)
		} else {
			return errors.New("either Index.Range or Index.File must be set, but not both.")
		}
		// Create an annotation for each index.
		if err := l.cw.Annotate(&chunk.Annotation{
			RefDataRefs: refDataRefs,
			Data:        idx,
		}); err != nil {
			return err
		}
		_, err := l.pbw.Write(idx)
		return errors.EnsureStack(err)
	})
}

// setupLevels ensures that len(w.levels) > 0 before releasing it
// setupLevels takes levelsMu in read and write mode as needed, and releases it before returning
func (w *Writer) setupLevel(idx *Index, i int) {
	// first check with the read lock.
	length := w.numLevels()
	if i < length {
		// this should be the common case
		return
	}
	// then get the write lock, check again and maybe create another level.
	w.levelsMu.Lock()
	defer w.levelsMu.Unlock()
	if i >= len(w.levels) {
		cw := w.chunks.NewWriter(w.ctx, w.tmpID, w.callback(i), chunk.WithRollingHashConfig(averageBits, int64(i)))
		lw := &levelWriter{
			cw:       cw,
			pbw:      pbutil.NewWriter(cw),
			firstIdx: idx,
		}
		w.levels = append(w.levels, lw)
	}
}

func (w *Writer) callback(level int) chunk.WriterCallback {
	return func(annotations []*chunk.Annotation) error {
		// TODO: What's going on here? Why would a callback be called with 0 annotations.
		//       Nothing immediately breaks when this is left out.
		if len(annotations) == 0 {
			return nil
		}
		return w.withLevelOut(level, func(lw *levelWriter) error {
			// Extract first and last index and setup file range.
			idx := annotations[0].Data.(*Index)
			dataRef := annotations[0].NextDataRef
			// Edge case handling.
			if len(annotations) > 1 {
				// Skip the first index if it started in the previous chunk.
				if proto.Equal(lw.lastCbIdx, idx) {
					idx = annotations[1].Data.(*Index)
					dataRef = annotations[1].NextDataRef
				}
			}
			idx = proto.Clone(idx).(*Index)
			dataRef = proto.Clone(dataRef).(*chunk.DataRef)
			idx.File = nil
			lw.lastCbIdx = proto.Clone(annotations[len(annotations)-1].Data.(*Index)).(*Index)
			lastPath := lw.lastCbIdx.Path
			if lw.lastCbIdx.Range != nil {
				lastPath = lw.lastCbIdx.Range.LastPath
			}
			idx.Range = &Range{
				Offset:   dataRef.OffsetBytes,
				LastPath: lastPath,
				ChunkRef: chunk.Reference(dataRef),
			}
			// Write index entry in next index level.
			return w.writeIndex(idx, level+1)
		})
	}
}

// Close finishes the index, and returns the serialized top index level.
func (w *Writer) Close() (ret *Index, retErr error) {
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
		var retIdx *Index
		if err := w.withLevelIn(i, func(l *levelWriter) error {
			if proto.Equal(l.firstIdx, l.lastIdx) {
				retIdx = l.firstIdx
				return nil
			}
			return l.cw.Close()
		}); err != nil {
			return nil, err
		}
		if retIdx != nil {
			return retIdx, nil
		}
	}
	return nil, nil
}

// numLevels get the number of levels while holding levelsMu and returns it.
func (w *Writer) numLevels() int {
	w.levelsMu.RLock()
	defer w.levelsMu.RUnlock()
	return len(w.levels)
}

// withLevelIn calls fn with a levelWriter, and the input lock held for
// that level writer.
// w.levelsMu is not held during this call
func (w *Writer) withLevelIn(i int, fn func(lw *levelWriter) error) error {
	w.levelsMu.RLock()
	// TODO (brendon): I don't like the possiblity panicing here without defering the unlock
	lw := w.levels[i]
	w.levelsMu.RUnlock()
	// It's very important that we don't have w.levelsMu held here, otherwise this could deadlock.
	lw.muIn.Lock()
	defer lw.muIn.Unlock()
	return fn(lw)
}

// withLevelIn calls fn with a levelWriter, and the output lock held for
// that level writer.
// w.levelsMu is not held during this call
func (w *Writer) withLevelOut(i int, fn func(lw *levelWriter) error) error {
	w.levelsMu.RLock()
	// TODO (brendon): I don't like the possiblity panicing here without defering the unlock
	lw := w.levels[i]
	w.levelsMu.RUnlock()
	// It's very important that we don't have w.levelsMu held here, otherwise this could deadlock.
	lw.muOut.Lock()
	defer lw.muOut.Unlock()
	return fn(lw)
}
