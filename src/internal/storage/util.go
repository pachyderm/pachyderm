package storage

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"google.golang.org/grpc"
	"testing"
)

func NewTestFilesetServer(t testing.TB, db *pachsql.DB) *Server {
	ctx := pctx.TestContext(t)
	b, buckURL := dockertestenv.NewTestBucket(ctx, t)
	t.Log("bucket", buckURL)
	s, err := New(ctx, Env{
		DB:     db,
		Bucket: b,
		Config: pachconfig.StorageConfiguration{},
	})
	require.NoError(t, err)
	return s
}

func NewTestFilesetClient(t testing.TB, db *pachsql.DB) storage.FilesetClient {
	s := NewTestFilesetServer(t, db)
	gc := grpcutil.NewTestClient(t, func(gs *grpc.Server) {
		storage.RegisterFilesetServer(gs, s)
	})
	return storage.NewFilesetClient(gc)
}
