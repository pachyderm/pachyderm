package index

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"modernc.org/mathutil"
)

// Reader is used for reading a multilevel index.
type Reader struct {
	ctx     context.Context
	objC    obj.Client
	chunks  *chunk.Storage
	path    string
	filter  *pathFilter
	levels  []pbutil.Reader
	peekIdx *Index
	done    bool
}

type pathFilter struct {
	pathRange *PathRange
	prefix    string
}

// NewReader create a new Reader.
// (bryce) might make sense to register copy function here.
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

func (r *Reader) Peek() (*Index, error) {
	if err := r.setup(); err != nil {
		return nil, err
	}
	var err error
	if r.peekIdx == nil {
		r.peekIdx, err = r.next()
	}
	return r.peekIdx, err

}

func (r *Reader) setup() error {
	if r.done {
		return io.EOF
	}
	if r.levels == nil {
		// Setup top level reader.
		pbr, err := topLevel(r.ctx, r.objC, r.path)
		if err != nil {
			return err
		}
		r.levels = []pbutil.Reader{pbr}
	}
	return nil
}

func topLevel(ctx context.Context, objC obj.Client, path string) (pbr pbutil.Reader, retErr error) {
	objR, err := objC.Reader(ctx, path, 0, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := objR.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, objR); err != nil {
		return nil, err
	}
	return pbutil.NewReader(buf), nil
}

func (r *Reader) Next() (*Index, error) {
	if err := r.setup(); err != nil {
		return nil, err
	}
	if r.peekIdx != nil {
		idx := r.peekIdx
		r.peekIdx = nil
		return idx, nil
	}
	return r.next()
}

func (r *Reader) next() (*Index, error) {
	for {
		pbr := r.levels[len(r.levels)-1]
		idx := &Index{}
		if err := pbr.Read(idx); err != nil {
			return nil, err
		}
		// Return if done.
		if r.atEnd(idx.Path) {
			r.done = true
			return nil, io.EOF
		}
		// Handle lowest level index.
		if idx.Range == nil {
			// Skip to the starting index.
			if !r.atStart(idx.Path) {
				continue
			}
			return idx, nil
		}
		// Skip to the starting index.
		if !r.atStart(idx.Range.LastPath) {
			continue
		}
		r.levels = append(r.levels, pbutil.NewReader(newLevelReader(pbr, r.chunks.NewReader(r.ctx), idx)))
	}
}

// atStart returns true when the name is in the valid range for a filter (always true if no filter is set).
// For a range filter, this means the name is >= to the lower bound.
// For a prefix filter, this means the name is >= to the prefix.
func (r *Reader) atStart(name string) bool {
	if r.filter == nil {
		return true
	}
	if r.filter.pathRange != nil {
		return name >= r.filter.pathRange.Lower
	}
	return name >= r.filter.prefix
}

// atEnd returns true when the name is past the valid range for a filter (always false if no filter is set).
// For a range filter, this means the name is > than the upper bound.
// For a prefix filter, this means the name does not have the prefix and a name with the prefix cannot show up after it.
func (r *Reader) atEnd(name string) bool {
	if r.filter == nil {
		return false
	}
	if r.filter.pathRange != nil {
		return name > r.filter.pathRange.Upper
	}
	// Name is past a prefix when the first len(prefix) bytes are greater than the prefix
	// (use len(name) bytes for comparison when len(name) < len(prefix)).
	// A simple greater than check would not suffice here for the prefix filter functionality
	// (for example, if the index consisted of the paths "a", "ab", "abc", and "b", then a
	// reader with the prefix filter set to "a" would end at the "ab" path rather than the "b" path).
	cmpSize := mathutil.Min(len(name), len(r.filter.prefix))
	return name[:cmpSize] > r.filter.prefix[:cmpSize]
}

type levelReader struct {
	parent pbutil.Reader
	cr     *chunk.Reader
	idx    *Index
	buf    *bytes.Buffer
}

func newLevelReader(parent pbutil.Reader, cr *chunk.Reader, idx *Index) *levelReader {
	return &levelReader{
		parent: parent,
		cr:     cr,
		idx:    idx,
	}
}

func (lr *levelReader) Read(data []byte) (int, error) {
	if err := lr.setup(); err != nil {
		return 0, err
	}
	var bytesRead int
	for len(data) > 0 {
		if lr.buf.Len() == 0 {
			if err := lr.next(); err != nil {
				return bytesRead, err
			}
		}
		n, _ := lr.buf.Read(data)
		bytesRead += n
		data = data[n:]
	}
	return bytesRead, nil
}

func (lr *levelReader) setup() error {
	if lr.buf == nil {
		lr.cr.NextDataRefs(lr.idx.DataOp.DataRefs)
		lr.buf = &bytes.Buffer{}
		if err := lr.cr.Get(lr.buf); err != nil {
			return err
		}
		// Skip offset bytes to get to first index entry in chunk.
		lr.buf = bytes.NewBuffer(lr.buf.Bytes()[lr.idx.Range.Offset:])
	}
	return nil
}

func (lr *levelReader) next() error {
	lr.idx.Reset()
	if err := lr.parent.Read(lr.idx); err != nil {
		return err
	}
	lr.cr.NextDataRefs(lr.idx.DataOp.DataRefs)
	lr.buf.Reset()
	return lr.cr.Get(lr.buf)
}

func (r *Reader) Iterate(f func(*Index) error, pathBound ...string) error {
	for {
		idx, err := r.Peek()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !BeforeBound(idx.Path, pathBound...) {
			return nil
		}
		if err := f(idx); err != nil {
			return err
		}
		if _, err := r.Next(); err != nil {
			return err
		}
	}
}

// BeforeBound checks if the passed in string is before the string bound (exclusive).
// The string bound is optional, so if no string bound is passed then it returns true.
func BeforeBound(str string, strBound ...string) bool {
	return len(strBound) == 0 || strings.Compare(str, strBound[0]) < 0
}

// GetTopLevelIndex gets the top level index entry for a file set, which contains metadata
// for the file set.
func GetTopLevelIndex(ctx context.Context, objC obj.Client, path string) (*Index, error) {
	pbr, err := topLevel(ctx, objC, path)
	if err != nil {
		return nil, err
	}
	idx := &Index{}
	if err := pbr.Read(idx); err != nil {
		return nil, err
	}
	return idx, nil
}

//// Copy is the basic data structure to represent a copy of data from
//// a reader to a writer.
//type Copy struct {
//	level int
//	raw   *chunk.Copy
//	hdrs  []*Header
//}
//
//// ReadCopyFunc returns a function for copying data from the reader.
//func (r *Reader) ReadCopyFunc(pathBound ...string) func() (*Copy, error) {
//	level := -1
//	var offset int64
//	var done bool
//	return func() (*Copy, error) {
//		if done {
//			return nil, io.EOF
//		}
//		// Setup levels, initialize first level.
//		if err := r.setupLevels(); err != nil {
//			return nil, err
//		}
//		if level < 0 {
//			level = len(r.levels) - 1
//		}
//		c := &Copy{level: r.wLevel(level)}
//		cr := r.levels[level].cr
//		// Handle index entries that span multiple chunks.
//		if offset > 0 {
//			raw, err := cr.ReadCopy(offset)
//			if err != nil {
//				return nil, err
//			}
//			c.raw = raw
//		}
//		// (bryce) this is janky, but we need the current header when copying a level above.
//		if r.levels[level].currHdr != nil {
//			r.levels[level].peekHdr, r.levels[level].currHdr = r.levels[level].currHdr, nil
//		}
//		// While not past a split point, get index entries to copy.
//		pastSplit := false
//		cr.OnSplit(func() { pastSplit = true })
//		for !pastSplit {
//			hdr, err := r.peek(level)
//			if err != nil {
//				if err == io.EOF {
//					done = true
//					return c, nil
//				}
//				return nil, err
//			}
//			// Stop copying when the last referenced (directly or indirectly) content
//			// chunk has a path that is >= the path bound.
//			if !BeforeBound(hdr.Idx.LastPathChunk, pathBound...) {
//				if hdr.Idx.Range == nil {
//					done = true
//					return c, nil
//				}
//				level++
//				offset = hdr.Idx.Range.Offset
//				return c, nil
//			}
//			c.hdrs = append(c.hdrs, hdr)
//			_, err = r.next(level)
//			if err != nil {
//				return nil, err
//			}
//		}
//		level--
//		return c, nil
//	}
//}
//
//// wLevel converts a reader level to the corresponding writer level.
//// The reason this is necessary is because the corresponding writer level is the inverse of the reader level.
//// The reader levels start at the top level index entry while the writer levels start at the content level.
//func (r *Reader) wLevel(rLevel int) int {
//	return len(r.levels) - 1 - rLevel
//}
