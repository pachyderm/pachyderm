package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

var SupportedDBSpecs = []DBSpec{postgreSQLSpec{}, mySQLSpec{}}

type setIDer interface {
	SetID(int16)
}

type DBSpec interface {
	fmt.Stringer
	Create(ctx context.Context, t *testing.T) (db *sqlx.DB, dbName string, tableName string)
	Schema() string
	TestRow() setIDer
}

type postgreSQLSpec struct{}

func (s postgreSQLSpec) String() string { return "PostgreSQL" }

func (s postgreSQLSpec) Create(ctx context.Context, t *testing.T) (*sqlx.DB, string, string) {
	const tableName = "test_table"
	db, dbName := dockertestenv.NewEphemeralPostgresDB(ctx, t)
	return db, dbName, tableName
}

func (s postgreSQLSpec) Schema() string { return "public" }

func (s postgreSQLSpec) TestRow() setIDer {
	return &pachsql.TestRow{}
}

type mySQLSpec struct {
	dbName string
}

func (s mySQLSpec) String() string { return "MySQL" }

func (s mySQLSpec) Create(ctx context.Context, t *testing.T) (*sqlx.DB, string, string) {
	const tableName = "test_table"
	var db *sqlx.DB
	db, s.dbName = dockertestenv.NewEphemeralMySQLDB(ctx, t)
	return db, s.dbName, tableName
}

func (s mySQLSpec) Schema() string { return s.dbName }

func (s mySQLSpec) TestRow() setIDer {
	return &pachsql.TestRow{}
}

func AddFuzzFuncs(fz *fuzz.Fuzzer) {
	fz.Funcs(
		func(ti *time.Time, co fuzz.Continue) {
			*ti = time.Now()
		},
		func(x *sql.NullInt64, co fuzz.Continue) {
			if co.RandBool() {
				x.Valid = true
				x.Int64 = co.Int63()
			} else {
				x.Valid = false
			}
		},
		func(x *sql.NullString, co fuzz.Continue) {
			n := co.Intn(10)
			if n < 3 {
				x.Valid = true
				x.String = co.RandString()
			} else if n < 6 {
				x.Valid = true
				x.String = ""
			} else {
				x.Valid = false
			}
		},
		fuzz.UnicodeRange{First: '!', Last: '~'}.CustomStringFuzzFunc(),
		func(x *interface{}, co fuzz.Continue) {
			switch co.Intn(6) {
			case 0:
				*x = int16(co.Int())
			case 1:
				*x = int32(co.Int())
			case 2:
				*x = int64(co.Int())
			case 3:
				*x = co.Float32()
			case 4:
				*x = co.Float64()
			case 5:
				*x = co.RandString()
			}
		},
	)
}

func GenerateTestData(db *pachsql.DB, tableName string, n int, row setIDer) error {
	fz := fuzz.New()
	AddFuzzFuncs(fz)
	insertStatement := fmt.Sprintf("INSERT INTO %s %s VALUES %s", tableName, formatColumns(row), formatValues(row, db))
	for i := 0; i < n; i++ {
		fz.Fuzz(row)
		row.SetID(int16(i))
		if _, err := db.Exec(insertStatement, makeArgs(row)...); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

func formatColumns(x interface{}) string {
	var process func(reflect.Type) []string
	process = func(t reflect.Type) []string {
		var cols []string
		if t.Kind() == reflect.Ptr {
			return process(t.Elem())
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				cols = append(cols, process(field.Type)...)
				continue
			}
			col := field.Tag.Get("column")
			cols = append(cols, col)
		}
		return cols
	}
	return "(" + strings.Join(process(reflect.TypeOf(x)), ", ") + ")"
}

func formatValues(x interface{}, db *pachsql.DB) string {
	var process func(reflect.Type) []string
	process = func(t reflect.Type) []string {
		var cols []string
		if t.Kind() == reflect.Ptr {
			return process(t.Elem())
		}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				cols = append(cols, process(field.Type)...)
				continue
			}
			cols = append(cols, pachsql.Placeholder(db.DriverName(), i))
		}
		return cols
	}
	return fmt.Sprintf("(%s)", strings.Join(process(reflect.TypeOf(x)), ", "))
}

func makeArgs(x interface{}) []interface{} {
	var process func(reflect.Value) []interface{}
	process = func(v reflect.Value) []interface{} {
		var vals []interface{}
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				vals = append(vals, process(v.Field(i))...)
				continue
			}
			vals = append(vals, v.Field(i).Interface())
		}
		return vals
	}
	return process(reflect.ValueOf(x))
}
