package dbutil

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type withTxConfig struct {
	sql.TxOptions
	MaxRetries int
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
		c.ReadOnly = true
	}
}

// WithMaxRetries sets the number of times to retry a transaction because of a serialization failure.
func WithMaxRetries(n int) WithTxOption {
	return func(c *withTxConfig) {
		c.MaxRetries = n
	}
}

// WithTx calls cb with a transaction,
// The transaction is committed IFF cb returns nil.
// If cb returns an error the transaction is rolled back.
func WithTx(ctx context.Context, db *sqlx.DB, cb func(tx *sqlx.Tx) error, opts ...WithTxOption) (retErr error) {
	c := &withTxConfig{
		TxOptions: sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
		MaxRetries: 100,
	}
	for _, opt := range opts {
		opt(c)
	}
	for i := 0; i < c.MaxRetries; i++ {
		tx, err := db.BeginTxx(ctx, &c.TxOptions)
		if err != nil {
			return err
		}
		err = tryTxFunc(tx, cb)
		if isSerializationFailure(err) {
			retErr = err
			time.Sleep(time.Millisecond * 500)
			continue
		}
		return err
	}
	return retErr
}

func tryTxFunc(tx *sqlx.Tx, cb func(tx *sqlx.Tx) error) (retErr error) {
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

func isSerializationFailure(err error) bool {
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}
	return pqErr.Code.Name() == "serialization_failure"
}
