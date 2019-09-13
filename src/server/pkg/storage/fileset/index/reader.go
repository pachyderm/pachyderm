package index

import (
	"bytes"
	"context"
	"io"
	"math"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

type levelReader struct {
	cr               *chunk.Reader
	tr               *tar.Reader
	currHdr, peekHdr *Header
}

type pathFilter struct {
	pathRange *PathRange
	prefix    string
}

// Reader is used for reading a multi-level index.
type Reader struct {
	ctx    context.Context
	objC   obj.Client
	chunks *chunk.Storage
	path   string
	filter *pathFilter
	levels []*levelReader
	done   bool
}

// NewReader create a new Reader.
func NewReader(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string, opts ...Option) *Reader {
	r := &Reader{
		ctx:    ctx,
		objC:   objC,
		chunks: chunks,
		path:   path,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Next gets the next header in the index.
func (r *Reader) Next() (*Header, error) {
	if err := r.setupLevels(); err != nil {
		return nil, err
	}
	if r.levels[0].peekHdr != nil {
		hdr := r.levels[0].peekHdr
		r.levels[0].peekHdr = nil
		return hdr, nil
	}
	if r.done {
		return nil, io.EOF
	}
	return r.next(len(r.levels) - 1)
}

func (r *Reader) setupLevels() error {
	if r.levels != nil {
		return nil
	}
	// Setup top level.
	objR, err := r.objC.Reader(r.ctx, r.path, 0, 0)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, objR); err != nil {
		return err
	}
	if err := objR.Close(); err != nil {
		return err
	}
	r.levels = []*levelReader{&levelReader{tr: tar.NewReader(buf)}}
	// Traverse until we reach the first entry in the lowest level.
	for {
		hdr, err := r.peek(len(r.levels) - 1)
		if err != nil {
			return err
		}
		// Return when we are at the lowest level.
		if hdr.Hdr.Typeflag == indexType {
			return nil
		}
		_, err = r.next(len(r.levels) - 1)
		if err != nil {
			return err
		}
		// Setup next level.
		// (bryce) this whole process of reading at the offset is janky, needs to be re-thought.
		cr := r.chunks.NewReader(r.ctx, r.callback(len(r.levels)))
		dataRef := hdr.Idx.DataOp.DataRefs[0]
		dataRef.OffsetBytes = hdr.Idx.Range.Offset
		dataRef.SizeBytes -= hdr.Idx.Range.Offset
		cr.NextRange([]*chunk.DataRef{dataRef})
		r.levels = append(r.levels, &levelReader{
			cr: cr,
			tr: tar.NewReader(cr),
		})
	}
}

func (r *Reader) peek(level int) (*Header, error) {
	l := r.levels[level]
	if l.peekHdr != nil {
		return l.peekHdr, nil
	}
	var err error
	l.peekHdr, err = r.next(level)
	return l.peekHdr, err
}

func (r *Reader) next(level int) (*Header, error) {
	l := r.levels[level]
	if l.peekHdr != nil {
		hdr := l.peekHdr
		l.peekHdr = nil
		return hdr, nil
	}
	for {
		hdr, err := l.tr.Next()
		if err != nil {
			return nil, err
		}
		// Return if done.
		if r.atEnd(hdr.Name) {
			r.done = true
			return nil, io.EOF
		}
		// Handle lowest level index.
		if hdr.Typeflag == indexType {
			// Skip to the starting header.
			if !r.atStart(hdr.Name) {
				continue
			}
			return deserialize(l.tr, hdr)
		}
		// Handle index level above lowest.
		fullHdr, err := deserialize(l.tr, hdr)
		if err != nil {
			return nil, err
		}
		// Skip to the starting header.
		if !r.atStart(fullHdr.Idx.Range.LastPath) {
			continue
		}
		return fullHdr, nil
	}
}

func (r *Reader) atStart(name string) bool {
	if r.filter == nil {
		return true
	}
	if r.filter.pathRange != nil {
		return name >= r.filter.pathRange.Lower
	}
	return name >= r.filter.prefix
}

func (r *Reader) atEnd(name string) bool {
	if r.filter == nil {
		return false
	}
	if r.filter.pathRange != nil {
		return name > r.filter.pathRange.Upper
	}
	cmpSize := int64(math.Min(float64(len(name)), float64(len(r.filter.prefix))))
	return name[:cmpSize] > r.filter.prefix[:cmpSize]
}

// Peek peeks ahead in the index.
func (r *Reader) Peek() (*Header, error) {
	if err := r.setupLevels(); err != nil {
		return nil, err
	}
	if r.levels[0].peekHdr != nil {
		return r.levels[0].peekHdr, nil
	}
	var err error
	r.levels[0].peekHdr, err = r.Next()
	return r.levels[0].peekHdr, err
}

func (r *Reader) callback(level int) chunk.ReaderFunc {
	return func() ([]*chunk.DataRef, error) {
		hdr, err := r.next(level - 1)
		if err != nil {
			return nil, err
		}
		r.levels[level-1].currHdr = hdr
		return hdr.Idx.DataOp.DataRefs, nil
	}
}

// Copy is the basic data structure to represent a copy of data from
// a reader to a writer.
type Copy struct {
	level int
	raw   *chunk.Copy
	hdrs  []*Header
}

// ReadCopyFunc returns a function for copying data from the reader.
func (r *Reader) ReadCopyFunc(pathBound ...string) func() (*Copy, error) {
	level := -1
	var offset int64
	var done bool
	return func() (*Copy, error) {
		if done {
			return nil, io.EOF
		}
		// Setup levels, initialize first level.
		if err := r.setupLevels(); err != nil {
			return nil, err
		}
		if level < 0 {
			level = len(r.levels) - 1
		}
		c := &Copy{level: r.wLevel(level)}
		cr := r.levels[level].cr
		// Handle index entries that span multiple chunks.
		if offset > 0 {
			raw, err := cr.ReadCopy(offset)
			if err != nil {
				return nil, err
			}
			c.raw = raw
		}
		// (bryce) this is janky, but we need the current header when copying a level above.
		if r.levels[level].currHdr != nil {
			r.levels[level].peekHdr, r.levels[level].currHdr = r.levels[level].currHdr, nil
		}
		// While not past a split point, get index entries to copy.
		pastSplit := false
		cr.OnSplit(func() { pastSplit = true })
		for !pastSplit {
			hdr, err := r.peek(level)
			if err != nil {
				if err == io.EOF {
					done = true
					return c, nil
				}
				return nil, err
			}
			// Stop copying when the last referenced (directly or indirectly) content
			// chunk has a path that is >= the path bound.
			if !BeforeBound(hdr.Idx.LastPathChunk, pathBound...) {
				if hdr.Idx.Range == nil {
					done = true
					return c, nil
				}
				level++
				offset = hdr.Idx.Range.Offset
				return c, nil
			}
			c.hdrs = append(c.hdrs, hdr)
			_, err = r.next(level)
			if err != nil {
				return nil, err
			}
		}
		level--
		return c, nil
	}
}

func (r *Reader) wLevel(rLevel int) int {
	return len(r.levels) - 1 - rLevel
}

// BeforeBound checks if the passed in string is before the string bound (exclusive).
// The string bound is optional, so if no string bound is passed then it returns true.
func BeforeBound(str string, strBound ...string) bool {
	return len(strBound) == 0 || strings.Compare(str, strBound[0]) < 0
}

// Close closes the reader.
func (r *Reader) Close() error {
	for i := 1; i < len(r.levels); i++ {
		if err := r.levels[i].cr.Close(); err != nil {
			return err
		}
	}
	return nil
}

func deserialize(tr *tar.Reader, hdr *tar.Header) (*Header, error) {
	data := &bytes.Buffer{}
	buf := grpcutil.GetBuffer()
	defer grpcutil.PutBuffer(buf)
	if _, err := io.CopyBuffer(data, tr, buf); err != nil {
		return nil, err
	}
	idx := &Index{}
	if err := proto.Unmarshal(data.Bytes(), idx); err != nil {
		return nil, err
	}
	return &Header{
		Hdr: hdr,
		Idx: idx,
	}, nil
}
