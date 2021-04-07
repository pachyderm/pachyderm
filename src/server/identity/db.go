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
