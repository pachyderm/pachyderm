package snapshot

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/klauspost/compress/zstd"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshotdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
	"google.golang.org/protobuf/testing/protocmp"
)

func validateDumpFileset(ctx context.Context, t *testing.T, s *fileset.Storage, fsHandle *fileset.Handle) {
	t.Helper()
	fs, err := s.Open(ctx, []*fileset.Handle{fsHandle})
	if err != nil {
		t.Fatalf("open fileset (%v): %v", fsHandle.HexString(), err)
	}
	var lines []string
	if err := fs.Iterate(ctx, func(f fileset.File) error {
		if got, want := f.Index().Path, "dump.sql.zst"; got != want {
			return errors.Errorf("invalid file in fileset:\n  got: %v\n want: %v", got, want)
		}
		r, w := io.Pipe()
		zr, err := zstd.NewReader(r)
		if err != nil {
			return errors.Wrap(err, "new zstd reader")
		}
		doneCh := make(chan error)
		go func() {
			s := bufio.NewScanner(zr)
			for s.Scan() {
				lines = append(lines, s.Text())
			}
			doneCh <- s.Err()
			close(doneCh)
		}()
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
	if len(lines) < 10 {
		t.Fatalf("too few lines in buf:\n%v", strings.Join(lines, "\n"))
	}
	var looksLikeDump bool
	for i := 0; i < 10; i++ {
		line := lines[i]
		t.Logf("database dump line: %v", line)
		if strings.Contains(line, "-- PostgreSQL database dump") {
			looksLikeDump = true
		}
	}
	if !looksLikeDump {
		t.Errorf("the file we got out of PFS does not look like a database dump; see debug logs above")
	}
}

func TestCreateAndRestoreSnaphot(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewMigratedTestDB(t, clusterstate.DesiredClusterState)
	tracker := track.NewPostgresTracker(db)
	storage := fileset.NewStorage(fileset.NewPostgresStore(db), tracker, chunk.NewStorage(kv.NewMemStore(), nil, db, tracker))

	// Create a fileset so we have some data to backup.
	w := storage.NewWriter(ctx)
	if err := w.Add("test", "", strings.NewReader("this is a test")); err != nil {
		t.Fatalf("add test file: %v", err)
	}
	testDataFsHandle, err := w.Close()
	if err != nil {
		t.Fatalf("close testdata fileset: %v", err)
	}

	// Create a snapshot.
	s := &Snapshotter{DB: db, Storage: storage}

	snapID, err := s.CreateSnapshot(ctx, CreateSnapshotOptions{})
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}

	// Check that we can read the snapshot we just made.
	var got *snapshot.SnapshotInfo
	want := new(snapshot.SnapshotInfo)
	want.Id = int64(snapID)
	want.ChunksetId = 1
	want.PachydermVersion = "v0.0.0"
	if err := dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		var err error
		got, err = snapshotdb.GetSnapshot(ctx, tx, int64(snapID))
		if err != nil {
			return errors.Wrap(err, "GetSnapshot")
		}
		return nil
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	t.Logf("fileset handle: %v", got.GetSqlDumpFilesetId())
	var fsToken fileset.Token
	if err := fsToken.Scan(got.GetSqlDumpFilesetId()); err != nil {
		t.Fatalf("parse token %v: %v", got.GetSqlDumpFilesetId(), err)
	}
	got.CreatedAt = nil
	got.SqlDumpFilesetId = ""
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("snapshot row (-want +got):\n%s", diff)
	}
	// Validate the database dump fileset we created in CreateSnapshot.
	validateDumpFileset(ctx, t, storage, fileset.NewHandle(fsToken))

	// Restore the snapshot.
	if err := s.RestoreSnapshot(ctx, snapID, RestoreSnapshotOptions{}); err != nil {
		t.Errorf("snapshot not restorable: %v", err)
	}
	got = nil
	if err := dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		var err error
		got, err = snapshotdb.GetSnapshot(ctx, tx, int64(snapID))
		if err != nil {
			return errors.Wrap(err, "GetSnapshot (after restore)")
		}
		return nil
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	// Validate the database backup fileset created during restoration.
	fsToken = fileset.Token{}
	if err := fsToken.Scan(got.GetSqlDumpFilesetId()); err != nil {
		t.Fatalf("parse token %v: %v", got.GetSqlDumpFilesetId(), err)
	}
	validateDumpFileset(ctx, t, storage, fileset.NewHandle(fsToken))

	// Drop the chunkset that's keeping the backed up chunks alive (if the restore failed), just
	// to make sure that it's the original references that are keeping the backed up filesets
	// around.
	if err := dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		if _, err := tx.ExecContext(ctx, `truncate table recovery.snapshots`); err != nil {
			return errors.Wrap(err, "truncate snapshots table")
		}
		if err := storage.DropChunkSet(ctx, tx, fileset.ChunkSetID(got.ChunksetId)); err != nil {
			return errors.Wrapf(err, "DropChunkSet(%v)", got.ChunksetId)
		}
		return nil
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}

	// GC to ensure our dropped refs drop the chunks; if chunks are dropped, then the read below
	// will fail.
	if err := storage.NewGC(time.Second).RunUntilEmpty(ctx); err != nil {
		t.Errorf("gc: %v", err)
	}

	// Check that the fileset created at the beginning of the test is fully readable after
	// restoring.
	fs, err := storage.Open(ctx, []*fileset.Handle{testDataFsHandle})
	if err != nil {
		t.Fatalf("open testdata fileset: %v", err)
	}
	var buf bytes.Buffer
	if err := fs.Iterate(ctx, func(f fileset.File) error {
		if err := f.Content(ctx, &buf); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("read testdata filest: %v", err)
	}
	if got, want := buf.String(), "this is a test"; got != want {
		t.Errorf("test data in snapshotted fileset:\n  got: %v\n want: %v", got, want)
	}
}

func TestVersionCompatibility(t *testing.T) {
	testData := []struct {
		snapshot, running string
		ok                bool
	}{
		{"", "", true},
		{"foo", "bar", true},
		{"bar", "foo", true},
		{"", "v1.2.3", true},
		{"v1.2.3", "", true},
		{"v2.12.1", "v2.12.0", false},
		{"v2.12.0", "v2.12.1", true},
		{"v2.13.0", "v2.12.0", false},
		{"v2.12.0", "v2.13.0", true},
		{"v2.12.0-pre.gb26915d9b7", "v2.12.0-pre.gb26915d9b7", true},
		{"v2.12.0-pre.gffffffffff", "v2.12.0-pre.g0000000000", true},
		{"v2.12.0-pre.g0000000000", "v2.12.0-pre.gaaaaaaaaaa", true},
		{"v2.12.0-pre.gb26915d9b7", "v2.12.0", true},
		{"v2.12.0-pre.gb26915d9b7", "v2.12.1", true},
		{"v2.12.0-pre.gb26915d9b7", "v2.13.0", true},
		{"v2.12.0", "v2.12.0-pre.gb26915d9b7", false},
		{"v2.12.1", "v2.12.0-pre.gb26915d9b7", false},
		{"v2.13.0", "v2.12.0-pre.gb26915d9b7", false},
		{"v2.12.0-pre.abc123", "v2.13.0-pre.123abc", true},
	}

	for _, test := range testData {
		err := checkVersionCompatibility(test.snapshot, test.running)
		if got, want := err == nil, test.ok; got != want {
			t.Errorf("restore %v onto %v:\n  got: %v (%v)\n want: %v", test.snapshot, test.running, got, err, want)
		}
	}
}
