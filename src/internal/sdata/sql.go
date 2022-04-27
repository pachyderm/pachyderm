package sdata

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// limit the number of rows
const rowLimit = 1000

// SQLTupleWriter writes tuples to a SQL database.
type SQLTupleWriter struct {
	tx              *pachsql.Tx
	tableInfo       *pachsql.TableInfo
	insertStatement string
	buf             []interface{}
}

func (m *SQLTupleWriter) WriteTuple(t Tuple) error {
	if len(m.buf)/len(m.tableInfo.Columns) >= rowLimit {
		m.Flush()
	}
	for i := range t {
		v := reflect.New(reflect.TypeOf(t[i]).Elem())
		v.Elem().Set(reflect.ValueOf(t[i]).Elem())
		m.buf = append(m.buf, v.Interface())
	}
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
	// TODO handle result
	_, err = stmt.Exec(m.buf...)
	if err != nil {
		return errors.EnsureStack(err)
	}
	m.buf = m.buf[:0]
	return nil
}

// GeneratePreparedStatement generates a prepared statement based the amount of data in the buffer.
func (m *SQLTupleWriter) GeneratePreparedStatement() (*pachsql.Stmt, error) {
	if len(m.buf) == 0 {
		return nil, nil
	}
	driverName := m.tx.DriverName()
	placeholders := []string{} // a list of (?, ?, ...)

	// construct list of placeholders by accumulating elements into a row first
	row := []string{}
	for i := 0; i < len(m.buf); i++ {
		row = append(row, pachsql.Placeholder(driverName, i))
		if (i+1)%len(m.tableInfo.Columns) == 0 {
			placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(row, ", ")))
			row = row[:0]
		}
	}
	sqlStr := m.insertStatement + strings.Join(placeholders, ", ")
	stmt, err := m.tx.Preparex(sqlStr)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return stmt, nil
}

func NewSQLTupleWriter(tx *pachsql.Tx, tableInfo *pachsql.TableInfo) *SQLTupleWriter {
	insertStatement := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES ",
		tableInfo.Schema,
		tableInfo.Name,
		strings.Join(tableInfo.ColumnNames(), ", "))
	return &SQLTupleWriter{tx, tableInfo, insertStatement, []interface{}{}}
}
