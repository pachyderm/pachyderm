package server

import (
	"github.com/pachyderm/pachyderm/src/client/transaction"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
)

// APIServer represents and api server.
type APIServer interface {
	transaction.APIServer
}

// NewAPIServer creates an APIServer.
func NewAPIServer(env *serviceenv.ServiceEnv, etcdPrefix string, memoryRequest int64) (APIServer, error) {
	return newAPIServer(env, etcdPrefix, memoryRequest)
}
