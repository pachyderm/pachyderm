package index

import (
	"context"

	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

var (
	averageBits = 20
)

type levelWriter struct {
	cw      *chunk.Writer
	pbw     pbutil.Writer
	lastIdx *Index
}

type meta struct {
	idx   *Index
	level int
}

// Writer is used for creating a multilevel index into a serialized file set.
// Each index level is a stream of byte length encoded index entries that are stored in chunk storage.
type Writer struct {
	ctx    context.Context
	objC   obj.Client
	chunks *chunk.Storage
	path   string
	root   *Index
	levels []*levelWriter
	closed bool
}

// NewWriter create a new Writer.
func NewWriter(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string) *Writer {
	return &Writer{
		ctx:    ctx,
		objC:   objC,
		chunks: chunks,
		path:   path,
	}
}

func (w *Writer) WriteIndexes(idxs []*Index) error {
	w.setupLevels()
	return w.writeIndexes(idxs, 0)
}

func (w *Writer) setupLevels() {
	// Setup the first index level.
	if w.levels == nil {
		cw := w.chunks.NewWriter(w.ctx, averageBits, w.callback(0), 0)
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
			Meta: &meta{
				idx:   idx,
				level: level,
			},
		})
		if _, err := l.pbw.Write(idx); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) callback(level int) chunk.WriterFunc {
	return func(chunkRef *chunk.DataRef, annotations []*chunk.Annotation) error {
		if len(annotations) == 0 {
			return nil
		}
		lw := w.levels[level]
		// Extract first and last index and setup file range.
		idx := annotations[0].Meta.(*meta).idx
		offset := annotations[0].Offset
		// Edge case handling.
		if len(annotations) > 1 {
			// Skip the first index if it started in the previous chunk.
			if lw.lastIdx != nil && idx.Path == lw.lastIdx.Path {
				idx = annotations[1].Meta.(*meta).idx
				offset = annotations[1].Offset
			}
		}
		lw.lastIdx = annotations[len(annotations)-1].Meta.(*meta).idx
		// Set standard fields in index.
		lastPath := lw.lastIdx.Path
		if lw.lastIdx.Range != nil {
			lastPath = lw.lastIdx.Range.LastPath
		}
		idx.Range = &Range{
			Offset:   offset,
			LastPath: lastPath,
		}
		idx.DataOp = &DataOp{DataRefs: []*chunk.DataRef{chunkRef}}
		// Set the root index when the writer is closed and we are at the top index level.
		if w.closed {
			w.root = idx
		}
		// Create next index level if it does not exist.
		if level == len(w.levels)-1 {
			cw := w.chunks.NewWriter(w.ctx, averageBits, w.callback(level+1), int64(level+1))
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
func (w *Writer) Close() error {
	w.closed = true
	// Note: new levels can be created while closing, so the number of iterations
	// necessary can increase as the levels are being closed. Levels stop getting
	// created when the top level chunk writer has been closed and the number of
	// annotations and chunks it has is one (one annotation in one chunk).
	for i := 0; i < len(w.levels); i++ {
		l := w.levels[i]
		if err := l.cw.Close(); err != nil {
			return err
		}
		// (bryce) this method of terminating the index can create garbage (level
		// above the final level).
		if l.cw.AnnotationCount() == 1 && l.cw.ChunkCount() == 1 {
			break
		}
	}
	// Write the final index level to the path.
	objW, err := w.objC.Writer(w.ctx, w.path)
	if err != nil {
		return err
	}
	if _, err := pbutil.NewWriter(objW).Write(w.root); err != nil {
		return err
	}
	return objW.Close()
}
