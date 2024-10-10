// Package recovery implements subsystem-independent disaster recovery.
package recovery

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/icholy/replace"
	"github.com/jmoiron/sqlx"
	"github.com/klauspost/compress/zstd"
	"github.com/pachyderm/pachyderm/v2/src/internal/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/version"
	uuid "github.com/satori/go.uuid"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
	"golang.org/x/text/transform"
)

var (
	// Variables replaced by the linker during Bazel builds.
	pgdump, psql, libpq, libldap, liblber, libsasl string
)

const (
	SQLDumpFilename = "dump.sql.zst"
)

func pgDumpPath() string {
	x, err := runfiles.Rlocation(pgdump)
	if err != nil || pgdump == "" {
		return "pg_dump" // Probably not built with bazel, look in $PATH and hope for the best.
	}
	return x
}

func psqlPath() string {
	x, err := runfiles.Rlocation(psql)
	if err != nil || pgdump == "" {
		return "psql"
	}
	return x
}

type SnapshotID int64

type snapshot struct {
	ID               SnapshotID `db:"id"`
	ChunksetID       int64      `db:"chunkset_id"`
	PachydermVersion string     `db:"pachyderm_version"`
	SQLDumpFileSetID uuid.UUID  `db:"sql_dump_fileset_id"`
}

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
func addDatabaseDump(ctx context.Context, tx *sqlx.Tx, snapshotID SnapshotID, filesetID fileset.ID) error {
	result, err := tx.ExecContext(ctx, `update recovery.snapshots set sql_dump_fileset_id=$1 where id=$2`, filesetID[:], snapshotID)
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

func getSnapshotRow(ctx context.Context, tx *sqlx.Tx, id SnapshotID) (result snapshot, _ error) {
	if err := sqlx.GetContext(ctx, tx, &result, `select id, chunkset_id, pachyderm_version, sql_dump_fileset_id from recovery.snapshots where id=$1`, id); err != nil {
		return snapshot{}, errors.Wrap(err, "select row")
	}
	return
}

// dumpDatabase runs pg_dump on the provided database, writing the zstd-compressed content to w.
func dumpDatabase(ctx context.Context, db *pachsql.DB, w io.Writer) (retErr error) {
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
	cmd := exec.CommandContext(ctx, pgDumpPath(), "-d", dsn, "-v", "--clean", "--if-exists", "--serializable-deferrable")
	cmd.Stdin = nil
	cmd.Stdout = transform.NewWriter(zw, replace.String("SET transaction_timeout = 0;", ""))
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "pg_dump.stderr"), log.DebugLevel)
	cmd.Env = cmd.Environ()
	if extra := bazel.LibraryPath(cmd.Environ(), libpq, libldap, liblber, libsasl); extra != "" {
		cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+extra)
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "run pg_dump")
	}
	return nil
}

// Snapshotter can take and restore Pachyderm snapshots.
type Snapshotter struct {
	DB         *pachsql.DB      // Databaase connection.
	Storage    *fileset.Storage // Fileset storage.
	EtcdClient *etcd.Client     // etcd client (for running migrations).
}

// CreateSnapshot creates a snapshot.
//
// TODO(jrockway): Clean up the half-created snapshot if we return with an error.
func (s *Snapshotter) CreateSnapshot(rctx context.Context) (id SnapshotID, retErr error) {
	rctx, done := log.SpanContext(rctx, "CreateSnapshot")
	defer done(log.Errorp(&retErr))

	// Add snapshot row.
	log.Debug(rctx, "adding snapshot row")
	if err := dbutil.WithTx(rctx, s.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		id, err = createSnapshotRow(ctx, tx, s.Storage)
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
		err := dumpDatabase(ctx, s.DB, w)
		w.CloseWithError(err) //nolint:errcheck
		select {
		case doneCh <- errors.Wrap(err, "dumpDatabase"):
		case <-ctx.Done():
			log.Debug(ctx, "database dump errored after context was done", zap.Error(err))
		}
	}(dctx)

	// Upload the database dump bytes as they arrive.
	var closedFileSet bool
	fw := s.Storage.NewWriter(rctx)
	defer func() {
		if closedFileSet {
			return
		}
		if _, err := fw.Close(); err != nil {
			errors.JoinInto(&retErr, errors.Wrap(err, "closed abandoned fileset"))
		}
	}()
	if err := fw.Add(SQLDumpFilename, "", r); err != nil {
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
	fsID, err := fw.Close()
	if err != nil {
		return 0, errors.Wrap(err, "close finished database dump fileset")
	}
	if fsID == nil {
		return 0, errors.New("fileset ID was nil")
	}

	// Adjust the snapshot to reference this fileset.
	log.Debug(rctx, "adjusting snapshot to reference database dump fileset", zap.Stringer("snapshot", id), zap.String("fileset_id", fsID.HexString()))
	if err := dbutil.WithTx(rctx, s.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		if err := addDatabaseDump(ctx, tx, id, *fsID); err != nil {
			return errors.Wrapf(err, "addDatabaseDump(%v)", fsID.HexString())
		}
		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "update snapshot row with database dump")
	}

	// Done.
	return id, nil
}

// restoreDatabase runs psql on the provided pg_dump output, reading zstd-compressed database
// content from r.
func restoreDatabase(ctx context.Context, db *pachsql.DB, r io.Reader) (retErr error) {
	ctx, done := log.SpanContext(ctx, "restoreDatabase")
	defer done(log.Errorp(&retErr))

	zr, err := zstd.NewReader(r)
	if err != nil {
		return errors.Wrap(err, "new zstd decoder")
	}
	defer zr.Close() // cannot error

	dsn, err := pachsql.ConnStringFromConn(ctx, db)
	if err != nil {
		return errors.Wrap(err, "get psql connection string from existing database connection")
	}
	cmd := exec.CommandContext(ctx, psqlPath(), "-d", dsn, "--single-transaction", "--set", "ON_ERROR_STOP=on")
	cmd.Stdin = zr
	cmd.Stdout = log.WriterAt(pctx.Child(ctx, "psql.stdout"), log.DebugLevel)
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "psql.stderr"), log.DebugLevel)
	cmd.Env = cmd.Environ()
	if extra := bazel.LibraryPath(cmd.Environ(), libpq, libldap, liblber, libsasl); extra != "" {
		cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+extra)
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "run psql")
	}
	return nil
}

// RestoreSnapshotOptions controls the behavior of the RestoreSnapshot function.
type RestoreSnapshotOptions struct {
	IgnoreVersionCompatibility bool // If true, allow restoring newer database dumps into an older Pachyderm.
}

// RestoreSnapshot restores the database state to that represented by the provided snapshot.  After
// the restore, the snapshot will remain restorable, even though it did not technically exist at the
// time it was created ;)
func (s *Snapshotter) RestoreSnapshot(rctx context.Context, id SnapshotID, opts RestoreSnapshotOptions) (retErr error) {
	rctx, done := log.SpanContext(rctx, "RestoreSnapshot")
	defer done(log.Errorp(&retErr))

	log.Debug(rctx, "reading snapshot metadata")
	var snap snapshot
	if err := dbutil.WithTx(rctx, s.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		snap, err = getSnapshotRow(ctx, tx, id)
		if err != nil {
			return errors.Wrap(err, "get snapshot row")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "read metadata: WithTx")
	}
	log.Debug(rctx, "got metadata", zap.String("pachyderm_version", snap.PachydermVersion), zap.Int64("chunkset_id", snap.ChunksetID), zap.Stringer("sql_dump_fileset_id", snap.SQLDumpFileSetID))

	if !opts.IgnoreVersionCompatibility {
		switch semver.Compare(snap.PachydermVersion, version.Version.String()) {
		// TODO(jrockway): Handle v0.0.0.
		case 1:
			return errors.Errorf("pachyderm version of snapshot %v is newer than the current running version %v", snap.PachydermVersion, version.Version.String())
		}
		log.Debug(rctx, "database dump is compatible with running pachyderm", zap.String("snapshot_version", snap.PachydermVersion), zap.Stringer("running_version", version.Version))
	}

	log.Debug(rctx, "downloading database dump to temporary file")
	fs, err := s.Storage.Open(rctx, []fileset.ID{fileset.ID(snap.SQLDumpFileSetID)})
	if err != nil {
		return errors.Wrapf(err, "open sql dump fileset %s", snap.SQLDumpFileSetID)
	}
	var tmp string
	if err := fs.Iterate(rctx, func(f fileset.File) (retErr error) {
		path := f.Index().Path
		log.Debug(rctx, "reading file from database dump fileset", zap.Stringer("fileset_id", snap.SQLDumpFileSetID), zap.String("path", path))
		if got, want := path, SQLDumpFilename; got != want {
			return errors.Errorf("unexpected file in database dump fileset %s: got %v want %v", snap.SQLDumpFileSetID, got, want)
		}
		fh, err := os.CreateTemp("", fmt.Sprintf("snapshot-%v-*", id))
		if err != nil {
			return errors.Wrapf(err, "create tmp file to store database dump")
		}
		defer errors.Close(&retErr, fh, "close tmp database dump file")
		if err := f.Content(rctx, fh); err != nil {
			return errors.Wrapf(err, "read file %v content", path)
		}
		log.Debug(rctx, "finished reading database dump ok", zap.String("path", path))
		tmp = fh.Name()
		return nil
	}); err != nil {
		return errors.Wrapf(err, "iterate over sql dump fileset %s", snap.SQLDumpFileSetID)
	}
	defer errors.Invoke1(&retErr, os.Remove, tmp, "cleanup database dump tmp file")
	log.Debug(rctx, "downloaded database dump ok", zap.String("path", tmp))

	log.Debug(rctx, "restoring database")
	restore, err := os.Open(tmp)
	if err != nil {
		return errors.Wrap(err, "open database dump (to restore)")
	}
	defer errors.Close(&retErr, restore, "close database dump file (opened for restore)")
	if err := restoreDatabase(rctx, s.DB, restore); err != nil {
		return errors.Wrap(err, "restore database")
	}
	log.Debug(rctx, "finished restoring database")

	log.Debug(rctx, "running migrations")
	menv := migrations.Env{
		WithTableLocks: true,
		EtcdClient:     s.EtcdClient,
	}
	if err := migrations.ApplyMigrations(rctx, s.DB, menv, clusterstate.DesiredClusterState); err != nil {
		return errors.Wrap(err, "apply database migrations from snapshot version to running version")
	}
	log.Debug(rctx, "ran migrations ok")

	log.Debug(rctx, "saving database dump to PFS")
	save, err := os.Open(tmp)
	if err != nil {
		return errors.Wrap(err, "open database dump (to save)")
	}
	defer errors.Close(&retErr, save, "close database dump file (opened to save)")
	fw := s.Storage.NewWriter(rctx)
	if err := fw.Add(SQLDumpFilename, "", save); err != nil {
		if _, closeErr := fw.Close(); closeErr != nil {
			errors.JoinInto(&err, errors.Wrap(closeErr, "close fileset writer"))
		}
		return errors.Wrap(err, "add database dump fileset content")
	}
	fsID, err := fw.Close()
	if err != nil {
		return errors.Wrap(err, "close database dump fileset writer")
	}
	log.Debug(rctx, "saved databse dump to PFS ok", zap.String("fileset_id", fsID.HexString()))

	log.Debug(rctx, "adding dump fileset to newly-restored snapshot row")
	if err := dbutil.WithTx(rctx, s.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		if err := addDatabaseDump(ctx, tx, id, *fsID); err != nil {
			return errors.Wrapf(err, "edit snapshot id=%v to contain fileset %s", id, snap.SQLDumpFileSetID)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "update snapshot: WithTx")
	}
	log.Debug(rctx, "snapshot state updated ok")

	return nil
}
