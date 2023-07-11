package sdata

import (
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// rowLimit limits the number of rows per batch in an INSERT statement
const rowLimit = 1000

// SQLTupleWriter writes tuples to a SQL database.
type SQLTupleWriter struct {
	tx              *pachsql.Tx
	tableInfo       *pachsql.TableInfo
	insertStatement string
	buf             []Tuple
}

func (m *SQLTupleWriter) WriteTuple(t Tuple) error {
	if len(m.buf) >= rowLimit {
		m.Flush()
	}
	m.buf = append(m.buf, CloneTuple(t))
	return nil
}

func (m *SQLTupleWriter) Flush() error {
	if len(m.buf) == 0 {
		return nil
	}
	stmt, err := m.GeneratePreparedStatement()
	if err != nil {
		return errors.EnsureStack(err)
	}
	// flatten list of Tuple
	var values Tuple
	for r := range m.buf {
		for c := range m.buf[r] { //nolint:gosimple
			values = append(values, m.buf[r][c])
		}
	}
	_, err = stmt.Exec(values...)
	if err != nil {
		return errors.EnsureStack(err)
	}
	m.buf = m.buf[:0]
	return nil
}

// GeneratePreparedStatement generates a prepared statement based the amount of data in the buffer.
// This can be used to execute a batched INSERT.
func (m *SQLTupleWriter) GeneratePreparedStatement() (*pachsql.Stmt, error) {
	if len(m.buf) == 0 {
		return nil, nil
	}
	var placeholders []string // a list of (?, ?, ...)

	for r := range m.buf {
		// construct list of placeholders by accumulating elements into a placeholderRow first
		var placeholderRow []string
		for c := range m.buf[r] {
			i := r*len(m.buf[r]) + c
			placeholderRow = append(placeholderRow, pachsql.Placeholder(m.tableInfo.Driver, i))
		}
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(placeholderRow, ", ")))
	}
	sqlStr := m.insertStatement + strings.Join(placeholders, ", ")
	stmt, err := m.tx.Preparex(sqlStr)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return stmt, nil
}

func NewSQLTupleWriter(tx *pachsql.Tx, tableInfo *pachsql.TableInfo) *SQLTupleWriter {
	var s string
	if tableInfo.Driver == "snowflake" {
		var (
			vv []string
			cc []string
		)
		for i, col := range tableInfo.Columns {
			var v string
			if col.DataType == "VARIANT" {
				v = fmt.Sprintf(`to_variant(COLUMN%d)`, i+1)
			} else {
				v = fmt.Sprintf(`COLUMN%d`, i+1)
			}
			vv = append(vv, v)
			cc = append(cc, col.Name)
		}
		s = fmt.Sprintf(`INSERT INTO %s.%s (%s) SELECT %s FROM VALUES `, tableInfo.Schema, tableInfo.Name, strings.Join(cc, ","), strings.Join(vv, ","))
	} else {
		s = fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES ",
			tableInfo.Schema,
			tableInfo.Name,
			strings.Join(tableInfo.ColumnNames(), ", "))
	}
	return &SQLTupleWriter{tx, tableInfo, s, []Tuple{}}
}
