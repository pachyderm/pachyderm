package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(env serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string) (*validatedAPIServer, error) {
	a, err := newAPIServer(env, txnEnv, etcdPrefix)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env), nil
}
