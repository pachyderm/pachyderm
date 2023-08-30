package v2_5_0

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

type row struct {
	key  string
	cols []any
}

type postgresBatcher struct {
	tx      *pachsql.Tx
	table   string
	columns []string
	set     string
	asNames string
	rows    []row
	max     int
	num     int
}

// newPostgresBatcher creates a new batcher.  It expects that one column will be
// named “key,” and that there will be no column named “old_key.”
func newPostgresBatcher(tx *pachsql.Tx, table string, columns ...string) *postgresBatcher {
	var (
		ss = make([]string, len(columns))
		qc = make([]string, len(columns))
	)
	for i, col := range columns {
		if col == "proto" {
			// It’s a code smell that newPostgresBatcher knows about
			// decoding a proto column, while the caller of Add must
			// know about encoding it.
			ss[i] = fmt.Sprintf(`%s = decode(x.%s, 'hex')`, col, col)
		} else {
			ss[i] = fmt.Sprintf(`%s = x.%s`, col, col)
		}
		qc[i] = col
	}
	qc = append(qc, "old_key")
	return &postgresBatcher{
		tx:      tx,
		table:   table,
		columns: columns,
		set:     strings.Join(ss, ", "),
		asNames: strings.Join(qc, ", "),
		max:     100,
	}
}

// Add adds a row to be updated. It expects that the columns will be in the
// order originally provided to newPostgresBatcher, plus a final old key column.
func (pb *postgresBatcher) Add(ctx context.Context, key string, cols []any) error {
	if len(cols) != len(pb.columns)+1 {
		return errors.Errorf("got %d columns; expected %d", len(cols), len(pb.columns))
	}
	pb.rows = append(pb.rows, row{key, cols})
	if len(pb.rows) < pb.max {
		return nil
	}
	return pb.execute(ctx)
}

func (pb *postgresBatcher) execute(ctx context.Context) error {
	if len(pb.rows) == 0 {
		return nil
	}
	log.Info(ctx, "executing postgres statement batch",
		zap.String("number", strconv.Itoa(pb.num)),
		zap.String("size", strconv.Itoa(len(pb.rows))),
	)
	pb.num++
	var (
		placeholders []string
		i            = 1
		args         []any
	)
	for _, row := range pb.rows {
		var placeholder []string
		for j := range pb.columns {
			placeholder = append(placeholder, fmt.Sprintf("$%d", i))
			i++
			args = append(args, row.cols[j])
		}
		placeholder = append(placeholder, fmt.Sprintf("$%d", i))
		i++
		args = append(args, row.cols[len(row.cols)-1]) // grab the old key
		placeholders = append(placeholders, "("+strings.Join(placeholder, ", ")+")")
	}

	updateStmt := fmt.Sprintf(`UPDATE collections."%s" AS p SET %s FROM (VALUES %s) AS x (%s) WHERE p."key" = x."old_key"`, pb.table, pb.set, strings.Join(placeholders, ", "), pb.asNames)
	if _, err := pb.tx.ExecContext(ctx, updateStmt, args...); err != nil {
		return errors.Wrapf(err, "could not execute %s", updateStmt)
	}
	pb.rows = nil
	return nil
}

func (pb *postgresBatcher) Finish(ctx context.Context) error {
	return pb.execute(ctx)
}
