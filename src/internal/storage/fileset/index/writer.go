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
	cw                *chunk.Writer
	pbw               pbutil.Writer
	firstIdx, lastIdx *Index
}

type data struct {
	idx   *Index
	level int
}

// Writer is used for creating a multilevel index into a serialized file set.
// Each index level is a stream of byte length encoded index entries that are stored in chunk storage.
type Writer struct {
	ctx    context.Context
	chunks *chunk.Storage
	tmpID  string

	levelsMu sync.RWMutex
	levels   []*levelWriter
	done     chan struct{}
}

// NewWriter create a new Writer.
func NewWriter(ctx context.Context, chunks *chunk.Storage, tmpID string) *Writer {
	return &Writer{
		ctx:    ctx,
		chunks: chunks,
		tmpID:  tmpID,
		done:   make(chan struct{}, 1),
	}
}

// WriteIndex writes an index entry.
func (w *Writer) WriteIndex(idx *Index) error {
	return w.writeIndex(idx, 0)
}

func (w *Writer) writeIndex(idx *Index, level int) error {
	idx = proto.Clone(idx).(*Index)
	w.setupLevel(idx, level)
	l := w.getLevel(level)
	var refDataRefs []*chunk.DataRef
	if idx.Range != nil {
		refDataRefs = []*chunk.DataRef{idx.Range.ChunkRef}
	}
	// TODO: I think we need to clear this field since it is reused.
	// Maybe we could restructure this into a clone, with the field cleared.
	if idx.File != nil {
		refDataRefs = append(refDataRefs, idx.File.DataRefs...)
	}
	// Create an annotation for each index.
	if err := l.cw.Annotate(&chunk.Annotation{
		RefDataRefs: refDataRefs,
		Data: &data{
			idx:   idx,
			level: level,
		},
	}); err != nil {
		return err
	}
	_, err := l.pbw.Write(idx)
	l.lastIdx = idx
	return errors.EnsureStack(err)
}

func (w *Writer) setupLevel(idx *Index, level int) {
	if level == w.numLevels() {
		cw := w.chunks.NewWriter(w.ctx, w.tmpID, w.callback(level), chunk.WithRollingHashConfig(averageBits, int64(level)))
		lw := &levelWriter{
			cw:  cw,
			pbw: pbutil.NewWriter(cw),
		}
		w.createLevel(lw)
		if lw.firstIdx == nil {
			lw.firstIdx = proto.Clone(idx).(*Index)
		}
	}
}

func (w *Writer) callback(level int) chunk.WriterCallback {
	return func(annotations []*chunk.Annotation) error {
		// TODO: What's going on here? Why would a callback be called with 0 annotations.
		//       Nothing immediately breaks when this is left out.
		if len(annotations) == 0 {
			return nil
		}
		select {
		case <-w.done:
			return nil
		default:
		}

		lw := w.getLevel(level)
		// Extract first and last index and setup file range.
		idx := proto.Clone(annotations[0].Data.(*data).idx).(*Index)
		idx.File = nil
		dataRef := proto.Clone(annotations[0].NextDataRef).(*chunk.DataRef)
		// Edge case handling.
		if len(annotations) > 1 {
			// Skip the first index if it started in the previous chunk.
			if lw.lastIdx != nil && idx.Path == lw.lastIdx.Path {
				idx = proto.Clone(annotations[1].Data.(*data).idx).(*Index)
				dataRef = proto.Clone(annotations[1].NextDataRef).(*chunk.DataRef)
			}
		}

		lastIdx := proto.Clone(annotations[len(annotations)-1].Data.(*data).idx).(*Index)
		lastPath := lastIdx.Path
		if lastIdx.Range != nil {
			lastPath = lastIdx.Range.LastPath
		}
		// Set standard fields in index.
		idx.Range = &Range{
			Offset:   dataRef.OffsetBytes,
			LastPath: lastPath,
			ChunkRef: chunk.Reference(dataRef),
		}
		// Write index entry in next index level.
		return w.writeIndex(idx, level+1)
	}
}

// Close finishes the index, and returns the serialized top index level.
func (w *Writer) Close() (ret *Index, retErr error) {
	// Note: new levels can be created while closing, so the number of iterations
	// necessary can increase as the levels are being closed. Levels stop getting
	// created when the top level chunk writer has been closed and the level chunk
	// writer's first and last index are the same.
	// Duplicate indeces may occur if a split point is found in a written index.
	for i := 0; i < w.numLevels(); i++ {
		l := w.getLevel(i)
		if l.firstIdx.Path == l.lastIdx.Path {
			w.done <- struct{}{}
			return l.firstIdx, nil
		}
		if err := l.cw.Close(); err != nil {
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
