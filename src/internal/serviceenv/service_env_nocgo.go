//go:build !cgo
// +build !cgo

package serviceenv

import (
	dex_sql "github.com/dexidp/dex/storage/sql"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
)

// InitDexDB initiates the connection to postgres to populate the Dex DB.
// This is only required in pods where the identity service is running,
// otherwise it creates an extra unnecessary postgres connection.
//
// NB: This should only be compiled without cgo, to avoid linking in sqlite
// which requires a newish libc
func (env *NonblockingServiceEnv) InitDexDB() {
	env.dexDBEg.Go(func() (err error) {
		env.dexDB, err = (&dex_sql.Postgres{
			NetworkDB: dex_sql.NetworkDB{
				Database: env.Config().IdentityServerDatabase,
				User:     env.Config().PostgresUser,
				Password: env.Config().PostgresPassword,
				Host:     env.Config().PGBouncerHost,
				Port:     uint16(env.Config().PGBouncerPort),
			},
			SSL: dex_sql.SSL{
				Mode: dbutil.SSLModeDisable,
			},
		}).Open(env.Logger().WithField("source", "identity-db"))
		return
	})
}
