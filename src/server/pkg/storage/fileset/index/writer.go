package index

import (
	"context"
	"sync"

	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

var (
	averageBits = 20
)

// TODO might want to move this into the chunk storage layer as a default tag.
const (
	indexTag = ""
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

// WriteIndexes writes a set of index entries.
func (w *Writer) WriteIndexes(idxs []*Index) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.setupLevels()
	return w.writeIndexes(idxs, 0)
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

func (w *Writer) writeIndexes(idxs []*Index, level int) error {
	l := w.levels[level]
	for _, idx := range idxs {
		// Create an annotation for each index.
		l.cw.Annotate(&chunk.Annotation{
			RefDataRefs: idx.DataOp.DataRefs,
			Data: &data{
				idx:   idx,
				level: level,
			},
		})
		l.cw.Tag(indexTag)
		if _, err := l.pbw.Write(idx); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) callback(level int) chunk.WriterFunc {
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
		}
		idx.DataOp = &DataOp{DataRefs: []*chunk.DataRef{chunk.Reference(dataRef, indexTag)}}
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
		return w.writeIndexes([]*Index{idx}, level+1)
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
