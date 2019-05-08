package index

import (
	"archive/tar"
	"bytes"
	"context"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

const (
	indexType        = 'i'
	rangeType        = 'r'
	defaultRangeSize = int64(chunk.MB)
)

// Header is a wrapper for a tar header and index.
type Header struct {
	Hdr *tar.Header
	Idx *Index
}

type levelWriter struct {
	cw *chunk.Writer
	tw *tar.Writer
}

// Writer is used for creating a multi-level index into a serialized FileSet.
// Each index level consists of compressed tar stream chunks.
// Each index tar entry has the full index in the content section.
type Writer struct {
	ctx       context.Context
	chunks    *chunk.Storage
	root      *Header
	levels    []*levelWriter
	rangeSize int64
	closed    bool
	lastPath  string
}

// NewWriter create a new Writer.
// rangeSize should not be used except for testing purposes, the defaultRangeSize will
// be used in a real deployment.
func NewWriter(ctx context.Context, chunks *chunk.Storage, rangeSize ...int64) *Writer {
	rSize := defaultRangeSize
	if len(rangeSize) > 0 {
		rSize = rangeSize[0]
	}
	return &Writer{
		ctx:       ctx,
		chunks:    chunks,
		rangeSize: rSize,
	}
}

// WriteHeader writes a Header to the index.
func (w *Writer) WriteHeader(hdr *Header) error {
	// Sets up the root header and first level.
	if w.root == nil {
		w.root = hdr
		cw := w.chunks.NewWriter(w.ctx)
		cw.StartRange(w.callback(hdr, 0))
		w.levels = append(w.levels, &levelWriter{
			cw: cw,
			tw: tar.NewWriter(cw),
		})
	}
	w.lastPath = hdr.Hdr.Name
	if hdr.Idx == nil {
		hdr.Idx = &Index{}
	}
	hdr.Hdr.Typeflag = indexType
	return w.writeHeader(hdr, 0)
}

func (w *Writer) writeHeader(hdr *Header, level int) error {
	l := w.levels[level]
	// Start new range if past range size, and propagate first header up index levels.
	if l.cw.RangeSize() > w.rangeSize {
		l.cw.StartRange(w.callback(hdr, level))
	}
	return w.serialize(l.tw, hdr, level)
}

func (w *Writer) serialize(tw *tar.Writer, hdr *Header, level int) error {
	// Create file range if above lowest index level.
	if level > 0 {
		hdr.Idx.Range = &Range{}
		hdr.Idx.Range.LastPath = w.lastPath
	}
	// Serialize and write additional metadata.
	idx, err := proto.Marshal(hdr.Idx)
	if err != nil {
		return err
	}
	hdr.Hdr.Size = int64(len(idx))
	if err := tw.WriteHeader(hdr.Hdr); err != nil {
		return err
	}
	if _, err = tw.Write(idx); err != nil {
		return err
	}
	return tw.Flush()
}

func (w *Writer) callback(hdr *Header, level int) func([]*chunk.DataRef) error {
	return func(dataRefs []*chunk.DataRef) error {
		hdr.Hdr.Typeflag = rangeType
		// Used to communicate data refs for final index level to Close function.
		if w.closed && w.levels[level].cw.RangeCount() == 1 {
			w.root.Idx.DataOp = &DataOp{DataRefs: dataRefs}
			return nil
		}
		// Create next index level if it does not exist.
		if level == len(w.levels)-1 {
			cw := w.chunks.NewWriter(w.ctx)
			cw.StartRange(w.callback(hdr, level+1))
			w.levels = append(w.levels, &levelWriter{
				cw: cw,
				tw: tar.NewWriter(cw),
			})
		}
		// Write index entry in next level index.
		hdr.Idx.DataOp = &DataOp{DataRefs: dataRefs}
		return w.writeHeader(hdr, level+1)
	}
}

// Close finishes the index, and returns the serialized top level index.
func (w *Writer) Close() (r io.Reader, retErr error) {
	w.closed = true
	// Note: new levels can be created while closing, so the number of iterations
	// necessary can increase as the levels are being closed. The number of ranges
	// will decrease per level as long as the range size is in general larger than
	// a serialized header. Levels stop getting created when the top level chunk
	// writer has been closed and the number of ranges it has is one.
	for i := 0; i < len(w.levels); i++ {
		l := w.levels[i]
		if err := l.tw.Close(); err != nil {
			return nil, err
		}
		if err := l.cw.Close(); err != nil {
			return nil, err
		}
	}
	// Write the final index level that will be readable
	// by the caller.
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)
	defer func() {
		if err := tw.Close(); err != nil && retErr != nil {
			retErr = err
		}
	}()
	if err := w.serialize(tw, w.root, len(w.levels)); err != nil {
		return nil, err
	}
	return buf, nil
}
