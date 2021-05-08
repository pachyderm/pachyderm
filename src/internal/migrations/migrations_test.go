package migrations

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestMigration(t *testing.T) {
	db := testutil.NewTestDB(t)
	state := InitialState().
		Apply("test 1", func(ctx context.Context, env Env) error {
			// NoOp
			return nil
		}).
		Apply("test 2", func(ctx context.Context, env Env) error {
			_, err := env.Tx.ExecContext(ctx, `CREATE TABLE test_table1 (id BIGSERIAL PRIMARY KEY, field1 TEXT, field2 TEXT);`)
			return err
		}).
		Apply("test 3", func(ctx context.Context, env Env) error {
			_, err := env.Tx.ExecContext(ctx, `CREATE TABLE test_table2 (id BIGSERIAL PRIMARY KEY, field1 TEXT, field2 TEXT);`)
			return err
		})
	ctx := context.Background()
	func() {
		eg, ctx := errgroup.WithContext(ctx)
		const numWaiters = 10
		for i := 0; i < numWaiters; i++ {
			eg.Go(func() error {
				return BlockUntil(ctx, db, state)
			})
		}
		eg.Go(func() error {
			time.Sleep(time.Second)
			return ApplyMigrations(ctx, db, Env{}, state)
		})
		require.NoError(t, eg.Wait())
	}()
	var count int
	require.NoError(t, db.GetContext(ctx, &count, `SELECT count(*) FROM migrations`))
	assert.Equal(t, state.Number()+1, count)
	var max int
	require.NoError(t, db.GetContext(ctx, &max, `SELECT max(id) FROM migrations`))
	assert.Equal(t, state.Number(), max)
}
