package clusterstate

import (
	"context"
	"crypto/rand"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"strconv"
	"testing"
)

func Test_v2_11_0_ClusterState(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDirectDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}

	// Pre-migration
	// Note that we are applying 2.6 migration here because we need to create collections.repos table
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_6_0))
	setupTestData(t, ctx, db)

	// Apply migrations up to and including 2.11.0
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_11_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_11_0))
}

func Test_v2_11_0_ClusterStateWithDanglingCommits(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDirectDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}

	// Pre-migration
	// Note that we are applying 2.6 migration here because we need to create collections.repos table
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_6_0))
	setupTestData(t, ctx, db)
	addDanglingTotals(t, ctx, db)

	// Apply migrations up to and including 2.11.0
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_11_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_11_0))
}

func newFilesetId() fileset.ID {
	id := fileset.ID{}
	if _, err := rand.Read(id[:]); err != nil {
		panic(err)
	}
	return id
}

func addDanglingTotals(t *testing.T, ctx context.Context, db *sqlx.DB) {
	for i := 1; i <= 5; i++ {
		_, err := db.ExecContext(ctx, `INSERT INTO pfs.commit_totals (commit_id, fileset_id) VALUES ($1, $2)`, "fake_commit_"+strconv.Itoa(i), newFilesetId())
		require.NoError(t, err)
	}
}
