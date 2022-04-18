package sdata

import (
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// SQLTupleWriter writes tuples to a SQL database.
type SQLTupleWriter struct {
	tx              *pachsql.Tx
	insertStatement string
}

func (m *SQLTupleWriter) WriteTuple(t Tuple) error {
	_, err := m.tx.Exec(m.insertStatement, t...)
	return errors.EnsureStack(err)
}

func (m *SQLTupleWriter) Flush() error {
	return nil
}

func NewSQLTupleWriter(tx *pachsql.Tx, tableInfo pachsql.TableInfo) *SQLTupleWriter {
	placeholders := make([]string, len(tableInfo.Columns))
	for i := range tableInfo.Columns {
		placeholders[i] = pachsql.Placeholder(tx.DriverName(), i)
	}
	insertStatement := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
		tableInfo.Schema,
		tableInfo.Name,
		strings.Join(tableInfo.ColumnNames(), ", "),
		strings.Join(placeholders, ", "))
	return &SQLTupleWriter{tx, insertStatement}
}
