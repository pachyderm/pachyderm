package pachsql_test

import (
	"context"
	"fmt"
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
					{"c_id", "SMALLINT", false},
					{"c_smallint", "SMALLINT", false},
					{"c_int", "INTEGER", false},
					{"c_bigint", "BIGINT", false},
					{"c_float", "DOUBLE PRECISION", false},
					{"c_numeric_int", "NUMERIC", false},
					{"c_numeric_float", "NUMERIC", false},
					{"c_varchar", "CHARACTER VARYING", false},
					{"c_time", "TIMESTAMP WITHOUT TIME ZONE", false},
					{"c_smallint_null", "SMALLINT", true},
					{"c_int_null", "INTEGER", true},
					{"c_bigint_null", "BIGINT", true},
					{"c_float_null", "DOUBLE PRECISION", true},
					{"c_numeric_int_null", "NUMERIC", true},
					{"c_numeric_float_null", "NUMERIC", true},
					{"c_varchar_null", "CHARACTER VARYING", true},
					{"c_time_null", "TIMESTAMP WITHOUT TIME ZONE", true},
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
					{"c_id", "SMALLINT", false},
					{"c_smallint", "SMALLINT", false},
					{"c_int", "INT", false},
					{"c_bigint", "BIGINT", false},
					{"c_float", "FLOAT", false},
					{"c_numeric_int", "DECIMAL", false},
					{"c_numeric_float", "DECIMAL", false},
					{"c_varchar", "VARCHAR", false},
					{"c_time", "TIMESTAMP", false},
					{"c_smallint_null", "SMALLINT", true},
					{"c_int_null", "INT", true},
					{"c_bigint_null", "BIGINT", true},
					{"c_float_null", "FLOAT", true},
					{"c_numeric_int_null", "DECIMAL", true},
					{"c_numeric_float_null", "DECIMAL", true},
					{"c_varchar_null", "VARCHAR", true},
					{"c_time_null", "TIMESTAMP", true},
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
					{"C_ID", "NUMBER", false},
					{"C_SMALLINT", "NUMBER", false},
					{"C_INT", "NUMBER", false},
					{"C_BIGINT", "NUMBER", false},
					{"C_FLOAT", "FLOAT", false},
					{"C_NUMERIC_INT", "NUMBER", false},
					{"C_NUMERIC_FLOAT", "NUMBER", false},
					{"C_VARCHAR", "TEXT", false},
					{"C_TIME", "TIMESTAMP_NTZ", false},
					{"C_SMALLINT_NULL", "NUMBER", true},
					{"C_INT_NULL", "NUMBER", true},
					{"C_BIGINT_NULL", "NUMBER", true},
					{"C_FLOAT_NULL", "FLOAT", true},
					{"C_NUMERIC_INT_NULL", "NUMBER", true},
					{"C_NUMERIC_FLOAT_NULL", "NUMBER", true},
					{"C_VARCHAR_NULL", "TEXT", true},
					{"C_TIME_NULL", "TIMESTAMP_NTZ", true},
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
			info, err := pachsql.GetTableInfo(ctx, db, fmt.Sprintf("%s.test_table", tc.Expected.Schema))
			require.NoError(t, err)
			require.Len(t, info.Columns, reflect.TypeOf(pachsql.TestRow{}).NumField())
			require.Equal(t, tc.Expected, info)
		})
	}
}
