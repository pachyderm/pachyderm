package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

// Env is the dependencies required for a license API Server
type Env struct {
	DB       *pachsql.DB
	Listener collection.PostgresListener
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv) Env {
	return Env{
		DB:       senv.GetDBClient(),
		Listener: senv.GetPostgresListener(),
	}
}
