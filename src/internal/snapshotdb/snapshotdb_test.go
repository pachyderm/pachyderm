package snapshotdb

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

type dependencies struct {
	ctx context.Context
	db  *pachsql.DB
	tx  *pachsql.Tx
	s   *fileset.Storage
}

func DB(t testing.TB) (context.Context, *pachsql.DB) {
	t.Helper()
	ctx := pctx.Child(pctx.TestContext(t), t.Name())
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	return ctx, db
}

func FilesetStorage(t testing.TB, db *pachsql.DB) *fileset.Storage {
	t.Helper()
	tracker := track.NewPostgresTracker(db)
	s := fileset.NewStorage(fileset.NewPostgresStore(db), tracker, chunk.NewStorage(kv.NewMemStore(), nil, db, tracker))
	return s
}

func withDependencies(t *testing.T, f func(d dependencies)) {
	ctx, db := DB(t)
	s := FilesetStorage(t, db)
	err := dbutil.WithTx(ctx, db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		f(dependencies{ctx: ctx, db: db, tx: sqlTx, s: s})
		return nil
	})
	require.NoError(t, err)
}
