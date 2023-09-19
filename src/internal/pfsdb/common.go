package pfsdb

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

type sortOrder string

const (
	SortAscend  = sortOrder("ASC")
	SortDescend = sortOrder("DESC")
)

type (
	ModelType interface{ Repo | Commit | Branch }
)

type pageIterator[T ModelType] struct {
	query         string
	values        []any
	limit, offset uint64
	page          []T
	pageIdx       int
}

func newPageIterator[T ModelType](ctx context.Context, tx *pachsql.Tx, query string, values []any, startPage, pageSize uint64) pageIterator[T] {
	return pageIterator[T]{
		query:  query,
		values: values,
		limit:  pageSize,
		offset: startPage * pageSize,
	}
}

func (i *pageIterator[T]) nextPage(ctx context.Context, tx *pachsql.Tx) (err error) {
	var page []T
	query := i.query + fmt.Sprintf("\nLIMIT %d OFFSET %d", i.limit, i.offset)
	if err := tx.SelectContext(ctx, &page, query, i.values...); err != nil {
		return errors.Wrap(err, "getting page")
	}
	if len(page) == 0 {
		return stream.EOS()
	}
	i.page = page
	i.pageIdx = 0
	i.offset += i.limit
	return nil
}

func (i *pageIterator[T]) hasNext() bool {
	return i.pageIdx < len(i.page)
}

func (i *pageIterator[T]) next(ctx context.Context, tx *pachsql.Tx) (*T, error) {
	if !i.hasNext() {
		if err := i.nextPage(ctx, tx); err != nil {
			return nil, err
		}
	}
	t := i.page[i.pageIdx]
	i.pageIdx++
	return &t, nil
}
