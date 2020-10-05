package dbutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

const (
	defaultPostgresHost = "localhost"
	defaultPostgresPort = 32228
)

func WithTestDB(t *testing.T, fn func(db *sqlx.DB)) {
	dsn := fmt.Sprintf("host=%s port=%d user=postgres sslmode=disable", defaultPostgresHost, defaultPostgresPort)
	db := sqlx.MustOpen("postgres", dsn)
	dbName := fmt.Sprintf("test_%d", time.Now().UnixNano())
	db.MustExec("CREATE DATABASE " + dbName)
	db2 := sqlx.MustOpen("postgres", dsn+" database="+dbName)
	fn(db2)
	require.Nil(t, db2.Close())
	db.MustExec("DROP DATABASE " + dbName)
	require.Nil(t, db.Close())
}
