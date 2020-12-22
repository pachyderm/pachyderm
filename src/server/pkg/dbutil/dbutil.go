package dbutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// set this to true if you want to keep the database around
var devDontDropDatabase = false

const (
	// DefaultHost is the default host.
	DefaultHost = "127.0.0.1"
	// DefaultPort is the default port.
	DefaultPort = 32228
	// DefaultUser is the default user
	DefaultUser = "postgres"
	// DefaultDBName is the default DB name.
	DefaultDBName = "pgc"
)

// NewTestDB connects to postgres using the default settings, creates a database with a unique name
// then calls cb with a sqlx.DB configured to use the newly created database.
// After cb returns the database is dropped.
func NewTestDB(t testing.TB) *sqlx.DB {
	db, err := NewDB()
	require.NoError(t, err)
	dbName := fmt.Sprintf("test_%d", time.Now().UnixNano())
	db.MustExec("CREATE DATABASE " + dbName)
	t.Log("database", dbName, "successfully created")
	t.Cleanup(func() {
		if !devDontDropDatabase {
			db.MustExec("DROP DATABASE " + dbName)
		}
		require.NoError(t, db.Close())
	})
	db2, err := NewDB(WithDBName(dbName))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db2.Close())
	})
	return db2
}

type dBConfig struct {
	host           string
	port           int
	user, password string
	name           string
}

// NewDB creates a new DB.
func NewDB(opts ...Option) (*sqlx.DB, error) {
	dbc := &dBConfig{
		host: DefaultHost,
		port: DefaultPort,
		user: DefaultUser,
		name: DefaultDBName,
	}
	for _, opt := range opts {
		opt(dbc)
	}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbc.host, dbc.port, dbc.user, dbc.password, dbc.name)
	return sqlx.Open("postgres", dsn)
}
