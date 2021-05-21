package identity

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// CreateUsersTable sets up the postgres table which tracks active IDP users
func CreateUsersTable(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS identity.users (
	email VARCHAR(4096) PRIMARY KEY,
	last_authenticated TIMESTAMP,
	enabled BOOLEAN
);
`)
	return errors.EnsureStack(err)
}

// CreateConfigTable sets up the postgres table which stores IDP configuration.
// Dex usually loads config from a file, but reconfiguring via RPCs makes it
// faster for users to iterate on finding the correct values.
// `issuer` must be a well-known URL where all pachds can reach this server.
func CreateConfigTable(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS identity.config (
	id INT PRIMARY KEY,
	issuer VARCHAR(4096)
);
`)
	return err
}
