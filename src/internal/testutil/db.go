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

const (
	DefaultPostgresHost     = "127.0.0.1"
	DefaultPostgresPort     = 30228
	DefaultPGBouncerHost    = DefaultPostgresHost
	DefaultPGBouncerPort    = 30229
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

// TestDatabaseDeployment represents a deployment of postgres, and databases may
// be created for individual tests.
type TestDatabaseDeployment interface {
	NewDatabase(t testing.TB) (*sqlx.DB, col.PostgresListener)
	NewDatabaseConfig(t testing.TB) serviceenv.ConfigOption
}

func postgresHost() string {
	if host, ok := os.LookupEnv("POSTGRES_HOST"); ok {
		return host
	}
	return DefaultPostgresHost
}

func postgresPort() int {
	if port, ok := os.LookupEnv("POSTGRES_PORT"); ok {
		if portInt, err := strconv.Atoi(port); err == nil {
			return portInt
		}
	}
	return DefaultPostgresPort
}

func pgBouncerHost() string {
	if host, ok := os.LookupEnv("PG_BOUNCER_HOST"); ok {
		return host
	}
	return DefaultPGBouncerHost
}

func pgBouncerPort() int {
	if port, ok := os.LookupEnv("PG_BOUNCER_PORT"); ok {
		if portInt, err := strconv.Atoi(port); err == nil {
			return portInt
		}
	}
	return DefaultPGBouncerPort
}

// NewTestDBConfig creates an ephemeral database scoped to the life of the test, without connecting to it.
// It returns a serviceenv.ConfigOption which can be used to configure the environment to connect directly, and indirectly to the database.
func NewTestDBConfig(t testing.TB) serviceenv.ConfigOption {
	db := openEphemeralDB(t,
		dbutil.WithDBName(DefaultPostgresDatabase),
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(DefaultPostgresUser, DefaultPostgresPassword),
		dbutil.WithHostPort(postgresHost(), postgresPort()),
	)
	dbName := createEphemeralDB(t, db)
	return func(c *serviceenv.Configuration) {
		// common
		c.PostgresDBName = dbName

		// direct
		c.PostgresHost = postgresHost()
		c.PostgresPort = postgresPort()
		// pg_bouncer
		c.PGBouncerHost = pgBouncerHost()
		c.PGBouncerPort = pgBouncerPort()

		c.PostgresUser = DefaultPostgresUser
	}
}

func NewTestDirectDBOptions(t testing.TB) []dbutil.Option {
	host, port := postgresHost(), postgresPort()
	opts := []dbutil.Option{
		dbutil.WithHostPort(host, port),
		dbutil.WithUserPassword(DefaultPostgresUser, DefaultPostgresPassword),
		dbutil.WithMaxOpenConns(1),
		dbutil.WithDBName(DefaultPostgresDatabase),
	}
	db := openEphemeralDB(t, opts...)
	dbName := createEphemeralDB(t, db)

	return []dbutil.Option{
		dbutil.WithUserPassword(DefaultPostgresUser, DefaultPostgresPassword),
		dbutil.WithDBName(dbName),
		dbutil.WithMaxOpenConns(maxOpenConnsPerPool),

		dbutil.WithHostPort(host, port),
	}
}

// NewTestDBOptions creates an ephemeral db and returns options that can be used to connect to it.
func NewTestDBOptions(t testing.TB) []dbutil.Option {
	host, port := pgBouncerHost(), pgBouncerPort()
	opts := []dbutil.Option{
		dbutil.WithHostPort(host, port),
		dbutil.WithUserPassword(DefaultPostgresUser, DefaultPostgresPassword),
		dbutil.WithMaxOpenConns(1),
		dbutil.WithDBName(DefaultPostgresDatabase),
	}
	db := openEphemeralDB(t, opts...)
	dbName := createEphemeralDB(t, db)

	return []dbutil.Option{
		dbutil.WithUserPassword(DefaultPostgresUser, DefaultPostgresPassword),
		dbutil.WithDBName(dbName),
		dbutil.WithMaxOpenConns(maxOpenConnsPerPool),

		dbutil.WithHostPort(host, port),
	}
}

// NewTestDB connects to postgres using the default settings, creates a database
// with a unique name then returns a sqlx.DB configured to use the newly created
// database. After the test or suite finishes, the database is dropped.
func NewTestDB(t testing.TB) *sqlx.DB {
	return openEphemeralDB(t, NewTestDBOptions(t)...)
}

// openEphemeralDB connects to a database using opts and returns it.
// the database will be cleaned up at the end of the test.
func openEphemeralDB(t testing.TB, opts ...dbutil.Option) *sqlx.DB {
	db, err := dbutil.NewDB(opts...)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}

// createEphemeralDB creates a new database using db with a lifetime scoped to the test t
// and returns its name
func createEphemeralDB(t testing.TB, db *sqlx.DB) string {
	dbName := ephemeralDBName(t)
	_, err := db.Exec(`CREATE DATABASE ` + dbName)
	require.NoError(t, err)
	if cleanup {
		t.Cleanup(func() {
			_, err := db.Exec(fmt.Sprintf(`DROP DATABASE %s WITH (FORCE)`, dbName))
			require.NoError(t, err)
		})
	}
	t.Log("database", dbName, "successfully created")
	return dbName
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
	// now := time.Now()
	// test_<date>T<time>_<random int>
	// return fmt.Sprintf("test_%04d%02d%02dT%02d%02d%02d_%04x",
	// 	now.Year(), now.Month(), now.Day(),
	// 	now.Hour(), now.Minute(), now.Second(),
	// 	rand.Uint32())
}
