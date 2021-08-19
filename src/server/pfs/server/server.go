package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(senv serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string) (pfs.APIServer, error) {
	env, err := EnvFromServiceEnv(senv)
	if err != nil {
		return nil, err
	}
	env.TxnEnv = txnEnv
	env.EtcdPrefix = etcdPrefix
	a, err := newAPIServer(*env)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env.AuthServer), nil
}
