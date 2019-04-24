package index

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

const (
	MetaKey          = "fileinfo"
	Base             = "base"
	defaultRangeSize = int64(chunk.MB)
)

// Header is a wrapper for a tar header with additional metadata in the form of a FileInfo.
type Header struct {
	hdr  *tar.Header
	info *FileInfo
}

type levelWriter struct {
	cw *chunk.Writer
	tw *tar.Writer
}

// Writer is used for creating a multi-level index into a serialized fileset.
// Each index level consists of compressed tar stream chunks (specifically just the tar headers for the underlying index/content tar stream).
// Each tar header in an index has additional metadata stored in a base64 encoding of a serialized FileInfo in a PAXRecords under the key defined by MetaKey.
type Writer struct {
	ctx       context.Context
	chunks    *chunk.Storage
	root      *Header
	levels    []*levelWriter
	rangeSize int64
	closed    bool
}

// NewWriter create a new Writer.
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
		cw.RangeStart(w.callback(hdr, 0))
		w.levels = append(w.levels, &levelWriter{
			cw: cw,
			tw: tar.NewWriter(cw),
		})
	}
	if hdr.hdr.PAXRecords == nil {
		hdr.hdr.PAXRecords = make(map[string]string)
	}
	hdr.hdr.PAXRecords[Base] = ""
	return w.writeHeader(hdr, 0)
}

func (w *Writer) writeHeader(hdr *Header, level int) error {
	l := w.levels[level]
	// Start new range if past range size, and propagate header up index levels.
	if l.cw.RangeSize() > w.rangeSize {
		l.cw.RangeStart(w.callback(hdr, level))
	}
	// Serialize, write, then clear the keys that correspond to this write.
	if err := serialize(hdr); err != nil {
		return err
	}
	if err := l.tw.WriteHeader(hdr.hdr); err != nil {
		return err
	}
	delete(hdr.hdr.PAXRecords, Base)
	delete(hdr.hdr.PAXRecords, MetaKey)
	return nil
}

func serialize(hdr *Header) error {
	bytes, err := proto.Marshal(hdr.info)
	if err != nil {
		return err
	}
	hdr.hdr.PAXRecords[MetaKey] = base64.StdEncoding.EncodeToString(bytes)
	return nil
}

func (w *Writer) callback(hdr *Header, level int) func([]*chunk.DataRef) error {
	return func(dataRefs []*chunk.DataRef) error {
		// Used to communicate data refs for final index level to Close function.
		if w.closed && w.levels[level].cw.RangeCount() == 1 {
			w.root.info.DataOps = refsToOps(dataRefs)
			return nil
		}
		// Create next index level if it does not exist.
		if level == len(w.levels)-1 {
			cw := w.chunks.NewWriter(w.ctx)
			cw.RangeStart(w.callback(hdr, level+1))
			w.levels = append(w.levels, &levelWriter{
				cw: cw,
				tw: tar.NewWriter(cw),
			})
		}
		// Write index entry in next level index.
		hdr.info.DataOps = refsToOps(dataRefs)
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
		level := w.levels[i]
		if err := level.tw.Close(); err != nil {
			return nil, err
		}
		if err := level.cw.Close(); err != nil {
			return nil, err
		}
	}
	if err := serialize(w.root); err != nil {
		return nil, err
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
	if err := tw.WriteHeader(w.root.hdr); err != nil {
		return nil, err
	}
	return buf, nil
}
