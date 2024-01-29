package storage

import (
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

func TestServer(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	b, buckURL := dockertestenv.NewTestBucket(ctx, t)
	t.Log("bucket", buckURL)

	tracker := track.NewTestTracker(t, db)
	fileset.NewTestStorage(ctx, t, db, tracker)

	var config pachconfig.StorageConfiguration
	config.StorageDiskCacheSize = 3
	config.StorageDownloadConcurrencyLimit = 3
	config.StorageUploadConcurrencyLimit = 3
	s, err := New(Env{
		DB:     db,
		Bucket: b,
	}, config)
	require.NoError(t, err)

	w := s.Filesets.NewWriter(ctx)
	require.NoError(t, w.Add("test.txt", "", strings.NewReader("hello world")))
	id, err := w.Close()
	require.NoError(t, err)
	t.Log(id)

	require.NoError(t, stream.ForEach(ctx, obj.NewKeyIterator(b, ""), func(k string) error {
		if !strings.HasPrefix(k, chunkPrefix) {
			t.Errorf("object key %q does not have prefix %q", k, chunkPrefix)
		}
		return nil
	}))
}
