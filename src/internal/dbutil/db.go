package dbutil

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultHost is the default host.
	DefaultHost = "127.0.0.1"
	// DefaultPort is the default port.
	DefaultPort = 32228
	// DefaultUser is the default user
	DefaultUser = "postgres"
	// DefaultDBName is the default DB name.
	DefaultDBName = "pgc"
	// DefaultMaxOpenConns is the argument passed to SetMaxOpenConns
	DefaultMaxOpenConns = 3
)

type dBConfig struct {
	host           string
	port           int
	user, password string
	name           string
	maxOpenConns   int
}

// NewDB creates a new DB.
func NewDB(opts ...Option) (*sqlx.DB, error) {
	dbc := &dBConfig{
		host:         DefaultHost,
		port:         DefaultPort,
		user:         DefaultUser,
		name:         DefaultDBName,
		maxOpenConns: DefaultMaxOpenConns,
	}
	for _, opt := range opts {
		opt(dbc)
	}
	fields := map[string]string{
		"sslmode": "disable",
	}
	if dbc.host != "" {
		fields["host"] = dbc.host
	}
	if dbc.port != 0 {
		fields["port"] = strconv.Itoa(dbc.port)
	}
	if dbc.name != "" {
		fields["dbname"] = dbc.name
	}
	if dbc.user != "" {
		fields["user"] = dbc.user
	}
	if dbc.password != "" {
		fields["password"] = dbc.password
	}
	var dsnParts []string
	for k, v := range fields {
		dsnParts = append(dsnParts, k+"="+v)
	}
	dsn := strings.Join(dsnParts, " ")
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if dbc.maxOpenConns != 0 {
		db.SetMaxOpenConns(dbc.maxOpenConns)
	}
	return db, nil
}

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

// WithTx calls cb with a transaction,
// The transaction is committed IFF cb returns nil.
// If cb returns an error the transaction is rolled back.
func WithTx(ctx context.Context, db *sqlx.DB, cb func(tx *sqlx.Tx) error, opts ...WithTxOption) (retErr error) {
	c := &withTxConfig{
		TxOptions: sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
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

// Interface is the common interface exposed by *sqlx.Tx and *sqlx.DB
type Interface interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
