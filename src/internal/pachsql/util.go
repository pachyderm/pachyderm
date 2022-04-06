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
type TestRow struct {
	Id int `sql:"c_id"`

	Smallint int16     `sql:"c_smallint"`
	Int      int32     `sql:"c_int"`
	Bigint   int64     `sql:"c_bigint"`
	Float    float32   `sql:"c_float"`
	Varchar  string    `sql:"c_varchar"`
	Time     time.Time `sql:"c_time"`

	SmallintNull sql.NullInt16   `sql:"c_smallint_null"`
	IntNull      sql.NullInt32   `sql:"c_int_null"`
	BigintNull   sql.NullInt64   `sql:"c_bigint_null"`
	FloatNull    sql.NullFloat64 `sql:"c_float_null"`
	VarcharNull  sql.NullString  `sql:"c_varchar_null"`
	TimeNull     sql.NullTime    `sql:"c_time_null"`
}

// CreateTestTable creates a test table at name in the database
func CreateTestTable(db *DB, name string) error {
	q := fmt.Sprintf(`CREATE TABLE %s (
		c_id INT PRIMARY KEY NOT NULL,

		c_smallint SMALLINT NOT NULL,
		c_int INT NOT NULL,
		c_bigint BIGINT NOT NULL,
		c_float FLOAT NOT NULL,
		c_varchar VARCHAR(100) NOT NULL,
		c_time TIMESTAMP NOT NULL,

		c_smallint_null SMALLINT NULL,
		c_int_null INT NULL,
		c_bigint_null BIGINT NULL,
		c_float_null FLOAT NULL,
		c_varchar_null VARCHAR(100) NULL,
		c_time_null TIMESTAMP NULL
	)`, name)
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
		col := field.Tag.Get("sql")
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
