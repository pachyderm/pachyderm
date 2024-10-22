package snapshot

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

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
	"github.com/pachyderm/pachyderm/v2/src/internal/pgjsontypes"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshotdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
	"github.com/pachyderm/pachyderm/v2/src/version"
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
func createSnapshotRow(ctx context.Context, tx *sqlx.Tx, s *fileset.Storage, metadata pgjsontypes.StringMap) (result SnapshotID, sql string, _ error) {
	var b strings.Builder // b contains SQL to recreate this function call inside a psql script.

	// Create (and dump) chunkset.
	chunksetID, err := s.CreateChunkSet(ctx, tx)
	if err != nil {
		return 0, "", errors.Wrap(err, "create chunkset")
	}
	// Write out psql to create only this chunkset.  (Pre-existing chunksets and the primary key
	// sequence are handled by pg_dump.)
	b.WriteRune('\n')
	b.WriteString("COPY storage.chunksets (id) FROM stdin;\n")
	fmt.Fprintf(&b, "%v\n", int64(chunksetID))
	b.WriteString(`\.` + "\n\n")

	// Create (and dump) snapshot row.
	if err := tx.GetContext(ctx, &result, `insert into recovery.snapshots (chunkset_id, pachyderm_version, metadata) values ($1, $2, $3) returning id`, chunksetID, version.Version.Canonical(), metadata); err != nil {
		return 0, "", errors.Wrap(err, "create snapshot row")
	}
	// Write out psql to create only this snapshot row.  (Pre-existing snapshot rows are dumped
	// by pg_dump.)
	b.WriteString("COPY recovery.snapshots (id, chunkset_id, pachyderm_version, metadata) FROM stdin;\n")
	js := `{}` // TODO(MLDM-142): escape JSON for this
	fmt.Fprintf(&b, "%v\t%v\t%v\t%s\n", int64(result), int64(chunksetID), version.Version.Canonical(), js)
	b.WriteString(`\.` + "\n\n")

	return result, b.String(), nil
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

// Snapshotter can take and restore Pachyderm snapshots.
type Snapshotter struct {
	DB         *pachsql.DB      // Databaase connection.
	Storage    *fileset.Storage // Fileset storage.
	EtcdClient *etcd.Client     // etcd client (for running migrations).
}

// dumpDatabase runs pg_dump on the provided database, writing the content to w.
func (s *Snapshotter) dumpDatabase(ctx context.Context, snapshot string, w io.Writer) (retErr error) {
	ctx, done := log.SpanContext(ctx, "dumpDatabase")
	defer done(log.Errorp(&retErr))

	dsn, err := pachsql.ConnStringFromConn(ctx, s.DB)
	if err != nil {
		return errors.Wrap(err, "get psql/pg_dump connection string from existing database connection")
	}
	cmd := exec.CommandContext(ctx, pgDumpPath(), "-d", dsn, "-v", "--clean", "--if-exists", "--snapshot", snapshot, "--exclude-table-data="+s.Storage.DumpTrackerTablePattern())
	cmd.Stdin = nil
	tw := transform.NewWriter(w, replace.String("SET transaction_timeout = 0;", ""))
	defer errors.Close(&retErr, tw, "close transform writer")
	cmd.Stdout = tw
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

// dumpTx dumps the database in a single transaction.  It creates a chunkset, dumps the database
// with pg_dump in the same transaction, dumps the sql to create the chunkset (which pg_dump can't
// see, because it can't read writes in this txn), and dumps the storage tracker (for the same
// reason).  This yields a completely atomic snapshot; no tracker entries or chunks can change
// during the dump.  (If they do, the txn rolls back with a serialization error.)
func (s *Snapshotter) dumpTx(ctx context.Context, tx *pachsql.Tx, opts CreateSnapshotOptions) (id SnapshotID, dumpFile *os.File, retErr error) {
	ctx, done := log.SpanContext(ctx, "snapshotTx")
	defer done(log.Errorp(&retErr))

	// Add snapshot row.
	log.Debug(ctx, "adding snapshot row")
	id, snapshotSQL, err := createSnapshotRow(ctx, tx, s.Storage, opts.Metadata)
	if err != nil {
		return 0, nil, errors.Wrap(err, "createSnapshotRow")
	}
	log.Debug(ctx, "snapshot row added; getting pg_snapshot id", zap.Stringer("snapshot_id", id))

	// Get the postgres tx snapshot ID.
	var snapshot string
	if err := sqlx.GetContext(ctx, tx, &snapshot, `select pg_export_snapshot()`); err != nil {
		return 0, nil, errors.Wrap(err, "select pg_export_snapshot()")
	}
	log.Debug(ctx, "pg_snapshot ok", zap.String("pg_snapshot", snapshot))

	// Create a temp file for the database dump.
	fh, err := os.CreateTemp("", "create-snapshot")
	if err != nil {
		return 0, nil, errors.Wrap(err, "create tmp file for database dump")
	}
	defer func() {
		if retErr != nil {
			errors.JoinInto(&retErr, errors.Wrap(os.Remove(fh.Name()), "delete unused tmp file"))
			errors.JoinInto(&retErr, errors.Wrap(fh.Close(), "close unused tmp file"))
		}
	}()
	zw, err := zstd.NewWriter(fh, zstd.WithEncoderLevel(zstd.SpeedBetterCompression))
	if err != nil {
		return 0, nil, errors.Wrap(err, "new zstd encoder")
	}
	defer errors.Close(&retErr, zw, "close zstd encoder")

	// Dump the database, starting at the same postgres snapshot as this transaction is using.
	log.Debug(ctx, "dumping database", zap.Stringer("snapshot_id", id), zap.String("pg_snapshot", snapshot))
	if err := s.dumpDatabase(ctx, snapshot, zw); err != nil {
		return 0, nil, errors.Wrap(err, "dump database")
	}
	log.Debug(ctx, "database dump finished ok", zap.String("dump_file", fh.Name()))

	// Add new information added in this tx (by this package) to the dump.
	log.Debug(ctx, "adding this tx to the dump")
	if _, err := io.Copy(zw, strings.NewReader(snapshotSQL)); err != nil {
		return 0, nil, errors.Wrap(err, "add footer")
	}

	// Add new information added in this tx (by the storage system) to the dump.
	buf := bufio.NewWriter(zw)
	if err := s.Storage.DumpTracker(ctx, tx, buf); err != nil {
		return 0, nil, errors.Wrap(err, "dump tracker")
	}
	if err := buf.Flush(); err != nil {
		return 0, nil, errors.Wrap(err, "flush buffer")
	}
	log.Debug(ctx, "added this tx to dump ok")
	return id, fh, nil
}

// CreateSnapshotOptions controls snapshotting behavior.
type CreateSnapshotOptions struct {
	Metadata pgjsontypes.StringMap // Metadata to add to the snapshot.
}

// CreateSnapshot creates a snapshot.
func (s *Snapshotter) CreateSnapshot(rctx context.Context, opts CreateSnapshotOptions) (_ SnapshotID, retErr error) {
	rctx, done := log.SpanContext(rctx, "CreateSnapshot")
	defer done(log.Errorp(&retErr))

	var id SnapshotID
	var fh *os.File
	defer func() {
		if fh != nil {
			errors.JoinInto(&retErr, errors.Wrap(os.Remove(fh.Name()), "delete tmp file"))
			errors.JoinInto(&retErr, errors.Wrap(fh.Close(), "close tmp file"))
		}
	}()
	if err := dbutil.WithTx(rctx, s.DB, func(ctx context.Context, tx *pachsql.Tx) (retErr error) {
		var err error
		id, fh, err = s.dumpTx(ctx, tx, opts)
		if err != nil {
			return errors.Wrap(err, "dumpTx")
		}
		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "WithTx(snapshotBgTx)")
	}

	// Seek to the start of the database dump.
	if _, err := fh.Seek(0, 0); err != nil {
		return 0, errors.Wrap(err, "seek to start of database dump")
	}

	// Upload the database dump bytes.
	log.Debug(rctx, "uploading database dump fileset")
	var closedFileSet bool
	fw := s.Storage.NewWriter(rctx)
	defer func() {
		if closedFileSet {
			return
		}
		if _, err := fw.Close(); err != nil {
			errors.JoinInto(&retErr, errors.Wrap(err, "close abandoned fileset"))
		}
	}()
	if err := fw.Add(SQLDumpFilename, "", fh); err != nil {
		return 0, errors.Wrap(err, "create fileset containing database dump")
	}
	closedFileSet = true
	fsHandle, err := fw.Close()
	if err != nil {
		return 0, errors.Wrap(err, "close finished database dump fileset")
	}
	if fsHandle == nil {
		return 0, errors.New("fileset handle was nil")
	}
	log.Debug(rctx, "database dump fileset uploaded ok", zap.String("fileset_token", fsHandle.Token().HexString()))

	// Adjust the snapshot to reference this fileset.
	log.Debug(rctx, "adjusting snapshot to reference database dump fileset", zap.Stringer("snapshot", id), zap.String("fileset_token", fsHandle.Token().HexString()))
	if err := dbutil.WithTx(rctx, s.DB, func(ctx context.Context, tx *pachsql.Tx) error {
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

var removeGitVersion = regexp.MustCompile(`-pre.g([0-9a-f]{10})$`)

func checkVersionCompatibility(snapshot, running string) error {
	if !strings.HasPrefix(snapshot, "v") || !strings.HasPrefix(running, "v") {
		// No version information available.
		return nil
	}
	// semver.Compare treats -pre.g<commit> as something that's ordered, but it's not.
	snapshot = removeGitVersion.ReplaceAllString(snapshot, "-pre.gXXX")
	running = removeGitVersion.ReplaceAllString(running, "-pre.gXXX")
	switch semver.Compare(snapshot, running) {
	case 1:
		return errors.Errorf("pachyderm version of snapshot (%v) is newer than the current running version (%v)", snapshot, running)
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
	var snap *snapshot.SnapshotInfo
	if err := dbutil.WithTx(rctx, s.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		snap, err = snapshotdb.GetSnapshot(ctx, tx, int64(id))
		if err != nil {
			return errors.Wrap(err, "get snapshot row")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "read metadata: WithTx")
	}
	log.Debug(rctx, "got metadata", log.Proto("snapshot", snap))

	if !opts.IgnoreVersionCompatibility {
		if err := checkVersionCompatibility(snap.PachydermVersion, version.Version.Canonical()); err != nil {
			return err
		}
		log.Debug(rctx, "database dump is compatible with running pachyderm", zap.String("snapshot_version", snap.PachydermVersion), zap.String("running_version", version.Version.Canonical()))
	}

	var token fileset.Token
	if err := token.Scan(snap.GetSqlDumpFilesetId()); err != nil {
		return errors.Wrapf(err, "parse fileset token %v", snap.GetSqlDumpFilesetId())
	}

	log.Debug(rctx, "downloading database dump to temporary file")
	fs, err := s.Storage.Open(rctx, []*fileset.Handle{fileset.NewHandle(token)})
	if err != nil {
		return errors.Wrapf(err, "open sql dump fileset %s", snap.GetSqlDumpFilesetId())
	}
	var tmp string
	if err := fs.Iterate(rctx, func(f fileset.File) (retErr error) {
		path := f.Index().Path
		log.Debug(rctx, "reading file from database dump fileset", zap.String("fileset_id", snap.GetSqlDumpFilesetId()), zap.String("path", path))
		if got, want := path, SQLDumpFilename; got != want {
			return errors.Errorf("unexpected file in database dump fileset %s: got %v want %v", snap.GetSqlDumpFilesetId(), got, want)
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
		return errors.Wrapf(err, "iterate over sql dump fileset %s", snap.GetSqlDumpFilesetId())
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
	fsHandle, err := fw.Close()
	if err != nil {
		return errors.Wrap(err, "close database dump fileset writer")
	}
	log.Debug(rctx, "saved databse dump to PFS ok", zap.String("fileset_token", fsHandle.Token().HexString()))

	log.Debug(rctx, "adding dump fileset to newly-restored snapshot row")
	if err := dbutil.WithTx(rctx, s.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		if err := addDatabaseDump(ctx, tx, id, fsHandle.Token()); err != nil {
			return errors.Wrapf(err, "edit snapshot id=%v to contain fileset %s", id, snap.GetSqlDumpFilesetId())
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "update snapshot: WithTx")
	}
	log.Debug(rctx, "snapshot state updated ok")

	return nil
}
