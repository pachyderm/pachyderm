package pfsdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

type (
	SortOrder        int
	FilterExpression string
)

const (
	SortNone SortOrder = iota
	SortAscend
	SortDescend
)

const (
	Equal   FilterExpression = "%s = %s"
	ValueIn FilterExpression = "%s in (%s)"
)

type Filter[T CommitField | BranchField] struct {
	Field      T
	Value      any
	Values     []any // for "in" expressions
	Expression FilterExpression
}

func (f *Filter[T]) String() string {
	if f.Expression == ValueIn && len(f.Values) > 0 {
		values := make([]string, len(f.Values))
		for i, value := range f.Values {
			values[i] = fmt.Sprintf("%v", value)
		}
		return fmt.Sprintf(string(f.Expression), f.Field, strings.Join(values, ","))
	}
	return fmt.Sprintf(string(f.Expression), f.Field, f.Value)
}

type QueryBuilder[T CommitField | BranchField] struct {
	baseQuery     string
	AndFilters    []Filter[T]
	GroupBy       string
	OrderBy       string
	Limit, Offset uint64
}

func (c *QueryBuilder[T]) Query() string {
	query := c.baseQuery
	if len(c.AndFilters) > 0 {
		filters := make([]string, len(c.AndFilters))
		for i, filter := range c.AndFilters {
			filters[i] = filter.String()
		}
		query += fmt.Sprintf("\nWHERE %s", strings.Join(filters, " AND "))
	}
	if c.GroupBy != "" {
		query += "\nGROUP BY " + c.GroupBy
	}
	if c.OrderBy != "" {
		query += "\nORDER BY " + c.OrderBy
	}
	return query
}

// T is any db struct we use to deserialize rows into.
type pageIterator[T any, U any] struct {
	db            *pachsql.DB
	baseQuery     string
	limit, offset uint64
	page          []*U
	pageIndex     int
}

func newPageIterator[T any, U any, S CommitField | BranchField](ctx context.Context, db *pachsql.DB, config *QueryBuilder[S]) (pageIterator[T, U], error) {
	iter := pageIterator[T, U]{
		db:        db,
		baseQuery: config.Query(),
		limit:     config.Limit,
		offset:    config.Offset,
	}
	return iter, nil
}

func (i *pageIterator[T, U]) next(ctx context.Context, tansformer func(context.Context, *pachsql.Tx, *T) (*U, error)) (*U, error) {
	if i.pageIndex >= len(i.page) {
		query := i.baseQuery + fmt.Sprintf("\nLIMIT %d OFFSET %d", i.limit, i.offset)
		var rows []T
		var newPage []*U
		// open a transaction and get the next page of results
		if err := dbutil.WithTx(ctx, i.db, func(ctx context.Context, tx *pachsql.Tx) error {
			if err := tx.SelectContext(ctx, &rows, query); err != nil {
				return err
			}
			for _, row := range rows {
				row := row
				u, err := tansformer(ctx, tx, &row)
				if err != nil {
					return err
				}
				newPage = append(newPage, u)
			}
			return nil
		}); err != nil {
			return nil, err
		}
		if len(newPage) == 0 {
			return nil, stream.EOS()
		}
		i.page = newPage
		i.pageIndex = 0
		i.offset += i.limit
	}
	item := i.page[i.pageIndex]
	i.pageIndex++
	return item, nil
}
