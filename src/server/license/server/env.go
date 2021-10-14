package server

import (
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/sirupsen/logrus"
)

// Env is the dependencies required for a license API Server
type Env struct {
	DB       *sqlx.DB
	Listener collection.PostgresListener

	Logger *logrus.Logger
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv) Env {
	return Env{
		DB:       senv.GetDBClient(),
		Listener: senv.GetPostgresListener(),
		Logger:   senv.Logger(),
	}
}
