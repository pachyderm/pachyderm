package server

import (
	"github.com/dexidp/dex/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	logrus "github.com/sirupsen/logrus"
)

// Env is the set of dependencies required by the API server
type Env struct {
	env serviceenv.ServiceEnv
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv) Env {
	return Env{
		env: senv,
	}
}

// Delegations to the service environment.  These are explicit in order to make
// it clear which parts of the service environment are relied upon.
func (e Env) DB() *pachsql.DB                   { return e.env.GetDBClient() }
func (e Env) DexStorage() storage.Storage       { return e.env.GetDexDB() }
func (e Env) Logger() *logrus.Logger            { return e.env.Logger() }
func (e Env) Config() *serviceenv.Configuration { return e.env.Config() }
