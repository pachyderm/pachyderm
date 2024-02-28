package pfsdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

type sortOrder string

const (
	SortOrderNone = sortOrder("")
	SortOrderAsc  = sortOrder("ASC")
	SortOrderDesc = sortOrder("DESC")
)

type (
	ModelType interface {
		Repo | Commit | Branch | Project
		GetCreatedAtUpdatedAt() CreatedAtUpdatedAt
	}
	ColumnName interface {
		string | projectColumn | branchColumn | commitColumn | repoColumn
	}
)

type OrderByColumn[T ColumnName] struct {
	Column T
	Order  sortOrder
}

func OrderByQuery[T ColumnName](orderBys ...OrderByColumn[T]) string {
	if len(orderBys) == 0 {
		return ""
	}
	values := make([]string, len(orderBys))
	for i, col := range orderBys {
		values[i] = fmt.Sprintf("%s %s", col.Column, col.Order)
	}
	return "ORDER BY " + strings.Join(values, ", ")
}

type pageIterator[T ModelType] struct {
	query                   string
	values                  []any
	limit, offset, maxPages uint64
	page                    []T
	pageIdx                 int
	lastTimestamp           time.Time
	revision                int64
	pagesSeen               int
}

// if maxPages == 0, then interpret as unlimited pages
func newPageIterator[T ModelType](ctx context.Context, query string, values []any, startPage, pageSize, maxPages uint64) pageIterator[T] {
	return pageIterator[T]{
		query:    query,
		values:   values,
		revision: -1, // first revision should be 0 and we increment before returning.
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
	i.pageIdx = 0
	i.offset += i.limit
	return nil
}

func (i *pageIterator[T]) hasNext() bool {
	return i.pageIdx < len(i.page)
}

func (i *pageIterator[T]) next(ctx context.Context, extCtx sqlx.ExtContext) (*T, int64, error) {
	if !i.hasNext() {
		if err := i.nextPage(ctx, extCtx); err != nil {
			return nil, 0, err
		}
	}
	t := i.page[i.pageIdx]
	createdAt := t.GetCreatedAtUpdatedAt().CreatedAt
	if i.lastTimestamp.Before(createdAt) {
		i.revision++
		i.lastTimestamp = createdAt
	}
	i.pageIdx++
	return &t, i.revision, nil
}

func IsNotFoundError(err error) bool {
	return errors.As(err, &RepoNotFoundError{}) ||
		errors.As(err, &ProjectNotFoundError{}) ||
		errors.As(err, &CommitNotFoundError{}) ||
		errors.As(err, &BranchNotFoundError{})
}
