package index

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

const (
	indexType = 'i'
	rangeType = 'r'
)

var (
	averageBits = 20
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

type meta struct {
	hdr   *Header
	level int
}

// Writer is used for creating a multi-level index into a serialized FileSet.
// Each index level consists of compressed tar stream chunks.
// Each index tar entry has the full index in the content section.
// (bryce) need to figure out how to handle small chunk without a start point for a header.
type Writer struct {
	ctx    context.Context
	objC   obj.Client
	chunks *chunk.Storage
	path   string
	root   *Header
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

// WriteHeader writes a Header to the index.
func (w *Writer) WriteHeader(hdr *Header) error {
	// Setup the first level.
	if w.levels == nil {
		cw := w.chunks.NewWriter(w.ctx, averageBits, w.callback(0))
		w.levels = append(w.levels, &levelWriter{
			cw: cw,
			tw: tar.NewWriter(cw),
		})
	}
	hdr.Hdr.Typeflag = indexType
	return w.writeHeader(hdr, 0)
}

func (w *Writer) writeHeader(hdr *Header, level int) error {
	l := w.levels[level]
	l.cw.Annotate(&chunk.Annotation{
		RefDataRefs: hdr.Idx.DataOp.DataRefs,
		Meta: &meta{
			hdr:   hdr,
			level: level,
		},
	})
	return w.serialize(l.tw, hdr)
}

func (w *Writer) serialize(tw *tar.Writer, hdr *Header) error {
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

func (w *Writer) callback(level int) chunk.WriterFunc {
	return func(dataRefs []*chunk.DataRef, annotations []*chunk.Annotation) error {
		// Extract first and last header and setup file range.
		hdr := annotations[0].Meta.(*meta).hdr
		lastHdr := annotations[len(annotations)-1].Meta.(*meta).hdr
		var lastPath string
		if lastHdr.Hdr.Typeflag == indexType {
			lastPath = lastHdr.Hdr.Name
		} else {
			lastPath = lastHdr.Idx.Range.LastPath
		}
		hdr.Hdr.Typeflag = rangeType
		hdr.Idx.Range = &Range{
			Offset:   annotations[0].Offset,
			LastPath: lastPath,
		}
		hdr.Idx.DataOp = &DataOp{DataRefs: dataRefs}
		// Set the root header when the writer is closed and we are at the top level index.
		if w.closed && w.levels[level].cw.ChunkCount() == 1 {
			w.root = hdr
			return nil
		}
		// Create next index level if it does not exist.
		if level == len(w.levels)-1 {
			cw := w.chunks.NewWriter(w.ctx, averageBits, w.callback(level+1))
			w.levels = append(w.levels, &levelWriter{
				cw: cw,
				tw: tar.NewWriter(cw),
			})
		}
		// Write index entry in next level index.
		return w.writeHeader(hdr, level+1)
	}
}

// Close finishes the index, and returns the serialized top level index.
func (w *Writer) Close() error {
	w.closed = true
	// Note: new levels can be created while closing, so the number of iterations
	// necessary can increase as the levels are being closed. The number of ranges
	// will decrease per level as long as the range size is in general larger than
	// a serialized header. Levels stop getting created when the top level chunk
	// writer has been closed and the number of ranges it has is one.
	for i := 0; i < len(w.levels); i++ {
		l := w.levels[i]
		if err := l.tw.Close(); err != nil {
			return err
		}
		if err := l.cw.Close(); err != nil {
			return err
		}
	}
	// Write the final index level to the path.
	objW, err := w.objC.Writer(w.ctx, w.path)
	if err != nil {
		return err
	}
	tw := tar.NewWriter(objW)
	if err := w.serialize(tw, w.root); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return objW.Close()
}
