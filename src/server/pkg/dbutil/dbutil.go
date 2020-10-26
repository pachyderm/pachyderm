package dbutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// set this to true if you want to keep the database around
var devDontDropDatabase = false

const (
	DefaultPostgresHost = "127.0.0.1"
	DefaultPostgresPort = 32228
	TestPostgresUser    = "postgres"
)

// WithTestDB connects to postgres using the default settings, creates a database with a unique name
// then calls cb with a sqlx.DB configured to use the newly created database.
// After cb returns the database is dropped.
func WithTestDB(t testing.TB, cb func(db *sqlx.DB)) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s sslmode=disable", DefaultPostgresHost, DefaultPostgresPort, TestPostgresUser)
	db := sqlx.MustOpen("postgres", dsn)
	dbName := fmt.Sprintf("test_%d", time.Now().UnixNano())
	db.MustExec("CREATE DATABASE " + dbName)
	t.Log("database", dbName, "successfully created")
	db2 := sqlx.MustOpen("postgres", dsn+" dbname="+dbName)
	cb(db2)
	require.Nil(t, db2.Close())
	if !devDontDropDatabase {
		db.MustExec("DROP DATABASE " + dbName)
	}
	require.Nil(t, db.Close())
}

type DBParams struct {
	Host       string
	Port       int
	User, Pass string
	DBName     string
}

func NewDB(x DBParams) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", x.Host, x.Port, x.User, x.Pass, x.DBName)
	return sqlx.Open("postgres", dsn)
}

// NewGORMDB creates a database client.
// TODO: Remove GOARM and switch to sql x.
func NewGORMDB(host, port, user, pass, dbName string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbName)
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// TODO Determine reasonable defaults.
	// db.LogMode(false)
	// db.DB().SetMaxOpenConns(3)
	// db.DB().SetMaxIdleConns(2)
	return db, nil
}
