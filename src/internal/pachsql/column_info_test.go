package pachsql_test

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestGetTableColumns(t *testing.T) {
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
	}
	ctx := context.Background()
	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			db := tc.NewDB(t)
			// Create Table
			infos, err := pachsql.GetTableColumns(ctx, db, "test_table")
			require.NoError(t, err)
			require.Len(t, infos, 3)
		})
	}
}
