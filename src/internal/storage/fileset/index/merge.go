package index

import (
	"context"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

func Merge(ctx context.Context, storage *chunk.Storage, indexes []*Index, cb func(*Index) error) error {
	var ss []stream.Stream
	for _, index := range indexes {
		ir := NewReader(storage, nil, index)
		iterateFunc := func(cb func(interface{}) error) error {
			return ir.Iterate(ctx, func(index *Index) error {
				return cb(index)
			})
		}
		ss = append(ss, &indexStream{
			iterator: miscutil.NewIterator(ctx, iterateFunc),
		})
	}
	pq := stream.NewPriorityQueue(ss, compare)
	return pq.Iterate(func(ss []stream.Stream) error {
		for _, s := range ss {
			if err := cb(s.(*indexStream).index); err != nil {
				return err
			}
		}
		return nil
	})
}

type indexStream struct {
	iterator *miscutil.Iterator
	index    *Index
}

func (is *indexStream) Next() error {
	data, err := is.iterator.Next()
	if err != nil {
		return err
	}
	is.index = data.(*Index)
	return nil
}

func compare(s1, s2 stream.Stream) int {
	idx1 := s1.(*indexStream).index
	idx2 := s2.(*indexStream).index
	if idx1.Path == idx2.Path {
		return strings.Compare(idx1.File.Datum, idx2.File.Datum)
	}
	return strings.Compare(idx1.Path, idx2.Path)
}
