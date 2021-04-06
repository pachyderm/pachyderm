package auth

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
)

// TODO(acohen4): Decide on hash as primary key or unique column
// CreateTokensTable sets up the postgres table which tracks active clusters
func CreateAuthTokensTable(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS auth.auth_tokens (
	token_hash VARCHAR(4096) PRIMARY KEY,
	subject VARCHAR(64) NOT NULL,
	source VARCHAR(64) NOT NULL,
	expiration TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`)
	return err
}
