package sdata

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
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
			t.Run(fmt.Sprintf("%s-%s", dbSpec.Name, writerSpec.Name), func(t *testing.T) {
				db := dbSpec.New(t)
				_, err := db.Exec(testutil.SeedCarsTable)
				require.NoError(t, err)
				rows, err := db.Query(`SELECT * FROM public.cars`)
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
