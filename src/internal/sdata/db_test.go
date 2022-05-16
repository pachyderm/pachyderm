package sdata

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testsnowflake"
)

type dbSpec interface {
	fmt.Stringer
	// create creates an ephemeral database and a test table, returning a
	// connection and the names.
	create(t *testing.T) (db *sqlx.DB, dbName string, tableName string)
	setup(db *sqlx.DB, tableName string, rows int) error
	schema() string
}

var supportedDBSpecs = []dbSpec{postgreSQLSpec{}, mySQLSpec{}, snowflakeSpec{}}

type postgreSQLSpec struct{}

func (s postgreSQLSpec) String() string { return "PostgreSQL" }

func (s postgreSQLSpec) create(t *testing.T) (*sqlx.DB, string, string) {
	const tableName = "test_table"
	db, dbName := dockertestenv.NewEphemeralPostgresDB(t)
	require.NoError(t, pachsql.CreateTestTable(db, tableName, pachsql.TestRow{}))

	return db, dbName, tableName
}

func (s postgreSQLSpec) setup(db *sqlx.DB, tableName string, rows int) error {
	return pachsql.GenerateTestData(db, tableName, rows)
}

func (s postgreSQLSpec) schema() string { return "public" }

type mySQLSpec struct {
	dbName string
}

func (s mySQLSpec) String() string { return "MySQL" }

func (s mySQLSpec) create(t *testing.T) (*sqlx.DB, string, string) {
	const tableName = "test_table"
	var db *sqlx.DB
	db, s.dbName = dockertestenv.NewEphemeralMySQLDB(t)
	require.NoError(t, pachsql.CreateTestTable(db, tableName, pachsql.TestRow{}))
	return db, s.dbName, tableName
}

func (s mySQLSpec) setup(db *sqlx.DB, tableName string, rows int) error {
	return pachsql.GenerateTestData(db, tableName, rows)
}

func (s mySQLSpec) schema() string { return s.dbName }

type snowflakeSpec struct{}

func (s snowflakeSpec) String() string { return "Snowflake" }

func (s snowflakeSpec) create(t *testing.T) (*sqlx.DB, string, string) {
	const tableName = "test_table"
	db, dbName := testsnowflake.NewEphemeralSnowflakeDB(t)
	require.NoError(t, pachsql.CreateTestTable(db, tableName, struct{ pachsql.TestRow }{}))
	return db, dbName, tableName
}

func (s snowflakeSpec) setup(db *sqlx.DB, tableName string, rows int) error {
	return pachsql.GenerateTestData(db, tableName, rows)
}

func (s snowflakeSpec) schema() string { return "public" }
