package pfsdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
	ValueIn FilterExpression = "%s in (%s)"
)

type Expression interface {
	QueryPart() string
}

type Filter struct {
	Field    string
	Value    any
	Values   []any // for
	Function FilterExpression
}

func (f *Filter) QueryPart() string {
	if f.Function == ValueIn && len(f.Values) > 0 {
		values := make([]string, len(f.Values))
		for i, value := range f.Values {
			values[i] = fmt.Sprintf("%v", value)
		}
		return fmt.Sprintf(string(f.Function), f.Field, strings.Join(values, ","))
	}
	return fmt.Sprintf(string(f.Function), f.Field, f.Value)
}

type ListResourceConfig struct {
	BaseQuery  string
	AndFilters []Filter
	GroupBy    Expression
	OrderBy    Expression
}

func (c *ListResourceConfig) Query() string {
	query := c.BaseQuery
	if len(c.AndFilters) > 0 {
		query += "\nWHERE "
		for i, filter := range c.AndFilters {
			query += filter.QueryPart()
			if i < len(c.AndFilters)-1 {
				query += " AND "
			}
		}
	}
	if c.GroupBy != nil {
		groupBy := c.GroupBy.QueryPart()
		if groupBy != "" {
			query += "\nGROUP BY " + groupBy
		}
	}
	if c.OrderBy != nil {
		orderBy := c.OrderBy.QueryPart()
		if orderBy != "" {
			query += "\nORDER BY " + orderBy
		}
	}
	return query
}

// T can be a CommitPair
type Iterator[T any, A any] struct {
	db            *pachsql.DB
	baseQuery     string
	limit, offset uint64

	items    []A
	curIndex int
}

func NewIterator[T any, A any](ctx context.Context, db *pachsql.DB, config *ListResourceConfig, limit, offset uint64) (*Iterator[T, A], error) {
	iter := &Iterator[T, A]{
		db:        db,
		baseQuery: config.Query(),
		limit:     limit,
		offset:    offset,
	}
	return iter, nil
}

func (i *Iterator[T, A]) Next(ctx context.Context, dst *A) error {
	if dst == nil {
		return errors.New("dst cannot be nil")
	}
	if i.curIndex >= len(i.items) {
		query := i.baseQuery + fmt.Sprintf("\nLIMIT %d OFFSET %d", i.limit, i.offset)
		if err := dbutil.WithTx(ctx, i.db, func(ctx context.Context, tx *pachsql.Tx) error {
			var rows []A
			if err := tx.SelectContext(ctx, &rows, query); err != nil {
				return err
			}
			i.items = rows
			return nil
		}); err != nil {
			return err
		}
		if len(i.items) == 0 {
			return stream.EOS()
		}
		i.curIndex = 0
		i.offset += i.limit
	}
	*dst = i.items[i.curIndex]
	i.curIndex++
	return nil
}
