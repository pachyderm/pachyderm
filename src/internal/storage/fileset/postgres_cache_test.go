package fileset

import (
	"strconv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func newTestCache(t *testing.T, db *pachsql.DB, tr track.Tracker, maxSize int) *Cache {
	ctx := pctx.TestContext(t)
	tx := db.MustBegin()
	tx.MustExec(`CREATE SCHEMA IF NOT EXISTS storage`)
	require.NoError(t, CreatePostgresCacheV1(ctx, tx))
	require.NoError(t, tx.Commit())
	return NewCache(db, tr, maxSize)
}

func TestPostgresCache(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	storage := NewTestStorage(ctx, t, db, tr)
	maxSize := 5
	cache := newTestCache(t, db, tr, maxSize)
	ids := make([]ID, maxSize+1)
	put := func(i int) {
		id, err := storage.newWriter(ctx, WithTTL(track.ExpireNow)).Close()
		require.NoError(t, err)
		valueProto := &TestCacheValue{FileSetId: id.HexString()}
		data, err := proto.Marshal(valueProto)
		require.NoError(t, err)
		value := &anypb.Any{
			TypeUrl: "/" + string(proto.MessageName(valueProto)),
			Value:   data,
		}
		tag := "odd"
		if i%2 == 0 {
			tag = "even"
		}
		require.NoError(t, cache.Put(ctx, strconv.Itoa(i), value, []ID{*id}, tag))
		ids[i] = *id
	}
	checkExists := func(i int) {
		value, err := cache.Get(ctx, strconv.Itoa(i))
		require.NoError(t, err)
		valueProto := &TestCacheValue{}
		require.NoError(t, value.UnmarshalTo(valueProto))
		require.Equal(t, ids[i].HexString(), valueProto.FileSetId)
		exists, err := storage.exists(ctx, ids[i])
		require.NoError(t, err)
		require.True(t, exists)
	}
	checkNotExists := func(i int) {
		_, err := cache.Get(ctx, strconv.Itoa(i))
		require.YesError(t, err)
		exists, err := storage.exists(ctx, ids[i])
		require.NoError(t, err)
		require.False(t, exists)
	}
	// Fill the cache and confirm that the entries are retrievable and correct.
	// Also, confirm the referenced file sets still exist after gc.
	for i := 0; i < maxSize; i++ {
		put(i)
	}
	gc := storage.NewGC(time.Second)
	_, err := gc.RunOnce(ctx)
	require.NoError(t, err)
	for i := 0; i < maxSize; i++ {
		checkExists(i)
	}
	// Add one more entry to confirm that the oldest entry is evicted and the
	// rest of the entries are retrievable and correct.
	// Also, confirm the oldest entry's file set is deleted and the others
	// still exist.
	put(maxSize)
	_, err = gc.RunOnce(ctx)
	require.NoError(t, err)
	checkNotExists(0)
	for i := 1; i < maxSize+1; i++ {
		checkExists(i)
	}
	// Clear the odd index entries and confirm they no longer exist.
	require.NoError(t, cache.Clear(ctx, "odd"))
	_, err = gc.RunOnce(ctx)
	require.NoError(t, err)
	for i := 1; i < maxSize+1; i++ {
		if i%2 != 0 {
			checkNotExists(i)
			continue
		}
		checkExists(i)
	}
}
