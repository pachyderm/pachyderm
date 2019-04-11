package server

import (
	"github.com/pachyderm/pachyderm/src/client/transaction"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps/server"
)

// APIServer represents an api server.
type APIServer interface {
	transaction.APIServer
}

// NewAPIServer creates an APIServer.
func NewAPIServer(
	env *serviceenv.ServiceEnv,
	etcdPrefix string,
	pfsServer *pfsserver.APIServer,
	ppsServer *ppsserver.APIServer,
	authServer *authserver.APIServer,
	memoryRequest int64,
) (APIServer, error) {
	return newAPIServer(env, etcdPrefix, pfsServer, ppsServer, authServer, memoryRequest)
}
