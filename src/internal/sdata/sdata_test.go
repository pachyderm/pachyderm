package sdata

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testsnowflake"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
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
			)

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
			dockertestenv.NewPostgres,
		},
		{
			"MySQL",
			dockertestenv.NewMySQL,
		},
		{
			"Snowflake",
			testsnowflake.NewSnowSQL,
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
				tableName := "test_table"
				nRows := 10
				setupTable(t, db, tableName, nRows)
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

func TestSQLTupleWriter(suite *testing.T) {
	testcases := []struct {
		Name  string
		NewDB func(testing.TB, string) *sqlx.DB
	}{
		{
			"Postgres",
			dockertestenv.NewEphemeralPostgresDB,
		},
		{
			"MySQL",
			dockertestenv.NewEphemeralMySQLDB,
		},
		{
			"Snowflake",
			testsnowflake.NewEphemeralSnowflakeDB,
		},
	}
	for _, tc := range testcases {
		suite.Run(tc.Name, func(t *testing.T) {
			dbName := testutil.GenerateEphermeralDBName(t)
			tableName := "test_table"
			db := tc.NewDB(t, dbName)
			require.NoError(t, pachsql.CreateTestTable(db, tableName, pachsql.TestRow{}))

			ctx := context.Background()
			schema := "public"
			if tc.Name == "MySQL" {
				schema = dbName
			}
			tableInfo, err := pachsql.GetTableInfo(ctx, db, fmt.Sprintf("%s.%s", schema, tableName))
			require.NoError(t, err)

			tx, err := db.Beginx()
			require.NoError(t, err)
			defer tx.Rollback()

			// Generate fake data
			fz := fuzz.New()
			fz.RandSource(rand.NewSource(0))
			fz.Funcs(func(ti *time.Time, co fuzz.Continue) {
				// for mysql compatibility
				*ti = time.Now()
			})

			tuple := newTupleFromTestRow(pachsql.TestRow{})
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

func setupTable(t testing.TB, db *pachsql.DB, name string, n int) {
	require.NoError(t, pachsql.CreateTestTable(db, name, pachsql.TestRow{}))
	require.NoError(t, pachsql.GenerateTestData(db, name, n))
}

func newTupleFromTestRow(row interface{}) Tuple {
	result := Tuple{}
	rval := reflect.TypeOf(row)
	for i := 0; i < rval.NumField(); i++ {
		v := reflect.New(rval.Field(i).Type).Interface()
		result = append(result, v)
	}
	return result
}
