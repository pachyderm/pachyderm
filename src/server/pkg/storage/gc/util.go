package gc

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"golang.org/x/sync/errgroup"
)

func NewDB(host, port, user, pass, dbname string) (*gorm.DB, error) {
	db, err := dbutil.NewGORMDB(host, port, user, pass, dbname)
	if err != nil {
		return nil, err
	}
	if err := initializeDb(db); err != nil {
		return nil, err
	}
	return db, nil
}

// WithTestDB creates a local database client for testing during the lifetime of
// the callback.
func WithTestDB(t *testing.T, f func(*gorm.DB) error) (retErr error) {
	dbutil.WithTestDB(t, func(db *sqlx.DB) {
		var dbName string
		if err := db.Get(&dbName, `SELECT current_database()`); err != nil {
			retErr = err
			return
		}
		if err := db.Close(); err != nil {
			retErr = err
			return
		}
		db2, err := dbutil.NewGORMDB(dbutil.DefaultPostgresHost, strconv.Itoa(dbutil.DefaultPostgresPort), "postgres", "password-ignored", dbName)
		if err != nil {
			retErr = err
			return
		}
		defer db2.Close()
		if err := initializeDb(db2); err != nil {
			retErr = err
			return
		}
		if err := clearData(db2); err != nil {
			retErr = err
			return
		}
		if err := f(db2); err != nil {
			retErr = err
			return
		}
	})
	return retErr
}

func clearData(db *gorm.DB) error {
	if err := db.Exec("DELETE FROM chunks *").Error; err != nil {
		return err
	}
	return db.Exec("DELETE FROM refs *").Error

}

// WithLocalGarbageCollector creates a local garbage collector client for testing during the lifetime of
// the callback.
func WithLocalGarbageCollector(t *testing.T, f func(context.Context, obj.Client, Client) error, opts ...Option) error {
	return obj.WithLocalClient(func(objClient obj.Client) error {
		return WithTestDB(t, func(db *gorm.DB) error {
			return WithGarbageCollector(objClient, db, func(ctx context.Context, client Client) error {
				return f(ctx, objClient, client)
			}, opts...)
		})
	})
}

// WithGarbageCollector creates a garbage collector client for testing during the lifetime of
// the callback.
func WithGarbageCollector(objClient obj.Client, db *gorm.DB, f func(context.Context, Client) error, opts ...Option) error {
	client, err := NewClient(db)
	if err != nil {
		return err
	}
	// TODO May want to pipe a real context through here.
	cancelCtx, cancel := context.WithCancel(context.Background())
	eg, gcContext := errgroup.WithContext(cancelCtx)
	eg.Go(func() error {
		return Run(gcContext, objClient, db, opts...)
	})
	eg.Go(func() error {
		defer cancel()
		return f(gcContext, client)
	})
	return eg.Wait()
}

func isRetriableError(err error) bool {
	pqErr := &pq.Error{}
	if errors.As(err, &pqErr) {
		return pqErr.Code.Class().Name() == "transaction_rollback"
	}
	return false
}

type statementFunc func(*gorm.DB) *gorm.DB

func runTransaction(ctx context.Context, db *gorm.DB, stmtFuncs []statementFunc) error {
	for {
		err := tryTransaction(ctx, db, stmtFuncs)
		if isRetriableError(err) {
			continue
		}
		return err
	}
}

func tryTransaction(ctx context.Context, db *gorm.DB, stmtFuncs []statementFunc) error {
	txn := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	// So we don't leak in case of a panic somewhere unexpected
	defer func() {
		if r := recover(); r != nil {
			txn.Rollback()
		}
	}()
	if err := txn.Error; err != nil {
		return err
	}
	for _, stmtFunc := range stmtFuncs {
		if err := stmtFunc(txn).Error; err != nil {
			txn.Rollback()
			return err
		}
	}
	if err := txn.Commit().Error; err != nil {
		txn.Rollback()
		return err
	}
	return nil
}

func retry(ctx context.Context, name string, f func() error) error {
	return backoff.RetryUntilCancel(ctx, f, backoff.NewExponentialBackOff(), func(err error, _ time.Duration) error {
		fmt.Printf("%v: %v\n", name, err)
		return nil
	})
}
