package pjsdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

//TODO(Fahad): Too much copy/paste from pfsdb, see if we can move the iterator code into another package which we can share
// between this and other db packages.

// OrderByColumn is used to configure iteratorRequests.
// values for Column and SortOrder are exposed as constants
// with prefixes Column_ such as ColumnID and SortOrder_ such as SortOrderAsc.
type OrderByColumn struct {
	Column    jobColumn
	SortOrder sortOrder
}

// these types are unexported to force callers of this API to use the exported column and sort orders.
type jobColumn string
type sortOrder string

const (
	ColumnID       = jobColumn("j.id")
	ColumnParent   = jobColumn("j.parent")
	ColumnSpecHash = jobColumn("j.program_hash")

	SortOrderAsc    = sortOrder("ASC")
	SortOrderDesc   = sortOrder("DESC")
	defaultPageSize = 100
)

type Iterable interface {
	jobRecord | queueRecord
}

type IteratorConfiguration struct {
	StartPage uint64
	PageSize  uint64          // defaults to 100 if not set
	OrderBys  []OrderByColumn // defaults to id ASC if not set.
}

func (c *IteratorConfiguration) orderBy() string {
	if len(c.OrderBys) == 0 {
		c.OrderBys = []OrderByColumn{{Column: ColumnID, SortOrder: SortOrderAsc}}
	}
	values := make([]string, len(c.OrderBys))
	for i, col := range c.OrderBys {
		values[i] = fmt.Sprintf("%s %s", col.Column, col.SortOrder)
	}
	return "\nORDER BY " + strings.Join(values, ", ")
}

type pageIterator[T Iterable] struct {
	query                   string
	values                  []any
	limit, offset, maxPages uint64
	page                    []T
	pageIndex               int
	pagesSeen               int
}

// if maxPages == 0, then interpret as unlimited pages
func newPageIterator[T Iterable](query string, values []any, startPage, pageSize, maxPages uint64) pageIterator[T] {
	return pageIterator[T]{
		query:    query,
		values:   values,
		limit:    pageSize,
		offset:   startPage * pageSize,
		maxPages: maxPages,
	}
}

func (i *pageIterator[T]) nextPage(ctx context.Context, extCtx sqlx.ExtContext) (err error) {
	defer func() { i.pagesSeen++ }()
	if i.maxPages > 0 && i.pagesSeen >= int(i.maxPages) {
		return stream.EOS()
	}
	var page []T
	query := i.query + fmt.Sprintf("\nLIMIT %d OFFSET %d", i.limit, i.offset)
	if err := sqlx.SelectContext(ctx, extCtx, &page, query, i.values...); err != nil {
		return errors.Wrap(err, "getting page")
	}
	if len(page) == 0 {
		return stream.EOS()
	}
	i.page = page
	i.pageIndex = 0
	i.offset += i.limit
	return nil
}

func (i *pageIterator[T]) hasNext() bool {
	return i.pageIndex < len(i.page)
}

func (i *pageIterator[T]) next(ctx context.Context, extCtx sqlx.ExtContext) (*T, error) {
	if !i.hasNext() {
		if err := i.nextPage(ctx, extCtx); err != nil {
			return nil, err
		}
	}
	t := i.page[i.pageIndex]
	i.pageIndex++
	return &t, nil
}
