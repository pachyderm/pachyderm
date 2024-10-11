package recovery

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/klauspost/compress/zstd"
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

func TestCreateSnaphot(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewMigratedTestDB(t, clusterstate.DesiredClusterState)
	tracker := track.NewPostgresTracker(db)
	s := fileset.NewStorage(fileset.NewPostgresStore(db), tracker, chunk.NewStorage(kv.NewMemStore(), nil, db, tracker))

	snapID, err := CreateSnapshot(ctx, db, s)
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}

	var got, want struct {
		ChunksetID       int64     `db:"chunkset_id"`
		PachydermVersion string    `db:"pachyderm_version"`
		SQLDumpFileSetID uuid.UUID `db:"sql_dump_fileset_id"`
	}
	want.ChunksetID = 1
	want.PachydermVersion = version.Version.String()
	if err := dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		if err := tx.GetContext(ctx, &got, `select chunkset_id, pachyderm_version, sql_dump_fileset_id from recovery.snapshots where id=$1`, snapID); err != nil {
			return errors.Wrap(err, "read snapshot")
		}
		return nil
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	fsid := fileset.ID(got.SQLDumpFileSetID[:])
	got.SQLDumpFileSetID = uuid.UUID{}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("snapshot row (-want +got):\n%s", diff)
	}

	fs, err := s.Open(ctx, []fileset.ID{fsid})
	if err != nil {
		t.Fatalf("open fileset (%v): %v", got.SQLDumpFileSetID.String(), err)
	}
	var buf bytes.Buffer
	if err := fs.Iterate(ctx, func(f fileset.File) error {
		if got, want := f.Index().Path, "dump.sql.zst"; got != want {
			return errors.Errorf("invalid file in filest:\n  got: %v\n want: %v", got, want)
		}
		r, w := io.Pipe()
		zr, err := zstd.NewReader(r)
		doneCh := make(chan error)
		go func() {
			_, err := io.Copy(&buf, zr)
			doneCh <- err
			close(doneCh)
		}()
		if err != nil {
			return errors.Wrap(err, "new zstd reader")
		}
		if err := f.Content(ctx, w); err != nil {
			w.CloseWithError(err) //nolint:errcheck
			return errors.Wrap(err, "read content")
		}
		w.Close() //nolint:errcheck
		if err := <-doneCh; err != nil {
			return errors.Wrap(err, "copy to buffer")
		}
		return nil
	}); err != nil {
		t.Errorf("iterate over fileset: %v", err)
	}
	var looksLikeDump bool
	scan := bufio.NewScanner(&buf)
	for i := 0; i < 10; i++ {
		if !scan.Scan() {
			t.Fatal("too few lines in buf")
		}
		line := scan.Text()
		t.Logf("database dump line: %v", line)
		if strings.Contains(line, "-- PostgreSQL database dump") {
			looksLikeDump = true
		}
	}
	if !looksLikeDump {
		t.Errorf("the file we got out of PFS does not look like a database dump; see debug logs above")
	}
}
