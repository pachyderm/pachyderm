//go:build unit_test

package v2_7_0

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestCore(t *testing.T) {
	ctx := pctx.TestContext(t)
	db, _ := dockertestenv.NewEphemeralPostgresDB(ctx, t)
	defer db.Close()

	// TODO: pre-migration state
	// TODO: populate pre-migration state
	// run migration
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	require.NoError(t, SetupCore(ctx, tx))

	require.NoError(t, tx.Commit())

	// TODO: check post-migration state
}
