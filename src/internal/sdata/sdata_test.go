package sdata

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// TestMaterializeSQL checks that rows can be materialized from all the supported databases,
// with all the supported writers.
// It does not check that the writers themselves output in the correct format.
func TestMaterializeSQL(t *testing.T) {
	dbSpecs := []struct {
		Name string
		New  func(t testing.TB) *sqlx.DB
	}{
		{
			"Postgres",
			dockertestenv.NewTestDB,
		},
		{
			"MySQL",
			dockertestenv.NewMySQL,
		},
	}
	writerSpecs := []struct {
		Name string
		New  func(io.Writer, []string) TupleWriter
	}{
		{
			"JSON",
			func(w io.Writer, names []string) TupleWriter {
				return NewJSONWriter(w, names)
			},
		},
		{
			"CSV",
			func(w io.Writer, names []string) TupleWriter {
				return NewCSVWriter(w, names)
			},
		},
	}
	for _, dbSpec := range dbSpecs {
		for _, writerSpec := range writerSpecs {
			testName := fmt.Sprintf("%s-%s", dbSpec.Name, writerSpec.Name)
			t.Run(testName, func(t *testing.T) {
				db := dbSpec.New(t)
				setupTable(t, db)
				rows, err := db.Query(`SELECT * FROM test_data`)
				require.NoError(t, err)
				defer rows.Close()
				buf := &bytes.Buffer{}
				colNames, err := rows.Columns()
				require.NoError(t, err)
				w := writerSpec.New(buf, colNames)
				_, err = MaterializeSQL(w, rows)
				require.NoError(t, err)
				t.Log(buf.String())
			})
		}
	}
}

func setupTable(t testing.TB, db *pachsql.DB) {
	type rowType struct {
		// id int

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
	_, err := db.Exec(`CREATE TABLE test_data (
		id SERIAL PRIMARY KEY NOT NULL,

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
	)`)
	require.NoError(t, err)
	const N = 10
	fz := fuzz.New()
	fz.Funcs(func(ti *time.Time, co fuzz.Continue) {
		*ti = time.Now()
	})
	fz.Funcs(fuzz.UnicodeRange{First: '!', Last: '~'}.CustomStringFuzzFunc())
	for i := 0; i < N; i++ {
		var x rowType
		x.Time = time.Now()
		if i > 0 {
			fz.Fuzz(&x)
		}
		_, err = db.Exec(`INSERT INTO test_data `+formatColumns(x)+` VALUES `+formatValues(x, db), makeArgs(x)...)
		require.NoError(t, err)
	}
}

func formatColumns(x interface{}) string {
	var cols []string
	rty := reflect.TypeOf(x)
	for i := 0; i < rty.NumField(); i++ {
		field := rty.Field(i)
		if field.Name == "id" {
			continue
		}
		col := "c_" + toSnakeCase(field.Name)
		cols = append(cols, col)
	}
	return "(" + strings.Join(cols, ", ") + ")"
}

func formatValues(x interface{}, db *pachsql.DB) string {
	var placeholder func(i int) string
	switch db.DriverName() {
	case "pgx":
		placeholder = func(i int) string { return "$" + strconv.Itoa(i+1) }
	case "mysql":
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
	rty := reflect.TypeOf(x)
	for i := 0; i < rval.NumField(); i++ {
		if rty.Field(i).Name == "id" {
			continue
		}
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
