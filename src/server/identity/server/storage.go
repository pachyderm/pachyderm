package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"

	dex_storage "github.com/dexidp/dex/storage"
	dex_sql "github.com/dexidp/dex/storage/sql"
	log "github.com/sirupsen/logrus"
)

// NewStorageProvider instantiates a new postgres storage provider for Dex.
// This creates a new postgres connection, and it can be shared by multiple
// servers.
func NewStorageProvider(env serviceenv.ServiceEnv) (dex_storage.Storage, error) {
	return (&dex_sql.Postgres{
		NetworkDB: dex_sql.NetworkDB{
			Database: env.Config().IdentityServerDatabase,
			User:     env.Config().IdentityServerUser,
			Password: env.Config().IdentityServerPassword,
			Host:     env.Config().PostgresServiceHost,
			Port:     uint16(env.Config().PostgresServicePort),
		},
		SSL: dex_sql.SSL{
			Mode: env.Config().PostgresServiceSSL,
		},
	}).Open(log.NewEntry(log.New()).WithField("source", "identity-db"))
}
