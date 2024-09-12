package migrationutils

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

// pg_batcher.go
//
// This module implements a general purpose "batcher" for executing multiple SQL statements with a single
// Postgres call.  The intention is that this can be used by any migration code that needs to perform a
// large number of SQL updates and wishes to do so in a more performant way.
//
// Current limitations:
//    - Only UPDATE statements are supported.
//    - Only one column is supported in the WHERE clause.
//
// It is likely that we'll want to enhance this in the future to support INSERT and/or DELETE statements,
// and to allow for multiple columns in the WHERE clause. The functions have been written with this in mind
// to make them easy to expand upon.

// row contains the values to be used for an insert/update of one row in Postgres.
type row struct {
	columnValues []any // Contains the values to be inserted/updated for any given row
	whereValues  []any // Contains the values to be used in the WHERE clause
}

// postgresBatcher is an object to be used for caching rows of insert/update statements and sending them to Postgres in batches.
type postgresBatcher struct {
	tx        *pachsql.Tx // The transaction object
	action    string      // One of the following: INSERT, UPDATE, DELETE (currently only UPDATE is implemented)
	table     string      // The name of the table
	columns   []string    // The list of columns whose values will be set by the query
	wColumns  []string    // The list of columns used in the WHERE clause
	setString string      // The SET clause for the query (ex: "col1 = x.col1, col2 = x.col2")
	rows      []row       // An array of row objects with values for insert/update/delete
	max       int         // The maximum number of rows per batch
	batchNum  int         // A count of the number of batches executed (for info logging only)
}

// NewPostgresBatcher creates a new batcher.
func NewPostgresBatcher(tx *pachsql.Tx, action string, table string, columns []string, wColumns []string, batchSize int) (*postgresBatcher, error) {

	// Validate the action; currently only UPDATE is
	switch strings.ToUpper(action) {
	case "UPDATE":
	case "INSERT", "DELETE":
		return nil, errors.New("INSERT and DELETE are not currently implemented")
	default:
		return nil, errors.Errorf("%q: invalid action", action)
	}

	// Validate that at least 1 column and no more than 1 wColumn have been provided.
	if len(columns) == 0 {
		return nil, errors.Errorf("at least one column must be provided; none were given")
	}
	if len(wColumns) > 1 {
		return nil, errors.Errorf("currently only 1 column may be used in the WHERE clause, but %d were given", len(wColumns))
	}

	// Array where each element corresponds to the set string for one column (ex: "col1 = x.col1")
	ss := make([]string, len(columns))

	// Loop through the columns, creating the set string for each
	for i, col := range columns {
		if col == "proto" {
			// It’s a code smell that newPostgresBatcher knows about
			// decoding a proto column, while the caller of Add must
			// know about encoding it.
			ss[i] = fmt.Sprintf(`%s = decode(x.%s, 'hex')`, col, col)
		} else {
			ss[i] = fmt.Sprintf(`%s = x.%s`, col, col)
		}
	}
	// Create and return the batcher object
	return &postgresBatcher{
		tx:        tx,
		action:    action,
		table:     table,
		columns:   columns,
		wColumns:  wColumns,
		setString: strings.Join(ss, ", "),
		max:       batchSize,
	}, nil
}

// Add adds a row of data to the batch.
func (pb *postgresBatcher) Add(ctx context.Context, columnValues []any, whereValues []any) error {
	// Make sure that the number of values to be set matches the number of columns
	if got, want := len(columnValues), len(pb.columns); got != want {
		return errors.Errorf("column value count: got %v, but want %v", got, want)
	}
	// Make sure that the number of WHERE clause values matches the number of columns used in the WHERE clause
	if got, want := len(whereValues), len(pb.wColumns); got != want {
		return errors.Errorf("WHERE clause value count: got %v, but want %v", got, want)
	}

	// Append the row with its associated values
	pb.rows = append(pb.rows, row{columnValues, whereValues})

	// If we haven't yet reached the max number of rows for the batch, simply return.
	if len(pb.rows) < pb.max {
		return nil
	}
	// Otherwise, execute the batch and return
	return pb.execute(ctx)
}

// execute will compose the final query string from the rows in the batch and submit it to Postgres.
//
// The final query will look something like the following:
//
//	UPDATE collections.repos AS p
//	SET key = x.key, proto = decode(x.proto, 'hex'), idx_Value = x.idx_Value
//	FROM (VALUES (($1, $2, $3, $4), ($5, $6, $7, $8)))
//	AS x (key, proto, idx_Value, key_where)
//	WHERE p.key = x.key_where
func (pb *postgresBatcher) execute(ctx context.Context) error {
	// If there's not anything to send, just return
	if len(pb.rows) == 0 {
		return nil
	}
	// Increment the batch number
	pb.batchNum++
	// Log the batch number and the number of rows it contains
	log.Info(ctx, "Executing Postgres statements",
		zap.String("Batch number", strconv.Itoa(pb.batchNum)),
		zap.String("Size", strconv.Itoa(len(pb.rows))),
	)
	var (
		placeholders []string // Each element will be the string of placeholders for a row (ex: "($1, $2, $3, $4)")
		i            = 1      // The current placeholder count across all rows (used to generate the $ numbers)
		values       []any    // Each element will be one value, corresponding to one placeholder
	)
	// Loop through all of the rows, constructing the placeholders and values arrays
	for _, row := range pb.rows {
		// Each element of placeholder will be a single placeholder string for the current row (ex: "$1")
		var placeholder []string
		// Loop through each column, creating the placeholder string and appending the value to the values array
		for j := range pb.columns {
			placeholder = append(placeholder, fmt.Sprintf("$%d", i))
			values = append(values, row.columnValues[j])
			i++ // Increment the placeholder count
		}
		// Loop through each WHERE column, creating the placeholder string and appending the value to the values array
		for j := range pb.wColumns {
			placeholder = append(placeholder, fmt.Sprintf("$%d", i))
			values = append(values, row.whereValues[j])
			i++ // Increment the placeholder count
		}

		// Join all of the placeholders for this row with commas separating them, enclose the result in parens,
		// and append to the overall placeholders array
		placeholders = append(placeholders, "("+strings.Join(placeholder, ", ")+")")
	}

	// Construct the query string
	stmt := fmt.Sprintf(`UPDATE %s AS p `, pb.table)
	stmt = stmt + fmt.Sprintf(`SET %s `, pb.setString)
	stmt = stmt + fmt.Sprintf(`FROM (VALUES %s) `, strings.Join(placeholders, ", "))
	stmt = stmt + fmt.Sprintf(`AS x (%s, %s_where) `, strings.Join(pb.columns, ", "), pb.wColumns[0])

	if len(pb.wColumns) > 0 {
		stmt = stmt + fmt.Sprintf(`WHERE p.%s = x.%s_where `, pb.wColumns[0], pb.wColumns[0])
	}

	if _, err := pb.tx.ExecContext(ctx, stmt, values...); err != nil {
		return errors.Wrapf(err, "could not execute %s", stmt)
	}
	pb.rows = nil
	return nil
}

// Finish is called at the end to ensure that the final batch is sent.
func (pb *postgresBatcher) Finish(ctx context.Context) error {
	return pb.execute(ctx)
}

// SimplePostgresBatcher is an alternative batcher that buffers statements 'stmts' terminated by
// semicolons until it hits the max number of statements. The statements are then executed in one
// round-trip query to postgres.
type SimplePostgresBatcher struct {
	tx    *pachsql.Tx
	stmts []string
	max   int
	num   int
}

func NewSimplePostgresBatcher(tx *pachsql.Tx) *SimplePostgresBatcher {
	return &SimplePostgresBatcher{
		tx:  tx,
		max: 100,
	}
}

// Add inserts a statement to the SimplePostgresBatcher's internal buffer. If the number
// of statements exceeds the maximum after adding, the batcher executes the statements.
// SimplePostgresBatcher doesn't support buffering arguments, the arguments are sanitized then
// their values are inlined.
//
// Add does not support SqlNull<X> structs like SqlNullString.
func (pb *SimplePostgresBatcher) Add(ctx context.Context, stmt string, args ...interface{}) error {
	query, err := NewQuery(stmt)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("creating batch query: %s", stmt))
	}
	sanitizedQuery, err := query.Sanitize(args...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("sanitizing query: args: %v", args))
	}
	pb.stmts = append(pb.stmts, sanitizedQuery)
	if len(pb.stmts) < pb.max {
		return nil
	}
	return pb.execute(ctx)
}

func (pb *SimplePostgresBatcher) execute(ctx context.Context) error {
	if len(pb.stmts) == 0 {
		return nil
	}
	log.Info(ctx, "executing postgres statement batch",
		zap.String("number", strconv.Itoa(pb.num)),
		zap.String("size", strconv.Itoa(len(pb.stmts))),
	)
	pb.num++
	if _, err := pb.tx.ExecContext(ctx, strings.Join(pb.stmts, ";")); err != nil {
		return errors.EnsureStack(err)
	}
	pb.stmts = nil
	return nil
}

// Flush triggers the SimplePostgresBatcher to execute any remaining buffered statements.
func (pb *SimplePostgresBatcher) Flush(ctx context.Context) error {
	return pb.execute(ctx)
}
