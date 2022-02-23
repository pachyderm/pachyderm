package index

import (
	"context"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
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
	closed   chan struct{}
}

// NewWriter create a new Writer.
func NewWriter(ctx context.Context, chunks *chunk.Storage, tmpID string) *Writer {
	return &Writer{
		ctx:    ctx,
		chunks: chunks,
		tmpID:  tmpID,
		closed: make(chan struct{}, 1),
	}
}

// WriteIndex writes an index entry.
func (w *Writer) WriteIndex(idx *Index) error {
	w.setupLevels()
	return w.writeIndex(idx, 0)
}

func (w *Writer) setupLevels() {
	// Setup the first index level.
	if w.numLevels() == 0 {
		cw := w.chunks.NewWriter(w.ctx, w.tmpID, w.callback(0), chunk.WithRollingHashConfig(averageBits, 0))
		w.createLevel(&levelWriter{
			cw:  cw,
			pbw: pbutil.NewWriter(cw),
		})
	}
}

func (w *Writer) writeIndex(idx *Index, level int) error {
	idx = proto.Clone(idx).(*Index)
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
	return errors.EnsureStack(err)
}

func (w *Writer) callback(level int) chunk.WriterCallback {
	return func(annotations []*chunk.Annotation) error {
		// TODO: what's going on here?
		if len(annotations) == 0 {
			return nil
		}

		select {
		case <-w.closed:
			return nil
		default:
		}

		lw := w.getLevel(level)
		// Extract first and last index and setup file range.
		idx := proto.Clone(annotations[0].Data.(*data).idx).(*Index)
		dataRef := proto.Clone(annotations[0].NextDataRef).(*chunk.DataRef)
		// Edge case handling.
		if len(annotations) > 1 {
			// Skip the first index if it started in the previous chunk.
			if lw.lastIdx != nil && idx.Path == lw.lastIdx.Path {
				idx = proto.Clone(annotations[1].Data.(*data).idx).(*Index)
				dataRef = proto.Clone(annotations[1].NextDataRef).(*chunk.DataRef)
			}
		}
		if lw.firstIdx == nil {
			lw.firstIdx = proto.Clone(annotations[0].Data.(*data).idx).(*Index)
		}
		lw.lastIdx = proto.Clone(annotations[len(annotations)-1].Data.(*data).idx).(*Index)
		// Set standard fields in index.
		lastPath := lw.lastIdx.Path
		if lw.lastIdx.Range != nil {
			lastPath = lw.lastIdx.Range.LastPath
		}
		idx.Range = &Range{
			Offset:   dataRef.OffsetBytes,
			LastPath: lastPath,
			ChunkRef: chunk.Reference(dataRef),
		}
		idx.File = nil
		// Create next index level if it does not exist.
		if level == w.numLevels()-1 {
			cw := w.chunks.NewWriter(w.ctx, uuid.NewWithoutDashes(), w.callback(level+1), chunk.WithRollingHashConfig(averageBits, int64(level+1)))
			w.createLevel(&levelWriter{
				cw:  cw,
				pbw: pbutil.NewWriter(cw),
			})
		}
		// Write index entry in next index level.
		return w.writeIndex(idx, level+1)
	}
}

// Close finishes the index, and returns the serialized top index level.
func (w *Writer) Close() (ret *Index, retErr error) {
	// Note: new levels can be created while closing, so the number of iterations
	// necessary can increase as the levels are being closed. Levels stop getting
	// created when the top level chunk writer has been closed and the number of
	// annotations and chunks it has is one (one annotation in one chunk).
	for i := 0; i < w.numLevels(); i++ {
		l := w.getLevel(i)
		if proto.Equal(l.firstIdx, l.lastIdx) {
			w.closed <- struct{}{}
			return l.firstIdx, nil
		}
		if err := l.cw.Close(); err != nil {
			return nil, err
		}
	}
	return nil, errors.New("index writer failed to close.")
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
