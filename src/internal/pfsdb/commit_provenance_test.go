package pfsdb

import (
	"context"
	"sort"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestCommitSetProvenance(t *testing.T) {
	db, _ := dockertestenv.NewEphemeralPostgresDB(t)
	defer db.Close()
	ctx := context.Background()
	// setup schema
	tx, err := db.Beginx()
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `CREATE SCHEMA pfs`)
	require.NoError(t, err)
	require.NoError(t, SetupCommitProvenanceV0(ctx, tx))
	require.NoError(t, tx.Commit())
	// setup basic commit graph of
	//                -- C@y
	//               /
	// A@v <- B@x <--
	//               \
	//                -- D@z
	//               /
	// E@w <---------
	tx, err = db.Beginx()
	require.NoError(t, err)
	require.NoError(t, AddCommit(ctx, tx, "A@v", "v"))
	require.NoError(t, AddCommit(ctx, tx, "B@x", "x"))
	require.NoError(t, AddCommitProvenance(ctx, tx, "B@x", "A@v"))
	require.NoError(t, AddCommit(ctx, tx, "C@y", "y"))
	require.NoError(t, AddCommitProvenance(ctx, tx, "C@y", "B@x"))
	require.NoError(t, AddCommit(ctx, tx, "E@w", "w"))
	require.NoError(t, AddCommit(ctx, tx, "D@z", "z"))
	require.NoError(t, AddCommitProvenance(ctx, tx, "D@z", "B@x"))
	require.NoError(t, AddCommitProvenance(ctx, tx, "D@z", "E@w"))
	require.NoError(t, tx.Commit())
	// assert commit set provenance
	tx, err = db.Beginx()
	require.NoError(t, err)
	defer tx.Commit()
	yProv, err := CommitSetProvenance(ctx, tx, "y")
	require.NoError(t, err)
	sort.Strings(yProv)
	require.ElementsEqual(t,
		[]string{"A@v", "B@x"},
		yProv)
	ySubv, err := CommitSetSubvenance(ctx, tx, "y")
	require.NoError(t, err)
	sort.Strings(ySubv)
	require.ElementsEqual(t,
		[]string{},
		ySubv)

	zProv, err := CommitSetProvenance(ctx, tx, "z")
	require.NoError(t, err)
	sort.Strings(zProv)
	require.ElementsEqual(t,
		[]string{"A@v", "B@x", "E@w"},
		zProv)

	xSubv, err := CommitSetSubvenance(ctx, tx, "x")
	require.NoError(t, err)
	sort.Strings(xSubv)
	require.ElementsEqual(t,
		[]string{"C@y", "D@z"},
		xSubv)
}
