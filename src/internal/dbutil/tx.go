package dbutil

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/sirupsen/logrus"
)

type withTxConfig struct {
	sql.TxOptions
}

// WithTxOption parameterizes the WithTx function
type WithTxOption func(c *withTxConfig)

// WithIsolationLevel runs the transaction with the specified isolation level.
func WithIsolationLevel(x sql.IsolationLevel) WithTxOption {
	return func(c *withTxConfig) {
		c.TxOptions.Isolation = x
	}
}

// WithReadOnly causes WithTx to run the transaction as read only
func WithReadOnly() WithTxOption {
	return func(c *withTxConfig) {
		c.TxOptions.ReadOnly = true
	}
}

// WithTx calls cb with a transaction,
// The transaction is committed IFF cb returns nil.
// If cb returns an error the transaction is rolled back.
func WithTx(ctx context.Context, db *sqlx.DB, cb func(tx *sqlx.Tx) error, opts ...WithTxOption) error {
	c := &withTxConfig{
		TxOptions: sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.InitialInterval = 10 * time.Millisecond
	backoffStrategy.MaxElapsedTime = 0
	return backoff.RetryUntilCancel(ctx, func() error {
		tx, err := db.BeginTxx(ctx, &c.TxOptions)
		if err != nil {
			return err
		}
		return tryTxFunc(tx, cb)
	}, backoffStrategy, func(err error, _ time.Duration) error {
		if isTransactionError(err) {
			return nil
		}
		return err
	})
}

func tryTxFunc(tx *sqlx.Tx, cb func(tx *sqlx.Tx) error) error {
	if err := cb(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logrus.Error(rbErr)
		}
		return err
	}
	return tx.Commit()
}

func isTransactionError(err error) bool {
	pqerr := &pq.Error{}
	if errors.As(err, pqerr) {
		return pgerrcode.IsTransactionRollback(string(pqerr.Code))
	}
	return IsErrTransactionConflict(err)
}

// ErrTransactionConflict should be used by user code to indicate a conflict in
// the transaction that should be reattempted.
type ErrTransactionConflict struct{}

func (err ErrTransactionConflict) Is(other error) bool {
	_, ok := other.(ErrTransactionConflict)
	return ok
}

func (err ErrTransactionConflict) Error() string {
	return "transaction conflict, will be reattempted"
}

// IsErrTransactionConflict determines if an error is an ErrTransactionConflict error
func IsErrTransactionConflict(err error) bool {
	return errors.Is(err, ErrTransactionConflict{})
}
