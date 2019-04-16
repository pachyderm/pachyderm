package server

import (
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs/server"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps/server"
)

// TransactionEnv contains the APIServer instances for each subsystem that may
// be involved in running transactions so that they can make calls to each other
// without leaving the context of a transaction.  This is a separate object
// because there are cyclic dependencies between APIServer instances.
type TransactionEnv struct {
	auth *authserver.APIServer
	pfs  *pfsserver.APIServer
	pps  *ppsserver.APIServer
}

// Initialize stores the references to APIServer instances in the TransactionEnv
func (env *TransactionEnv) Initialize(
	auth *authserver.APIServer,
	pfs *pfsserver.APIServer,
	pps *ppsserver.APIServer,
) {
	env.auth = auth
	env.pfs = pfs
	env.pps = pps
}

func (env *TransactionEnv) AuthServer() *authserver.APIServer {
	return env.auth
}

func (env *TransactionEnv) PfsServer() *pfsserver.APIServer {
	return env.pfs
}

func (env *TransactionEnv) PpsServer() *ppsserver.APIServer {
	return env.pps
}
