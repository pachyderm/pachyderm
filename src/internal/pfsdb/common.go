package pfsdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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
	Equal   FilterOperator = "%s = ?"
	ValueIn FilterOperator = "%s in (?)"
)

type ModelType interface {
	Repo | Commit | Branch
}

type FieldType interface {
	CommitField | BranchField
}

type ProtoType interface {
	pfs.CommitInfo | pfs.BranchInfo
}
type Filter[T CommitField | BranchField] struct {
	Field  T
	Value  any
	Values []any // for ValueIn expressions, is there a better way?
	Op     FilterOperator
}

func (f *Filter[T]) String() (formatted string, err error) {
	formatted = fmt.Sprintf(string(f.Op), f.Field)
	if f.Op == ValueIn {
		formatted, f.Values, err = sqlx.In(formatted, f.Values) // Thank god for this!
	}
	return
}

type AndFilters[T FieldType] []Filter[T]

func (f AndFilters[T]) String() (string, error) {
	var parts []string
	for _, filter := range f {
		part, err := filter.String()
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, " AND "), nil
}

type QueryBuilder[T FieldType] struct {
	baseQuery   string
	queryParams []any

	AndFilters AndFilters[T]
	GroupBy    string
	OrderBy    string
	// Limit, Offset uint64
}

func (c *QueryBuilder[T]) Build() (string, error) {
	query := c.baseQuery
	condition, err := c.AndFilters.String()
	if err != nil {
		return "", err
	}
	for _, filter := range c.AndFilters {
		if filter.Op == ValueIn {
			c.queryParams = append(c.queryParams, filter.Values...)
		} else {
			c.queryParams = append(c.queryParams, filter.Value)
		}
	}
	if condition != "" {
		query += "\nWHERE " + condition
	}
	// TODO implement GROUP BY and ORDER BY in a type safe way
	if c.GroupBy != "" {
		query += "\nGROUP BY " + c.GroupBy
	}
	if c.OrderBy != "" {
		query += "\nORDER BY " + c.OrderBy
	}
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	return query, nil
}

type pageIterator[T ModelType, U any] struct {
	db            *pachsql.DB
	baseQuery     string
	queryParams   []any
	limit, offset uint64
	page          []*U
	pageIndex     int
}

func newPageIterator[T ModelType, U ProtoType, S FieldType](ctx context.Context, db *pachsql.DB, qb QueryBuilder[S], limit, offset uint64) (pageIterator[T, U], error) {
	baseQuery, err := qb.Build()
	if err != nil {
		return pageIterator[T, U]{}, err
	}
	iter := pageIterator[T, U]{
		db:          db,
		baseQuery:   baseQuery,
		queryParams: qb.queryParams,
		limit:       limit,
		offset:      offset,
	}
	return iter, nil
}

func (i *pageIterator[T, U]) next(ctx context.Context, tansformer func(context.Context, *pachsql.Tx, *T) (*U, error)) (*U, error) {
	if i.pageIndex >= len(i.page) {
		query := i.baseQuery + fmt.Sprintf("\nLIMIT %d OFFSET %d", i.limit, i.offset)
		var rows []T
		var rowsConverted []*U
		// open a transaction and get the next page of results
		if err := dbutil.WithTx(ctx, i.db, func(ctx context.Context, tx *pachsql.Tx) error {
			if err := tx.SelectContext(ctx, &rows, query, i.queryParams...); err != nil {
				return errors.Wrap(err, "getting page")
			}
			for _, row := range rows {
				row := row
				x, err := tansformer(ctx, tx, &row)
				if err != nil {
					return err
				}
				rowsConverted = append(rowsConverted, x)
			}
			return nil
		}); err != nil {
			return nil, err
		}
		if len(rowsConverted) == 0 {
			return nil, stream.EOS()
		}
		i.page = rowsConverted
		i.pageIndex = 0
		i.offset += i.limit
	}
	item := i.page[i.pageIndex]
	i.pageIndex++
	return item, nil
}
