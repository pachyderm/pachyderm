package index

import (
	"bytes"
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"modernc.org/mathutil"
)

// Reader is used for reading a multilevel index.
type Reader struct {
	objC   obj.Client
	chunks *chunk.Storage
	path   string
	filter *pathFilter
}

type pathFilter struct {
	pathRange *PathRange
	prefix    string
}

// NewReader create a new Reader.
func NewReader(objC obj.Client, chunks *chunk.Storage, path string, opts ...Option) *Reader {
	r := &Reader{
		objC:   objC,
		chunks: chunks,
		path:   path,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Iterate iterates over the indexes.
func (r *Reader) Iterate(ctx context.Context, cb func(*Index) error) error {
	// Setup top level reader.
	pbr, err := topLevel(ctx, r.objC, r.path)
	if err != nil {
		return err
	}
	levels := []pbutil.Reader{pbr}
	for {
		pbr := levels[len(levels)-1]
		idx := &Index{}
		if err := pbr.Read(idx); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		// TODO An empty fileset is represented by no referenced data
		// stored in the semantic path for the fileset. We should probably spend some more time
		// thinking through the implications of this representation.
		if idx.Range == nil && idx.FileOp == nil {
			return nil
		}
		// Return if done.
		if r.atEnd(idx.Path) {
			return nil
		}
		// Handle lowest level index.
		if idx.Range == nil {
			// Skip to the starting index.
			if !r.atStart(idx.Path) {
				continue
			}
			resolveDataOps(idx)
			if err := cb(idx); err != nil {
				return err
			}
			continue
		}
		// Skip to the starting index.
		if !r.atStart(idx.Range.LastPath) {
			continue
		}
		levels = append(levels, pbutil.NewReader(newLevelReader(ctx, pbr, r.chunks, idx)))
	}
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

// atStart returns true when the name is in the valid range for a filter (always true if no filter is set).
// For a range filter, this means the name is >= to the lower bound.
// For a prefix filter, this means the name is >= to the prefix.
func (r *Reader) atStart(name string) bool {
	if r.filter == nil {
		return true
	}
	if r.filter.pathRange != nil && r.filter.pathRange.Lower != "" {
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
	if r.filter.pathRange != nil && r.filter.pathRange.Upper != "" {
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
	ctx    context.Context
	parent pbutil.Reader
	chunks *chunk.Storage
	idx    *Index
	buf    *bytes.Buffer
}

func newLevelReader(ctx context.Context, parent pbutil.Reader, chunks *chunk.Storage, idx *Index) *levelReader {
	return &levelReader{
		ctx:    ctx,
		parent: parent,
		chunks: chunks,
		idx:    idx,
	}
}

// Read reads data from an index level.
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
		r := lr.chunks.NewReader(lr.ctx, []*chunk.DataRef{lr.idx.Range.ChunkRef})
		lr.buf = &bytes.Buffer{}
		if err := r.Get(lr.buf); err != nil {
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
	r := lr.chunks.NewReader(lr.ctx, []*chunk.DataRef{lr.idx.Range.ChunkRef})
	lr.buf.Reset()
	return r.Get(lr.buf)
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
