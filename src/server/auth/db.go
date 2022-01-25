package auth

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// CreateTokensTable sets up the postgres table which tracks active clusters
func CreateAuthTokensTable(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS auth.auth_tokens (
	token_hash VARCHAR(4096) PRIMARY KEY,
	subject VARCHAR(64) NOT NULL,
	expiration TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX subject_index
ON auth.auth_tokens (subject);
`)
	return errors.EnsureStack(err)
}
