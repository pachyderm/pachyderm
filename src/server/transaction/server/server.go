package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

// APIServer represents an api server.
type APIServer interface {
	transaction.APIServer
	txnenv.TransactionServer
}

type Env struct {
	DB         *pachsql.DB
	PGListener collection.PostgresListener
	TxnEnv     *transactionenv.TransactionEnv
}

// NewAPIServer creates an APIServer.
func NewAPIServer(env Env) (APIServer, error) {
	return newAPIServer(env)
}
