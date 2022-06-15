package sdata

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// TestFormatParse is a round trip from a Tuple through formatting and parsing
// back to a Tuple again.
func TestFormatParse(t *testing.T) {
	testCases := []struct {
		Name string
		NewW func(w io.Writer, fieldNames []string) TupleWriter
		NewR func(r io.Reader, fieldNames []string) TupleReader
	}{
		{
			Name: "CSV",
			NewW: func(w io.Writer, _ []string) TupleWriter {
				return NewCSVWriter(w, nil)
			},
			NewR: func(r io.Reader, _ []string) TupleReader {
				return NewCSVParser(r)
			},
		},
		{
			Name: "JSON",
			NewW: func(w io.Writer, fieldNames []string) TupleWriter {
				return NewJSONWriter(w, fieldNames)
			},
			NewR: func(r io.Reader, fieldNames []string) TupleReader {
				return NewJSONParser(r, fieldNames)
			},
		},
	}
	newTuple := func() Tuple {
		a := int64(0)
		b := float64(0)
		c := ""
		d := sql.NullInt64{}
		e := false
		f := sql.NullString{}
		return Tuple{&a, &b, &c, &d, &e, &f}
	}
	fieldNames := []string{"a", "b", "c", "d", "e", "f"}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			const N = 10
			buf := &bytes.Buffer{}
			fz := fuzz.New()
			fz.RandSource(rand.NewSource(0))
			addFuzzFuncs(fz)

			var expected []Tuple
			w := tc.NewW(buf, fieldNames)
			for i := 0; i < N; i++ {
				x := newTuple()
				for i := range x {
					fz.Fuzz(x[i])
				}
				err := w.WriteTuple(x)
				require.NoError(t, err)
				expected = append(expected, x)
			}
			require.NoError(t, w.Flush())

			var actual []Tuple
			r := tc.NewR(buf, fieldNames)
			for i := 0; i < N; i++ {
				y := newTuple()
				err := r.Next(y)
				require.NoError(t, err)
				actual = append(actual, y)
			}
			require.Len(t, actual, len(expected))
			for i := range actual {
				require.Equal(t, expected[i], actual[i])
			}
		})
	}
}

type setIDer interface {
	SetID(int16)
}

func generateTestData(db *pachsql.DB, tableName string, n int, row setIDer) error {
	fz := fuzz.New()
	addFuzzFuncs(fz)
	var insertStatement string
	if db.DriverName() == "snowflake" {
		var process func(reflect.Type, int) []string
		process = func(t reflect.Type, acc int) []string {
			var asClauses []string
			if t.Kind() == reflect.Ptr {
				return process(t.Elem(), acc)
			}
			for i := 0; i < t.NumField(); i++ {
				var asClause string
				f := t.Field(i)
				if f.Anonymous && f.Type.Kind() == reflect.Struct {
					asClauses = append(asClauses, process(f.Type, acc+len(asClauses))...)
					continue
				}
				if f.Type.Kind() == reflect.Interface && f.Type.NumMethod() == 0 {
					asClause = fmt.Sprintf(`to_variant(COLUMN%d) as %s`, acc+len(asClauses)+1, f.Tag.Get("column"))
				} else {
					asClause = fmt.Sprintf(`COLUMN%d as %s`, acc+len(asClauses)+1, f.Tag.Get("column"))
				}
				asClauses = append(asClauses, asClause)
			}
			return asClauses
		}
		insertStatement = fmt.Sprintf(`INSERT INTO %s SELECT %s FROM VALUES %s`, tableName, strings.Join(process(reflect.TypeOf(row), 0), ","), formatValues(row, db))
	} else {
		insertStatement = fmt.Sprintf("INSERT INTO %s %s VALUES %s", tableName, formatColumns(row), formatValues(row, db))

	}
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

// TestMaterializeSQL checks that rows can be materialized from all the supported databases,
// with all the supported writers.
// It does not check that the writers themselves output in the correct format.
func TestMaterializeSQL(t *testing.T) {
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
	for _, dbSpec := range supportedDBSpecs {
		for _, writerSpec := range writerSpecs {
			testName := fmt.Sprintf("%v-%s", dbSpec, writerSpec.Name)
			t.Run(testName, func(t *testing.T) {
				db, _, tableName := dbSpec.create(t)
				require.NoError(t, pachsql.CreateTestTable(db, tableName, dbSpec.testRow()))
				nRows := 10
				if err := generateTestData(db, tableName, nRows, dbSpec.testRow()); err != nil {
					t.Fatalf("could not setup database: %v", err)
				}
				rows, err := db.Query(fmt.Sprintf(`SELECT * FROM %s`, tableName))
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

func TestSQLTupleWriter(t *testing.T) {
	for _, dbSpec := range supportedDBSpecs {
		t.Run(dbSpec.String(), func(t *testing.T) {
			var (
				ctx              = context.Background()
				db, _, tableName = dbSpec.create(t)
			)
			require.NoError(t, pachsql.CreateTestTable(db, tableName, dbSpec.testRow()))
			tableInfo, err := pachsql.GetTableInfo(ctx, db, fmt.Sprintf("%s.%s", dbSpec.schema(), tableName))

			require.NoError(t, err)

			tx, err := db.Beginx()
			require.NoError(t, err)
			defer tx.Rollback()

			// Generate fake data
			fz := fuzz.New()
			fz.RandSource(rand.NewSource(0))
			addFuzzFuncs(fz)

			tuple := newTupleFromTestRow(dbSpec.testRow())
			w := NewSQLTupleWriter(tx, tableInfo)
			nRows := 3
			for i := 0; i < nRows; i++ {
				for j := range tuple {
					fz.Fuzz(tuple[j])
				}
				// key part we are testing
				require.NoError(t, w.WriteTuple(tuple))
			}
			require.NoError(t, w.Flush())
			require.NoError(t, tx.Commit())

			// assertions
			var count int
			require.NoError(t, db.QueryRow(fmt.Sprintf("select count(*) from %s", tableName)).Scan(&count))
			require.Equal(t, nRows, count)
		})
	}
}

func TestCSVNull(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewCSVWriter(buf, nil)
	row := Tuple{
		&sql.NullString{String: "null", Valid: true},
		&sql.NullString{String: `""`, Valid: true},
		&sql.NullString{String: "", Valid: false},
		&sql.NullString{String: "", Valid: true},
	}
	expected := `null,"""""",,""` + "\n"
	w.WriteTuple(row)
	w.Flush()
	require.Equal(t, expected, buf.String())

	r := NewCSVParser(buf)
	row2 := Tuple{
		&sql.NullString{},
		&sql.NullString{},
		&sql.NullString{},
		&sql.NullString{},
	}
	err := r.Next(row2)
	require.NoError(t, err)
	require.Equal(t, row, row2)
}

func newTupleFromTestRow(row interface{}) Tuple {
	var process func(reflect.Type) Tuple
	process = func(t reflect.Type) Tuple {
		var rr Tuple
		if t.Kind() == reflect.Ptr {
			return process(t.Elem())
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				rr = append(rr, process(f.Type)...)
				continue
			}
			rr = append(rr, reflect.New(f.Type).Interface())
		}
		return rr
	}
	return process(reflect.TypeOf(row))
}
