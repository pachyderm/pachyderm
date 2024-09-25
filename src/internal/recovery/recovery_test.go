package recovery

import (
	"context"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

func TestSnapshotDatabase(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewMigratedTestDB(t, clusterstate.DesiredClusterState)
	tracker := track.NewPostgresTracker(db)
	s := fileset.NewStorage(fileset.NewPostgresStore(db), tracker, chunk.NewStorage(kv.NewMemStore(), nil, db, tracker))
	var snapID SnapshotID
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		snapID, err = createSnapshotRow(ctx, tx, s)
		return errors.Wrap(err, "createSnapshotRow")
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	if snapID < 1 {
		t.Errorf("snapshot id < 1: %v", snapID)
	}

	w := s.NewWriter(ctx)
	if err := w.Add("dump.sql", "", strings.NewReader("hello, world")); err != nil {
		t.Fatalf("writer.Add(dump.sql): %v", err)
	}
	fsID, err := w.Close()
	if err != nil {
		t.Fatalf("writer.Close: %v", err)
	}
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		return errors.Wrap(addDatabaseDump(ctx, tx, snapID, *fsID), "addDatabaseDump")
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
}
