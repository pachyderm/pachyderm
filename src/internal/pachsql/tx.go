package pachsql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Tx struct {
	tx  *sqlx.Tx
	ctx context.Context
}

func (tx *Tx) DriverName() string {
	return tx.tx.DriverName()
}

func (tx *Tx) Exec(ctx context.Context, q string, args ...any) (sql.Result, error) {
	if err := tx.checkContext(ctx); err != nil {
		return nil, err
	}
	return tx.tx.ExecContext(tx.ctx, q, args...)
}

func (tx *Tx) Query(ctx context.Context, q string, args ...any) (*sqlx.Rows, error) {
	if err := tx.checkContext(ctx); err != nil {
		return nil, err
	}
	return tx.tx.QueryxContext(tx.ctx, q, args...)
}

func (tx *Tx) Get(ctx context.Context, dst any, q string, args ...any) error {
	if err := tx.checkContext(ctx); err != nil {
		return err
	}
	return tx.tx.GetContext(tx.ctx, dst, q, args...)
}

func (tx *Tx) Select(ctx context.Context, dst any, q string, args ...any) error {
	if err := tx.checkContext(ctx); err != nil {
		return err
	}
	return tx.tx.SelectContext(tx.ctx, dst, q, args...)
}

func (tx *Tx) checkContext(ctx context.Context) error {
	if inTx := getCtxInTx(tx.ctx); !inTx {
		return errors.New("external context passed to transaction methods")
	}
	return nil
}

func DoTx(ctx context.Context, db *sqlx.DB, fn func(ctx context.Context, tx *Tx) error) (retErr error) {
	if getCtxInTx(ctx) {
		return errors.New("pachsql: nested transaction detected")
	}
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, tx.Rollback())
		}
	}()

	tx2 := &Tx{ctx: ctx, tx: tx}
	ctx2 := newTxContext(ctx)
	// TODO: the hard part, will copy from dbutil
	if err := fn(ctx2, tx2); err != nil {
		return err
	}
	return tx.Commit()
}

func DoTx1[T any](ctx context.Context, db *sqlx.DB, fn func(ctx context.Context, tx *Tx) (T, error)) (T, error) {
	var ret T
	err := DoTx(ctx, db, func(ctx context.Context, tx *Tx) error {
		var err error
		ret, err = fn(ctx, tx)
		return err
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return ret, err
}

var _ context.Context = &txContext{}

type txContext struct {
	parent context.Context
	done   chan struct{}
}

func newTxContext(ctx context.Context) txContext {
	ctx = setCtxInTx(ctx)
	done := make(chan struct{})
	close(done)
	return txContext{parent: ctx, done: done}
}

func (tx txContext) Value(x any) any {
	return tx.parent.Value(x)
}

func (tx txContext) Deadline() (time.Time, bool) {
	return time.Time{}, true
}

func (tx txContext) Done() <-chan struct{} {
	return tx.done
}

func (tx txContext) Err() error {
	return errors.New("pachsql: detected blocking operation in transaction")
}

type ctxKeyInTx struct{}

func setCtxInTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyInTx{}, struct{}{})
}

func getCtxInTx(ctx context.Context) bool {
	v := ctx.Value(ctxKeyInTx{})
	return v != nil
}
