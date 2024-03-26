package testutil

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// set this to false if you want to keep the database around
var cleanup = true

// OpenDB connects to a database using opts and returns it.
// The database will be closed at the end of the test.
func OpenDB(t testing.TB, opts ...dbutil.Option) *pachsql.DB {
	db, err := dbutil.NewDB(opts...)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}

// OpenDBURL connects to a database using u and returns it.
// The database will be closed at the end of the test.
func OpenDBURL(t testing.TB, u pachsql.URL, password string) *pachsql.DB {
	db, err := pachsql.OpenURL(u, password)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}

// CreateEphemeralDB creates a new database using db with a lifetime scoped to the test t.
func CreateEphemeralDB(t testing.TB, db *pachsql.DB, dbName string) {
	ctx := pctx.TestContext(t)
	err := CreateEphemeralDBNontest(ctx, db, dbName)
	require.NoError(t, err)
	if cleanup {
		t.Cleanup(func() {
			err := CleanupEphemeralDB(ctx, db, dbName)
			require.NoError(t, err)
		})
	}
	t.Log("database", dbName, "successfully created")
}

// CreateEphermealDBNontest creates a new database outside of a test.
func CreateEphemeralDBNontest(ctx context.Context, db *pachsql.DB, dbName string) error {
	if _, err := db.ExecContext(ctx, `CREATE DATABASE `+dbName); err != nil {
		return errors.Wrap(err, "create database")
	}
	return nil
}

// CleanupEphemeralDB cleans up a database.
func CleanupEphemeralDB(ctx context.Context, db *pachsql.DB, dbName string) error {
	q := fmt.Sprintf("DROP DATABASE %s", dbName)
	if db.DriverName() == "pgx" || db.DriverName() == "postgres" {
		q += " WITH (FORCE)"
	}
	if _, err := db.ExecContext(ctx, q); err != nil {
		return errors.Wrap(err, "drop database")
	}
	return nil
}

// GenerateEphemeralDBName generates a random name suitable for use as a Postgres identifier
func GenerateEphemeralDBName() string {
	buf := [8]byte{}
	if _, err := rand.Reader.Read(buf[:]); err != nil {
		panic(fmt.Sprintf("rand.Reader.Read: %v", err))
	}

	// TODO: it looks like postgres is truncating identifiers to 32 bytes,
	// it should be 64 but we might be passing the name as non-ascii, i'm not really sure.
	// for now just use a random int, but it would be nice to go back to names with a timestamp.
	return fmt.Sprintf("test_%08x", buf)
}
