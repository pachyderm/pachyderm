package pachsql_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testsnowflake"
)

func TestGetTableInfo(t *testing.T) {
	type testCase struct {
		Name     string
		NewDB    func(t testing.TB) *pachsql.DB
		Expected *pachsql.TableInfo
	}
	tcs := []testCase{
		{
			Name:  "Postgres",
			NewDB: dockertestenv.NewPostgres,
			Expected: &pachsql.TableInfo{
				"test_table",
				"public",
				[]pachsql.ColumnInfo{
					{"c_id", "integer", false},
					{"c_smallint", "smallint", false},
					{"c_int", "integer", false},
					{"c_bigint", "bigint", false},
					{"c_float", "double precision", false},
					{"c_varchar", "character varying", false},
					{"c_time", "timestamp without time zone", false},
					{"c_smallint_null", "smallint", true},
					{"c_int_null", "integer", true},
					{"c_bigint_null", "bigint", true},
					{"c_float_null", "double precision", true},
					{"c_varchar_null", "character varying", true},
					{"c_time_null", "timestamp without time zone", true},
				},
			},
		},
		{
			Name:  "MySQL",
			NewDB: dockertestenv.NewMySQL,
			Expected: &pachsql.TableInfo{
				"test_table",
				"",
				[]pachsql.ColumnInfo{
					{"c_id", "int", false},
					{"c_smallint", "smallint", false},
					{"c_int", "int", false},
					{"c_bigint", "bigint", false},
					{"c_float", "float", false},
					{"c_varchar", "varchar", false},
					{"c_time", "timestamp", false},
					{"c_smallint_null", "smallint", true},
					{"c_int_null", "int", true},
					{"c_bigint_null", "bigint", true},
					{"c_float_null", "float", true},
					{"c_varchar_null", "varchar", true},
					{"c_time_null", "timestamp", true},
				},
			},
		},
		{
			Name:  "Snowflake",
			NewDB: testsnowflake.NewSnowSQL,
			Expected: &pachsql.TableInfo{
				"test_table",
				"public",
				[]pachsql.ColumnInfo{
					{"C_ID", "NUMBER", false},
					{"C_SMALLINT", "NUMBER", false},
					{"C_INT", "NUMBER", false},
					{"C_BIGINT", "NUMBER", false},
					{"C_FLOAT", "FLOAT", false},
					{"C_VARCHAR", "TEXT", false},
					{"C_TIME", "TIMESTAMP_NTZ", false},
					{"C_SMALLINT_NULL", "NUMBER", true},
					{"C_INT_NULL", "NUMBER", true},
					{"C_BIGINT_NULL", "NUMBER", true},
					{"C_FLOAT_NULL", "FLOAT", true},
					{"C_VARCHAR_NULL", "TEXT", true},
					{"C_TIME_NULL", "TIMESTAMP_NTZ", true},
				},
			},
		},
	}
	ctx := context.Background()
	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			db := tc.NewDB(t)
			require.NoError(t, pachsql.CreateTestTable(db, "test_table"))
			info, err := pachsql.GetTableInfo(ctx, db, "test_table")
			require.NoError(t, err)
			require.Len(t, info.Columns, reflect.TypeOf(pachsql.TestRow{}).NumField())
			require.Equal(t, tc.Expected, info)
		})
	}
}
