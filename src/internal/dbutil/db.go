package dbutil

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	// DefaultHost is the default host.
	DefaultHost = "127.0.0.1"
	// DefaultPort is the default port.
	DefaultPort = 5432
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
	var dsn string
	if uri := os.Getenv("DATABASE_URL"); uri != "" {
		dsn = uri
	} else {
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
		dsn = strings.Join(dsnParts, " ")
	}
	fmt.Printf("DSN: %s", dsn)
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if dbc.maxOpenConns != 0 {
		db.SetMaxOpenConns(dbc.maxOpenConns)
	}
	return db, nil
}

// Interface is the common interface exposed by *sqlx.Tx and *sqlx.DB
type Interface interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
