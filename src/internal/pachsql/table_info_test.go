package pachsql_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testsnowflake"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestGetTableInfo(suite *testing.T) {
	type testCase struct {
		Name     string
		NewDB    func(testing.TB, string) *pachsql.DB
		Expected *pachsql.TableInfo
	}
	tcs := []testCase{
		{
			Name:  "Postgres",
			NewDB: dockertestenv.NewEphemeralPostgresDB,
			Expected: &pachsql.TableInfo{
				"test_table",
				"public",
				[]pachsql.ColumnInfo{
					{"c_id", "INTEGER", false, 32, 0},
					{"c_smallint", "SMALLINT", false, 16, 0},
					{"c_int", "INTEGER", false, 32, 0},
					{"c_bigint", "BIGINT", false, 64, 0},
					{"c_float", "DOUBLE PRECISION", false, 53, 0},
					// {"c_numeric", "NUMERIC", false, 0, 0},
					{"c_numeric_int", "NUMERIC", false, 20, 0},
					{"c_numeric_float", "NUMERIC", false, 20, 2},
					{"c_varchar", "CHARACTER VARYING", false, 0, 0},
					{"c_time", "TIMESTAMP WITHOUT TIME ZONE", false, 0, 0},
					{"c_smallint_null", "SMALLINT", true, 16, 0},
					{"c_int_null", "INTEGER", true, 32, 0},
					{"c_bigint_null", "BIGINT", true, 64, 0},
					{"c_float_null", "DOUBLE PRECISION", true, 53, 0},
					{"c_varchar_null", "CHARACTER VARYING", true, 0, 0},
					{"c_time_null", "TIMESTAMP WITHOUT TIME ZONE", true, 0, 0},
				},
			},
		},
		{
			Name:  "MySQL",
			NewDB: dockertestenv.NewEphemeralMySQLDB,
			Expected: &pachsql.TableInfo{
				"test_table",
				"MySQL doesn't have schema, use database name instead",
				[]pachsql.ColumnInfo{
					{"c_id", "INT", false, 10, 0},
					{"c_smallint", "SMALLINT", false, 5, 0},
					{"c_int", "INT", false, 10, 0},
					{"c_bigint", "BIGINT", false, 19, 0},
					{"c_float", "FLOAT", false, 12, 0},
					// {"c_numeric", "DECIMAL", false, 10, 0},
					{"c_numeric_int", "DECIMAL", false, 20, 0},
					{"c_numeric_float", "DECIMAL", false, 20, 2},
					{"c_varchar", "VARCHAR", false, 0, 0},
					{"c_time", "TIMESTAMP", false, 0, 0},
					{"c_smallint_null", "SMALLINT", true, 5, 0},
					{"c_int_null", "INT", true, 10, 0},
					{"c_bigint_null", "BIGINT", true, 19, 0},
					{"c_float_null", "FLOAT", true, 12, 0},
					{"c_varchar_null", "VARCHAR", true, 0, 0},
					{"c_time_null", "TIMESTAMP", true, 0, 0},
				},
			},
		},
		{
			Name:  "Snowflake",
			NewDB: testsnowflake.NewEphemeralSnowflakeDB,
			Expected: &pachsql.TableInfo{
				"test_table",
				"public",
				[]pachsql.ColumnInfo{
					{"C_ID", "NUMBER", false, 38, 0},
					{"C_SMALLINT", "NUMBER", false, 38, 0},
					{"C_INT", "NUMBER", false, 38, 0},
					{"C_BIGINT", "NUMBER", false, 38, 0},
					{"C_FLOAT", "FLOAT", false, 0, 0},
					// {"C_NUMERIC", "NUMBER", false, 38, 0},
					{"C_NUMERIC_INT", "NUMBER", false, 20, 0},
					{"C_NUMERIC_FLOAT", "NUMBER", false, 20, 2},
					{"C_VARCHAR", "TEXT", false, 0, 0},
					{"C_TIME", "TIMESTAMP_NTZ", false, 0, 0},
					{"C_SMALLINT_NULL", "NUMBER", true, 38, 0},
					{"C_INT_NULL", "NUMBER", true, 38, 0},
					{"C_BIGINT_NULL", "NUMBER", true, 38, 0},
					{"C_FLOAT_NULL", "FLOAT", true, 0, 0},
					{"C_VARCHAR_NULL", "TEXT", true, 0, 0},
					{"C_TIME_NULL", "TIMESTAMP_NTZ", true, 0, 0},
				},
			},
		},
	}
	ctx := context.Background()
	for _, tc := range tcs {
		suite.Run(tc.Name, func(t *testing.T) {
			dbName := testutil.GenerateEphermeralDBName(t)
			if tc.Name == "MySQL" {
				tc.Expected.Schema = dbName
			}
			db := tc.NewDB(t, dbName)
			require.NoError(t, pachsql.CreateTestTable(db, "test_table", pachsql.TestRow{}))
			info, err := pachsql.GetTableInfo(ctx, db, "test_table")
			require.NoError(t, err)
			require.Len(t, info.Columns, reflect.TypeOf(pachsql.TestRow{}).NumField())
			require.Equal(t, tc.Expected, info)
		})
	}
}
