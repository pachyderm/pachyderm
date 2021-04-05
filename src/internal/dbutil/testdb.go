package dbutil

import (
	"crypto/rand"
	"fmt"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

const (
	// DefaultHost is the default host used by NewTestDB.
	DefaultTestHost = "127.0.0.1"
	// DefaultPort is the default port used by NewTestDB.
	DefaultTestPort = 5432
	// DefaultUser is the default user used by NewTestDB.
	DefaultTestUser = "postgres"
)

// set this to true if you want to keep the database around
var devDontDropDatabase = false

const postgresMaxConnections = 100

// we want to divide the total number of connections we can have up among the
// concurrently running tests
var maxOpenConnsPerPool = postgresMaxConnections / runtime.GOMAXPROCS(0)

// NewTestDB connects to postgres using the default settings, creates a database with a unique name
// then calls cb with a sqlx.DB configured to use the newly created database.
// After cb returns the database is dropped.
func NewTestDB(t testing.TB) *sqlx.DB {
	dbName := ephemeralDBName()
	require.NoError(t, withDB(func(db *sqlx.DB) error {
		db.MustExec("CREATE DATABASE " + dbName)
		t.Log("database", dbName, "successfully created")
		return nil
	}))
	if !devDontDropDatabase {
		t.Cleanup(func() {
			require.NoError(t, withDB(func(db *sqlx.DB) error {
				db.MustExec("DROP DATABASE " + dbName)
				t.Log("database", dbName, "successfully deleted")
				return nil
			}))
		})
	}
	db2, err := NewDB(WithDBName(dbName))
	require.NoError(t, err)
	db2.SetMaxOpenConns(maxOpenConnsPerPool)
	t.Cleanup(func() {
		require.NoError(t, db2.Close())
	})
	return db2
}

// withDB creates a database connection that is scoped to the passed in callback.
func withDB(cb func(*sqlx.DB) error, opts ...Option) (retErr error) {
	opts = append([]Option{
		WithHostPort(DefaultTestHost, DefaultTestPort),
	}, opts...)
	db, err := NewDB(opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(db)
}

func ephemeralDBName() string {
	buf := [8]byte{}
	if n, err := rand.Reader.Read(buf[:]); err != nil || n < 8 {
		panic(err)
	}
	// TODO: it looks like postgres is truncating identifiers to 32 bytes,
	// it should be 64 but we might be passing the name as non-ascii, i'm not really sure.
	// for now just use a random int, but it would be nice to go back to names with a timestamp.
	return fmt.Sprintf("test_%08x", buf)
	//now := time.Now()
	// test_<date>T<time>_<random int>
	// return fmt.Sprintf("test_%04d%02d%02dT%02d%02d%02d_%04x",
	// 	now.Year(), now.Month(), now.Day(),
	// 	now.Hour(), now.Minute(), now.Second(),
	// 	rand.Uint32())
}
