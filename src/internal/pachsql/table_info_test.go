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
		Name  string
		NewDB func(t testing.TB) *pachsql.DB
	}
	tcs := []testCase{
		{
			Name:  "Postgres",
			NewDB: dockertestenv.NewPostgres,
		},
		{
			Name:  "MySQL",
			NewDB: dockertestenv.NewMySQL,
		},
		{
			Name:  "Snowflake",
			NewDB: testsnowflake.NewSnowSQL,
		},
	}
	ctx := context.Background()
	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			db := tc.NewDB(t)
			require.NoError(t, pachsql.CreateTestTable(db, "test_table"))
			info, err := pachsql.GetTableInfo(ctx, db, "test_table")
			require.NoError(t, err)
			t.Log(info)
			require.Len(t, info.Columns, reflect.TypeOf(pachsql.TestRow{}).NumField())
		})
	}
}
