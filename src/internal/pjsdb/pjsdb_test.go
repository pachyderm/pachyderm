package pjsdb_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

type dependencies struct {
	ctx context.Context
	db  *pachsql.DB
	tx  *pachsql.Tx
	s   *fileset.Storage
}

func DB(t *testing.T) (context.Context, *pachsql.DB) {
	t.Helper()
	ctx := pctx.Child(pctx.TestContext(t), t.Name())
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	return ctx, db
}

func FilesetStorage(t *testing.T, db *pachsql.DB) *fileset.Storage {
	t.Helper()
	store := kv.NewFSStore(filepath.Join(t.TempDir(), "obj-store"), 512, chunk.DefaultMaxChunkSize)
	tr := track.NewPostgresTracker(db)
	return fileset.NewStorage(fileset.NewPostgresStore(db), tr, chunk.NewStorage(store, nil, db, tr))
}

func withTx(t *testing.T, ctx context.Context, db *pachsql.DB, s *fileset.Storage, f func(d dependencies)) {
	t.Helper()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	f(dependencies{ctx: ctx, db: db, tx: tx, s: s})
	require.NoError(t, tx.Commit())
}

func withDependencies(t *testing.T, f func(d dependencies)) {
	ctx, db := DB(t)
	withTx(t, ctx, db, FilesetStorage(t, db), func(d dependencies) {
		f(d)
	})
}

func mockFileset(t *testing.T, d dependencies, path, value string) *fileset.ID {
	var id *fileset.ID
	uw, err := d.s.NewUnorderedWriter(d.ctx)
	require.NoError(t, err)
	err = uw.Put(d.ctx, path, "", true, bytes.NewReader([]byte(value)))
	require.NoError(t, err)
	id, err = uw.Close(d.ctx)
	require.NoError(t, err)
	return id
}

func makeReq(t *testing.T, d dependencies, parent pjsdb.JobID, mutate func(req *pjsdb.CreateJobRequest)) *pjsdb.CreateJobRequest {
	spec := `!#/bin/bash; ls /input/;`
	specFilesetID := mockFileset(t, d, "/spec.py", spec)
	inputFilesetID := mockFileset(t, d, "/input/0.txt", `pachyderm`)
	createRequest := &pjsdb.CreateJobRequest{
		Spec:   specFilesetID.HexString(),
		Input:  inputFilesetID.HexString(),
		Parent: parent,
	}
	if mutate != nil {
		mutate(createRequest)
	}
	return createRequest
}

func createJobWithFilesets(t *testing.T, d dependencies, parent pjsdb.JobID, spec, input *fileset.ID) pjsdb.JobID {
	createRequest := &pjsdb.CreateJobRequest{
		Spec:   spec.HexString(),
		Parent: parent,
	}
	if input != nil {
		createRequest.Input = input.HexString()
	}
	req, err := createRequest.Sanitize(d.ctx, d.tx, d.s)
	require.NoError(t, err)
	id, err := pjsdb.CreateJob(d.ctx, d.tx, req)
	require.NoError(t, err)
	return id
}
