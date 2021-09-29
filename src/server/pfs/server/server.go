package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// NewAPIServer creates an APIServer and runs the master process.
func NewAPIServer(env serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string) (pfs.APIServer, error) {
	a, err := newAPIServer(env, txnEnv, etcdPrefix)
	if err != nil {
		return nil, err
	}
	// Setup PFS master
	go a.driver.master(env.Context())
	return newValidatedAPIServer(a, env), nil
}

// NewSidecarAPIServer creates an APIServer.
func NewSidecarAPIServer(env serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string) (pfs.APIServer, error) {
	a, err := newAPIServer(env, txnEnv, etcdPrefix)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env), nil
}