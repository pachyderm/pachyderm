package pachsql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	fuzz "github.com/google/gofuzz"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// Placeholder returns a placeholder for the given driver
// assuming i placeholders have been provided before it.  It is 0 indexed in this way.
//
// This means that the first Postgres placeholder is `$1` when i = 0.
// This is more ergonimic for contructing a list of arguments since i = len(args)
// but is perhaps unintuitive for those familiar with Postgres.
func Placeholder(driverName string, i int) string {
	switch driverName {
	case "pgx":
		return "$" + strconv.Itoa(i+1)
	case "mysql", "snowflake":
		return "?"
	default:
		panic(driverName)
	}
}

// SplitTableSchema splits the tablePath on the first . and interprets the first part
// as the schema, if the driver supports schemas.
func SplitTableSchema(driver string, tablePath string) (schemaName string, tableName string) {
	if parts := strings.SplitN(tablePath, ".", 2); len(parts) == 2 {
		schemaName = parts[0]
		tableName = parts[1]
	} else {
		tableName = tablePath
		switch driver {
		case "mysql":
		case "pgx", "snowflake":
			schemaName = "public"
		default:
			panic(fmt.Sprintf("driver not supported: %s", driver))
		}
	}
	return schemaName, tableName
}

// TestRow is the type of a row in the test table
// struct tag: sql:"<column_name>,<data_type>,<table_constraint>"
type TestRow struct {
	Id int `sql:"c_id,INT PRIMARY KEY,NOT NULL"`

	Smallint int16     `sql:"c_smallint,SMALLINT,NOT NULL"`
	Int      int32     `sql:"c_int,INT,NOT NULL"`
	Bigint   int64     `sql:"c_bigint,BIGINT,NOT NULL"`
	Float    float32   `sql:"c_float,FLOAT,NOT NULL"`
	Varchar  string    `sql:"c_varchar,VARCHAR(100),NOT NULL"`
	Time     time.Time `sql:"c_time,TIMESTAMP,NOT NULL"`

	SmallintNull sql.NullInt16   `sql:"c_smallint_null,SMALLINT,NULL"`
	IntNull      sql.NullInt32   `sql:"c_int_null,INT,NULL"`
	BigintNull   sql.NullInt64   `sql:"c_bigint_null,BIGINT,NULL"`
	FloatNull    sql.NullFloat64 `sql:"c_float_null,FLOAT,NULL"`
	VarcharNull  sql.NullString  `sql:"c_varchar_null,VARCHAR(100),NULL"`
	TimeNull     sql.NullTime    `sql:"c_time_null,TIMESTAMP,NULL"`
}

// CreateTestTable creates a test table at name in the database
func CreateTestTable(db *DB, name string) error {
	t := reflect.TypeOf(TestRow{})
	cols := []string{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		cols = append(cols, strings.Join(strings.Split(field.Tag.Get("sql"), ","), " "))
	}
	q := fmt.Sprintf(`CREATE TABLE %s (%s)`, name, strings.Join(cols, ", "))
	_, err := db.Exec(q)
	return errors.EnsureStack(err)
}

func GenerateTestData(db *DB, tableName string, n int) error {
	fz := fuzz.New()
	// support mysql
	fz.Funcs(func(ti *time.Time, co fuzz.Continue) {
		*ti = time.Now()
	})
	fz.Funcs(fuzz.UnicodeRange{First: '!', Last: '~'}.CustomStringFuzzFunc())
	var row TestRow
	insertStatement := fmt.Sprintf("INSERT INTO test_data %s VALUES %s", formatColumns(row), formatValues(row, db))
	for i := 0; i < n; i++ {
		fz.Fuzz(&row)
		row.Id = i
		row.Time = time.Now() // support mysql
		if _, err := db.Exec(insertStatement, makeArgs(row)...); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

func formatColumns(x interface{}) string {
	var cols []string
	rty := reflect.TypeOf(x)
	for i := 0; i < rty.NumField(); i++ {
		field := rty.Field(i)
		col := strings.Split(field.Tag.Get("sql"), ",")[0]
		cols = append(cols, col)
	}
	return "(" + strings.Join(cols, ", ") + ")"
}

func formatValues(x interface{}, db *DB) string {
	var ret string
	for i := 0; i < reflect.TypeOf(x).NumField(); i++ {
		if i > 0 {
			ret += ", "
		}
		ret += Placeholder(db.DriverName(), i)
	}
	return "(" + ret + ")"
}

func makeArgs(x interface{}) []interface{} {
	var vals []interface{}
	rval := reflect.ValueOf(x)
	for i := 0; i < rval.NumField(); i++ {
		v := rval.Field(i).Interface()
		vals = append(vals, v)
	}
	return vals
}
