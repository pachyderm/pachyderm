package index

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

func Merge(ctx context.Context, storage *chunk.Storage, indexes []*Index, cb func(*Index) error) error {
	var its []stream.Peekable[*Index]
	for _, index := range indexes {
		ir := NewReader(storage, nil, index)
		it := stream.NewFromForEach(ctx, copyIndex, func(cb func(*Index) error) error {
			return ir.Iterate(ctx, func(index *Index) error {
				return cb(index)
			})
		})
		peekIt := stream.NewPeekable(it, copyIndex)
		its = append(its, peekIt)
	}
	m := stream.NewMerger(its, LessThan)
	return stream.ForEach[stream.Merged[*Index]](ctx, m, func(x stream.Merged[*Index]) error {
		for _, idx := range x.Values {
			if err := cb(idx); err != nil {
				return err
			}
		}
		return nil
	})
}

func LessThan(a, b *Index) bool {
	if a.Path != b.Path {
		return a.Path < b.Path
	}
	return a.File.Datum < b.File.Datum
}

func copyIndex(dst, src **Index) {
	*dst = *src
}
