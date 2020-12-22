package server

import (
	"github.com/pachyderm/pachyderm/src/transaction"
	"github.com/pachyderm/pachyderm/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/internal/transactionenv"
)

// APIServer represents an api server.
type APIServer interface {
	transaction.APIServer
	txnenv.TransactionServer
}

// NewAPIServer creates an APIServer.
func NewAPIServer(
	env *serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	etcdPrefix string,
) (APIServer, error) {
	return newAPIServer(env, txnEnv, etcdPrefix)
}
