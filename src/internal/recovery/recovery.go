// Package recovery implements subsystem-independent disaster recovery.
package recovery

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/jmoiron/sqlx"
	"github.com/klauspost/compress/zstd"
	"github.com/pachyderm/pachyderm/v2/src/internal/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"go.uber.org/zap"
)

var (
	// Variables replaced by the linker during Bazel builds.
	pgdump, libpq, libldap, liblber, libsasl string
)

func pgDumpPath() string {
	x, err := runfiles.Rlocation(pgdump)
	if err != nil || pgdump == "" {
		return "pg_dump" // Probably not built with bazel, look in $PATH and hope for the best.
	}
	return x
}

type SnapshotID int64

// String implements fmt.Stringer.
func (id SnapshotID) String() string {
	var invalid string
	if id < 1 {
		// a postgres bigserial is 1-2^63 stored in a signed int64.
		invalid = "Invalid"
	}
	return fmt.Sprintf("%vSnapshotID(%v)", invalid, int64(id))
}

// createSnapshotRow creates a snapshot row containing a chunkset referencing all live data.
func createSnapshotRow(ctx context.Context, tx *sqlx.Tx, s *fileset.Storage) (result SnapshotID, _ error) {
	chunksetID, err := s.CreateChunkSet(ctx, tx)
	if err != nil {
		return 0, errors.Wrap(err, "create chunkset")
	}
	if err := tx.GetContext(ctx, &result, `insert into recovery.snapshots (chunkset_id, pachyderm_version) values ($1, $2) returning id`, chunksetID, version.Version.String()); err != nil {
		return 0, errors.Wrap(err, "create snapshot row")
	}
	return result, nil
}

// addDatabaseDump updates a snapshot row to contain a reference to a database dump.
func addDatabaseDump(ctx context.Context, tx *sqlx.Tx, snapshotID SnapshotID, filesetToken fileset.Token) error {
	result, err := tx.ExecContext(ctx, `update recovery.snapshots set sql_dump_fileset_id=$1 where id=$2`, filesetToken[:], snapshotID)
	if err != nil {
		return errors.Wrap(err, "update snapshot to contain database dump fileset")
	}
	got, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get affected row count")
	}
	if got != 1 {
		return errors.Errorf("rows affected by snapshot update: got %v want 1", got)
	}
	return nil
}

// dumpDatabase runs pg_dump on the provided database, writing the zstd-compressed content to w.
// The writer is not closed by this function.
func dumpDatabase(ctx context.Context, db *pachsql.DB, w io.WriteCloser) (retErr error) {
	ctx, done := log.SpanContext(ctx, "dumpDatabase")
	defer done(log.Errorp(&retErr))

	zw, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedBetterCompression))
	if err != nil {
		return errors.Wrap(err, "new zstd encoder")
	}
	defer errors.Close(&retErr, zw, "close zstd encoder")

	dsn, err := pachsql.ConnStringFromConn(ctx, db)
	if err != nil {
		return errors.Wrap(err, "get psql/pg_dump connection string from existing database connection")
	}
	cmd := exec.CommandContext(ctx, pgDumpPath(), "-d", dsn, "-v")
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stdout = zw
	cmd.Stderr = log.WriterAt(ctx, log.DebugLevel)
	cmd.Env = cmd.Environ()
	if extra := bazel.LibraryPath(cmd.Environ(), libpq, libldap, liblber, libsasl); extra != "" {
		cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+extra)
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "run pg_dump")
	}
	return nil
}

// CreateSnapshot creates a snapshot.
//
// TODO(jrockway): Clean up the half-created snapshot if we return with an error.
func CreateSnapshot(rctx context.Context, db *pachsql.DB, s *fileset.Storage) (id SnapshotID, retErr error) {
	rctx, done := log.SpanContext(rctx, "CreateSnapshot")
	defer done(log.Errorp(&retErr))

	// Add snapshot row.
	log.Debug(rctx, "adding snapshot row")
	if err := dbutil.WithTx(rctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		id, err = createSnapshotRow(ctx, tx, s)
		if err != nil {
			return errors.Wrap(err, "createSnapshotRow")
		}
		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "add snapshot row")
	}

	// Dump database and upload to PFS.
	log.Debug(rctx, "dumping database", zap.Stringer("snapshot", id))
	dctx, cancel := pctx.WithCancel(rctx)
	doneCh := make(chan error)
	r, w := io.Pipe()
	defer r.Close() //nolint:errcheck

	// Background job to feed database dump bytes into `r`.
	go func(ctx context.Context) {
		defer close(doneCh)
		defer cancel()
		err := dumpDatabase(ctx, db, w)
		w.CloseWithError(err) //nolint:errcheck
		select {
		case doneCh <- errors.Wrap(err, "dumpDatabase"):
		case <-ctx.Done():
			log.Debug(ctx, "database dump errored after context was done", zap.Error(err))
		}
	}(dctx)

	// Upload the database dump bytes as they arrive.
	var closedFileSet bool
	fw := s.NewWriter(rctx)
	defer func() {
		if closedFileSet {
			return
		}
		if _, err := fw.Close(); err != nil {
			errors.JoinInto(&retErr, errors.Wrap(err, "closed abandoned fileset"))
		}
	}()
	if err := fw.Add("dump.sql.zst", "", r); err != nil {
		cancel() // Kill the database dumper process if it's not dead yet.
		return 0, errors.Wrap(err, "create fileset containing database dump")
	}
	select {
	case <-rctx.Done():
		return 0, errors.Wrap(context.Cause(rctx), "context expired while waiting for dump to finish")
	case err, ok := <-doneCh:
		if err != nil {
			return 0, errors.Wrap(err, "wait for database dump to complete")
		}
		if !ok {
			return 0, errors.New("context expired before database dump could complete")
		}
	}
	closedFileSet = true
	fsHandle, err := fw.Close()
	if err != nil {
		return 0, errors.Wrap(err, "close finished database dump fileset")
	}
	if fsHandle == nil {
		return 0, errors.New("fileset handle was nil")
	}

	// Adjust the snapshot to reference this fileset.
	log.Debug(rctx, "adjusting snapshot to reference database dump fileset", zap.Stringer("snapshot", id), zap.String("fileset_token", fsHandle.Token().HexString()))
	if err := dbutil.WithTx(rctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		if err := addDatabaseDump(ctx, tx, id, fsHandle.Token()); err != nil {
			return errors.Wrapf(err, "addDatabaseDump(%v)", fsHandle.Token().HexString())
		}
		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "update snapshot row with database dump")
	}

	// Done.
	return id, nil
}
