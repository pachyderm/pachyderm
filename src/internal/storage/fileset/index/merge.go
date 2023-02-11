package index

import (
	"context"

	proto "github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

func Merge(ctx context.Context, storage *chunk.Storage, indexes []*Index, cb func(*Index) error) error {
	var its []stream.Peekable[*Index]
	for _, index := range indexes {
		ir := NewReader(storage, nil, index)
		iterateFunc := func(cb func(*Index) error) error {
			return ir.Iterate(ctx, func(index *Index) error {
				return cb(index)
			})
		}
		it := miscutil.NewIterator(ctx, iterateFunc)
		peekIt := stream.NewPeekable(it, copyIndex)
		its = append(its, peekIt)
	}
	m := stream.NewReducer(its, compareIndexes, copyIndex)
	return stream.ForEach[*Index](ctx, m, cb)
}

func compareIndexes(a, b *Index) bool {
	if a.Path != b.Path {
		return a.Path < b.Path
	}
	return a.File.Datum < b.File.Datum
}

func copyIndex(dst, src **Index) {
	*dst = proto.Clone(*src).(*Index)
}
