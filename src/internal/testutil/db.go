package testutil

import (
	"crypto/rand"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

// set this to false if you want to keep the database around
var cleanup = true

const postgresMaxConnections = 100

// we want to divide the total number of connections we can have up among the
// concurrently running tests
var maxOpenConnsPerPool = postgresMaxConnections / runtime.GOMAXPROCS(0)

// TestDatabaseDeployment represents a deployment of postgres, and databases may
// be created for individual tests.
type TestDatabaseDeployment interface {
	NewDatabase(t testing.TB) (*sqlx.DB, *col.PostgresListener)
	NewDatabaseConfig(t testing.TB) serviceenv.ConfigOption
}

func newDatabase(t testing.TB) string {
	dbName := ephemeralDBName(t)
	require.NoError(t, withDB(func(db *sqlx.DB) error {
		_, err := db.Exec("CREATE DATABASE " + dbName)
		require.NoError(t, err)
		t.Log("database", dbName, "successfully created")
		return nil
	}, dbutil.WithHostPort(dbHost(), dbPort())))
	if cleanup {
		t.Cleanup(func() {
			require.NoError(t, withDB(func(db *sqlx.DB) error {
				_, err := db.Exec("DROP DATABASE " + dbName)
				require.NoError(t, err)
				t.Log("database", dbName, "successfully deleted")
				return nil
			}, dbutil.WithHostPort(dbHost(), dbPort())))
		})
	}
	return dbName
}

func dbHost() string {
	if host, ok := os.LookupEnv("POSTGRES_HOST"); ok {
		return host
	}
	return dbutil.DefaultHost
}

func dbPort() int {
	if port, ok := os.LookupEnv("POSTGRES_PORT"); ok {
		if portInt, err := strconv.Atoi(port); err == nil {
			return portInt
		}
	}
	return dbutil.DefaultPort
}

// NewTestDB connects to postgres using the default settings, creates a database
// with a unique name then returns a sqlx.DB configured to use the newly created
// database. After the test or suite finishes, the database is dropped.
func NewTestDB(t testing.TB) *sqlx.DB {
	db2, err := dbutil.NewDB(dbutil.WithHostPort(dbHost(), dbPort()), dbutil.WithDBName(newDatabase(t)))
	require.NoError(t, err)
	db2.SetMaxOpenConns(maxOpenConnsPerPool)
	t.Cleanup(func() {
		require.NoError(t, db2.Close())
	})
	return db2
}

// NewTestDBConfig connects to postgres using the default settings, creates a
// database with a unique name then returns a ServiceEnv config option to
// connect to the new database. After test test or suite finishes, the database
// is dropped.
func NewTestDBConfig(t testing.TB) serviceenv.ConfigOption {
	dbName := newDatabase(t)
	return func(config *serviceenv.Configuration) {
		serviceenv.WithPostgresHostPort(dbHost(), dbPort())(config)
		config.PostgresDBName = dbName
	}
}

// withDB creates a database connection that is scoped to the passed in callback.
func withDB(cb func(*sqlx.DB) error, opts ...dbutil.Option) (retErr error) {
	db, err := dbutil.NewDB(opts...)
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

func ephemeralDBName(t testing.TB) string {
	buf := [8]byte{}
	n, err := rand.Reader.Read(buf[:])
	require.NoError(t, err)
	require.Equal(t, n, 8)

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
