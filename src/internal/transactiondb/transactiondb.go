// Package transactiondb contains the database schema that Pachyderm
// transactions use.
package transactiondb

import (
	"context"

	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

const (
	transactionsPrefix = "/transactions"
)

// Transactions returns a collection of open transactions
func Transactions(ctx context.Context, db *sqlx.DB, listener *col.PostgresListener) (col.PostgresCollection, error) {
	return col.NewPostgresCollection(
		ctx,
		db,
		listener,
		&transaction.TransactionInfo{},
		nil,
	)
}
