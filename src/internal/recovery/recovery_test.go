package recovery

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	"github.com/pachyderm/pachyderm/v2/src/version"
	uuid "github.com/satori/go.uuid"
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

	var got, want struct {
		ChunksetID       int64     `db:"chunkset_id"`
		PachydermVersion string    `db:"pachyderm_version"`
		SQLDumpFileSetID uuid.UUID `db:"sql_dump_fileset_id"`
	}
	want.ChunksetID = 1
	want.PachydermVersion = version.Version.String()
	want.SQLDumpFileSetID = uuid.UUID(*fsID)
	if err := dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		if err := tx.GetContext(ctx, &got, `select chunkset_id, pachyderm_version, sql_dump_fileset_id from recovery.snapshots where id=$1`, snapID); err != nil {
			return errors.Wrap(err, "read snapshot")
		}
		return nil
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("snapshot row (-want +got):\n%s", diff)
	}
}
