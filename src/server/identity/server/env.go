package server

import (
	"github.com/dexidp/dex/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

// Env is the set of dependencies required by the API server
type Env struct {
	DB         *pachsql.DB
	DexStorage storage.Storage
	Config     *serviceenv.Configuration
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv) Env {
	return Env{
		DB:         senv.GetDBClient(),
		DexStorage: senv.GetDexDB(),
		Config:     senv.Config(),
	}
}
