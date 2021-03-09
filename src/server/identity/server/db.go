package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/identity"

	"github.com/jmoiron/sqlx"
)

func listUsers(ctx context.Context, db *sqlx.DB) ([]*identity.User, error) {
	users := make([]*identity.User, 0)
	err := db.SelectContext(ctx, &users, "SELECT email, last_authenticated FROM identity.users WHERE enabled=true;")
	if err != nil {
		return nil, err
	}
	return users, nil
}

func addUserInTx(ctx context.Context, tx *sqlx.Tx, email string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO identity.users (email, last_authenticated, enabled) VALUES ($1, now(), true) ON CONFLICT(email) DO UPDATE SET last_authenticated=NOW()`, email)
	return err
}
