// Package transactiondb contains the database schema that Pachyderm
// transactions use.
package transactiondb

import (
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

const (
	transactionsCollectionName = "transactions"
)

// AllCollections returns a list of all the Transaction API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(transactionsCollectionName, nil, nil, nil, nil, nil),
	}
}

// Transactions returns a collection of open transactions
func Transactions(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		transactionsCollectionName,
		db,
		listener,
		&transaction.TransactionInfo{},
		nil,
		nil,
	)
}
