package dbutil

import (
	"crypto/rand"
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
