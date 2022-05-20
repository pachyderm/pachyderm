package sdata

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/testsnowflake"
)

type dbSpec interface {
	fmt.Stringer
	create(t *testing.T) (db *sqlx.DB, dbName string, tableName string)
	schema() string
	testRow() setIDer
}

var supportedDBSpecs = []dbSpec{postgreSQLSpec{}, mySQLSpec{}, snowflakeSpec{}}

type postgreSQLSpec struct{}

func (s postgreSQLSpec) String() string { return "PostgreSQL" }

func (s postgreSQLSpec) create(t *testing.T) (*sqlx.DB, string, string) {
	const tableName = "test_table"
	db, dbName := dockertestenv.NewEphemeralPostgresDB(t)
	return db, dbName, tableName
}

func (s postgreSQLSpec) schema() string { return "public" }

func (s postgreSQLSpec) testRow() setIDer {
	return &pachsql.TestRow{}
}

type mySQLSpec struct {
	dbName string
}

func (s mySQLSpec) String() string { return "MySQL" }

func (s mySQLSpec) create(t *testing.T) (*sqlx.DB, string, string) {
	const tableName = "test_table"
	var db *sqlx.DB
	db, s.dbName = dockertestenv.NewEphemeralMySQLDB(t)
	return db, s.dbName, tableName
}

func (s mySQLSpec) schema() string { return s.dbName }

func (s mySQLSpec) testRow() setIDer {
	return &pachsql.TestRow{}
}

type snowflakeSpec struct{}

func (s snowflakeSpec) String() string { return "Snowflake" }

func (s snowflakeSpec) create(t *testing.T) (*sqlx.DB, string, string) {
	const tableName = "test_table"
	db, dbName := testsnowflake.NewEphemeralSnowflakeDB(t)
	return db, dbName, tableName
}

func (s snowflakeSpec) schema() string { return "public" }

func (s snowflakeSpec) testRow() setIDer {
	return &snowflakeRow{}
}

type snowflakeRow struct {
	pachsql.TestRow
	Variant interface{} `column:"c_variant" dtype:"VARIANT" constraint:"NULL"`
}

func (row *snowflakeRow) SetID(id int16) {
	row.TestRow.SetID(id)
}
