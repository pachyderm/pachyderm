// +build !cgo

package serviceenv

import (
	dex_sql "github.com/dexidp/dex/storage/sql"
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
				User:     env.Config().IdentityServerUser,
				Password: env.Config().IdentityServerPassword,
				Host:     env.Config().PostgresHost,
				Port:     uint16(env.Config().PostgresPort),
			},
			SSL: dex_sql.SSL{
				Mode: env.Config().PostgresSSL,
			},
		}).Open(env.Logger().WithField("source", "identity-db"))
		return
	})
}
