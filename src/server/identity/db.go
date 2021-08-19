package identity

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// CreateUsersTable sets up the postgres table which tracks active IDP users
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CreateUsersTable(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
CREATE TABLE identity.users (
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
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CreateConfigTable(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
CREATE TABLE identity.config (
	id INT PRIMARY KEY,
	issuer VARCHAR(4096)
);
`)
	return err
}

// AddTokenExpiryConfig adds expiry fields for token lifespan to the server config
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func AddTokenExpiryConfig(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
ALTER TABLE identity.config ADD COLUMN
	id_token_expiry VARCHAR(4096)`)
	return err
}
