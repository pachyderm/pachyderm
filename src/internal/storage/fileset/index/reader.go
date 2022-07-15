package index

import (
	"bytes"
	"context"
	"io"

	proto "github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
)

// Reader is used for reading a multilevel index.
type Reader struct {
	chunks *chunk.Storage
	cache  *Cache
	filter *pathFilter
	topIdx *Index
	datum  string
}

type pathFilter struct {
	pathRange *PathRange
	prefix    string
}

// NewReader create a new Reader.
func NewReader(chunks *chunk.Storage, cache *Cache, topIdx *Index, opts ...Option) *Reader {
	r := &Reader{
		chunks: chunks,
		cache:  cache,
		topIdx: topIdx,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Iterate iterates over the indexes.
func (r *Reader) Iterate(ctx context.Context, cb func(*Index) error) error {
	if r.topIdx == nil {
		return nil
	}
	// Setup top level reader.
	pbr := r.topLevel()
	levels := []*level{&level{pbr: pbr}}
	for {
		pbr := levels[len(levels)-1].pbr
		idx := &Index{}
		if err := pbr.Read(idx); err != nil {
			if !errors.Is(err, io.EOF) {
				return errors.EnsureStack(err)
			}
			for {
				if len(levels) == 1 {
					return nil
				}
				idx = levels[len(levels)-1].lr.idx
				levels = levels[:len(levels)-1]
				if idx.File != nil {
					break
				}
			}
		}
		// Return if done.
		if atEnd(idx.Path, r.filter) {
			return nil
		}
		// Handle lowest level index.
		if idx.Range == nil {
			// Skip to the starting index.
			if !atStart(idx.Path, r.filter) {
				continue
			}
			if r.datum == "" || r.datum == idx.File.Datum {
				if err := cb(idx); err != nil {
					if errors.Is(err, errutil.ErrBreak) {
						return nil
					}
					return err
				}
			}
			continue
		}
		// Skip to the starting index.
		if !atStart(idx.Range.LastPath, r.filter) {
			continue
		}
		lr := newLevelReader(ctx, r, pbr, idx)
		levels = append(levels, &level{
			pbr: pbutil.NewReader(lr),
			lr:  lr,
		})
	}
}

func (r *Reader) topLevel() pbutil.Reader {
	buf := bytes.Buffer{}
	pbw := pbutil.NewWriter(&buf)
	pbw.Write(r.topIdx)
	return pbutil.NewReader(&buf)
}

// atStart returns true when the name is in the valid range for a filter (always true if no filter is set).
// For a range filter, this means the name is >= to the lower bound.
// For a prefix filter, this means the name is >= to the prefix.
func atStart(name string, filter *pathFilter) bool {
	if filter == nil {
		return true
	}
	if filter.pathRange != nil {
		return filter.pathRange.atStart(name)
	}
	return name >= filter.prefix
}

// atEnd returns true when the name is past the valid range for a filter (always false if no filter is set).
// For a range filter, this means the name is >= to the upper bound.
// For a prefix filter, this means the name does not have the prefix and a name with the prefix cannot show up after it.
func atEnd(name string, filter *pathFilter) bool {
	if filter == nil {
		return false
	}
	if filter.pathRange != nil {
		return filter.pathRange.atEnd(name)
	}
	// Name is past a prefix when the first len(prefix) bytes are greater than the prefix
	// (use len(name) bytes for comparison when len(name) < len(prefix)).
	// A simple greater than check would not suffice here for the prefix filter functionality
	// (for example, if the index consisted of the paths "a", "ab", "abc", and "b", then a
	// reader with the prefix filter set to "a" would end at the "ab" path rather than the "b" path).
	cmpSize := miscutil.Min(len(name), len(filter.prefix))
	return name[:cmpSize] > filter.prefix[:cmpSize]
}

type level struct {
	pbr pbutil.Reader
	lr  *levelReader
}

type levelReader struct {
	ctx    context.Context
	r      *Reader
	parent pbutil.Reader
	idx    *Index
	buf    *bytes.Buffer
}

func newLevelReader(ctx context.Context, r *Reader, parent pbutil.Reader, idx *Index) *levelReader {
	return &levelReader{
		ctx:    ctx,
		r:      r,
		parent: parent,
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
		lr.buf = &bytes.Buffer{}
		chunkRef := proto.Clone(lr.idx.Range.ChunkRef).(*chunk.DataRef)
		// Skip offset bytes to get to first index entry in chunk.
		// NOTE: Indexes can no longer span multiple chunks, but older
		// versions of Pachyderm could write indexes that span multiple chunks.
		chunkRef.OffsetBytes = lr.idx.Range.Offset
		if lr.r.cache != nil {
			return lr.r.cache.Get(lr.ctx, chunkRef, lr.r.filter, lr.buf)
		}
		cr := lr.r.chunks.NewReader(lr.ctx, []*chunk.DataRef{chunkRef})
		return cr.Get(lr.buf)
	}
	return nil
}

func (lr *levelReader) next() error {
	lr.idx.Reset()
	if err := lr.parent.Read(lr.idx); err != nil {
		return errors.EnsureStack(err)
	}
	if lr.idx.File != nil {
		return io.EOF
	}
	r := lr.r.chunks.NewReader(lr.ctx, []*chunk.DataRef{lr.idx.Range.ChunkRef})
	lr.buf.Reset()
	return r.Get(lr.buf)
}

// ShardCallback is a callback that returns a path range for each shard.
type ShardCallback = func(*PathRange) error

func (r *Reader) Shard(ctx context.Context, cb ShardCallback) error {
	if r.topIdx == nil || r.topIdx.File != nil {
		return cb(&PathRange{})
	}
	idxs := []*Index{r.topIdx}
Loop:
	for len(idxs) < 10 {
		nextIdxs, err := r.nextLevel(ctx, idxs)
		if err != nil {
			return err
		}
		// TODO: Unbalanced shards can result with files & ranges at the same level.
		for _, idx := range nextIdxs {
			if idx.File != nil {
				break Loop
			}
		}
		idxs = nextIdxs
	}
	for i, idx := range idxs {
		pathRange := &PathRange{}
		if i != 0 {
			pathRange.Lower = idx.Path
		}
		if i != len(idxs)-1 {
			pathRange.Upper = idxs[i+1].Path
		}
		if err := cb(pathRange); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) nextLevel(ctx context.Context, idxs []*Index) ([]*Index, error) {
	var dataRefs []*chunk.DataRef
	for _, idx := range idxs {
		dataRefs = append(dataRefs, idx.Range.ChunkRef)
	}
	cr := r.chunks.NewReader(ctx, dataRefs)
	buf := &bytes.Buffer{}
	if err := cr.Get(buf); err != nil {
		return nil, err
	}
	pbr := pbutil.NewReader(buf)
	var nextIdxs []*Index
	for {
		idx := &Index{}
		if err := pbr.Read(idx); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errors.EnsureStack(err)
		}
		nextIdxs = append(nextIdxs, idx)
	}
	return nextIdxs, nil
}
