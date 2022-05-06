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
	Id int16 `column:"c_id" dtype:"SMALLINT" constraint:"PRIMARY KEY NOT NULL"`

	Smallint     int16     `column:"c_smallint" dtype:"SMALLINT" constraint:"NOT NULL"`
	Int          int32     `column:"c_int" dtype:"INT" constraint:"NOT NULL"`
	Bigint       int64     `column:"c_bigint" dtype:"BIGINT" constraint:"NOT NULL"`
	Float        float32   `column:"c_float" dtype:"FLOAT" constraint:"NOT NULL"`
	NumericInt   int64     `column:"c_numeric_int" dtype:"NUMERIC(20,0)" constraint:"NOT NULL"`
	NumericFloat float64   `column:"c_numeric_float" dtype:"NUMERIC(20,19)" constraint:"NOT NULL"`
	Varchar      string    `column:"c_varchar" dtype:"VARCHAR(100)" constraint:"NOT NULL"`
	Time         time.Time `column:"c_time" dtype:"TIMESTAMP" constraint:"NOT NULL"`

	SmallintNull     sql.NullInt16   `column:"c_smallint_null" dtype:"SMALLINT" constraint:"NULL"`
	IntNull          sql.NullInt32   `column:"c_int_null" dtype:"INT" constraint:"NULL"`
	BigintNull       sql.NullInt64   `column:"c_bigint_null" dtype:"BIGINT" constraint:"NULL"`
	FloatNull        sql.NullFloat64 `column:"c_float_null" dtype:"FLOAT" constraint:"NULL"`
	NumericIntNull   sql.NullInt64   `column:"c_numeric_int_null" dtype:"NUMERIC(20,0)" constraint:"NULL"`
	NumericFloatNull sql.NullFloat64 `column:"c_numeric_float_null" dtype:"NUMERIC(20,19)" constraint:"NULL"`
	VarcharNull      sql.NullString  `column:"c_varchar_null" dtype:"VARCHAR(100)" constraint:"NULL"`
	TimeNull         sql.NullTime    `column:"c_time_null" dtype:"TIMESTAMP" constraint:"NULL"`
}

// CreateTestTable creates a test table at name in the database
func CreateTestTable(db *DB, name string, schema interface{}) error {
	t := reflect.TypeOf(schema)
	cols := []string{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		cols = append(cols, fmt.Sprintf("%s %s %s", field.Tag.Get("column"), field.Tag.Get("dtype"), field.Tag.Get("constraint")))
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
	insertStatement := fmt.Sprintf("INSERT INTO %s %s VALUES %s", tableName, formatColumns(row), formatValues(row, db))
	for i := 0; i < n; i++ {
		fz.Fuzz(&row)
		row.Id = int16(i)
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
		col := field.Tag.Get("column")
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
