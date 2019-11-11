package index

import (
	"context"
	"io"

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
	cw      *chunk.Writer
	tw      *tar.Writer
	lastHdr *Header
}

type meta struct {
	hdr   *Header
	level int
}

// Writer is used for creating a multi-level index into a serialized FileSet.
// Each index level consists of compressed tar stream chunks.
// Each index tar entry has the full index in the content section.
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

// WriteHeaders writes a set of Header to the index.
func (w *Writer) WriteHeaders(hdrs []*Header) error {
	w.setupLevels()
	for _, hdr := range hdrs {
		hdr.Hdr.Typeflag = indexType
	}
	return w.writeHeaders(hdrs, 0)
}

func (w *Writer) setupLevels() {
	// Setup the first level.
	if w.levels == nil {
		cw := w.chunks.NewWriter(w.ctx, averageBits, w.callback(0), 0)
		w.levels = append(w.levels, &levelWriter{
			cw: cw,
			tw: tar.NewWriter(cw),
		})
	}
}

func (w *Writer) writeHeaders(hdrs []*Header, level int) error {
	l := w.levels[level]
	for _, hdr := range hdrs {
		// Create an annotation for each header.
		l.cw.Annotate(&chunk.Annotation{
			RefDataRefs: hdr.Idx.DataOp.DataRefs,
			Meta: &meta{
				hdr:   hdr,
				level: level,
			},
		})
		if err := w.serialize(l.tw, hdr); err != nil {
			return err
		}
	}
	return nil
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
	return func(chunkRef *chunk.DataRef, annotations []*chunk.Annotation) error {
		if len(annotations) == 0 {
			return nil
		}
		lw := w.levels[level]
		// Extract first and last header and setup file range.
		hdr := annotations[0].Meta.(*meta).hdr
		offset := annotations[0].Offset
		// Edge case handling.
		if len(annotations) > 1 {
			// Skip the first header if it started in the previous chunk.
			if lw.lastHdr != nil && hdr.Hdr.Name == lw.lastHdr.Hdr.Name {
				hdr = annotations[1].Meta.(*meta).hdr
				offset = annotations[1].Offset
			}
		}
		lw.lastHdr = annotations[len(annotations)-1].Meta.(*meta).hdr
		// Set standard fields in index header.
		var lastPath string
		if lw.lastHdr.Hdr.Typeflag == indexType {
			lastPath = lw.lastHdr.Hdr.Name
		} else {
			lastPath = lw.lastHdr.Idx.Range.LastPath
		}
		hdr.Hdr.Typeflag = rangeType
		hdr.Idx.Range = &Range{
			Offset:   offset,
			LastPath: lastPath,
		}
		hdr.Idx.DataOp = &DataOp{DataRefs: []*chunk.DataRef{chunkRef}}
		hdr.Idx.LastPathChunk = lw.lastHdr.Idx.LastPathChunk
		// Set the root header when the writer is closed and we are at the top level index.
		if w.closed {
			w.root = hdr
		}
		// Create next index level if it does not exist.
		if level == len(w.levels)-1 {
			cw := w.chunks.NewWriter(w.ctx, averageBits, w.callback(level+1), int64(level+1))
			w.levels = append(w.levels, &levelWriter{
				cw: cw,
				tw: tar.NewWriter(cw),
			})
		}
		// Write index entry in next level index.
		return w.writeHeaders([]*Header{hdr}, level+1)
	}
}

// WriteCopyFunc executes a function for copying data to the writer.
func (w *Writer) WriteCopyFunc(f func() (*Copy, error)) error {
	w.setupLevels()
	for {
		c, err := f()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		// Finish level below.
		// (bryce) this is going to run twice during the copy process (going up the levels, then back down).
		// This is probably fine, but may be worth noting.
		if c.level > 0 {
			lw := w.levels[c.level-1]
			lw.tw = tar.NewWriter(lw.cw)
			if err := lw.cw.Flush(); err != nil {
				return err
			}
			lw.cw.Reset()
		}
		// Write the raw bytes first (handles bytes hanging over at the end)
		if c.raw != nil {
			if err := w.levels[c.level].cw.WriteCopy(c.raw); err != nil {
				return err
			}
		}
		// Write the headers to be copied.
		if c.hdrs != nil {
			if err := w.writeHeaders(c.hdrs, c.level); err != nil {
				return err
			}
		}
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
		if err := l.tw.Flush(); err != nil {
			return err
		}
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
	tw := tar.NewWriter(objW)
	if err := w.serialize(tw, w.root); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return objW.Close()
}
