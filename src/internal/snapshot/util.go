package snapshot

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"google.golang.org/grpc"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
)

func NewTestClient(t testing.TB) snapshot.APIClient {
	db := dockertestenv.NewMigratedTestDB(t, clusterstate.DesiredClusterState)
	tracker := track.NewPostgresTracker(db)
	s := fileset.NewStorage(fileset.NewPostgresStore(db), tracker, chunk.NewStorage(kv.NewMemStore(), nil, db, tracker))
	srv := NewAPIServer(Env{
		DB:    db,
		store: s,
	})
	gc := grpcutil.NewTestClient(t, func(s *grpc.Server) {
		snapshot.RegisterAPIServer(s, srv)
	})
	return snapshot.NewAPIClient(gc)
}
