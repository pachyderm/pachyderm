package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	pfsclient "github.com/pachyderm/pachyderm/v2/src/pfs"
)

// APIServer represents an api server.
type APIServer interface {
	pfsclient.APIServer
	txnenv.PfsTransactionServer
}

// NewAPIServer creates an APIServer.
func NewAPIServer(env serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string) (APIServer, error) {
	a, err := newAPIServer(env, txnEnv, etcdPrefix)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env), nil
}
