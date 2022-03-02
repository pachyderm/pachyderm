package sdata

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

var _ TupleWriter = &SQLTableLoader{}

type SQLTableLoader struct {
	tx    *pachsql.Tx
	query string
}

func NewSQLTableLoader(tx *pachsql.Tx, query string) *SQLTableLoader {
	return &SQLTableLoader{
		tx:    tx,
		query: query,
	}
}

func (tl *SQLTableLoader) WriteTuple(row Tuple) error {
	// TODO: batch statements
	_, err := tl.tx.Exec(tl.query, row...)
	return err
}

func (tl *SQLTableLoader) Flush() error {
	return nil
}
