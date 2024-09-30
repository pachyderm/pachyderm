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
	require.Equal(t, 1, countDeleted)

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

func TestDropChunkset(t *testing.T) {
	t.Run("chunkset does not exist", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		s, db := setUptest(t, ctx)
		// create a fileset
		w := s.NewWriter(ctx, fileset.WithTTL(time.Hour))
		require.NoError(t, w.Add("a.txt", "datum1", strings.NewReader("test data")))
		_, err := w.Close()
		require.NoError(t, err)
		// drop a chunkset that does not exist
		err = dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
			err := s.DropChunkSet(ctx, tx, 2)
			return err
		})
		require.YesError(t, err)
		require.Equal(t, "no chunkset found with the given id: 2", err.Error())
	})

	t.Run("happy path", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		s, db := setUptest(t, ctx)
		// create a fileset
		gc := s.NewGC(5 * time.Second)
		w := s.NewWriter(ctx, fileset.WithTTL(time.Hour))
		require.NoError(t, w.Add("a.txt", "datum1", strings.NewReader("test data")))
		filesetID, err := w.Close()
		require.NoError(t, err)
		// create a chunkset
		var chunkSetID fileset.ChunkSetID
		err = dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
			chunkSetID, err = s.CreateChunkSet(ctx, tx)
			return err
		})
		require.NoError(t, err)
		// drop the chunkset
		err = dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
			err := s.DropChunkSet(ctx, tx, chunkSetID)
			return err
		})
		require.NoError(t, err)
		// expire it
		require.NoError(t, s.Drop(ctx, *filesetID))
		// run the gc
		countDeleted, err := gc.RunOnce(ctx)
		require.NoError(t, err)
		// now the fileset can be deleted
		require.Equal(t, 1, countDeleted)
	})
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
