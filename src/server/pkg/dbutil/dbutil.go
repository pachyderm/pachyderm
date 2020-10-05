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
	defaultPostgresHost = "localhost"
	defaultPostgresPort = 32228
)

// WithTestDB connects to postgres using the default settings, creates a database with a unique name
// then calls cb with a sqlx.DB configured to use the newly created database.
// After cb returns the database is dropped.
func WithTestDB(t *testing.T, cb func(db *sqlx.DB)) {
	dsn := fmt.Sprintf("host=%s port=%d user=postgres sslmode=disable", defaultPostgresHost, defaultPostgresPort)
	db := sqlx.MustOpen("postgres", dsn)
	dbName := fmt.Sprintf("test_%d", time.Now().UnixNano())
	db.MustExec("CREATE DATABASE " + dbName)
	t.Log("database", dbName, "successfully created")
	db2 := sqlx.MustOpen("postgres", dsn+" database="+dbName)
	cb(db2)
	require.Nil(t, db2.Close())
	if !devDontDropDatabase {
		db.MustExec("DROP DATABASE " + dbName)
	}
	require.Nil(t, db.Close())
}
