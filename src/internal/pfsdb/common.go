package pfsdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

type (
	SortOrder      int
	FilterOperator string
)

const (
	SortNone SortOrder = iota
	SortAscend
	SortDescend
)

const (
	// Note that we use `?` as the placeholder for bindvar, but postgres uses $1, $2, etc.
	// We solve this by rebinding the query before executing it using sqlx.Rebind.
	Equal              FilterOperator = "%s = ?"
	GreaterThan        FilterOperator = "%s > ?"
	GreaterThanOrEqual FilterOperator = "%s >= ?"
	LessThan           FilterOperator = "%s < ?"
	LessThanOrEqual    FilterOperator = "%s <= ?"
	NotEqual           FilterOperator = "%s <> ?"
	ValueIn            FilterOperator = "%s in (?)"
)

type (
	ModelType interface{ Repo | Commit | Branch }
	FieldType interface{ CommitField | BranchField }
	PairType  interface{ CommitPair | BranchWithID }
)

type Filter[T FieldType] struct {
	Field  T
	Value  any
	values []any
	Op     FilterOperator
}

func (f *Filter[T]) QueryString() (formatted string, err error) {
	formatted = fmt.Sprintf(string(f.Op), f.Field)
	if f.Op == ValueIn {
		formatted, f.values, err = sqlx.In(formatted, f.Value)
		if err != nil {
			return "", errors.Wrapf(err, "expanding %s", formatted)
		}
	}
	return
}

type AndFilters[T FieldType] []*Filter[T]

func (f AndFilters[T]) QueryString() (string, error) {
	var parts []string
	for _, filter := range f {
		part, err := filter.QueryString()
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, " AND "), nil
}

type OrderBy[T FieldType] struct {
	Fields    []T
	SortOrder SortOrder
}

func (o *OrderBy[T]) QueryString() string {
	var parts []string
	for _, field := range o.Fields {
		parts = append(parts, string(field))
	}
	var order string
	switch o.SortOrder {
	case SortAscend:
		order = "ASC"
	case SortDescend:
		order = "DESC"
	}
	return strings.Join(parts, ", ") + " " + order
}

// T is the type used to deserialize rows from the database.
// U is the type returned by the iterator.
// The caller supplies the transform function, which converts T to U.
type pageIterator[T ModelType] struct {
	query         string
	values        []any
	limit, offset uint64
	page          []T
	pageIdx       int
	// transform     transformFn[T, U] // converts T to U, and is called for each item in the page
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
	// open a transaction and get the next page of results
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
