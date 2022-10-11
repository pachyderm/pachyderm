package index

import (
	"bytes"
	"context"
	"io"

	"github.com/docker/go-units"
	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
)

// Reader is used for reading a multilevel index.
type Reader struct {
	chunks      *chunk.Storage
	cache       *Cache
	filter      *pathFilter
	topIdx      *Index
	datum       string
	shardConfig *ShardConfig
}

// NewReader creates a new Reader.
func NewReader(chunks *chunk.Storage, cache *Cache, topIdx *Index, opts ...Option) *Reader {
	r := &Reader{
		chunks: chunks,
		cache:  cache,
		topIdx: topIdx,
		shardConfig: &ShardConfig{
			NumFiles:  1000000,
			SizeBytes: units.GB,
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Iterate iterates over the lowest level (file type) indexes.
func (r *Reader) Iterate(ctx context.Context, cb func(*Index) error) error {
	if r.topIdx == nil {
		return nil
	}
	traverseCb := func(idx *Index) (bool, error) {
		if atEnd(idx.Path, r.filter) {
			return false, errutil.ErrBreak
		}
		if idx.File != nil {
			if !atStart(idx.Path, r.filter) || !(r.datum == "" || r.datum == idx.File.Datum) {
				return false, nil
			}
			return false, cb(idx)
		}
		if !atStart(idx.Range.LastPath, r.filter) {
			return false, nil
		}
		return true, nil
	}
	_, err := r.traverse(ctx, r.topIdx, []byte{}, traverseCb)
	if errors.Is(err, errutil.ErrBreak) {
		err = nil
	}
	return err
}

// traverse implements traversal through a multilevel index.
// Traversal starts at the provided index and the callback is executed
// for each index entry encountered (range and file type).
// The callback can return true to traverse into the next level, otherwise
// the traversal will continue on the same level.
// The prependBytes and leftoverBytes logic is needed to handle index entries
// that span multiple chunks.
func (r *Reader) traverse(ctx context.Context, idx *Index, prependBytes []byte, cb func(*Index) (bool, error)) ([]byte, error) {
	if idx.File != nil {
		_, err := cb(idx)
		return []byte{}, err
	}
	buf := &bytes.Buffer{}
	buf.Write(prependBytes)
	if err := r.getChunk(ctx, idx, buf); err != nil {
		return nil, err
	}
	pbr := pbutil.NewReader(buf)
	nextPrependBytes := []byte{}
	for {
		leftoverBytes := buf.Bytes()
		idx := &Index{}
		if err := pbr.Read(idx); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return leftoverBytes, nil
			}
			return nil, errors.EnsureStack(err)
		}
		nextLevel, err := cb(idx)
		if err != nil {
			return nil, err
		}
		if nextLevel {
			nextPrependBytes, err = r.traverse(ctx, idx, nextPrependBytes, cb)
			if err != nil {
				return nil, err
			}
		}
	}
}

func (r *Reader) getChunk(ctx context.Context, idx *Index, w io.Writer) error {
	chunkRef := proto.Clone(idx.Range.ChunkRef).(*chunk.DataRef)
	// Skip offset bytes to get to first index entry in chunk.
	// NOTE: Indexes can no longer span multiple chunks, but older
	// versions of Pachyderm could write indexes that span multiple chunks.
	chunkRef.OffsetBytes = idx.Range.Offset
	if r.cache != nil {
		return r.cache.Get(ctx, chunkRef, r.filter, w)
	}
	cr := r.chunks.NewReader(ctx, []*chunk.DataRef{idx.Range.ChunkRef})
	return cr.Get(w)
}

type pathFilter struct {
	pathRange *PathRange
	prefix    string
}

// PathRange is a range of paths.
// The range is inclusive, exclusive: [Lower, Upper).
type PathRange struct {
	Lower, Upper string
}

func (r *PathRange) atStart(path string) bool {
	if r.Lower == "" {
		return true
	}
	return path >= r.Lower
}

func (r *PathRange) atEnd(path string) bool {
	if r.Upper == "" {
		return false
	}
	return path >= r.Upper
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

// ShardConfig is a sharding configuration.
// NumFiles is the number of files to target for each shard.
// SizeBytes is the size, in bytes, to target for each shard.
type ShardConfig struct {
	NumFiles  int64
	SizeBytes int64
}

// Shards creates shards for the index based on the sharding configuration provided to the reader.
// Sharding takes advantage of the NumFiles and SizeBytes index metadata to efficiently traverse the multilevel index.
// A subtree is traversed only when a split point exists within it, which we know based on the NumFiles and SizeBytes
// values at the root of each subtree.
func (r *Reader) Shards(ctx context.Context) ([]*PathRange, error) {
	if r.topIdx == nil || (r.topIdx.NumFiles == 0 && r.topIdx.SizeBytes == 0) {
		return []*PathRange{{}}, nil
	}
	var shards []*PathRange
	pathRange := &PathRange{}
	var numFiles, sizeBytes int64
	traverseCb := func(idx *Index) (bool, error) {
		if numFiles >= r.shardConfig.NumFiles || sizeBytes >= r.shardConfig.SizeBytes {
			pathRange.Upper = idx.Path
			shards = append(shards, pathRange)
			pathRange = &PathRange{
				Lower: idx.Path,
			}
			numFiles = 0
			sizeBytes = 0
		}
		if idx.Range != nil && (numFiles+idx.NumFiles > r.shardConfig.NumFiles || sizeBytes+idx.SizeBytes > r.shardConfig.SizeBytes) {
			return true, nil
		}
		numFiles += idx.NumFiles
		sizeBytes += idx.SizeBytes
		return false, nil
	}
	_, err := r.traverse(ctx, r.topIdx, []byte{}, traverseCb)
	if err != nil {
		return nil, err
	}
	shards = append(shards, pathRange)
	return shards, nil
}
