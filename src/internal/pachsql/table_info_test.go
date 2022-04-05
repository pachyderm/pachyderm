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
				[]pachsql.ColumnInfo{
					{"public", "test_table", "integer", false},
					{"public", "test_table", "smallint", false},
					{"public", "test_table", "integer", false},
					{"public", "test_table", "bigint", false},
					{"public", "test_table", "double precision", false},
					{"public", "test_table", "character varying", false},
					{"public", "test_table", "timestamp without time zone", false},
					{"public", "test_table", "smallint", true},
					{"public", "test_table", "integer", true},
					{"public", "test_table", "bigint", true},
					{"public", "test_table", "double precision", true},
					{"public", "test_table", "character varying", true},
					{"public", "test_table", "timestamp without time zone", true},
				},
			},
		},
		{
			Name:  "MySQL",
			NewDB: dockertestenv.NewMySQL,
			Expected: &pachsql.TableInfo{
				[]pachsql.ColumnInfo{
					{"public", "test_table", "int", false},
					{"public", "test_table", "smallint", false},
					{"public", "test_table", "int", false},
					{"public", "test_table", "bigint", false},
					{"public", "test_table", "float", false},
					{"public", "test_table", "varchar", false},
					{"public", "test_table", "timestamp", false},
					{"public", "test_table", "smallint", true},
					{"public", "test_table", "int", true},
					{"public", "test_table", "bigint", true},
					{"public", "test_table", "float", true},
					{"public", "test_table", "varchar", true},
					{"public", "test_table", "timestamp", true},
				},
			},
		},
		{
			Name:  "Snowflake",
			NewDB: testsnowflake.NewSnowSQL,
			Expected: &pachsql.TableInfo{
				[]pachsql.ColumnInfo{
					{"PUBLIC", "TEST_TABLE", "NUMBER", false},
					{"PUBLIC", "TEST_TABLE", "NUMBER", false},
					{"PUBLIC", "TEST_TABLE", "NUMBER", false},
					{"PUBLIC", "TEST_TABLE", "NUMBER", false},
					{"PUBLIC", "TEST_TABLE", "FLOAT", false},
					{"PUBLIC", "TEST_TABLE", "TEXT", false},
					{"PUBLIC", "TEST_TABLE", "TIMESTAMP_NTZ", false},
					{"PUBLIC", "TEST_TABLE", "NUMBER", true},
					{"PUBLIC", "TEST_TABLE", "NUMBER", true},
					{"PUBLIC", "TEST_TABLE", "NUMBER", true},
					{"PUBLIC", "TEST_TABLE", "FLOAT", true},
					{"PUBLIC", "TEST_TABLE", "TEXT", true},
					{"PUBLIC", "TEST_TABLE", "TIMESTAMP_NTZ", true},
				},
			},
		},
	}
	ctx := context.Background()
	for _, tc := range tcs {
		t.Log(fmt.Sprintf("Running: %s", tc.Name))
		t.Run(tc.Name, func(t *testing.T) {
			db := tc.NewDB(t)
			require.NoError(t, pachsql.CreateTestTable(db, "test_table"))
			info, err := pachsql.GetTableInfo(ctx, db, "test_table")
			require.NoError(t, err)
			t.Log(info)
			require.Len(t, info.Columns, reflect.TypeOf(pachsql.TestRow{}).NumField())
			require.Equal(t, tc.Expected, info)
		})
	}
}
