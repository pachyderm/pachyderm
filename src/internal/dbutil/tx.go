package dbutil

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// WithTx calls cb with a transaction,
// The transaction is committed IFF cb returns nil.
// If cb returns an error the transaction is rolled back.
func WithTx(ctx context.Context, db *sqlx.DB, cb func(tx *sqlx.Tx) error) (retErr error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logrus.Error(rbErr)
			}
		}
	}()
	if err := cb(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// Interface is the common interface exposed by *sqlx.Tx and *sqlx.DB
type Interface interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
