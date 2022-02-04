package fileset

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

func newTestCache(t *testing.T, db *pachsql.DB, tr track.Tracker, maxSize int) *Cache {
	ctx := context.Background()
	tx := db.MustBegin()
	tx.MustExec(`CREATE SCHEMA IF NOT EXISTS storage`)
	require.NoError(t, CreatePostgresCacheV1(ctx, tx))
	require.NoError(t, tx.Commit())
	return NewCache(db, tr, maxSize)
}

func TestPostgresCache(t *testing.T) {
	ctx := context.Background()
	db := dockertestenv.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	storage := NewTestStorage(t, db, tr)
	maxSize := 5
	cache := newTestCache(t, db, tr, maxSize)
	ids := make([]ID, maxSize+1)
	putFileSet := func(i int) {
		id, err := storage.newWriter(ctx, WithTTL(track.ExpireNow)).Close()
		require.NoError(t, err)
		valueProto := &TestCacheValue{FileSetId: id.HexString()}
		data, err := proto.Marshal(valueProto)
		require.NoError(t, err)
		value := &types.Any{
			TypeUrl: "/" + proto.MessageName(valueProto),
			Value:   data,
		}
		require.NoError(t, cache.Put(ctx, strconv.Itoa(i), value, []ID{*id}))
		ids[i] = *id
	}
	checkFileSet := func(i int) {
		value, err := cache.Get(ctx, strconv.Itoa(i))
		require.NoError(t, err)
		valueProto := &TestCacheValue{}
		require.NoError(t, types.UnmarshalAny(value, valueProto))
		require.Equal(t, ids[i].HexString(), valueProto.FileSetId)
		exists, err := storage.exists(ctx, ids[i])
		require.NoError(t, err)
		require.True(t, exists)
	}
	// Fill the cache and confirm that the entries are retrievable and correct.
	// Also, confirm the referenced file sets still exist after gc.
	for i := 0; i < maxSize; i++ {
		putFileSet(i)
	}
	gc := storage.NewGC(time.Second)
	_, err := gc.RunOnce(ctx)
	require.NoError(t, err)
	for i := 0; i < maxSize; i++ {
		checkFileSet(i)
	}
	// Add one more entry to confirm that the oldest entry is evicted and the
	// rest of the entries are retrievable and correct.
	// Also, confirm the oldest entry's file set is deleted and the others
	// still exist.
	putFileSet(maxSize)
	_, err = cache.Get(ctx, strconv.Itoa(0))
	require.YesError(t, err)
	_, err = gc.RunOnce(ctx)
	require.NoError(t, err)
	exists, err := storage.exists(ctx, ids[0])
	require.NoError(t, err)
	require.False(t, exists)
	for i := 1; i < maxSize+1; i++ {
		checkFileSet(i)
	}
}
