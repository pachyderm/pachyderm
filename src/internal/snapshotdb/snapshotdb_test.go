package snapshotdb_test

import (
	"context"
	"path/filepath"
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
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
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
	store := kv.NewFSStore(filepath.Join(t.TempDir(), "obj-store"), 512, chunk.DefaultMaxChunkSize)
	tr := track.NewPostgresTracker(db)
	return fileset.NewStorage(fileset.NewPostgresStore(db), tr, chunk.NewStorage(store, nil, db, tr))
}

func withTx(t testing.TB, ctx context.Context, db *pachsql.DB, s *fileset.Storage, f func(d dependencies)) {
	t.Helper()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	f(dependencies{ctx: ctx, db: db, tx: tx, s: s})
	require.NoError(t, tx.Commit())
}

func withDependencies(t *testing.T, f func(d dependencies)) {
	ctx, db := DB(t)
	withTx(t, ctx, db, FilesetStorage(t, db), func(d dependencies) {
		f(d)
	})
}
