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
		NewTable func(*pachsql.DB) error
		Expected *pachsql.TableInfo
	}
	tcs := []testCase{
		{
			Name:  "Postgres",
			NewDB: dockertestenv.NewPostgres,
			NewTable: func(db *pachsql.DB) error {
				return pachsql.CreateTestTable(db, "test_table", pachsql.TestRow{})
			},
			Expected: &pachsql.TableInfo{
				"test_table",
				"public",
				[]pachsql.ColumnInfo{
					{"c_id", "INTEGER", false},
					{"c_smallint", "SMALLINT", false},
					{"c_int", "INTEGER", false},
					{"c_bigint", "BIGINT", false},
					{"c_float", "DOUBLE PRECISION", false},
					{"c_varchar", "CHARACTER VARYING", false},
					{"c_time", "TIMESTAMP WITHOUT TIME ZONE", false},
					{"c_smallint_null", "SMALLINT", true},
					{"c_int_null", "INTEGER", true},
					{"c_bigint_null", "BIGINT", true},
					{"c_float_null", "DOUBLE PRECISION", true},
					{"c_varchar_null", "CHARACTER VARYING", true},
					{"c_time_null", "TIMESTAMP WITHOUT TIME ZONE", true},
				},
			},
		},
		{
			Name:  "MySQL",
			NewDB: dockertestenv.NewMySQL,
			NewTable: func(db *pachsql.DB) error {
				return pachsql.CreateTestTable(db, "public.test_table", pachsql.TestRow{})
			},
			Expected: &pachsql.TableInfo{
				"test_table",
				"public",
				[]pachsql.ColumnInfo{
					{"c_id", "INT", false},
					{"c_smallint", "SMALLINT", false},
					{"c_int", "INT", false},
					{"c_bigint", "BIGINT", false},
					{"c_float", "FLOAT", false},
					{"c_varchar", "VARCHAR", false},
					{"c_time", "TIMESTAMP", false},
					{"c_smallint_null", "SMALLINT", true},
					{"c_int_null", "INT", true},
					{"c_bigint_null", "BIGINT", true},
					{"c_float_null", "FLOAT", true},
					{"c_varchar_null", "VARCHAR", true},
					{"c_time_null", "TIMESTAMP", true},
				},
			},
		},
		{
			Name:  "Snowflake",
			NewDB: testsnowflake.NewSnowSQL,
			NewTable: func(db *pachsql.DB) error {
				return pachsql.CreateTestTable(db, "test_table", pachsql.TestRow{})
			},
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
			// For mysql, we created a database named public via NewMySQL
			require.NoError(t, tc.NewTable(db))
			info, err := pachsql.GetTableInfo(ctx, db, "test_table")
			require.NoError(t, err)
			require.Len(t, info.Columns, reflect.TypeOf(pachsql.TestRow{}).NumField())
			require.Equal(t, tc.Expected, info)
		})
	}
}
