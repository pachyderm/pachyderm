package fileset_test

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestCreateChunkset(t *testing.T) {
	ctx := pctx.TestContext(t)
	s, db := setUptest(t, ctx)

	// create a fileset
	gc := s.NewGC(5 * time.Second)
	w := s.NewWriter(ctx, fileset.WithTTL(time.Hour))
	require.NoError(t, w.Add("a.txt", "datum1", strings.NewReader("test data")))
	id, err := w.Close()
	require.NoError(t, err)
	// expire it
	require.NoError(t, s.Drop(ctx, *id))
	// run the gc
	countDeleted, err := gc.RunOnce(ctx)
	require.NoError(t, err)
	// gc sanity check. the fileset should be deleted
	t.Log(countDeleted)
	require.True(t, countDeleted > 0)

	// create fileset and rerun gc.
	w = s.NewWriter(ctx, fileset.WithTTL(time.Hour))
	require.NoError(t, w.Add("b.txt", "datum2", strings.NewReader("test data2")))
	id, err = w.Close()
	require.NoError(t, err)
	// create a chunkset
	err = dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		_, err := s.CreateChunkSet(ctx, tx)
		return err
	})
	require.NoError(t, err)
	// expire it
	require.NoError(t, s.Drop(ctx, *id))
	// run the gc
	countDeleted, err = gc.RunOnce(ctx)
	require.NoError(t, err)
	// the fileset can no longer be deleted
	require.True(t, countDeleted == 0)
}

func setUptest(t *testing.T, ctx context.Context) (*fileset.Storage, *pachsql.DB) {
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	tr := track.NewPostgresTracker(db)
	p := filepath.Join(t.TempDir(), "obj-store")
	kvStore := kv.NewFSStore(p, 512, chunk.DefaultMaxChunkSize)
	chunks := chunk.NewStorage(kvStore, nil, db, tr)
	store := fileset.NewPostgresStore(db)
	return fileset.NewStorage(store, tr, chunks), db
}
