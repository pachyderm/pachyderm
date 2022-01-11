package dbutil

import (
	"context"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultMaxOpenConns is the default maximum number of open connections; if you change
	// this, also consider changing the default from the environment in
	// serviceenv.GlobalConfiguration.
	DefaultMaxOpenConns = 10
	// DefaultMaxIdleConns is the default number of idle database connections to maintain.  (2
	// comes from the default in database/sql.go.)
	DefaultMaxIdleConns = 2
	// DefaultConnMaxLifetime is the default maximum amount of time a connection may be reused
	// for.  Defaults to no maximum.
	DefaultConnMaxLifetime = 0
	// DefaultConnMaxIdleTime is the default maximum amount of time a connection may be idle.
	// Defaults to no maximum.
	DefaultConnMaxIdleTime = 0
)

type dbConfig struct {
	host            string
	port            int
	user, password  string
	name            string
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	sslMode         string
}

func newConfig(opts ...Option) *dbConfig {
	dbc := &dbConfig{
		maxOpenConns:    DefaultMaxOpenConns,
		maxIdleConns:    DefaultMaxIdleConns,
		connMaxLifetime: DefaultConnMaxLifetime,
		connMaxIdleTime: DefaultConnMaxIdleTime,
		sslMode:         DefaultSSLMode,
	}
	for _, opt := range opts {
		opt(dbc)
	}
	return dbc
}

func getDSN(dbc *dbConfig) string {
	fields := map[string]string{
		"connect_timeout": "30",
		"sslmode":         dbc.sslMode,

		// https://github.com/jackc/pgx/issues/650#issuecomment-568212888
		// both of the options below are mentioned as solutions for working with pg_bouncer
		// prefer_simple_protocol causes some of our types to fail to serialize.
		"statement_cache_mode": "describe",
		//"prefer_simple_protocol": "true",
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
	return strings.Join(dsnParts, " ")
}

// GetDSN returns the string for connecting to the postgres instance with the
// parameters specified in 'opts'. This is needed because 'listen' operations
// are not supported in generic SQL libraries and they need to be run in a side
// session.
func GetDSN(opts ...Option) string {
	dbc := newConfig(opts...)
	return getDSN(dbc)
}

// NewDB creates a new DB.
func NewDB(opts ...Option) (*pachsql.DB, error) {
	dbc := newConfig(opts...)
	if dbc.name == "" {
		panic("must specify database name")
	}
	if dbc.host == "" {
		panic("must specify database host")
	}
	if dbc.user == "" {
		panic("must specify user")
	}
	dsn := getDSN(dbc)
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if dbc.maxOpenConns != 0 {
		db.SetMaxOpenConns(dbc.maxOpenConns)
	}
	// Always set these; 0 does not mean "use the default", it means "use zero".
	db.SetMaxIdleConns(dbc.maxIdleConns)
	db.SetConnMaxLifetime(dbc.connMaxLifetime)
	db.SetConnMaxIdleTime(dbc.connMaxIdleTime)
	return db, nil
}

// WaitUntilReady attempts to ping the database until the context is cancelled.
// Progress information is written to log
func WaitUntilReady(ctx context.Context, log *logrus.Logger, db *pachsql.DB) error {
	const period = time.Second
	const timeout = time.Second
	log.Infof("waiting for db to be ready...")
	return backoff.RetryUntilCancel(ctx, func() error {
		log.Debugf("pinging db...")
		ctx, cf := context.WithTimeout(ctx, timeout)
		defer cf()
		if err := db.PingContext(ctx); err != nil {
			log.Infof("db is not ready: %v", err)
			return errors.EnsureStack(err)
		}
		log.Infof("db is ready")
		return nil
	}, backoff.NewConstantBackOff(period), nil)
}
