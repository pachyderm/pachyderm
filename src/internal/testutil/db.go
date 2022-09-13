package testutil

import (
	"crypto/rand"
	"fmt"
	"runtime"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

const (
	DefaultPostgresHost     = "127.0.0.1"
	DefaultPostgresUser     = "pachyderm"
	DefaultPostgresPassword = "correcthorsebatterystaple"
	DefaultPostgresDatabase = "pachyderm"
)

// set this to false if you want to keep the database around
var cleanup = true

const postgresMaxConnections = 100

// we want to divide the total number of connections we can have up among the
// concurrently running tests
var maxOpenConnsPerPool = (postgresMaxConnections - 1) / runtime.GOMAXPROCS(0)

// NewTestDBOptions connects to postgres using opts, creates a database
// with a unique name then returns options to connect to the new database
// After t finishes, the database is dropped.
func NewTestDBOptions(t testing.TB, opts []dbutil.Option) []dbutil.Option {
	db := OpenDB(t, opts...)
	dbName := GenerateEphemeralDBName(t)
	CreateEphemeralDB(t, db, dbName)
	opts2 := []dbutil.Option{
		dbutil.WithMaxOpenConns(maxOpenConnsPerPool),
	}
	opts2 = append(opts2, opts...)
	opts2 = append(opts2, dbutil.WithDBName(dbName))
	return opts2
}

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
	_, err := db.Exec(`CREATE DATABASE ` + dbName)
	require.NoError(t, err)
	if cleanup {
		t.Cleanup(func() {
			q := fmt.Sprintf("DROP DATABASE %s", dbName)
			if db.DriverName() == "pgx" || db.DriverName() == "postgres" {
				q += " WITH (FORCE)"
			}
			_, err := db.Exec(q)
			require.NoError(t, err)
		})
	}
	t.Log("database", dbName, "successfully created")
}

func GenerateEphemeralDBName(t testing.TB) string {
	buf := [8]byte{}
	n, err := rand.Reader.Read(buf[:])
	require.NoError(t, err)
	require.Equal(t, n, 8)

	// TODO: it looks like postgres is truncating identifiers to 32 bytes,
	// it should be 64 but we might be passing the name as non-ascii, i'm not really sure.
	// for now just use a random int, but it would be nice to go back to names with a timestamp.
	return fmt.Sprintf("test_%08x", buf)
	// now := time.Now()
	// test_<date>T<time>_<random int>
	// return fmt.Sprintf("test_%04d%02d%02dT%02d%02d%02d_%04x",
	// 	now.Year(), now.Month(), now.Day(),
	// 	now.Hour(), now.Minute(), now.Second(),
	// 	rand.Uint32())
}
