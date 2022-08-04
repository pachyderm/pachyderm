package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/server/enterprise"
)

// Env is the dependencies required for a license API Server
type Env struct {
	env serviceenv.ServiceEnv
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv) *Env {
	return &Env{
		env: senv,
	}
}

// Delegations.  These are explicit in order to make it clear which parts of the
// service environment are relied upon.
func (e Env) DB() *pachsql.DB                        { return e.env.GetDBClient() }
func (e Env) Listener() collection.PostgresListener  { return e.env.GetPostgresListener() }
func (e Env) Config() *serviceenv.Configuration      { return e.env.Config() }
func (e Env) EnterpriseServer() enterprise.APIServer { return e.env.EnterpriseServer() }
