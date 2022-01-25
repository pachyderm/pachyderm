package index

import (
	"context"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

var (
	averageBits = 20
)

type levelWriter struct {
	cw      *chunk.Writer
	pbw     pbutil.Writer
	lastIdx *Index
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

	mu     sync.Mutex
	levels []*levelWriter
	closed bool
	root   *Index
}

// NewWriter create a new Writer.
func NewWriter(ctx context.Context, chunks *chunk.Storage, tmpID string) *Writer {
	return &Writer{
		ctx:    ctx,
		chunks: chunks,
		tmpID:  tmpID,
	}
}

// WriteIndex writes an index entry.
func (w *Writer) WriteIndex(idx *Index) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.setupLevels()
	return w.writeIndex(idx, 0)
}

func (w *Writer) setupLevels() {
	// Setup the first index level.
	if w.levels == nil {
		cw := w.chunks.NewWriter(w.ctx, w.tmpID, w.callback(0), chunk.WithRollingHashConfig(averageBits, 0))
		w.levels = append(w.levels, &levelWriter{
			cw:  cw,
			pbw: pbutil.NewWriter(cw),
		})
	}
}

func (w *Writer) writeIndex(idx *Index, level int) error {
	l := w.levels[level]
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
		w.mu.Lock()
		defer w.mu.Unlock()
		if len(annotations) == 0 {
			return nil
		}
		lw := w.levels[level]
		// Extract first and last index and setup file range.
		idx := annotations[0].Data.(*data).idx
		dataRef := annotations[0].NextDataRef
		// Edge case handling.
		if len(annotations) > 1 {
			// Skip the first index if it started in the previous chunk.
			if lw.lastIdx != nil && idx.Path == lw.lastIdx.Path {
				idx = annotations[1].Data.(*data).idx
				dataRef = annotations[1].NextDataRef
			}
		}
		lw.lastIdx = annotations[len(annotations)-1].Data.(*data).idx
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
		// Set the root index when the writer is closed and we are at the top index level.
		if w.closed {
			w.root = idx
		}
		// Create next index level if it does not exist.
		if level == len(w.levels)-1 {
			cw := w.chunks.NewWriter(w.ctx, uuid.NewWithoutDashes(), w.callback(level+1), chunk.WithRollingHashConfig(averageBits, int64(level+1)))
			w.levels = append(w.levels, &levelWriter{
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
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()

	// Note: new levels can be created while closing, so the number of iterations
	// necessary can increase as the levels are being closed. Levels stop getting
	// created when the top level chunk writer has been closed and the number of
	// annotations and chunks it has is one (one annotation in one chunk).
	for i := 0; i < len(w.levels); i++ {
		w.mu.Lock()
		l := w.levels[i]
		w.mu.Unlock()
		if err := l.cw.Close(); err != nil {
			return nil, err
		}
		if l.cw.AnnotationCount() == 1 && l.cw.ChunkCount() == 1 {
			break
		}
	}
	return w.root, nil
}
