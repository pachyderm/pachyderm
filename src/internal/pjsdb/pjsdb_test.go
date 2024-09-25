package pjsdb_test

import (
	"bytes"
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"math/rand"
	"path/filepath"
	"strconv"
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

func DB(t testing.TB) (context.Context, *pachsql.DB) {
	t.Helper()
	ctx := pctx.Child(pctx.TestContext(t), t.Name())
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	return ctx, db
}

func FilesetStorage(t testing.TB, db *pachsql.DB) *fileset.Storage {
	t.Helper()
	store := kv.NewFSStore(filepath.Join(t.TempDir(), "obj-store"), 512, chunk.DefaultMaxChunkSize)
	tr := track.NewPostgresTracker(db)
	return fileset.NewStorage(fileset.NewPostgresStore(db), tr, chunk.NewStorage(store, nil, db, tr))
}

func withTx(t testing.TB, ctx context.Context, db *pachsql.DB, s *fileset.Storage, f func(d dependencies)) {
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

func mockFileset(t testing.TB, d dependencies, path, value string) fileset.PinnedFileset {
	var id *fileset.ID
	uw, err := d.s.NewUnorderedWriter(d.ctx)
	require.NoError(t, err)
	err = uw.Put(d.ctx, path, "", true, bytes.NewReader([]byte(value)))
	require.NoError(t, err)
	id, err = uw.Close(d.ctx)
	require.NoError(t, err)
	pin, err := d.s.Pin(d.tx, *id)
	require.NoError(t, err)
	return pin
}

func mockAndHashFileset(t testing.TB, d dependencies, path, value string) (fileset.PinnedFileset, []byte) {
	fs := mockFileset(t, d, path, value)
	hasher := pachhash.New()
	_, err := hasher.Write([]byte(path))
	require.NoError(t, err)
	_, err = hasher.Write([]byte(strconv.Itoa(len(path))))
	require.NoError(t, err)
	_, err = hasher.Write([]byte(value))
	require.NoError(t, err)
	hash := hasher.Sum(nil)
	return fs, hash
}

func makeReq(t *testing.T, d dependencies, parent pjsdb.JobID, mutate func(req *pjsdb.CreateJobRequest)) pjsdb.CreateJobRequest {
	program := `!#/bin/bash; ls /input/;`
	programFileset, hash := mockAndHashFileset(t, d, "/program.py", program)
	numInputs := rand.Intn(10) + 1
	inputs := make([]fileset.PinnedFileset, 0)
	inputHashes := make([][]byte, 0)
	for i := 0; i < numInputs; i++ {
		inputFileset, inputHash := mockAndHashFileset(t, d, "/input/"+strconv.Itoa(i)+".txt", `pachyderm`)
		inputs = append(inputs, inputFileset)
		inputHashes = append(inputHashes, inputHash)
	}
	createRequest := &pjsdb.CreateJobRequest{
		Program:           programFileset,
		ProgramHash:       hash,
		Inputs:            inputs,
		InputHashes:       inputHashes,
		Parent:            parent,
		CacheWriteEnabled: true,
		CacheReadEnabled:  true,
	}
	if mutate != nil {
		mutate(createRequest)
	}
	return *createRequest
}

func createJobWithFilesets(t testing.TB, d dependencies, parent pjsdb.JobID, program fileset.PinnedFileset,
	targetHash []byte, inputs ...fileset.PinnedFileset) pjsdb.JobID {
	createRequest := pjsdb.CreateJobRequest{
		Program:     program,
		ProgramHash: targetHash,
		Inputs:      inputs,
		Parent:      parent,
	}
	id, err := pjsdb.CreateJob(d.ctx, d.tx, createRequest)
	require.NoError(t, err)
	return id
}
