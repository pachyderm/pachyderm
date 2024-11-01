package snapshot

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"strings"
	"testing"
)

func TestRestoreAndCheckStorage(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewMigratedTestDB(t, clusterstate.DesiredClusterState)
	tracker := track.NewPostgresTracker(db)
	chunks := chunk.NewStorage(kv.NewMemStore(), nil, db, tracker)
	storage := fileset.NewStorage(fileset.NewPostgresStore(db), tracker, chunks)

	// Create a fileset so we have some data to backup.
	w := storage.NewWriter(ctx)
	if err := w.Add("test", "", strings.NewReader("this is a test")); err != nil {
		t.Fatalf("add test file: %v", err)
	}
	_, err := w.Close()
	if err != nil {
		t.Fatalf("close testdata fileset: %v", err)
	}

	// Create a snapshot.
	s := &Snapshotter{DB: db, Storage: storage}

	snapID, err := s.CreateSnapshot(ctx, CreateSnapshotOptions{})
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	// Restore the snapshot.
	if err := s.RestoreSnapshot(ctx, snapID, RestoreSnapshotOptions{}); err != nil {
		t.Errorf("snapshot not restorable: %v", err)
	}

	_, err = chunks.Check(ctx, nil, nil, true)
	if err != nil {
		t.Fatalf("check storage: %v", err)
	}
}
