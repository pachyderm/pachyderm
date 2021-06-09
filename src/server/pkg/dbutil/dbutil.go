package dbutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// set this to true if you want to keep the database around
var devDontDropDatabase = false

const (
	// DefaultPostgresHost for tests
	DefaultPostgresHost = "127.0.0.1"
	// DefaultPostgresPort for tests
	DefaultPostgresPort = 32228
	// DefaultPostgresDatabase for tests
	DefaultPostgresDatabase = "pgc"
	// TestPostgresUser is the default postgres user
	TestPostgresUser = "pachyderm"
)

// NewTestDB connects to postgres using the default settings, creates a database with a unique name
// then calls cb with a sqlx.DB configured to use the newly created database.
// After cb returns the database is dropped.
func NewTestDB(t testing.TB) *sqlx.DB {
	host := os.Getenv("POSTGRES_SERVICE_HOST")
	if host == "" {
		host = DefaultPostgresHost
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, DefaultPostgresPort, TestPostgresUser, DefaultPostgresDatabase)
	db := sqlx.MustOpen("postgres", dsn)
	dbName := fmt.Sprintf("test_%d", time.Now().UnixNano())
	db.MustExec("CREATE DATABASE " + dbName)
	t.Log("database", dbName, "successfully created")
	t.Cleanup(func() {
		if !devDontDropDatabase {
			db.MustExec("DROP DATABASE " + dbName)
		}
		require.NoError(t, db.Close())
	})
	db2 := sqlx.MustOpen("postgres", dsn+" dbname="+dbName)
	t.Cleanup(func() {
		require.NoError(t, db2.Close())
	})
	return db2
}

// DBParams are parameters passed to the db constructor.
type DBParams struct {
	Host       string
	Port       int
	User, Pass string
	DBName     string
}

// NewDB returns a db created with the given parameters
func NewDB(x DBParams) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", x.Host, x.Port, x.User, x.Pass, x.DBName)
	return sqlx.Open("postgres", dsn)
}
