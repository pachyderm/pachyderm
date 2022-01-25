package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

func listUsers(ctx context.Context, db *pachsql.DB) ([]*identity.User, error) {
	users := make([]*identity.User, 0)
	err := db.SelectContext(ctx, &users, "SELECT email, last_authenticated FROM identity.users WHERE enabled=true;")
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return users, nil
}

func addUserInTx(ctx context.Context, tx *pachsql.Tx, email string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO identity.users (email, last_authenticated, enabled) VALUES ($1, now(), true) ON CONFLICT(email) DO UPDATE SET last_authenticated=NOW()`, email)
	return errors.EnsureStack(err)
}
