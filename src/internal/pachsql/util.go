package pachsql

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	fuzz "github.com/google/gofuzz"
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
	case "mysql":
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
		case "pgx":
			schemaName = "public"
		default:
			panic(driver)
		}
	}
	return schemaName, tableName
}

// TestRow is the type of a row in the test table
type TestRow struct {
	Id int

	Smallint int16
	Int      int32
	Bigint   int64
	Float    float32
	Varchar  string
	Time     time.Time

	SmallintNull sql.NullInt16
	IntNull      sql.NullInt32
	BigintNull   sql.NullInt64
	FloatNull    sql.NullFloat64
	VarcharNull  sql.NullString
	TimeNull     sql.NullTime
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
	return err
}

func LoadTestData(db *DB, tableName string, n int) error {
	fz := fuzz.New()
	fz.Funcs(func(ti *time.Time, co fuzz.Continue) {
		*ti = time.Now()
	})
	fz.Funcs(fuzz.UnicodeRange{First: '!', Last: '~'}.CustomStringFuzzFunc())
	for i := 0; i < n; i++ {
		var x TestRow
		x.Time = time.Now()
		if i > 0 {
			fz.Fuzz(&x)
		}
		x.Id = i
		insertStatement := `INSERT INTO test_data ` + formatColumns(x) + ` VALUES ` + formatValues(x, db)
		if _, err := db.Exec(insertStatement, makeArgs(x)...); err != nil {
			return err
		}
	}
	return nil
}

func formatColumns(x interface{}) string {
	var cols []string
	rty := reflect.TypeOf(x)
	for i := 0; i < rty.NumField(); i++ {
		field := rty.Field(i)
		col := "c_" + toSnakeCase(field.Name)
		cols = append(cols, col)
	}
	return "(" + strings.Join(cols, ", ") + ")"
}

func formatValues(x interface{}, db *DB) string {
	var placeholder func(i int) string
	switch db.DriverName() {
	case "pgx":
		placeholder = func(i int) string { return "$" + strconv.Itoa(i+1) }
	case "mysql", "snowflake":
		placeholder = func(int) string { return "?" }
	default:
		panic(db.DriverName())
	}
	var ret string
	for i := 0; i < reflect.TypeOf(x).NumField(); i++ {
		if i > 0 {
			ret += ", "
		}
		ret += placeholder(i)
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

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
